package legacy

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"image"
	"math"
	"os"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestCircleCollisionNativeWallIndexOriginal(t *testing.T) {
	blob, err := os.ReadFile("../server/testdata/circle_collision_startup_original.json")
	if err != nil {
		t.Fatal(err)
	}
	type state struct {
		New         [2]uint32 `json:"new"`
		Velocity    [2]uint32 `json:"velocity"`
		Accumulator [2]uint32 `json:"accumulator"`
	}
	type event struct {
		Service string   `json:"service"`
		Grid    []int    `json:"grid"`
		Normal  []uint32 `json:"normal"`
		First   state    `json:"first"`
	}
	var fixture struct {
		Tables []struct {
			Name string `json:"name"`
			Hex  string `json:"hex"`
		} `json:"tables"`
		Rows []struct {
			Mode  string `json:"mode"`
			Input struct {
				Name        string    `json:"name"`
				Kind        string    `json:"kind"`
				New         [2]uint32 `json:"first_new"`
				Position    [2]uint32 `json:"first_position"`
				Velocity    [2]uint32 `json:"first_velocity"`
				Accumulator [2]uint32 `json:"first_accumulator"`
				Mass        uint32    `json:"first_mass"`
				Radius      uint32    `json:"first_radius"`
				Type        uint16    `json:"first_type"`
				Spring      uint32    `json:"spring"`
				Collide     [4]uint32 `json:"collide"`
				WallType    byte      `json:"wall_type"`
			} `json:"input"`
			First     state   `json:"first"`
			Events    []event `json:"events"`
			Preserved bool    `json:"nonoutput_bytes_preserved"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	tables := make(map[string][]byte)
	for _, table := range fixture.Tables {
		data, err := hex.DecodeString(table.Hex)
		if err != nil {
			t.Fatal(err)
		}
		tables[table.Name] = data
	}
	defs := tables["wall-definitions"]
	if len(defs) != 264 || len(tables["wall-edge-masks"]) != 11 {
		t.Fatal("incomplete sealed wall tables")
	}
	for index := byte(0); index < 11; index++ {
		d := defs[int(index)*24:][:24]
		f := func(off int) float32 { return math.Float32frombits(binary.LittleEndian.Uint32(d[off:])) }
		want := server.CircleWallDefinition54FFC0{First: d[0], Second: d[12], FirstLo: f(4), FirstHi: f(8), SecondLo: f(16), SecondHi: f(20)}
		if got := circleWallDefinitionNative54FFC0(index); got != want {
			t.Fatalf("live C-owned wall definition %d differs: %+v original=%+v", index, got, want)
		}
	}
	rigidNativeTable(t, 292496, tables["wall-edge-masks"])
	s := &server.Server{}
	if s.Walls.Init() != 1 {
		t.Fatal("native wall allocation failed")
	}
	t.Cleanup(s.Walls.Free)
	wall := s.Walls.CreateAtGrid(image.Pt(15, 15))
	if wall == nil {
		t.Fatal("native wall insertion failed")
	}
	rigidNativeAddress(t, unsafe.Pointer(wall))
	selected := 0
	for _, row := range fixture.Rows {
		i := row.Input
		if row.Mode != "actual-startup" || i.Kind != "circle-wall" || !strings.HasPrefix(i.Name, "circle-wall/stock-") || !strings.HasSuffix(i.Name, "/velocity-0") || !(strings.Contains(i.Name, "/point-0/") || strings.Contains(i.Name, "/point-1/")) {
			continue
		}
		selected++
		t.Run(i.Name, func(t *testing.T) {
			wall.Dir0 = i.WallType
			obj, free := alloc.New(server.Object{})
			t.Cleanup(free)
			rigidNativeAddress(t, unsafe.Pointer(obj))
			*obj = server.Object{TypeInd: i.Type, NewPos: rigidNativePoint(i.New), PosVec: rigidNativePoint(i.Position), VelVec: rigidNativePoint(i.Velocity), Pos24: rigidNativePoint(i.Accumulator), Mass: math.Float32frombits(i.Mass), CollideP1: rigidNativePoint([2]uint32{i.Collide[0], i.Collide[1]}), CollideP2: rigidNativePoint([2]uint32{i.Collide[2], i.Collide[3]}), Field29: 0x13572468}
			obj.Shape.Kind = server.ShapeKindCircle
			obj.Shape.Circle.R = math.Float32frombits(i.Radius)
			want := rigidNativeExpected(obj, row.First.New, row.First.Velocity, row.First.Accumulator)
			r := circleWallRuntimeNative54FEF0()
			if math.Float32bits(r.Spring()) != i.Spring {
				t.Fatal("live C-owned spring differs from original input")
			}
			// Use the actual world methods with the actual C-owned definition
			// and coefficient providers. Only declared output services observe.
			r.WallType = s.Sub_57B500
			r.WallObject = s.Walls.GetWallAtGrid
			events := make([]event, 0)
			observe := func(e event) {
				e.First = state{rigidNativeBits(obj.NewPos), rigidNativeBits(obj.VelVec), rigidNativeBits(obj.Pos24)}
				events = append(events, e)
			}
			r.AddCollision = func(got *server.Object, normal types.Pointf) {
				if got != obj {
					t.Fatal("collision object changed")
				}
				w := rigidNativeBits(normal)
				observe(event{Service: "collision", Normal: []uint32{w[0], w[1]}})
			}
			r.OpenSecret = func(pos image.Point, got *server.Object) {
				if got != obj {
					t.Fatal("secret object changed")
				}
				observe(event{Service: "secret", Grid: []int{pos.X, pos.Y}})
			}
			server.CircleWallCollision54FEF0(obj, r)
			wantEvents := make([]event, 0)
			for _, e := range row.Events {
				if e.Service == "collision" || e.Service == "secret" {
					wantEvents = append(wantEvents, e)
				}
			}
			if !row.Preserved || !bytes.Equal(rigidNativeObjectBytes(obj), want) || !reflect.DeepEqual(events, wantEvents) {
				t.Fatalf("real native wall lookup differs: bytes=%t events=%+v original=%+v", bytes.Equal(rigidNativeObjectBytes(obj), want), events, wantEvents)
			}
		})
	}
	if selected != 22 {
		t.Fatalf("native indexed wall conditions=%d", selected)
	}
}
