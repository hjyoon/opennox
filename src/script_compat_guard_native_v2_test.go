package opennox

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/script"
	"github.com/opennox/libs/script/lua"
	"github.com/opennox/libs/script/lua/mapv0"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Inputs are C-owned, server-associated objects and real parsed AUD data.
// The interface and actual Lua VM use the production resolver, native order,
// action stack, and audio queue. No service or result is replaced by a hook.
type luaCompatGuardFixture struct {
	s          *Server
	unit       *server.Object
	update     *server.MonsterUpdateData
	definition *server.MonsterDef
	sounds     *[18]uint32
	owner      *server.Object
	other      *server.Object
	vm         *lua.VM
}

func luaCompatGuardNew(t *testing.T) *luaCompatGuardFixture {
	t.Helper()
	f := &luaCompatGuardFixture{s: scriptNSRoamFlagServer515C80(t)}
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameHost)
	f.s.SetFrame(37)
	f.s.SetTickRate(60)
	f.s.Audio.Init(f.s.Server)
	t.Cleanup(f.s.Audio.Free)
	luaCompatAudioDefinitions(t, f.s)
	if !f.s.Objs.Init(8) {
		t.Fatal("cannot initialize native object allocator")
	}
	t.Cleanup(f.s.Objs.FreeObjects)
	f.unit = f.s.Objs.NewObject(&server.ObjectType{})
	f.owner = f.s.Objs.NewObject(&server.ObjectType{})
	f.other = f.s.Objs.NewObject(&server.ObjectType{})
	f.owner.ObjClass, f.other.ObjClass = object.ClassPlayer, object.ClassSimple
	var free func()
	f.update, free = alloc.New(server.MonsterUpdateData{})
	t.Cleanup(free)
	f.definition, free = alloc.New(server.MonsterDef{})
	t.Cleanup(free)
	f.sounds, free = alloc.New([18]uint32{})
	t.Cleanup(free)
	f.sounds[17] = uint32(sound.SoundAnchorCast)
	id, freeID := alloc.CString("lua_guard_target")
	t.Cleanup(freeID)
	f.unit.IDPtr, f.unit.UpdateData = unsafe.Pointer(id), unsafe.Pointer(f.update)
	f.unit.ObjClass, f.unit.ObjFlags = object.ClassMonster, object.FlagEnabled
	f.unit.ObjOwner, f.unit.Field129 = f.owner, f.other
	f.unit.TypeInd = 0x1234
	f.unit.PosVec, f.unit.Direction2 = types.Ptf(41.25, -82.5), server.Dir16(0x1357)
	f.s.Objs.List = f.unit
	t.Cleanup(func() { f.s.Objs.List = nil })
	for _, pointer := range []unsafe.Pointer{unsafe.Pointer(f.unit), unsafe.Pointer(f.owner), unsafe.Pointer(f.other), unsafe.Pointer(f.update), unsafe.Pointer(f.definition), unsafe.Pointer(f.sounds), unsafe.Pointer(id)} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
			t.Fatalf("native Guard input is not above 4 GiB: %p", pointer)
		}
	}
	var game script.Game = noxScriptImpl{f.s}
	f.vm = lua.NewVM(nil, game, "")
	f.vm.InitAPI("Nox", mapv0.New)
	t.Cleanup(func() { _ = f.vm.Close() })
	if _, err := f.vm.Exec(`target = Nox.Object("lua_guard_target"); assert(type(target) == "userdata")`); err != nil {
		t.Fatal(err)
	}
	return f
}

func luaCompatGuardNoPanic(t *testing.T, call func()) (ok bool) {
	t.Helper()
	defer func() {
		if value := recover(); value != nil {
			t.Errorf("Guard panicked: %v", value)
		}
	}()
	call()
	return true
}

func TestLuaCompatGuardV2OriginalNativeOrder(t *testing.T) {
	f := luaCompatGuardNew(t)
	unitInitial := *f.unit
	ownerBefore := scriptNSRoamFlagBytes515C80(unsafe.Pointer(f.owner), unsafe.Sizeof(*f.owner))
	otherBefore := scriptNSRoamFlagBytes515C80(unsafe.Pointer(f.other), unsafe.Sizeof(*f.other))
	soundsBefore := scriptNSRoamFlagBytes515C80(unsafe.Pointer(f.sounds), unsafe.Sizeof(*f.sounds))
	classes := []struct {
		name     string
		class    object.Class
		subclass object.SubClass
		missile  bool
		weapon   uint32
		shoot    bool
	}{
		{"monster-melee", object.ClassMonster, 0, false, 0, false},
		{"monster-ranged", object.ClassMonster, 0, true, 0, true},
		{"npc-melee", object.ClassMonster, object.SubClass(object.MonsterNPC), true, 0, false},
		{"npc-ranged", object.ClassMonster, object.SubClass(object.MonsterNPC), false, 0x80, true},
		{"player", object.ClassPlayer, 0, false, 0, false},
		{"world", object.ClassSimple, 0, false, 0, false},
	}
	speeds := []struct {
		bits   uint32
		moving bool
	}{
		{0xbf800000, false}, {0x3c23d709, false}, {0x3c23d70a, true},
		{0x3f800000, true}, {0x7fc01234, false},
	}
	// GAME.EXE 00533B4E reads a signed WORD and stores a DWORD.
	directions := []struct {
		word uint16
		arg  uint32
	}{{0, 0}, {0x7fff, 0x7fff}, {0x8000, 0xffff8000}, {0xffff, 0xffffffff}}
	for _, class := range classes {
		for _, speed := range speeds {
			for _, blocked := range []bool{false, true} {
				for _, direction := range directions {
					for _, audible := range []bool{false, true} {
						t.Run(fmt.Sprintf("%s/speed-%08x/blocked-%t/facing-%04x/audio-%t", class.name, speed.bits, blocked, direction.word, audible), func(t *testing.T) {
							*f.unit = unitInitial
							f.unit.ObjClass, f.unit.ObjSubClass = class.class, class.subclass
							f.unit.SpeedBase, f.unit.Direction1 = math.Float32frombits(speed.bits), server.Dir16(direction.word)
							if blocked {
								f.unit.ObjFlags |= object.Flags(0x8000)
							}
							*f.definition = server.MonsterDef{}
							if class.missile {
								f.definition.MissileName148[0] = 'a'
							}
							*f.update = server.MonsterUpdateData{
								MonsterDef: f.definition, AIStackInd: 0, WeaponEquipFlags: class.weapon,
								StatusFlags: object.MonsterStatus(0x80000085), Aggression: .875, Aggression2: 9.25, SightRange: 321.5,
								Field2: 17, Field67: 18, Field74: 19, Field91: 20,
								Field120_0: 0xa5, Field120_1: 21, Field120_2: 22, Field120_3: 23,
								Field124: 51, Field137: 93, Field333: 0xdeadbeef,
							}
							f.update.AIStack[0] = server.AIStackItem{Action: uint32(ai.ACTION_IDLE), Args: [4]uintptr{0x11223344, 0x55667788, 0x9abcdef0, 0x76543210}}
							if unsafe.Sizeof(uintptr(0)) == 8 {
								f.update.AIStack[0].Args[3] |= uintptr(uint64(0x12345678) << 32)
							}
							if audible {
								f.update.SoundSet122 = unsafe.Pointer(f.sounds)
							}
							if class.class != object.ClassMonster {
								f.unit.UpdateData = nil
							}
							initial := scriptNSRoamFlagBytes515C80(unsafe.Pointer(f.update), unsafe.Sizeof(*f.update))
							want := bytes.Clone(initial)
							eligible := class.class == object.ClassMonster && speed.moving && !blocked
							if eligible {
								word := func(offset uintptr, value uint32) { binary.LittleEndian.PutUint32(want[offset:offset+4], value) }
								word(unsafe.Offsetof(f.update.Aggression), 0x3f000000)
								word(unsafe.Offsetof(f.update.SightRange), 0x437a0000)
								if class.shoot {
									word(unsafe.Offsetof(f.update.StatusFlags), 0x800000c5)
								}
								for _, offset := range []uintptr{unsafe.Offsetof(f.update.Field2), unsafe.Offsetof(f.update.Field67), unsafe.Offsetof(f.update.Field74), unsafe.Offsetof(f.update.Field91)} {
									word(offset, 0)
								}
								for _, offset := range []uintptr{unsafe.Offsetof(f.update.Field120_1), unsafe.Offsetof(f.update.Field120_2), unsafe.Offsetof(f.update.Field120_3)} {
									want[offset] = 0
								}
								word(unsafe.Offsetof(f.update.Field124), 37)
								word(unsafe.Offsetof(f.update.Field137), 37)
								stack := unsafe.Offsetof(f.update.AIStack)
								word(stack+unsafe.Offsetof(f.update.AIStack[0].Action), 4)
								for i, value := range []uint32{0x42250000, 0xc2a50000, direction.arg} {
									offset := stack + unsafe.Offsetof(f.update.AIStack[0].Args) + uintptr(i)*unsafe.Sizeof(uintptr(0))
									if unsafe.Sizeof(uintptr(0)) == 8 {
										binary.LittleEndian.PutUint64(want[offset:offset+8], uint64(value))
									} else {
										word(offset, value)
									}
								}
							}
							unitBefore := scriptNSRoamFlagBytes515C80(unsafe.Pointer(f.unit), unsafe.Sizeof(*f.unit))
							definitionBefore := scriptNSRoamFlagBytes515C80(unsafe.Pointer(f.definition), unsafe.Sizeof(*f.definition))
							var expectedEvents []luaCompatAudioSnapshot
							if eligible && audible {
								event := luaCompatAudioSnapshot{sound.SoundAnchorCast, 0x42250000, 0xc2a50000, f.unit, 0, 0}
								expectedEvents = []luaCompatAudioSnapshot{event, event}
							}
							gameBefore, engineBefore := noxflags.GetGame(), noxflags.GetEngine()
							var api script.OffensiveGroup = scrObject{asObjectS(f.unit)}
							for _, route := range []struct {
								name string
								call func()
							}{
								{"native-control", func() { legacy.Nox_xxx_orderUnit_533900(f.unit, f.unit, 3) }},
								{"legacy-interface", api.Guard},
								{"actual-Lua", func() {
									if _, err := f.vm.Exec(`target:Guard()`); err != nil {
										t.Errorf("actual Lua Guard: %v", err)
									}
								}},
							} {
								copy(unsafe.Slice((*byte)(unsafe.Pointer(f.update)), len(initial)), initial)
								f.s.Audio.Reset()
								f.s.AI.StackChanged = false
								if !luaCompatGuardNoPanic(t, func() { route.call(); route.call() }) {
									continue
								}
								if !bytes.Equal(want, scriptNSRoamFlagBytes515C80(unsafe.Pointer(f.update), unsafe.Sizeof(*f.update))) {
									t.Errorf("%s changed the original complete Guard update contract", route.name)
								}
								events, _ := luaCompatAudioEvents(t, f.s)
								if !reflect.DeepEqual(events, expectedEvents) {
									t.Errorf("%s real sound events=%#v, want %#v", route.name, events, expectedEvents)
								}
								if f.s.AI.StackChanged != eligible || f.s.Frame() != 37 || f.s.TickRate() != 60 || noxflags.GetGame() != gameBefore || noxflags.GetEngine() != engineBefore {
									t.Errorf("%s changed original stack/timing/flags contract", route.name)
								}
								for _, record := range []struct {
									pointer unsafe.Pointer
									size    uintptr
									before  []byte
								}{
									{unsafe.Pointer(f.unit), unsafe.Sizeof(*f.unit), unitBefore},
									{unsafe.Pointer(f.owner), unsafe.Sizeof(*f.owner), ownerBefore},
									{unsafe.Pointer(f.other), unsafe.Sizeof(*f.other), otherBefore},
									{unsafe.Pointer(f.definition), unsafe.Sizeof(*f.definition), definitionBefore},
									{unsafe.Pointer(f.sounds), unsafe.Sizeof(*f.sounds), soundsBefore},
								} {
									if !bytes.Equal(record.before, scriptNSRoamFlagBytes515C80(record.pointer, record.size)) {
										t.Errorf("%s modified object/owner/owned-list/definition/sound bytes", route.name)
									}
								}
							}
							f.s.Audio.Reset()
						})
					}
				}
			}
		}
	}
}

func TestLuaCompatGuardV2NilAndUnitGroup(t *testing.T) {
	t.Run("nil-wrapper", func(t *testing.T) { luaCompatGuardNoPanic(t, (scrObject{}).Guard) })
	t.Run("empty-group", func(t *testing.T) { luaCompatGuardNoPanic(t, script.NewUnitGroup("empty").Guard) })
	t.Run("ordered-members", func(t *testing.T) {
		f := luaCompatGuardNew(t)
		f.unit.SpeedBase, f.unit.Direction1 = 1, server.Dir16(0xffff)
		*f.update = server.MonsterUpdateData{MonsterDef: f.definition, SoundSet122: unsafe.Pointer(f.sounds), AIStackInd: 0, Aggression: .875, SightRange: 321.5}
		unit := scrObject{asObjectS(f.unit)}
		player := scrObject{asObjectS(f.owner)}
		group := script.NewUnitGroup("guard", scrObject{}, unit, player, unit)
		f.s.Audio.Reset()
		if !luaCompatGuardNoPanic(t, group.Guard) {
			return
		}
		events, _ := luaCompatAudioEvents(t, f.s)
		if len(events) != 2 || events[0].obj != f.unit || events[1].obj != f.unit || f.update.AIStackHead().Type() != ai.ACTION_GUARD || f.update.AIStackHead().ArgU32(2) != 0xffffffff || math.Float32bits(f.update.Aggression) != 0x3f000000 || math.Float32bits(f.update.SightRange) != 0x437a0000 {
			t.Fatal("group Guard did not preserve member order, nil/player no-ops, or repeated native orders")
		}
	})
}
