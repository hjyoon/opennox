package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/balance"
	noxflags "github.com/opennox/opennox/v1/common/flags"
)

func TestArmorCreate54C950NativeLayouts(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantTypeInd := uintptr(4)
	wantHealth := uintptr(556)
	wantModifierSize := uintptr(88)
	wantDurability := uintptr(52)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantTypeInd = 8
		wantHealth = 616
		wantModifierSize = 112
		wantDurability = 64
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.TypeInd", unsafe.Offsetof(Object{}.TypeInd), wantTypeInd},
		{"Object.HealthData", unsafe.Offsetof(Object{}.HealthData), wantHealth},
		{"Modifier size", unsafe.Sizeof(Modifier{}), wantModifierSize},
		{"Modifier.Durability52", unsafe.Offsetof(Modifier{}.Durability52), wantDurability},
		{"HealthData size", unsafe.Sizeof(HealthData{}), 20},
		{"HealthData.Cur", unsafe.Offsetof(HealthData{}.Cur), 0},
		{"HealthData.Max", unsafe.Offsetof(HealthData{}.Max), 4},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestArmorCreateNative54C950UsesNativeObjectAndModifier(t *testing.T) {
	health := &HealthData{
		Cur:     0x1111,
		Field2:  0x2222,
		Max:     0x3333,
		field6:  0x4444,
		field8:  0x55555555,
		field12: 0x66666666,
		Field16: 0x77777777,
	}
	obj := &Object{TypeInd: 0x3456, Field1_2: 0x789a, HealthData: health}
	definition := &Modifier{
		TypeInd:              0x3456,
		Effectiveness36:      0x36363636,
		Durability52:         0xabcd5678,
		Field56:              0x56565656,
		DamageCoeffOrArmor64: 64.5,
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for index, pointer := range []unsafe.Pointer{
			unsafe.Pointer(obj), unsafe.Pointer(health), unsafe.Pointer(definition),
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("pointer %d on %s/%s = %p, want native high address", index, runtime.GOOS, runtime.GOARCH, pointer)
			}
		}
	}
	findCalls := 0
	flagCalls := 0
	armorCreateNative54C950(obj, armorCreateNativeDeps54C950{
		findDefinition: func(typeInd uint16) *Modifier {
			findCalls++
			if typeInd != obj.TypeInd {
				t.Fatalf("type = %#x, want %#x", typeInd, obj.TypeInd)
			}
			return definition
		},
		gameFlag: func(flag uint32) int32 {
			flagCalls++
			if flag != armorCreateQuestFlag54C950 {
				t.Fatalf("flag = %#x, want %#x", flag, armorCreateQuestFlag54C950)
			}
			return 0
		},
		loadBalance: func(string) float32 {
			t.Fatal("non-Quest native path loaded balance")
			return 0
		},
		floatToInt: armorCreateRoundFloat32ToInt32_54C950,
	})
	if findCalls != 1 || flagCalls != 1 {
		t.Fatalf("lookup/flag calls = %d/%d, want 1/1", findCalls, flagCalls)
	}
	if health.Cur != 0x5678 || health.Max != 0x5678 {
		t.Fatalf("durability = %#x/%#x, want 0x5678/0x5678", health.Cur, health.Max)
	}
	if health.Field2 != 0x2222 || health.field6 != 0x4444 || health.field8 != 0x55555555 ||
		health.field12 != 0x66666666 || health.Field16 != 0x77777777 {
		t.Fatalf("neighboring health fields changed: %+v", health)
	}
	if obj.TypeInd != 0x3456 || obj.Field1_2 != 0x789a || obj.HealthData != health {
		t.Fatalf("neighboring object fields changed: Type=%#x Field=%#x Health=%p", obj.TypeInd, obj.Field1_2, obj.HealthData)
	}
	if definition.Durability52 != 0xabcd5678 || definition.Effectiveness36 != 0x36363636 ||
		definition.Field56 != 0x56565656 || definition.DamageCoeffOrArmor64 != 64.5 {
		t.Fatalf("definition changed: %+v", definition)
	}
	runtime.KeepAlive(obj)
	runtime.KeepAlive(health)
	runtime.KeepAlive(definition)
}

func TestServerArmorCreate54C950BindsQuestMultiplier(t *testing.T) {
	oldFlags := noxflags.GetGame()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameFlag(armorCreateQuestFlag54C950))

	srv := new(Server)
	definition := &Modifier{TypeInd: 0x45, Durability52: 3}
	srv.Modif.Dword_5d4594_251608 = definition
	srv.Balance.file = &balance.File{
		Global: balance.Config{
			"questdurabilitymultiplier": balance.Float(1.5),
		},
		Tags: make(map[balance.Tag]balance.Config),
	}
	health := &HealthData{Field2: 0xaaaa, field6: 0xbbbb}
	obj := &Object{TypeInd: 0x45, HealthData: health}

	srv.ArmorCreate54C950(obj)
	if health.Cur != 4 || health.Max != 4 {
		t.Fatalf("Quest durability = %d/%d, want ties-to-even 4/4", health.Cur, health.Max)
	}
	if health.Field2 != 0xaaaa || health.field6 != 0xbbbb {
		t.Fatalf("neighboring fields = %#x/%#x", health.Field2, health.field6)
	}
}
