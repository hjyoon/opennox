package legacy

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"math"
	"os"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type elevatorCollisionCABIService551AE0 struct {
	Server
	srv *server.Server
}

func (s *elevatorCollisionCABIService551AE0) S() *server.Server { return s.srv }

func TestElevatorCollisionCEntryOriginal(t *testing.T) {
	blob, err := os.ReadFile("../server/testdata/elevator_shaft_startup_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Rows []struct {
			Mode  string `json:"mode"`
			Input struct {
				Kind      string    `json:"kind"`
				Name      string    `json:"name"`
				Source    [2]uint32 `json:"source"`
				Target    [2]uint32 `json:"target"`
				Size      [2]uint32 `json:"size"`
				Corners   [8]uint32 `json:"corners"`
				Height    int32     `json:"height"`
				Z         uint32    `json:"z"`
				Vertical  uint32    `json:"vertical"`
				Flags     uint32    `json:"flags"`
				Class     uint32    `json:"candidate_class"`
				Type      uint16    `json:"candidate_type"`
				Move      int32     `json:"move"`
				CachedIDs [4]uint32 `json:"cached_ids"`
				CABI      bool      `json:"cabi"`
			} `json:"input"`
			Candidate struct {
				Z         uint32     `json:"z"`
				Vertical  uint32     `json:"vertical"`
				Flags     uint32     `json:"flags"`
				SyncMask  uint32     `json:"sync_mask"`
				SyncWords [32]uint32 `json:"sync_words"`
			} `json:"candidate"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	oldServer := GetServer
	srv := &server.Server{}
	GetServer = func() Server { return &elevatorCollisionCABIService551AE0{srv: srv} }
	t.Cleanup(func() { GetServer = oldServer })
	slots := memmap.Slice(0x5D4594, 2491792)[:20]
	oldSlots := bytes.Clone(slots)
	t.Cleanup(func() { copy(slots, oldSlots) })
	selected := 0
	for _, row := range fixture.Rows {
		i := row.Input
		if row.Mode != "actual-startup" || i.Kind != "elevator" || !(i.CABI || (strings.Contains(i.Name, "/type-90") && strings.HasSuffix(i.Name, "/initialized-1"))) {
			continue
		}
		selected++
		t.Run(i.Name, func(t *testing.T) {
			a, fa := alloc.New(server.Object{})
			b, fb := alloc.New(server.Object{})
			u, fu := alloc.New(server.ElevatorUpdateData{})
			for _, free := range []func(){fa, fb, fu} {
				t.Cleanup(free)
			}
			for _, ptr := range []unsafe.Pointer{unsafe.Pointer(a), unsafe.Pointer(b), unsafe.Pointer(u)} {
				rigidNativeAddress(t, ptr)
			}
			*a = server.Object{NewPos: rigidNativePoint(i.Source), UpdateData: unsafe.Pointer(u), Field29: 0x13572468}
			*u = server.ElevatorUpdateData{Field_4: uint32(i.Height)}
			a.Shape.Kind = server.ShapeKindBox
			a.Shape.Box.W, a.Shape.Box.H = math.Float32frombits(i.Size[0]), math.Float32frombits(i.Size[1])
			fields := []*float32{&a.Shape.Box.LeftTop, &a.Shape.Box.LeftBottom, &a.Shape.Box.LeftBottom2, &a.Shape.Box.LeftTop2, &a.Shape.Box.RightTop, &a.Shape.Box.RightBottom, &a.Shape.Box.RightBottom2, &a.Shape.Box.RightTop2}
			for index, p := range fields {
				*p = math.Float32frombits(i.Corners[index])
			}
			*b = server.Object{TypeInd: i.Type, NewPos: rigidNativePoint(i.Target), ZVal: math.Float32frombits(i.Z), Field27: math.Float32frombits(i.Vertical), ObjClass: object.Class(i.Class), ObjFlags: object.Flags(i.Flags), Field37: 0x12345678, Field38: 0x11223344, Field29: 0x87654321}
			b.Shape.Kind = server.ShapeKindCircle
			b.Shape.Circle.R = 5
			for index := range b.Field140 {
				b.Field140[index] = 0xa5a50000 + uint32(257*index)
			}
			beforeA := bytes.Clone(rigidNativeObjectBytes(a))
			beforeU := bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(u)), int(unsafe.Sizeof(*u))))
			wantB := bytes.Clone(rigidNativeObjectBytes(b))
			write := func(off uintptr, value uint32) { binary.LittleEndian.PutUint32(wantB[off:], value) }
			write(unsafe.Offsetof(b.ZVal), row.Candidate.Z)
			write(unsafe.Offsetof(b.Field27), row.Candidate.Vertical)
			write(unsafe.Offsetof(b.ObjFlags), row.Candidate.Flags)
			write(unsafe.Offsetof(b.Field38), row.Candidate.SyncMask)
			for index, value := range row.Candidate.SyncWords {
				write(unsafe.Offsetof(b.Field140)+uintptr(index*4), value)
			}
			for index, id := range i.CachedIDs {
				binary.LittleEndian.PutUint32(slots[index*4:], id)
			}
			binary.LittleEndian.PutUint32(slots[16:], 1)
			for index, name := range []string{"SmallFist", "MediumFist", "LargeFist", "Meteor"} {
				if srv.Types.IndByID(name) != 0 || uint32(elevatorCollisionTypeNative551BF0(name)) != i.CachedIDs[index] {
					t.Fatal("live cached IDs were replaced by fresh type lookups")
				}
			}
			beforeSlots := bytes.Clone(slots)
			elevatorCollisionCABI551AE0(a, b, i.Move)
			if !bytes.Equal(rigidNativeObjectBytes(a), beforeA) || !bytes.Equal(rigidNativeObjectBytes(b), wantB) || !bytes.Equal(unsafe.Slice((*byte)(unsafe.Pointer(u)), int(unsafe.Sizeof(*u))), beforeU) || !bytes.Equal(slots, beforeSlots) {
				t.Fatalf("native C entry/real raise differs: height=%08x original=%08x flags=%08x original=%08x", math.Float32bits(b.ZVal), row.Candidate.Z, uint32(b.ObjFlags), row.Candidate.Flags)
			}
		})
	}
	if selected != 16 {
		t.Fatalf("lift C entry conditions=%d", selected)
	}
}
