package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	playerlib "github.com/opennox/libs/player"
)

func TestActiveAbilityDeactivateNative4FC440Layout(t *testing.T) {
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
	if got := unsafe.Sizeof(Ability(0)); got != 4 {
		t.Fatalf("Ability width = %d, want 4", got)
	}
}

func activeAbilityDeactivateWarriorUnit4FC440() (*Object, *PlayerUpdateData, *Player) {
	player := new(Player)
	player.Info().SetPlayerClass(playerlib.Warrior)
	update := &PlayerUpdateData{Player: player}
	unit := &Object{
		ObjClass:   object.ClassPlayer,
		UpdateData: unsafe.Pointer(update),
	}
	return unit, update, player
}

func TestActiveAbilityDeactivateNative4FC440PreservesPointersSignedAbilityAndList(t *testing.T) {
	unit, update, player := activeAbilityDeactivateWarriorUnit4FC440()
	other := new(Object)
	duplicate := &ExecAbilityClass{
		Abil: Ability(math.MinInt32), Unit: unit, Frame: 19, Active: 0x55667788,
	}
	match := &ExecAbilityClass{
		Abil: Ability(math.MinInt32), Unit: unit, Frame: math.MaxUint32,
		Active: 0x89abcdef, Next: duplicate,
	}
	head := &ExecAbilityClass{
		Abil: Ability(math.MinInt32), Unit: other, Frame: 77,
		Active: math.MaxUint32, Next: match,
	}
	match.Prev = head
	duplicate.Prev = match
	a := serverAbilities{execList: head}

	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}
	a.Sub4FC440(unit, Ability(math.MinInt32))
	if match.Active != 0 {
		t.Fatalf("matching Active = %08x, want 00000000", match.Active)
	}
	if duplicate.Active != 0x55667788 {
		t.Fatalf("later duplicate Active = %08x, want 55667788", duplicate.Active)
	}
	if head.Active != math.MaxUint32 || a.execList != head || head.Next != match ||
		match.Next != duplicate || match.Prev != head || duplicate.Prev != match ||
		match.Frame != math.MaxUint32 || duplicate.Frame != 19 {
		t.Fatal("deactivation mutated another record, list topology, or deadline")
	}

	a.Sub4FC440(unit, Ability(math.MaxInt32))
	if duplicate.Active != 0x55667788 || head.Active != math.MaxUint32 {
		t.Fatal("signed-ability miss mutated Active")
	}
	runtime.KeepAlive(update)
	runtime.KeepAlive(player)
}

func TestActiveAbilityDeactivateNative4FC440ClassAndNilGates(t *testing.T) {
	t.Run("nil unit faults", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("nil unit did not fault at the class read")
			}
		}()
		a := serverAbilities{}
		a.Sub4FC440(nil, AbilityBerserk)
	})

	t.Run("non-Player returns before head", func(t *testing.T) {
		headLoads := 0
		activeAbilityDeactivateNative4FC440(
			new(Object),
			AbilityBerserk,
			activeAbilityDeactivateNativeDeps4FC440{
				loadExecHead: func() *ExecAbilityClass {
					headLoads++
					return new(ExecAbilityClass)
				},
			},
		)
		if headLoads != 0 {
			t.Fatalf("non-Player head loads = %d, want 0", headLoads)
		}
	})

	t.Run("nil UpdateData skips Player class", func(t *testing.T) {
		unit := &Object{ObjClass: object.ClassPlayer}
		record := &ExecAbilityClass{
			Unit: unit, Abil: AbilityHarpoon, Active: math.MaxUint32,
		}
		a := serverAbilities{execList: record}
		a.Sub4FC440(unit, AbilityHarpoon)
		if record.Active != 0 {
			t.Fatalf("nil UpdateData Active = %08x, want 00000000", record.Active)
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
		a.Sub4FC440(unit, AbilityBerserk)
	})

	t.Run("non-Warrior returns before head", func(t *testing.T) {
		player := new(Player)
		player.Info().SetPlayerClass(playerlib.Wizard)
		update := &PlayerUpdateData{Player: player}
		unit := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
		headLoads := 0
		activeAbilityDeactivateNative4FC440(
			unit,
			AbilityBerserk,
			activeAbilityDeactivateNativeDeps4FC440{
				loadExecHead: func() *ExecAbilityClass {
					headLoads++
					return &ExecAbilityClass{Unit: unit, Abil: AbilityBerserk, Active: 1}
				},
			},
		)
		if headLoads != 0 {
			t.Fatalf("non-Warrior head loads = %d, want 0", headLoads)
		}
		runtime.KeepAlive(update)
		runtime.KeepAlive(player)
	})

	t.Run("nil head leaves state unchanged", func(t *testing.T) {
		unit, update, player := activeAbilityDeactivateWarriorUnit4FC440()
		a := serverAbilities{}
		a.Sub4FC440(unit, AbilityBerserk)
		if a.execList != nil {
			t.Fatal("nil execution-list head changed")
		}
		runtime.KeepAlive(update)
		runtime.KeepAlive(player)
	})
}
