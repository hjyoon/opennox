package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestBuffApplyNative4FF380ObjectLayout(t *testing.T) {
	wantType, wantClass, wantSubclass, wantFlags := uintptr(4), uintptr(8), uintptr(12), uintptr(16)
	wantBuffs, wantDuration, wantPower := uintptr(340), uintptr(344), uintptr(408)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantType, wantClass, wantSubclass, wantFlags = 8, 12, 16, 20
		wantBuffs, wantDuration, wantPower = 344, 348, 412
	}
	fields := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"TypeInd", unsafe.Offsetof(Object{}.TypeInd), wantType},
		{"ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"ObjSubClass", unsafe.Offsetof(Object{}.ObjSubClass), wantSubclass},
		{"ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantFlags},
		{"Buffs", unsafe.Offsetof(Object{}.Buffs), wantBuffs},
		{"BuffsDur", unsafe.Offsetof(Object{}.BuffsDur), wantDuration},
		{"BuffsPower", unsafe.Offsetof(Object{}.BuffsPower), wantPower},
	}
	for _, field := range fields {
		if field.got != field.want {
			t.Errorf("Object.%s offset = %d, want %d", field.name, field.got, field.want)
		}
	}
	if unsafe.Sizeof(Object{}.TypeInd) != 2 || unsafe.Sizeof(Object{}.ObjClass) != 4 ||
		unsafe.Sizeof(Object{}.ObjSubClass) != 4 || unsafe.Sizeof(Object{}.ObjFlags) != 4 ||
		unsafe.Sizeof(Object{}.Buffs) != 4 || unsafe.Sizeof(Object{}.BuffsDur[0]) != 2 ||
		unsafe.Sizeof(Object{}.BuffsPower[0]) != 1 {
		t.Fatal("004FF380 native object fields lost their exact fixed widths")
	}
}

func buffApplyNativeTestDeps4FF380() buffApplyNativeDeps4FF380 {
	return buffApplyNativeDeps4FF380{
		loadHecubahTypeID:      func() uint32 { return 0x1234 },
		storeHecubahTypeID:     func(uint32) {},
		loadNecromancerTypeID:  func() uint32 { return 0x5678 },
		storeNecromancerTypeID: func(uint32) {},
		lookupTypeID:           func(string) uint32 { return 0 },
		gameFlag:               func(uint32) int32 { return 0 },
		audio:                  func(int32, *Object, int32, int32) {},
		buffOff:                func(*Object, EnchantID) int32 { return 0 },
		resetPlayerProtection:  func(*Player, uint32) {},
		enchantSpell:           func(int32) int32 { return 0 },
		spellAudio:             func(int32, int32) int32 { return 0 },
	}
}

func TestBuffApplyNative4FF380PreservesPointerWidthsAndLiveBuffReload(t *testing.T) {
	unit, freeUnit := alloc.New(Object{})
	t.Cleanup(freeUnit)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}
	unit.TypeInd = 0x9ABC
	unit.ObjClass = object.ClassClientPersist
	unit.Buffs = uint32(1) << 2
	for i := range unit.Field140 {
		unit.Field140[i] = 0xA5000080 | uint32(i<<12)
	}

	deps := buffApplyNativeTestDeps4FF380()
	var offCalls, audioCalls int
	deps.buffOff = func(got *Object, enchant EnchantID) int32 {
		offCalls++
		if got != unit || enchant != ENCHANT_INVISIBLE {
			t.Fatalf("BuffOff = (%p, %d), want (%p, %d)", got, enchant, unit, ENCHANT_INVISIBLE)
		}
		got.Buffs = uint32(1) << 4
		return -1
	}
	deps.enchantSpell = func(buff int32) int32 {
		if buff != 31 {
			t.Fatalf("EnchantSpell buff = %d, want 31", buff)
		}
		return 0x456
	}
	deps.spellAudio = func(spellID, selector int32) int32 {
		if spellID != 0x456 || selector != 1 {
			t.Fatalf("SpellAudio = (%#x, %d), want (0x456, 1)", spellID, selector)
		}
		return 0x789
	}
	deps.audio = func(id int32, got *Object, kind, code int32) {
		audioCalls++
		if id != 0x789 || got != unit || kind != 0 || code != 0 {
			t.Fatalf("Audio = (%#x, %p, %d, %d)", id, got, kind, code)
		}
	}

	buffApplyNative4FF380(unit, 31, -2, -3, deps)
	if offCalls != 1 || audioCalls != 1 {
		t.Fatalf("BuffOff/Audio calls = %d/%d, want 1/1", offCalls, audioCalls)
	}
	if unit.BuffsDur[31] != math.MaxUint16-1 || unit.BuffsPower[31] != math.MaxUint8-2 {
		t.Fatalf("duration/power = %#x/%#x, want exact low word/byte", unit.BuffsDur[31], unit.BuffsPower[31])
	}
	if unit.Buffs != uint32(1)<<4|uint32(1)<<31 {
		t.Fatalf("buffs = %#08x, want live post-BuffOff flags plus slot 31", unit.Buffs)
	}
	if unit.Field38 != math.MaxUint32 {
		t.Fatalf("sync marker = %#08x, want %#08x", unit.Field38, uint32(math.MaxUint32))
	}
	for i, value := range unit.Field140 {
		want := (uint32(0xA5000080)|uint32(i<<12))&0xFFFFF000 | 0x800000
		if value != want {
			t.Fatalf("Field140[%d] = %#08x, want %#08x", i, value, want)
		}
	}
	runtime.KeepAlive(unit)
}

func TestBuffApplyNative4FF380PlayerProtectionBoundary(t *testing.T) {
	player := &Player{ProtUnitBuffs: 0x13579BDF}
	update := &PlayerUpdateData{Player: player}
	unit := &Object{
		ObjClass:   object.ClassPlayer,
		UpdateData: unsafe.Pointer(update),
	}
	for i := range unit.Field140 {
		unit.Field140[i] = 0xA5000080 | uint32(i)
	}
	before := unit.Field140

	deps := buffApplyNativeTestDeps4FF380()
	calls := 0
	deps.resetPlayerProtection = func(got *Player, flags uint32) {
		calls++
		if got != player || flags != 1 {
			t.Fatalf("protection callback = (%p, %#x), want (%p, 1)", got, flags, player)
		}
		if unit.Field38 != math.MaxUint32 || unit.Buffs != 1 || unit.Field140 != before {
			t.Fatal("protection callback ran outside the original SetBuffFlags boundary")
		}
	}
	buffApplyNative4FF380(unit, 0, 7, 5, deps)
	if calls != 1 {
		t.Fatalf("protection callback count = %d, want 1", calls)
	}
}

func TestBuffApply4FF380ServerBindingUsesDedicatedTypeCachesBeforeNilGate(t *testing.T) {
	s := &Server{}
	s.Types.byID = map[string]*ObjectType{
		"hecubah":     {ind: 0x1234},
		"necromancer": {ind: 0x5678},
	}
	s.BuffApply4FF380(nil, math.MinInt32, -1, -1, BuffApplyRuntime4FF380{})
	if got := s.Types.buffApplyHecubahIDCached4FF380(); got != 0x1234 {
		t.Fatalf("Hecubah cache = %#x, want 0x1234", got)
	}
	if got := s.Types.buffApplyNecromancerIDCached4FF380(); got != 0x5678 {
		t.Fatalf("Necromancer cache = %#x, want 0x5678", got)
	}
	if s.Types.fast.hecubah != 0 || s.Types.fast.necrom != 0 {
		t.Fatal("004FF380 polluted the general Hecubah/Necromancer caches")
	}

	s.Types.byID["hecubah"] = &ObjectType{ind: 1}
	s.Types.byID["necromancer"] = &ObjectType{ind: 2}
	s.BuffApply4FF380(nil, math.MaxInt32, 1, 1, BuffApplyRuntime4FF380{})
	if s.Types.buffApplyHecubahIDCached4FF380() != 0x1234 ||
		s.Types.buffApplyNecromancerIDCached4FF380() != 0x5678 {
		t.Fatal("nonzero 004FF380 Hecubah cache did not skip both lookups")
	}
}

func TestBuffApplyNative4FF380InvalidSlotCannotCorruptObject(t *testing.T) {
	unit := &Object{ObjClass: object.ClassClientPersist}
	deps := buffApplyNativeTestDeps4FF380()
	offCalls := 0
	deps.buffOff = func(*Object, EnchantID) int32 {
		offCalls++
		return 0
	}
	defer func() {
		if recover() == nil {
			t.Fatal("buff slot 32 did not stop at a checked native array boundary")
		}
		if offCalls != 1 {
			t.Fatalf("BuffOff calls before bounds failure = %d, want 1", offCalls)
		}
		if unit.Buffs != 0 || unit.BuffsDur != [32]uint16{} || unit.BuffsPower != [32]uint8{} {
			t.Fatal("invalid slot mutated native object buff storage")
		}
	}()
	buffApplyNative4FF380(unit, 32, 7, 5, deps)
}
