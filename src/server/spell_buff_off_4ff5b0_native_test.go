package server

import (
	"fmt"
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestSpellBuffOffNative4FF5B0ObjectLayout(t *testing.T) {
	wantBuffs, wantDuration, wantPower := uintptr(340), uintptr(344), uintptr(408)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantBuffs, wantDuration, wantPower = 344, 348, 412
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Buffs offset", unsafe.Offsetof(Object{}.Buffs), wantBuffs},
		{"Buffs width", unsafe.Sizeof(Object{}.Buffs), 4},
		{"BuffsDur offset", unsafe.Offsetof(Object{}.BuffsDur), wantDuration},
		{"BuffsDur width", unsafe.Sizeof(Object{}.BuffsDur), 64},
		{"BuffsDur element width", unsafe.Sizeof(Object{}.BuffsDur[0]), 2},
		{"BuffsPower offset", unsafe.Offsetof(Object{}.BuffsPower), wantPower},
		{"BuffsPower width", unsafe.Sizeof(Object{}.BuffsPower), 32},
		{"BuffsPower element width", unsafe.Sizeof(Object{}.BuffsPower[0]), 1},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("Object %s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func spellBuffOffNativeTestDeps4FF5B0() spellBuffOffNativeDeps4FF5B0 {
	return spellBuffOffNativeDeps4FF5B0{
		loadBuffArg:   func(buff int32) int32 { return buff },
		loadUnitArg:   func(unit *Object) *Object { return unit },
		loadBuffs:     func(unit *Object) uint32 { return unit.Buffs },
		setBuffFlags:  func(unit *Object, flags uint32) { unit.Buffs = flags },
		storeDuration: func(unit *Object, buff int32, value uint16) { unit.BuffsDur[int(buff)] = value },
		storePower:    func(unit *Object, buff int32, value uint8) { unit.BuffsPower[int(buff)] = value },
		enchantSpell:  func(int32) int32 { return 0 },
		spellAudio:    func(int32, int32) int32 { return 0 },
		audio:         func(int32, *Object, int32, int32) {},
	}
}

func TestSpellBuffOffNative4FF5B0PreservesPointerWidthsAndAccessorOrder(t *testing.T) {
	unit, freeUnit := alloc.New(Object{})
	t.Cleanup(freeUnit)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}
	unit.Buffs = uint32(1) << 31
	unit.BuffsDur[31] = math.MaxUint16
	unit.BuffsPower[31] = math.MaxUint8

	var events []string
	deps := spellBuffOffNativeTestDeps4FF5B0()
	deps.loadBuffArg = func(buff int32) int32 {
		events = append(events, fmt.Sprintf("buff-%08x", uint32(buff)))
		return buff
	}
	deps.loadUnitArg = func(got *Object) *Object {
		events = append(events, "unit")
		if got != unit {
			t.Fatalf("unit = %p, want full native identity %p", got, unit)
		}
		return got
	}
	deps.loadBuffs = func(got *Object) uint32 {
		events = append(events, "buffs")
		if got != unit {
			t.Fatalf("LoadBuffs unit = %p, want %p", got, unit)
		}
		return got.Buffs
	}
	deps.setBuffFlags = func(got *Object, flags uint32) {
		events = append(events, "set")
		if got != unit || flags != 0 {
			t.Fatalf("SetBuffFlags = (%p, %#x), want (%p, 0)", got, flags, unit)
		}
		got.Buffs = flags
	}
	deps.storeDuration = func(got *Object, buff int32, value uint16) {
		events = append(events, "duration")
		if got != unit || buff != 31 || value != 0 {
			t.Fatalf("StoreDuration = (%p, %d, %#x), want (%p, 31, 0)", got, buff, value, unit)
		}
		got.BuffsDur[buff] = value
	}
	deps.storePower = func(got *Object, buff int32, value uint8) {
		events = append(events, "power")
		if got != unit || buff != 31 || value != 0 {
			t.Fatalf("StorePower = (%p, %d, %#x), want (%p, 31, 0)", got, buff, value, unit)
		}
		got.BuffsPower[buff] = value
	}
	deps.enchantSpell = func(buff int32) int32 {
		events = append(events, "spell")
		if buff != 31 {
			t.Fatalf("EnchantSpell buff = %d, want 31", buff)
		}
		return math.MinInt32
	}
	deps.spellAudio = func(spellID, selector int32) int32 {
		events = append(events, "spell-audio")
		if spellID != math.MinInt32 || selector != 2 {
			t.Fatalf("SpellAudio = (%#x, %d), want (%#x, 2)", spellID, selector, math.MinInt32)
		}
		return math.MaxInt32
	}
	deps.audio = func(id int32, got *Object, kind, code int32) {
		events = append(events, "audio")
		if id != math.MaxInt32 || got != unit || kind != 0 || code != 0 {
			t.Fatalf("Audio = (%#x, %p, %d, %d), want exact native arguments", id, got, kind, code)
		}
	}

	if got := spellBuffOffNative4FF5B0(unit, 31, deps); got != 0 {
		t.Fatalf("result = %#x, want zero", got)
	}
	want := []string{"buff-0000001f", "unit", "buffs", "set", "duration", "power", "spell", "spell-audio", "audio"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %q, want exact accessor order %q", events, want)
	}
	if unit.Buffs != 0 || unit.BuffsDur[31] != 0 || unit.BuffsPower[31] != 0 {
		t.Fatal("native slot 31 was not cleared")
	}
	runtime.KeepAlive(unit)
}

func TestSpellBuffOff4FF5B0ServerBinding(t *testing.T) {
	var s Server
	spellID := ENCHANT_LIGHT.Spell()
	s.Spells.byID = map[spell.ID]*SpellDef{
		spellID: {OffSound: sound.ID(987)},
	}
	unit := &Object{
		ObjClass: object.ClassClientPersist,
		Buffs:    uint32(1) << ENCHANT_LIGHT,
	}
	unit.BuffsDur[ENCHANT_LIGHT] = 0x9876
	unit.BuffsPower[ENCHANT_LIGHT] = 0xab
	for i := range unit.Field140 {
		unit.Field140[i] = 0xa5000080 | uint32(i<<12)
	}

	if got := s.SpellBuffOff4FF5B0(unit, int32(ENCHANT_LIGHT), SpellBuffOffRuntime4FF5B0{}); got != 0 {
		t.Fatalf("result = %#x, want zero", got)
	}
	if unit.Buffs != 0 || unit.BuffsDur[ENCHANT_LIGHT] != 0 || unit.BuffsPower[ENCHANT_LIGHT] != 0 {
		t.Fatalf("native buff state = %#x/%#x/%#x, want zero", unit.Buffs, unit.BuffsDur[ENCHANT_LIGHT], unit.BuffsPower[ENCHANT_LIGHT])
	}
	if unit.Field38 != math.MaxUint32 {
		t.Fatalf("sync marker = %#08x, want %#08x", unit.Field38, uint32(math.MaxUint32))
	}
	for i, value := range unit.Field140 {
		want := (uint32(0xa5000080)|uint32(i<<12))&0xfffff000 | 0x800000
		if value != want {
			t.Fatalf("Field140[%d] = %#08x, want %#08x", i, value, want)
		}
	}
	if len(s.Audio.delayedObj) != 1 {
		t.Fatalf("queued audio events = %d, want 1", len(s.Audio.delayedObj))
	}
	event := s.Audio.delayedObj[0]
	if event.ID != 987 || event.Obj != unit || event.Kind != 0 || event.Code != 0 {
		t.Fatalf("queued audio = %#v, want off sound 987 for %p", event, unit)
	}
}

func TestSpellBuffOff4FF5B0PlayerProtectionBoundary(t *testing.T) {
	player := &Player{ProtUnitBuffs: 0x13579bdf}
	update := &PlayerUpdateData{Player: player}
	unit := &Object{
		ObjClass:   object.ClassPlayer,
		UpdateData: unsafe.Pointer(update),
		Buffs:      uint32(1) << spellBuffOffNoAudio30_4FF5B0,
	}
	unit.BuffsDur[spellBuffOffNoAudio30_4FF5B0] = 0xffff
	unit.BuffsPower[spellBuffOffNoAudio30_4FF5B0] = 0xff
	for i := range unit.Field140 {
		unit.Field140[i] = 0xa5000080 | uint32(i)
	}
	beforeSync := unit.Field140

	calls := 0
	if got := (&Server{}).SpellBuffOff4FF5B0(unit, spellBuffOffNoAudio30_4FF5B0, SpellBuffOffRuntime4FF5B0{
		ResetPlayerProtection: func(got *Player, flags uint32) {
			calls++
			if got != player || flags != 0 {
				t.Fatalf("protection callback = (%p, %#x), want (%p, 0)", got, flags, player)
			}
			if unit.Field38 != math.MaxUint32 || unit.Buffs != 0 {
				t.Fatal("protection callback ran before the flag/sync writes")
			}
			if unit.BuffsDur[spellBuffOffNoAudio30_4FF5B0] != 0xffff || unit.BuffsPower[spellBuffOffNoAudio30_4FF5B0] != 0xff || unit.Field140 != beforeSync {
				t.Fatal("duration, power, or per-player sync stores preceded the protection callback")
			}
		},
	}); got != 0 {
		t.Fatalf("result = %#x, want zero", got)
	}
	if calls != 1 {
		t.Fatalf("protection callback count = %d, want 1", calls)
	}
	if unit.BuffsDur[spellBuffOffNoAudio30_4FF5B0] != 0 || unit.BuffsPower[spellBuffOffNoAudio30_4FF5B0] != 0 {
		t.Fatal("player buff storage was not cleared after protection callback")
	}
}

func TestSpellBuffOff4FF5B0InactiveReturnsMaskWithoutMutation(t *testing.T) {
	unit := &Object{Buffs: 0xa5a50000}
	unit.BuffsDur[0] = 0x1234
	unit.BuffsPower[0] = 0x56
	if got := (&Server{}).SpellBuffOff4FF5B0(unit, 32, SpellBuffOffRuntime4FF5B0{}); got != 1 {
		t.Fatalf("result = %#x, want masked bit 1", got)
	}
	if unit.Buffs != 0xa5a50000 || unit.BuffsDur[0] != 0x1234 || unit.BuffsPower[0] != 0x56 || unit.Field38 != 0 {
		t.Fatal("inactive enchant mutated native object state")
	}
}

func TestSpellBuffOff4FF5B0InvalidActiveSlotStopsAfterFlagClear(t *testing.T) {
	unit := &Object{ObjClass: object.ClassClientPersist, Buffs: 1}
	unit.BuffsDur[0] = 0x1234
	unit.BuffsPower[0] = 0x56
	defer func() {
		if recover() == nil {
			t.Fatal("active buff slot 32 did not stop at a checked native array boundary")
		}
		if unit.Buffs != 0 || unit.Field38 != math.MaxUint32 {
			t.Fatal("original SetBuffFlags prefix did not complete before bounds failure")
		}
		if unit.BuffsDur[0] != 0x1234 || unit.BuffsPower[0] != 0x56 {
			t.Fatal("invalid slot corrupted native buff arrays")
		}
	}()
	(&Server{}).SpellBuffOff4FF5B0(unit, 32, SpellBuffOffRuntime4FF5B0{})
}

func TestSpellBuffOff4FF5B0NativeNilObjectFaults(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("nil native object returned instead of faulting at buff load")
		}
	}()
	(&Server{}).SpellBuffOff4FF5B0(nil, 0, SpellBuffOffRuntime4FF5B0{})
}
