package server

import (
	"encoding/json"
	"math"
	"os"
	"reflect"
	"strconv"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func forceOriginalAllocation52EF30[T comparable](t *testing.T) *T {
	t.Helper()
	var zero T
	p, free := alloc.New(zero)
	t.Cleanup(free)
	if uintptr(unsafe.Pointer(p)) <= 0xffffffff {
		t.Fatal("native original comparison allocation must exceed4GiB")
	}
	return p
}

// The independent fixture retains all original results. The duration manager
// discards destructor EAX and then frees its record, so only those two dead
// outputs are canonically excluded. The native charge sidecar is compared
// with the original pointer slot on create/update; Field76 stays opaque.
func TestSpellForceOfNatureOriginalNative52EF30(t *testing.T) {
	if strconv.IntSize != 64 {
		t.Skip("native C-owned comparison above4GiB")
	}
	path := os.Getenv("NOX_FORCE_OF_NATURE_ORACLE_FIXTURE")
	if path == "" {
		path = "testdata/spell_force_of_nature_native_original_52ef30.json"
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Control uint32 `json:"control_word"`
		Cases   []struct {
			Input  map[string]any `json:"input"`
			Result map[string]any `json:"result"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Control != 0xe3f || len(fixture.Cases) != 925 {
		t.Fatal("sealed startup original corpus")
	}
	u := func(c map[string]any, key string) uint32 { return uint32(c[key].(float64)) }
	pt := func(value any) types.Pointf {
		p := value.([]any)
		return types.Ptf(math.Float32frombits(uint32(p[0].(float64))), math.Float32frombits(uint32(p[1].(float64))))
	}
	pointBits := func(p types.Pointf) []uint32 { return []uint32{math.Float32bits(p.X), math.Float32bits(p.Y)} }
	canonical := func(value any, destructor bool) any {
		encoded, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		var copy map[string]any
		if err := json.Unmarshal(encoded, &copy); err != nil {
			t.Fatal(err)
		}
		if destructor {
			copy["return_value"] = float64(0)
			copy["state"].(map[string]any)["charge"] = float64(-2)
		}
		return copy
	}
	for _, row := range fixture.Cases {
		c := row.Input
		t.Run(c["name"].(string), func(t *testing.T) {
			objects := make([]*Object, 8)
			for i := range objects {
				objects[i] = forceOriginalAllocation52EF30[Object](t)
				*objects[i] = Object{ObjClass: object.ClassMonster, SpeedCur: math.Float32frombits(u(c, "speed"))}
			}
			index := func(p *Object) int {
				if p == nil {
					return -1
				}
				for i, o := range objects {
					if p == o {
						return i
					}
				}
				t.Fatal("object outside sealed identity set")
				return -2
			}
			uses := []*WandUseData{forceOriginalAllocation52EF30[WandUseData](t), forceOriginalAllocation52EF30[WandUseData](t)}
			for i, actor := range []*Object{objects[0], objects[2]} {
				update := forceOriginalAllocation52EF30[PlayerUpdateData](t)
				actor.UpdateData = unsafe.Pointer(update)
				wand := objects[6+i]
				wand.ObjSubClass = object.SubClass(u(c, "wand_subclass"))
				wand.UseData = UseDataPtr{Ptr: unsafe.Pointer(uses[i])}
				uses[i].Flags = u(c, "wand_flags")
				if c["wand"].(bool) {
					update.EquippedWeapon = wand
				}
			}
			objects[0].ObjClass = object.Class(u(c, "class"))
			objects[2].ObjClass = object.ClassPlayer
			objects[0].Buffs = u(c, "buffs")
			objects[0].PosVec, objects[0].Direction1 = pt(c["caster_pos"]), Dir16(u(c, "direction"))
			objects[1].PosVec, objects[1].Direction1 = types.Ptf(71.125, 93.5), 192
			objects[2].PosVec, objects[2].Direction1 = types.Ptf(211.25, -41.5), 33
			objects[0].Shape.Kind = ShapeKind(u(c, "shape"))
			objects[0].Shape.Circle.R = math.Float32frombits(u(c, "radius"))
			objects[0].Shape.Box.W = math.Float32frombits(u(c, "width"))
			objects[0].Shape.Box.H = math.Float32frombits(u(c, "height"))
			type frame struct {
				Before [32]byte
				Record DurSpell
				After  [32]byte
			}
			f := forceOriginalAllocation52EF30[frame](t)
			for i := range f.Before {
				f.Before[i], f.After[i] = byte(0x51+i), byte(0xe7-i)
			}
			r := &f.Record
			*r = DurSpell{ID: 0xabcd, Level: 3, Caster16: objects[0], Obj24: objects[1], Flag20: u(c, "mode"), Pos: pt(c["pos"]), Frame68: u(c, "expiry"), Field72: int32(u(c, "field72")), Field76: 0x3fdccccc, Flags88: u(c, "flags88")}
			if c["caster_nil"].(bool) {
				r.Caster16 = nil
			}
			var charge *Object
			if c["charge"].(bool) {
				charge = objects[3]
			}
			frameNumber := u(c, "frame")
			state := func() map[string]any {
				return map[string]any{
					"caster": index(r.Caster16), "mode": r.Flag20, "anchor": index(r.Obj24), "pos": pointBits(r.Pos),
					"field72": uint32(r.Field72), "charge": index(charge), "flags88": r.Flags88, "frame": frameNumber, "expiry": r.Frame68,
					"caster_pos": pointBits(objects[0].PosVec), "caster_direction": uint16(objects[0].Direction1),
					"wand_flags": uses[0].Flags, "alternate_wand_flags": uses[1].Flags,
					"charge_pos": pointBits(objects[4].PosVec), "ball_pos": pointBits(objects[5].PosVec),
					"ball_velocity": pointBits(objects[5].VelVec), "ball_speed": math.Float32bits(objects[5].SpeedCur),
					"ball_direction": []uint16{uint16(objects[5].Direction1), uint16(objects[5].Direction2)},
				}
			}
			events := make([]map[string]any, 0)
			calls := make(map[string]uint32)
			changed := false
			mutation := c["mutation"].(map[string]any)
			event := func(kind string, values []any) {
				events = append(events, map[string]any{"kind": kind, "args": values, "state": state()})
				n := calls[kind]
				calls[kind]++
				if changed || mutation["stage"].(string) != kind || uint32(mutation["at"].(float64)) != n {
					return
				}
				changed = true
				switch mutation["action"].(string) {
				case "caster-c":
					r.Caster16 = objects[2]
				case "mode-one":
					r.Flag20 = 1
				case "mode-zero":
					r.Flag20 = 0
				case "caster-position":
					objects[0].PosVec = types.Ptf(8193.125, -4097.25)
				case "caster-direction":
					objects[0].Direction1 = 192
				case "record-position":
					r.Pos = types.Ptf(-73.25, 97.125)
				case "frame-launch":
					frameNumber = r.Frame68 - 1
				case "expiry-next":
					r.Frame68 = frameNumber + 1
				case "ball-speed":
					objects[5].SpeedCur = 3.125
				case "wand-flags":
					uses[0].Flags = 0xfffffffd
				default:
					t.Fatal("unknown sealed mutation")
				}
			}
			runtime := SpellForceOfNatureRuntime52EF30{
				NewObject: func(id string) *Object {
					event("new", []any{id})
					if !c["allocate"].(bool) {
						return nil
					}
					if id == "ForceOfNatureCharge" {
						return objects[4]
					}
					if id == "DeathBall" {
						return objects[5]
					}
					t.Fatal("unrecognized original allocation type")
					return nil
				},
				CreateAt: func(p, owner *Object, pos types.Pointf) {
					event("create", []any{index(p), index(owner), pointBits(pos), 0})
					p.PosVec = pos
				},
				DelayedDelete: func(p *Object) { event("delete", []any{index(p)}) },
				LoadCharge:    func(*DurSpell) *Object { return charge },
				StoreCharge:   func(_ *DurSpell, p *Object) { charge = p },
				CurrentFrame:  func() uint32 { return frameNumber },
				TraceRay: func(a, b types.Pointf, flags MapTraceFlags) bool {
					event("trace", []any{pointBits(a), pointBits(b), uint32(flags)})
					return c["trace"].(bool)
				},
				SetPlayerState: func(p *Object, s PlayerState) { event("state", []any{index(p), uint32(s)}) },
				EventObj:       func(p *Object) { event("audio", []any{38, index(p), 0, 0}) },
			}
			function := c["function"].(string)
			var result int32
			switch function {
			case "create":
				result = SpellForceOfNatureCreate52EF30(r, runtime)
			case "update":
				result = SpellForceOfNatureUpdate52EFD0(r, runtime)
			case "destroy":
				SpellForceOfNatureDestroy52F1D0(r, runtime)
			default:
				t.Fatal("unknown sealed entry")
			}
			actual := map[string]any{"return_value": uint32(result), "events": events, "state": state()}
			if !reflect.DeepEqual(canonical(actual, function == "destroy"), canonical(row.Result, function == "destroy")) {
				a, _ := json.Marshal(actual)
				w, _ := json.Marshal(row.Result)
				t.Fatalf("native=%s\noriginal=%s", a, w)
			}
			if r.Field76 != 0x3fdccccc {
				t.Fatal("native pointer leaked into opaque PE32 field")
			}
			for i := range f.Before {
				if f.Before[i] != byte(0x51+i) || f.After[i] != byte(0xe7-i) {
					t.Fatal("native duration canary")
				}
			}
		})
	}
}
