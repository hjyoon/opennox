package server

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	playerlib "github.com/opennox/libs/player"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

//go:embed testdata/spell_drain_mana_native_original_startup_52e210.json
var drainOriginalStartup52E210 []byte

//go:embed testdata/spell_drain_mana_native_original_nearest_52e210.json
var drainOriginalNearest52E210 []byte

type drainOriginalState52E210 struct {
	Caster        int          `json:"caster"`
	Source        int          `json:"source"`
	Glyph         uint32       `json:"glyph"`
	Target        int          `json:"target"`
	Ray           int          `json:"ray"`
	Carry         uint32       `json:"carry"`
	Expiry        uint32       `json:"expiry"`
	Level         uint32       `json:"level"`
	Position      [2]uint32    `json:"position"`
	Selected      int          `json:"selected"`
	Center        [2]uint32    `json:"center"`
	Nearest       uint32       `json:"nearest"`
	FPS           uint32       `json:"fps"`
	Frame         uint32       `json:"frame"`
	Quest         uint32       `json:"quest"`
	Classes       [4]uint32    `json:"classes"`
	Subclasses    [4]uint32    `json:"subclasses"`
	Flags         [4]uint32    `json:"flags"`
	TeamIDs       [4]uint8     `json:"team_ids"`
	SyncFlags     [4]uint32    `json:"sync_flags"`
	Positions     [4][2]uint32 `json:"positions"`
	Buffs         [4]uint32    `json:"buffs"`
	Health        [4]int       `json:"health"`
	DamageFrames  [4]uint32    `json:"damage_frames"`
	Updates       [4]int       `json:"updates"`
	Mana          [4]uint16    `json:"mana"`
	Maximum       [4]uint16    `json:"maximum"`
	ObeliskMana   [4]int32     `json:"obelisk_mana"`
	MonsterStatus [4]uint32    `json:"monster_status"`
	Players       [4]int       `json:"players"`
	PlayerClasses [4]uint8     `json:"player_classes"`
	PlayerFlags   [4]uint32    `json:"player_flags"`
	Multipliers   [3]uint32    `json:"multipliers"`
}

type drainOriginalInput52E210 struct {
	drainOriginalState52E210
	Name        string       `json:"name"`
	Kind        string       `json:"kind"`
	Range       string       `json:"range"`
	Coefficient string       `json:"coefficient"`
	Candidates  []int        `json:"candidates"`
	Enemy       [4][4]uint32 `json:"enemy"`
	SameTeam    [4][5]uint32 `json:"same_team"`
	Trace       uint32       `json:"trace"`
	Moved       uint32       `json:"moved"`
	Amount      int32        `json:"amount"`
	Mutation    struct {
		Stage  string `json:"stage"`
		At     int    `json:"at"`
		Action string `json:"action"`
	} `json:"mutation"`
}
type drainOriginalEvent52E210 struct {
	Kind  string                   `json:"kind"`
	Args  []json.RawMessage        `json:"args"`
	State drainOriginalState52E210 `json:"state"`
}
type drainOriginalResult52E210 struct {
	Ret    uint32                     `json:"ret"`
	State  drainOriginalState52E210   `json:"state"`
	Events []drainOriginalEvent52E210 `json:"events"`
}

func drainOriginalAllocation52E210[T comparable](t *testing.T) *T {
	t.Helper()
	var zero T
	size := int(unsafe.Sizeof(zero))
	raw, free := alloc.Make([]byte(nil), size+64)
	for i := 0; i < 32; i++ {
		raw[i], raw[size+32+i] = byte(0x31+i), byte(0xe7-i)
	}
	t.Cleanup(func() {
		defer free()
		for i := 0; i < 32; i++ {
			if raw[i] != byte(0x31+i) || raw[size+32+i] != byte(0xe7-i) {
				t.Errorf("native reference allocation canary %d", i)
			}
		}
	})
	return (*T)(unsafe.Pointer(&raw[32]))
}

// Identical before/after comparator. Reflection attaches new runtime fields
// only when present. Baseline scratch alone is unavailable and projected out.
// Original dying records retain +36; native private sidecars drop that link
// after the no-target stop. That separate lifetime policy is asserted below.
func TestSpellDrainManaOriginalNative52E210(t *testing.T) {
	for _, profile := range []struct {
		name    string
		control int
		data    []byte
		hash    string
	}{
		{"startup", 0xe3f, drainOriginalStartup52E210, "8e3cef4beae2ba8a220f788a90e38414149d40c1326e5ba3b13c6aa588a1273b"},
		{"nearest", 0x23f, drainOriginalNearest52E210, "29945c7bd3b5db1cabda44a793e561dc7a0a995816354f4c53e68dfee913fedb"},
	} {
		t.Run(profile.name, func(t *testing.T) {
			digest := sha256.Sum256(profile.data)
			if hex.EncodeToString(digest[:]) != profile.hash {
				t.Fatal("original fixture seal changed")
			}
			var fixture struct {
				Control int `json:"control_word"`
				Rows    []struct {
					Input  drainOriginalInput52E210  `json:"input"`
					Result drainOriginalResult52E210 `json:"result"`
				} `json:"rows"`
			}
			if err := json.Unmarshal(profile.data, &fixture); err != nil {
				t.Fatal(err)
			}
			if fixture.Control != profile.control || len(fixture.Rows) != 835 {
				t.Fatal("original profile/contexts")
			}
			for _, row := range fixture.Rows {
				t.Run(row.Input.Name, func(t *testing.T) {
					in := row.Input
					objs := drainOriginalAllocation52E210[[4]Object](t)
					infos := drainOriginalAllocation52E210[[4]Player](t)
					health := drainOriginalAllocation52E210[[4]HealthData](t)
					r := drainOriginalAllocation52E210[DurSpell](t)
					// A C-owned universal update frame holds the disjoint native
					// fields used by Player/Monster/Obelisk views. All pointers in
					// it point to other C allocations; no Go pointer is hidden in C.
					maxSize := max(unsafe.Sizeof(PlayerUpdateData{}), unsafe.Sizeof(MonsterUpdateData{}), unsafe.Sizeof(ObeliskUpdateData{}))
					type span struct{ off, size uintptr }
					spans := []span{{unsafe.Offsetof(PlayerUpdateData{}.ManaCur), 2}, {unsafe.Offsetof(PlayerUpdateData{}.ManaMax), 2}, {unsafe.Offsetof(PlayerUpdateData{}.Player), unsafe.Sizeof(uintptr(0))}, {unsafe.Offsetof(MonsterUpdateData{}.StatusFlags), 4}, {unsafe.Offsetof(ObeliskUpdateData{}.Mana), 4}}
					for i, a := range spans {
						for j, b := range spans {
							if i < j && a.off < b.off+b.size && b.off < a.off+a.size {
								t.Fatal("native update observation fields overlap")
							}
						}
					}
					var updates [4]unsafe.Pointer
					for i := range updates {
						data, free := alloc.Make([]byte(nil), int(maxSize)+64)
						for j := 0; j < 32; j++ {
							data[j], data[int(maxSize)+32+j] = byte(0x51+j), byte(0xd7-j)
						}
						t.Cleanup(func() {
							defer free()
							for j := 0; j < 32; j++ {
								if data[j] != byte(0x51+j) || data[int(maxSize)+32+j] != byte(0xd7-j) {
									t.Errorf("native update canary%d", j)
								}
							}
						})
						updates[i] = unsafe.Pointer(&data[32])
					}
					pd := func(i int) *PlayerUpdateData { return (*PlayerUpdateData)(updates[i]) }
					md := func(i int) *MonsterUpdateData { return (*MonsterUpdateData)(updates[i]) }
					od := func(i int) *ObeliskUpdateData { return (*ObeliskUpdateData)(updates[i]) }
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
						for i, v := range updates {
							if p == v {
								return i
							}
						}
						panic("foreign update")
					}
					pi := func(p *Player) int {
						for i := range infos {
							if p == &infos[i] {
								return i
							}
						}
						panic("foreign player")
					}
					for i := range objs {
						objs[i] = Object{ObjClass: object.Class(in.Classes[i]), ObjSubClass: object.SubClass(in.Subclasses[i]), ObjFlags: object.Flags(in.Flags[i]), PosVec: plasmaOriginalPoint531580(in.Positions[i]), Buffs: in.Buffs[i], Frame134: in.DamageFrames[i], UpdateData: updates[in.Updates[i]], Field38: in.SyncFlags[i]}
						objs[i].TeamVal.ID = TeamID(in.TeamIDs[i])
						if in.Health[i] != 0 {
							objs[i].HealthData = &health[i]
						}
						pd(i).ManaCur, pd(i).ManaMax, pd(i).Player = in.Mana[i], in.Maximum[i], &infos[in.Players[i]]
						md(i).StatusFlags = object.MonsterStatus(in.MonsterStatus[i])
						od(i).Mana = in.ObeliskMana[i]
						infos[i].Info().SetPlayerClass(playerlib.Class(in.PlayerClasses[i]))
						infos[i].Field3680 = in.PlayerFlags[i]
					}
					*r = DurSpell{Caster16: obj(in.Caster), Obj12: obj(in.Source), Flag20: in.Glyph, Target48: obj(in.Target), Field72: int32(in.Carry), Level: in.Level, Frame68: in.Expiry, Pos: plasmaOriginalPoint531580(in.Position)}
					ray, selected := obj(in.Ray), obj(in.Selected)
					center, nearest := plasmaOriginalPoint531580(in.Center), math.Float32frombits(in.Nearest)
					fps, frame, quest := in.FPS, in.Frame, in.Quest
					multipliers := in.Multipliers
					rt := SpellDrainManaRuntime52E210{}
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
							s.FieldByName("Center").Set(reflect.ValueOf(center))
							s.FieldByName("Target").Set(reflect.ValueOf(selected))
							s.FieldByName("Nearest").SetFloat(float64(nearest))
						}
					}
					loadScratch := func() {
						if scratch.IsValid() {
							s := scratch.Elem()
							center = s.FieldByName("Center").Interface().(types.Pointf)
							selected = s.FieldByName("Target").Interface().(*Object)
							nearest = float32(s.FieldByName("Nearest").Float())
						}
					}
					storeScratch()
					snapshot := func() drainOriginalState52E210 {
						loadScratch()
						s := drainOriginalState52E210{Caster: idx(r.Caster16), Source: idx(r.Obj12), Glyph: r.Flag20, Target: idx(r.Target48), Ray: idx(ray), Carry: uint32(r.Field72), Expiry: r.Frame68, Level: r.Level, Position: plasmaOriginalBits531580(r.Pos), Selected: idx(selected), Center: plasmaOriginalBits531580(center), Nearest: math.Float32bits(nearest), FPS: fps, Frame: frame, Quest: quest, Multipliers: multipliers}
						for i := range objs {
							o := &objs[i]
							s.Classes[i] = uint32(o.ObjClass)
							s.Subclasses[i] = uint32(o.ObjSubClass)
							s.Flags[i] = uint32(o.ObjFlags)
							s.TeamIDs[i] = uint8(o.TeamVal.ID)
							s.SyncFlags[i] = o.Field38
							s.Positions[i] = plasmaOriginalBits531580(o.PosVec)
							s.Buffs[i] = o.Buffs
							if o.HealthData != nil {
								s.Health[i] = 1
							}
							s.DamageFrames[i] = o.Frame134
							s.Updates[i] = ui(o.UpdateData)
							s.Mana[i] = pd(i).ManaCur
							s.Maximum[i] = pd(i).ManaMax
							s.ObeliskMana[i] = od(i).Mana
							s.MonsterStatus[i] = uint32(md(i).StatusFlags)
							s.Players[i] = pi(pd(i).Player)
							s.PlayerClasses[i] = uint8(infos[i].PlayerClass())
							s.PlayerFlags[i] = infos[i].Field3680
						}
						return s
					}
					events := make([]drainOriginalEvent52E210, 0)
					calls := map[string]int{}
					changed := false
					var enumCaster *Object
					event := func(kind string, args ...any) {
						raw := make([]json.RawMessage, len(args))
						for i, v := range args {
							raw[i], _ = json.Marshal(v)
						}
						events = append(events, drainOriginalEvent52E210{kind, raw, snapshot()})
						at := calls[kind]
						calls[kind]++
						m := in.Mutation
						if changed || m.Stage != kind || m.At != at {
							return
						}
						changed = true
						switch m.Action {
						case "caster-b":
							r.Caster16 = obj(2)
						case "caster-nil":
							r.Caster16 = nil
						case "source-b":
							r.Obj12 = obj(2)
						case "target-b":
							r.Target48 = obj(3)
						case "ray-b":
							ray = obj(3)
						case "ray-nil":
							ray = nil
						case "carry":
							r.Field72 = int32(math.Float32bits(-.75))
						case "level":
							r.Level = 0xffffffff
						case "glyph":
							r.Flag20 = 1
						case "expiry":
							r.Frame68 = 0x13579bdf
						case "record-position":
							r.Pos = types.Ptf(99.25, -88.5)
						case "caster-position":
							objs[0].PosVec = types.Ptf(103.25, 204.5)
						case "target-position":
							objs[1].PosVec = types.Ptf(111.25, 222.5)
						case "candidate-position":
							objs[2].PosVec = types.Ptf(11.25, -22.5)
						case "fps":
							fps = 60
						case "frame":
							frame = 0xfffffff9
						case "selected-b":
							selected = obj(3)
							storeScratch()
						case "nearest":
							nearest = 4
							storeScratch()
						case "center":
							center = types.Ptf(3, 4)
							storeScratch()
						case "caster-update-b":
							objs[0].UpdateData = updates[2]
						case "target-update-b":
							objs[1].UpdateData = updates[2]
						case "candidate-update-b":
							objs[2].UpdateData = updates[3]
						case "monster-update-b":
							objs[3].UpdateData = updates[2]
						case "caster-player-b":
							pd(0).Player = &infos[2]
						case "candidate-player-b":
							pd(2).Player = &infos[3]
						case "caster-class-zero":
							objs[0].ObjClass = 0
						case "caster-class-player":
							objs[0].ObjClass = object.ClassPlayer
						case "target-class-player":
							objs[1].ObjClass = object.ClassPlayer
						case "target-dead":
							objs[1].ObjFlags |= object.Flags(0x8020)
						case "candidate-status-zero":
							md(3).StatusFlags = 0
						case "candidate-flags-one":
							infos[2].Field3680 = 1
						case "candidate-class-zero":
							objs[2].ObjClass = 0
						case "caster-mana-full":
							pd(0).ManaCur = 100
						case "caster-mana-zero":
							pd(0).ManaCur = 0
						case "target-mana-zero":
							pd(2).ManaCur = 0
						case "obelisk-zero":
							od(1).Mana = 0
						case "obelisk-negative":
							od(1).Mana = -5
						case "maximum-zero":
							pd(0).ManaMax = 0
						case "quest-on":
							quest = 1
						case "quest-off":
							quest = 0
						case "class-wizard":
							infos[0].Info().SetPlayerClass(playerlib.Wizard)
						case "multiplier":
							multipliers[0] = math.Float32bits(1.9999999)
						case "team-change":
							objs[0].TeamVal.ID = 1
							objs[1].TeamVal.ID = 2
						default:
							t.Fatalf("unknown original mutation %q", m.Action)
						}
					}
					rt.Frame = func() uint32 { return frame }
					rt.TickRate = func() uint32 { return fps }
					balance := func(key string) float64 {
						if key != "ManaDrainRange" {
							t.Fatal("balance key")
						}
						enumCaster = r.Caster16
						event("balance", key)
						return plasmaOriginalDouble531580(in.Range)
					}
					levelBalance := func(key string, level uint32) float64 {
						if key != "ManaDrainCoeff" {
							t.Fatal("table key")
						}
						event("balance-level", key, level)
						return plasmaOriginalDouble531580(in.Coefficient)
					}
					rt.Balance = func(key string) float32 { return float32(balance(key)) }
					rt.BalanceLevel = func(key string, level uint32) float32 { return float32(levelBalance(key, level)) }
					if field := runtime.FieldByName("Balance64"); field.IsValid() {
						field.Set(reflect.ValueOf(balance))
					}
					if field := runtime.FieldByName("BalanceLevel64"); field.IsValid() {
						field.Set(reflect.ValueOf(levelBalance))
					}
					rt.ObjectsInCircle = func(p types.Pointf, radius float32, visit func(*Object) bool) {
						v := plasmaOriginalBits531580(p)
						caster := enumCaster
						event("circle", idx(caster), v[0], v[1], math.Float32bits(radius))
						for _, i := range in.Candidates {
							event("visit", i, idx(caster))
							if !visit(obj(i)) {
								break
							}
						}
					}
					rt.IsEnemy = func(a, b *Object) bool {
						i, j := idx(a), idx(b)
						value := in.Enemy[i][j]
						event("enemy", i, j)
						return value != 0
					}
					rt.SameTeam = func(a, b *Object) bool {
						i, j := idx(a), idx(b)
						value := in.SameTeam[i][j+1]
						event("same-team", i, j)
						return value != 0
					}
					rt.TraceRay = func(a, b types.Pointf, flags MapTraceFlags) bool {
						pa, pb := plasmaOriginalBits531580(a), plasmaOriginalBits531580(b)
						event("trace", pa[0], pa[1], pb[0], pb[1], uint32(flags))
						return in.Trace != 0
					}
					rt.PositionDelta = func(o *Object, p *types.Pointf) int32 {
						v := plasmaOriginalBits531580(*p)
						event("position", idx(o), v[0], v[1])
						return int32(in.Moved)
					}
					rt.QuestMode = func() bool { value := quest; event("quest", 4096); return value != 0 }
					rt.QuestManaScale = func(o *Object) float32 {
						cls := uint8(o.UpdateDataPlayer().Player.PlayerClass())
						if cls > 2 {
							return 0
						}
						return math.Float32frombits(multipliers[cls])
					}
					manaService := func(kind string, o *Object, amount int32) {
						event(kind, idx(o), amount)
						if o != nil && uint32(o.ObjClass)&4 != 0 {
							data := o.UpdateDataPlayer()
							delta := int64(amount)
							if kind == "mana-sub" {
								delta = -delta
							}
							value := int64(data.ManaCur) + delta
							value = min(int64(data.ManaMax), max(int64(0), value))
							data.ManaCur = uint16(value)
						}
					}
					rt.AddMana = func(o *Object, amount int16) { manaService("mana-add", o, int32(amount)) }
					if field := runtime.FieldByName("AddMana32"); field.IsValid() {
						field.Set(reflect.ValueOf(func(o *Object, amount int32) { manaService("mana-add", o, amount) }))
					}
					rt.SubMana = func(o *Object, amount int32) { manaService("mana-sub", o, amount) }
					rt.StartRay = func(*DurSpell) { event("start") }
					rt.StopRay = func(_ *DurSpell, o *Object) { event("stop", idx(o)) }
					rt.Audio = func(id uint16, o *Object) { event("audio", id, idx(o)) }
					rt.LoadRayTarget = func(*DurSpell) *Object { return ray }
					rt.StoreRayTarget = func(_ *DurSpell, o *Object) { ray = o }
					var result uint32
					switch in.Kind {
					case "update":
						result = uint32(SpellDrainManaUpdate52E210(r, rt))
					case "transfer":
						if spellDrainManaTransfer52E450(r.Caster16, r.Target48, in.Amount, rt) {
							result = 1
						}
					default:
						t.Fatal("original root")
					}
					got, want := drainOriginalResult52E210{result, snapshot(), events}, row.Result
					if in.Mutation.Stage != "" && !changed {
						t.Fatal("original mutation not reached")
					}
					if !scratch.IsValid() {
						clear := func(s *drainOriginalState52E210) { s.Selected = 0; s.Center = [2]uint32{}; s.Nearest = 0 }
						clear(&got.State)
						clear(&want.State)
						for i := range got.Events {
							clear(&got.Events[i].State)
						}
						for i := range want.Events {
							clear(&want.Events[i].State)
						}
					}
					if want.Ret == 1 && in.Kind == "update" {
						for _, e := range want.Events {
							if e.Kind == "stop" && e.State.Target == -1 {
								if got.State.Ray != -1 {
									t.Fatal("native dying sidecar must be released after no-target stop")
								}
								want.State.Ray = -1
								break
							}
						}
					}
					if !reflect.DeepEqual(got, want) {
						actual, _ := json.Marshal(got)
						expected, _ := json.Marshal(want)
						t.Fatalf("control=%#x got %s\nwant %s", profile.control, actual, expected)
					}
				})
			}
		})
	}
}
