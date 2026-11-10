package legacy

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/client"
	noxflags "github.com/opennox/opennox/v1/common/flags"
)

func TestGame3DoorByteAndPressurePlateReadOrder(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	noxflags.SetGame(1)
	defer func() { noxflags.ResetGame(); noxflags.SetGame(oldFlags) }()
	oldFactory := game3MapFactory
	defer func() { game3MapFactory = oldFactory }()
	cases := []struct {
		name   string
		kind   int
		status uint32
	}{{"door-zero", 0, 0}, {"door-high-word-only", 0, 0x100}, {"door-low-byte", 0, 1}, {"door-signed-byte", 0, 0x80}, {"pressure-width-height", 3, 0}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw := append([]byte{41, 0}, game3TestMapRecord(61, false)...)
			put := func(v uint32) { var p [4]byte; binary.LittleEndian.PutUint32(p[:], v); raw = append(raw, p[:]...) }
			if tc.kind == 0 {
				put(3)
				put(tc.status)
				put(0)
			} else {
				put(0xfffffff9)
				put(123)
				raw = append(raw, 1, 2, 3, 4, 5, 6)
			}
			reference, actual := game3TestDrawable(t), game3TestDrawable(t)
			for _, d := range []*client.Drawable{reference, actual} {
				d.Field_72 = 0x12345678
				d.Shape.Box.W = 9
				if tc.kind == 0 {
					d.ObjClass = object.Class(0x80)
				}
			}
			refCount, refCursor, refReads := game3TestReferenceKind(raw, reference, 17, tc.kind)
			h, cursor := game3TestMapIO(t, raw, actual, false)
			rw := h.rw
			observed := [][7]uint32{}
			h.rw = func(p []byte) {
				u := (*[2]byte)(unsafe.Pointer(&actual.Union))
				observed = append(observed, [7]uint32{uint32(*cursor), uint32(len(p)), actual.Field_72, uint32(actual.ObjFlags), math.Float32bits(actual.Shape.Box.W), uint32(u[0]), uint32(u[1])})
				rw(p)
			}
			h.attachWall = func(*client.Drawable, int32, int32) {}
			game3MapFactory = func() game3MapServices { return h }
			got := game3TestPublicLoad(17, tc.kind)
			if got != refCount || *cursor != refCursor || !bytes.Equal(game3TestSnapshot(actual), game3TestSnapshot(reference)) || !reflect.DeepEqual(observed, refReads) {
				t.Fatalf("current-C observable read/write order differs kind%d status%#x\nGo%v\nC%v", tc.kind, tc.status, observed, refReads)
			}
			if tc.kind == 0 && tc.status == 0x100 {
				u := (*[2]byte)(unsafe.Pointer(&actual.Union))
				if u[0] != 0 || u[1] != 0 {
					t.Fatal("door status must use the serialized low BYTE")
				}
			}
		})
	}
}

func TestGame3MovingSparkMixedSignedCoordinates(t *testing.T) {
	actual, reference := game3TestDrawable(t), game3TestDrawable(t)
	for _, d := range []*client.Drawable{actual, reference} {
		e := d.UnionEffect()
		e.Field_108 = 0xfffff000
		e.Field_109 = 0xfffff000
		e.Field_110 = 0
		e.Field_111 = 0
		e.Field_112 = 1
	}
	old := game3EffectsFactory
	defer func() { game3EffectsFactory = old }()
	calls := 0
	h := game3EffectServices{frame: func() uint32 { return 1 }, update: func(d *client.Drawable, p image.Point) { calls++; d.PosVec = p }, deleteSelf: func(*client.Drawable) {}}
	game3EffectsFactory = func() game3EffectServices { return h }
	want := game3TestReferenceMotion(reference)
	got := game3TestPublicMotion(actual)
	if got != want || calls != 1 || actual.PosVec != image.Pt(1048575, -1) || !bytes.Equal(game3TestSnapshot(actual), game3TestSnapshot(reference)) {
		t.Fatalf("DWORD X logical / signed temporary Y arithmetic: %+v ref%+v returns%d/%d", actual.PosVec, reference.PosVec, got, want)
	}
}

func TestGame3BounceLazyFrameReadAfterZ(t *testing.T) {
	old := game3EffectsFactory
	defer func() { game3EffectsFactory = old }()
	for _, tc := range []struct {
		z        int16
		velocity int8
		reads    int
		nextZ    int16
	}{{1, -3, 0, 2}, {-1, 3, 1, 2}, {0, 0, 1, 0}} {
		t.Run(fmt.Sprintf("z%d-v%d", tc.z, tc.velocity), func(t *testing.T) {
			d := game3TestDrawable(t)
			d.ZVal = uint16(tc.z)
			d.VelZ = tc.velocity
			reads := 0
			h := game3EffectServices{frame: func() uint32 {
				reads++
				if int16(d.ZVal) != tc.nextZ {
					t.Errorf("frame read before Z store: got%d want%d", int16(d.ZVal), tc.nextZ)
				}
				d.VelZ = 99
				return 1
			}}
			game3EffectsFactory = func() game3EffectServices { return h }
			game3TestPublicBounce(d)
			if reads != tc.reads {
				t.Fatalf("frame reads%d want%d", reads, tc.reads)
			}
			if tc.reads != 0 && d.VelZ != tc.velocity-1 {
				t.Fatal("velocity must be cached before frame getter")
			}
		})
	}
}
