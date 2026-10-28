package server

import (
	"encoding/json"
	"image"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/sound"
)

type lifetimeOriginalState53CC90 struct {
	Objects [4][12]uint32 `json:"objects"`
	Owners  [4]int        `json:"owners"`
	Updates [4]int        `json:"update_refs"`
	Players [2]int        `json:"player_refs"`
	Cursors [2][2]uint32  `json:"cursors"`
	Frame   uint32        `json:"frame"`
	FPS     uint32        `json:"fps"`
}
type lifetimeOriginalInput53CC90 struct {
	Name     string `json:"name"`
	Kind     int    `json:"kind"`
	Valid    bool   `json:"valid"`
	Sound    uint32 `json:"sound"`
	Mutation string `json:"mutation"`
	lifetimeOriginalState53CC90
}
type lifetimeOriginalEvent53CC90 struct {
	Service string  `json:"service"`
	Args    []int64 `json:"args"`
	lifetimeOriginalState53CC90
}
type lifetimeOriginalFixture53CC90 struct {
	SHA        string `json:"original_sha256"`
	Contexts   int    `json:"contexts"`
	Executions int    `json:"executions"`
	Rows       []struct {
		Input     lifetimeOriginalInput53CC90   `json:"input"`
		Output    lifetimeOriginalState53CC90   `json:"output"`
		Events    []lifetimeOriginalEvent53CC90 `json:"events"`
		Preserved bool                          `json:"nonoutput_bytes_preserved"`
	} `json:"rows"`
}

func TestLifetimeEffectsOriginal53CC90(t *testing.T) {
	blob, err := os.ReadFile("testdata/lifetime_effect_updates_53cc90_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture lifetimeOriginalFixture53CC90
	if err = json.Unmarshal(blob, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || fixture.Contexts != 853 || fixture.Executions != 1706 || len(fixture.Rows) != 853 {
		t.Fatal("original provenance")
	}
	var report *os.File
	if path := os.Getenv("OPENNOX_LIFETIME_STATE_REPORT"); path != "" {
		report, err = os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			t.Fatal(err)
		}
		defer report.Close()
	}
	for _, row := range fixture.Rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			if !row.Preserved {
				t.Fatal("original canary")
			}
			input := row.Input
			frame, fps := input.Frame, input.FPS
			var objs [4]*Object
			var updates [2]*PlayerUpdateData
			var players [2]*Player
			for n := range objs {
				objs[n] = new(Object)
			}
			for n := range updates {
				updates[n] = new(PlayerUpdateData)
				players[n] = new(Player)
			}
			ref := func(n int) *Object {
				if n < 0 {
					return nil
				}
				return objs[n]
			}
			objID := func(o *Object) int {
				if o == nil {
					return -1
				}
				for n, p := range objs {
					if p == o {
						return n
					}
				}
				t.Fatal("unknown object identity")
				return -2
			}
			updID := func(p unsafe.Pointer) int {
				if p == nil {
					return -1
				}
				for n, x := range updates {
					if unsafe.Pointer(x) == p {
						return n
					}
				}
				t.Fatal("unknown update identity")
				return -2
			}
			playerID := func(p *Player) int {
				if p == nil {
					return -1
				}
				for n, x := range players {
					if x == p {
						return n
					}
				}
				t.Fatal("unknown player identity")
				return -2
			}
			for n, v := range input.Objects {
				o := objs[n]
				o.ObjClass = object.Class(v[0])
				o.ObjFlags = object.Flags(v[1])
				o.PosVec = types.Ptf(math.Float32frombits(v[2]), math.Float32frombits(v[3]))
				o.PrevPos = types.Ptf(math.Float32frombits(v[4]), math.Float32frombits(v[5]))
				o.ZVal = math.Float32frombits(v[6])
				o.Field32 = v[7]
				o.Field34 = v[8]
				o.Pos39 = types.Ptf(math.Float32frombits(v[9]), math.Float32frombits(v[10]))
				o.Shape.Circle.R = math.Float32frombits(v[11])
				o.ObjOwner = ref(input.Owners[n])
				if k := input.Updates[n]; k >= 0 {
					o.UpdateData = unsafe.Pointer(updates[k])
				}
			}
			for n, p := range updates {
				if k := input.Players[n]; k >= 0 {
					p.Player = players[k]
				}
			}
			for n, p := range players {
				p.CursorVec = image.Pt(int(int32(input.Cursors[n][0])), int(int32(input.Cursors[n][1])))
			}
			snapshot := func() lifetimeOriginalState53CC90 {
				out := lifetimeOriginalState53CC90{Frame: frame, FPS: fps}
				for n, o := range objs {
					out.Objects[n] = [12]uint32{uint32(o.ObjClass), uint32(o.ObjFlags), math.Float32bits(o.PosVec.X), math.Float32bits(o.PosVec.Y), math.Float32bits(o.PrevPos.X), math.Float32bits(o.PrevPos.Y), math.Float32bits(o.ZVal), o.Field32, o.Field34, math.Float32bits(o.Pos39.X), math.Float32bits(o.Pos39.Y), math.Float32bits(o.Shape.Circle.R)}
					out.Owners[n] = objID(o.ObjOwner)
					out.Updates[n] = updID(o.UpdateData)
				}
				for n, p := range updates {
					out.Players[n] = playerID(p.Player)
				}
				for n, p := range players {
					out.Cursors[n] = [2]uint32{uint32(int32(p.CursorVec.X)), uint32(int32(p.CursorVec.Y))}
				}
				return out
			}
			events := make([]lifetimeOriginalEvent53CC90, 0)
			fxCount := 0
			event := func(name string, args ...int64) {
				events = append(events, lifetimeOriginalEvent53CC90{Service: name, Args: args, lifetimeOriginalState53CC90: snapshot()})
			}
			pointBits := func(p *types.Pointf) []int64 {
				return []int64{int64(math.Float32bits(p.X)), int64(math.Float32bits(p.Y))}
			}
			mutate := func(service string, p *types.Pointf, r *types.Rectf) {
				if service == "fx" {
					fxCount++
				}
				prefix := service + ":"
				if len(input.Mutation) < len(prefix) || input.Mutation[:len(prefix)] != prefix {
					return
				}
				if service == "fx" && fxCount != 1 {
					return
				}
				switch action := input.Mutation[len(prefix):]; action {
				case "owner":
					objs[0].ObjOwner = objs[2]
				case "owner-nil":
					objs[0].ObjOwner = nil
				case "update":
					objs[1].UpdateData = unsafe.Pointer(updates[1])
				case "player":
					updates[0].Player = players[1]
				case "cursor":
					players[0].CursorVec = image.Pt(0x7fffffff, -0x7fffffff)
				case "position":
					objs[0].PosVec = types.Ptf(500.1, -600.3)
				case "radius":
					objs[0].Shape.Circle.R = 53.7
				case "height":
					objs[0].ZVal = 201
				case "unmark":
					objs[0].ObjFlags = 0
				case "deadline":
					objs[0].Field34 = 0
				case "created":
					objs[0].Field32 = frame
				case "clock":
					frame = 0xffffffff
					fps = 0
				case "trace-point":
					r.Max = types.Ptf(17.3, -91.7)
				case "pos-pointer":
					*p = types.Ptf(611.3, -817.7)
				case "prev-position":
					objs[0].PrevPos = types.Ptf(99, 88)
				default:
					t.Fatal("unknown mutation", action)
				}
			}
			h := LifetimeEffectRuntime53CC90{
				Frame: func() uint32 { return frame }, FPS: func() uint32 { return fps },
				DelayedDelete: func(o *Object) {
					if o != objs[0] {
						t.Fatal("delete identity")
					}
					event("delete", 0)
					mutate("delete", nil, nil)
				},
				DamageAround: func(p *types.Pointf, a, b float32, d int, typ object.DamageType, o *Object, x Obj) {
					if o != objs[0] || x != nil {
						t.Fatal("damage identities")
					}
					args := []int64{int64(math.Float32bits(a)), int64(math.Float32bits(b)), int64(d), int64(typ), 0, -1}
					event("damage", append(args, pointBits(p)...)...)
					mutate("damage", p, nil)
				},
				BuffOff: func(o *Object, id EnchantID) { event("buff", int64(objID(o)), int64(id)); mutate("buff", nil, nil) },
				ValidPosition: func(p types.Pointf) bool {
					event("valid", pointBits(&p)...)
					mutate("valid", nil, nil)
					return input.Valid
				},
				Move: func(o *Object, p *types.Pointf) {
					if o != objs[0] {
						t.Fatal("move identity")
					}
					event("move", append([]int64{0}, pointBits(p)...)...)
					mutate("move", p, nil)
					o.PosVec = *p
				},
				SpellSound: func(id spell.ID, kind int) sound.ID {
					event("sound", int64(id), int64(kind))
					mutate("sound", nil, nil)
					return sound.ID(input.Sound)
				},
				Audio: func(id sound.ID, o *Object) {
					event("audio", int64(uint32(id)), int64(objID(o)), 0, 0)
					mutate("audio", nil, nil)
				},
				TraceRay: func(r *types.Rectf, flags MapTraceFlags) bool {
					args := append(pointBits(&r.Min), pointBits(&r.Max)...)
					event("trace", append(args, 0, 0, int64(flags))...)
					mutate("trace", nil, r)
					return input.Valid
				},
				Scorch: func(p *types.Pointf, kind int) {
					event("scorch", append([]int64{int64(kind)}, pointBits(p)...)...)
					mutate("scorch", p, nil)
				},
				PointFX: func(id uint8, p *types.Pointf) {
					event("fx", append([]int64{int64(id)}, pointBits(p)...)...)
					mutate("fx", p, nil)
				},
				Earthquake: func(p *types.Pointf, n int) {
					event("quake", append([]int64{int64(n)}, pointBits(p)...)...)
					mutate("quake", p, nil)
				},
			}
			LifetimeEffectUpdateRaw53CC90(LifetimeEffectKind53CC90(input.Kind), objs[0], h)
			output := snapshot()
			if report != nil {
				if e := json.NewEncoder(report).Encode(map[string]any{"name": input.Name, "output": output, "events": events, "output_matches": reflect.DeepEqual(output, row.Output), "events_match": reflect.DeepEqual(events, row.Events)}); e != nil {
					t.Fatal(e)
				}
			}
			if !reflect.DeepEqual(output, row.Output) {
				t.Errorf("native output differs from unchanged original")
			}
			if !reflect.DeepEqual(events, row.Events) {
				t.Errorf("service arguments/order/state differ: got %d events want %d", len(events), len(row.Events))
			}
		})
	}
}
