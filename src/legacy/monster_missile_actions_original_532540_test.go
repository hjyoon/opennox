package legacy

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
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

type missileWord532540 struct {
	Kind   string `json:"kind"`
	Value  uint32 `json:"value"`
	Index  int    `json:"index"`
	Update int    `json:"update"`
}

func (w missileWord532540) MarshalJSON() ([]byte, error) {
	v := map[string]any{"kind": w.Kind}
	switch w.Kind {
	case "u32", "byte":
		v["value"] = w.Value
	case "update", "definition", "sound", "projectile":
		v["index"] = w.Index
	case "slot":
		v["index"], v["update"] = w.Index, w.Update
	}
	return json.Marshal(v)
}

type missileSlot532540 struct {
	Action  uint32              `json:"action"`
	Args    []missileWord532540 `json:"args"`
	Started uint32              `json:"started"`
}
type missileUpdate532540 struct {
	Head       int8                `json:"head"`
	Reset      []uint32            `json:"reset_fields"`
	Definition missileWord532540   `json:"definition"`
	Sound      missileWord532540   `json:"sound"`
	Deadline   uint32              `json:"deadline"`
	Slots      []missileSlot532540 `json:"slots"`
}
type missileProjectile532540 struct {
	Position   []uint32 `json:"position_bits"`
	Velocity   []uint32 `json:"velocity_bits"`
	Directions []int16  `json:"directions"`
	Speed      uint32   `json:"speed"`
}
type missileRow532540 struct {
	Name   string           `json:"name"`
	Entry  uint32           `json:"entry"`
	Family string           `json:"family"`
	Input  meleeInput532130 `json:"input"`
	Output struct {
		Returned     missileWord532540         `json:"return_value"`
		Class        uint32                    `json:"cls"`
		Subclass     uint32                    `json:"subclass"`
		Live         missileWord532540         `json:"unit_update"`
		Field34      uint32                    `json:"field34"`
		UnitPosition []uint32                  `json:"unit_position_bits"`
		Direction    int16                     `json:"direction"`
		Buffs        uint32                    `json:"buffs"`
		Durations    []uint16                  `json:"durations"`
		Powers       []uint8                   `json:"powers"`
		Updates      []missileUpdate532540     `json:"updates"`
		Projectiles  []missileProjectile532540 `json:"projectiles"`
		Frame        uint32                    `json:"frame"`
		RNG          uint32                    `json:"rng"`
		Changed      uint32                    `json:"stack_changed"`
		Events       []map[string]any          `json:"events"`
		Pushes       []meleePush532130         `json:"pushes"`
	} `json:"output"`
}

// Reuse the byte-identical, already qualified guarded allocator and byte
// expectation helpers. All missile setup, services and observations are below.
type missileFixture532540 struct {
	*meleeFixture532130
	row *missileRow532540
}

func (f *missileFixture532540) projectile(i int) *server.Object { return &f.objects[i+2].Value }
func (f *missileFixture532540) word(w missileWord532540) uintptr {
	switch w.Kind {
	case "nil":
		return 0
	case "u32", "byte":
		return uintptr(w.Value)
	case "unit":
		return uintptr(unsafe.Pointer(f.unit()))
	case "target":
		return uintptr(unsafe.Pointer(&f.objects[1].Value))
	case "projectile":
		return uintptr(unsafe.Pointer(f.projectile(w.Index)))
	case "update":
		return uintptr(unsafe.Pointer(&f.updates[w.Index].Value))
	case "definition":
		return uintptr(unsafe.Pointer(&f.defs[w.Index].Value))
	case "sound":
		return uintptr(unsafe.Pointer(&f.sounds[w.Index].Value[0]))
	case "slot":
		return uintptr(unsafe.Pointer(&f.updates[w.Update].Value.AIStack[w.Index]))
	}
	panic("unknown original missile word " + w.Kind)
}
func (f *missileFixture532540) role(p uintptr) missileWord532540 {
	for _, w := range []missileWord532540{{Kind: "nil"}, {Kind: "unit"}, {Kind: "target"}, {Kind: "projectile", Index: 0}, {Kind: "projectile", Index: 1}, {Kind: "update", Index: 0}, {Kind: "update", Index: 1}, {Kind: "definition", Index: 0}, {Kind: "definition", Index: 1}, {Kind: "sound", Index: 0}, {Kind: "sound", Index: 1}} {
		if f.word(w) == p {
			return w
		}
	}
	for i := range f.updates {
		for j := range f.updates[i].Value.AIStack {
			w := missileWord532540{Kind: "slot", Update: i, Index: j}
			if f.word(w) == p {
				return w
			}
		}
	}
	if p > math.MaxUint32 {
		panic(fmt.Sprintf("unmapped native missile pointer%#x", p))
	}
	return missileWord532540{Kind: "u32", Value: uint32(p)}
}
func (f *missileFixture532540) reset(row *missileRow532540) {
	f.row = row
	f.events = nil
	f.pushes = nil
	c := row.Input
	for _, r := range f.regions {
		b := unsafe.Slice((*byte)(r.Pointer), int(r.Size))
		for i := range b {
			b[i] = byte(i*31 + 19)
		}
	}
	for i, g := range f.objects {
		g.Value = f.prototypes[i]
		g.Value.ObjClass = 0
		g.Value.ObjSubClass = 0
		g.Value.ObjFlags = 0
		g.Value.PosVec = types.Pointf{}
		g.Value.VelVec = types.Pointf{}
		g.Value.UpdateData = nil
		g.Value.Direction1 = 0
		g.Value.Direction2 = 0
	}
	u := f.unit()
	u.TypeInd = 117
	u.NetCode = 118
	u.ObjClass = object.Class(c.number("cls"))
	u.ObjSubClass = object.SubClass(c.number("subclass"))
	pos := c.vector("unit_pos")
	u.PosVec = types.Pointf{X: math.Float32frombits(pos[0]), Y: math.Float32frombits(pos[1])}
	u.Direction1 = server.Dir16(c.number("direction"))
	u.Field34 = 97
	u.Shape.Circle.R = math.Float32frombits(c.number("radius"))
	u.Buffs = c.number("buffs")
	u.BuffsDur = [32]uint16{}
	u.BuffsPower = [32]uint8{}
	for _, e := range []int{0, 23} {
		u.BuffsDur[e] = uint16(0x5511 + e)
		u.BuffsPower[e] = byte(0x41 + e)
	}
	u.UpdateData = unsafe.Pointer(&f.updates[0].Value)
	target := &f.objects[1].Value
	pos = c.vector("target_pos")
	target.PosVec = types.Pointf{X: math.Float32frombits(pos[0]), Y: math.Float32frombits(pos[1])}
	pos = c.vector("target_velocity")
	target.VelVec = types.Pointf{X: math.Float32frombits(pos[0]), Y: math.Float32frombits(pos[1])}
	for i, g := range f.updates {
		d := &g.Value
		*d = server.MonsterUpdateData{}
		m := &f.defs[i].Value
		*m = server.MonsterDef{}
		d.MonsterDef = m
		if c.number("sound") != 0 {
			d.SoundSet122 = unsafe.Pointer(&f.sounds[i].Value[0])
		}
		d.Field128 = c.number("deadline")
		if i == 1 {
			d.Field128 = 701
		}
		d.CurrentEnemy = target
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
			item := &d.AIStack[j]
			item.Action = 1
			for k := range item.Args {
				item.Args[k] = uintptr(0x61000000 + uint32(i)*0x10000 + uint32(j)*0x100 + uint32(k)*0x10 + 19)
			}
			item.Field5 = 0
		}
		if d.AIStackInd >= 0 {
			item := &d.AIStack[d.AIStackInd]
			item.Action = c.number("action")
			item.Field5 = c.number("started")
			pos = c.vector("head_pos")
			item.Args[0] = uintptr(pos[0])
			item.Args[1] = uintptr(pos[1])
			item.Args[2] = 0
			if c.number("targeted") != 0 {
				item.Args[2] = uintptr(unsafe.Pointer(target))
			}
			if d.AIStackInd > 0 {
				d.AIStack[0].Action = c.number("root_action")
			}
			if c.number("condition_below") != 0 {
				for j := 1; j < int(d.AIStackInd); j++ {
					d.AIStack[j].Action = 41
				}
			}
		}
		copy(m.MissileName148[:], []byte("FixtureMissileA"))
		if i == 1 {
			copy(m.MissileName148[:], []byte("FixtureMissileB"))
		}
		m.MissileAttackRange212 = math.Float32frombits(c.number("range"))
		m.MissileAttackFrame216 = c.number("strike_frame")
		m.MissileAttackDelayMin220 = c.number("delay_min")
		m.MissileAttackDelayMax224 = c.number("delay_max")
		if i == 1 {
			m.MissileAttackRange212 = 19
			m.MissileAttackDelayMin220 = 31
			m.MissileAttackDelayMax224 = 35
		}
		for j := range f.sounds[i].Value {
			f.sounds[i].Value[j] = uint32(600 + 100*i + j)
		}
	}
	for i := 0; i < 2; i++ {
		o := f.projectile(i)
		o.PosVec = types.Ptf(float32(701+i), float32(801+i))
		o.VelVec = types.Ptf(17, -19)
		o.Direction1 = 77
		o.Direction2 = 78
		o.SpeedCur = math.Float32frombits(c.number("speed"))
	}
	f.srv.SetFrame(c.number("frame"))
	f.srv.SetTickRate(c.number("fps"))
	f.srv.Rand.Logic.Reset(int(c.number("rng")))
	f.srv.AI.StackChanged = c.number("initial_changed") != 0
	for i := range f.global {
		f.global[i] = 0x6b
	}
	*f.selected = unsafe.Pointer(f.projectile(1))
	*f.distance = 17
	f.globalBefore = append([]byte(nil), f.global...)
	f.before = nil
	for _, r := range f.regions {
		f.before = append(f.before, bytes.Clone(unsafe.Slice((*byte)(r.Pointer), int(r.Size))))
	}
}
func (f *missileFixture532540) mutate(effect string, cached *server.MonsterUpdateData) {
	u := f.unit()
	switch effect {
	case "none":
	case "swap-update":
		u.UpdateData = unsafe.Pointer(&f.updates[1].Value)
	case "clear-class":
		u.ObjClass = 0
	case "swap-update-and-clear-class":
		u.UpdateData = unsafe.Pointer(&f.updates[1].Value)
		u.ObjClass = 0
	case "advance-frame":
		f.srv.SetFrame(f.srv.Frame() + 3)
	case "replace-sound":
		cached.SoundSet122 = unsafe.Pointer(&f.sounds[1].Value[0])
	case "null-sound":
		cached.SoundSet122 = nil
	case "replace-definition":
		cached.MonsterDef = &f.defs[1].Value
	case "complete-cached":
		cached.Field120_3 = 0x81
	case "change-head":
		cached.AIStackInd = 2
	case "cap-cached-head":
		cached.AIStackInd = 22
	case "change-head-args":
		item := &cached.AIStack[cached.AIStackInd]
		item.Args[0] = uintptr(math.Float32bits(410))
		item.Args[1] = uintptr(math.Float32bits(510))
		item.Args[2] = 0
	case "change-direction":
		u.Direction1 = 96
	case "change-position":
		u.PosVec = types.Ptf(325, 305)
	default:
		panic("unknown original missile callback effect " + effect)
	}
}
func (f *missileFixture532540) callback(name string) func(*server.Object) {
	return func(u *server.Object) {
		if u != f.unit() {
			panic("missile callback unit")
		}
		cached := (*server.MonsterUpdateData)(u.UpdateData)
		effect := f.row.Input.text(name + "_effect")
		f.event(map[string]any{"service": name, "class_before": uint32(u.ObjClass), "live_update_before": f.role(uintptr(u.UpdateData)), "effect": effect})
		f.mutate(effect, cached)
	}
}
func (f *missileFixture532540) runtime() server.MonsterMissileNativeRuntime532540 {
	audio := func(id uint32, u *server.Object) {
		if u != f.unit() {
			panic("missile audio unit")
		}
		f.event(map[string]any{"service": "audio", "sound": id, "live_update": f.role(uintptr(u.UpdateData))})
		f.mutate(f.row.Input.text("audio_effect"), &f.updates[0].Value)
	}
	return server.MonsterMissileNativeRuntime532540{
		AudioEvent: audio,
		BuffOff: func(u *server.Object, enchant server.EnchantID) {
			if u != f.unit() {
				panic("missile buff unit")
			}
			e := int32(enchant)
			server.SpellBuffOff4FF5B0(server.SpellBuffOffHooks4FF5B0[*server.Object]{
				LoadBuffArg: func() int32 { return e }, LoadUnitArg: func() *server.Object { return u }, LoadBuffs: func(v *server.Object) uint32 { return v.Buffs },
				SetBuffFlags: func(v *server.Object, flags uint32) {
					v.Buffs = flags
					f.event(map[string]any{"service": "publish-buffs", "flags": flags})
					count := 0
					for _, ev := range f.events {
						if ev["service"] == "publish-buffs" {
							count++
						}
					}
					if count == 1 {
						f.mutate(f.row.Input.text("buff_effect"), &f.updates[0].Value)
					}
				},
				StoreDuration: func(v *server.Object, e int32, n uint16) { v.BuffsDur[e] = n }, StorePower: func(v *server.Object, e int32, n uint8) { v.BuffsPower[e] = n },
				EnchantSpell: func(e int32) int32 {
					if e == 0 {
						return 10
					}
					if e == 23 {
						return 11
					}
					panic("undeclared missile enchant resource")
				},
				SpellAudio: func(spell, field int32) int32 {
					if field != 2 {
						panic("missile spell audio field")
					}
					if spell == 10 {
						return 900
					}
					if spell == 11 {
						return 923
					}
					panic("undeclared missile spell resource")
				},
				Audio: func(id int32, v *server.Object, kind, code int32) {
					if kind != 0 || code != 0 {
						panic("missile buff audio flags")
					}
					audio(uint32(id), v)
				},
			})
		},
		PlayerAttack: func(u *server.Object) int {
			f.callback("player-attack")(u)
			return int(int32(f.row.Input.number("attack_return")))
		},
		NewObjectByTypeID: func(name string) *server.Object {
			if name != "FixtureMissileA" && name != "FixtureMissileB" {
				panic("missile allocation type " + name)
			}
			var o *server.Object
			if n := f.row.Input.number("allocation"); n != 0 {
				o = f.projectile(int(n) - 1)
			}
			f.event(map[string]any{"service": "allocate", "type_name": name, "returned": f.role(uintptr(unsafe.Pointer(o)))})
			f.mutate(f.row.Input.text("allocate_effect"), &f.updates[0].Value)
			return o
		},
		MapTraceRay: func(from, to types.Pointf, flags server.MapTraceFlags) bool {
			if flags != 5 {
				panic("missile trace flags")
			}
			result := f.row.Input.number("trace")
			f.event(map[string]any{"service": "trace", "line_bits": []uint32{math.Float32bits(from.X), math.Float32bits(from.Y), math.Float32bits(to.X), math.Float32bits(to.Y)}, "flags": uint32(flags), "result": result})
			f.mutate(f.row.Input.text("trace_effect"), &f.updates[0].Value)
			return result != 0
		},
		CreateObjectAt: func(o, owner *server.Object, pos types.Pointf) {
			if owner != f.unit() || (o != f.projectile(0) && o != f.projectile(1)) {
				panic("missile create pointers")
			}
			f.event(map[string]any{"service": "create", "projectile": f.role(uintptr(unsafe.Pointer(o))), "position_bits": []uint32{math.Float32bits(pos.X), math.Float32bits(pos.Y)}})
			o.PosVec = pos
			f.mutate(f.row.Input.text("create_effect"), &f.updates[0].Value)
		},
		DelayedDelete: func(o *server.Object) {
			if o != f.projectile(0) && o != f.projectile(1) {
				panic("missile delete pointer")
			}
			f.event(map[string]any{"service": "delete", "projectile": f.role(uintptr(unsafe.Pointer(o)))})
			f.mutate(f.row.Input.text("delete_effect"), &f.updates[0].Value)
		},
		CancelAction: func(ai.ActionType) func(*server.Object) {
			if f.row.Input.number("callback_present") == 0 {
				return nil
			}
			return f.callback("cancel")
		},
		EndAction: func(ai.ActionType) func(*server.Object) {
			if f.row.Input.number("callback_present") == 0 {
				return nil
			}
			return f.callback("end")
		},
		ObservePush: func(action ai.ActionType, cached *server.MonsterUpdateData, result *server.AIStackItem) {
			f.pushes = append(f.pushes, meleePush532130{Action: uint32(action), Live: f.meleeFixture532130.role(uintptr(unsafe.Pointer(cached))), Returned: f.meleeFixture532130.role(uintptr(unsafe.Pointer(result)))})
		},
	}
}
func (f *missileFixture532540) validate(t *testing.T, result uintptr, control, checkReturn bool) {
	t.Helper()
	expected := make([][]byte, len(f.before))
	for i, b := range f.before {
		expected[i] = bytes.Clone(b)
	}
	w := f.row.Output
	u := f.unit()
	put16 := func(p unsafe.Pointer, n uint16) {
		var b [2]byte
		binary.LittleEndian.PutUint16(b[:], n)
		f.expectBytes(expected, p, b[:])
	}
	if !control {
		if len(w.UnitPosition) != 2 || len(w.Updates) != 2 || len(w.Projectiles) != 2 || len(w.Durations) != 2 || len(w.Powers) != 2 {
			t.Fatal("original missile output dimensions")
		}
		f.expect32(expected, unsafe.Pointer(&u.ObjClass), w.Class)
		f.expect32(expected, unsafe.Pointer(&u.ObjSubClass), w.Subclass)
		f.expectWord(expected, unsafe.Pointer(&u.UpdateData), f.word(w.Live))
		f.expect32(expected, unsafe.Pointer(&u.Field34), w.Field34)
		f.expect32(expected, unsafe.Pointer(&u.PosVec.X), w.UnitPosition[0])
		f.expect32(expected, unsafe.Pointer(&u.PosVec.Y), w.UnitPosition[1])
		put16(unsafe.Pointer(&u.Direction1), uint16(w.Direction))
		f.expect32(expected, unsafe.Pointer(&u.Buffs), w.Buffs)
		for i, e := range []int{0, 23} {
			put16(unsafe.Pointer(&u.BuffsDur[e]), w.Durations[i])
			f.expect8(expected, unsafe.Pointer(&u.BuffsPower[e]), w.Powers[i])
		}
		for i, state := range w.Updates {
			if len(state.Reset) != 9 || len(state.Slots) != 24 {
				t.Fatal("original missile update dimensions")
			}
			d := &f.updates[i].Value
			f.expect8(expected, unsafe.Pointer(&d.AIStackInd), byte(state.Head))
			f.expectWord(expected, unsafe.Pointer(&d.MonsterDef), f.word(state.Definition))
			f.expectWord(expected, unsafe.Pointer(&d.SoundSet122), f.word(state.Sound))
			f.expect32(expected, unsafe.Pointer(&d.Field128), state.Deadline)
			for j, p := range []*uint32{&d.Field2, &d.Field67, &d.Field74, &d.Field91} {
				f.expect32(expected, unsafe.Pointer(p), state.Reset[j])
			}
			for j, p := range []*byte{&d.Field120_1, &d.Field120_2, &d.Field120_3} {
				f.expect8(expected, unsafe.Pointer(p), byte(state.Reset[j+4]))
			}
			f.expect32(expected, unsafe.Pointer(&d.Field124), state.Reset[7])
			f.expect32(expected, unsafe.Pointer(&d.Field137), state.Reset[8])
			for j, slot := range state.Slots {
				if len(slot.Args) != 4 {
					t.Fatal("original missile slot dimensions")
				}
				item := &d.AIStack[j]
				f.expect32(expected, unsafe.Pointer(&item.Action), slot.Action)
				f.expect32(expected, unsafe.Pointer(&item.Field5), slot.Started)
				for k, arg := range slot.Args {
					f.expectWord(expected, unsafe.Pointer(&item.Args[k]), f.word(arg))
				}
			}
		}
		for i, state := range w.Projectiles {
			if len(state.Position) != 2 || len(state.Velocity) != 2 || len(state.Directions) != 2 {
				t.Fatal("original missile projectile dimensions")
			}
			o := f.projectile(i)
			for j, p := range []*float32{&o.PosVec.X, &o.PosVec.Y} {
				f.expect32(expected, unsafe.Pointer(p), state.Position[j])
			}
			for j, p := range []*float32{&o.VelVec.X, &o.VelVec.Y} {
				f.expect32(expected, unsafe.Pointer(p), state.Velocity[j])
			}
			put16(unsafe.Pointer(&o.Direction1), uint16(state.Directions[0]))
			put16(unsafe.Pointer(&o.Direction2), uint16(state.Directions[1]))
			f.expect32(expected, unsafe.Pointer(&o.SpeedCur), state.Speed)
		}
		if checkReturn {
			if f.row.Entry == 0x532610 {
				result &= 255
			}
			if want := f.word(w.Returned); result != want {
				t.Errorf("native missile return=%#x, want original%+v=%#x", result, w.Returned, want)
			}
		}
		if f.srv.Frame() != w.Frame || f.srv.Rand.Logic.Index() != int(w.RNG) || f.srv.AI.StackChanged != (w.Changed != 0) {
			t.Errorf("native missile frame/RNG/stack=%d/%d/%v, original=%d/%d/%d", f.srv.Frame(), f.srv.Rand.Logic.Index(), f.srv.AI.StackChanged, w.Frame, w.RNG, w.Changed)
		}
		if !reflect.DeepEqual(f.events, w.Events) && !(len(f.events) == 0 && len(w.Events) == 0) {
			t.Errorf("native missile events=%v; original=%v", f.events, w.Events)
		}
		if !reflect.DeepEqual(f.pushes, w.Pushes) && !(len(f.pushes) == 0 && len(w.Pushes) == 0) {
			t.Errorf("native missile real pushes=%v; original=%v", f.pushes, w.Pushes)
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
			t.Errorf("guarded missile region%d differs in%d bytes; first+%d", i, count, first)
		}
	}
	if !bytes.Equal(f.global, f.globalBefore) || *f.selected != unsafe.Pointer(f.projectile(1)) || *f.distance != 17 {
		t.Error("unrelated global pointer, packed distance or neighbors changed")
	}
}

type missileLegacyServer532540 struct {
	Server
	fixture *missileFixture532540
}

func (s *missileLegacyServer532540) S() *server.Server { return s.fixture.srv }
func (s *missileLegacyServer532540) MonsterMissileNativeRuntime532540() server.MonsterMissileNativeRuntime532540 {
	return s.fixture.runtime()
}
func TestMissileActionsOriginalActualC532540(t *testing.T) {
	path := os.Getenv("NOX_MISSILE_ACTIONS_ORIGINAL_REFERENCE")
	if path == "" {
		path = "../server/testdata/monster_missile_actions_original_532540.json.gz"
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(raw)
	if hex.EncodeToString(digest[:]) != "93dc8dc054be8d2180a4f169ae33ba1f37229c6fc1cffc85903ae5444b06119b" {
		t.Fatal("unsealed original missile actions reference")
	}
	z, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	var ref struct {
		Schema   string             `json:"schema"`
		Original string             `json:"original_sha256"`
		Control  uint16             `json:"control_word"`
		Rows     []missileRow532540 `json:"rows"`
	}
	if err := json.NewDecoder(z).Decode(&ref); err != nil {
		t.Fatal(err)
	}
	if ref.Schema != "original-monster-missile-actions-532540-v1" || ref.Original != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" || ref.Control != 0xe3f || len(ref.Rows) != 4150 {
		t.Fatal("original missile reference dimensions")
	}
	f := &missileFixture532540{meleeFixture532130: meleeNewFixture532130(t)}
	engine, game := noxflags.GetEngine(), noxflags.GetGame()
	noxflags.UnsetEngine(noxflags.EngineShowAI)
	noxflags.UnsetGame(game)
	t.Cleanup(func() {
		noxflags.UnsetEngine(noxflags.GetEngine())
		noxflags.SetEngine(engine)
		noxflags.UnsetGame(noxflags.GetGame())
		noxflags.SetGame(game)
	})
	control := os.Getenv("NOX_MISSILE_ACTIONS_CONTROL_ONLY") == "true"
	entries := map[uint32]int{}
	for i := range ref.Rows {
		row := &ref.Rows[i]
		entries[row.Entry]++
		t.Run(row.Name, func(t *testing.T) {
			f.reset(row)
			old := GetServer
			GetServer = func() Server { return &missileLegacyServer532540{fixture: f} }
			defer func() { GetServer = old }()
			result := uintptr(0)
			defer func() {
				if v := recover(); v != nil {
					t.Errorf("native panic for original-defined missile context: %v", v)
				}
				f.validate(t, result, control, true)
			}()
			if !control {
				result = ccall.CallUPtrPtr(missilePublicCBefore532540(row.Entry), f.unit().CObj())
			}
		})
	}
	if !reflect.DeepEqual(entries, map[uint32]int{0x532540: 1364, 0x532610: 2786}) {
		t.Fatal("original missile per-entry dimensions")
	}
	t.Logf("original actual C missile-action contexts=%d control-only=%v native-returns-projectiles-cached-context-and-events=true", len(ref.Rows), control)
}
