package server

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

//go:embed testdata/spell_energy_bolt_native_original_startup_52e820.json
var energyBoltOriginalStartup52E820 []byte

//go:embed testdata/spell_energy_bolt_native_original_nearest_52e820.json
var energyBoltOriginalNearest52E820 []byte

type energyBoltOriginalState52E820 struct {
	Caster       int          `json:"caster"`
	Source       int          `json:"source"`
	Flag20       uint32       `json:"flag20"`
	Target       int          `json:"target"`
	Ray          int          `json:"ray"`
	Carry        uint32       `json:"carry"`
	Level        uint32       `json:"level"`
	Expiry       uint32       `json:"expiry"`
	Start        uint32       `json:"start"`
	Position     [2]uint32    `json:"position"`
	FPS          uint32       `json:"fps"`
	Frame        uint32       `json:"frame"`
	Mode         uint32       `json:"mode"`
	Center       [2]uint32    `json:"center"`
	Selected     int          `json:"selected"`
	Nearest      uint32       `json:"nearest"`
	Classes      [8]uint32    `json:"classes"`
	Subclasses   [8]uint32    `json:"subclasses"`
	Flags        [8]uint32    `json:"flags"`
	Positions    [8][2]uint32 `json:"positions"`
	Directions   [8]uint16    `json:"directions"`
	Buffs        [8]uint32    `json:"buffs"`
	Health       [8]int       `json:"health"`
	DamageFrames [8]uint32    `json:"damage_frames"`
	Updates      [8]int       `json:"updates"`
	Cursors      [8]int       `json:"cursors"`
	Mana         [8]uint16    `json:"mana"`
}
type energyBoltOriginalInput52E820 struct {
	energyBoltOriginalState52E820
	Name           string            `json:"name"`
	Kind           string            `json:"kind"`
	PositionResult uint32            `json:"position_result"`
	Facing         [8][8]uint32      `json:"facing"`
	Enemy          [8][8]uint32      `json:"enemy"`
	LOS            [8][8]uint32      `json:"los"`
	DistanceBits64 [8][8]string      `json:"distance_bits64"`
	Balances       map[string]string `json:"balances"`
	DamageBits64   string            `json:"damage_bits64"`
	CastSound      uint16            `json:"cast_sound"`
	Candidates     []int             `json:"candidates"`
	Mutation       struct {
		Stage  string `json:"stage"`
		At     int    `json:"at"`
		Action string `json:"action"`
	} `json:"mutation"`
}
type energyBoltOriginalEvent52E820 struct {
	Kind  string                        `json:"kind"`
	Args  []json.RawMessage             `json:"args"`
	State energyBoltOriginalState52E820 `json:"state"`
}
type energyBoltOriginalResult52E820 struct {
	Ret    uint32                          `json:"ret"`
	State  energyBoltOriginalState52E820   `json:"state"`
	Events []energyBoltOriginalEvent52E820 `json:"events"`
}

// Both original PC53 contexts are explicit. The existing Go runtime's default
// nearest behavior remains tested alongside the production startup chop mode.
// Reflection makes the identical comparator runnable before the new runtime
// fields exist; only that baseline's unavailable scratch observations project
// out. All argument/state/selection behavior remains compared on the baseline.
func TestSpellEnergyBoltOriginalNative52E820(t *testing.T) {
	for _, profile := range []struct {
		name    string
		control int
		data    []byte
		hash    string
	}{
		{"startup", 0xe3f, energyBoltOriginalStartup52E820, "8f5342ee7d774f19382adde3eb5ebab917165e086b418b883de12502c630ae21"},
		{"nearest", 0x23f, energyBoltOriginalNearest52E820, "78db99fcf3ee85c43ace711cedb82760df79e00692a220b6563f06974a0d74a8"},
	} {
		t.Run(profile.name, func(t *testing.T) {
			digest := sha256.Sum256(profile.data)
			if hex.EncodeToString(digest[:]) != profile.hash {
				t.Fatal("sealed original fixture hash")
			}
			var fixture struct {
				OriginalSHA256 string `json:"original_sha256"`
				ControlWord    int    `json:"control_word"`
				Rows           []struct {
					Input  energyBoltOriginalInput52E820  `json:"input"`
					Result energyBoltOriginalResult52E820 `json:"result"`
				} `json:"rows"`
			}
			if err := json.Unmarshal(profile.data, &fixture); err != nil {
				t.Fatal(err)
			}
			if fixture.OriginalSHA256 != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || fixture.ControlWord != profile.control || len(fixture.Rows) != 622 {
				t.Fatal("original identity/profile/contexts")
			}
			for _, row := range fixture.Rows {
				t.Run(row.Input.Name, func(t *testing.T) {
					in := row.Input
					objs := new([8]Object)
					updates := new([8]PlayerUpdateData)
					health := new([8]HealthData)
					obj := func(i int) *Object {
						if i < 0 {
							return nil
						}
						return &objs[i]
					}
					idx := func(o *Object) int {
						if o == nil {
							return -1
						}
						for i := range objs {
							if o == &objs[i] {
								return i
							}
						}
						panic("foreign object")
					}
					ui := func(p unsafe.Pointer) int {
						for i := range updates {
							if p == unsafe.Pointer(&updates[i]) {
								return i
							}
						}
						panic("foreign update")
					}
					for i := range objs {
						objs[i] = Object{ObjClass: object.Class(in.Classes[i]), ObjSubClass: object.SubClass(in.Subclasses[i]), ObjFlags: object.Flags(in.Flags[i]), PosVec: plasmaOriginalPoint531580(in.Positions[i]), Direction1: Dir16(in.Directions[i]), Buffs: in.Buffs[i], Frame134: in.DamageFrames[i], UpdateData: unsafe.Pointer(&updates[in.Updates[i]])}
						if in.Health[i] != 0 {
							objs[i].HealthData = &health[i]
						}
						updates[i].CursorObj = obj(in.Cursors[i])
						updates[i].ManaCur = in.Mana[i]
					}
					r := &DurSpell{Caster16: obj(in.Caster), Obj12: obj(in.Source), Flag20: in.Flag20, Target48: obj(in.Target), Field72: int32(in.Carry), Level: in.Level, Frame68: in.Expiry, Frame60: in.Start, Pos: plasmaOriginalPoint531580(in.Position)}
					ray := obj(in.Ray)
					fps, frame := in.FPS, in.Frame
					mode, center, selected, nearest := in.Mode, plasmaOriginalPoint531580(in.Center), obj(in.Selected), math.Float32frombits(in.Nearest)
					rt := SpellEnergyBoltRuntime52E820{}
					runtime := reflect.ValueOf(&rt).Elem()
					scratch := reflect.Value{}
					if field := runtime.FieldByName("Selection"); field.IsValid() {
						scratch = reflect.New(field.Type().Elem())
						field.Set(scratch)
					}
					if field := runtime.FieldByName("RoundTowardZero"); field.IsValid() {
						field.SetBool(profile.control == 0xe3f)
					}
					storeScratch := func() {
						if scratch.IsValid() {
							s := scratch.Elem()
							s.FieldByName("Mode").SetUint(uint64(mode))
							s.FieldByName("Center").Set(reflect.ValueOf(center))
							s.FieldByName("Target").Set(reflect.ValueOf(selected))
							s.FieldByName("Nearest").SetFloat(float64(nearest))
						}
					}
					loadScratch := func() {
						if scratch.IsValid() {
							s := scratch.Elem()
							mode = uint32(s.FieldByName("Mode").Uint())
							center = s.FieldByName("Center").Interface().(types.Pointf)
							selected = s.FieldByName("Target").Interface().(*Object)
							nearest = float32(s.FieldByName("Nearest").Float())
						}
					}
					storeScratch()
					snapshot := func() energyBoltOriginalState52E820 {
						loadScratch()
						s := energyBoltOriginalState52E820{Caster: idx(r.Caster16), Source: idx(r.Obj12), Flag20: r.Flag20, Target: idx(r.Target48), Ray: idx(ray), Carry: uint32(r.Field72), Level: r.Level, Expiry: r.Frame68, Start: r.Frame60, Position: plasmaOriginalBits531580(r.Pos), FPS: fps, Frame: frame, Mode: mode, Center: plasmaOriginalBits531580(center), Selected: idx(selected), Nearest: math.Float32bits(nearest)}
						for i := range objs {
							o, d := &objs[i], &updates[i]
							s.Classes[i] = uint32(o.ObjClass)
							s.Subclasses[i] = uint32(o.ObjSubClass)
							s.Flags[i] = uint32(o.ObjFlags)
							s.Positions[i] = plasmaOriginalBits531580(o.PosVec)
							s.Directions[i] = uint16(o.Direction1)
							s.Buffs[i] = o.Buffs
							if o.HealthData != nil {
								s.Health[i] = 1
							}
							s.DamageFrames[i] = o.Frame134
							s.Updates[i] = ui(o.UpdateData)
							s.Cursors[i] = idx(d.CursorObj)
							s.Mana[i] = d.ManaCur
						}
						return s
					}
					events := make([]energyBoltOriginalEvent52E820, 0)
					calls := map[string]int{}
					changed := false
					glyphPath := false
					event := func(kind string, args ...any) {
						raw := make([]json.RawMessage, len(args))
						for i, v := range args {
							raw[i], _ = json.Marshal(v)
						}
						events = append(events, energyBoltOriginalEvent52E820{kind, raw, snapshot()})
						at := calls[kind]
						calls[kind]++
						m := in.Mutation
						if changed || m.Stage != kind || m.At != at {
							return
						}
						changed = true
						switch m.Action {
						case "caster-b":
							r.Caster16 = obj(1)
						case "caster-nil":
							r.Caster16 = nil
						case "source-b":
							r.Obj12 = obj(1)
						case "target-b":
							r.Target48 = obj(3)
						case "target-nil":
							r.Target48 = nil
						case "ray-b":
							ray = obj(4)
						case "ray-nil":
							ray = nil
						case "carry":
							r.Field72 = int32(math.Float32bits(-0.75))
						case "level":
							r.Level = 0xffffffff
						case "flag20":
							r.Flag20 = 1
						case "expiry":
							r.Frame68 = 0x13579bdf
						case "start":
							r.Frame60 = 0xfffffff0
						case "record-position":
							r.Pos = types.Ptf(99.25, -88.5)
						case "fps":
							fps = 60
						case "frame":
							frame = 0xfffffff9
						case "selected-b":
							selected = obj(5)
							storeScratch()
						case "nearest":
							nearest = 4
							storeScratch()
						case "mode":
							mode = 1
							storeScratch()
						case "center":
							center = types.Ptf(3, 4)
							storeScratch()
						case "caster-update-b":
							objs[0].UpdateData = unsafe.Pointer(&updates[1])
						case "cursor-b":
							updates[0].CursorObj = obj(3)
						case "caster-position":
							objs[0].PosVec = types.Ptf(3, 4)
						case "target-position":
							objs[2].PosVec = types.Ptf(1.25, -2.5)
						case "target-dead":
							objs[2].ObjFlags |= object.Flags(0x8020)
						case "caster-class-zero":
							objs[0].ObjClass = 0
						case "caster-class-monster":
							objs[0].ObjClass = object.ClassMonster
						case "mana-zero":
							updates[0].ManaCur = 0
						case "mana-one":
							updates[0].ManaCur = 1
						default:
							t.Fatalf("unknown mutation %q", m.Action)
						}
					}
					rt.Frame = func() uint32 { return frame }
					rt.TickRate = func() uint32 { return fps }
					balance := func(key string) float64 {
						event("balance", key)
						if key == "LightningRange" {
							glyphPath = r.Flag20 != 0
						}
						return plasmaOriginalDouble531580(in.Balances[key])
					}
					levelBalance := func(key string, level uint32) float64 {
						event("balance-level", key, level)
						return plasmaOriginalDouble531580(in.DamageBits64)
					}
					rt.Balance = func(key string) float32 { return float32(balance(key)) }
					rt.BalanceLevel = func(key string, level uint32) float32 { return float32(levelBalance(key, level)) }
					if field := runtime.FieldByName("Balance64"); field.IsValid() {
						field.Set(reflect.ValueOf(balance))
					}
					if field := runtime.FieldByName("BalanceLevel64"); field.IsValid() {
						field.Set(reflect.ValueOf(levelBalance))
					}
					rt.CancelSpell = func(id int32, o *Object) { event("cancel", id, idx(o)) }
					rt.PositionDelta = func(o *Object, p *types.Pointf) int32 {
						v := plasmaOriginalBits531580(*p)
						event("position", idx(o), v[0], v[1])
						return int32(in.PositionResult)
					}
					rt.InFront = func(a, b *Object) bool {
						i, j := idx(a), idx(b)
						pa, pb := plasmaOriginalBits531580(a.PosVec), plasmaOriginalBits531580(b.PosVec)
						direction := int16(a.Direction1)
						event("facing", i, int32(direction), j, pa[0], pa[1], pb[0], pb[1])
						return in.Facing[i][j]&1 != 0
					}
					distance := func(a, b *Object) float64 {
						i, j := idx(a), idx(b)
						event("distance", i, j)
						return plasmaOriginalDouble531580(in.DistanceBits64[i][j])
					}
					if field := runtime.FieldByName("Distance"); field.IsValid() {
						field.Set(reflect.ValueOf(distance))
					}
					rt.IsEnemy = func(a, b *Object) bool { i, j := idx(a), idx(b); event("enemy", i, j); return in.Enemy[i][j] != 0 }
					rt.CanInteract = func(a, b *Object) bool { i, j := idx(a), idx(b); event("los", i, j); return in.LOS[i][j] != 0 }
					rt.ObjectsInCircle = func(p types.Pointf, radius float32, visit func(*Object) bool) {
						caster := r.Caster16
						v := plasmaOriginalBits531580(p)
						event("circle", idx(caster), v[0], v[1], math.Float32bits(radius))
						for _, i := range in.Candidates {
							event("visit", i, idx(caster))
							if !visit(obj(i)) {
								break
							}
						}
					}
					rt.StartRay = func(*DurSpell) { event("start") }
					rt.StopRay = func(_ *DurSpell, o *Object) { event("stop", idx(o)) }
					rt.Damage = func(a, b *Object, amount int32) { event("damage", idx(a), idx(b), amount) }
					rt.PointFX = func(code uint8, p types.Pointf) {
						owner := -1
						if in.Kind != "create" {
							if glyphPath {
								loadScratch()
								owner = idx(selected)
							} else {
								owner = idx(r.Target48)
							}
						}
						v := plasmaOriginalBits531580(p)
						event("fx", code, owner, v[0], v[1])
					}
					rt.Audio = func(id uint16, o *Object) { event("audio", id, idx(o)) }
					rt.SetPlayerState = func(o *Object, state PlayerState) { event("state", idx(o), state) }
					rt.CastSound = func() uint16 { event("cast-sound"); return in.CastSound }
					rt.ManaSub = func(o *Object, amount int32) {
						event("mana-sub", idx(o), amount)
						if o.ObjClass.Has(object.ClassPlayer) {
							d := o.UpdateDataPlayer()
							v := int32(d.ManaCur) - amount
							if v < 0 {
								v = 0
							}
							d.ManaCur = uint16(v)
						}
					}
					rt.LoadRayTarget = func(*DurSpell) *Object { return ray }
					rt.StoreRayTarget = func(_ *DurSpell, o *Object) { ray = o }
					var ret uint32
					switch in.Kind {
					case "create":
						ret = uint32(SpellEnergyBoltCreate52E820(r, rt))
					case "update":
						ret = uint32(SpellEnergyBoltUpdate52E850(r, rt))
					default:
						t.Fatal("unknown root")
					}
					if in.Mutation.Stage != "" && !changed {
						t.Fatal("declared mutation not reached")
					}
					got, want := energyBoltOriginalResult52E820{ret, snapshot(), events}, row.Result
					if !scratch.IsValid() {
						t.Log("pre-port baseline lacks native selection scratch; only those observations projected out")
						clear := func(s *energyBoltOriginalState52E820) {
							s.Mode = 0
							s.Center = [2]uint32{}
							s.Selected = 0
							s.Nearest = 0
						}
						// Glyph FX receives coordinates by value; its scratch-derived owner
						// label is unavailable on the baseline, while coordinates still compare.
						if glyphPath {
							for i := range got.Events {
								if got.Events[i].Kind == "fx" {
									got.Events[i].Args[1] = json.RawMessage("0")
								}
							}
							for i := range want.Events {
								if want.Events[i].Kind == "fx" {
									want.Events[i].Args[1] = json.RawMessage("0")
								}
							}
						}
						clear(&got.State)
						clear(&want.State)
						for i := range got.Events {
							clear(&got.Events[i].State)
						}
						for i := range want.Events {
							clear(&want.Events[i].State)
						}
					}
					if !reflect.DeepEqual(got, want) {
						actual, _ := json.Marshal(got)
						expected, _ := json.Marshal(want)
						t.Fatalf("%s got %s\nwant %s", fmt.Sprintf("control=%#x", profile.control), actual, expected)
					}
				})
			}
		})
	}
}
