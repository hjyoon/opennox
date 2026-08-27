package server

import (
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

const (
	monsterDieCoopFlag50A3D0  = uint32(0x0800)
	monsterDieQuestFlag50A3D0 = uint32(0x1000)
)

// MonsterDieRuntime50A3D0 contains the services that remain owned by the
// outer game package while GAME.EXE 0050A3D0 manipulates native-width Object
// and Player pointers in package server.
type MonsterDieRuntime50A3D0 struct {
	GameFlag        func(uint32) bool
	IsZombie        func(*Object) bool
	ObserveClear    func(*Object)
	QuestPrepare    func(*Object)
	RemoveShadow    func(*Object)
	RandomInt       func(int, int) int
	SetDecayTime    func(*Object, uint32)
	NetFxShield     func(int, *Object)
	UnmarkMinimap   func(int, *Object, uint32)
	DropAllItems    func(*Object)
	AwardSoloKill   func(*Object)
	CreditQuestKill func(*Object)
	Unsupported     func(string, *Object)
}

func monsterDieUnsupported50A3D0(runtime MonsterDieRuntime50A3D0, reason string, unit *Object) bool {
	if runtime.Unsupported != nil {
		runtime.Unsupported(reason, unit)
	}
	return false
}

// MonsterDieNative50A3D0 restores the ordinary monster-death dispatcher at
// GAME.EXE 0050A3D0. All admission checks that depend on an outer callback are
// completed before the action stack or object flags are changed.
func (s *Server) MonsterDieNative50A3D0(unit *Object, runtime MonsterDieRuntime50A3D0) bool {
	if unit == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassMonster) {
		return false
	}
	update := unit.UpdateDataMonster()
	quest := runtime.GameFlag != nil && runtime.GameFlag(monsterDieQuestFlag50A3D0)
	coop := runtime.GameFlag != nil && runtime.GameFlag(monsterDieCoopFlag50A3D0)
	zombie := runtime.IsZombie != nil && runtime.IsZombie(unit)

	var observers []*Object
	for playerUnit := s.Players.FirstUnit(); playerUnit != nil; playerUnit = s.Players.NextUnit(playerUnit) {
		player := playerUnit.ControllingPlayer()
		if player != nil && player.ObserveTarget() == unit {
			observers = append(observers, playerUnit)
		}
	}
	if len(observers) != 0 && runtime.ObserveClear == nil {
		return monsterDieUnsupported50A3D0(runtime, "observed monster", unit)
	}
	if quest && runtime.QuestPrepare == nil {
		return monsterDieUnsupported50A3D0(runtime, "Quest death preparation", unit)
	}
	if !zombie && unit.ObjFlags.Has(object.FlagShadow) && runtime.RemoveShadow == nil {
		return monsterDieUnsupported50A3D0(runtime, "shadow removal", unit)
	}
	needsDecay := !zombie && (quest || update.StatusFlags.Has(object.MonStatusSummoned))
	if needsDecay && (runtime.RandomInt == nil || runtime.SetDecayTime == nil) {
		return monsterDieUnsupported50A3D0(runtime, "decay scheduling", unit)
	}
	owner := unit.ObjOwner
	ownerPlayer := owner != nil && owner.Class().Has(object.ClassPlayer)
	if !zombie && ownerPlayer && (runtime.NetFxShield == nil || runtime.UnmarkMinimap == nil) {
		return monsterDieUnsupported50A3D0(runtime, "owner minimap cleanup", unit)
	}
	dropsItems := !zombie && uint32(unit.SubClass())&0x2000 == 0 && unit.InvFirstItem != nil
	if dropsItems && runtime.DropAllItems == nil {
		return monsterDieUnsupported50A3D0(runtime, "inventory drop", unit)
	}
	killer := unit.Obj130.FindOwnerChainPlayer()
	awardsSolo := !zombie && !coop && !quest && update.Field547 == 2 && update.Field546 == 2 &&
		killer != nil && killer.Class().Has(object.ClassPlayer)
	if awardsSolo && runtime.AwardSoloKill == nil {
		return monsterDieUnsupported50A3D0(runtime, "solo kill award", unit)
	}
	creditsQuest := !zombie && quest && killer != nil && killer.Class().Has(object.ClassPlayer)
	if creditsQuest && runtime.CreditQuestKill == nil {
		return monsterDieUnsupported50A3D0(runtime, "Quest kill credit", unit)
	}

	if quest {
		runtime.QuestPrepare(unit)
	}
	for _, playerUnit := range observers {
		runtime.ObserveClear(playerUnit)
	}
	unit.ClearActionStack()
	unit.MonsterPushAction(ai.ACTION_DEAD)
	unit.MonsterPushAction(ai.ACTION_DYING)
	if zombie {
		return true
	}

	unit.ObjFlags &^= object.FlagMissileHit
	if unit.ObjFlags.Has(object.FlagShadow) {
		runtime.RemoveShadow(unit)
	}
	unit.SetBuffFlags(0, nil)
	for i := range unit.BuffsDur {
		unit.BuffsDur[i] = 0
		unit.BuffsPower[i] = 0
	}
	if needsDecay {
		minimum, maximum := 5, 8
		if update.StatusFlags.Has(object.MonStatusSummoned) {
			minimum, maximum = 10, 20
		}
		runtime.SetDecayTime(unit, s.TickRate()*uint32(runtime.RandomInt(minimum, maximum)))
	}
	if ownerPlayer {
		unit.ObjSubClass &^= 0x80
		player := owner.ControllingPlayer()
		if player != nil {
			index := player.Index()
			runtime.NetFxShield(index, unit)
			runtime.UnmarkMinimap(index, unit, 1)
		}
	}
	unit.ObjSubClass &^= 0x100
	s.ObjTransferSlaves(unit)
	s.ObjClearOwner(unit)
	if dropsItems {
		runtime.DropAllItems(unit)
	}
	if awardsSolo {
		runtime.AwardSoloKill(killer)
	}
	if creditsQuest {
		runtime.CreditQuestKill(killer)
	}
	return true
}
