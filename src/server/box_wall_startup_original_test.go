package server

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

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type boxWallOriginalEvent5504B0 struct {
	Service     string   `json:"service"`
	Grid        []int32  `json:"grid"`
	Flags       uint32   `json:"flags"`
	Impulse     []uint32 `json:"impulse_bits32"`
	New         []uint32 `json:"new_bits32"`
	Velocity    []uint32 `json:"velocity_bits32"`
	Accumulator []uint32 `json:"accumulator_bits32"`
}

func TestBoxWall5504B0StartupOriginal(t *testing.T) {
	testBoxWall5504B0Original(t, false)
}

func TestBoxWall5504B0NativeWallIndex(t *testing.T) {
	testBoxWall5504B0Original(t, true)
}

func testBoxWall5504B0Original(t *testing.T, native bool) {
	blob, err := os.ReadFile("testdata/box_wall_startup_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		SHA         string            `json:"original_sha256"`
		Unchanged   bool              `json:"original_unchanged"`
		CW          uint16            `json:"actual_startup_control_word"`
		Contexts    int               `json:"contexts"`
		Faults      []json.RawMessage `json:"memory_faults"`
		Definitions string            `json:"original_wall_definitions_hex"`
		Rows        []struct {
			Mode  string `json:"mode"`
			Input struct {
				Name        string    `json:"name"`
				Entry       string    `json:"entry"`
				Class       uint32    `json:"class_bits"`
				Collide     [4]uint32 `json:"collide_bits32"`
				New         [2]uint32 `json:"new_bits32"`
				Previous    [2]uint32 `json:"previous_bits32"`
				Velocity    [2]uint32 `json:"velocity_bits32"`
				Size        [2]uint32 `json:"size_bits32"`
				Accumulator [2]uint32 `json:"accumulator_bits32"`
				Mass        uint32    `json:"mass_bits32"`
				Impulse     [2]uint32 `json:"impulse_bits32"`
				WallType    byte      `json:"wall_type"`
				Grid        [2]int    `json:"wall_grid"`
				AllWalls    bool      `json:"all_walls"`
				Mutation    int       `json:"mutation"`
			} `json:"input"`
			Events      []boxWallOriginalEvent5504B0 `json:"events"`
			New         [2]uint32                    `json:"new_bits32"`
			Velocity    [2]uint32                    `json:"velocity_bits32"`
			Accumulator [2]uint32                    `json:"accumulator_bits32"`
			Class       uint32                       `json:"class_bits"`
			Mass        uint32                       `json:"mass_bits32"`
			Size        [2]uint32                    `json:"size_bits32"`
			Previous    [2]uint32                    `json:"previous_bits32"`
			Definitions string                       `json:"definitions_hex"`
			Preserved   bool                         `json:"nonoutput_bytes_preserved"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || !fixture.Unchanged || fixture.CW != 0xe3f || fixture.Contexts != 591 || len(fixture.Rows) != 1182 || len(fixture.Faults) != 0 {
		t.Fatal("incomplete unchanged original box/wall observations")
	}
	if native {
		// Standalone server tests initialize the sealed stock definitions in the
		// real memory table, then restore every byte. World lookup stays real.
		const base, offset, need = 0x587000, 292520, 292520 + 11*24
		if blob := memmap.BlobByAddr(base); blob == nil {
			memmap.RegisterBlobData(base, "box_wall_original_test", make([]byte, need))
		} else if len(blob.Data) < need {
			previous := *blob
			data := make([]byte, need)
			copy(data, blob.Data)
			blob.Data, blob.Size = data, uintptr(len(data))
			t.Cleanup(func() { *blob = previous })
		}
		table := memmap.Slice(base, offset)[:11*24]
		previous := bytes.Clone(table)
		t.Cleanup(func() { copy(table, previous) })
		defs, err := hex.DecodeString(fixture.Definitions)
		if err != nil {
			t.Fatal(err)
		}
		copy(table, defs)
	}
	point := func(w [2]uint32) types.Pointf {
		return types.Ptf(math.Float32frombits(w[0]), math.Float32frombits(w[1]))
	}
	bits := func(p types.Pointf) []uint32 { return []uint32{math.Float32bits(p.X), math.Float32bits(p.Y)} }
	selected := 0
	for _, row := range fixture.Rows {
		if row.Mode != "actual-startup" {
			continue
		}
		if native && (!strings.HasPrefix(row.Input.Name, "stock-") || !(strings.Contains(row.Input.Name, "/point-0/") || strings.Contains(row.Input.Name, "/point-1/"))) {
			continue
		}
		selected++
		t.Run(row.Input.Name, func(t *testing.T) {
			i := row.Input
			obj, free := alloc.New(Object{})
			t.Cleanup(free)
			if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(obj)) <= math.MaxUint32 {
				t.Fatal("native box below 4 GiB")
			}
			*obj = Object{ObjClass: object.Class(i.Class), NewPos: point(i.New), PrevPos: point(i.Previous), VelVec: point(i.Velocity), Pos24: point(i.Accumulator), Mass: math.Float32frombits(i.Mass), CollideP1: point([2]uint32{i.Collide[0], i.Collide[1]}), CollideP2: point([2]uint32{i.Collide[2], i.Collide[3]}), Field29: 0x13572468}
			obj.Shape.Kind = ShapeKindBox
			obj.Shape.Box.W, obj.Shape.Box.H = math.Float32frombits(i.Size[0]), math.Float32frombits(i.Size[1])
			raw := func() []byte { return unsafe.Slice((*byte)(unsafe.Pointer(obj)), int(unsafe.Sizeof(*obj))) }
			expected := bytes.Clone(raw())
			write32 := func(p unsafe.Pointer, w uint32) {
				offset := uintptr(p) - uintptr(unsafe.Pointer(obj))
				binary.LittleEndian.PutUint32(expected[offset:], w)
			}
			writePoint := func(p *types.Pointf, w [2]uint32) {
				write32(unsafe.Pointer(&p.X), w[0])
				write32(unsafe.Pointer(&p.Y), w[1])
			}
			writePoint(&obj.NewPos, row.New)
			writePoint(&obj.VelVec, row.Velocity)
			writePoint(&obj.Pos24, row.Accumulator)
			writePoint(&obj.PrevPos, row.Previous)
			write32(unsafe.Pointer(&obj.ObjClass), row.Class)
			write32(unsafe.Pointer(&obj.Mass), row.Mass)
			write32(unsafe.Pointer(&obj.Shape.Box.W), row.Size[0])
			write32(unsafe.Pointer(&obj.Shape.Box.H), row.Size[1])
			defs, err := hex.DecodeString(fixture.Definitions)
			if err != nil || len(defs) != 11*24 {
				t.Fatal("invalid original wall definitions")
			}
			events := make([]boxWallOriginalEvent5504B0, 0)
			state := func(service string) boxWallOriginalEvent5504B0 {
				return boxWallOriginalEvent5504B0{Service: service, New: bits(obj.NewPos), Velocity: bits(obj.VelVec), Accumulator: bits(obj.Pos24)}
			}
			collisions := 0
			h := boxWallCollisionHooks5504B0{
				spring: func() float32 { return boxWallSpring5504B0 },
				wallType: func(grid image.Point, flags byte) int8 {
					events = append(events, boxWallOriginalEvent5504B0{Service: "wall", Grid: []int32{int32(grid.X), int32(grid.Y)}, Flags: uint32(flags)})
					if i.Mutation == 3 {
						obj.NewPos = types.Ptf(354.25, 348.5)
						obj.PrevPos = types.Ptf(340.5, 329.75)
					}
					if i.AllWalls || grid == image.Pt(i.Grid[0], i.Grid[1]) {
						return int8(i.WallType)
					}
					return -1
				},
				wallDef: func(typ byte) boxWallDef5504B0 {
					d := defs[int(typ)*24:][:24]
					f := func(off int) float32 { return math.Float32frombits(binary.LittleEndian.Uint32(d[off:])) }
					return boxWallDef5504B0{firstEnabled: d[0], firstMask: d[1], firstOffset: f(4), firstLength: f(8), secondEnabled: d[12], secondMask: d[13], secondOffset: f(16), secondLength: f(20)}
				},
				addCollision: func(got *Object, impulse types.Pointf) {
					if got != obj {
						t.Fatal("collision object changed")
					}
					collisions++
					e := state("collision")
					e.Impulse = bits(impulse)
					events = append(events, e)
					if collisions == 1 && i.Mutation == 1 {
						obj.Mass = 5.5
						obj.VelVec = types.Ptf(2.5, -3.25)
						obj.NewPos = types.Ptf(371.25, 359.5)
						obj.Shape.Box.W = 24
						obj.Shape.Box.H = 14
					}
					if collisions == 1 && i.Mutation == 2 {
						defs[int(i.WallType)*24+12] = 0
					}
				},
				openSecret: func(grid image.Point, got *Object) {
					if got != obj {
						t.Fatal("secret-contact object changed")
					}
					e := state("secret")
					e.Grid = []int32{int32(grid.X), int32(grid.Y)}
					events = append(events, e)
					if i.Mutation == 4 {
						obj.ObjClass |= 0x400000
					}
				},
			}
			wantEvents := row.Events
			if native {
				s := newDirectedVisionMap(t)
				w, freeWall := alloc.New(Wall{})
				t.Cleanup(freeWall)
				*w = Wall{X5: 15, Y6: 15, Dir0: i.WallType}
				if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(w)) <= math.MaxUint32 {
					t.Fatal("native wall below 4 GiB")
				}
				ind, ok := wallArrayInd(w.GridPos())
				if !ok {
					t.Fatal("invalid wall grid")
				}
				s.Walls.byPos[ind] = w
				if w.GridPos() != image.Pt(i.Grid[0], i.Grid[1]) || s.Sub_57B500(w.GridPos(), 0x40) != int8(i.WallType) {
					t.Fatal("real wall lookup did not resolve stock definition")
				}
				if !bytes.Equal(memmap.Slice(0x587000, 292520)[:11*24], defs) {
					t.Fatal("live stock definition table changed during fixture setup")
				}
				t.Logf("real indexed wall=%v type=%d first=%d len=%08x second=%d len=%08x bounds=%v..%v grid=%d..%d", w.GridPos(), w.Dir0, memmap.Uint8(0x587000, 292520+uintptr(i.WallType)*24), math.Float32bits(memmap.Float32(0x587000, 292528+uintptr(i.WallType)*24)), memmap.Uint8(0x587000, 292532+uintptr(i.WallType)*24), math.Float32bits(memmap.Float32(0x587000, 292540+uintptr(i.WallType)*24)), obj.CollideP1, obj.CollideP2, boxWallGrid5504B0(obj.CollideP1.X), boxWallGrid5504B0(obj.CollideP2.X))
				wantEvents = make([]boxWallOriginalEvent5504B0, 0)
				for _, event := range row.Events {
					if event.Service != "wall" {
						wantEvents = append(wantEvents, event)
					}
				}
				s.BoxWallCollision5504B0(obj, h.spring, h.addCollision, h.openSecret)
			} else if i.Entry == "impulse" {
				obj.Sub548600(point(i.Impulse))
			} else {
				boxWallCollision5504B0(obj, h)
			}
			if !row.Preserved || !bytes.Equal(raw(), expected) || !reflect.DeepEqual(events, wantEvents) || hex.EncodeToString(defs) != row.Definitions {
				t.Fatalf("native new/velocity/accumulator=%08x/%08x/%08x original=%08x/%08x/%08x; bytes preserved=%t; events=%+v original=%+v; definitions preserved=%t", bits(obj.NewPos), bits(obj.VelVec), bits(obj.Pos24), row.New, row.Velocity, row.Accumulator, bytes.Equal(raw(), expected), events, row.Events, hex.EncodeToString(defs) == row.Definitions)
			}
		})
	}
	want := 591
	if native {
		want = 66
	}
	if selected != want {
		t.Fatalf("startup conditions=%d", selected)
	}
}
