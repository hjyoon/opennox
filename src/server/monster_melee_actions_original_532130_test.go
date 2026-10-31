package server

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type meleeInput532130 map[string]json.RawMessage

func (v meleeInput532130) number(k string) uint32 {
	var n int64
	if raw := v[k]; raw != nil {
		if err := json.Unmarshal(raw, &n); err != nil {
			panic(err)
		}
	}
	return uint32(n)
}
func (v meleeInput532130) text(k string) string {
	var s string
	if raw := v[k]; raw != nil {
		if err := json.Unmarshal(raw, &s); err != nil {
			panic(err)
		}
	}
	return s
}
func (v meleeInput532130) vector(k string) []uint32 {
	var n []uint32
	if err := json.Unmarshal(v[k], &n); err != nil {
		panic(err)
	}
	return n
}

type meleeWord532130 struct {
	Kind   string `json:"kind"`
	Value  uint32 `json:"value"`
	Index  int    `json:"index"`
	Update int    `json:"update"`
}

func (w meleeWord532130) MarshalJSON() ([]byte, error) {
	v := map[string]any{"kind": w.Kind}
	switch w.Kind {
	case "u32", "byte":
		v["value"] = w.Value
	case "candidate", "update", "sound":
		v["index"] = w.Index
	case "slot":
		v["index"], v["update"] = w.Index, w.Update
	}
	return json.Marshal(v)
}

type meleeSlot532130 struct {
	Action  uint32            `json:"action"`
	Args    []meleeWord532130 `json:"args"`
	Started uint32            `json:"started"`
}
type meleeUpdateState532130 struct {
	Head     int8              `json:"head"`
	Reset    []uint32          `json:"reset_fields"`
	Deadline uint32            `json:"deadline"`
	Stamina  uint8             `json:"stamina"`
	Status   uint32            `json:"status"`
	Sound    meleeWord532130   `json:"sound"`
	Slots    []meleeSlot532130 `json:"slots"`
}
type meleePush532130 struct {
	Action   uint32          `json:"action"`
	Live     meleeWord532130 `json:"live_update"`
	Returned meleeWord532130 `json:"returned"`
}
type meleeRow532130 struct {
	Name   string           `json:"name"`
	Entry  uint32           `json:"entry"`
	Family string           `json:"family"`
	Input  meleeInput532130 `json:"input"`
	Output struct {
		Returned   meleeWord532130          `json:"return_value"`
		Class      uint32                   `json:"class"`
		Subclass   uint32                   `json:"subclass"`
		Live       meleeWord532130          `json:"unit_update"`
		Field34    uint32                   `json:"field34"`
		Buffs      uint32                   `json:"buffs"`
		Durations  []uint16                 `json:"durations"`
		Powers     []uint8                  `json:"powers"`
		Updates    []meleeUpdateState532130 `json:"updates"`
		Selection  meleeWord532130          `json:"selection"`
		Distance   uint32                   `json:"distance"`
		Frame      uint32                   `json:"frame"`
		RNG        uint32                   `json:"rng"`
		Changed    uint32                   `json:"stack_changed"`
		PlayerLink meleeWord532130          `json:"player_link"`
		Events     []map[string]any         `json:"events"`
		Pushes     []meleePush532130        `json:"pushes"`
	} `json:"output"`
}
type meleeGuard532130[T any] struct {
	Before [64]byte
	Value  T
	After  [64]byte
}
type meleeRegion532130 struct {
	Pointer unsafe.Pointer
	Size    uintptr
}

func meleeAlloc532130[T comparable](t *testing.T, f *meleeFixture532130) *meleeGuard532130[T] {
	t.Helper()
	p, free := alloc.New(meleeGuard532130[T]{})
	t.Cleanup(free)
	f.regions = append(f.regions, meleeRegion532130{unsafe.Pointer(p), unsafe.Sizeof(*p)})
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(p)) <= math.MaxUint32 {
		t.Fatal("all guarded native records must exceed4GiB")
	}
	return p
}

type meleeFixture532130 struct {
	srv              *Server
	objects          [4]*meleeGuard532130[Object]
	prototypes       [4]Object
	updates          [2]*meleeGuard532130[MonsterUpdateData]
	defs             [2]*meleeGuard532130[MonsterDef]
	sounds           [2]*meleeGuard532130[[17]uint32]
	player           *meleeGuard532130[PlayerUpdateData]
	playerInfo       *meleeGuard532130[Player]
	candidateMonster [3]*meleeGuard532130[MonsterUpdateData]
	candidatePlayer  [3]*meleeGuard532130[PlayerUpdateData]
	candidateInfo    [3]*meleeGuard532130[Player]
	strikeToken      *meleeGuard532130[uint32]
	regions          []meleeRegion532130
	before           [][]byte
	global           []byte
	globalBefore     []byte
	selected         *unsafe.Pointer
	distance         *float32
	row              *meleeRow532130
	events           []map[string]any
	pushes           []meleePush532130
}

func meleeNewFixture532130(t *testing.T) *meleeFixture532130 {
	f := &meleeFixture532130{srv: New(nil, nil, strman.New())}
	t.Cleanup(f.srv.Close)
	if !f.srv.Objs.Init(4) {
		t.Fatal("server object owner allocation")
	}
	t.Cleanup(f.srv.Objs.FreeObjects)
	for i := range f.objects {
		f.prototypes[i] = *f.srv.Objs.NewObject(&ObjectType{})
		f.objects[i] = meleeAlloc532130[Object](t, f)
	}
	for i := range f.updates {
		f.updates[i] = meleeAlloc532130[MonsterUpdateData](t, f)
		f.defs[i] = meleeAlloc532130[MonsterDef](t, f)
		f.sounds[i] = meleeAlloc532130[[17]uint32](t, f)
	}
	f.player = meleeAlloc532130[PlayerUpdateData](t, f)
	f.playerInfo = meleeAlloc532130[Player](t, f)
	for i := range f.candidateMonster {
		f.candidateMonster[i] = meleeAlloc532130[MonsterUpdateData](t, f)
		f.candidatePlayer[i] = meleeAlloc532130[PlayerUpdateData](t, f)
		f.candidateInfo[i] = meleeAlloc532130[Player](t, f)
	}
	f.strikeToken = meleeAlloc532130[uint32](t, f)
	f.selected = memmap.PtrPtr(0x5d4594, 2487948)
	f.distance = memmap.PtrFloat32(0x5d4594, 2487952)
	f.global = unsafe.Slice((*byte)(unsafe.Add(unsafe.Pointer(f.distance), -68)), 136)
	old, oldGlobal := *f.selected, append([]byte(nil), f.global...)
	t.Cleanup(func() { copy(f.global, oldGlobal); *f.selected = old })
	return f
}
func (f *meleeFixture532130) unit() *Object { return &f.objects[0].Value }
func (f *meleeFixture532130) word(w meleeWord532130) uintptr {
	switch w.Kind {
	case "nil":
		return 0
	case "u32", "byte":
		return uintptr(w.Value)
	case "unit":
		return uintptr(unsafe.Pointer(f.unit()))
	case "player":
		return uintptr(unsafe.Pointer(&f.player.Value))
	case "candidate":
		return uintptr(unsafe.Pointer(&f.objects[w.Index+1].Value))
	case "update":
		return uintptr(unsafe.Pointer(&f.updates[w.Index].Value))
	case "sound":
		return uintptr(unsafe.Pointer(&f.sounds[w.Index].Value[0]))
	case "slot":
		return uintptr(unsafe.Pointer(&f.updates[w.Update].Value.AIStack[w.Index]))
	default:
		panic("unmapped original role " + w.Kind)
	}
}
func (f *meleeFixture532130) role(p uintptr) meleeWord532130 {
	if p == 0 {
		return meleeWord532130{Kind: "nil"}
	}
	if p == uintptr(unsafe.Pointer(f.unit())) {
		return meleeWord532130{Kind: "unit"}
	}
	if p == uintptr(unsafe.Pointer(&f.player.Value)) {
		return meleeWord532130{Kind: "player"}
	}
	for i, pointer := range f.objects[1:] {
		if p == uintptr(unsafe.Pointer(&pointer.Value)) {
			return meleeWord532130{Kind: "candidate", Index: i}
		}
	}
	for i, pointer := range f.updates {
		if p == uintptr(unsafe.Pointer(&pointer.Value)) {
			return meleeWord532130{Kind: "update", Index: i}
		}
		for j := range pointer.Value.AIStack {
			if p == uintptr(unsafe.Pointer(&pointer.Value.AIStack[j])) {
				return meleeWord532130{Kind: "slot", Update: i, Index: j}
			}
		}
	}
	for i, pointer := range f.sounds {
		if p == uintptr(unsafe.Pointer(&pointer.Value[0])) {
			return meleeWord532130{Kind: "sound", Index: i}
		}
	}
	if p > math.MaxUint32 {
		panic(fmt.Sprintf("unmapped native pointer%#x", p))
	}
	return meleeWord532130{Kind: "u32", Value: uint32(p)}
}
func (f *meleeFixture532130) event(values map[string]any) {
	raw, err := json.Marshal(values)
	if err != nil {
		panic(err)
	}
	var normalized map[string]any
	if err := json.Unmarshal(raw, &normalized); err != nil {
		panic(err)
	}
	f.events = append(f.events, normalized)
}
func (f *meleeFixture532130) mutate(effect string, cached *MonsterUpdateData) {
	u := f.unit()
	switch effect {
	case "none":
	case "swap-update":
		u.UpdateData = unsafe.Pointer(&f.updates[1].Value)
	case "swap-update-and-clear-class":
		u.UpdateData = unsafe.Pointer(&f.updates[1].Value)
		u.ObjClass = 0
	case "clear-class":
		u.ObjClass = 0
	case "class-player":
		u.ObjClass = object.ClassPlayer
	case "replace-sound":
		cached.SoundSet122 = unsafe.Pointer(&f.sounds[1].Value[0])
	case "null-sound":
		cached.SoundSet122 = nil
	case "clear-bot":
		cached.StatusFlags &^= object.MonsterStatus(0x20000)
	case "set-bot":
		cached.StatusFlags |= object.MonsterStatus(0x20000)
	case "advance-frame":
		f.srv.SetFrame(f.srv.Frame() + 3)
	case "change-player-link":
		f.player.Value.Field73 = &f.updates[1].Value
	case "complete-cached":
		cached.Field120_3 = 0x81
	case "cap-cached-head":
		cached.AIStackInd = 22
	default:
		panic("unknown controlled effect " + effect)
	}
}
func (f *meleeFixture532130) callback(name string) func(*Object) {
	return func(u *Object) {
		if u != f.unit() {
			panic("callback native unit changed")
		}
		cached := (*MonsterUpdateData)(u.UpdateData)
		effect := f.row.Input.text(name + "_effect")
		f.event(map[string]any{"service": name, "class_before": uint32(u.ObjClass), "live_update_before": f.role(uintptr(u.UpdateData)), "effect": effect})
		f.mutate(effect, cached)
	}
}
func (f *meleeFixture532130) runtime() MonsterMeleeNativeRuntime532130 {
	audio := func(id uint32, u *Object) {
		if u != f.unit() {
			panic("audio native unit changed")
		}
		f.event(map[string]any{"service": "audio", "sound": id, "live_update": f.role(uintptr(u.UpdateData))})
	}
	return MonsterMeleeNativeRuntime532130{
		AudioEvent: audio,
		BuffOff: func(u *Object, enchant EnchantID) {
			if u != f.unit() {
				panic("buff native unit changed")
			}
			e := int32(enchant)
			SpellBuffOff4FF5B0(SpellBuffOffHooks4FF5B0[*Object]{
				LoadBuffArg: func() int32 { return e }, LoadUnitArg: func() *Object { return u }, LoadBuffs: func(v *Object) uint32 { return v.Buffs },
				SetBuffFlags: func(v *Object, flags uint32) {
					v.Buffs = flags
					f.event(map[string]any{"service": "publish-buffs", "flags": flags})
					count := 0
					for _, event := range f.events {
						if event["service"] == "publish-buffs" {
							count++
						}
					}
					if count == 1 {
						f.mutate(f.row.Input.text("buff_effect"), &f.updates[0].Value)
					}
				},
				StoreDuration: func(v *Object, e int32, n uint16) { v.BuffsDur[e] = n }, StorePower: func(v *Object, e int32, n uint8) { v.BuffsPower[e] = n },
				EnchantSpell: func(e int32) int32 {
					if e == 0 {
						return 10
					}
					if e == 23 {
						return 11
					}
					panic("undeclared enchant resource")
				},
				SpellAudio: func(spell, field int32) int32 {
					if field != 2 {
						panic("audio resource field")
					}
					if spell == 10 {
						return 900
					}
					if spell == 11 {
						return 923
					}
					panic("undeclared spell resource")
				},
				Audio: func(id int32, v *Object, kind, code int32) {
					if kind != 0 || code != 0 {
						panic("audio service flags")
					}
					audio(uint32(id), v)
				},
			})
		},
		PlayerAttack: func(u *Object) int {
			if u != f.unit() {
				panic("player attack unit")
			}
			effect := f.row.Input.text("player-attack_effect")
			f.event(map[string]any{"service": "player-attack", "class_before": uint32(u.ObjClass), "live_update_before": f.role(uintptr(u.UpdateData)), "effect": effect})
			f.mutate(effect, &f.updates[0].Value)
			return int(int32(f.row.Input.number("attack_return")))
		},
		CanStrike: func(fn unsafe.Pointer) bool { return fn == unsafe.Pointer(&f.strikeToken.Value) },
		Strike: func(u *Object, fn unsafe.Pointer) int {
			if u != f.unit() || fn != unsafe.Pointer(&f.strikeToken.Value) {
				panic("strike native arguments")
			}
			effect := f.row.Input.text("strike_effect")
			f.event(map[string]any{"service": "strike", "class_before": uint32(u.ObjClass), "live_update_before": f.role(uintptr(u.UpdateData)), "effect": effect})
			f.mutate(effect, &f.updates[0].Value)
			return int(int32(f.row.Input.number("strike_return")))
		},
		EachInRect: func(rect types.Rectf, fn func(*Object) bool) {
			f.event(map[string]any{"service": "world", "rect": []uint32{math.Float32bits(rect.Min.X), math.Float32bits(rect.Min.Y), math.Float32bits(rect.Max.X), math.Float32bits(rect.Max.Y)}})
			f.mutate(f.row.Input.text("world_effect"), &f.updates[0].Value)
			orders := map[string][]int{"empty": {}, "friend": {0}, "enemy": {1}, "friend-then-enemy": {0, 1}, "enemy-then-friend": {1, 0}, "dead-front": {0, 1}, "behind": {0}, "zero-vector": {0}, "angle-boundary": {0}}
			order, ok := orders[f.row.Input.text("world")]
			if !ok {
				panic("unknown original world input")
			}
			for _, i := range order {
				if !fn(&f.objects[i+1].Value) {
					break
				}
			}
		},
		IsEnemy: func(u, v *Object) bool {
			got := f.srv.IsEnemyTo(u, v)
			n := uint32(0)
			if got {
				n = 1
			}
			f.event(map[string]any{"service": "original-enemy", "unit": f.role(uintptr(unsafe.Pointer(u))), "candidate": f.role(uintptr(unsafe.Pointer(v))), "result": n})
			return got
		},
		CancelAction: func(ai.ActionType) func(*Object) {
			if f.row.Input.number("callback_present") == 0 {
				return nil
			}
			return f.callback("cancel")
		},
		EndAction: func(ai.ActionType) func(*Object) {
			if f.row.Input.number("callback_present") == 0 {
				return nil
			}
			return f.callback("end")
		},
		ObservePush: func(action ai.ActionType, cached *MonsterUpdateData, result *AIStackItem) {
			f.pushes = append(f.pushes, meleePush532130{Action: uint32(action), Live: f.role(uintptr(unsafe.Pointer(cached))), Returned: f.role(uintptr(unsafe.Pointer(result)))})
		},
	}
}

// The accessor is resolved only in the post-implementation run. Before-port
// controls validate the allocations without installing behavior or calling C.
type meleeWeaponSetter532130 interface {
	SetNPCWeaponForMeleeUpdate532130(*MonsterUpdateData, *Object)
}

func (f *meleeFixture532130) reset(row *meleeRow532130, control bool) {
	f.row = row
	f.events = nil
	f.pushes = nil
	c := row.Input
	for _, r := range f.regions {
		for i := range unsafe.Slice((*byte)(r.Pointer), int(r.Size)) {
			unsafe.Slice((*byte)(r.Pointer), int(r.Size))[i] = byte(i*31 + 19)
		}
	}
	for i, g := range f.objects {
		g.Value = f.prototypes[i]
		g.Value.TypeInd = 117
		g.Value.NetCode = uint32(118 + i)
		g.Value.ObjClass = object.ClassMonster
		g.Value.ObjSubClass = 0
		g.Value.ObjFlags = 0
		g.Value.TeamVal = ObjectTeam{}
	}
	u := f.unit()
	u.ObjClass = object.Class(c.number("cls"))
	u.ObjSubClass = object.SubClass(c.number("subclass"))
	p := c.vector("unit_pos")
	u.PosVec = types.Pointf{X: math.Float32frombits(p[0]), Y: math.Float32frombits(p[1])}
	u.Direction1 = Dir16(c.number("direction"))
	u.Field34 = 97
	u.Buffs = c.number("buffs")
	u.BuffsDur = [32]uint16{}
	u.BuffsPower = [32]uint8{}
	for _, e := range []int{0, 23} {
		u.BuffsDur[e] = uint16(0x5511 + e)
		u.BuffsPower[e] = byte(0x41 + e)
	}
	u.UpdateData = unsafe.Pointer(&f.updates[0].Value)
	for i, g := range f.updates {
		d := &g.Value
		*d = MonsterUpdateData{}
		m := &f.defs[i].Value
		*m = MonsterDef{}
		d.MonsterDef = m
		if c.number("sound") != 0 {
			d.SoundSet122 = unsafe.Pointer(&f.sounds[i].Value[0])
		}
		d.Field128 = c.number("deadline")
		d.StatusFlags = object.MonsterStatus(c.number("status"))
		d.Stamina = byte(c.number("stamina"))
		if i == 1 {
			d.Field128 = 701
			d.StatusFlags = object.MonsterStatus(c.number("alt_status"))
			d.Stamina = 73
		}
		d.WeaponEquipFlags = c.number("weapons")
		d.Field545 = &f.player.Value
		d.CurrentEnemy = &f.objects[1].Value
		repeated := uint32(0x71717171) + uint32(i)*0x01010101
		d.Field2 = repeated
		d.Field67 = repeated
		d.Field74 = repeated
		d.Field91 = repeated
		d.Field124 = repeated
		d.Field137 = repeated
		d.Field120_1 = byte(c.number("animation"))
		d.Field120_2 = byte(c.number("strike_done"))
		d.Field120_3 = byte(c.number("complete"))
		if i == 1 {
			d.Field120_3 = byte(c.number("alt_complete"))
		}
		d.AIStackInd = int8(c.number("head"))
		for j := range d.AIStack {
			d.AIStack[j].Action = 1
			for k := range d.AIStack[j].Args {
				d.AIStack[j].Args[k] = uintptr(0x61000000 + uint32(i)*0x10000 + uint32(j)*0x100 + uint32(k)*0x10 + 19)
			}
		}
		if d.AIStackInd >= 0 {
			d.AIStack[d.AIStackInd].Action = c.number("action")
			d.AIStack[d.AIStackInd].Field5 = c.number("started")
			if d.AIStackInd > 0 {
				d.AIStack[0].Action = c.number("root_action")
			}
			if c.number("condition_below") != 0 {
				for j := 1; j < int(d.AIStackInd); j++ {
					d.AIStack[j].Action = 41
				}
			}
		}
		m.MeleeAttackFrame108 = c.number("strike_frame")
		m.MeleeAttackRange112 = math.Float32frombits(c.number("range"))
		m.MeleeAttackDelayMin128 = c.number("delay_min")
		m.MeleeAttackDelayMax132 = c.number("delay_max")
		if c.number("strike_present") != 0 {
			m.MeleeStrikeFunc236 = unsafe.Pointer(&f.strikeToken.Value)
		}
		for j := range f.sounds[i].Value {
			f.sounds[i].Value[j] = uint32(600 + 100*i + j)
		}
	}
	f.player.Value = PlayerUpdateData{}
	f.playerInfo.Value = Player{}
	f.player.Value.Player = &f.playerInfo.Value
	f.player.Value.Field73 = &f.updates[0].Value
	for i := range f.candidateMonster {
		f.candidateMonster[i].Value = MonsterUpdateData{}
		f.candidatePlayer[i].Value = PlayerUpdateData{}
		f.candidateInfo[i].Value = Player{}
		v := &f.objects[i+1].Value
		v.PosVec = types.Ptf(float32(330+10*i), 300)
		v.UpdateData = unsafe.Pointer(&f.candidateMonster[i].Value)
		if i == 1 {
			v.ObjClass = object.ClassPlayer
			v.UpdateData = unsafe.Pointer(&f.candidatePlayer[i].Value)
			f.candidatePlayer[i].Value.Player = &f.candidateInfo[i].Value
		}
	}
	f.objects[3].Value.ObjSubClass = object.SubClass(0x4000)
	switch c.text("world") {
	case "dead-front":
		f.objects[1].Value.ObjFlags = object.FlagDead
	case "behind":
		f.objects[1].Value.PosVec.X = 310
	case "zero-vector":
		f.objects[1].Value.PosVec.X = 320
	case "angle-boundary":
		f.objects[1].Value.PosVec = types.Ptf(321, 301.7320508)
	}
	if !control && f.nativeAvailable() {
		setter, ok := any(u).(meleeWeaponSetter532130)
		if !ok {
			panic("native cached-weapon field accessor not implemented")
		}
		for _, g := range f.updates {
			var weapon *Object
			if c.number("hammer") != 0 {
				weapon = &f.objects[3].Value
			}
			setter.SetNPCWeaponForMeleeUpdate532130(&g.Value, weapon)
		}
	}
	f.srv.SetFrame(c.number("frame"))
	f.srv.SetTickRate(c.number("fps"))
	f.srv.Rand.Logic.Reset(int(c.number("rng")))
	f.srv.AI.StackChanged = c.number("initial_changed") != 0
	for i := range f.global {
		f.global[i] = 0x6b
	}
	*f.selected = unsafe.Pointer(&f.objects[3].Value)
	*f.distance = 17
	f.globalBefore = append([]byte(nil), f.global...)
	f.before = nil
	for _, r := range f.regions {
		f.before = append(f.before, append([]byte(nil), unsafe.Slice((*byte)(r.Pointer), int(r.Size))...))
	}
}

func (f *meleeFixture532130) expectBytes(expected [][]byte, p unsafe.Pointer, raw []byte) {
	addr := uintptr(p)
	for i, r := range f.regions {
		start := uintptr(r.Pointer)
		if start <= addr && addr+uintptr(len(raw)) <= start+r.Size {
			copy(expected[i][addr-start:], raw)
			return
		}
	}
	panic("expected output is not in a guarded native region")
}
func (f *meleeFixture532130) expect32(expected [][]byte, p unsafe.Pointer, n uint32) {
	var raw [4]byte
	binary.LittleEndian.PutUint32(raw[:], n)
	f.expectBytes(expected, p, raw[:])
}
func (f *meleeFixture532130) expect8(expected [][]byte, p unsafe.Pointer, n byte) {
	f.expectBytes(expected, p, []byte{n})
}
func (f *meleeFixture532130) expectWord(expected [][]byte, p unsafe.Pointer, n uintptr) {
	var raw [8]byte
	if unsafe.Sizeof(uintptr(0)) == 8 {
		binary.LittleEndian.PutUint64(raw[:], uint64(n))
		f.expectBytes(expected, p, raw[:])
	} else {
		binary.LittleEndian.PutUint32(raw[:], uint32(n))
		f.expectBytes(expected, p, raw[:4])
	}
}
func (f *meleeFixture532130) validate(t *testing.T, result uintptr, control bool) {
	t.Helper()
	expected := make([][]byte, len(f.before))
	for i, before := range f.before {
		expected[i] = append([]byte(nil), before...)
	}
	w := &f.row.Output
	u := f.unit()
	if !control {
		if len(w.Updates) != 2 || len(w.Durations) != 2 || len(w.Powers) != 2 {
			t.Fatal("original output dimensions")
		}
		f.expect32(expected, unsafe.Pointer(&u.ObjClass), w.Class)
		f.expect32(expected, unsafe.Pointer(&u.ObjSubClass), w.Subclass)
		f.expectWord(expected, unsafe.Pointer(&u.UpdateData), f.word(w.Live))
		f.expect32(expected, unsafe.Pointer(&u.Field34), w.Field34)
		f.expect32(expected, unsafe.Pointer(&u.Buffs), w.Buffs)
		for i, e := range []int{0, 23} {
			var raw [2]byte
			binary.LittleEndian.PutUint16(raw[:], w.Durations[i])
			f.expectBytes(expected, unsafe.Pointer(&u.BuffsDur[e]), raw[:])
			f.expect8(expected, unsafe.Pointer(&u.BuffsPower[e]), w.Powers[i])
		}
		for i, state := range w.Updates {
			d := &f.updates[i].Value
			if len(state.Reset) != 9 || len(state.Slots) != 24 {
				t.Fatal("original stack/reset dimensions")
			}
			f.expect8(expected, unsafe.Pointer(&d.AIStackInd), byte(state.Head))
			for j, p := range []*uint32{&d.Field2, &d.Field67, &d.Field74, &d.Field91} {
				f.expect32(expected, unsafe.Pointer(p), state.Reset[j])
			}
			for j, p := range []*uint8{&d.Field120_1, &d.Field120_2, &d.Field120_3} {
				f.expect8(expected, unsafe.Pointer(p), byte(state.Reset[4+j]))
			}
			f.expect32(expected, unsafe.Pointer(&d.Field124), state.Reset[7])
			f.expect32(expected, unsafe.Pointer(&d.Field137), state.Reset[8])
			f.expect32(expected, unsafe.Pointer(&d.Field128), state.Deadline)
			f.expect8(expected, unsafe.Pointer(&d.Stamina), state.Stamina)
			f.expect32(expected, unsafe.Pointer(&d.StatusFlags), state.Status)
			f.expectWord(expected, unsafe.Pointer(&d.SoundSet122), f.word(state.Sound))
			for j, slot := range state.Slots {
				if len(slot.Args) != 4 {
					t.Fatal("original stack argument dimensions")
				}
				item := &d.AIStack[j]
				f.expect32(expected, unsafe.Pointer(&item.Action), slot.Action)
				f.expect32(expected, unsafe.Pointer(&item.Field5), slot.Started)
				for k, arg := range slot.Args {
					f.expectWord(expected, unsafe.Pointer(&item.Args[k]), f.word(arg))
				}
			}
		}
		f.expectWord(expected, unsafe.Pointer(&f.player.Value.Field73), f.word(w.PlayerLink))
		want := f.word(w.Returned)
		if f.row.Entry == 0x532440 {
			result &= 255
		}
		if f.nativeAvailable() && result != want {
			t.Errorf("native return=%#x; original role=%+v=%#x", result, w.Returned, want)
		}
		if f.srv.Frame() != w.Frame || f.srv.Rand.Logic.Index() != int(w.RNG) || f.srv.AI.StackChanged != (w.Changed != 0) {
			t.Errorf("native frame/RNG/changed=%d/%d/%v; original=%d/%d/%d", f.srv.Frame(), f.srv.Rand.Logic.Index(), f.srv.AI.StackChanged, w.Frame, w.RNG, w.Changed)
		}
		if !reflect.DeepEqual(f.events, w.Events) && !(len(f.events) == 0 && len(w.Events) == 0) {
			t.Errorf("native service events=%v; original=%v", f.events, w.Events)
		}
		if !reflect.DeepEqual(f.pushes, w.Pushes) && !(len(f.pushes) == 0 && len(w.Pushes) == 0) {
			t.Errorf("native real push attempts=%v; original=%v", f.pushes, w.Pushes)
		}
	}
	for i, r := range f.regions {
		after := unsafe.Slice((*byte)(r.Pointer), int(r.Size))
		if !bytes.Equal(after, expected[i]) {
			count, first := 0, -1
			for j := range after {
				if after[j] != expected[i][j] {
					count++
					if first < 0 {
						first = j
					}
				}
			}
			t.Errorf("guarded native region%d has%d differing bytes; first+%d", i, count, first)
		}
	}
	wantGlobal := append([]byte(nil), f.globalBefore...)
	wantSelected := uintptr(unsafe.Pointer(&f.objects[3].Value))
	wantDistance := math.Float32bits(17)
	if !control {
		wantSelected = f.word(w.Selection)
		wantDistance = w.Distance
		binary.LittleEndian.PutUint32(wantGlobal[68:], wantDistance)
		if unsafe.Sizeof(uintptr(0)) == 4 {
			binary.LittleEndian.PutUint32(wantGlobal[64:], uint32(wantSelected))
		}
	}
	if uintptr(*f.selected) != wantSelected || math.Float32bits(*f.distance) != wantDistance || !bytes.Equal(f.global, wantGlobal) {
		t.Errorf("native selected/global/distance mismatch; selected=%p want=%#x distance=%08x want=%08x", *f.selected, wantSelected, math.Float32bits(*f.distance), wantDistance)
	}
}

type meleeNativeExecutor532130 interface {
	MonsterMeleeNative532130(uint32, *Object, MonsterMeleeBlockerState532390, MonsterMeleeNativeRuntime532130) uintptr
}

func (f *meleeFixture532130) nativeAvailable() bool {
	_, ok := any(f.srv).(meleeNativeExecutor532130)
	return ok
}

type meleeControlledAction532130 struct {
	kind    ai.ActionType
	fixture *meleeFixture532130
}

func (a *meleeControlledAction532130) Type() ai.ActionType { return a.kind }
func (a *meleeControlledAction532130) Start(*Object)       {}
func (a *meleeControlledAction532130) Update(*Object)      {}
func (a *meleeControlledAction532130) End(u *Object)       { a.fixture.callback("end")(u) }
func (a *meleeControlledAction532130) Cancel(u *Object)    { a.fixture.callback("cancel")(u) }
func (f *meleeFixture532130) executeExistingOrNative() uintptr {
	runtime := f.runtime()
	u := f.unit()
	if native, ok := any(f.srv).(meleeNativeExecutor532130); ok {
		return native.MonsterMeleeNative532130(f.row.Entry, u, MonsterMeleeBlockerState532390{Selected: f.selected, Distance: f.distance}, runtime)
	}
	old := aiActions
	aiActions = make(map[ai.ActionType]AIAction)
	if f.row.Input.number("callback_present") != 0 {
		for _, id := range []ai.ActionType{0, 1, 4, 24, 27, 31, 41, 50, 60, 0xffffffff} {
			aiActions[id] = &meleeControlledAction532130{kind: id, fixture: f}
		}
	}
	defer func() { aiActions = old }()
	if f.row.Entry == 0x532130 {
		monsterActionMeleeStart532130(u, monsterActionMeleeStartHooks532130{
			frame: f.srv.Frame, tickRate: f.srv.TickRate, random: f.srv.Rand.Logic.IntClamp,
			audio: runtime.AudioEvent, buffOff: runtime.BuffOff, weaponStamina: WeaponStaminaByType4F7E80,
			eachInRect: runtime.EachInRect, isEnemy: runtime.IsEnemy, pop: u.MonsterPopAction,
			push: func(action ai.ActionType, args ...any) *AIStackItem {
				cached := (*MonsterUpdateData)(u.UpdateData)
				result := u.monsterPushActionWithCancel50A260(action, "go", 0, runtime.CancelAction)
				result.SetArgs(args...)
				runtime.ObservePush(action, cached, result)
				return result
			},
		})
	} else {
		monsterActionMeleeUpdate532440(u, monsterActionMeleeUpdateHooks532440{
			audio: runtime.AudioEvent, pop: u.MonsterPopAction, playerAttack: runtime.PlayerAttack,
			canStrike: runtime.CanStrike, strike: runtime.Strike,
		})
	}
	return 0 // Existing Go's handled-bool API exposes no original word.
}
func TestMeleeActionsOriginalGo532130(t *testing.T) {
	path := os.Getenv("NOX_MELEE_ACTIONS_ORIGINAL_REFERENCE")
	if path == "" {
		path = "testdata/monster_melee_actions_original_532130.json.gz"
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(raw)
	if hex.EncodeToString(digest[:]) != "23aaab9349edc5b10cce887ed21c9564dee5b8df1f2bcf9ef15d4831b2dd30b3" {
		t.Fatal("unsealed original melee actions reference")
	}
	z, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	var ref struct {
		Schema   string           `json:"schema"`
		Original string           `json:"original_sha256"`
		Control  uint16           `json:"control_word"`
		Rows     []meleeRow532130 `json:"rows"`
	}
	if err := json.NewDecoder(z).Decode(&ref); err != nil {
		t.Fatal(err)
	}
	if ref.Schema != "original-monster-melee-actions-532130-v2" || ref.Original != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || ref.Control != 0xe3f || len(ref.Rows) != 3664 {
		t.Fatal("original provenance/control/dimensions")
	}

	f := meleeNewFixture532130(t)
	engine, game := noxflags.GetEngine(), noxflags.GetGame()
	noxflags.UnsetEngine(noxflags.EngineShowAI)
	noxflags.UnsetGame(game)
	t.Cleanup(func() {
		noxflags.UnsetEngine(noxflags.GetEngine())
		noxflags.SetEngine(engine)
		noxflags.UnsetGame(noxflags.GetGame())
		noxflags.SetGame(game)
	})
	control := os.Getenv("NOX_MELEE_ACTIONS_CONTROL_ONLY") == "true"
	entries := map[uint32]int{}
	for i := range ref.Rows {
		row := &ref.Rows[i]
		entries[row.Entry]++
		t.Run(row.Name, func(t *testing.T) {
			f.reset(row, control)
			result := uintptr(0)
			defer func() {
				if v := recover(); v != nil {
					t.Errorf("Go panic for original-defined context: %v", v)
				}
				f.validate(t, result, control)
			}()
			if !control {
				result = f.executeExistingOrNative()
			}
		})
	}
	if !reflect.DeepEqual(entries, map[uint32]int{0x532130: 2255, 0x532440: 1409}) {
		t.Fatal("original per-entry dimensions")
	}
	t.Logf("original Go melee-action contexts=%d control-only=%v original-word-exposed=%v", len(ref.Rows), control, f.nativeAvailable())
}
