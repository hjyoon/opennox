package legacy

import (
	"bytes"
	"crypto/sha256"
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
	"github.com/opennox/libs/player"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/resource_spells_original_52bf20_52bf50_52dd50_v2.json.gz
var resourceSpellsOriginal []byte

//go:embed testdata/resource_spells_x87_52dd50_v1.json.gz
var resourceSpellsX87 []byte

type resourceSpellX87Row struct {
	Index             int    `json:"index"`
	Input             uint64 `json:"input"`
	Factor            uint32 `json:"factor"`
	Multiply          bool   `json:"multiply"`
	OriginalSpill     uint32 `json:"original_spill"`
	OriginalConverted int32  `json:"original_converted"`
	Spill             uint32 `json:"spill"`
	Converted         int32  `json:"converted"`
}

func resourceSpellX87References(t *testing.T, cases []resourceSpellOriginalCase) map[int]uint32 {
	t.Helper()
	var reference struct {
		Original string                `json:"original_fixture_sha256"`
		Rows     []resourceSpellX87Row `json:"arithmetic_rows"`
		Changes  map[int]uint32        `json:"signaling_NaN_corrections"`
	}
	var envelope map[string]json.RawMessage
	npcVoiceOriginalDecode(t, resourceSpellsX87, "dfc548589c5cfb31412159315eaaf0aa3e3447424f031e809cd5b1bcaa2ff64f", "a04c882a9b643b54ec8ac196380f39dfcaaa4214cece8de4ce672d49bbcae952", 4, &envelope)
	for key, dest := range map[string]any{"original_fixture_sha256": &reference.Original, "arithmetic_rows": &reference.Rows, "signaling_NaN_corrections": &reference.Changes} {
		if err := json.Unmarshal(envelope[key], dest); err != nil {
			t.Fatal("independent x87 reference decode", key, err)
		}
	}
	if reference.Original != "4b315c8c6fbeb059247385d7c77334f534a94712a32007b8dbe0bff8c0bb3f9c" || len(reference.Rows) != 261 || len(reference.Changes) != 6 {
		t.Fatal("independent x87 reference scope differs")
	}
	seen := map[int]bool{}
	for _, r := range reference.Rows {
		if r.Index < 0 || r.Index >= len(cases) || seen[r.Index] {
			t.Fatal("invalid or duplicate x87 reference index", r.Index)
		}
		seen[r.Index] = true
		c := cases[r.Index]
		var convert, adjust *resourceSpellEvent
		for j := range c.Events {
			switch c.Events[j].Kind {
			case "original-convert":
				convert = &c.Events[j]
			case "adjust-hp":
				adjust = &c.Events[j]
			}
		}
		if convert == nil || adjust == nil || r.Input != c.Input.Balance || r.OriginalSpill != uint32(convert.Args[0].(float64)) || r.OriginalConverted != int32(adjust.Args[1].(float64)) || r.Converted != r.OriginalConverted {
			t.Fatal("x87 inputs/results differ from the sealed original row", r.Index)
		}
		var factor uint32
		var multiply bool
		s := convert.Before
		if caster := c.Input.Caster; caster >= 0 && s.Flags[caster]&4 != 0 {
			class := s.Classes[s.PlayerRefs[s.UpdateRefs[caster]]]
			if slot, ok := map[byte]int{0: 0, 2: 1, 1: 2}[class]; ok {
				factor, multiply = s.Factors[slot], true
			}
		}
		if r.Factor != factor || r.Multiply != multiply {
			t.Fatal("x87 multiplier inputs differ from live original owners", r.Index)
		}
		if changed, ok := reference.Changes[r.Index]; ok {
			if r.Index < 338 || r.Index > 343 || r.Input != 0x7ff0000000000001 || r.OriginalSpill != 0xffc00000 || r.Spill != 0x7fc00000 || changed != r.Spill {
				t.Fatal("x87 difference exceeds independently observed SNaN rows", r.Index)
			}
		} else if r.Spill != r.OriginalSpill {
			t.Fatal("unaccounted x87/original mismatch", r.Index)
		}
		spill, converted, native := resourceSpellX87C(r.Input, r.Factor, r.Multiply)
		if native != (runtime.GOARCH == "amd64" || runtime.GOARCH == "386") || native && (spill != r.Spill || converted != r.Converted) {
			t.Fatal("local x87 sequence disagrees with sealed independent results", r.Index, spill, converted, native)
		}
	}
	return reference.Changes
}

type resourceSpellState struct {
	Target     int      `json:"target"`
	Flags      []uint32 `json:"class_flags"`
	UpdateRefs []int    `json:"update_refs"`
	PlayerRefs []int    `json:"player_refs"`
	Mana       []uint16 `json:"maximum_mana"`
	Classes    []byte   `json:"player_classes"`
	Factors    []uint32 `json:"class_multiplier_bits"`
	Position   []uint32 `json:"position_bits"`
}
type resourceSpellMutation struct {
	Stage  string `json:"stage"`
	At     int    `json:"at"`
	Action string `json:"action"`
}
type resourceSpellEvent struct {
	Kind   string             `json:"kind"`
	Args   []any              `json:"args"`
	Before resourceSpellState `json:"before"`
}
type resourceSpellOriginalCase struct {
	Input struct {
		Name      string                `json:"name"`
		Kind      string                `json:"kind"`
		Target    int                   `json:"target"`
		Second    int                   `json:"second"`
		Caster    int                   `json:"caster"`
		Fourth    int                   `json:"fourth"`
		Spell     int32                 `json:"spell"`
		Level     int32                 `json:"level"`
		HP        []uint32              `json:"hp"`
		MaximumHP []uint32              `json:"maximum_hp"`
		Cost      uint32                `json:"cost"`
		Sound     uint32                `json:"sound"`
		Balance   uint64                `json:"balance_bits64"`
		Mutation  resourceSpellMutation `json:"mutation"`
	} `json:"input"`
	Result       int32                   `json:"result"`
	Events       []resourceSpellEvent    `json:"events"`
	Mutations    []resourceSpellMutation `json:"mutations"`
	Before       resourceSpellState      `json:"before"`
	After        resourceSpellState      `json:"after"`
	CallerBefore string                  `json:"original_caller_before_hex"`
	CallerAfter  string                  `json:"original_caller_after_hex"`
	HeapBefore   string                  `json:"original_heap_before_sha256"`
	HeapAfter    string                  `json:"original_heap_after_sha256"`
}

func resourceSpellCases(t *testing.T) []resourceSpellOriginalCase {
	t.Helper()
	var rows []resourceSpellOriginalCase
	npcVoiceOriginalDecode(t, resourceSpellsOriginal, "3bfa7592c8e94b82098feda21ce8cdc6a72c14354d26b363c7ba2572f5ed6d7e", "4b315c8c6fbeb059247385d7c77334f534a94712a32007b8dbe0bff8c0bb3f9c", 354, &rows)
	return rows
}
func resourceSpellOriginalHeap(s resourceSpellState) []byte {
	out := make([]byte, 65536)
	for i := range out {
		out[i] = byte(i*31 + 19)
	}
	objectAddress := func(i int) uint32 {
		if i < 0 {
			return 0
		}
		return 0x21001000 + uint32(i)*0x1000
	}
	for i := 0; i < 3; i++ {
		binary.LittleEndian.PutUint32(out[0x1000+i*0x1000+8:], s.Flags[i])
		binary.LittleEndian.PutUint32(out[0x1000+i*0x1000+748:], 0x21005000+uint32(s.UpdateRefs[i])*0x400)
		binary.LittleEndian.PutUint16(out[0x5000+i*0x400+8:], s.Mana[i])
		binary.LittleEndian.PutUint32(out[0x5000+i*0x400+276:], 0x21007000+uint32(s.PlayerRefs[i])*0x1000)
		out[0x7000+i*0x1000+2251] = s.Classes[i]
	}
	binary.LittleEndian.PutUint32(out[0xd000:], objectAddress(s.Target))
	binary.LittleEndian.PutUint32(out[0xd004:], s.Position[0])
	binary.LittleEndian.PutUint32(out[0xd008:], s.Position[1])
	return out
}
func resourceSpellVerifyOriginal(t *testing.T, c resourceSpellOriginalCase) {
	t.Helper()
	if fmt.Sprintf("%x", sha256.Sum256(resourceSpellOriginalHeap(c.Before))) != c.HeapBefore || fmt.Sprintf("%x", sha256.Sum256(resourceSpellOriginalHeap(c.After))) != c.HeapAfter {
		t.Fatal("sealed original whole heap differs from declared fields and canaries")
	}
	a, b := npcVoiceUnhex(t, c.CallerBefore), npcVoiceUnhex(t, c.CallerAfter)
	if len(a) != 28 || len(b) != 28 || !bytes.Equal(a[:20], b[:20]) || !bytes.Equal(a[24:], b[24:]) {
		t.Fatal("original caller owner escaped declared float spill")
	}
	converted := false
	for _, e := range c.Events {
		if e.Kind == "original-convert" {
			converted = true
			if binary.LittleEndian.Uint32(b[20:]) != uint32(e.Args[0].(float64)) {
				t.Fatal("original binary32 caller spill differs from unchanged converter input")
			}
		}
	}
	if !converted && !bytes.Equal(a, b) {
		t.Fatal("nonfloating path changed original caller")
	}
}

type resourceSpellArena struct {
	raw                                         []byte
	objects                                     [3]*server.Object
	updates                                     [3]*server.PlayerUpdateData
	players                                     [3]*server.Player
	objectOffsets, updateOffsets, playerOffsets [3]int
	arg                                         *server.SpellAcceptArg
	argOffset                                   int
	factors                                     [3]float32
	layout                                      [13]uintptr
}

func resourceSpellArenaNew(t *testing.T, placement int) *resourceSpellArena {
	t.Helper()
	raw, free := alloc.MakeHigh([]byte{}, 0x40000)
	t.Cleanup(free)
	v := &resourceSpellArena{raw: raw, layout: resourceSpellLayoutC(), argOffset: 0x3e000 + placement}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&raw[0])) <= math.MaxUint32 {
		t.Fatal("actual C owners must be foreign pointers above4GiB")
	}
	if unsafe.Sizeof(server.Object{}) > 0x3000 || unsafe.Sizeof(server.PlayerUpdateData{}) > 0x3000 || unsafe.Sizeof(server.Player{}) > 0x7000 {
		t.Fatal("foreign owner strides too small")
	}
	for i := 0; i < 3; i++ {
		v.objectOffsets[i] = 256 + i*0x4000 + placement
		v.updateOffsets[i] = 0x10000 + i*0x4000 + placement
		v.playerOffsets[i] = 0x20000 + i*0x8000 + placement
		v.objects[i] = (*server.Object)(unsafe.Pointer(&raw[v.objectOffsets[i]]))
		v.updates[i] = (*server.PlayerUpdateData)(unsafe.Pointer(&raw[v.updateOffsets[i]]))
		v.players[i] = (*server.Player)(unsafe.Pointer(&raw[v.playerOffsets[i]]))
	}
	v.arg = (*server.SpellAcceptArg)(unsafe.Pointer(&raw[v.argOffset]))
	return v
}
func (v *resourceSpellArena) object(i int) *server.Object {
	if i < 0 {
		return nil
	}
	return v.objects[i]
}
func (v *resourceSpellArena) index(p *server.Object) int {
	if p == nil {
		return -1
	}
	for i, x := range v.objects {
		if p == x {
			return i
		}
	}
	panic("unmapped native resource Object")
}
func (v *resourceSpellArena) state() resourceSpellState {
	s := resourceSpellState{Target: v.index(v.arg.Obj), Position: []uint32{math.Float32bits(v.arg.Pos.X), math.Float32bits(v.arg.Pos.Y)}}
	for i := 0; i < 3; i++ {
		s.Flags = append(s.Flags, uint32(v.objects[i].ObjClass))
		u := -1
		for j, p := range v.updates {
			if unsafe.Pointer(p) == v.objects[i].UpdateData {
				u = j
			}
		}
		if u < 0 {
			panic("unmapped native update")
		}
		s.UpdateRefs = append(s.UpdateRefs, u)
		pl := -1
		for j, p := range v.players {
			if p == v.updates[i].Player {
				pl = j
			}
		}
		if pl < 0 {
			panic("unmapped native Player")
		}
		s.PlayerRefs = append(s.PlayerRefs, pl)
		s.Mana = append(s.Mana, v.updates[i].ManaMax)
		s.Classes = append(s.Classes, byte(v.players[i].Info().PlayerClass()))
		s.Factors = append(s.Factors, math.Float32bits(v.factors[i]))
	}
	return s
}
func (v *resourceSpellArena) set(s resourceSpellState) {
	for i := 0; i < 3; i++ {
		v.objects[i].ObjClass = object.Class(s.Flags[i])
		v.objects[i].UpdateData = unsafe.Pointer(v.updates[s.UpdateRefs[i]])
		v.updates[i].ManaMax = s.Mana[i]
		v.updates[i].Player = v.players[s.PlayerRefs[i]]
		v.players[i].Info().SetPlayerClass(player.Class(s.Classes[i]))
		v.factors[i] = math.Float32frombits(s.Factors[i])
	}
	v.arg.Obj = v.object(s.Target)
	v.arg.Pos = types.Pointf{X: math.Float32frombits(s.Position[0]), Y: math.Float32frombits(s.Position[1])}
}
func (v *resourceSpellArena) expected(before []byte, s resourceSpellState) []byte {
	out := bytes.Clone(before)
	pointer := func(pos int, p unsafe.Pointer) {
		if v.layout[12] == 8 {
			binary.LittleEndian.PutUint64(out[pos:], uint64(uintptr(p)))
		} else {
			binary.LittleEndian.PutUint32(out[pos:], uint32(uintptr(p)))
		}
	}
	for i := 0; i < 3; i++ {
		binary.LittleEndian.PutUint32(out[v.objectOffsets[i]+int(v.layout[1]):], s.Flags[i])
		pointer(v.objectOffsets[i]+int(v.layout[2]), unsafe.Pointer(v.updates[s.UpdateRefs[i]]))
		binary.LittleEndian.PutUint16(out[v.updateOffsets[i]+int(v.layout[4]):], s.Mana[i])
		pointer(v.updateOffsets[i]+int(v.layout[5]), unsafe.Pointer(v.players[s.PlayerRefs[i]]))
		out[v.playerOffsets[i]+int(v.layout[7])] = s.Classes[i]
	}
	pointer(v.argOffset+int(v.layout[9]), unsafe.Pointer(v.object(s.Target)))
	binary.LittleEndian.PutUint32(out[v.argOffset+int(v.layout[10]):], s.Position[0])
	binary.LittleEndian.PutUint32(out[v.argOffset+int(v.layout[11]):], s.Position[1])
	return out
}
func (v *resourceSpellArena) run(t *testing.T, route int, c resourceSpellOriginalCase, correctedSpill *uint32) {
	t.Helper()
	for i := range v.raw {
		v.raw[i] = 0xa5
	}
	v.set(c.Before)
	before := bytes.Clone(v.raw)
	if !reflect.DeepEqual(v.state(), c.Before) {
		t.Fatal("native input state normalization")
	}
	var got []resourceSpellEvent
	var mutations []resourceSpellMutation
	counts := map[string]int{}
	emit := func(kind string, args ...any) {
		for i, a := range args {
			switch x := a.(type) {
			case int:
				args[i] = float64(x)
			case int32:
				args[i] = float64(x)
			case uint32:
				args[i] = float64(x)
			}
		}
		got = append(got, resourceSpellEvent{Kind: kind, Args: args, Before: v.state()})
		if kind == "original-convert" {
			return
		}
		at := counts[kind]
		counts[kind]++
		m := c.Input.Mutation
		if m.Stage != kind || m.At != at {
			return
		}
		mutations = append(mutations, m)
		switch m.Action {
		case "target1":
			v.arg.Obj = v.objects[1]
		case "target2":
			v.arg.Obj = v.objects[2]
		case "target-nil":
			v.arg.Obj = nil
		case "caster-nonplayer":
			v.objects[c.Input.Caster].ObjClass = 0
		case "caster-class2":
			v.players[c.Input.Caster].Info().SetPlayerClass(player.Conjurer)
		case "caster-update1":
			v.objects[c.Input.Caster].UpdateData = unsafe.Pointer(v.updates[1])
		case "update-player1":
			v.updates[c.Input.Caster].Player = v.players[1]
		case "maximum-mana":
			v.updates[0].ManaMax = 0x2211
		case "factor":
			v.factors[0] = 0.25
		default:
			t.Fatal("unknown sealed mutation", m)
		}
	}
	deps := resourceSpellDeps52BF20{
		setMaximumHP: func(p *server.Object) { emit("set-max-hp", v.index(p)) }, addMana: func(p *server.Object, n int32) { emit("add-mana", v.index(p), n) },
		hp: func(p *server.Object) uint16 {
			i := v.index(p)
			var n uint16
			if i >= 0 {
				n = uint16(c.Input.HP[i])
			}
			emit("hp", i)
			return n
		}, maximumHP: func(p *server.Object) uint16 {
			i := v.index(p)
			var n uint16
			if i >= 0 {
				n = uint16(c.Input.MaximumHP[i])
			}
			emit("max-hp", i)
			return n
		},
		cost: func(id, which int32) int32 { emit("cost", id, which); return int32(c.Input.Cost) }, recharge: func(p *server.Object, n int32) { emit("recharge", v.index(p), n) },
		balance: func(key string) float64 { emit("balance", key); return math.Float64frombits(c.Input.Balance) }, multiplier: func(class byte) float32 {
			switch class {
			case 0:
				return v.factors[0]
			case 2:
				return v.factors[1]
			case 1:
				return v.factors[2]
			default:
				t.Fatal("unexpected multiplier class", class)
				return 0
			}
		},
		convert: func(value float32) int32 {
			emit("original-convert", math.Float32bits(value))
			return resourceSpellFloatToInt52DD50(value)
		}, adjustHP: func(p *server.Object, n int32) { emit("adjust-hp", v.index(p), n) }, sound: func(id, which int32) int32 { emit("sound", id, which); return int32(c.Input.Sound) }, audio: func(id int32, p *server.Object) { emit("audio", id, v.index(p), 0, 0) },
	}
	old := resourceSpellDepsForCall52BF20
	resourceSpellDepsForCall52BF20 = func() resourceSpellDeps52BF20 { return deps }
	defer func() { resourceSpellDepsForCall52BF20 = old }()
	mode, ok := map[string]int{"health": 0, "mana": 1, "lesser": 2}[c.Input.Kind]
	if !ok {
		t.Fatal("unknown original root")
	}
	result := resourceSpellActualC(route, mode, c.Input.Spell, v.object(c.Input.Second), v.object(c.Input.Caster), v.object(c.Input.Fourth), v.arg, c.Input.Level)
	if result != c.Result {
		t.Fatal("actual C/native Go result", result, c.Result)
	}
	if len(got) != len(c.Events) {
		t.Fatalf("service count: got%+v want%+v", got, c.Events)
	}
	for i := range got {
		expected := c.Events[i]
		if correctedSpill != nil && expected.Kind == "original-convert" {
			expected.Args = []any{float64(*correctedSpill)}
		}
		if !reflect.DeepEqual(got[i], expected) {
			t.Fatalf("ordered service%d: got%+v want%+v", i, got[i], expected)
		}
	}
	if len(mutations) != len(c.Mutations) {
		t.Fatal("provider mutation count", mutations, c.Mutations)
	}
	for i := range mutations {
		if mutations[i] != c.Mutations[i] {
			t.Fatal("provider mutation", mutations, c.Mutations)
		}
	}
	if !reflect.DeepEqual(v.state(), c.After) || !bytes.Equal(v.raw, v.expected(before, c.After)) {
		t.Fatal("whole native owners/pointer fields/argument positions or canaries differ")
	}
}
func TestResourceSpellsNativeLayoutCABI(t *testing.T) {
	layout := resourceSpellLayoutC()
	if layout[0] > unsafe.Sizeof(server.Object{}) || layout[3] > unsafe.Sizeof(server.PlayerUpdateData{}) || layout[6] > unsafe.Sizeof(server.Player{}) {
		t.Fatal("C owner exceeds Go foreign allocation", layout)
	}
	if layout[1] != unsafe.Offsetof(server.Object{}.ObjClass) || layout[2] != unsafe.Offsetof(server.Object{}.UpdateData) || layout[4] != unsafe.Offsetof(server.PlayerUpdateData{}.ManaMax) || layout[5] != unsafe.Offsetof(server.PlayerUpdateData{}.Player) || layout[8] != unsafe.Sizeof(server.SpellAcceptArg{}) || layout[9] != unsafe.Offsetof(server.SpellAcceptArg{}.Obj) || layout[10] != unsafe.Offsetof(server.SpellAcceptArg{}.Pos) || layout[11] != unsafe.Offsetof(server.SpellAcceptArg{}.Pos)+4 || layout[12] != unsafe.Sizeof(uintptr(0)) {
		t.Fatal("independent C and Go native field layouts differ", layout)
	}
	v := resourceSpellArenaNew(t, 0)
	for i := range v.raw {
		v.raw[i] = 0xa5
	}
	before := bytes.Clone(v.raw)
	v.players[0].Info().SetPlayerClass(player.Wizard)
	before[v.playerOffsets[0]+int(layout[7])] = byte(player.Wizard)
	if !bytes.Equal(before, v.raw) {
		t.Fatal("native C Player class offset differs from Go setter")
	}
}
func TestResourceSpellsNativeOriginalCABI(t *testing.T) {
	rows := resourceSpellCases(t)
	for _, c := range rows {
		resourceSpellVerifyOriginal(t, c)
	}
	corrections := resourceSpellX87References(t, rows)
	for route, name := range []string{"direct-public-C", "actual-shared-callback", "production-Go-wrapper"} {
		for _, placement := range []int{0, 8, 24} {
			t.Run(fmt.Sprintf("%s/placement-%d", name, placement), func(t *testing.T) {
				v := resourceSpellArenaNew(t, placement)
				for i, c := range rows {
					var corrected *uint32
					if spill, ok := corrections[i]; ok {
						corrected = &spill
					}
					t.Run(fmt.Sprintf("%03d-%s", i, c.Input.Name), func(t *testing.T) { v.run(t, route, c, corrected) })
				}
				runtime.GC()
			})
		}
	}
	t.Log("unchanged original resource spell cases=354, roots=351bytes, original converter=17bytes, original startup trace=319bytes; exact canonical control0xe3f; three foreign placements and direct/shared-C/production-Go routes; declared service transcripts and whole native canaries")
}
