package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	playerlib "github.com/opennox/libs/player"
	"github.com/opennox/libs/types"
)

func TestWarcryProximityScanNative4FC4C0Layout(t *testing.T) {
	type layout struct {
		playerUnit  uintptr
		playerClass uintptr
		objectX     uintptr
		objectY     uintptr
		pointSize   uintptr
		abilitySize uintptr
	}
	wants := map[uintptr]layout{
		4: {
			playerUnit: 2056, playerClass: 2251,
			objectX: 56, objectY: 60, pointSize: 8, abilitySize: 4,
		},
		8: {
			playerUnit: 2056, playerClass: 2255,
			objectX: 60, objectY: 64, pointSize: 8, abilitySize: 4,
		},
	}
	ptrSize := unsafe.Sizeof(uintptr(0))
	want, ok := wants[ptrSize]
	if !ok {
		t.Fatalf("unsupported pointer size %d", ptrSize)
	}
	got := layout{
		playerUnit:  unsafe.Offsetof(Player{}.PlayerUnit),
		playerClass: unsafe.Offsetof(Player{}.info) + unsafe.Offsetof(PlayerInfo{}.playerClass),
		objectX:     unsafe.Offsetof(Object{}.PosVec) + unsafe.Offsetof(types.Pointf{}.X),
		objectY:     unsafe.Offsetof(Object{}.PosVec) + unsafe.Offsetof(types.Pointf{}.Y),
		pointSize:   unsafe.Sizeof(types.Pointf{}),
		abilitySize: unsafe.Sizeof(Ability(0)),
	}
	if got != want {
		t.Fatalf("native layout on %s/%s = %+v, want %+v", runtime.GOOS, runtime.GOARCH, got, want)
	}
}

func TestWarcryProximityScanNative4FC4C0PreservesPointersAndReloadsUnit(t *testing.T) {
	target := &Object{PosVec: types.Ptf(10, 20)}
	original := &Object{PosVec: types.Ptf(1000, 1000)}
	replacement := &Object{PosVec: types.Ptf(13, 24)}
	player := &Player{PlayerUnit: original}
	player.Info().SetPlayerClass(playerlib.Warrior)

	var (
		activeCalls int
		mapCalls    int
	)
	got := warcryProximityScanNative4FC4C0(target, warcryProximityScanNativeDeps4FC4C0{
		firstPlayer: func() *Player {
			return player
		},
		nextPlayer: func(got *Player) *Player {
			t.Fatalf("next called after successful map check with %p", got)
			return nil
		},
		isAbilityActive: func(gotUnit *Object, ability Ability) int32 {
			activeCalls++
			if gotUnit != original || ability != AbilityWarcry {
				t.Fatalf("ability check = (%p,%d), want (%p,%d)", gotUnit, ability, original, AbilityWarcry)
			}
			player.PlayerUnit = replacement
			return math.MinInt32
		},
		mapCheck: func(gotUnit, gotTarget *Object) int32 {
			mapCalls++
			if gotUnit != replacement || gotTarget != target {
				t.Fatalf("map check = (%p,%p), want (%p,%p)", gotUnit, gotTarget, replacement, target)
			}
			return math.MinInt32
		},
	})
	if got != 1 || activeCalls != 1 || mapCalls != 1 {
		t.Fatalf("scan/ability/map = %d/%d/%d, want 1/1/1", got, activeCalls, mapCalls)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]uintptr{
			"target": uintptr(unsafe.Pointer(target)), "original": uintptr(unsafe.Pointer(original)),
			"replacement": uintptr(unsafe.Pointer(replacement)), "player": uintptr(unsafe.Pointer(player)),
		} {
			if pointer <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, pointer)
			}
		}
	}
	runtime.KeepAlive(target)
	runtime.KeepAlive(original)
	runtime.KeepAlive(replacement)
	runtime.KeepAlive(player)
}

func TestWarcryProximityScanNative4FC4C0ClassGateAndNilFault(t *testing.T) {
	t.Run("non-Warrior skips ability", func(t *testing.T) {
		unit := new(Object)
		player := &Player{PlayerUnit: unit}
		player.Info().SetPlayerClass(playerlib.Wizard)
		abilityCalls := 0
		got := warcryProximityScanNative4FC4C0(new(Object), warcryProximityScanNativeDeps4FC4C0{
			firstPlayer: func() *Player { return player },
			nextPlayer:  func(*Player) *Player { return nil },
			isAbilityActive: func(*Object, Ability) int32 {
				abilityCalls++
				return 1
			},
			mapCheck: func(*Object, *Object) int32 {
				t.Fatal("map check reached for non-Warrior")
				return 0
			},
		})
		if got != 0 || abilityCalls != 0 {
			t.Fatalf("scan/ability calls = %d/%d, want 0/0", got, abilityCalls)
		}
	})

	t.Run("nil target faults at first position read", func(t *testing.T) {
		unit := new(Object)
		player := &Player{PlayerUnit: unit}
		player.Info().SetPlayerClass(playerlib.Warrior)
		defer func() {
			if recover() == nil {
				t.Fatal("nil target did not preserve the original position-read fault")
			}
		}()
		warcryProximityScanNative4FC4C0(nil, warcryProximityScanNativeDeps4FC4C0{
			firstPlayer:     func() *Player { return player },
			nextPlayer:      func(*Player) *Player { return nil },
			isAbilityActive: func(*Object, Ability) int32 { return 1 },
			mapCheck:        func(*Object, *Object) int32 { return 1 },
		})
	})
}

func TestWarcryProximityScan4FC4C0BindsPlayersAndAbilityList(t *testing.T) {
	unit := &Object{ObjClass: 4, PosVec: types.Ptf(300, 0)}
	target := &Object{PosVec: types.Ptf(0, 0)}
	s := new(Server)
	s.Players.list = []Player{{Active: 1, PlayerInd: 0, PlayerUnit: unit}}
	s.Players.list[0].Info().SetPlayerClass(playerlib.Warrior)
	s.Abils.s = s
	s.Abils.execList = &ExecAbilityClass{Unit: unit, Abil: AbilityWarcry}

	// The exact 300-unit distance rejects before MapTraceVision, allowing this
	// test to exercise the real Player iterator and ability list without map
	// fixture state obscuring either binding.
	if got := s.Abils.WarcryProximityScan4FC4C0(target); got != 0 {
		t.Fatalf("scan = %d, want distance rejection", got)
	}
}
