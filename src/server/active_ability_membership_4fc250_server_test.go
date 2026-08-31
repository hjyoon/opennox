package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	playerlib "github.com/opennox/libs/player"
)

func TestActiveAbilityMembershipNative4FC250Layout(t *testing.T) {
	type layout struct {
		objectClass      uintptr
		objectUpdateData uintptr
		updatePlayer     uintptr
		playerClass      uintptr
		execAbility      uintptr
		execUnit         uintptr
		execNext         uintptr
		execSize         uintptr
	}
	wants := map[uintptr]layout{
		4: {
			objectClass: 8, objectUpdateData: 748, updatePlayer: 276,
			playerClass: 2251, execAbility: 0, execUnit: 4, execNext: 16, execSize: 24,
		},
		8: {
			objectClass: 12, objectUpdateData: 872, updatePlayer: 336,
			playerClass: 2255, execAbility: 0, execUnit: 8, execNext: 24, execSize: 40,
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
		execNext:         unsafe.Offsetof(ExecAbilityClass{}.Next),
		execSize:         unsafe.Sizeof(ExecAbilityClass{}),
	}
	if got != want {
		t.Fatalf("native layout = %+v, want %+v", got, want)
	}
	if got := unsafe.Sizeof(Ability(0)); got != 4 {
		t.Fatalf("Ability width = %d, want 4", got)
	}
}

func activeAbilityMembershipWarriorUnit4FC250() (*Object, *PlayerUpdateData, *Player) {
	player := new(Player)
	player.Info().SetPlayerClass(playerlib.Warrior)
	update := &PlayerUpdateData{Player: player}
	unit := &Object{
		ObjClass:   object.ClassPlayer,
		UpdateData: unsafe.Pointer(update),
	}
	return unit, update, player
}

func TestActiveAbilityMembershipNative4FC250PreservesPointersSignedAbilityAndList(t *testing.T) {
	unit, update, player := activeAbilityMembershipWarriorUnit4FC250()
	other := new(Object)
	tail := &ExecAbilityClass{Unit: unit, Abil: AbilityBerserk, Frame: 31, Active: 1}
	match := &ExecAbilityClass{
		Abil: Ability(math.MinInt32), Unit: unit, Frame: math.MaxUint32,
		Active: 0, Next: tail,
	}
	head := &ExecAbilityClass{
		Abil: Ability(math.MinInt32), Unit: other, Frame: 77,
		Active: 1, Next: match,
	}
	match.Prev = head
	tail.Prev = match
	a := serverAbilities{execList: head}

	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}
	if !a.IsActive(unit, Ability(math.MinInt32)) {
		t.Fatal("inactive record with signed minimum ability was not found")
	}
	if a.IsActive(unit, Ability(math.MaxInt32)) {
		t.Fatal("nonmatching signed ability was reported active")
	}
	if a.execList != head || head.Next != match || match.Next != tail ||
		match.Prev != head || tail.Prev != match || match.Active != 0 ||
		match.Frame != math.MaxUint32 {
		t.Fatal("membership query mutated list topology, Active, or deadline")
	}
	runtime.KeepAlive(update)
	runtime.KeepAlive(player)
}

func TestActiveAbilityMembershipNative4FC250ClassAndNilGates(t *testing.T) {
	t.Run("nil unit faults", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("nil unit did not fault at the class read")
			}
		}()
		a := serverAbilities{}
		a.IsActive(nil, AbilityBerserk)
	})

	t.Run("non-Player returns before head", func(t *testing.T) {
		headLoads := 0
		got := activeAbilityMembershipNative4FC250(
			new(Object),
			AbilityBerserk,
			activeAbilityMembershipNativeDeps4FC250{
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
		if !a.IsActive(unit, AbilityHarpoon) {
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
		a.IsActive(unit, AbilityBerserk)
	})

	t.Run("non-Warrior returns before head", func(t *testing.T) {
		player := new(Player)
		player.Info().SetPlayerClass(playerlib.Wizard)
		update := &PlayerUpdateData{Player: player}
		unit := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
		headLoads := 0
		got := activeAbilityMembershipNative4FC250(
			unit,
			AbilityBerserk,
			activeAbilityMembershipNativeDeps4FC250{
				loadExecHead: func() *ExecAbilityClass {
					headLoads++
					return &ExecAbilityClass{Unit: unit, Abil: AbilityBerserk}
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
		unit, update, player := activeAbilityMembershipWarriorUnit4FC250()
		a := serverAbilities{}
		if a.IsActive(unit, AbilityBerserk) {
			t.Fatal("nil execution-list head reported a match")
		}
		runtime.KeepAlive(update)
		runtime.KeepAlive(player)
	})
}
