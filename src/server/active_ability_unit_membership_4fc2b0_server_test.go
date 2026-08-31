package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	playerlib "github.com/opennox/libs/player"
)

func TestActiveAbilityUnitMembershipNative4FC2B0Layout(t *testing.T) {
	type layout struct {
		objectClass      uintptr
		objectUpdateData uintptr
		updatePlayer     uintptr
		playerClass      uintptr
		execAbility      uintptr
		execUnit         uintptr
		execFrame        uintptr
		execActive       uintptr
		execNext         uintptr
		execPrev         uintptr
		execSize         uintptr
	}
	wants := map[uintptr]layout{
		4: {
			objectClass: 8, objectUpdateData: 748, updatePlayer: 276,
			playerClass: 2251, execAbility: 0, execUnit: 4, execFrame: 8,
			execActive: 12, execNext: 16, execPrev: 20, execSize: 24,
		},
		8: {
			objectClass: 12, objectUpdateData: 872, updatePlayer: 336,
			playerClass: 2255, execAbility: 0, execUnit: 8, execFrame: 16,
			execActive: 20, execNext: 24, execPrev: 32, execSize: 40,
		},
	}
	ptrSize := unsafe.Sizeof(uintptr(0))
	want, ok := wants[ptrSize]
	if !ok {
		t.Fatalf("unsupported pointer size %d", ptrSize)
	}
	got := layout{
		objectClass:      unsafe.Offsetof(Object{}.ObjClass),
		objectUpdateData: unsafe.Offsetof(Object{}.UpdateData),
		updatePlayer:     unsafe.Offsetof(PlayerUpdateData{}.Player),
		playerClass:      unsafe.Offsetof(Player{}.info) + unsafe.Offsetof(PlayerInfo{}.playerClass),
		execAbility:      unsafe.Offsetof(ExecAbilityClass{}.Abil),
		execUnit:         unsafe.Offsetof(ExecAbilityClass{}.Unit),
		execFrame:        unsafe.Offsetof(ExecAbilityClass{}.Frame),
		execActive:       unsafe.Offsetof(ExecAbilityClass{}.Active),
		execNext:         unsafe.Offsetof(ExecAbilityClass{}.Next),
		execPrev:         unsafe.Offsetof(ExecAbilityClass{}.Prev),
		execSize:         unsafe.Sizeof(ExecAbilityClass{}),
	}
	if got != want {
		t.Fatalf("native layout = %+v, want %+v", got, want)
	}
}

func activeAbilityUnitMembershipWarriorUnit4FC2B0() (*Object, *PlayerUpdateData, *Player) {
	player := new(Player)
	player.Info().SetPlayerClass(playerlib.Warrior)
	update := &PlayerUpdateData{Player: player}
	unit := &Object{
		ObjClass:   object.ClassPlayer,
		UpdateData: unsafe.Pointer(update),
	}
	return unit, update, player
}

func TestActiveAbilityUnitMembershipNative4FC2B0PreservesNativePointersAndList(t *testing.T) {
	unit, update, player := activeAbilityUnitMembershipWarriorUnit4FC2B0()
	other := new(Object)
	match := &ExecAbilityClass{
		Abil: Ability(math.MinInt32), Unit: unit, Frame: math.MaxUint32, Active: 0,
	}
	head := &ExecAbilityClass{
		Abil: AbilityHarpoon, Unit: other, Frame: 77, Active: 1, Next: match,
	}
	match.Prev = head
	a := serverAbilities{execList: head}

	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}
	if !a.IsAnyActive(unit) {
		t.Fatal("inactive record with signed minimum ability was not found")
	}
	if a.IsAnyActive(new(Object)) {
		t.Fatal("unlisted unit was reported present")
	}
	if a.execList != head || head.Next != match || match.Prev != head ||
		match.Abil != Ability(math.MinInt32) || match.Active != 0 || match.Frame != math.MaxUint32 {
		t.Fatal("unit membership query mutated ignored fields or list topology")
	}
	runtime.KeepAlive(update)
	runtime.KeepAlive(player)
}

func TestActiveAbilityUnitMembershipNative4FC2B0ClassAndNilGates(t *testing.T) {
	t.Run("nil unit faults", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("nil unit did not fault at the class read")
			}
		}()
		a := serverAbilities{}
		a.IsAnyActive(nil)
	})

	t.Run("non-Player returns before head", func(t *testing.T) {
		headLoads := 0
		got := activeAbilityUnitMembershipNative4FC2B0(
			new(Object),
			activeAbilityUnitMembershipNativeDeps4FC2B0{
				loadExecHead: func() *ExecAbilityClass {
					headLoads++
					return new(ExecAbilityClass)
				},
			},
		)
		if got != 0 || headLoads != 0 {
			t.Fatalf("non-Player result/head loads = %d/%d, want 0/0", got, headLoads)
		}
	})

	t.Run("nil UpdateData skips Player class", func(t *testing.T) {
		unit := &Object{ObjClass: object.ClassPlayer}
		a := serverAbilities{execList: &ExecAbilityClass{
			Unit: unit, Abil: AbilityHarpoon, Active: 0,
		}}
		if !a.IsAnyActive(unit) {
			t.Fatal("nil UpdateData did not preserve the executable's membership path")
		}
	})

	t.Run("nil Player faults", func(t *testing.T) {
		update := new(PlayerUpdateData)
		unit := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
		defer func() {
			if recover() == nil {
				t.Fatal("nil Player did not fault at the class-byte read")
			}
		}()
		a := serverAbilities{}
		a.IsAnyActive(unit)
	})

	t.Run("non-Warrior returns before head", func(t *testing.T) {
		player := new(Player)
		player.Info().SetPlayerClass(playerlib.Wizard)
		update := &PlayerUpdateData{Player: player}
		unit := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
		headLoads := 0
		got := activeAbilityUnitMembershipNative4FC2B0(
			unit,
			activeAbilityUnitMembershipNativeDeps4FC2B0{
				loadExecHead: func() *ExecAbilityClass {
					headLoads++
					return &ExecAbilityClass{Unit: unit}
				},
			},
		)
		if got != 0 || headLoads != 0 {
			t.Fatalf("non-Warrior result/head loads = %d/%d, want 0/0", got, headLoads)
		}
		runtime.KeepAlive(update)
		runtime.KeepAlive(player)
	})

	t.Run("nil head misses", func(t *testing.T) {
		unit, update, player := activeAbilityUnitMembershipWarriorUnit4FC2B0()
		a := serverAbilities{}
		if a.IsAnyActive(unit) {
			t.Fatal("nil execution-list head reported a match")
		}
		runtime.KeepAlive(update)
		runtime.KeepAlive(player)
	})
}
