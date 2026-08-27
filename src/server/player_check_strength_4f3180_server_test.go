package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestPlayerCheckStrength4F3180NativeLayout(t *testing.T) {
	checks32 := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.size", unsafe.Sizeof(Object{}), 780},
		{"Object.TypeInd", unsafe.Offsetof(Object{}.TypeInd), 4},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), 8},
		{"Modifier.size", unsafe.Sizeof(Modifier{}), 88},
		{"Modifier.TypeInd", unsafe.Offsetof(Modifier{}.TypeInd), 4},
		{"Modifier.ReqStrength60", unsafe.Offsetof(Modifier{}.ReqStrength60), 60},
		{"Modifier.Next80", unsafe.Offsetof(Modifier{}.Next80), 80},
	}
	checks64 := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.size", unsafe.Sizeof(Object{}), 928},
		{"Object.TypeInd", unsafe.Offsetof(Object{}.TypeInd), 8},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), 12},
		{"Modifier.size", unsafe.Sizeof(Modifier{}), 112},
		{"Modifier.TypeInd", unsafe.Offsetof(Modifier{}.TypeInd), 8},
		{"Modifier.ReqStrength60", unsafe.Offsetof(Modifier{}.ReqStrength60), 72},
		{"Modifier.Next80", unsafe.Offsetof(Modifier{}.Next80), 96},
	}
	checks := checks64
	if unsafe.Sizeof(uintptr(0)) == 4 {
		checks = checks32
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestPlayerCheckStrengthNative4F3180PreservesPointersAndLiveState(t *testing.T) {
	player := &Object{ObjClass: object.ClassPlayer}
	item := &Object{ObjClass: object.ClassWeapon, TypeInd: 7}
	armor := &Modifier{ReqStrength60: 42}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for index, pointer := range []unsafe.Pointer{unsafe.Pointer(player), unsafe.Pointer(item), unsafe.Pointer(armor)} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("pointer %d = %p, want native high address", index, pointer)
			}
		}
	}

	got := playerCheckStrengthNative4F3180(player, item, playerCheckStrengthNativeDeps4F3180{
		getUnitStrength: func(gotPlayer *Object) int32 {
			if gotPlayer != player {
				t.Fatalf("player = %p, want %p", gotPlayer, player)
			}
			item.ObjClass = object.ClassArmor
			item.TypeInd = 0xbeef
			return 42
		},
		findArmorDef: func(typeInd uint16) *Modifier {
			if typeInd != 0xbeef {
				t.Fatalf("armor type = %#x, want 0xbeef", typeInd)
			}
			return armor
		},
		findWeaponDef: func(uint16) *Modifier {
			t.Fatal("weapon lookup reached after live armor mutation")
			return nil
		},
	})
	if got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	runtime.KeepAlive(player)
	runtime.KeepAlive(item)
	runtime.KeepAlive(armor)
}

func TestServerPlayerCheckStrength4F3180UsesNativeDefinitions(t *testing.T) {
	s := new(Server)
	weapon := &Modifier{TypeInd: 0x1234, ReqStrength60: 37}
	armor := &Modifier{TypeInd: 0x5678, ReqStrength60: 38}
	s.Modif.Dword_5d4594_251600 = weapon
	s.Modif.Dword_5d4594_251608 = armor

	playerData := &Player{}
	playerData.Info().SetField2239(37)
	player := &Object{
		ObjClass:   object.ClassPlayer,
		UpdateData: unsafe.Pointer(&PlayerUpdateData{Player: playerData}),
	}
	weaponItem := &Object{ObjClass: object.ClassWeapon, TypeInd: 0x1234}
	armorItem := &Object{ObjClass: object.ClassArmor, TypeInd: 0x5678}

	if got := s.PlayerCheckStrength4F3180(player, weaponItem); got != 1 {
		t.Fatalf("weapon result = %d, want 1", got)
	}
	if got := s.PlayerCheckStrength4F3180(player, armorItem); got != 0 {
		t.Fatalf("armor result = %d, want 0", got)
	}
	playerData.Info().SetField2239(38)
	if got := s.PlayerCheckStrength4F3180(player, armorItem); got != 1 {
		t.Fatalf("armor equality result = %d, want 1", got)
	}
	missing := &Object{ObjClass: object.ClassWeapon, TypeInd: 0xffff}
	if got := s.PlayerCheckStrength4F3180(player, missing); got != 0 {
		t.Fatalf("missing definition result = %d, want 0", got)
	}
}
