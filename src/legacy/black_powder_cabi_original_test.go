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

type powderCEventOriginal struct {
	Service string   `json:"service"`
	Args    []uint32 `json:"args"`
}
type powderCService53C9A0 struct {
	Server
	srv    *server.Server
	events []powderCEventOriginal
	source *server.Object
}

func (s *powderCService53C9A0) S() *server.Server { return s.srv }
func (s *powderCService53C9A0) DelayedDelete(p *server.Object) {
	if p != s.source {
		panic("foreign delayed-delete object")
	}
	s.events = append(s.events, powderCEventOriginal{Service: "delete", Args: []uint32{0}})
}
func (s *powderCService53C9A0) Nox_xxx_mapDamageUnitsAround(p types.Pointf, outer, inner float32, damage int, kind object.DamageType, source *server.Object, excluded server.Obj, walls bool) {
	if source != s.source || excluded != nil || walls {
		panic("foreign damage identity or wall mode")
	}
	s.events = append(s.events, powderCEventOriginal{Service: "damage", Args: []uint32{math.Float32bits(p.X), math.Float32bits(p.Y), math.Float32bits(outer), math.Float32bits(inner), uint32(damage), uint32(kind), 0, math.MaxUint32}})
}

func TestBlackPowderCEntryOriginal(t *testing.T) {
	blob, err := os.ReadFile("../server/testdata/black_powder_startup_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		SHA        string `json:"original_sha256"`
		Contexts   int    `json:"contexts"`
		Executions int    `json:"root_executions"`
		Rows       []struct {
			Mode  string `json:"mode"`
			Input struct {
				Kind          string    `json:"kind"`
				Name          string    `json:"name"`
				Creation      uint32    `json:"creation"`
				Fuse          uint32    `json:"fuse"`
				Frame         uint32    `json:"frame"`
				FPS           uint32    `json:"fps"`
				Position      [2]uint32 `json:"position"`
				RNG           int       `json:"rng"`
				MutationEvent string    `json:"mutation_event"`
			} `json:"input"`
			Objects   [][6]uint32            `json:"objects"`
			RNG       int                    `json:"rng"`
			Events    []powderCEventOriginal `json:"events"`
			Preserved bool                   `json:"nonoutput_bytes_preserved"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &f); err != nil {
		t.Fatal(err)
	}
	if f.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || f.Contexts != 471 || f.Executions != 942 || len(f.Rows) != 942 {
		t.Fatal("incomplete original fixture")
	}
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	shim := &powderCService53C9A0{srv: srv}
	oldServer, oldWalls := GetServer, GetDoDamageWalls
	GetServer = func() Server { return shim }
	GetDoDamageWalls = func() bool { return false }
	t.Cleanup(func() { GetServer = oldServer; GetDoDamageWalls = oldWalls })
	counts := map[string]int{}
	for _, row := range f.Rows {
		i := row.Input
		if row.Mode != "actual-startup" || i.MutationEvent != "" || (i.Kind == "barrel" && i.Fuse != i.Creation && i.Fuse == i.Frame) {
			continue
		}
		// Actual production C/Go bindings cover timing and the burn damage call.
		// Flame allocation, tracing and lifecycle are explicit services in the
		// separate full server matrix and are not claimed by this C subset.
		counts[i.Kind]++
		t.Run(i.Name, func(t *testing.T) {
			p, free := alloc.Make([]byte(nil), int(unsafe.Sizeof(server.Object{}))+64)
			t.Cleanup(free)
			for n := range p {
				p[n] = 0xa5
			}
			source := (*server.Object)(unsafe.Pointer(&p[32]))
			if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(source)) <= math.MaxUint32 {
				t.Fatal("native C object below 4 GiB")
			}
			*source = server.Object{PosVec: rigidNativePoint(i.Position), Field32: i.Creation, Field34: i.Fuse, Field38: 0x11223344, ObjFlags: 0x40}
			want := bytes.Clone(p)
			binary.LittleEndian.PutUint32(want[32+unsafe.Offsetof(source.Field34):], row.Objects[0][3])
			srv.SetFrame(i.Frame)
			srv.SetTickRate(i.FPS)
			srv.Rand.Logic.Reset(i.RNG)
			shim.source = source
			shim.events = make([]powderCEventOriginal, 0)
			if i.Kind == "barrel" {
				blackPowderBarrelC53C9A0(source)
			} else if i.Kind == "burn" {
				blackPowderBurnC53CCB0(source)
			} else {
				t.Fatal("unknown C entry")
			}
			events := make([]powderCEventOriginal, 0)
			for _, e := range row.Events {
				if e.Service == "damage" || e.Service == "delete" {
					events = append(events, e)
				}
			}
			if !row.Preserved || !bytes.Equal(p, want) || srv.Rand.Logic.Index() != row.RNG || !reflect.DeepEqual(shim.events, events) {
				t.Fatalf("source frame=%t RNG=%d/%d effects=%+v original=%+v", bytes.Equal(p, want), srv.Rand.Logic.Index(), row.RNG, shim.events, events)
			}
		})
	}
	if counts["barrel"] != 38 || counts["burn"] != 39 || len(counts) != 2 {
		t.Fatalf("C coverage=%v", counts)
	}
}
