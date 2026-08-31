package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestPlayerActionStateNative4FA2B0Layouts(t *testing.T) {
	wantUpdate := uintptr(748)
	wantPlayer := uintptr(276)
	wantUseData := uintptr(736)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantUpdate = 872
		wantPlayer = 336
		wantUseData = 848
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdate},
		{"PlayerUpdateData.State", unsafe.Offsetof(PlayerUpdateData{}.State), 88},
		{"PlayerUpdateData.EquippedWeapon", unsafe.Offsetof(PlayerUpdateData{}.EquippedWeapon), 104},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantPlayer},
		{"Player.WeaponEquip", unsafe.Offsetof(Player{}.WeaponEquip), 4},
		{"Player.Field8", unsafe.Offsetof(Player{}.Field8), 8},
		{"Object.UseData", unsafe.Offsetof(Object{}.UseData), wantUseData},
		{"UseDataPtr.Ptr", unsafe.Offsetof(UseDataPtr{}.Ptr), 0},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s offset = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestPlayerActionStateNative4FA2B0PreservesPointerChain(t *testing.T) {
	var useData [playerWeaponUseFlagsOffset4FA2B0 + 1]byte
	useData[playerWeaponUseFlagsOffset4FA2B0] = 2
	weapon := &Object{UseData: UseDataPtr{Ptr: unsafe.Pointer(&useData[0])}}
	player := &Player{WeaponEquip: 0x10000}
	update := &PlayerUpdateData{State: PlayerState1, EquippedWeapon: weapon, Player: player}
	unit := &Object{UpdateData: unsafe.Pointer(update)}
	s := new(Server)
	s.Abils.Reset()
	animationCalls := 0

	got := playerActionStateNative4FA2B0(s, unit, func(uint32) int32 {
		animationCalls++
		return -1
	})
	if got != 29 {
		t.Fatalf("result = %d, want 29", got)
	}
	if animationCalls != 0 {
		t.Fatalf("weapon animation calls = %d, want 0", animationCalls)
	}
}

func TestPlayerActionStateNative4FA2B0UsesNativeAbilities(t *testing.T) {
	player := &Player{WeaponEquip: 4}
	update := &PlayerUpdateData{State: PlayerState22, Player: player}
	unit := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	s := new(Server)
	s.Abils.Reset()
	s.Abils.SetExecHead(&ExecAbilityClass{Unit: unit, Abil: AbilityWarcry, Active: 1})

	got := playerActionStateNative4FA2B0(s, unit, func(uint32) int32 {
		t.Fatal("active warcry reached weapon animation")
		return -1
	})
	if got != 46 {
		t.Fatalf("result = %d, want 46", got)
	}
}

func TestPlayerActionStateNative4FA2B0InjectsWeaponTableLookup(t *testing.T) {
	player := &Player{WeaponEquip: 1 << 7, Field8: 0xff}
	update := &PlayerUpdateData{State: PlayerState14, Player: player}
	unit := &Object{UpdateData: unsafe.Pointer(update)}
	s := new(Server)
	s.Abils.Reset()
	var gotEquip uint32

	got := playerActionStateNative4FA2B0(s, unit, func(equip uint32) int32 {
		gotEquip = equip
		return 0x12345678
	})
	if got != 0x12345678 {
		t.Fatalf("result = %#x, want %#x", got, int32(0x12345678))
	}
	if gotEquip != player.WeaponEquip {
		t.Fatalf("weapon flags = %#x, want %#x", gotEquip, player.WeaponEquip)
	}
}

func TestPlayerActionStateNative4FA2B0DoesNotGuardUpdate(t *testing.T) {
	s := new(Server)
	s.Abils.Reset()
	defer func() {
		if recover() == nil {
			t.Fatal("nil update did not preserve the original fault contract")
		}
	}()
	_ = playerActionStateNative4FA2B0(s, &Object{}, func(uint32) int32 { return 0 })
}
