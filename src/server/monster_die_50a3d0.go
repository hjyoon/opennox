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

func monsterDieRuntimeReady50A3D0(runtime MonsterDieRuntime50A3D0, unit *Object) bool {
	var missing string
	switch {
	case runtime.GameFlag == nil:
		missing = "GameFlag"
	case runtime.IsZombie == nil:
		missing = "IsZombie"
	case runtime.ObserveClear == nil:
		missing = "ObserveClear"
	case runtime.QuestPrepare == nil:
		missing = "QuestPrepare"
	case runtime.RemoveShadow == nil:
		missing = "RemoveShadow"
	case runtime.RandomInt == nil:
		missing = "RandomInt"
	case runtime.SetDecayTime == nil:
		missing = "SetDecayTime"
	case runtime.NetFxShield == nil:
		missing = "NetFxShield"
	case runtime.UnmarkMinimap == nil:
		missing = "UnmarkMinimap"
	case runtime.DropAllItems == nil:
		missing = "DropAllItems"
	case runtime.AwardSoloKill == nil:
		missing = "AwardSoloKill"
	case runtime.CreditQuestKill == nil:
		missing = "CreditQuestKill"
	default:
		return true
	}
	return monsterDieUnsupported50A3D0(runtime, "missing "+missing, unit)
}

// MonsterDieNative50A3D0 restores the monster-death dispatcher at GAME.EXE
// 0050A3D0. Once the outer services are bound, all flags and object links are
// read at the same points as the original so callbacks can mutate later state.
func (s *Server) MonsterDieNative50A3D0(unit *Object, runtime MonsterDieRuntime50A3D0) bool {
	if unit == nil || unit.UpdateData == nil || !unit.Class().Has(object.ClassMonster) {
		return false
	}
	if !monsterDieRuntimeReady50A3D0(runtime, unit) {
		return false
	}
	update := unit.UpdateDataMonster()
	if runtime.GameFlag(monsterDieQuestFlag50A3D0) {
		runtime.QuestPrepare(unit)
	}
	for playerUnit := s.Players.FirstUnit(); playerUnit != nil; playerUnit = s.Players.NextUnit(playerUnit) {
		if playerUnit.ControllingPlayer().ObserveTarget() == unit {
			runtime.ObserveClear(playerUnit)
		}
	}
	unit.ClearActionStack()
	unit.MonsterPushAction(ai.ACTION_DEAD)
	unit.MonsterPushAction(ai.ACTION_DYING)
	if runtime.IsZombie(unit) {
		return true
	}

	unit.ObjFlags &^= object.FlagMissileHit
	runtime.RemoveShadow(unit)
	unit.UnitBuffClear4FF580(UnitBuffClearRuntime4FF580{})
	if int8(uint8(update.StatusFlags)) < 0 {
		runtime.SetDecayTime(unit, s.TickRate()*uint32(runtime.RandomInt(10, 20)))
	} else if runtime.GameFlag(monsterDieQuestFlag50A3D0) {
		runtime.SetDecayTime(unit, s.TickRate()*uint32(runtime.RandomInt(5, 8)))
	}
	owner := unit.ObjOwner
	if owner != nil && owner.Class().Has(object.ClassPlayer) {
		unit.ObjSubClass &^= 0x80
		player := owner.ControllingPlayer()
		index := player.Index()
		runtime.NetFxShield(index, unit)
		runtime.UnmarkMinimap(index, unit, 1)
	}
	unit.ObjSubClass &^= 0x100
	s.ObjTransferSlaves(unit)
	s.ObjClearOwner(unit)
	if uint32(unit.SubClass())&0x2000 == 0 {
		runtime.DropAllItems(unit)
	}
	if !runtime.GameFlag(monsterDieCoopFlag50A3D0) && update.Field547 == 2 && update.Field546 == 2 {
		if source := unit.Obj130; source != nil {
			killer := source.FindOwnerChainPlayer()
			if killer.Class().Has(object.ClassPlayer) {
				runtime.AwardSoloKill(killer)
			}
		}
	}
	if runtime.GameFlag(monsterDieQuestFlag50A3D0) {
		if source := unit.Obj130; source != nil {
			killer := source.FindOwnerChainPlayer()
			if killer.Class().Has(object.ClassPlayer) {
				runtime.CreditQuestKill(killer)
			}
		}
	}
	return true
}
