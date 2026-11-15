package legacy

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/monster_cast_original_541300_5413b0_541490_v1.json.gz
var monsterCastOriginal []byte

type monsterCastObjectState struct {
	Class      uint32    `json:"class_bits"`
	Subclass   uint32    `json:"subclass"`
	Position   [2]uint32 `json:"position"`
	Velocity   [2]uint32 `json:"velocity"`
	Directions [2]uint16 `json:"directions"`
	Update     int       `json:"update"`
}

type monsterCastUpdateState struct {
	Flags      uint32 `json:"flags"`
	Accuracy   uint32 `json:"accuracy"`
	Frame      byte   `json:"frame"`
	Ready      byte   `json:"ready"`
	Index      int8   `json:"stack_index"`
	HeadSpell  uint32 `json:"head_spell"`
	HeadArg2   uint32 `json:"head_arg2"`
	HeadArg3   uint32 `json:"head_arg3"`
	Definition int    `json:"definition"`
}

type monsterCastArgumentState struct {
	Object   int       `json:"object"`
	Position [2]uint32 `json:"position"`
}

type monsterCastState struct {
	Objects          []monsterCastObjectState  `json:"objects"`
	Updates          []monsterCastUpdateState  `json:"updates"`
	PlayerLinks      []int                     `json:"player_links"`
	DefinitionFrames []uint32                  `json:"definition_frames"`
	VoiceCast        uint32                    `json:"voice_cast"`
	Argument         *monsterCastArgumentState `json:"argument"`
	Output           [2]uint32                 `json:"output"`
	OriginalRNGIndex *uint32                   `json:"original_rng_index,omitempty"`
}

type monsterCastEvent struct {
	Kind   string           `json:"kind"`
	Args   []any            `json:"args"`
	Before monsterCastState `json:"before"`
}

type monsterCastCase struct {
	Input struct {
		Name    string                   `json:"name"`
		Root    string                   `json:"root"`
		Spell   int32                    `json:"spell"`
		Mode    int32                    `json:"mode"`
		Target  int                      `json:"target"`
		Objects []monsterCastObjectState `json:"objects"`
		Updates []struct {
			Flags           uint32    `json:"flags"`
			Accuracy        uint32    `json:"accuracy"`
			Frame           byte      `json:"frame"`
			Ready           byte      `json:"ready"`
			Index           int8      `json:"stack_index"`
			Spell           uint32    `json:"spell"`
			Position        [2]uint32 `json:"position"`
			DefinitionFrame uint32    `json:"definition_frame"`
		} `json:"updates"`
		ArgumentPosition [2]uint32 `json:"argument_position"`
		Direction        uint16    `json:"direction"`
		CastReturn       uint32    `json:"cast_return"`
		MorphReturn      uint32    `json:"morph_return"`
		BuffReturn       uint32    `json:"buff_return"`
		VoiceAvailable   bool      `json:"voice_available"`
		VoiceCast        uint32    `json:"voice_cast"`
		Random           [3]uint64 `json:"random_bits64"`
		OutputAlias      string    `json:"output_alias"`
		RNGIndex         uint32    `json:"rng_index"`
		Mutations        []struct {
			Stage  string `json:"stage"`
			At     int    `json:"at"`
			Action string `json:"action"`
			Value  uint32 `json:"value"`
		} `json:"mutations"`
	} `json:"input"`
	Before       monsterCastState   `json:"before"`
	Events       []monsterCastEvent `json:"events"`
	After        monsterCastState   `json:"after"`
	Return       *byte              `json:"defined_return_byte"`
	CallerBefore string             `json:"caller_before_hex"`
	CallerAfter  string             `json:"caller_after_hex"`
}

func monsterCastCases(t *testing.T) []monsterCastCase {
	t.Helper()
	var rows []monsterCastCase
	npcVoiceOriginalDecode(t, monsterCastOriginal, "0c369861ffd59e92e9017fd642a43e59277b3a2f2ee491eb4e6399fcca4dedca", "ecf60d2dc2cafb1ef28fb4f898ae2804e92307739e17b13974702ba81f18e4db", 428, &rows)
	for _, c := range rows {
		a, b := npcVoiceUnhex(t, c.CallerBefore), npcVoiceUnhex(t, c.CallerAfter)
		if len(a) != len(b) || len(a) < 12 || (c.Input.Root != "recoil" && !bytes.Equal(a, b)) || (c.Input.Root == "recoil" && (!bytes.Equal(a[:4], b[:4]) || !bytes.Equal(a[8:], b[8:]))) {
			t.Fatal("original caller escaped its declared recoil float temporary", c.Input.Name)
		}
	}
	return rows
}

type monsterCastArena struct {
	raw         []byte
	objects     [3]*server.Object
	updates     [2]*server.MonsterUpdateData
	players     [2]*server.PlayerUpdateData
	definitions [2]*server.MonsterDef
	voice       unsafe.Pointer
	arg         *server.SpellAcceptArg
	out         *types.Pointf
	separateOut *types.Pointf
	activeArg   *server.SpellAcceptArg
	mode        int32
	layout      [24]uintptr
	headChanges map[int]uint32
	randomIndex func() uint32
}

func monsterCastArenaNew(t *testing.T, placement int) *monsterCastArena {
	t.Helper()
	raw, free := alloc.MakeHigh([]byte{}, 0x20000)
	t.Cleanup(free)
	v := &monsterCastArena{raw: raw, layout: monsterCastLayoutC()}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&raw[0])) <= math.MaxUint32 {
		t.Fatal("native foreign arena must be above4GiB")
	}
	if unsafe.Sizeof(server.MonsterUpdateData{}) >= 0x4000 || unsafe.Sizeof(server.Object{}) >= 0x2000 {
		t.Fatal("foreign owner stride too small")
	}
	for i := range v.objects {
		v.objects[i] = (*server.Object)(unsafe.Pointer(&raw[0x100+i*0x2000+placement]))
	}
	for i := range v.updates {
		v.updates[i] = (*server.MonsterUpdateData)(unsafe.Pointer(&raw[0x8000+i*0x4000+placement]))
		v.players[i] = (*server.PlayerUpdateData)(unsafe.Pointer(&raw[0x10000+i*0x1000+placement]))
		v.definitions[i] = (*server.MonsterDef)(unsafe.Pointer(&raw[0x14000+i*0x1000+placement]))
	}
	v.voice = unsafe.Pointer(&raw[0x16000+placement])
	v.arg = (*server.SpellAcceptArg)(unsafe.Pointer(&raw[0x18000+placement]))
	v.separateOut = (*types.Pointf)(unsafe.Pointer(&raw[0x19000+placement]))
	v.out = v.separateOut
	return v
}

func (v *monsterCastArena) object(i int) *server.Object {
	if i < 0 {
		return nil
	}
	return v.objects[i]
}
func (v *monsterCastArena) objectIndex(p *server.Object) int {
	if p == nil {
		return -1
	}
	for i, o := range v.objects {
		if o == p {
			return i
		}
	}
	panic("unmapped native cast object")
}
func (v *monsterCastArena) updatePointer(i int) unsafe.Pointer {
	if i < 0 {
		return nil
	}
	if i < 2 {
		return unsafe.Pointer(v.updates[i])
	}
	return unsafe.Pointer(v.players[i-2])
}
func (v *monsterCastArena) updateIndex(p unsafe.Pointer) int {
	if p == nil {
		return -1
	}
	for i := 0; i < 4; i++ {
		if p == v.updatePointer(i) {
			return i
		}
	}
	panic("unmapped native cast update")
}
func (v *monsterCastArena) argumentState(a *server.SpellAcceptArg) *monsterCastArgumentState {
	if a == nil {
		return nil
	}
	return &monsterCastArgumentState{Object: v.objectIndex(a.Obj), Position: [2]uint32{math.Float32bits(a.Pos.X), math.Float32bits(a.Pos.Y)}}
}
func (v *monsterCastArena) state() monsterCastState {
	s := monsterCastState{Objects: make([]monsterCastObjectState, 3), Updates: make([]monsterCastUpdateState, 2), PlayerLinks: make([]int, 2), DefinitionFrames: make([]uint32, 2), VoiceCast: *(*uint32)(unsafe.Add(v.voice, 56)), Argument: v.argumentState(v.activeArg), Output: [2]uint32{math.Float32bits(v.out.X), math.Float32bits(v.out.Y)}}
	if v.randomIndex != nil {
		value := v.randomIndex()
		s.OriginalRNGIndex = &value
	}
	for i, o := range v.objects {
		s.Objects[i] = monsterCastObjectState{Class: uint32(o.ObjClass), Subclass: uint32(o.ObjSubClass), Position: [2]uint32{math.Float32bits(o.PosVec.X), math.Float32bits(o.PosVec.Y)}, Velocity: [2]uint32{math.Float32bits(o.VelVec.X), math.Float32bits(o.VelVec.Y)}, Directions: [2]uint16{uint16(o.Direction1), uint16(o.Direction2)}, Update: v.updateIndex(o.UpdateData)}
	}
	for i, u := range v.updates {
		h := &u.AIStack[int(u.AIStackInd)]
		arg2 := uint32(h.Args[2])
		if v.mode == 0 {
			index := v.objectIndex(h.ArgObj(2))
			arg2 = 0
			if index >= 0 {
				arg2 = 0x21001000 + uint32(index)*0x1000
			}
		}
		def := -1
		for j, d := range v.definitions {
			if d == u.MonsterDef {
				def = j
			}
		}
		s.Updates[i] = monsterCastUpdateState{Flags: uint32(u.StatusFlags), Accuracy: math.Float32bits(u.Field330), Frame: u.Field120_1, Ready: u.Field120_2, Index: u.AIStackInd, HeadSpell: h.ArgU32(0), HeadArg2: arg2, HeadArg3: h.ArgU32(3), Definition: def}
		for j, m := range v.updates {
			if m == v.players[i].Field73 {
				s.PlayerLinks[i] = j
			}
		}
		s.DefinitionFrames[i] = v.definitions[i].MissileAttackFrame216
	}
	return s
}

func monsterCastPoint(bits [2]uint32) types.Pointf {
	return types.Pointf{X: math.Float32frombits(bits[0]), Y: math.Float32frombits(bits[1])}
}
func (v *monsterCastArena) prepare(c monsterCastCase) {
	for i := range v.raw {
		v.raw[i] = 0xa5
	}
	v.mode = c.Input.Mode
	v.activeArg = nil
	v.headChanges = map[int]uint32{}
	for i, s := range c.Before.Objects {
		o := v.objects[i]
		o.ObjClass = object.Class(s.Class)
		o.ObjSubClass = object.SubClass(s.Subclass)
		o.PosVec = monsterCastPoint(s.Position)
		o.VelVec = monsterCastPoint(s.Velocity)
		o.Direction1 = server.Dir16(s.Directions[0])
		o.Direction2 = server.Dir16(s.Directions[1])
		o.UpdateData = v.updatePointer(s.Update)
	}
	for i, s := range c.Input.Updates {
		u := v.updates[i]
		u.StatusFlags = object.MonsterStatus(s.Flags)
		u.Field330 = math.Float32frombits(s.Accuracy)
		u.Field120_1 = s.Frame
		u.Field120_2 = s.Ready
		u.AIStackInd = s.Index
		u.MonsterDef = v.definitions[i]
		u.SoundSet122 = v.voice
		u.Field545 = v.players[i]
		v.players[i].Field73 = u
		v.definitions[i].MissileAttackFrame216 = s.DefinitionFrame
		for j := range u.AIStack {
			h := &u.AIStack[j]
			h.Action = 0x12345678
			h.Args = [4]uintptr{uintptr(s.Spell + uint32(j)), 0xaabbccdd, uintptr(s.Position[0]), uintptr(s.Position[1])}
			h.Field5 = 0xdeadbeef
			if c.Input.Mode == 0 {
				h.Args[2] = uintptr(unsafe.Pointer(v.object(c.Input.Target)))
			}
		}
	}
	*(*uint32)(unsafe.Add(v.voice, 56)) = c.Input.VoiceCast
	v.arg.Obj = v.object(c.Input.Target)
	v.arg.Pos = monsterCastPoint(c.Input.ArgumentPosition)
	if c.Input.Root == "direct" {
		v.activeArg = v.arg
	}
	switch c.Input.OutputAlias {
	case "target-position":
		v.out = &v.objects[1].PosVec
	case "target-velocity":
		v.out = &v.objects[1].VelVec
	default:
		v.out = v.separateOut
		*v.out = types.Pointf{X: -9.25, Y: 12.75}
	}
}

func (v *monsterCastArena) offset(p unsafe.Pointer) int {
	return int(uintptr(p) - uintptr(unsafe.Pointer(&v.raw[0])))
}
func (v *monsterCastArena) expected(before []byte, c monsterCastCase) []byte {
	out := bytes.Clone(before)
	put := func(at int, value uint32) { binary.LittleEndian.PutUint32(out[at:], value) }
	putPtr := func(at int, value uintptr) {
		if unsafe.Sizeof(value) == 8 {
			binary.LittleEndian.PutUint64(out[at:], uint64(value))
		} else {
			put(at, uint32(value))
		}
	}
	s := c.After
	for i, o := range s.Objects {
		a := v.offset(unsafe.Pointer(v.objects[i]))
		put(a+int(v.layout[1]), o.Class)
		put(a+int(v.layout[2]), o.Subclass)
		for j := 0; j < 2; j++ {
			put(a+int(v.layout[3])+4*j, o.Position[j])
			put(a+int(v.layout[4])+4*j, o.Velocity[j])
		}
		binary.LittleEndian.PutUint16(out[a+int(v.layout[5]):], o.Directions[0])
		binary.LittleEndian.PutUint16(out[a+int(v.layout[6]):], o.Directions[1])
		putPtr(a+int(v.layout[7]), uintptr(v.updatePointer(o.Update)))
	}
	for at, value := range v.headChanges {
		put(at, value)
	}
	for i, u := range s.Updates {
		a := v.offset(unsafe.Pointer(v.updates[i]))
		out[a+int(v.layout[14])] = u.Frame
		out[a+int(v.layout[15])] = u.Ready
		out[a+int(v.layout[18])] = byte(u.Index)
		put(a+int(v.layout[20]), u.Accuracy)
		put(a+int(v.layout[21]), u.Flags)
		putPtr(a+int(v.layout[16]), uintptr(unsafe.Pointer(v.definitions[u.Definition])))
		h := a + int(v.layout[19]) + int(u.Index)*int(v.layout[11])
		putPtr(h+int(v.layout[12]), uintptr(u.HeadSpell))
		arg2 := uintptr(u.HeadArg2)
		if c.Input.Mode == 0 {
			index := -1
			if u.HeadArg2 != 0 {
				index = int((u.HeadArg2 - 0x21001000) / 0x1000)
			}
			arg2 = uintptr(unsafe.Pointer(v.object(index)))
		}
		putPtr(h+int(v.layout[12])+2*int(unsafe.Sizeof(uintptr(0))), arg2)
		putPtr(h+int(v.layout[12])+3*int(unsafe.Sizeof(uintptr(0))), uintptr(u.HeadArg3))
		putPtr(v.offset(unsafe.Pointer(v.players[i]))+int(v.layout[22]), uintptr(unsafe.Pointer(v.updates[s.PlayerLinks[i]])))
		put(v.offset(unsafe.Pointer(v.definitions[i]))+int(v.layout[23]), s.DefinitionFrames[i])
	}
	put(v.offset(v.voice)+56, s.VoiceCast)
	if c.Input.Root == "direct" && s.Argument != nil {
		a := v.offset(unsafe.Pointer(v.arg))
		putPtr(a+int(v.layout[9]), uintptr(unsafe.Pointer(v.object(s.Argument.Object))))
		put(a+int(v.layout[10]), s.Argument.Position[0])
		put(a+int(v.layout[10])+4, s.Argument.Position[1])
	}
	a := v.offset(unsafe.Pointer(v.out))
	put(a, s.Output[0])
	put(a+4, s.Output[1])
	return out
}

func (v *monsterCastArena) run(t *testing.T, route int, c monsterCastCase) {
	t.Helper()
	v.prepare(c)
	before := bytes.Clone(v.raw)
	if !reflect.DeepEqual(v.state(), c.Before) {
		t.Fatalf("native normalized input got%+v want%+v", v.state(), c.Before)
	}
	got := make([]monsterCastEvent, 0)
	counts := map[string]int{}
	emit := func(kind string, args ...any) {
		got = append(got, monsterCastEvent{Kind: kind, Args: args, Before: v.state()})
		at := counts[kind]
		counts[kind]++
		for _, m := range c.Input.Mutations {
			if m.Stage != kind || m.At != at {
				continue
			}
			switch m.Action {
			case "cached-flags":
				v.updates[0].StatusFlags = object.MonsterStatus(m.Value)
			case "unit-update":
				v.objects[0].UpdateData = v.updatePointer(int(m.Value))
			case "player-link":
				v.players[0].Field73 = v.updates[m.Value]
			case "unit-position":
				v.objects[0].PosVec = types.Pointf{X: 81.1, Y: -29.3}
			case "target-position":
				v.objects[1].PosVec = types.Pointf{X: -47.1, Y: 82.3}
			case "target-velocity":
				v.objects[1].VelVec = types.Pointf{X: -0.7, Y: 1.3}
			case "accuracy":
				v.updates[0].Field330 = math.Float32frombits(m.Value)
			case "head-spell":
				h := &v.updates[0].AIStack[int(v.updates[0].AIStackInd)]
				h.Args[0] = uintptr(m.Value)
				v.headChanges[v.offset(unsafe.Pointer(&h.Args[0]))] = m.Value
			case "stack-index":
				v.updates[0].AIStackInd = int8(m.Value)
			case "argument":
				if v.activeArg == nil {
					t.Fatal("argument mutation without selected record")
				}
				v.activeArg.Obj = v.objects[2]
				v.activeArg.Pos = types.Pointf{X: 91.1, Y: -38.3}
			case "output":
				*v.out = types.Pointf{X: -11.3, Y: 21.1}
			default:
				t.Fatal("unknown sealed cast service mutation", m)
			}
		}
	}
	rt := server.MonsterCastRuntime541300{
		MorphTo: func(unit *server.Object) byte {
			if unit != v.objects[0] {
				t.Fatal("morph owner")
			}
			emit("morph-to", 0)
			unit.ObjClass = unit.ObjClass&^object.ClassMonster | object.ClassPlayer
			unit.ObjSubClass = 0
			unit.UpdateData = unsafe.Pointer(v.players[0])
			return byte(unit.ObjClass)
		},
		MorphFrom: func(unit *server.Object) byte {
			if unit != v.objects[0] {
				t.Fatal("reverse morph owner")
			}
			emit("morph-from", 0)
			index := v.updateIndex(unit.UpdateData)
			if index < 2 {
				unit.UpdateData = unsafe.Pointer(v.updates[index])
			} else {
				unit.UpdateData = unsafe.Pointer(v.players[index-2].Field73)
			}
			unit.ObjClass = unit.ObjClass&^object.ClassPlayer | object.ClassMonster
			unit.ObjSubClass = 16
			return byte(c.Input.MorphReturn)
		},
		Direction: func(unit *server.Object, pos *types.Pointf) {
			if unit != v.objects[0] {
				t.Fatal("direction owner")
			}
			v.activeArg = (*server.SpellAcceptArg)(unsafe.Add(unsafe.Pointer(pos), -int(unsafe.Offsetof(server.SpellAcceptArg{}.Pos))))
			emit("direction", 0, [2]uint32{math.Float32bits(pos.X), math.Float32bits(pos.Y)})
			unit.Direction2 = server.Dir16(c.Input.Direction)
		},
		Cast: func(id int32, unit *server.Object, arg *server.SpellAcceptArg) int32 {
			if unit != v.objects[0] {
				t.Fatal("cast owner")
			}
			v.activeArg = arg
			emit("cast", uint32(id), 0, v.argumentState(arg))
			return int32(c.Input.CastReturn)
		},
		Buff: func(unit *server.Object, buff int32) int32 {
			emit("buff", v.objectIndex(unit), buff)
			return int32(c.Input.BuffReturn)
		},
		Voice: func(unit *server.Object) unsafe.Pointer {
			emit("voice", v.objectIndex(unit))
			if c.Input.VoiceAvailable {
				return v.voice
			}
			return nil
		},
		Audio: func(id uint32, unit *server.Object) { emit("audio", id, v.objectIndex(unit), 0, 0) },
		RandomFloat: func(min, max float32) float64 {
			at := counts["random"]
			if at >= 3 {
				t.Fatal("extra RNG call")
			}
			emit("random", math.Float32bits(min), math.Float32bits(max))
			return math.Float64frombits(c.Input.Random[at])
		},
	}
	previous := monsterCastRuntimeForCall541300
	monsterCastRuntimeForCall541300 = func() server.MonsterCastRuntime541300 { return rt }
	defer func() { monsterCastRuntimeForCall541300 = previous }()
	result := monsterCastActualC(route, c.Input.Root, c.Input.Spell, v.objects[0], v.object(c.Input.Target), v.arg, c.Input.Mode, v.out)
	if c.Return != nil && result != *c.Return {
		t.Fatalf("native direct return=%#x want%#x", result, *c.Return)
	}
	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	var normalized []monsterCastEvent
	if err = json.Unmarshal(encoded, &normalized); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(normalized, c.Events) {
		for i := 0; i < len(normalized) && i < len(c.Events); i++ {
			if !reflect.DeepEqual(normalized[i], c.Events[i]) {
				t.Fatalf("ordered event%d got%+v want%+v", i, normalized[i], c.Events[i])
			}
		}
		t.Fatalf("ordered event count got%d want%d", len(normalized), len(c.Events))
	}
	if !reflect.DeepEqual(v.state(), c.After) {
		t.Fatalf("native final state got%+v want%+v", v.state(), c.After)
	}
	expected := v.expected(before, c)
	if !bytes.Equal(v.raw, expected) {
		for i, b := range v.raw {
			if b != expected[i] {
				t.Fatalf("whole native arena mismatch offset%#x got%#x want%#x", i, b, expected[i])
			}
		}
	}
}

func TestMonsterCastNativeLayoutCABI(t *testing.T) {
	v := monsterCastLayoutC()
	o := server.Object{}
	u := server.MonsterUpdateData{}
	a := server.AIStackItem{}
	arg := server.SpellAcceptArg{}
	want := [24]uintptr{v[0], unsafe.Offsetof(o.ObjClass), unsafe.Offsetof(o.ObjSubClass), unsafe.Offsetof(o.PosVec), unsafe.Offsetof(o.VelVec), unsafe.Offsetof(o.Direction1), unsafe.Offsetof(o.Direction2), unsafe.Offsetof(o.UpdateData), unsafe.Sizeof(arg), unsafe.Offsetof(arg.Obj), unsafe.Offsetof(arg.Pos), unsafe.Sizeof(a), unsafe.Offsetof(a.Args), unsafe.Offsetof(a.Field5), unsafe.Offsetof(u.Field120_1), unsafe.Offsetof(u.Field120_2), unsafe.Offsetof(u.MonsterDef), unsafe.Offsetof(u.SoundSet122), unsafe.Offsetof(u.AIStackInd), unsafe.Offsetof(u.AIStack), unsafe.Offsetof(u.Field330), unsafe.Offsetof(u.StatusFlags), unsafe.Offsetof(server.PlayerUpdateData{}.Field73), unsafe.Offsetof(server.MonsterDef{}.MissileAttackFrame216)}
	if v != want || v[0] > unsafe.Sizeof(o) {
		t.Fatal("independent native C/Go layouts differ", v, want)
	}
}

func TestMonsterCastNativeOriginalCABI(t *testing.T) {
	rows := monsterCastCases(t)
	for route, name := range []string{"direct-public-C", "typed-C-callback", "production-Go-wrapper"} {
		for _, placement := range []int{0, 8, 24} {
			t.Run(fmt.Sprintf("%s/placement%d", name, placement), func(t *testing.T) {
				v := monsterCastArenaNew(t, placement)
				for i, c := range rows {
					t.Run(fmt.Sprintf("%03d-%s", i, c.Input.Name), func(t *testing.T) { v.run(t, route, c) })
				}
				runtime.GC()
			})
		}
	}
	t.Log("unchanged original monster-cast cases=428 roots=511bytes startup=319bytes control0xe3f; direct return byte and action/recoil effects, ordered external services, cached update/head links, output aliases, three foreign placements and C/typed-callback/Go routes, complete native arena canaries")
}
