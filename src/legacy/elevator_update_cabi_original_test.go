package legacy

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestElevatorUpdateCEntryOriginal(t *testing.T) {
	blob, err := os.ReadFile("../server/testdata/elevator_passenger_startup_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Rows []struct {
			Mode  string `json:"mode"`
			Input struct {
				Kind            string           `json:"kind"`
				Name            string           `json:"name"`
				Source          [2]uint32        `json:"source"`
				Target          [2]uint32        `json:"target"`
				Corners         [8]uint32        `json:"corners"`
				Size            [2]uint32        `json:"size"`
				CandidateKind   server.ShapeKind `json:"candidate_kind"`
				CandidateSize   [2]uint32        `json:"candidate_size"`
				Radius          uint32           `json:"radius"`
				Height          int64            `json:"height"`
				LinkedHeight    int64            `json:"linked_height"`
				State           byte             `json:"state"`
				LinkedState     byte             `json:"linked_state"`
				AlternateState  byte             `json:"alternate_state"`
				AlternateHeight int64            `json:"alternate_height"`
				ShaftState      byte             `json:"shaft_state"`
				Z               uint32           `json:"z"`
				Flags           uint32           `json:"flags"`
				Class           uint32           `json:"cls"`
				Linked          bool             `json:"linked"`
				Last            uint32           `json:"last"`
				TickRate        uint32           `json:"tick_rate"`
				Ticks           []uint32         `json:"ticks"`
				Material        uint16           `json:"material"`
				Subclass        uint32           `json:"subclass"`
				Direction       int32            `json:"direction"`
				MutationEvent   string           `json:"mutation_event"`
			} `json:"input"`
			Events []struct {
				Service  string    `json:"service"`
				Object   int       `json:"object"`
				Position [2]uint32 `json:"position"`
			} `json:"events"`
			Objects       [][]uint32 `json:"objects"`
			ObjectUpdates []int      `json:"object_updates"`
			Updates       [][5]int64 `json:"updates"`
			Return        uint8      `json:"return_byte"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	if !srv.Objs.Init(16) {
		t.Fatal("native object class initialization")
	}
	t.Cleanup(srv.Objs.FreeObjects)
	srv.Map.Init()
	t.Cleanup(srv.Map.Free)
	oldServer, oldMove := GetServer, Nox_xxx_unitMove_4E7010
	GetServer = func() Server { return &elevatorCollisionCABIService551AE0{srv: srv} }
	t.Cleanup(func() { GetServer = oldServer; Nox_xxx_unitMove_4E7010 = oldMove })
	count := 0
	for _, row := range fixture.Rows {
		i := row.Input
		if row.Mode != "actual-startup" || i.MutationEvent != "" || len(i.Ticks) != 1 {
			continue
		}
		moving := false
		for _, e := range row.Events {
			if e.Service == "move" || e.Service == "raise" {
				moving = true
			}
		}
		if (i.Kind == "elevator" || i.Kind == "shaft") && moving {
			continue
		}
		count++
		t.Run(i.Name, func(t *testing.T) {
			var objs [5]*server.Object
			var ud [4]*server.ElevatorUpdateData
			for n := range objs {
				p := srv.Objs.NewObject(&server.ObjectType{})
				if p == nil {
					t.Fatal("allocate native server object")
				}
				objs[n] = p
				rigidNativeAddress(t, unsafe.Pointer(p))
				p.TypeInd = 40
				p.Field37 = 0x12345678
				p.Field38 = 0x11223344
				p.Field29 = 0x87654321
				for k := range p.Field140 {
					p.Field140[k] = 0xa5a50000 + uint32(257*k)
				}
				pos := i.Target
				if n == 0 {
					pos = i.Source
				} else if n >= 3 {
					pos = [2]uint32{math.Float32bits(float32(900 + 300*(n-3))), math.Float32bits(float32(901 + 300*(n-3)))}
				}
				p.PosVec = rigidNativePoint(pos)
				p.NewPos = p.PosVec
				p.PrevPos = p.PosVec
				p.Shape.Kind = server.ShapeKindBox
				p.Shape.Box.W, p.Shape.Box.H = math.Float32frombits(i.Size[0]), math.Float32frombits(i.Size[1])
				fields := []*float32{&p.Shape.Box.LeftTop, &p.Shape.Box.LeftBottom, &p.Shape.Box.LeftBottom2, &p.Shape.Box.LeftTop2, &p.Shape.Box.RightTop, &p.Shape.Box.RightBottom, &p.Shape.Box.RightBottom2, &p.Shape.Box.RightTop2}
				for k, f := range fields {
					*f = math.Float32frombits(i.Corners[k])
				}
				if n == 1 || n == 2 {
					p.Shape.Kind = i.CandidateKind
					p.Shape.Box.W, p.Shape.Box.H = math.Float32frombits(i.CandidateSize[0]), math.Float32frombits(i.CandidateSize[1])
					p.Shape.Circle.R = math.Float32frombits(i.Radius)
					p.ObjClass = object.Class(i.Class)
					p.ObjFlags = 0x100040
					p.ZVal = math.Float32frombits(i.Z)
					p.Field27 = 3.5
				}
			}
			owner := objs[0]
			owner.ObjFlags = object.Flags(i.Flags)
			owner.Field34 = i.Last
			owner.Material = i.Material
			owner.ObjSubClass = object.SubClass(i.Subclass)
			rawUD := func(p *server.ElevatorUpdateData) []byte {
				return unsafe.Slice((*byte)(unsafe.Pointer(p)), int(unsafe.Sizeof(*p)))
			}
			for n := range ud {
				p, free := alloc.New(server.ElevatorUpdateData{})
				t.Cleanup(free)
				ud[n] = p
				rigidNativeAddress(t, unsafe.Pointer(p))
				for k := range rawUD(p) {
					rawUD(p)[k] = 0xa5
				}
				p.Field_0 = 0x10203040 + uint32(n)
				p.Field_1 = 0
				p.Field_2 = 0x55667700 + uint32(n)
				p.Field_3 = [4]byte{i.State, i.AlternateState, i.LinkedState, i.AlternateState}[n]
				if n == 0 && i.Kind == "shaft" {
					p.Field_3 = i.ShaftState
				}
				p.Field_4 = uint32([4]int64{i.Height, i.AlternateHeight, i.LinkedHeight, i.AlternateHeight}[n])
			}
			owner.UpdateData = unsafe.Pointer(ud[0])
			objs[3].UpdateData = unsafe.Pointer(ud[2])
			objs[4].UpdateData = unsafe.Pointer(ud[3])
			link := objs[3]
			if !i.Linked {
				link = nil
			}
			owner.SetElevatorLinkFor(unsafe.Pointer(ud[0]), link)
			objs[3].SetElevatorLinkFor(unsafe.Pointer(ud[2]), owner)
			objs[4].SetElevatorLinkFor(unsafe.Pointer(ud[3]), owner)
			t.Cleanup(func() {
				for _, p := range objs {
					srv.Objs.FreeObject(p)
				}
			})
			offsets := []uintptr{unsafe.Offsetof(owner.ObjClass), unsafe.Offsetof(owner.ObjSubClass), unsafe.Offsetof(owner.ObjFlags), unsafe.Offsetof(owner.Material), unsafe.Offsetof(owner.PosVec), unsafe.Offsetof(owner.PosVec) + 4, unsafe.Offsetof(owner.NewPos), unsafe.Offsetof(owner.NewPos) + 4, unsafe.Offsetof(owner.PrevPos), unsafe.Offsetof(owner.PrevPos) + 4, unsafe.Offsetof(owner.ZVal), unsafe.Offsetof(owner.Field27), unsafe.Offsetof(owner.Field34), unsafe.Offsetof(owner.Field37), unsafe.Offsetof(owner.Field38)}
			for k := 0; k < 32; k++ {
				offsets = append(offsets, unsafe.Offsetof(owner.Field140)+uintptr(4*k))
			}
			wantObjs := make([][]byte, 5)
			wantUD := make([][]byte, 4)
			for n, p := range objs {
				want := bytes.Clone(rigidNativeObjectBytes(p))
				wantObjs[n] = want
				for k, off := range offsets {
					binary.LittleEndian.PutUint32(want[off:], row.Objects[n][k])
				}
			}
			for n, p := range ud {
				want := bytes.Clone(rawUD(p))
				wantUD[n] = want
				for k, v := range row.Updates[n] {
					if k != 1 {
						binary.LittleEndian.PutUint32(want[k*4:], uint32(v))
					}
				}
			}
			type move struct {
				Object   int
				Position [2]uint32
			}
			gotMoves := make([]move, 0)
			wantMoves := make([]move, 0)
			for _, e := range row.Events {
				if e.Service == "move" {
					wantMoves = append(wantMoves, move{e.Object, e.Position})
				}
			}
			Nox_xxx_unitMove_4E7010 = func(p *server.Object, pos types.Pointf) {
				index := -1
				for n, v := range objs {
					if p == v {
						index = n
					}
				}
				gotMoves = append(gotMoves, move{index, [2]uint32{math.Float32bits(pos.X), math.Float32bits(pos.Y)}})
				p.PosVec = pos
				p.NewPos = pos
				p.PrevPos = pos
			}
			srv.SetFrame(i.Ticks[0])
			srv.SetTickRate(i.TickRate)
			switch i.Kind {
			case "elevator":
				elevatorUpdateCABI53B5D0(owner)
			case "shaft":
				if got := elevatorShaftUpdateCABI53B380(owner); got != row.Return {
					t.Fatalf("shaft return=%d original=%d", got, row.Return)
				}
			case "up":
				elevatorPassengerUpCABI53B750(objs[1], owner)
			case "down":
				elevatorPassengerDownCABI53B410(objs[1], owner)
			case "audio":
				elevatorAudioCABI53B490(owner, i.Direction)
			default:
				t.Fatal("unknown C entry")
			}
			if !reflect.DeepEqual(gotMoves, wantMoves) {
				t.Fatalf("native movement=%+v original=%+v", gotMoves, wantMoves)
			}
			for n, p := range objs {
				if !bytes.Equal(rigidNativeObjectBytes(p), wantObjs[n]) {
					t.Fatalf("C entry object %d bytes differ", n)
				}
			}
			for n, p := range ud {
				if !bytes.Equal(rawUD(p), wantUD[n]) {
					t.Fatalf("C entry update %d bytes differ", n)
				}
			}
			if owner.ElevatorLinkFor(unsafe.Pointer(ud[0])) != link {
				t.Fatal("native C link changed")
			}
		})
	}
	if count != 1143 {
		t.Fatalf("C entry conditions=%d", count)
	}
}
