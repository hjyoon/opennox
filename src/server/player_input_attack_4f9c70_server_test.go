package server

import (
	"math"
	"testing"
	"unsafe"
)

func TestPlayerInputAttackNative4F9C70Layouts(t *testing.T) {
	wantUpdate := uintptr(748)
	wantField34 := uintptr(136)
	wantAnim := uintptr(236)
	wantPlayer := uintptr(276)
	wantCursor := uintptr(288)
	wantUseData := uintptr(736)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantUpdate = 872
		wantField34 = 140
		wantAnim = 296
		wantPlayer = 336
		wantCursor = 360
		wantUseData = 848
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.Field34", unsafe.Offsetof(Object{}.Field34), wantField34},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdate},
		{"Object.UseData", unsafe.Offsetof(Object{}.UseData), wantUseData},
		{"PlayerUpdateData.State", unsafe.Offsetof(PlayerUpdateData{}.State), 88},
		{"PlayerUpdateData.EquippedWeapon", unsafe.Offsetof(PlayerUpdateData{}.EquippedWeapon), 104},
		{"PlayerUpdateData.Field59_0", unsafe.Offsetof(PlayerUpdateData{}.Field59_0), wantAnim},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantPlayer},
		{"PlayerUpdateData.CursorObj", unsafe.Offsetof(PlayerUpdateData{}.CursorObj), wantCursor},
		{"Player.WeaponEquip", unsafe.Offsetof(Player{}.WeaponEquip), 4},
		{"WandUseData.Flags", unsafe.Offsetof(WandUseData{}.Flags), 96},
		{"WandUseData.Charge", unsafe.Offsetof(WandUseData{}.Charge), 108},
		{"WandUseData.MaxCharge", unsafe.Offsetof(WandUseData{}.MaxCharge), 109},
		{"WandUseData size", unsafe.Sizeof(WandUseData{}), 116},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestPlayerInputAttackNative4F9C70ReloadsNativeWeaponAndUse(t *testing.T) {
	var oldUseToken, newUseToken byte
	oldUsePointer := unsafe.Pointer(&oldUseToken)
	newUsePointer := unsafe.Pointer(&newUseToken)
	oldUseCalls := 0
	newUseCalls := 0

	useData := &WandUseData{Charge: 1}
	oldWeapon := &Object{
		Use:     UseFuncPtr{Ptr: oldUsePointer},
		UseData: UseDataPtr{Ptr: unsafe.Pointer(useData)},
	}
	newWeapon := &Object{Use: UseFuncPtr{Ptr: newUsePointer}}
	player := &Player{WeaponEquip: 0x10000}
	update := &PlayerUpdateData{
		State:          PlayerState1,
		EquippedWeapon: oldWeapon,
		Field59_0:      0xff,
		Player:         player,
	}
	unit := &Object{UpdateData: unsafe.Pointer(update)}

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]unsafe.Pointer{
			"unit": unsafe.Pointer(unit), "update": unsafe.Pointer(update),
			"player": unsafe.Pointer(player), "old weapon": unsafe.Pointer(oldWeapon),
			"new weapon": unsafe.Pointer(newWeapon), "use data": unsafe.Pointer(useData),
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("%s pointer = %p, want native address above 4 GiB", name, pointer)
			}
		}
	}

	objUse.Register(oldUsePointer, func(*Object, *Object) bool {
		oldUseCalls++
		return false
	})
	objUse.Register(newUsePointer, func(gotOwner, gotWeapon *Object) bool {
		newUseCalls++
		if gotOwner != unit || gotWeapon != newWeapon {
			t.Fatalf("Use args = %p/%p, want %p/%p", gotOwner, gotWeapon, unit, newWeapon)
		}
		return false
	})

	s := &Server{}
	s.SetFrame(0xfedcba98)
	var buffCalls []int32
	setCalls := 0
	s.PlayerInputAttack4F9C70(unit, PlayerInputAttackRuntime4F9C70{
		SetState: func(gotUnit *Object, state PlayerState) bool {
			setCalls++
			if gotUnit != unit || state != PlayerState1 {
				t.Fatalf("SetState args = %p/%d, want %p/1", gotUnit, state, unit)
			}
			update.EquippedWeapon = newWeapon
			return true
		},
		BuffOff: func(gotUnit *Object, buff int32) int32 {
			if gotUnit != unit {
				t.Fatalf("BuffOff unit = %p, want %p", gotUnit, unit)
			}
			buffCalls = append(buffCalls, buff)
			return int32(0x7000) + buff
		},
	})

	if oldUseCalls != 0 || newUseCalls != 1 || setCalls != 1 {
		t.Fatalf("old/new/set calls = %d/%d/%d, want 0/1/1", oldUseCalls, newUseCalls, setCalls)
	}
	if len(buffCalls) != 2 || buffCalls[0] != 0 || buffCalls[1] != 23 {
		t.Fatalf("buff calls = %v, want [0 23]", buffCalls)
	}
	if unit.Field34 != 0xfedcba98 || update.Field59_0 != 0 {
		t.Fatalf("frame/anim = %#x/%d, want %#x/0", unit.Field34, update.Field59_0, uint32(0xfedcba98))
	}
}

func TestPlayerAimsAtEnemyNative4F9DC0NilContracts(t *testing.T) {
	s := &Server{}
	if got := s.PlayerAimsAtEnemy4F9DC0(nil); got != 0 {
		t.Fatalf("nil unit result = %d, want 0", got)
	}
	update := &PlayerUpdateData{}
	unit := &Object{UpdateData: unsafe.Pointer(update)}
	if got := s.PlayerAimsAtEnemy4F9DC0(unit); got != 1 {
		t.Fatalf("nil cursor result = %d, want 1", got)
	}
}
