package server

import (
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

const monsterMainPassiveAggressionLimit547210 = float32(0.079999998)

const monsterMainRetreatBenignStatus547210 = object.MonStatusCanSeeFriends | object.MonStatusRunning | object.MonStatusAlwaysRun

type MonsterMainRuntime547210 struct {
	AudioEvent         func(id uint32, unit *Object)
	ScriptCallback     func(block *ScriptCallback, caller, trigger *Object, event ScriptEventType)
	GUICursorActive    func() bool
	FindObjectAtCursor func(player *Object) *Object
}

// MonsterMainNative547210 handles the pointer-safe portions of GAME.EXE
// 00547210 whose result is an immediate return or a provable no-op. It leaves
// every state-changing branch to the legacy implementation until that branch
// has been ported separately.
//
// The first three cases are exact early returns from the original function:
// the staggered IDLE/GUARD throttle, an uninterruptible dependency below the
// stack head, and FlagDead. The final case covers passive monsters for which
// every later branch predicate is known to be false.
func (s *Server) MonsterMainNative547210(unit *Object) bool {
	return s.MonsterMainNativeRuntime547210(unit, MonsterMainRuntime547210{})
}

func (s *Server) MonsterMainNativeRuntime547210(unit *Object, runtime MonsterMainRuntime547210) bool {
	if unit == nil || unit.UpdateData == nil || !unit.ObjClass.Has(object.ClassMonster) {
		return false
	}
	update := unit.UpdateDataMonster()
	if update.AIStackInd < 0 || int(update.AIStackInd) >= len(update.AIStack) {
		return false
	}
	head := update.AIStackHead()
	if (head.Type() == ai.ACTION_IDLE || head.Type() == ai.ACTION_GUARD) &&
		(byte(s.Frame())+byte(unit.NetCode)-byte(update.Field137))&0xf != 0 {
		return true
	}
	for i := int(update.AIStackInd) - 1; i >= 0; i-- {
		if update.AIStack[i].Type() == ai.DEPENDENCY_UNINTERRUPTABLE {
			return true
		}
	}
	if unit.ObjFlags.Has(object.FlagDead) {
		return true
	}
	if s.monsterMainConversation547210(unit, update, runtime) {
		return true
	}
	if s.monsterMainRetreat547210(unit, update, runtime) {
		return true
	}
	if s.monsterMainPassiveCasterNoop547210(unit, update) {
		return true
	}
	if s.monsterMainPassiveNoop547210(unit, update) {
		return true
	}
	if s.monsterMainPassiveRetreatRoamTracking547210(unit, update) {
		return true
	}
	if s.monsterMainPassiveRetreatStackNoop547210(unit, update) {
		return true
	}
	if s.monsterMainRoamTracking547210(unit, update) {
		return true
	}
	if s.monsterMainAmbientIdleNoop547210(unit, update) {
		return true
	}
	if s.monsterMainDialogNoop547210(unit, update) {
		return true
	}
	if s.monsterMainScriptedFaceNoop547210(unit, update) {
		return true
	}
	if s.monsterMainWaitNoop547210(unit, update) {
		return true
	}
	if s.monsterMainQuiescentNoop547210(unit, update) {
		return true
	}
	return s.MonsterMainPassiveShopkeeper547210(unit)
}

// monsterMainConversation547210 restores the co-op under-cursor transition at
// GAME.EXE 00547287..005473E9. It runs before every buff, combat, retreat, and
// ambient branch in monster main AI.
func (s *Server) monsterMainConversation547210(unit *Object, update *MonsterUpdateData, runtime MonsterMainRuntime547210) bool {
	if !noxflags.HasGame(noxflags.GameModeCoop) ||
		runtime.GUICursorActive == nil || runtime.GUICursorActive() ||
		unit.Field5&0x10 == 0 || update.HasAction(ai.DEPENDENCY_TIME) {
		return false
	}
	host := s.Players.HostUnit()
	if host == nil || host.ObjFlags.Has(object.FlagNoUpdate) ||
		host.UpdateData == nil || !host.ObjClass.Has(object.ClassPlayer) {
		return false
	}
	hostUpdate := host.UpdateDataPlayer()
	player := hostUpdate.Player
	if player == nil {
		return false
	}
	dx := float64(player.CursorVec.X) - float64(unit.PosVec.X)
	dy := float64(player.CursorVec.Y) - float64(unit.PosVec.Y)
	if dx*dx+dy*dy >= 100 || runtime.FindObjectAtCursor == nil || runtime.FindObjectAtCursor(host) != unit {
		return false
	}

	unit.PrevPos = unit.PosVec
	unit.VelVec = types.Pointf{}
	unit.ForceVec = types.Pointf{}
	unit.Pos24 = types.Pointf{}
	unit.MonsterPushAction(ai.DEPENDENCY_NOT_MOVED)
	unit.MonsterPushAction(ai.ACTION_WAIT_RELATIVE, s.TickRate())
	unit.MonsterPushAction(ai.DEPENDENCY_UNDER_CURSOR)
	unit.MonsterPushAction(ai.ACTION_WAIT_RELATIVE, 999999)
	unit.MonsterPushAction(ai.ACTION_FACE_OBJECT, host)

	if runtime.AudioEvent != nil && update.SoundSet122 != nil && hostUpdate.Trade70 == nil && hostUpdate.DialogWith == nil {
		runtime.AudioEvent(*(*uint32)(unsafe.Add(update.SoundSet122, 4)), unit)
	}
	return true
}

// monsterMainPassiveRetreatStackNoop547210 handles the non-moving action
// temporarily placed above a low-aggression RETREAT action (most commonly the
// WAIT inserted when ROAM cannot find a waypoint). In 00547210, aggression
// below 0.08 disables threat and food reactions, while the retained RETREAT
// entry suppresses a duplicate retreat. Moving heads are left to the tracking
// routine above because they also update Field124..126.
func (s *Server) monsterMainPassiveRetreatStackNoop547210(unit *Object, update *MonsterUpdateData) bool {
	if update.AIStackInd < 0 || !update.HasAction(ai.ACTION_RETREAT) ||
		unit.Buffs != 0 || update.Aggression >= monsterMainPassiveAggressionLimit547210 ||
		update.StatusFlags.HasAny(object.MonStatusCanCastSpells|object.MonStatusCanBlock|object.MonStatusBot) ||
		update.MonsterDef == nil || update.MonsterDef.StatusFlags92&object.MonStatusCanDodge != 0 ||
		!s.monsterMainConversationImpossible547210(unit, update) {
		return false
	}
	if unit.ObjSubClass.AsMonster().Has(object.MonsterNPC) &&
		(update.WeaponEquipFlags&0x400 != 0 || update.ArmorEquipFlags&0x3000000 != 0) {
		return false
	}
	switch update.AIStackHead().Type() {
	case ai.ACTION_MOVE_TO, ai.ACTION_FAR_MOVE_TO, ai.ACTION_MOVE_TO_HOME, ai.ACTION_ROAM, ai.ACTION_FLEE:
		return false
	default:
		return true
	}
}

// monsterMainPassiveRetreatRoamTracking547210 covers the moving action placed
// above RETREAT and its food dependencies by 005455E0. Aggression below 0.08
// suppresses the enemy, food-use, and sound branches in 00547210; an existing
// RETREAT action suppresses another retreat transition. The only remaining
// work before the action update is the original movement-progress bookkeeping.
func (s *Server) monsterMainPassiveRetreatRoamTracking547210(unit *Object, update *MonsterUpdateData) bool {
	if update.AIStackInd < 0 || !update.HasAction(ai.ACTION_RETREAT) ||
		unit.Buffs != 0 || update.Aggression >= monsterMainPassiveAggressionLimit547210 ||
		update.StatusFlags.HasAny(object.MonStatusCanCastSpells|object.MonStatusCanBlock|object.MonStatusBot) ||
		update.MonsterDef == nil || update.MonsterDef.StatusFlags92&object.MonStatusCanDodge != 0 ||
		!unit.ObjFlags.Has(object.FlagEnabled) || unit.ObjFlags.Has(object.FlagDestroyed) ||
		!s.monsterMainConversationImpossible547210(unit, update) {
		return false
	}
	switch update.AIStackHead().Type() {
	case ai.ACTION_MOVE_TO, ai.ACTION_FAR_MOVE_TO, ai.ACTION_MOVE_TO_HOME, ai.ACTION_ROAM, ai.ACTION_FLEE:
	default:
		return false
	}
	if unit.ObjSubClass.AsMonster().Has(object.MonsterNPC) &&
		(update.WeaponEquipFlags&0x400 != 0 || update.ArmorEquipFlags&0x3000000 != 0) {
		return false
	}
	health := unit.HealthData
	if health == nil || health.Cur < health.Field2 {
		return false
	}
	dx := float64(math.Float32frombits(update.Field125)) - float64(unit.PosVec.X)
	dy := float64(math.Float32frombits(update.Field126)) - float64(unit.PosVec.Y)
	if dx*dx+dy*dy > 225.0 {
		update.Field124 = s.Frame()
		update.Field125 = math.Float32bits(unit.PosVec.X)
		update.Field126 = math.Float32bits(unit.PosVec.Y)
		return true
	}
	return s.Frame()-update.Field124 <= s.TickRate()/2
}

// monsterMainRetreat547210 ports the low-aggression, unbuffed portion of the
// retreat transition at GAME.EXE 00547210. The retained guards rule out every
// earlier state-changing branch before reproducing the exact action stack,
// retreat sound, and script callback order.
func (s *Server) monsterMainRetreat547210(unit *Object, update *MonsterUpdateData, runtime MonsterMainRuntime547210) bool {
	if update.AIStackInd != 0 ||
		(update.AIStack[0].Type() != ai.ACTION_IDLE && update.AIStack[0].Type() != ai.ACTION_GUARD) ||
		unit.Buffs != 0 || update.CurrentEnemy != nil || update.StatusFlags&^monsterMainRetreatBenignStatus547210 != 0 ||
		!s.monsterMainConversationImpossible547210(unit, update) ||
		update.Aggression >= monsterMainPassiveAggressionLimit547210 ||
		unit.HealthData == nil || unit.HealthData.Max == 0 || unit.SpeedBase < 0.0099999998 ||
		s.Frame()-update.Field127 < 3*s.TickRate() ||
		update.HasAction(ai.ACTION_FLEE) || update.HasAction(ai.ACTION_RETREAT) ||
		update.HasAction(ai.ACTION_RETREAT_TO_MASTER) ||
		float64(unit.HealthData.Cur)/float64(unit.HealthData.Max) > float64(update.RetreatLevel) {
		return false
	}
	unit.MonsterPushAction(ai.DEPENDENCY_NOT_CORNERED)
	retreat := ai.ACTION_RETREAT
	if noxflags.HasGame(noxflags.GameModeCoop) &&
		(update.StatusFlags.Has(object.MonStatusSummoned) || unit.ObjSubClass.AsMonster().Has(object.MonsterMonitor)) {
		retreat = ai.ACTION_RETREAT_TO_MASTER
	}
	unit.MonsterPushAction(retreat)
	if runtime.AudioEvent != nil && update.SoundSet122 != nil {
		runtime.AudioEvent(*(*uint32)(unsafe.Add(update.SoundSet122, 13*4)), unit)
	}
	if runtime.ScriptCallback != nil {
		runtime.ScriptCallback(&update.ScriptRetreat, nil, unit, NoxEventMonsterMoveXXX)
	}
	return true
}

// monsterMainPassiveCasterNoop547210 covers the low-aggression IDLE/GUARD
// state used by War01A's Wizard2. In GAME.EXE 00547210, a caster in this
// state can still react to a targeted magic missile through 005408D0 even
// though every ordinary combat and food-search branch is disabled by its
// aggression. Keep that inversion test exact and conservatively reject every
// other state-changing branch before declaring the tick a no-op.
func (s *Server) monsterMainPassiveCasterNoop547210(unit *Object, update *MonsterUpdateData) bool {
	if update.AIStackInd != 0 ||
		(update.AIStack[0].Type() != ai.ACTION_IDLE && update.AIStack[0].Type() != ai.ACTION_GUARD) ||
		unit.HasEnchant(ENCHANT_CONFUSED) || unit.HasEnchant(ENCHANT_AFRAID) ||
		!s.monsterMainConversationImpossible547210(unit, update) ||
		update.Aggression >= monsterMainPassiveAggressionLimit547210 ||
		!update.StatusFlags.Has(object.MonStatusCanCastSpells) ||
		update.StatusFlags.Has(object.MonStatusBot) ||
		update.MonsterDef == nil || update.MonsterDef.StatusFlags92&object.MonStatusCanDodge != 0 {
		return false
	}
	if unit.ObjSubClass.AsMonster().Has(object.MonsterNPC) {
		if update.WeaponEquipFlags&0x400 != 0 || update.ArmorEquipFlags&0x3000000 != 0 {
			return false
		}
	} else if update.StatusFlags.Has(object.MonStatusCanBlock) {
		return false
	}
	health := unit.HealthData
	if health == nil || health.Cur < health.Field2 {
		return false
	}
	if health.Max != 0 && unit.SpeedBase >= 0.0099999998 &&
		s.Frame()-update.Field127 >= 3*s.TickRate() &&
		!update.HasAction(ai.ACTION_FLEE) && !update.HasAction(ai.ACTION_RETREAT) &&
		!update.HasAction(ai.ACTION_RETREAT_TO_MASTER) &&
		float64(health.Cur)/float64(health.Max) <= float64(update.RetreatLevel) {
		return false
	}
	return !s.monsterMainInversionCanAct547210(unit, update)
}

// monsterMainInversionCanAct547210 implements the guards and missile query at
// GAME.EXE 005408D0. A matching missile is enough to reject the no-op path;
// the spell-selection portion is intentionally left to the remaining port.
func (s *Server) monsterMainInversionCanAct547210(unit *Object, update *MonsterUpdateData) bool {
	if !unit.ObjFlags.Has(object.FlagEnabled) ||
		!update.StatusFlags.Has(object.MonStatusCanCastSpells) ||
		unit.HasEnchant(ENCHANT_ANTI_MAGIC) ||
		s.Frame() < update.Field363 {
		return false
	}
	if action := update.AIStackHead().Type(); action >= ai.ACTION_CAST_SPELL_ON_OBJECT && action <= ai.ACTION_CAST_DURATION_SPELL {
		return false
	}
	// A missing balance table is only possible in isolated unit tests. Treat
	// it as actionable so production code never proves a no-op with an
	// unknown InversionRange.
	if s.Balance.file == nil {
		return true
	}
	return s.monsterMainInversionThreat547210(unit, float32(s.Balance.Float("InversionRange")*0.5))
}

func (s *Server) monsterMainInversionThreat547210(unit *Object, radius float32) bool {
	found := false
	s.Map.EachMissileInCircle(unit.PosVec, radius, func(missile *Object) bool {
		if missile.UpdateData != nil && missile.ObjSubClass.AsMissile().Has(object.MissileMagic) &&
			missile.UpdateDataMissile().Target == unit {
			found = true
			return false
		}
		return true
	})
	return found
}

// monsterMainScriptedFaceNoop547210 covers passive monsters temporarily facing
// a script location or object outside the dialog-owned state above. Low
// aggression rules out target acquisition, combat casting, and food search;
// the remaining gates explicitly exclude shield blocking, definition-driven
// dodging, mimic behavior, and an eligible low-health retreat.
func (s *Server) monsterMainScriptedFaceNoop547210(unit *Object, update *MonsterUpdateData) bool {
	if update.AIStackInd < 0 || int(update.AIStackInd) >= len(update.AIStack) {
		return false
	}
	head := update.AIStackHead().Type()
	if head != ai.ACTION_FACE_LOCATION && head != ai.ACTION_FACE_OBJECT ||
		unit.Buffs != 0 || update.CurrentEnemy != nil ||
		!s.monsterMainConversationImpossible547210(unit, update) ||
		update.Aggression >= monsterMainPassiveAggressionLimit547210 ||
		update.StatusFlags.HasAny(object.MonStatusCanCastSpells|object.MonStatusBot) ||
		update.MonsterDef == nil || update.MonsterDef.StatusFlags92&object.MonStatusCanDodge != 0 {
		return false
	}
	if unit.ObjSubClass.AsMonster().Has(object.MonsterNPC) {
		if update.WeaponEquipFlags&0x400 != 0 || update.ArmorEquipFlags&0x3000000 != 0 {
			return false
		}
	} else if update.StatusFlags.Has(object.MonStatusCanBlock) {
		return false
	}
	health := unit.HealthData
	if health == nil || health.Cur < health.Field2 {
		return false
	}
	if health.Max != 0 && unit.SpeedBase >= 0.0099999998 &&
		s.Frame()-update.Field127 >= 3*s.TickRate() &&
		!update.HasAction(ai.ACTION_FLEE) && !update.HasAction(ai.ACTION_RETREAT) &&
		!update.HasAction(ai.ACTION_RETREAT_TO_MASTER) &&
		float64(health.Cur)/float64(health.Max) <= float64(update.RetreatLevel) {
		return false
	}
	return true
}

// monsterMainAmbientIdleNoop547210 covers the IDLE interval between ROAM and
// WAIT for War01A's medium-low-aggression ambient creatures. GAME.EXE does
// not acquire an enemy below the 0.33000001 aggression boundary. A retained
// enemy is only actionable after the three-second move-attempt cooldown and
// inside FleeRange; full health also rules out the cadence-driven food scan.
func (s *Server) monsterMainAmbientIdleNoop547210(unit *Object, update *MonsterUpdateData) bool {
	if update.AIStackInd != 0 || update.AIStack[0].Type() != ai.ACTION_IDLE ||
		unit.Buffs != 0 || update.StatusFlags != 0 ||
		update.WeaponEquipFlags != 0 || update.ArmorEquipFlags != 0 ||
		update.MonsterDef == nil || update.MonsterDef.StatusFlags92&object.MonStatusCanDodge != 0 ||
		unit.ObjSubClass.AsMonster().Has(object.MonsterNPC) ||
		!s.monsterMainConversationImpossible547210(unit, update) ||
		update.Aggression <= monsterMainPassiveAggressionLimit547210 || update.Aggression >= 0.33000001 {
		return false
	}
	health := unit.HealthData
	if health == nil || health.Cur < health.Field2 {
		return false
	}
	moveCooldown := s.Frame()-update.Field127 < 3*s.TickRate()
	if health.Max != 0 && unit.SpeedBase >= 0.0099999998 && !moveCooldown &&
		!update.HasAction(ai.ACTION_FLEE) && !update.HasAction(ai.ACTION_RETREAT) &&
		!update.HasAction(ai.ACTION_RETREAT_TO_MASTER) &&
		float64(health.Cur)/float64(health.Max) <= float64(update.RetreatLevel) {
		return false
	}
	if enemy := update.CurrentEnemy; enemy != nil && unit.SpeedBase >= 0.0099999998 && !moveCooldown {
		delta := enemy.PosVec.Sub(unit.PosVec)
		if float64(delta.X*delta.X+delta.Y*delta.Y) < float64(update.FleeRange)*float64(update.FleeRange) {
			return false
		}
	}
	if byte(s.Frame())&0xf == 0 && health.Max != 0 && health.Cur < health.Max {
		return false
	}
	return true
}

// monsterMainDialogNoop547210 keeps a passive NPC's scripted face-target
// action running while the host player is already in that NPC's dialog. The
// main AI has no independent state change left in this exact dialog state;
// the action dispatcher below it still turns the NPC toward the player.
func (s *Server) monsterMainDialogNoop547210(unit *Object, update *MonsterUpdateData) bool {
	if !noxflags.HasGame(noxflags.GameModeCoop) || update.AIStackInd < 0 ||
		int(update.AIStackInd) >= len(update.AIStack) ||
		update.AIStackHead().Type() != ai.ACTION_FACE_OBJECT ||
		unit.Field5&0x10 == 0 || unit.Buffs != 0 || update.CurrentEnemy != nil ||
		update.StatusFlags&^object.MonStatusCanSeeFriends != 0 ||
		update.WeaponEquipFlags != 0 || update.ArmorEquipFlags != 0 ||
		unit.InvFirstItem != nil || update.MonsterDef == nil ||
		update.MonsterDef.StatusFlags92&object.MonStatusCanDodge != 0 ||
		update.Aggression >= monsterMainPassiveAggressionLimit547210 {
		return false
	}
	host := s.Players.HostUnit()
	if host == nil || host.UpdateData == nil || !host.ObjClass.Has(object.ClassPlayer) ||
		host.UpdateDataPlayer().DialogWith != unit {
		return false
	}
	health := unit.HealthData
	return health != nil && health.Cur >= health.Field2 &&
		(health.Max == 0 || unit.SpeedBase < 0.0099999998 ||
			float64(health.Cur)/float64(health.Max) > float64(update.RetreatLevel))
}

// monsterMainWaitNoop547210 covers the short WAIT inserted by ambient ROAM.
// Outside the original 16-frame scan cadence, this exact passive state cannot
// acquire food or an enemy, cast, retreat, block, dodge, or change its stack.
func (s *Server) monsterMainWaitNoop547210(unit *Object, update *MonsterUpdateData) bool {
	if update.AIStackInd != 0 || update.AIStack[0].Type() != ai.ACTION_WAIT ||
		unit.Buffs != 0 ||
		update.StatusFlags != 0 || update.WeaponEquipFlags != 0 || update.ArmorEquipFlags != 0 ||
		unit.InvFirstItem != nil || update.MonsterDef == nil ||
		update.MonsterDef.StatusFlags92&object.MonStatusCanDodge != 0 ||
		!s.monsterMainConversationImpossible547210(unit, update) {
		return false
	}
	health := unit.HealthData
	if health == nil || health.Cur < health.Field2 {
		return false
	}
	// The 16-frame cadence can only enter the edible search when the unit's
	// health is below Max. A full-health WAIT therefore remains an exact no-op
	// on the cadence frame used by War01A's ambient fish.
	if byte(s.Frame())&0xf == 0 && health.Max != 0 && health.Cur < health.Max {
		return false
	}
	// The retreat branch is suppressed for three seconds after the last move
	// attempt. This is the state used by newly loaded War01A NPCs whose current
	// health has not yet been raised from zero to Max.
	if health.Max != 0 && unit.SpeedBase >= 0.0099999998 &&
		s.Frame()-update.Field127 >= 3*s.TickRate() &&
		float64(health.Cur)/float64(health.Max) <= float64(update.RetreatLevel) {
		return false
	}
	if enemy := update.CurrentEnemy; enemy != nil {
		delta := enemy.PosVec.Sub(unit.PosVec)
		if float64(delta.X*delta.X+delta.Y*delta.Y) < float64(update.FleeRange)*float64(update.FleeRange) {
			return false
		}
	}
	return true
}

// monsterMainRoamTracking547210 covers the medium-low-aggression ROAM path
// used by War01A's ambient fish. With every combat, spell, retreat, hunger,
// block, dodge, conversation, and mimic predicate ruled out, the only work
// left in GAME.EXE 00547210 is its movement-progress bookkeeping. The
// frustration branch remains unhandled because it can push randomized AI
// actions.
func (s *Server) monsterMainRoamTracking547210(unit *Object, update *MonsterUpdateData) bool {
	if update.AIStackInd != 0 || update.AIStack[0].Type() != ai.ACTION_ROAM ||
		unit.Buffs != 0 || update.StatusFlags != 0 ||
		update.MonsterDef == nil || update.MonsterDef.StatusFlags92&object.MonStatusCanDodge != 0 ||
		unit.ObjSubClass.AsMonster().Has(object.MonsterNPC) ||
		!unit.ObjFlags.Has(object.FlagEnabled) || unit.ObjFlags.Has(object.FlagDestroyed) ||
		!s.monsterMainConversationImpossible547210(unit, update) ||
		update.Aggression < 0.079999998 || update.Aggression > 0.33000001 ||
		unit.SpeedBase < 0.0099999998 {
		return false
	}
	health := unit.HealthData
	if health == nil || health.Cur < health.Field2 ||
		health.Max != 0 && float64(health.Cur)/float64(health.Max) <= float64(update.RetreatLevel) {
		return false
	}
	// A current enemy only changes this low-aggression path when it is within
	// FleeRange (and the move-attempt cooldown has expired). Reject the whole
	// distance range conservatively so the native path never hides a flee
	// transition; farther enemies leave the bookkeeping path unchanged.
	if enemy := update.CurrentEnemy; enemy != nil {
		delta := enemy.PosVec.Sub(unit.PosVec)
		if float64(delta.X*delta.X+delta.Y*delta.Y) < float64(update.FleeRange)*float64(update.FleeRange) {
			return false
		}
	}
	dx := float64(math.Float32frombits(update.Field125)) - float64(unit.PosVec.X)
	dy := float64(math.Float32frombits(update.Field126)) - float64(unit.PosVec.Y)
	if dx*dx+dy*dy > 225.0 {
		update.Field124 = s.Frame()
		update.Field125 = math.Float32bits(unit.PosVec.X)
		update.Field126 = math.Float32bits(unit.PosVec.Y)
		return true
	}
	return s.Frame()-update.Field124 <= s.TickRate()/2
}

// monsterMainQuiescentNoop547210 covers an IDLE/GUARD monster for which all
// state-changing predicates in 00547210 are false. The missile check is
// deliberately conservative: any nearby missile leaves the unit to the
// remaining implementation, even when the original geometry test would have
// rejected that missile.
func (s *Server) monsterMainQuiescentNoop547210(unit *Object, update *MonsterUpdateData) bool {
	if update.AIStackInd != 0 ||
		(update.AIStack[0].Type() != ai.ACTION_IDLE && update.AIStack[0].Type() != ai.ACTION_GUARD) ||
		unit.Buffs != 0 || update.CurrentEnemy != nil ||
		!s.monsterMainConversationImpossible547210(unit, update) ||
		update.StatusFlags.HasAny(object.MonStatusCanCastSpells|object.MonStatusBot) ||
		update.MonsterDef == nil {
		return false
	}
	health := unit.HealthData
	if health == nil || health.Cur < health.Field2 {
		return false
	}
	if health.Max != 0 && unit.SpeedBase >= 0.0099999998 &&
		s.Frame()-update.Field127 >= 3*s.TickRate() &&
		float64(health.Cur)/float64(health.Max) <= float64(update.RetreatLevel) {
		return false
	}
	hasNearbyMissile := false
	s.Map.EachMissileInCircle(unit.PosVec, 100, func(*Object) bool {
		hasNearbyMissile = true
		return false
	})
	return !hasNearbyMissile
}

func (s *Server) monsterMainConversationImpossible547210(unit *Object, update *MonsterUpdateData) bool {
	if !noxflags.HasGame(noxflags.GameModeCoop) || unit.Field5&0x10 == 0 {
		return true
	}
	host := s.Players.HostUnit()
	if host == nil || host.ObjFlags.Has(object.FlagNoUpdate) || update.HasAction(ai.DEPENDENCY_TIME) {
		return true
	}
	if !host.ObjClass.Has(object.ClassPlayer) || host.UpdateData == nil {
		return false
	}
	player := host.UpdateDataPlayer().Player
	if player == nil {
		return false
	}
	dx := float64(player.CursorVec.X) - float64(unit.PosVec.X)
	dy := float64(player.CursorVec.Y) - float64(unit.PosVec.Y)
	return dx*dx+dy*dy >= 100.0
}

// monsterMainPassiveNoop547210 recognizes the low-aggression IDLE/GUARD state
// used by War01A's Jennifer and Horrendous. These checks deliberately describe
// branch predicates, not monster types: expanding the set without proving
// every remaining 00547210 branch would hide an unported AI transition.
func (s *Server) monsterMainPassiveNoop547210(unit *Object, update *MonsterUpdateData) bool {
	if update.AIStackInd != 0 ||
		(update.AIStack[0].Type() != ai.ACTION_IDLE && update.AIStack[0].Type() != ai.ACTION_GUARD) ||
		unit.Buffs != 0 ||
		!(update.Aggression < monsterMainPassiveAggressionLimit547210) {
		return false
	}
	// CurrentEnemy is only inspected by the movement/flee block guarded by
	// !sub_534440, and inventory is only searched by the final edible block
	// behind the same guard. Both are therefore inert below aggression 0.08.
	// In co-op, xstatus 0x10 can open the under-cursor conversation stack.
	if noxflags.HasGame(noxflags.GameModeCoop) && unit.Field5&0x10 != 0 {
		return false
	}
	// Spell inversion, missile blocking, and mimic morphing are independent of
	// aggression and therefore must be ruled out explicitly.
	if update.StatusFlags.HasAny(object.MonStatusCanCastSpells | object.MonStatusCanBlock | object.MonStatusBot) {
		return false
	}
	// NPC weapon and armor flags can enter either shield-block branch even when
	// StatusFlags itself has no blocking capability.
	if uint8(unit.ObjSubClass)&0x10 != 0 &&
		(update.WeaponEquipFlags&0x400 != 0 || update.ArmorEquipFlags&0x3000000 != 0) {
		return false
	}
	// The retreat branch is skipped for Max==0, stationary monsters, or health
	// strictly above the configured threshold. Use float64 to match the x87
	// comparison of the two uint16 health values against the stored float32.
	health := unit.HealthData
	if health == nil {
		return false
	}
	if health.Max != 0 && unit.SpeedBase >= 0.0099999998 &&
		s.Frame()-update.Field127 >= 3*s.TickRate() &&
		!update.HasAction(ai.ACTION_FLEE) && !update.HasAction(ai.ACTION_RETREAT) &&
		!update.HasAction(ai.ACTION_RETREAT_TO_MASTER) &&
		float64(health.Cur)/float64(health.Max) <= float64(update.RetreatLevel) {
		return false
	}
	return true
}

// MonsterMainPassiveShopkeeper547210 handles the exact no-op path used by a
// passive shopkeeper in a regular game. The original function reaches no
// state-changing branch when the shopkeeper is idle or guarding, healthy,
// unbuffed, unarmed, has no target, and has no monster status capabilities
// enabled.
//
// It returns false for every other state so that callers cannot accidentally
// treat this narrow, verified path as the complete 00547210 implementation.
func (s *Server) MonsterMainPassiveShopkeeper547210(unit *Object) bool {
	if unit == nil || unit.UpdateData == nil ||
		!unit.ObjClass.Has(object.ClassMonster) ||
		!unit.ObjSubClass.AsMonster().Has(object.MonsterShopkeeper) ||
		noxflags.HasGame(noxflags.GameModeQuest) ||
		(noxflags.HasGame(noxflags.GameModeCoop) && unit.Field5&0x10 != 0) ||
		unit.ObjFlags.HasAny(object.FlagDead|object.FlagDestroyed) ||
		unit.Buffs != 0 || unit.InvFirstItem != nil {
		return false
	}
	update := unit.UpdateDataMonster()
	if update.AIStackInd != 0 ||
		(update.AIStack[0].Type() != ai.ACTION_IDLE && update.AIStack[0].Type() != ai.ACTION_GUARD) ||
		update.CurrentEnemy != nil || update.StatusFlags != 0 ||
		update.WeaponEquipFlags != 0 || update.ArmorEquipFlags != 0 ||
		update.MonsterDef == nil || update.MonsterDef.StatusFlags92&8 != 0 {
		return false
	}
	health := unit.HealthData
	if health == nil || health.Cur < health.Field2 || health.Max != 0 && health.Cur < health.Max {
		return false
	}
	return true
}
