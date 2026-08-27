package server

import (
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func TestMonsterDieNative50A3D0OrdinaryCoopMonster(t *testing.T) {
	s := new(Server)
	s.handle = atomic.AddUintptr(&serverLast, 1)
	servers.Store(s.handle, s)
	t.Cleanup(func() { servers.Delete(s.handle) })
	s.SetTickRate(30)

	update := &MonsterUpdateData{
		AIStackInd:  0,
		StatusFlags: object.MonStatusAlert | object.MonStatusInjured | object.MonStatusRunning,
	}
	update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_FIGHT)}
	unit := &Object{
		ObjClass:     object.ClassMonster,
		ObjSubClass:  0x302,
		ObjFlags:     object.FlagEnabled | object.FlagDead | object.FlagMissileHit,
		Buffs:        0x1234,
		UpdateData:   unsafe.Pointer(update),
		serverHandle: s.handle,
	}
	unit.BuffsDur[3], unit.BuffsPower[3] = 77, 8

	if !s.MonsterDieNative50A3D0(unit, MonsterDieRuntime50A3D0{
		GameFlag: func(flag uint32) bool { return flag == monsterDieCoopFlag50A3D0 },
		IsZombie: func(*Object) bool { return false },
		Unsupported: func(reason string, _ *Object) {
			t.Fatalf("ordinary co-op death rejected: %s", reason)
		},
	}) {
		t.Fatal("ordinary co-op death was not handled")
	}
	if unit.ObjFlags.Has(object.FlagMissileHit) || !unit.ObjFlags.Has(object.FlagDead) {
		t.Fatalf("death flags = %#x", unit.ObjFlags)
	}
	if unit.Buffs != 0 || unit.BuffsDur[3] != 0 || unit.BuffsPower[3] != 0 {
		t.Fatalf("buff state = %#x/%d/%d", unit.Buffs, unit.BuffsDur[3], unit.BuffsPower[3])
	}
	if uint32(unit.ObjSubClass) != 0x202 {
		t.Fatalf("monster subclass = %#x, want 0x202", unit.ObjSubClass)
	}
	if update.AIStackInd != 1 || update.AIStack[0].Type() != ai.ACTION_DEAD ||
		update.AIStack[1].Type() != ai.ACTION_DYING {
		t.Fatalf("death action stack = %#v", update.GetAIStack())
	}
}

func TestMonsterDieNative50A3D0PreflightsObservedMonster(t *testing.T) {
	s := new(Server)
	s.handle = atomic.AddUintptr(&serverLast, 1)
	servers.Store(s.handle, s)
	t.Cleanup(func() { servers.Delete(s.handle) })

	update := &MonsterUpdateData{AIStackInd: 0}
	unit := &Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(update), serverHandle: s.handle}
	update.AIStack[0].Action = uint32(ai.ACTION_IDLE)
	player := Player{Active: 1, PlayerInd: 7, CameraFollowObj: unit, Field3680: 2}
	playerUpdate := &PlayerUpdateData{Player: &player}
	playerUnit := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(playerUpdate), serverHandle: s.handle}
	player.PlayerUnit = playerUnit
	s.Players.list = []Player{player}
	// The slice copy above gives the live server entry a distinct address.
	s.Players.list[0].PlayerUnit.UpdateDataPlayer().Player = &s.Players.list[0]

	before := *unit
	beforeUpdate := *update
	var reason string
	if s.MonsterDieNative50A3D0(unit, MonsterDieRuntime50A3D0{
		GameFlag: func(uint32) bool { return true },
		IsZombie: func(*Object) bool { return false },
		Unsupported: func(got string, _ *Object) {
			reason = got
		},
	}) {
		t.Fatal("missing ObserveClear callback was accepted")
	}
	if reason != "observed monster" {
		t.Fatalf("unsupported reason = %q", reason)
	}
	if *unit != before || *update != beforeUpdate {
		t.Fatal("failed preflight mutated monster")
	}
}
