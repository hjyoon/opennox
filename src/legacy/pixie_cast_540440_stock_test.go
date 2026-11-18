package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

type pixieCastInput540440 struct {
	Name, Mutation      string
	Warm, Type, Radius  uint32
	Owned, Level, Spell int32
	Desired             [2]float64
	Position            [2]uint32
	Dir                 [4]int
	Rays, Alloc         [4]json.RawMessage
	Target              bool
	Frame, TickRate     uint32
	Lifetime            int
}

type pixieCastObject540440 struct {
	Update, Owner, Target      int
	Direction, Spell, Deadline uint32
	LastVisible                uint32 `json:"last_visible"`
	Velocity, Origin           [2]uint32
}

type pixieCastResult540440 struct {
	Events [][]any
	Cache  uint32
	Pixies [4]pixieCastObject540440
	Result int
}

func TestCastPixiesFrozenOriginal540440(t *testing.T) {
	content, err := os.ReadFile("testdata/pixie_cast_original_v1.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	if hash := fmt.Sprintf("%x", sha256.Sum256(content)); hash != "ae714f5c28c9064eb9ad46688d5a0d1dbae0bc4a3b5c036e9bce9af8605cc46d" {
		t.Fatalf("stock Pixie corpus hash = %s", hash)
	}
	reader, err := gzip.NewReader(bytes.NewReader(content))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	var report struct {
		Startup       string         `json:"startup_control_word"`
		Guards        bool           `json:"object_and_register_guards"`
		OriginalSizes map[string]int `json:"original_sizes"`
		Entered       map[string]int
		Rows          []struct {
			Input    pixieCastInput540440
			Observed pixieCastResult540440
		}
	}
	if err := json.NewDecoder(reader).Decode(&report); err != nil {
		t.Fatal(err)
	}
	if report.Startup != "0xe3f" || !report.Guards ||
		report.OriginalSizes["pixie"] != 459 || report.OriginalSizes["converter"] != 39 ||
		report.Entered["pixie"] != 44 || report.Entered["converter"] != 71 || len(report.Rows) != 44 {
		t.Fatal("unexpected original Pixie proof")
	}
	for _, row := range report.Rows {
		t.Run(row.Input.Name, func(t *testing.T) { comparePixieCast540440(t, row.Input, row.Observed) })
	}
}

func comparePixieCast540440(t *testing.T, in pixieCastInput540440, want pixieCastResult540440) {
	t.Helper()
	bits := math.Float32bits
	f32 := math.Float32frombits
	caster := game4ConsumerValue(t, server.Object{PosVec: types.Ptf(f32(in.Position[0]), f32(in.Position[1]))})
	caster.Shape.Circle.R = f32(in.Radius)
	owner, target := game4ConsumerValue(t, server.Object{}), game4ConsumerValue(t, server.Object{})
	ignored := game4ConsumerValue(t, server.SpellAcceptArg{})
	guard := server.PixieUpdateData{Field8: 0xcafeb109, Field16: 0x98310765}
	alt := game4ConsumerValue(t, guard)
	var pixies [4]*server.Object
	var updates [4]*server.PixieUpdateData
	for i := range pixies {
		updates[i] = game4ConsumerValue(t, guard)
		pixies[i] = game4ConsumerValue(t, server.Object{
			UpdateData: unsafe.Pointer(updates[i]), VelVec: types.Ptf(1.125, -2.125),
		})
	}
	objects := append([]*server.Object{nil, caster, owner, target}, pixies[:]...)
	objectRole := func(o *server.Object) int {
		index := slices.Index(objects, o)
		if index < 0 {
			t.Fatalf("unexpected native object %p", o)
		}
		return index
	}
	updateRole := func(p *server.PixieUpdateData) int {
		if p == alt {
			return 12
		}
		index := slices.Index(updates[:], p)
		if index < 0 {
			t.Fatalf("unexpected native update record %p", p)
		}
		return 8 + index
	}
	enabled := func(value json.RawMessage) bool { return string(value) == "true" || string(value) == "1" }
	events := make([][]any, 0)
	cache, tick, frame := in.Warm, in.TickRate, in.Frame
	tableCalls, directions, allocations := 0, 0, 0
	runtime := server.PixieCastRuntime540440{
		PixieType: func() int {
			if cache == 0 {
				events = append(events, []any{"type", "Pixie"})
				cache = in.Type
				if in.Mutation == "lookup-caster-radius" {
					caster.Shape.Circle.R = 99
				}
			}
			return int(cache)
		},
		CountOwned: func(o *server.Object, typ int32) int32 {
			events = append(events, []any{"owned", objectRole(o), uint32(typ)})
			return in.Owned
		},
		Desired: func(level int) float64 {
			item := in.Desired[tableCalls]
			tableCalls++
			events = append(events, []any{"table", uint32(level), item})
			return item
		},
		RandomInt: func(min, max int) int {
			switch {
			case min == 0 && max == 255:
				value := in.Dir[directions]
				directions++
				events = append(events, []any{"direction", value})
				return value
			case min == 30 && max == 90:
				events = append(events, []any{"lifetime", in.Lifetime})
				if in.Mutation == "lifetime-clock" {
					tick, frame = 72, 17423
				}
				return in.Lifetime
			default:
				t.Fatalf("random range %d..%d", min, max)
				return 0
			}
		},
		Direction: func(direction uint8) types.Pointf {
			return map[uint8]types.Pointf{
				0: types.Ptf(0.6, 0.8), 64: types.Ptf(-0.8, 0.6),
				255: types.Ptf(0.70710677, -0.70710677), 127: types.Ptf(0.81, -0.59),
			}[direction]
		},
		TraceRay: func(from, to types.Pointf) bool {
			events = append(events, []any{"trace", bits(from.X), bits(from.Y), bits(to.X), bits(to.Y)})
			if in.Mutation == "trace-moves-caster" {
				caster.PosVec = types.Ptf(250.1234, 888.0123)
			}
			return enabled(in.Rays[directions-1])
		},
		NewPixie: func() *server.Object {
			var pixie *server.Object
			if enabled(in.Alloc[allocations]) {
				pixie = pixies[allocations]
			}
			allocations++
			events = append(events, []any{"new", objectRole(pixie)})
			return pixie
		},
		CreateAt: func(pixie, o *server.Object, pos types.Pointf) {
			events = append(events, []any{"create", objectRole(pixie), objectRole(o), bits(pos.X), bits(pos.Y), updateRole(pixie.UpdateDataPixie())})
			if in.Mutation == "create-replaces-update" {
				pixie.UpdateData = unsafe.Pointer(alt)
			}
		},
		FindTarget: func(pixie, o *server.Object) *server.Object {
			if pixie.VelVec != (types.Pointf{}) {
				t.Fatal("target search ran before clearing velocity")
			}
			angle := uint32(pixie.Direction1) | uint32(pixie.Direction2)<<16
			events = append(events, []any{"target", objectRole(pixie), objectRole(o), angle, updateRole(pixie.UpdateDataPixie())})
			if in.Target {
				return target
			}
			return nil
		},
		Frame:    func() uint32 { return frame },
		TickRate: func() uint32 { return tick },
		PlayCastAudio: func(id spell.ID, o *server.Object) {
			events = append(events, []any{"sound-id", uint32(id)}, []any{"sound", 777, objectRole(o)})
		},
	}
	oldRuntime, oldGetServer := pixieCastRuntime540440, GetServer
	t.Cleanup(func() { pixieCastRuntime540440, GetServer = oldRuntime, oldGetServer })
	pixieCastRuntime540440 = func(Server) server.PixieCastRuntime540440 { return runtime }
	GetServer = func() Server { return nil }
	result := pixieCastStockFixture540440(in.Spell, target, owner, caster, ignored, in.Level)
	gotEvents, err := json.Marshal(events)
	if err != nil {
		t.Fatal(err)
	}
	wantEvents, err := json.Marshal(want.Events)
	if err != nil {
		t.Fatal(err)
	}
	if result != want.Result || cache != want.Cache || !bytes.Equal(gotEvents, wantEvents) {
		t.Fatalf("result %d cache %x events %s; original %d %x %s", result, cache, gotEvents, want.Result, want.Cache, wantEvents)
	}
	if *alt != guard {
		t.Fatal("Pixie cast overwrote the replacement update record")
	}
	for i, pixie := range pixies {
		u := updates[i]
		got := pixieCastObject540440{
			Update: updateRole(pixie.UpdateDataPixie()), Owner: objectRole(u.Owner), Target: objectRole(u.Target),
			Direction: uint32(pixie.Direction1) | uint32(pixie.Direction2)<<16, Spell: uint32(u.SpellID),
			Deadline: u.Deadline, LastVisible: u.LastOwnerVisibleFrame,
			Velocity: [2]uint32{bits(pixie.VelVec.X), bits(pixie.VelVec.Y)},
			Origin:   [2]uint32{bits(pixie.Pos39.X), bits(pixie.Pos39.Y)},
		}
		if got != want.Pixies[i] || u.Field8 != guard.Field8 || u.Field16 != guard.Field16 {
			t.Fatalf("Pixie %d %+v; original %+v", i, got, want.Pixies[i])
		}
	}
}
