package server

import (
	"fmt"
	"reflect"
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func completeMonsterDieRuntime50A3D0() MonsterDieRuntime50A3D0 {
	return MonsterDieRuntime50A3D0{
		GameFlag:        func(uint32) bool { return false },
		IsZombie:        func(*Object) bool { return false },
		ObserveClear:    func(*Object) {},
		QuestPrepare:    func(*Object) {},
		RemoveShadow:    func(*Object) {},
		RandomInt:       func(minimum, _ int) int { return minimum },
		SetDecayTime:    func(*Object, uint32) {},
		NetFxShield:     func(int, *Object) {},
		UnmarkMinimap:   func(int, *Object, uint32) {},
		DropAllItems:    func(*Object) {},
		AwardSoloKill:   func(*Object) {},
		CreditQuestKill: func(*Object) {},
	}
}

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

	runtime := completeMonsterDieRuntime50A3D0()
	runtime.GameFlag = func(flag uint32) bool { return flag == monsterDieCoopFlag50A3D0 }
	var removedShadow, dropped int
	runtime.RemoveShadow = func(got *Object) {
		removedShadow++
		if got != unit {
			t.Fatalf("RemoveShadow object = %p, want %p", got, unit)
		}
	}
	runtime.DropAllItems = func(got *Object) {
		dropped++
		if got != unit {
			t.Fatalf("DropAllItems object = %p, want %p", got, unit)
		}
	}
	runtime.Unsupported = func(reason string, _ *Object) {
		t.Fatalf("ordinary co-op death rejected: %s", reason)
	}
	if !s.MonsterDieNative50A3D0(unit, runtime) {
		t.Fatal("ordinary co-op death was not handled")
	}
	if removedShadow != 1 || dropped != 1 {
		t.Fatalf("unconditional callbacks = shadow:%d drop:%d, want 1/1", removedShadow, dropped)
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
	if reason != "missing ObserveClear" {
		t.Fatalf("unsupported reason = %q", reason)
	}
	if *unit != before || *update != beforeUpdate {
		t.Fatal("failed preflight mutated monster")
	}
}

func TestMonsterDieNative50A3D0QuestPreservesLiveOrderAndRewards(t *testing.T) {
	s := new(Server)
	s.handle = atomic.AddUintptr(&serverLast, 1)
	servers.Store(s.handle, s)
	t.Cleanup(func() { servers.Delete(s.handle) })
	s.SetTickRate(30)

	update := &MonsterUpdateData{
		AIStackInd: 0,
		Field546:   2,
		Field547:   2,
	}
	update.AIStack[0] = AIStackItem{Action: uint32(ai.ACTION_FIGHT)}
	unit := &Object{
		ObjClass:     object.ClassMonster,
		ObjSubClass:  0x1182,
		ObjFlags:     object.FlagEnabled | object.FlagDead | object.FlagMissileHit,
		Buffs:        0x1234,
		UpdateData:   unsafe.Pointer(update),
		serverHandle: s.handle,
	}
	unit.BuffsDur[0], unit.BuffsPower[0] = 12, 3

	observer := Player{Active: 1, PlayerInd: 7, CameraFollowObj: unit, Field3680: 2}
	observerUnit := &Object{ObjClass: object.ClassPlayer, serverHandle: s.handle}
	observer.PlayerUnit = observerUnit
	observerUnit.UpdateData = unsafe.Pointer(&PlayerUpdateData{Player: &observer})
	s.Players.list = []Player{observer}
	s.Players.list[0].PlayerUnit.UpdateDataPlayer().Player = &s.Players.list[0]

	ownerPlayer := &Player{Active: 1, PlayerInd: 9}
	owner := &Object{
		ObjClass:     object.ClassPlayer,
		UpdateData:   unsafe.Pointer(&PlayerUpdateData{Player: ownerPlayer}),
		serverHandle: s.handle,
	}
	ownerPlayer.PlayerUnit = owner
	owner.Field129 = unit
	unit.ObjOwner = owner

	awardPlayer := &Player{Active: 1, PlayerInd: 10}
	awardKiller := &Object{
		ObjClass:     object.ClassPlayer,
		UpdateData:   unsafe.Pointer(&PlayerUpdateData{Player: awardPlayer}),
		serverHandle: s.handle,
	}
	awardPlayer.PlayerUnit = awardKiller
	questPlayer := &Player{Active: 1, PlayerInd: 11, field4664: 4, field4692: 1}
	questKiller := &Object{
		ObjClass:     object.ClassPlayer,
		UpdateData:   unsafe.Pointer(&PlayerUpdateData{Player: questPlayer}),
		serverHandle: s.handle,
	}
	questPlayer.PlayerUnit = questKiller
	unit.Obj130 = awardKiller

	var events []string
	record := func(event string) { events = append(events, event) }
	runtime := completeMonsterDieRuntime50A3D0()
	runtime.GameFlag = func(flag uint32) bool {
		record(fmt.Sprintf("flag:%#x", flag))
		return flag == monsterDieQuestFlag50A3D0
	}
	runtime.QuestPrepare = func(got *Object) {
		record("quest-prepare")
		if got != unit {
			t.Fatalf("QuestPrepare object = %p, want %p", got, unit)
		}
	}
	runtime.ObserveClear = func(got *Object) {
		record("observe-clear")
		if got != observerUnit {
			t.Fatalf("ObserveClear object = %p, want %p", got, observerUnit)
		}
	}
	runtime.IsZombie = func(got *Object) bool {
		record("is-zombie")
		if got != unit || update.AIStackInd != 1 || update.AIStack[0].Type() != ai.ACTION_DEAD ||
			update.AIStack[1].Type() != ai.ACTION_DYING {
			t.Fatalf("IsZombie state = %p/%#v", got, update.GetAIStack())
		}
		return false
	}
	runtime.RemoveShadow = func(got *Object) {
		record("remove-shadow")
		if got != unit || got.ObjFlags.Has(object.FlagMissileHit) {
			t.Fatalf("RemoveShadow state = %p/%#x", got, got.ObjFlags)
		}
	}
	runtime.RandomInt = func(minimum, maximum int) int {
		record(fmt.Sprintf("random:%d:%d", minimum, maximum))
		return 7
	}
	runtime.SetDecayTime = func(got *Object, frames uint32) {
		record(fmt.Sprintf("decay:%d", frames))
		if got != unit || frames != 210 {
			t.Fatalf("SetDecayTime = %p/%d, want %p/210", got, frames, unit)
		}
	}
	runtime.NetFxShield = func(index int, got *Object) {
		record(fmt.Sprintf("shield:%d", index))
		if index != 9 || got != unit {
			t.Fatalf("NetFxShield = %d/%p, want 9/%p", index, got, unit)
		}
	}
	runtime.UnmarkMinimap = func(index int, got *Object, flags uint32) {
		record(fmt.Sprintf("unmark:%d:%d", index, flags))
		if index != 9 || got != unit || flags != 1 {
			t.Fatalf("UnmarkMinimap = %d/%p/%d", index, got, flags)
		}
	}
	runtime.DropAllItems = func(got *Object) {
		record("drop")
		if got != unit || got.InvFirstItem != nil {
			t.Fatalf("DropAllItems = %p/inventory %p", got, got.InvFirstItem)
		}
	}
	runtime.AwardSoloKill = func(got *Object) {
		record("award")
		if got != awardKiller {
			t.Fatalf("AwardSoloKill = %p, want %p", got, awardKiller)
		}
		unit.Obj130 = questKiller
	}
	runtime.CreditQuestKill = func(got *Object) {
		record("credit")
		if got != questKiller {
			t.Fatalf("CreditQuestKill = %p, want live %p", got, questKiller)
		}
		got.RecordMonsterKilled4D6170()
	}

	if !s.MonsterDieNative50A3D0(unit, runtime) {
		t.Fatal("Quest death was not handled")
	}
	wantEvents := []string{
		"flag:0x1000", "quest-prepare", "observe-clear", "is-zombie", "remove-shadow",
		"flag:0x1000", "random:5:8", "decay:210", "shield:9", "unmark:9:1", "drop",
		"flag:0x800", "award", "flag:0x1000", "credit",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	if unit.ObjOwner != nil || owner.Field129 != nil {
		t.Fatalf("ownership = owner %p/head %p, want nil/nil", unit.ObjOwner, owner.Field129)
	}
	if uint32(unit.ObjSubClass) != 0x1002 {
		t.Fatalf("subclass = %#x, want 0x1002", unit.ObjSubClass)
	}
	if unit.Buffs != 0 || unit.BuffsDur[0] != 0 || unit.BuffsPower[0] != 0 {
		t.Fatalf("buff state = %#x/%d/%d", unit.Buffs, unit.BuffsDur[0], unit.BuffsPower[0])
	}
	if questPlayer.field4664 != 5 || questPlayer.field4692 != 5 {
		t.Fatalf("Quest credit = %d/%#x, want 5/0x5", questPlayer.field4664, questPlayer.field4692)
	}
}

func TestMonsterDieNative50A3D0SummonedDecaySkipsQuestRecheck(t *testing.T) {
	s := new(Server)
	s.handle = atomic.AddUintptr(&serverLast, 1)
	servers.Store(s.handle, s)
	t.Cleanup(func() { servers.Delete(s.handle) })
	s.SetTickRate(30)
	update := &MonsterUpdateData{AIStackInd: 0, StatusFlags: object.MonStatusSummoned}
	unit := &Object{
		ObjClass: object.ClassMonster, ObjSubClass: 0x2000,
		UpdateData: unsafe.Pointer(update), serverHandle: s.handle,
	}
	runtime := completeMonsterDieRuntime50A3D0()
	var flags []uint32
	runtime.GameFlag = func(flag uint32) bool {
		flags = append(flags, flag)
		return false
	}
	runtime.RandomInt = func(minimum, maximum int) int {
		if minimum != 10 || maximum != 20 {
			t.Fatalf("summoned random range = %d..%d", minimum, maximum)
		}
		return 12
	}
	runtime.SetDecayTime = func(got *Object, frames uint32) {
		if got != unit || frames != 360 {
			t.Fatalf("summoned decay = %p/%d, want %p/360", got, frames, unit)
		}
	}
	if !s.MonsterDieNative50A3D0(unit, runtime) {
		t.Fatal("summoned death was not handled")
	}
	wantFlags := []uint32{monsterDieQuestFlag50A3D0, monsterDieCoopFlag50A3D0, monsterDieQuestFlag50A3D0}
	if !reflect.DeepEqual(flags, wantFlags) {
		t.Fatalf("flag reads = %#v, want %#v", flags, wantFlags)
	}
}

func TestRecordMonsterKilled4D6170DestroyedGuard(t *testing.T) {
	player := &Player{field4664: 2, field4692: 1}
	unit := &Object{
		ObjClass:   object.ClassPlayer,
		ObjFlags:   object.FlagDestroyed,
		UpdateData: unsafe.Pointer(&PlayerUpdateData{Player: player}),
	}
	unit.RecordMonsterKilled4D6170()
	if player.field4664 != 2 || player.field4692 != 1 {
		t.Fatalf("destroyed credit = %d/%#x", player.field4664, player.field4692)
	}
	unit.ObjFlags &^= object.FlagDestroyed
	unit.RecordMonsterKilled4D6170()
	if player.field4664 != 3 || player.field4692 != 5 {
		t.Fatalf("live credit = %d/%#x, want 3/0x5", player.field4664, player.field4692)
	}
}
