package legacy

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/monster_generator_damage_original_4e27d0_v1.json.gz
var generatorDamageOriginal []byte

type generatorDamageObjectState struct {
	Flags       uint32    `json:"flags"`
	Status      uint32    `json:"status"`
	Position    [2]uint32 `json:"position"`
	EffectFrame uint32    `json:"effect_frame"`
	Health      int       `json:"health"`
	Update      int       `json:"update"`
}

type generatorDamageState struct {
	Objects []generatorDamageObjectState `json:"objects"`
	Health  [][3]uint16                  `json:"health"`
	Updates [][2]uint32                  `json:"updates"`
}

type generatorDamageEvent struct {
	Kind         string               `json:"kind"`
	Args         []uint32             `json:"args"`
	Before       generatorDamageState `json:"before"`
	After        generatorDamageState `json:"after"`
	PointAlias   string               `json:"point_alias"`
	PointBits    [2]uint32            `json:"point_bits"`
	CachedUpdate int                  `json:"cached_update"`
}

type generatorDamageCase struct {
	Input struct {
		Name           string      `json:"name"`
		Flags          uint32      `json:"flags"`
		Status         uint32      `json:"status"`
		Position       [2]uint32   `json:"position"`
		WeaponPosition [2]uint32   `json:"weapon_position"`
		Source         int         `json:"source"`
		Weapon         int         `json:"weapon"`
		Health         [][3]uint16 `json:"health"`
		Update         int         `json:"update"`
		HealthOwner    int         `json:"health_owner"`
		EffectFrame    uint32      `json:"effect_frame"`
		Frame          uint32      `json:"frame"`
		Damage         int32       `json:"damage"`
		DamageType     int32       `json:"damage_type"`
		HPReturns      [2]uint32   `json:"hp_returns"`
		DefaultReturn  uint32      `json:"default_return"`
		PointWrite     *[2]uint32  `json:"point_write"`
		Mutations      []struct {
			Stage  string `json:"stage"`
			At     int    `json:"at"`
			Action string `json:"action"`
			Value  uint32 `json:"value"`
		} `json:"mutations"`
	} `json:"input"`
	Before generatorDamageState   `json:"before"`
	After  generatorDamageState   `json:"after"`
	Events []generatorDamageEvent `json:"events"`
	Return uint32                 `json:"return_u32"`
}

type generatorDamageArena struct {
	raw      []byte
	objects  [3]*server.Object
	updates  [2]*server.MonsterGenUpdateData
	health   [2]*server.HealthData
	layout   [19]uintptr
	expected []byte
}

func generatorDamageArenaNew(t *testing.T, placement int) *generatorDamageArena {
	t.Helper()
	raw, free := alloc.MakeHigh([]byte{}, 0x10000)
	t.Cleanup(free)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&raw[0])) <= math.MaxUint32 {
		t.Fatal("native foreign owners must be above 4 GiB")
	}
	v := &generatorDamageArena{raw: raw, layout: generatorDamageLayoutC()}
	for i := range v.objects {
		v.objects[i] = (*server.Object)(unsafe.Pointer(&raw[0x1000+i*0x1000+placement]))
	}
	for i := range v.updates {
		v.updates[i] = (*server.MonsterGenUpdateData)(unsafe.Pointer(&raw[0x8000+i*0x200+placement]))
		v.health[i] = (*server.HealthData)(unsafe.Pointer(&raw[0x6000+i*0x100+placement]))
	}
	return v
}

func (v *generatorDamageArena) object(index int) *server.Object {
	if index < 0 {
		return nil
	}
	return v.objects[index]
}

func (v *generatorDamageArena) offset(p unsafe.Pointer) int {
	return int(uintptr(p) - uintptr(unsafe.Pointer(&v.raw[0])))
}

func (v *generatorDamageArena) state() generatorDamageState {
	s := generatorDamageState{Objects: make([]generatorDamageObjectState, 3), Health: make([][3]uint16, 2), Updates: make([][2]uint32, 2)}
	for i, o := range v.objects {
		h, u := -1, -1
		for j, p := range v.health {
			if o.HealthData == p {
				h = j
			}
		}
		for j, p := range v.updates {
			if o.UpdateData == unsafe.Pointer(p) {
				u = j
			}
		}
		if o.HealthData != nil && h < 0 || o.UpdateData != nil && u < 0 {
			panic("unmapped native generator link")
		}
		s.Objects[i] = generatorDamageObjectState{uint32(o.ObjFlags), o.Field5, [2]uint32{math.Float32bits(o.PosVec.X), math.Float32bits(o.PosVec.Y)}, o.Frame134, h, u}
	}
	for i, h := range v.health {
		s.Health[i] = [3]uint16{h.Cur, h.Field2, h.Max}
	}
	for i, u := range v.updates {
		s.Updates[i] = [2]uint32{u.Field48, u.FuncInd52}
	}
	return s
}

func (v *generatorDamageArena) prepare(t *testing.T, c generatorDamageCase) {
	t.Helper()
	for i := range v.raw {
		v.raw[i] = byte(i*31 + 19)
	}
	put32 := func(at int, value uint32) { binary.LittleEndian.PutUint32(v.raw[at:at+4], value) }
	putPointer := func(at int, value unsafe.Pointer) {
		if unsafe.Sizeof(uintptr(0)) == 8 {
			binary.LittleEndian.PutUint64(v.raw[at:at+8], uint64(uintptr(value)))
		} else {
			put32(at, uint32(uintptr(value)))
		}
	}
	for i, o := range v.objects {
		at := v.offset(unsafe.Pointer(o))
		flags, status := uint32(0), uint32(0x13579bdf)
		xy := c.Input.Position
		if i == 0 {
			flags, status = c.Input.Flags, c.Input.Status
		}
		if i == 2 {
			xy = c.Input.WeaponPosition
		}
		put32(at+int(v.layout[1]), flags)
		put32(at+int(v.layout[2]), status)
		put32(at+int(v.layout[3]), xy[0])
		put32(at+int(v.layout[4]), xy[1])
		put32(at+int(v.layout[5]), c.Input.EffectFrame)
		var h, u unsafe.Pointer
		if i == 0 {
			h = unsafe.Pointer(v.health[c.Input.HealthOwner])
			u = unsafe.Pointer(v.updates[c.Input.Update])
		}
		putPointer(at+int(v.layout[6]), h)
		putPointer(at+int(v.layout[7]), u)
	}
	for i, h := range v.health {
		at := v.offset(unsafe.Pointer(h))
		binary.LittleEndian.PutUint16(v.raw[at+int(v.layout[10]):], c.Input.Health[i][0])
		binary.LittleEndian.PutUint16(v.raw[at+int(v.layout[16]):], c.Input.Health[i][1])
		binary.LittleEndian.PutUint16(v.raw[at+int(v.layout[11]):], c.Input.Health[i][2])
	}
	for i, u := range v.updates {
		at := v.offset(unsafe.Pointer(u))
		put32(at+int(v.layout[13]), 0xa1000000+uint32(i))
		put32(at+int(v.layout[14]), 0xb2000000+uint32(i))
	}
	generatorDamageBindC(v.objects[0])
	if got := v.state(); !reflect.DeepEqual(got, c.Before) {
		t.Fatalf("native input state got%+v want%+v", got, c.Before)
	}
	v.expected = bytes.Clone(v.raw)
}

func (v *generatorDamageArena) checkArena(t *testing.T) {
	t.Helper()
	if bytes.Equal(v.raw, v.expected) {
		return
	}
	for i := range v.raw {
		if v.raw[i] != v.expected[i] {
			t.Fatalf("undeclared native arena write at%d: got%02x want%02x", i, v.raw[i], v.expected[i])
		}
	}
}

func (v *generatorDamageArena) run(t *testing.T, route int, c generatorDamageCase) {
	t.Helper()
	v.prepare(t, c)
	next, frames := 0, 0
	counts := make(map[string]int)
	owner := func(p *server.Object, original uint32) bool {
		if original == 0 {
			return p == nil
		}
		for i := range v.objects {
			if original == 0x21001000+uint32(i)*0x1000 {
				return p == v.objects[i]
			}
		}
		t.Fatalf("unmapped original owner %#x", original)
		return false
	}
	event := func(kind string, args int, builtin func(*generatorDamageEvent)) *generatorDamageEvent {
		t.Helper()
		v.checkArena(t)
		if next >= len(c.Events) {
			t.Fatalf("extra native service %s", kind)
		}
		e := &c.Events[next]
		next++
		if e.Kind != kind || len(e.Args) != args {
			t.Fatalf("native service %s/%d want%+v", kind, args, *e)
		}
		if got := v.state(); !reflect.DeepEqual(got, e.Before) {
			t.Fatalf("%s before got%+v want%+v", kind, got, e.Before)
		}
		counts[kind]++
		if builtin != nil {
			builtin(e)
		}
		for _, m := range c.Input.Mutations {
			at := m.At
			if at == 0 {
				at = 1
			}
			if m.Stage != kind || at != counts[kind] {
				continue
			}
			u := v.objects[0]
			switch m.Action {
			case "flags":
				u.ObjFlags = object.Flags(m.Value)
			case "status":
				u.Field5 = m.Value
			case "position-x":
				u.PosVec.X = math.Float32frombits(m.Value)
			case "position-y":
				u.PosVec.Y = math.Float32frombits(m.Value)
			case "update":
				u.UpdateData = unsafe.Pointer(v.updates[m.Value])
			case "health":
				u.HealthData = v.health[m.Value]
			case "effect-frame":
				u.Frame134 = m.Value
			default:
				i := int(m.Action[len(m.Action)-1] - '0')
				if strings.HasPrefix(m.Action, "health-cur-") {
					v.health[i].Cur = uint16(m.Value)
				} else if strings.HasPrefix(m.Action, "health-max-") {
					v.health[i].Max = uint16(m.Value)
				} else {
					t.Fatal("unknown mutation", m.Action)
				}
			}
		}
		if got := v.state(); !reflect.DeepEqual(got, e.After) {
			t.Fatalf("%s after got%+v want%+v", kind, got, e.After)
		}
		v.expected = bytes.Clone(v.raw)
		return e
	}
	rt := server.MonsterGeneratorDamageNativeRuntime4E27D0{
		Frame: func() uint32 { frames++; return c.Input.Frame },
		PointFX: func(op, subtype byte, point *types.Pointf) {
			event("point-fx", 3, func(e *generatorDamageEvent) {
				if op != byte(e.Args[0]) || subtype != byte(e.Args[1]) || point == nil {
					t.Fatal("native FX signature", op, subtype, point)
				}
				alias := "temporary"
				if point == &v.objects[0].PosVec {
					alias = "target-position"
				}
				address, base := uintptr(unsafe.Pointer(point)), uintptr(unsafe.Pointer(&v.raw[0]))
				if alias == "temporary" && address >= base && address < base+uintptr(len(v.raw)) {
					t.Fatal("native temporary aliases a foreign owner")
				}
				if bits := [2]uint32{math.Float32bits(point.X), math.Float32bits(point.Y)}; alias != e.PointAlias || bits != e.PointBits {
					t.Fatalf("native FX got%s/%v want%s/%v", alias, bits, e.PointAlias, e.PointBits)
				}
				if c.Input.PointWrite != nil {
					point.X = math.Float32frombits(c.Input.PointWrite[0])
					point.Y = math.Float32frombits(c.Input.PointWrite[1])
				}
			})
		},
		Audio: func(id int32, unit *server.Object) {
			event("audio", 4, func(e *generatorDamageEvent) {
				if uint32(id) != e.Args[0] || !owner(unit, e.Args[1]) || e.Args[2] != 0 || e.Args[3] != 0 {
					t.Fatal("native audio args")
				}
			})
		},
		GetHP: func(unit *server.Object) uint16 {
			event("get-hp", 1, func(e *generatorDamageEvent) {
				if !owner(unit, e.Args[0]) {
					t.Fatal("native HP owner")
				}
			})
			return uint16(c.Input.HPReturns[counts["get-hp"]-1])
		},
		Default: func(unit, source, weapon *server.Object, damage, typ int32) int32 {
			event("default-damage", 5, func(e *generatorDamageEvent) {
				if !owner(unit, e.Args[0]) || !owner(source, e.Args[1]) || !owner(weapon, e.Args[2]) || uint32(damage) != e.Args[3] || uint32(typ) != e.Args[4] {
					t.Fatal("native default args", unit, source, weapon, damage, typ)
				}
			})
			return int32(c.Input.DefaultReturn)
		},
		Script: func(block *server.ScriptCallback, source, unit *server.Object, code server.ScriptEventType) {
			event("script", 3, func(e *generatorDamageEvent) {
				if e.CachedUpdate < 0 || e.CachedUpdate >= 2 || unsafe.Pointer(block) != unsafe.Pointer(&v.updates[e.CachedUpdate].Field48) || !owner(source, e.Args[1]) || !owner(unit, e.Args[2]) || code != server.NoxEventGeneratorDamage {
					t.Fatal("cached native script block/owners or explicit native event23")
				}
			})
		},
		SetStatus: func(unit *server.Object, mask uint32) {
			event("set-status", 2, func(e *generatorDamageEvent) {
				if !owner(unit, e.Args[0]) || mask != e.Args[1] {
					t.Fatal("native set status args")
				}
				unit.Field5 |= mask
			})
		},
		UnsetStatus: func(unit *server.Object, mask uint32) {
			event("unset-status", 2, func(e *generatorDamageEvent) {
				if !owner(unit, e.Args[0]) || mask != e.Args[1] {
					t.Fatal("native unset status args")
				}
				unit.Field5 &= ^mask
			})
		},
	}
	old := monsterGeneratorDamageRuntimeForCall4E27D0
	monsterGeneratorDamageRuntimeForCall4E27D0 = func() server.MonsterGeneratorDamageNativeRuntime4E27D0 { return rt }
	defer func() { monsterGeneratorDamageRuntimeForCall4E27D0 = old }()
	got := generatorDamageActualCall(route, v.objects[0], v.object(c.Input.Source), v.object(c.Input.Weapon), c.Input.Damage, c.Input.DamageType)
	want := c.Return
	if route == 2 && want != 0 {
		want = 1
	}
	if uint32(got) != want {
		t.Fatalf("native return %#x want%#x", uint32(got), want)
	}
	if next != len(c.Events) {
		t.Fatal("missing native services", next, len(c.Events))
	}
	if expectedFrames := 1; frames != expectedFrames && c.Input.Flags&0x8020 == 0 || c.Input.Flags&0x8020 != 0 && frames != 0 {
		t.Fatal("original single frame load/gate", frames)
	}
	if state := v.state(); !reflect.DeepEqual(state, c.After) {
		t.Fatalf("native final got%+v want%+v", state, c.After)
	}
	v.checkArena(t)
	runtime.KeepAlive(v.raw)
}

func TestMonsterGeneratorDamageNativeLayout(t *testing.T) {
	o, h, u := server.Object{}, server.HealthData{}, server.MonsterGenUpdateData{}
	want := [19]uintptr{unsafe.Offsetof(o.ScriptPickup) + unsafe.Sizeof(o.ScriptPickup), unsafe.Offsetof(o.ObjFlags), unsafe.Offsetof(o.Field5), unsafe.Offsetof(o.PosVec), unsafe.Offsetof(o.PosVec) + 4, unsafe.Offsetof(o.Frame134), unsafe.Offsetof(o.HealthData), unsafe.Offsetof(o.UpdateData), unsafe.Offsetof(o.Damage), unsafe.Sizeof(h), unsafe.Offsetof(h.Cur), unsafe.Offsetof(h.Max), unsafe.Sizeof(u), unsafe.Offsetof(u.Field48), unsafe.Offsetof(u.FuncInd52), unsafe.Sizeof(server.ScriptCallback{}), unsafe.Offsetof(h.Field2), unsafe.Sizeof(types.Pointf{}), unsafe.Offsetof(types.Pointf{}.Y)}
	if got := generatorDamageLayoutC(); got != want || got[0] > unsafe.Sizeof(o) {
		t.Fatalf("independent C/native Go shared layouts got%v want%v", got, want)
	}
}

func TestMonsterGeneratorDamageNativeOriginal(t *testing.T) {
	var rows []generatorDamageCase
	npcVoiceOriginalDecode(t, generatorDamageOriginal, "72700654d9c174a572bb197f3cb91a4ac2510a6854aae7bd60f5eaffe43bd1ad", "28a4d6c6a77b092066bf4474af49c987e306aee46c74d0c235b2dc85e83348ff", 1230, &rows)
	for route, name := range []string{"public-C", "typed-C-callback", "registered-Go-damage"} {
		for _, placement := range []int{0, 64, 256} {
			t.Run(fmt.Sprintf("%s/placement%d", name, placement), func(t *testing.T) {
				v := generatorDamageArenaNew(t, placement)
				for _, c := range rows {
					t.Run(c.Input.Name, func(t *testing.T) { v.run(t, route, c) })
				}
			})
		}
	}
	t.Log("unchanged original generator damage1230cases;468root+58normalize+17float-int+17double-float+319startup bytes, control0xe3f. Complete native arena, ordered services/mutations, cached update/script identity and live health/status, exact C int return versus production bool observation. Only original frame global4bytes seeded; original script call3args versus explicit native event23 binding.")
}
