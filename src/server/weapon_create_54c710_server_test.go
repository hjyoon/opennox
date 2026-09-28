package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestWeaponCreate54C710NativeLayouts(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantTypeInd := uintptr(4)
	wantClass := uintptr(8)
	wantSubclass := uintptr(12)
	wantHealth := uintptr(556)
	wantInitData := uintptr(692)
	wantUseData := uintptr(736)
	wantModifierSize := uintptr(88)
	wantDurability := uintptr(52)
	wantInitSize := uintptr(20)
	wantInitField16 := uintptr(16)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantTypeInd = 8
		wantClass = 12
		wantSubclass = 16
		wantHealth = 616
		wantInitData = 760
		wantUseData = 848
		wantModifierSize = 112
		wantDurability = 64
		wantInitSize = 40
		wantInitField16 = 32
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.TypeInd", unsafe.Offsetof(Object{}.TypeInd), wantTypeInd},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.ObjSubClass", unsafe.Offsetof(Object{}.ObjSubClass), wantSubclass},
		{"Object.HealthData", unsafe.Offsetof(Object{}.HealthData), wantHealth},
		{"Object.InitData", unsafe.Offsetof(Object{}.InitData), wantInitData},
		{"Object.UseData", unsafe.Offsetof(Object{}.UseData), wantUseData},
		{"Modifier size", unsafe.Sizeof(Modifier{}), wantModifierSize},
		{"Modifier.Durability52", unsafe.Offsetof(Modifier{}.Durability52), wantDurability},
		{"ModifierInitData size", unsafe.Sizeof(ModifierInitData{}), wantInitSize},
		{"ModifierInitData.Field16", unsafe.Offsetof(ModifierInitData{}.Field16), wantInitField16},
		{"AmmoUseData size", unsafe.Sizeof(AmmoUseData{}), 3},
		{"AmmoUseData.Charge0", unsafe.Offsetof(AmmoUseData{}.Charge0), 0},
		{"AmmoUseData.Charge1", unsafe.Offsetof(AmmoUseData{}.Charge1), 1},
		{"AmmoUseData.Field2", unsafe.Offsetof(AmmoUseData{}.Field2), 2},
		{"WandUseData size", unsafe.Sizeof(WandUseData{}), 116},
		{"WandUseData.Charge", unsafe.Offsetof(WandUseData{}.Charge), 108},
		{"WandUseData.MaxCharge", unsafe.Offsetof(WandUseData{}.MaxCharge), 109},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestWeaponCreateNative54C710UsesNativePointers(t *testing.T) {
	health := &HealthData{Cur: 0x1111, Field2: 0x2222, Max: 0x3333, field6: 0x4444}
	initData := &ModifierInitData{Field16: 0x51515151}
	useData := new(WandUseData)
	useData.Charge = 3
	useData.MaxCharge = 5
	obj := &Object{
		TypeInd:     12,
		Field1_2:    0x789a,
		ObjClass:    object.Class(weaponCreateAmmoClass54C710 | weaponCreateStaffClass54C710),
		ObjSubClass: object.SubClass(weaponCreateAmmoChargeMask54C710 | weaponCreateDoubleChargeMask54C710),
		HealthData:  health,
		InitData:    unsafe.Pointer(initData),
		UseData:     UseDataPtr{Ptr: unsafe.Pointer(useData)},
	}
	definition := &Modifier{TypeInd: 12, Durability52: 0xabcd0003}
	vampirism := new(ModifierEff)
	lightning := new(ModifierEff)
	cache := new(weaponCreateTypeCache54C710)

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for index, pointer := range []unsafe.Pointer{
			unsafe.Pointer(obj), unsafe.Pointer(health), unsafe.Pointer(initData),
			unsafe.Pointer(useData), unsafe.Pointer(definition), unsafe.Pointer(vampirism),
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("pointer %d on %s/%s = %p, want native high address", index, runtime.GOOS, runtime.GOARCH, pointer)
			}
		}
	}

	weaponCreateNative54C710(obj, weaponCreateNativeDeps54C710{
		cache: cache,
		findDefinition: func(typeInd uint16) *Modifier {
			if typeInd != 12 {
				t.Fatalf("definition type = %d, want 12", typeInd)
			}
			return definition
		},
		resolveObjectType: func(name string) uint32 {
			return map[string]uint32{
				weaponCreateHeartType54C710:     11,
				weaponCreateWierdlingType54C710: 12,
				weaponCreateOrbType54C710:       13,
			}[name]
		},
		modifierID: func(name string) int {
			switch name {
			case weaponCreateWierdlingModifier54C710:
				return 2
			case weaponCreateWierdlingSecond54C710:
				return 3
			default:
				t.Fatalf("modifier name = %q", name)
				return 0
			}
		},
		modifierDesc: func(id int) *ModifierEff {
			if id == 2 {
				return vampirism
			}
			if id == 3 {
				return lightning
			}
			t.Fatalf("modifier id = %d", id)
			return nil
		},
		gameFlag: func(flag uint32) int32 {
			if flag != weaponCreateQuestFlag54C710 {
				t.Fatalf("game flag = %#x", flag)
			}
			return 1
		},
		loadBalance: func(key string) float32 {
			switch key {
			case weaponCreateDurabilityKey54C710:
				return 1.5
			case weaponCreateAmmoQuestKey54C710:
				return 6.5
			case weaponCreateStaffChargeKey54C710:
				return 1.5
			default:
				t.Fatalf("balance key = %q", key)
				return 0
			}
		},
		floatToInt: armorCreateRoundFloat32ToInt32_54C950,
	})

	if *cache != (weaponCreateTypeCache54C710{heart: 11, wierdling: 12, orb: 13}) {
		t.Fatalf("type cache = %+v", *cache)
	}
	if health.Cur != 4 || health.Max != 4 {
		t.Fatalf("durability = %d/%d, want 4/4", health.Cur, health.Max)
	}
	if health.Field2 != 0x2222 || health.field6 != 0x4444 {
		t.Fatalf("neighboring health fields changed: %+v", health)
	}
	if initData.Modifiers[2] != vampirism || initData.Modifiers[3] != lightning || initData.Field16 != 0x51515151 {
		t.Fatalf("modifier init data = %+v", initData)
	}
	if raw := (*[116]byte)(unsafe.Pointer(useData)); raw[0] != 6 || raw[1] != 6 || raw[2] != 0 {
		t.Fatalf("ammo bytes = %v, want [6 6 0]", raw[:3])
	}
	if useData.Charge != 9 || useData.MaxCharge != 15 {
		t.Fatalf("staff charge = %d/%d, want 9/15", useData.Charge, useData.MaxCharge)
	}
	if obj.TypeInd != 12 || obj.Field1_2 != 0x789a || obj.InitData != unsafe.Pointer(initData) || obj.UseData.Ptr != unsafe.Pointer(useData) {
		t.Fatalf("neighboring object fields changed")
	}
	if definition.Durability52 != 0xabcd0003 {
		t.Fatalf("definition durability changed: %#x", definition.Durability52)
	}
	runtime.KeepAlive(obj)
	runtime.KeepAlive(health)
	runtime.KeepAlive(initData)
	runtime.KeepAlive(useData)
	runtime.KeepAlive(definition)
}
