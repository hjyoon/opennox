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

func TestRewardArmorNativeLayouts4F0E80(t *testing.T) {
	wantInitSize := uintptr(20)
	wantField16 := uintptr(16)
	wantModifierSize := uintptr(144)
	wantAllowArmor := uintptr(32)
	wantAllowPos := uintptr(36)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantInitSize = 40
		wantField16 = 32
		wantModifierSize = 208
		wantAllowArmor = 52
		wantAllowPos = 56
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
		{"ModifierEff.AllowArmor32", unsafe.Offsetof(ModifierEff{}.AllowArmor32), wantAllowArmor},
		{"ModifierEff.AllowPos36", unsafe.Offsetof(ModifierEff{}.AllowPos36), wantAllowPos},
		{"reward object TypeInd", unsafe.Sizeof(rewardObjectDefinition4F0640{}.TypeInd), 4},
		{"reward modifier pointer", unsafe.Sizeof(rewardModifierDefinition4F0640{}.Modifier), unsafe.Sizeof(uintptr(0))},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestRewardArmorNativeUsesNativeModifierPointers4F0E80(t *testing.T) {
	quality := &ModifierEff{AllowArmor32: 0x40}
	material := &ModifierEff{AllowArmor32: 0x40}
	first := &ModifierEff{AllowArmor32: 0x40, AllowPos36: 1}
	second := &ModifierEff{AllowArmor32: 0x40, AllowPos36: 2}
	created := &Object{}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		pointers := []unsafe.Pointer{
			unsafe.Pointer(quality), unsafe.Pointer(material), unsafe.Pointer(first),
			unsafe.Pointer(second), unsafe.Pointer(created),
		}
		for index, pointer := range pointers {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("pointer %d does not exercise a native high address: %p", index, pointer)
			}
		}
	}

	tables := rewardArmorNativeTables4F0E80{
		objects: []rewardObjectDefinition4F0640{
			{Weight: 1, Name: "Armor", TypeInd: 7, Kind: 2, Slots: 16}, {},
		},
		armorQuality: []rewardModifierDefinition4F0640{
			{Name: "Quality", Modifier: quality, Slots: 16}, {},
		},
		material: []rewardModifierDefinition4F0640{
			{Name: "Material", Modifier: material, Slots: 16}, {},
		},
		enchantments: []rewardModifierDefinition4F0640{
			{Group: 1, Name: "First", Modifier: first, Slots: 16},
			{Group: 2, Name: "Second", Modifier: second, Slots: 4},
			{},
		},
	}
	wantBounds := [][2]int32{{0, 0}, {2, 4}, {0, 0}, {0, 0}, {0, 0}, {3, 4}, {0, 0}}
	results := []int32{0, 4, 0, 0, 0, 4, 0}
	var events []string
	call := 0
	var applied ModifierInitData
	got := rewardArmorNative4F0E80(8, rewardArmorNativeDeps4F0E80{
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
		armorTypeMask: func(typeInd uint32) uint32 {
			events = append(events, "armor-mask")
			if typeInd != 7 {
				t.Fatalf("armor-mask type = %d, want 7", typeInd)
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
		applyModifiers: func(object *Object, attrs *ModifierInitData) {
			events = append(events, "apply")
			if object != created {
				t.Fatalf("apply object = %p, want %p", object, created)
			}
			applied = *attrs
		},
	})
	if got != created || applied.Modifiers != [4]*ModifierEff{quality, material, first, second} || applied.Field16 != 0 {
		t.Fatalf("result/attributes = %p/%#v, want native modifier identities and zero Field16", got, applied)
	}
	if call != len(wantBounds) {
		t.Fatalf("RNG calls = %d, want %d", call, len(wantBounds))
	}
	wantPrefix := []string{"slots", "allowed", "rng", "allowed", "armor-mask", "create"}
	if len(events) < len(wantPrefix) || !slices.Equal(events[:len(wantPrefix)], wantPrefix) || events[len(events)-1] != "apply" {
		t.Fatalf("events = %v, want prefix %v and final apply", events, wantPrefix)
	}
}

func TestRewardArmorNativeNilModifierFaultBoundary4F0E80(t *testing.T) {
	created := &Object{}
	createCalls := 0
	applyCalls := 0
	rngResults := []int32{0, 3}
	defer func() {
		fault := recover()
		if fault == nil || createCalls != 1 || applyCalls != 0 || len(rngResults) != 0 {
			t.Fatalf("recover/create/apply/RNG = %v/%d/%d/%d, want panic/1/0/0", fault, createCalls, applyCalls, len(rngResults))
		}
	}()
	rewardArmorNative4F0E80(6, rewardArmorNativeDeps4F0E80{
		tables: rewardArmorNativeTables4F0E80{
			objects: []rewardObjectDefinition4F0640{
				{Weight: 1, Name: "Armor", TypeInd: 7, Kind: 2, Slots: 8}, {},
			},
			armorQuality: []rewardModifierDefinition4F0640{{Name: "Unresolved", Slots: 8}, {}},
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
		armorTypeMask:     func(uint32) uint32 { return 0x40 },
		createObject: func(uint32) *Object {
			createCalls++
			return created
		},
		applyModifiers: func(*Object, *ModifierInitData) { applyCalls++ },
	})
}

func TestRewardArmorModifierApplyTeamBaseCache4E4990(t *testing.T) {
	srv := new(Server)
	srv.Types.byID = map[string]*ObjectType{"teambase": {ind: 9}}
	modifier := &ModifierEff{}
	attrs := &ModifierInitData{Modifiers: [4]*ModifierEff{modifier}}
	newTeamBase := func(typeInd uint16) (*Object, *ModifierInitData) {
		destination := new(ModifierInitData)
		return &Object{
			ObjClass: object.ClassPlayer,
			TypeInd:  typeInd,
			InitData: unsafe.Pointer(destination),
		}, destination
	}
	firstObject, firstDestination := newTeamBase(9)
	if !srv.applyModifierAttrs4E4990(firstObject, attrs) || firstDestination.Modifiers[0] != modifier {
		t.Fatal("first TeamBase modifier application failed")
	}
	if srv.Modif.teamBaseTypeInd4E4990 != 9 {
		t.Fatalf("cached TeamBase = %d, want 9", srv.Modif.teamBaseTypeInd4E4990)
	}
	srv.Types.byID["teambase"] = &ObjectType{ind: 10}
	secondObject, secondDestination := newTeamBase(9)
	if !srv.applyModifierAttrs4E4990(secondObject, attrs) || secondDestination.Modifiers[0] != modifier {
		t.Fatal("cached TeamBase ID was not reused")
	}

	emptyServer := new(Server)
	emptyObject := &Object{ObjClass: object.ClassWeapon}
	if emptyServer.applyModifierAttrs4E4990(emptyObject, &ModifierInitData{}) || emptyServer.Modif.teamBaseTypeInd4E4990 != 0 {
		t.Fatal("empty ordinary attributes reached TeamBase lookup or application")
	}

	retryServer := new(Server)
	newForcedWand := func() *Object {
		return &Object{
			ObjClass:    object.ClassWand | object.ClassPlayer,
			ObjSubClass: object.SubClass(0x00010000),
			InitData:    unsafe.Pointer(new(ModifierInitData)),
		}
	}
	if !retryServer.applyModifierAttrs4E4990(newForcedWand(), &ModifierInitData{}) || retryServer.Modif.teamBaseTypeInd4E4990 != 0 {
		t.Fatal("forced empty attributes did not apply with a retryable zero TeamBase lookup")
	}
	retryServer.Types.byID = map[string]*ObjectType{"teambase": {ind: 11}}
	if !retryServer.applyModifierAttrs4E4990(newForcedWand(), &ModifierInitData{}) || retryServer.Modif.teamBaseTypeInd4E4990 != 11 {
		t.Fatal("zero TeamBase lookup was not retried and cached")
	}
}

func TestRewardArmorModifierApplyFaultBoundaries4E4990(t *testing.T) {
	t.Run("nil object faults before attributes and TeamBase", func(t *testing.T) {
		srv := new(Server)
		srv.Types.byID = map[string]*ObjectType{"teambase": {ind: 9}}
		defer func() {
			if recover() == nil || srv.Modif.teamBaseTypeInd4E4990 != 0 {
				t.Fatalf("nil object did not fault before cache: cache=%d", srv.Modif.teamBaseTypeInd4E4990)
			}
		}()
		srv.applyModifierAttrs4E4990(nil, &ModifierInitData{})
	})

	t.Run("nil ordinary attributes fault before TeamBase", func(t *testing.T) {
		srv := new(Server)
		srv.Types.byID = map[string]*ObjectType{"teambase": {ind: 9}}
		defer func() {
			if recover() == nil || srv.Modif.teamBaseTypeInd4E4990 != 0 {
				t.Fatalf("nil attributes did not fault before cache: cache=%d", srv.Modif.teamBaseTypeInd4E4990)
			}
		}()
		srv.applyModifierAttrs4E4990(&Object{ObjClass: object.ClassWeapon}, nil)
	})

	t.Run("nil forced attributes fault after TeamBase and NeedSync", func(t *testing.T) {
		srv := new(Server)
		srv.Types.byID = map[string]*ObjectType{"teambase": {ind: 9}}
		obj := &Object{
			ObjClass:    object.ClassWand | object.ClassPlayer,
			ObjSubClass: object.SubClass(0x00010000),
			InitData:    unsafe.Pointer(new(ModifierInitData)),
		}
		defer func() {
			if recover() == nil || srv.Modif.teamBaseTypeInd4E4990 != 9 || obj.Field38 != math.MaxUint32 {
				t.Fatalf("forced fault cache/sync = %d/%#x, want 9/MaxUint32", srv.Modif.teamBaseTypeInd4E4990, obj.Field38)
			}
		}()
		srv.applyModifierAttrs4E4990(obj, nil)
	})
}

func TestRewardArmorServerUsesLogicRNGAndObjectFactory4F0E80(t *testing.T) {
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
	armorType := &ObjectType{
		s:            &srv.Types,
		ind:          1,
		id:           "Armor",
		allowed:      true,
		class:        object.ClassArmor,
		flags:        object.FlagNoCollide,
		InitData:     unsafe.Pointer(templateInit),
		InitDataSize: unsafe.Sizeof(*templateInit),
	}
	srv.Types.byID = map[string]*ObjectType{"armor": armorType}
	srv.Types.byInd = []*ObjectType{nil, armorType}
	srv.Armor.table[0] = armorRecord{TypeInd: 1, Bit: 0x40}
	srv.rewardDefinitions.Objects = [58]rewardObjectDefinition4F0640{
		{Weight: 1, Name: "Armor", TypeInd: 1, Kind: 2, Slots: 1},
		{},
	}

	beforeRNG := srv.Rand.Logic.Index()
	got := srv.RewardArmor4F0E80(nil, 0)
	if got == nil || got.TypeInd != 1 || !got.Class().Has(object.ClassArmor) {
		t.Fatalf("server result = %#v, want native armor type 1", got)
	}
	if got.InitData == nil || (*ModifierInitData)(got.InitData).HasModifiers() {
		t.Fatalf("server modifier init data = %p/%#v, want allocated empty native data", got.InitData, (*ModifierInitData)(got.InitData))
	}
	if index, want := srv.Rand.Logic.Index(), (beforeRNG+1)%4096; index != want {
		t.Fatalf("logic RNG index = %d, want %d after object draw", index, want)
	}
}
