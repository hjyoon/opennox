package legacy

import (
	"bytes"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
	"math"
	"reflect"
	"testing"
	"unsafe"
)

func game5HighObject(t *testing.T, code uint32) *server.Object {
	t.Helper()
	obj, free := alloc.NewHigh(server.Object{})
	t.Cleanup(free)
	obj.NetCode = code
	obj.Collide = game5HitCallback()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(obj)) <= math.MaxUint32 {
		t.Fatal("high native object required")
	}
	return obj
}
func TestGame5CollisionQueuePublicCNativeOwners(t *testing.T) {
	handles.InitHigh()
	t.Cleanup(handles.Release)
	first, second := game5HighObject(t, 255), game5HighObject(t, 2)
	saved := Nox_xxx_collide_4FDF90
	t.Cleanup(func() { Nox_xxx_collide_4FDF90 = saved; circleCollisionCABIReleaseFixturePool() })
	var enchantOrder []int
	Nox_xxx_collide_4FDF90 = func(a, b *server.Object) {
		enchantOrder = append(enchantOrder, game5HitCount())
		if a != first && a != second || b != first && b != second {
			t.Fatal("enchant owner identity")
		}
	}
	normal := types.Ptf(3, 4)
	game5HitInit(true)
	game5HitAdd(first, second, 0, &normal)
	game5HitAdd(second, first, 0, &types.Pointf{X: 100, Y: 200})
	game5HitAdd(first, second, 0, nil) // Duplicate detection precedes the normal read.
	game5HitFlush()
	if game5HitCount() != 2 || !reflect.DeepEqual(enchantOrder, []int{1, 2}) {
		t.Fatal("deduplication or ordered callbacks", game5HitCount(), enchantOrder)
	}
	a, b, n := game5HitRecord(0)
	if a != uintptr(unsafe.Pointer(first)) || b != uintptr(unsafe.Pointer(second)) || n != normal {
		t.Fatal("forward native C callback", a, b, n)
	}
	a, b, n = game5HitRecord(1)
	if a != uintptr(unsafe.Pointer(second)) || b != uintptr(unsafe.Pointer(first)) || n != types.Ptf(-7, 9) {
		t.Fatal("reverse callback must reload the live normal", a, b, n)
	}
	for token := uintptr(0); token <= 5; token++ {
		game5HitInit(false)
		enchantOrder = nil
		game5HitAdd(first, nil, token, &normal)
		game5HitFlush()
		want := 0
		if token == 0 {
			want = 1
		}
		if game5HitCount() != want || len(enchantOrder) != 0 {
			t.Fatal("wall token contract", token, game5HitCount(), enchantOrder)
		}
	}
}

func TestGame5DoorQueuePublicCHighData(t *testing.T) {
	for _, tc := range []struct {
		name                         string
		fraction, delta              int16
		velocity                     float32
		current, target, wantCurrent int32
		wantFraction                 int16
	}{
		{"wrap-positive", 254, 5, 0, 0, 0, 0, 3},
		{"wrap-negative", 2, -5, 0, 0, 0, 31, 253},
		{"advance", 255, 0, 2, 0, 0, 0, 3},
		{"retreat", 1, 0, -2, 0, 0, 31, 254},
		{"positive-stop", 100, 0, 2, 12, 0, 12, 100},
		{"negative-stop", 160, 0, -2, 20, 0, 20, 160},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw, free := alloc.MallocHigh(116)
			defer free()
			frame := unsafe.Slice((*byte)(raw), 116)
			for i := range frame {
				frame[i] = byte(13*i + 19)
			}
			d := (*server.DoorUpdateData)(unsafe.Add(raw, 32))
			d.Queued = 0
			d.FractionalDir = tc.fraction
			d.AngularVelocity = tc.velocity
			d.CurrentDirection = tc.current
			d.TargetDirection = tc.target
			want := bytes.Clone(frame)
			expected := (*server.DoorUpdateData)(unsafe.Pointer(&want[32]))
			expected.FractionalDir = tc.wantFraction
			expected.CurrentDirection = tc.wantCurrent
			expected.Queued = 0
			expected.AngularVelocity = 0
			o := game5HighObject(t, 1)
			o.UpdateData = unsafe.Pointer(d)
			defer func() { o.UpdateData = nil }()
			game5DoorDelta(o, tc.delta)
			game5DoorQueue(d) // A second enqueue must not apply velocity twice.
			if d.Queued != 1 {
				t.Fatal("queue marker")
			}
			game5DoorFlush()
			game5DoorFlush()
			if !bytes.Equal(frame, want) {
				t.Fatalf("whole native door owner differs: %+v want %+v", *d, *expected)
			}
		})
	}
}

func TestGame5GeometryPublicCOutputsAndGuards(t *testing.T) {
	segment, freeSegment := alloc.NewHigh([4]float32{})
	defer freeSegment()
	*segment = [4]float32{-2, 0, 2, 0}
	rect, freeRect := alloc.NewHigh([4]float32{})
	defer freeRect()
	*rect = [4]float32{-1, -1, 1, 1}
	raw, free := alloc.MallocHigh(64)
	defer free()
	frame := unsafe.Slice((*byte)(raw), 64)
	for i := range frame {
		frame[i] = 0xa5
	}
	out := (*types.Pointf)(unsafe.Add(raw, 16))
	if n := game5IntersectionC(segment, rect, out, 2); n != 2 {
		t.Fatal("crossing count", n)
	}
	points := unsafe.Slice(out, 2)
	if points[0] != types.Ptf(1, 0) || points[1] != types.Ptf(-1, 0) {
		t.Fatal("edge order", points)
	}
	for i, v := range frame {
		if (i < 16 || i >= 32) && v != 0xa5 {
			t.Fatal("output guard", i)
		}
	}
	before := bytes.Clone(frame)
	if game5AxisC(false, segment, -1, 1, 2, out) {
		t.Fatal("parallel line intersects")
	}
	if !bytes.Equal(frame, before) {
		t.Fatal("rejected intersection wrote output")
	}
	if n := game5IntersectionC(segment, rect, nil, 0); n != 0 {
		t.Fatal("zero-limit output access", n)
	}
	for _, limit := range []int32{0, -1} {
		if n := game5IntersectionC(nil, nil, nil, limit); n != 0 {
			t.Fatal("nonpositive limit must not access any buffer", limit, n)
		}
	}
}
