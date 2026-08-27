package server

import (
	"math"
	"runtime"
	"slices"
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/prand"
)

func TestRewardWeaponNativeLayouts4F14E0(t *testing.T) {
	wantInitSize := uintptr(20)
	wantField16 := uintptr(16)
	wantModifierSize := uintptr(144)
	wantAllowWeapons := uintptr(28)
	wantAllowPos := uintptr(36)
	wantObjectClass := uintptr(8)
	wantObjectSubClass := uintptr(12)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantInitSize = 40
		wantField16 = 32
		wantModifierSize = 208
		wantAllowWeapons = 48
		wantAllowPos = 56
		wantObjectClass = 12
		wantObjectSubClass = 16
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"ModifierInitData size", unsafe.Sizeof(ModifierInitData{}), wantInitSize},
		{"ModifierInitData.Modifiers", unsafe.Offsetof(ModifierInitData{}.Modifiers), 0},
		{"ModifierInitData.Field16", unsafe.Offsetof(ModifierInitData{}.Field16), wantField16},
		{"ModifierEff size", unsafe.Sizeof(ModifierEff{}), wantModifierSize},
		{"ModifierEff.AllowWeapons28", unsafe.Offsetof(ModifierEff{}.AllowWeapons28), wantAllowWeapons},
		{"ModifierEff.AllowPos36", unsafe.Offsetof(ModifierEff{}.AllowPos36), wantAllowPos},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantObjectClass},
		{"Object.ObjSubClass", unsafe.Offsetof(Object{}.ObjSubClass), wantObjectSubClass},
		{"reward object TypeInd", unsafe.Sizeof(rewardObjectDefinition4F0640{}.TypeInd), 4},
		{"reward modifier pointer", unsafe.Sizeof(rewardModifierDefinition4F0640{}.Modifier), unsafe.Sizeof(uintptr(0))},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestRewardWeaponNativeUsesNativeModifierPointers4F14E0(t *testing.T) {
	power := &ModifierEff{AllowWeapons28: 0x40}
	material := &ModifierEff{AllowWeapons28: 0x40}
	first := &ModifierEff{AllowWeapons28: 0x40, AllowPos36: 1}
	second := &ModifierEff{AllowWeapons28: 0x40, AllowPos36: 2}
	created := &Object{}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		pointers := []unsafe.Pointer{
			unsafe.Pointer(power), unsafe.Pointer(material), unsafe.Pointer(first),
			unsafe.Pointer(second), unsafe.Pointer(created),
		}
		for index, pointer := range pointers {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("pointer %d does not exercise a native high address: %p", index, pointer)
			}
		}
	}

	tables := rewardWeaponNativeTables4F14E0{
		objects: []rewardObjectDefinition4F0640{
			{Weight: 1, Name: "Weapon", TypeInd: 7, Kind: 1, Slots: 16}, {},
		},
		weaponPower: []rewardModifierDefinition4F0640{
			{Name: "Power", Modifier: power, Slots: 16}, {},
		},
		material: []rewardModifierDefinition4F0640{
			{Name: "Material", Modifier: material, Slots: 16}, {},
		},
		enchantments: []rewardModifierDefinition4F0640{
			{Group: 1, Name: "First", Modifier: first, Slots: 20},
			{Group: 2, Name: "Second", Modifier: second, Slots: 20},
			{},
		},
	}
	wantBounds := [][2]int32{{0, 0}, {2, 4}, {0, 0}, {0, 0}, {3, 4}, {0, 0}, {3, 4}, {0, 0}}
	results := []int32{0, 4, 0, 0, 4, 0, 4, 0}
	var events []string
	call := 0
	var applied ModifierInitData
	got := rewardWeaponNative4F14E0(8, rewardWeaponNativeDeps4F14E0{
		tables: tables,
		pickSlots: func(stage uint32) uint32 {
			events = append(events, "slots")
			if stage != 8 {
				t.Fatalf("stage = %d, want 8", stage)
			}
			return 16
		},
		randomInt: func(minimum, maximum int32) int32 {
			events = append(events, "rng")
			if call >= len(wantBounds) || [2]int32{minimum, maximum} != wantBounds[call] {
				t.Fatalf("RNG call %d bounds = %d..%d, want %v", call, minimum, maximum, wantBounds)
			}
			result := results[call]
			call++
			return result
		},
		objectTypeAllowed: func(typeInd uint32) bool {
			events = append(events, "allowed")
			return typeInd == 7
		},
		weaponTypeMask: func(typeInd uint32) uint32 {
			events = append(events, "weapon-mask")
			if typeInd != 7 {
				t.Fatalf("weapon-mask type = %d, want 7", typeInd)
			}
			return 0x40
		},
		createObject: func(typeInd uint32) *Object {
			events = append(events, "create")
			if typeInd != 7 {
				t.Fatalf("create type = %d, want 7", typeInd)
			}
			return created
		},
		loadReplenishment: func() *ModifierEff {
			t.Fatal("ordinary weapon loaded Replenishment1")
			return nil
		},
		applyModifiers: func(object *Object, attrs *ModifierInitData) {
			events = append(events, "apply")
			if object != created {
				t.Fatalf("apply object = %p, want %p", object, created)
			}
			applied = *attrs
		},
	})
	if got != created || applied.Modifiers != [4]*ModifierEff{power, material, first, second} || applied.Field16 != 0 {
		t.Fatalf("result/attributes = %p/%#v, want native modifier identities and zero Field16", got, applied)
	}
	if call != len(wantBounds) {
		t.Fatalf("RNG calls = %d, want %d", call, len(wantBounds))
	}
	wantPrefix := []string{"slots", "allowed", "rng", "allowed", "weapon-mask", "create"}
	if len(events) < len(wantPrefix) || !slices.Equal(events[:len(wantPrefix)], wantPrefix) || events[len(events)-1] != "apply" {
		t.Fatalf("events = %v, want prefix %v and final apply", events, wantPrefix)
	}
}

func TestRewardWeaponNativeReplenishingWand4F14E0(t *testing.T) {
	replenishment := &ModifierEff{}
	created := &Object{
		ObjClass:    object.ClassWand,
		ObjSubClass: object.SubClass(0x00010000),
	}
	var applied ModifierInitData
	loadCalls := 0
	got := rewardWeaponNative4F14E0(1, rewardWeaponNativeDeps4F14E0{
		tables: rewardWeaponNativeTables4F14E0{
			objects:      []rewardObjectDefinition4F0640{{Weight: 1, Name: "#Wand", TypeInd: 7, Kind: 1, Slots: 1}, {}},
			weaponPower:  []rewardModifierDefinition4F0640{{}},
			material:     []rewardModifierDefinition4F0640{{}},
			enchantments: []rewardModifierDefinition4F0640{{}},
		},
		pickSlots:         func(uint32) uint32 { return 1 },
		randomInt:         func(minimum, _ int32) int32 { return minimum },
		objectTypeAllowed: func(uint32) bool { return true },
		weaponTypeMask:    func(uint32) uint32 { return 0x40 },
		createObject:      func(uint32) *Object { return created },
		loadReplenishment: func() *ModifierEff {
			loadCalls++
			return replenishment
		},
		applyModifiers: func(object *Object, attrs *ModifierInitData) {
			if object != created {
				t.Fatalf("apply object = %p, want %p", object, created)
			}
			applied = *attrs
		},
	})
	if got != created || loadCalls != 1 || applied.Modifiers != [4]*ModifierEff{nil, nil, replenishment, nil} {
		t.Fatalf("result/load/attributes = %p/%d/%#v", got, loadCalls, applied)
	}
}

func TestRewardWeaponNativeNilModifierFaultBoundary4F14E0(t *testing.T) {
	created := &Object{}
	createCalls := 0
	applyCalls := 0
	rngResults := []int32{0, 1, 21}
	defer func() {
		fault := recover()
		if fault == nil || createCalls != 1 || applyCalls != 0 || len(rngResults) != 0 {
			t.Fatalf("recover/create/apply/RNG = %v/%d/%d/%d, want panic/1/0/0", fault, createCalls, applyCalls, len(rngResults))
		}
	}()
	rewardWeaponNative4F14E0(6, rewardWeaponNativeDeps4F14E0{
		tables: rewardWeaponNativeTables4F14E0{
			objects: []rewardObjectDefinition4F0640{
				{Weight: 1, Name: "Weapon", TypeInd: 7, Kind: 1, Slots: 8}, {},
			},
			weaponPower:  []rewardModifierDefinition4F0640{{Name: "Unresolved", Slots: 8}, {}},
			material:     []rewardModifierDefinition4F0640{{}},
			enchantments: []rewardModifierDefinition4F0640{{}},
		},
		pickSlots: func(uint32) uint32 { return 8 },
		randomInt: func(int32, int32) int32 {
			result := rngResults[0]
			rngResults = rngResults[1:]
			return result
		},
		objectTypeAllowed: func(uint32) bool { return true },
		weaponTypeMask:    func(uint32) uint32 { return 0x40 },
		createObject: func(uint32) *Object {
			createCalls++
			return created
		},
		loadReplenishment: func() *ModifierEff { return nil },
		applyModifiers:    func(*Object, *ModifierInitData) { applyCalls++ },
	})
}

func TestRewardWeaponServerUsesLogicRNGAndObjectFactory4F14E0(t *testing.T) {
	srv := new(Server)
	srv.handle = atomic.AddUintptr(&serverLast, 1)
	servers.Store(srv.handle, srv)
	t.Cleanup(func() { servers.Delete(srv.handle) })
	srv.Rand.Logic = prand.New(2011)
	srv.Objs.init(srv.handle)
	if !srv.Objs.Init(2) {
		t.Fatal("object allocator initialization failed")
	}
	t.Cleanup(srv.Objs.FreeObjects)

	templateInit := &ModifierInitData{}
	weaponType := &ObjectType{
		s:            &srv.Types,
		ind:          1,
		id:           "Weapon",
		allowed:      true,
		class:        object.ClassWeapon,
		flags:        object.FlagNoCollide,
		InitData:     unsafe.Pointer(templateInit),
		InitDataSize: unsafe.Sizeof(*templateInit),
	}
	srv.Types.byID = map[string]*ObjectType{"weapon": weaponType}
	srv.Types.byInd = []*ObjectType{nil, weaponType}
	srv.Weapons.table[0] = weaponRecord{TypeInd: 1, Bit: 0x40}
	srv.rewardDefinitions.Objects = [58]rewardObjectDefinition4F0640{
		{Weight: 1, Name: "Weapon", TypeInd: 1, Kind: 1, Slots: 1},
		{},
	}

	marker := &Object{NetCode: 0xa5a5a5a5}
	beforeRNG := srv.Rand.Logic.Index()
	got := srv.RewardWeapon4F14E0(marker, 0)
	if got == nil || got.TypeInd != 1 || !got.Class().Has(object.ClassWeapon) {
		t.Fatalf("server result = %#v, want native weapon type 1", got)
	}
	if got.InitData == nil || (*ModifierInitData)(got.InitData).HasModifiers() {
		t.Fatalf("server modifier init data = %p/%#v, want allocated empty native data", got.InitData, (*ModifierInitData)(got.InitData))
	}
	if marker.NetCode != 0xa5a5a5a5 {
		t.Fatal("ignored marker was modified")
	}
	if index, want := srv.Rand.Logic.Index(), (beforeRNG+1)%4096; index != want {
		t.Fatalf("logic RNG index = %d, want %d after object draw", index, want)
	}
}

func TestRewardWeaponServerResolvesReplenishment4F14E0(t *testing.T) {
	srv := new(Server)
	srv.handle = atomic.AddUintptr(&serverLast, 1)
	servers.Store(srv.handle, srv)
	t.Cleanup(func() { servers.Delete(srv.handle) })
	srv.Rand.Logic = prand.New(2011)
	srv.Objs.init(srv.handle)
	if !srv.Objs.Init(1) {
		t.Fatal("object allocator initialization failed")
	}
	t.Cleanup(srv.Objs.FreeObjects)

	templateInit := &ModifierInitData{}
	wandType := &ObjectType{
		s:            &srv.Types,
		ind:          1,
		id:           "Wand",
		allowed:      true,
		class:        object.ClassWand,
		subclass:     object.SubClass(0x00010000),
		flags:        object.FlagNoCollide,
		InitData:     unsafe.Pointer(templateInit),
		InitDataSize: unsafe.Sizeof(*templateInit),
	}
	srv.Types.byID = map[string]*ObjectType{"wand": wandType}
	srv.Types.byInd = []*ObjectType{nil, wandType}
	srv.Weapons.table[0] = weaponRecord{TypeInd: 1, Bit: 0x40}
	srv.rewardDefinitions.Objects = [58]rewardObjectDefinition4F0640{
		{Weight: 1, Name: "#Wand", TypeInd: 1, Kind: 1, Slots: 1},
		{},
	}
	name := append([]byte("Replenishment1"), 0)
	replenishment := &ModifierEff{name0: &name[0], ind4: 7}
	srv.Modif.types[0] = replenishment

	got := srv.RewardWeapon4F14E0(nil, 0)
	if got == nil || !got.Class().Has(object.ClassWand) {
		t.Fatalf("server result = %#v, want native wand", got)
	}
	attrs := got.InitDataModifier()
	if attrs.Modifiers != [4]*ModifierEff{nil, nil, replenishment, nil} || attrs.Field16 != 0 {
		t.Fatalf("wand attributes = %#v, want native Replenishment1 in slot 2", *attrs)
	}
	if got.Field38 != math.MaxUint32 {
		t.Fatalf("wand sync field = %#x, want MaxUint32", got.Field38)
	}
	runtime.KeepAlive(name)
}
