package legacy

/*
#include "defs.h"
#include "GAME1_1.h"
#include "GAME2_2.h"
#include "GAME3_3.h"
#include "GAME4.h"
#include "GAME4_1.h"
#include "GAME4_2.h"
#include "GAME4_3.h"
#include "GAME5.h"
#include "server__script__script.h"
extern unsigned int dword_5d4594_2489460;
*/
import "C"
import (
	"math"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

var (
	Nox_xxx_gameSetAudioFadeoutMb_501AC0 func(v int)
	Nox_xxx_unitUpdateMonster_50A5C0     func(a1 *server.Object)
)

type Nox_player_polygon_check_data struct {
	Field_0 [35]uint32
}

func init() {
	for typ, a := range map[ai.ActionType]struct {
		Start, Update, End, Cancel unsafe.Pointer
	}{
		ai.ACTION_ESCORT:            {Update: C.nox_xxx_mobActionEscort_546430, End: C.sub_546410, Cancel: C.sub_546420},
		ai.ACTION_GUARD:             {Update: C.nox_xxx_mobActionGuard_546010},
		ai.ACTION_HUNT:              {Update: C.nox_xxx_mobActionHunt_5449D0},
		ai.ACTION_RETREAT:           {Update: C.nox_xxx_mobActionRetreat_545440},
		ai.ACTION_MOVE_TO:           {Update: C.nox_xxx_mobActionMoveTo_5443F0},
		ai.ACTION_FAR_MOVE_TO:       {Update: C.nox_xxx_mobActionMoveToFar_5445C0},
		ai.ACTION_DODGE:             {Update: C.nox_xxx_mobActionDodge_544640},
		ai.ACTION_ROAM:              {Start: C.sub_545790, Update: C.nox_xxx_mobActionRoam_5457E0, Cancel: C.sub_5457C0},
		ai.ACTION_PICKUP_OBJECT:     {Update: C.nox_xxx_mobActionPickupObject_544B90},
		ai.ACTION_RETREAT_TO_MASTER: {Start: C.sub_5456B0, Update: C.sub_5456D0, End: C.sub_5456C0},
		ai.ACTION_FIGHT:             {Start: C.nox_xxx_mobActionFightStart_531E20, Update: C.nox_xxx_mobActionFight_531EC0, End: C.sub_531E90},
		ai.ACTION_MELEE_ATTACK:      {Start: C.nox_xxx_mobActionMelee1_532130, Update: C.nox_xxx_mobActionMeleeAtt_532440, Cancel: C.nox_ai_action_pop_532100},
		ai.ACTION_MISSILE_ATTACK:    {Start: C.sub_532540, Update: C.nox_xxx_mobActionMissileAtt_532610, Cancel: C.nox_ai_action_pop_532100},
		ai.ACTION_BLOCK_ATTACK:      {Update: C.nox_xxx_monsterShieldBlockStart_532070, Cancel: C.nox_ai_action_pop_532100},
		ai.ACTION_BLOCK_FINISH:      {Update: C.nox_xxx_monsterShieldBlockStop_5320E0, Cancel: C.nox_ai_action_pop_532100},
		ai.ACTION_WEAPON_BLOCK:      {Update: C.sub_532110, Cancel: C.nox_ai_action_pop_532100},
		ai.ACTION_FLEE:              {Start: C.sub_544740, Update: C.nox_xxx_mobActionFlee_544760, End: C.sub_544750},
		ai.ACTION_FACE_LOCATION:     {Update: C.sub_545210, Cancel: C.nox_ai_action_pop_532100},
		ai.ACTION_FACE_OBJECT:       {Update: C.sub_545300, Cancel: C.nox_ai_action_pop_532100},
		ai.ACTION_FACE_ANGLE:        {Update: C.sub_545340, Cancel: C.nox_ai_action_pop_532100},
		ai.ACTION_SET_ANGLE:         {Update: C.sub_5453E0, Cancel: C.nox_ai_action_pop_532100},
		ai.ACTION_RANDOM_WALK:       {Update: C.nox_xxx_mobActionRandomWalk_545020},
		ai.ACTION_DYING:             {Start: C.nox_xxx_mobGenericDeath_544C40, Update: C.sub_544D60, End: C.nox_xxx_zombieBurnDeleteCheck_544CA0},
		ai.ACTION_DEAD:              {Start: C.nox_xxx_mobActionDead1_544D80, Update: C.nox_xxx_mobActionDead2_544EC0},
		ai.ACTION_GET_UP:            {Update: C.nox_xxx_mobActionGetUp_534A90},
		ai.ACTION_CONFUSED:          {Update: C.nox_xxx_mobActionConfuse_545140},
		ai.ACTION_MOVE_TO_HOME:      {Start: C.nox_xxx_mobActionReturnToHome_544920, Update: C.sub_544950, End: C.sub_544930, Cancel: C.sub_544940},
	} {
		server.RegisterAIAction(cgoAIAction{typ: typ, start: a.Start, update: a.Update, end: a.End, cancel: a.Cancel})
	}
}

type cgoAIAction struct {
	typ                        ai.ActionType
	start, update, end, cancel unsafe.Pointer
}

func (a cgoAIAction) Type() ai.ActionType {
	return a.typ
}

func (a cgoAIAction) Start(u *server.Object) {
	switch a.typ {
	case ai.ACTION_FIGHT:
		GetServer().S().MonsterActionFightStart531E20(u, server.MonsterActionFightStartRuntime531E20{
			AudioEvent: func(id uint32, unit *server.Object) {
				C.nox_xxx_aud_501960(C.int(id), asObjectC(unit), 0, 0)
			},
			ScriptCallback: func(block *server.ScriptCallback, caller, trigger *server.Object, event server.ScriptEventType) {
				GetServer().NoxScriptC().ScriptCallback(block, caller, trigger, event)
			},
			CopyFrameCounter: func() {
				*memmap.PtrUint32(0x5D4594, 2487684) = GetServer().S().Frame()
			},
			UpdateSight: Nox_xxx_unitUpdateSightMB_5281F0,
		})
		return
	case ai.ACTION_MELEE_ATTACK:
		if GetServer().S().MonsterActionMeleeStart532130(u, monsterActionMeleeRuntime532130()) {
			return
		}
	case ai.ACTION_FLEE:
		GetServer().S().MonsterActionRunStart534750(u)
		return
	case ai.ACTION_ROAM:
		GetServer().S().MonsterActionRoamStart545790(u)
		return
	case ai.ACTION_DYING:
		if GetServer().S().MonsterActionDyingStart544C40(u, monsterActionDyingRuntime544C40()) {
			return
		}
		if unsafe.Sizeof(uintptr(0)) != 4 {
			GetServer().DelayedDelete(u)
			return
		}
	case ai.ACTION_DEAD:
		if GetServer().S().MonsterActionDeadStart544D80(u, monsterActionDeadRuntime544D80()) {
			return
		}
		if unsafe.Sizeof(uintptr(0)) != 4 {
			GetServer().DelayedDelete(u)
			return
		}
	}
	if a.start != nil {
		ccall.CallVoidPtr(a.start, u.CObj())
	}
}

func (a cgoAIAction) Update(u *server.Object) {
	switch a.typ {
	case ai.ACTION_FIGHT:
		if GetServer().S().MonsterActionFight531EC0(u) {
			return
		}
	case ai.ACTION_MELEE_ATTACK:
		if GetServer().S().MonsterActionMeleeUpdate532440(u, monsterActionMeleeRuntime532130()) {
			return
		}
	case ai.ACTION_RETREAT:
		GetServer().S().MonsterActionRetreat545440(u)
		return
	case ai.ACTION_MOVE_TO:
		s := GetServer()
		s.S().MonsterActionMoveTo5443F0(u, s.Nox_xxx_creatureSetDetailedPath_50D220)
		return
	case ai.ACTION_FLEE:
		s := GetServer()
		s.S().MonsterActionFlee544760(u, s.Nox_xxx_generateRetreatPath_50CA00)
		return
	case ai.ACTION_FACE_LOCATION:
		GetServer().S().MonsterActionFaceLocation545210(u)
		return
	case ai.ACTION_FACE_OBJECT:
		GetServer().S().MonsterActionFaceObject545300(u)
		return
	case ai.ACTION_FACE_ANGLE:
		GetServer().S().MonsterActionFaceAngle545340(u)
		return
	case ai.ACTION_SET_ANGLE:
		GetServer().S().MonsterActionSetAngle5453E0(u)
		return
	case ai.ACTION_RANDOM_WALK:
		GetServer().S().MonsterActionRandomWalk545020(u, Nox_xxx_tileNFromPoint_411160)
		return
	case ai.ACTION_DYING:
		if GetServer().S().MonsterActionDyingUpdate544D60(u) {
			return
		}
	case ai.ACTION_DEAD:
		if GetServer().S().MonsterActionDeadUpdate544EC0(u, monsterActionDeadRuntime544D80()) {
			return
		}
		if unsafe.Sizeof(uintptr(0)) != 4 {
			GetServer().DelayedDelete(u)
			return
		}
	}
	if a.typ == ai.ACTION_GUARD {
		s := GetServer()
		s.S().MonsterActionGuard546010(u, Sub_5466F0)
		return
	}
	if a.typ == ai.ACTION_ROAM {
		s := GetServer()
		s.S().MonsterActionRoam5457E0(u, Sub_5466F0, s.Sub_50CB20, s.Nox_xxx_creatureSetDetailedPath_50D220)
		return
	}
	if a.update != nil {
		ccall.CallVoidPtr(a.update, u.CObj())
	}
}

func (a cgoAIAction) End(u *server.Object) {
	switch a.typ {
	case ai.ACTION_FIGHT:
		GetServer().S().MonsterActionFightEnd531E90(u)
		return
	case ai.ACTION_FLEE:
		GetServer().S().MonsterActionRunEnd534780(u)
		return
	case ai.ACTION_DYING:
		if GetServer().S().MonsterActionDyingEnd544CA0(u, monsterActionDyingRuntime544C40()) {
			return
		}
		if unsafe.Sizeof(uintptr(0)) != 4 {
			GetServer().DelayedDelete(u)
			return
		}
	}
	if a.end != nil {
		ccall.CallVoidPtr(a.end, u.CObj())
	}
}

func monsterActionDyingRuntime544C40() server.MonsterActionDyingRuntime544C40 {
	srv := GetServer()
	s := srv.S()
	return server.MonsterActionDyingRuntime544C40{
		AudioEvent: func(id uint32, unit *server.Object) {
			s.Audio.EventObj(sound.ID(id), unit, 0, 0)
		},
		ScriptCallback: func(block *server.ScriptCallback, caller, trigger *server.Object, event server.ScriptEventType) {
			srv.NoxScriptC().ScriptCallback(block, caller, trigger, event)
		},
		CanDieFunc: func(unsafe.Pointer) bool {
			return false
		},
		IsZombie: s.IsZombie,
		Unsupported: func(reason string, unit *server.Object) {
			if s.Log != nil {
				s.Log.Error("Monster ACTION_DYING native branch is not ported", "reason", reason, "unit_ptr", uintptr(unit.CObj()))
			}
		},
	}
}

func monsterActionDeadRuntime544D80() server.MonsterActionDeadRuntime544D80 {
	srv := GetServer()
	s := srv.S()
	return server.MonsterActionDeadRuntime544D80{
		IsZombie: s.IsZombie,
		CanDeadFunc: func(fnc unsafe.Pointer) bool {
			return fnc == unsafe.Pointer(C.sub_54A250)
		},
		DeadFunc: func(fnc unsafe.Pointer, unit *server.Object) {
			if fnc == unsafe.Pointer(C.sub_54A250) {
				s.Nox_xxx_netSendPointFx_522FF0(netmsg.MSG_FX_BLUE_SPARKS, unit.Pos())
			}
		},
		RemoveUpdatable: s.Objs.RemoveFromUpdatable,
		DelayedDelete:   srv.DelayedDelete,
		Unsupported: func(reason string, unit *server.Object) {
			if s.Log != nil {
				s.Log.Error("Monster ACTION_DEAD native branch is not ported", "reason", reason, "unit_ptr", uintptr(unit.CObj()))
			}
		},
	}
}

func (a cgoAIAction) Cancel(u *server.Object) {
	switch a.typ {
	case ai.ACTION_MELEE_ATTACK:
		u.MonsterPopAction()
		return
	case ai.ACTION_FACE_LOCATION, ai.ACTION_FACE_OBJECT, ai.ACTION_FACE_ANGLE, ai.ACTION_SET_ANGLE:
		u.MonsterPopAction()
		return
	}
	if a.typ == ai.ACTION_ROAM {
		GetServer().S().MonsterActionRoamCancel5457C0(u)
		return
	}
	if a.cancel != nil {
		ccall.CallVoidPtr(a.cancel, u.CObj())
	}
}

func monsterActionMeleeCanStrike532440(fnc unsafe.Pointer) bool {
	return fnc == unsafe.Pointer(C.nox_xxx_strikeMonsterDefault_549380) ||
		fnc == unsafe.Pointer(C.nox_xxx_strikeSpider_549BC0) ||
		fnc == unsafe.Pointer(C.nox_xxx_strikeSpittingSpider_549CA0)
}

func monsterActionMeleeRuntime532130() server.MonsterActionMeleeRuntime532130 {
	return server.MonsterActionMeleeRuntime532130{
		AudioEvent: func(id uint32, unit *server.Object) {
			C.nox_xxx_aud_501960(C.int(id), asObjectC(unit), 0, 0)
		},
		BuffOff:   Nox_xxx_spellBuffOff_4FF5B0,
		CanStrike: monsterActionMeleeCanStrike532440,
		Strike: func(unit *server.Object, fnc unsafe.Pointer) int {
			if !monsterActionMeleeCanStrike532440(fnc) {
				return 0
			}
			if fnc == unsafe.Pointer(C.nox_xxx_strikeMonsterDefault_549380) {
				return GetServer().S().MonsterStrikeDefault549380(unit, server.MonsterStrikeDefaultRuntime549380{
					Damage: func(target, source, attacker *server.Object, damage int, damageType object.DamageType) bool {
						return target.CallDamage(source, attacker, damage, damageType)
					},
					ApplyForce: func(target *server.Object, origin types.Pointf, force float64) {
						GetServer().ApplyForce(target, origin, force)
					},
				})
			}
			return GetServer().S().MonsterStrikeSpider549BC0(unit, server.MonsterStrikeSpiderRuntime549BC0{
				Damage: func(target, source, attacker *server.Object, damage int, damageType object.DamageType) bool {
					return target.CallDamage(source, attacker, damage, damageType)
				},
				ApplyForce: func(target *server.Object, origin types.Pointf, force float64) {
					GetServer().ApplyForce(target, origin, force)
				},
				ActivatePoison: Nox_xxx_activatePoison_4EE7E0,
				PriorityMessage: func(target *server.Object, id strman.ID, value byte) {
					GetServer().S().NetPriMsgToPlayer(target, id, value)
				},
			})
		},
	}
}

//export nox_ai_debug_print
func nox_ai_debug_print(str *C.char) {
	if noxflags.HasEngine(noxflags.EngineShowAI) {
		ai.Log.Printf("%s", GoString(str))
	}
}

//export sub_545E60
func sub_545E60(a1c *nox_object_t) int32 { return int32(asObjectS(a1c).Sub_545E60()) }

//export nox_xxx_gameSetAudioFadeoutMb_501AC0
func nox_xxx_gameSetAudioFadeoutMb_501AC0(v_cgo int32) {
	v := int(v_cgo)
	Nox_xxx_gameSetAudioFadeoutMb_501AC0(v)
}

//export nox_xxx_monsterPopAction_50A160
func nox_xxx_monsterPopAction_50A160(a1 *nox_object_t) int32 {
	return int32(asObjectS(a1).MonsterPopAction())
}

//export nox_xxx_monsterPushAction_50A260_impl
func nox_xxx_monsterPushAction_50A260_impl(u *nox_object_t, act_cgo int32, file *C.char, line_cgo int32) unsafe.Pointer {
	act := int(act_cgo)
	line := int(line_cgo)
	return asObjectS(u).MonsterPushActionImpl(ai.ActionType(act), GoString(file), line).C()
}

//export nox_xxx_unitUpdateMonster_50A5C0
func nox_xxx_unitUpdateMonster_50A5C0(a1 *nox_object_t) {
	Nox_xxx_unitUpdateMonster_50A5C0(asObjectS(a1))
}

func monsterSightKillable528190(obj *server.Object) bool {
	if obj == nil || obj.HealthData == nil {
		return false
	}
	return obj.HealthData.Cur != 0 || obj.HealthData.Max == 0
}

func monsterLostSightNative528560(monster *server.Object, index int) bool {
	if monster == nil || !monster.Class().Has(object.ClassMonster) {
		return false
	}
	ud := monster.UpdateDataMonster()
	if index < 0 || index >= ud.MonsterSeenEnemyCount528560() {
		return false
	}
	target := ud.SeenEnemies[index]
	if target != nil {
		if noxflags.HasEngine(noxflags.EngineShowAI) {
			ai.Log.Printf("%d: Lost sight of type %d(#%d)\n", GetServer().S().Frame(), target.TypeInd, target.NetCode)
		}
		GetServer().NoxScriptC().ScriptCallback(&ud.ScriptLostEnemy, target, monster, server.NoxEventMonsterLostEnemy)
	}
	_, ok := ud.MonsterRemoveSeenEnemyAt528560(index)
	return ok
}

func monsterSelectEnemyNative528610(monster *server.Object) {
	if monster == nil || !monster.Class().Has(object.ClassMonster) {
		return
	}
	ud := monster.UpdateDataMonster()
	ud.CurrentEnemy = nil
	bestDistance := float32(100000000.0)
	for i := 0; i < ud.MonsterSeenEnemyCount528560(); i++ {
		target := ud.SeenEnemies[i]
		if target == ud.PreferredEnemy {
			ud.CurrentEnemy = target
			return
		}
		if target == nil || !GetServer().S().IsEnemyTo(monster, target) || !monsterSightKillable528190(target) {
			continue
		}
		delta := target.Pos().Sub(monster.Pos())
		distance := delta.X*delta.X + delta.Y*delta.Y
		if distance < bestDistance {
			bestDistance = distance
			ud.CurrentEnemy = target
		}
	}
}

func monsterSeeEnemyNative5287B0(monster, target *server.Object) {
	if monster == nil || target == nil || !monster.Class().Has(object.ClassMonster) {
		return
	}
	ud := monster.UpdateDataMonster()
	if ud.MonsterSeenEnemyCount528560() == len(ud.SeenEnemies) {
		farthestIndex := -1
		farthestDistance := float32(0)
		for i, seen := range ud.SeenEnemies {
			if seen == nil {
				continue
			}
			delta := seen.Pos().Sub(monster.Pos())
			distance := delta.X*delta.X + delta.Y*delta.Y
			if distance > farthestDistance {
				farthestDistance = distance
				farthestIndex = i
			}
		}
		delta := target.Pos().Sub(monster.Pos())
		if farthestIndex < 0 || farthestDistance <= delta.X*delta.X+delta.Y*delta.Y {
			return
		}
		monsterLostSightNative528560(monster, farthestIndex)
	}
	if !ud.MonsterAppendSeenEnemy5287B0(target) {
		return
	}

	s := GetServer().S()
	frame := s.Frame()
	if frame > ud.Field134 && s.IsEnemyTo(monster, target) && (!s.IsZombie(monster) || !monster.Flags().Has(object.FlagDead)) {
		if soundSet := ud.SoundSet122; soundSet != nil {
			sound := *(*uint32)(unsafe.Add(soundSet, 68))
			if sound != 0 {
				C.nox_xxx_aud_501960(C.int(sound), asObjectC(monster), 0, 0)
			}
		}
		ud.Field134 = frame + uint32(s.Rand.Logic.IntClamp(int(2*s.TickRate()), int(4*s.TickRate())))
	}
	GetServer().NoxScriptC().ScriptCallback(&ud.ScriptEnemySighted, target, monster, server.NoxEventMonsterSeeEnemy)
}

func monsterUpdateSeenNative5286D0(candidate, monster *server.Object) {
	if candidate == nil || monster == nil || candidate == monster || !monster.Class().Has(object.ClassMonster) {
		return
	}
	if !candidate.Class().HasAny(object.ClassMonster|object.ClassPlayer) || candidate.Flags().HasAny(object.FlagDead|object.FlagDestroyed) {
		return
	}
	ud := monster.UpdateDataMonster()
	s := GetServer().S()
	if !ud.StatusFlags.Has(object.MonsterStatus(0x400)) && !s.IsEnemyTo(monster, candidate) {
		return
	}
	if ud.MonsterHasSeenEnemy528950(candidate) {
		return
	}
	if !ud.StatusFlags.Has(object.MonsterStatus(0x100)) {
		delta := candidate.Pos().Sub(monster.Pos())
		distance := float32(math.Sqrt(float64(delta.X*delta.X+delta.Y*delta.Y))) + 0.001
		direction := monster.Direction1.Vec()
		if delta.Y/distance*direction.Y+delta.X/distance*direction.X < 0.5 {
			return
		}
	}
	if s.CanInteract(monster, candidate, 0) {
		monsterSeeEnemyNative5287B0(monster, candidate)
	}
}

func monsterUpdateSightNative5281F0(monster *server.Object) {
	if monster == nil || !monster.Class().Has(object.ClassMonster) {
		return
	}
	s := GetServer().S()
	if monster.Flags().Has(object.FlagDead) && !s.IsZombie(monster) {
		return
	}
	ud := monster.UpdateDataMonster()
	sightRange := float32(250.0)
	if noxflags.HasGame(noxflags.GameModeQuest) {
		sightRange = 640.0
	}
	if ud.SightRange > sightRange {
		sightRange = ud.SightRange
	}

	frame := s.Frame()
	checkInteraction := frame-ud.Field303 > 2*s.TickRate()
	if checkInteraction {
		ud.Field303 = frame
	}
	changed := false
	for i := 0; i < ud.MonsterSeenEnemyCount528560(); {
		target := ud.SeenEnemies[i]
		remove := target == nil
		if target != nil {
			delta := monster.Pos().Sub(target.Pos())
			moved := monster.Pos().Sub(monster.PrevPos)
			limit := sightRange + 30.0
			remove = target.Flags().HasAny(object.FlagDead|object.FlagDestroyed) ||
				!s.CanSee(monster, target, 0) ||
				delta.X*delta.X+delta.Y*delta.Y > limit*limit ||
				moved.X*moved.X+moved.Y*moved.Y > 1000.0 ||
				checkInteraction && !s.CanInteract(monster, target, 0)
		}
		if remove {
			monsterLostSightNative528560(monster, i)
			changed = true
			continue
		}
		i++
	}
	if ud.CurrentEnemy != nil && ud.CurrentEnemy.HasEnchant(server.ENCHANT_CHARMING) {
		changed = true
	}
	if (ud.CurrentEnemy == nil || frame-ud.Field301 > 2*s.TickRate()) &&
		(ud.Field302 <= frame || frame == memmap.Uint32(0x5D4594, 2487684)) {
		s.Map.EachObjInCircle(monster.Pos(), sightRange, func(candidate *server.Object) bool {
			monsterUpdateSeenNative5286D0(candidate, monster)
			return true
		})
		ud.Field301 = frame
		ud.Field303 = frame
		changed = true
	}
	if changed {
		oldNetCode := uint32(0)
		if ud.CurrentEnemy != nil {
			oldNetCode = ud.CurrentEnemy.NetCode
		}
		monsterSelectEnemyNative528610(monster)
		if ud.CurrentEnemy != nil && oldNetCode != 0 && oldNetCode != ud.CurrentEnemy.NetCode {
			ud.Field300 = oldNetCode
		}
	}
	if ud.Field301 != frame {
		return
	}
	if ud.StatusFlags.Has(object.MonsterStatus(0x400)) || noxflags.HasGame(noxflags.GameOnline) || ud.CurrentEnemy != nil {
		ud.Field302 = frame + uint32(s.Rand.Logic.IntClamp(5, 10))
		return
	}
	delayBase := 5 * s.TickRate()
	distance := s.Sub5336D0(monster)
	ud.Field131 = math.Float32bits(float32(distance))
	switch {
	case distance < 0:
		ud.Field302 = frame + delayBase
	case distance > float64(sightRange):
		ud.Field302 = frame + uint32((distance-float64(sightRange))*float64(delayBase)/(1000.0-float64(sightRange))) + 10
	default:
		ud.Field302 = frame + uint32(s.Rand.Logic.IntClamp(5, 10))
	}
}

//export nox_server_unit_update_sight_native
func nox_server_unit_update_sight_native(monster *nox_object_t) {
	monsterUpdateSightNative5281F0(asObjectS(monster))
}

//export nox_server_ai_lost_sight_native
func nox_server_ai_lost_sight_native(monster *nox_object_t, index C.int) C.int {
	return C.int(bool2int(monsterLostSightNative528560(asObjectS(monster), int(index))))
}

//export nox_server_monster_select_enemy_native
func nox_server_monster_select_enemy_native(monster *nox_object_t) {
	monsterSelectEnemyNative528610(asObjectS(monster))
}

//export nox_server_monster_update_seen_native
func nox_server_monster_update_seen_native(candidate, monster *nox_object_t) {
	monsterUpdateSeenNative5286D0(asObjectS(candidate), asObjectS(monster))
}

//export nox_server_monster_see_enemy_native
func nox_server_monster_see_enemy_native(monster, target *nox_object_t) {
	monsterSeeEnemyNative5287B0(asObjectS(monster), asObjectS(target))
}

//export nox_server_monster_remove_seen_native
func nox_server_monster_remove_seen_native(monster, target *nox_object_t) C.int {
	obj := asObjectS(monster)
	targ := asObjectS(target)
	if obj == nil || targ == nil || !obj.Class().Has(object.ClassMonster) {
		return 0
	}
	index := obj.UpdateDataMonster().MonsterSeenEnemyIndex528950(targ)
	return C.int(bool2int(monsterLostSightNative528560(obj, index)))
}

//export nox_server_monster_has_seen_native
func nox_server_monster_has_seen_native(monster, target *nox_object_t) C.int {
	obj := asObjectS(monster)
	if obj == nil || target == nil || !obj.Class().Has(object.ClassMonster) {
		return 0
	}
	return C.int(bool2int(obj.UpdateDataMonster().MonsterHasSeenEnemy528950(asObjectS(target))))
}

//export nox_xxx_monsterClearActionStack_50A3A0
func nox_xxx_monsterClearActionStack_50A3A0(a1 *nox_object_t) {
	asObjectS(a1).ClearActionStack()
}

//export nox_xxx_checkMobAction_50A0D0
func nox_xxx_checkMobAction_50A0D0(a1 *nox_object_t, a2_cgo int32) int32 {
	a2 := int(a2_cgo)
	return int32(bool2int(asObjectS(a1).UpdateDataMonster().HasAction(ai.ActionType(a2))))
}

//export nox_xxx_generateRetreatPath_50CA00
func nox_xxx_generateRetreatPath_50CA00(ptr unsafe.Pointer, sz_cgo int32, obj *nox_object_t, p *C.float2) int32 {
	sz := int(sz_cgo)
	return int32(GetServer().Nox_xxx_generateRetreatPath_50CA00(unsafe.Slice((*types.Pointf)(ptr), sz), asObjectS(obj), (*types.Pointf)(unsafe.Pointer(p))))
}

//export nox_xxx_creatureSetDetailedPath_50D220
func nox_xxx_creatureSetDetailedPath_50D220(obj *nox_object_t, p *C.float2) {
	GetServer().Nox_xxx_creatureSetDetailedPath_50D220(asObjectS(obj), (*types.Pointf)(unsafe.Pointer(p)))
}

//export sub_50B810
func sub_50B810(obj *nox_object_t, p *C.float2) int32 {
	return int32(bool2int(GetServer().Sub_50B810(asObjectS(obj), (*types.Pointf)(unsafe.Pointer(p)))))
}

//export sub_50CB20
func sub_50CB20(obj *nox_object_t, p *C.float2) *nox_waypoint_t {
	return asWaypointC(GetServer().Sub_50CB20(asObjectS(obj), (*types.Pointf)(unsafe.Pointer(p))))
}

//export sub_50B500
func sub_50B500() {
	GetServer().S().AI.Paths.Sub_50B500()
}

//export sub_50B510
func sub_50B510() {
	GetServer().S().AI.Paths.Sub_50B510()
}

//export sub_50CB00
func sub_50CB00() int32 {
	points := GetServer().S().AI.Paths.Points()
	return int32(len(points))
}

//export sub_50CB10
func sub_50CB10() unsafe.Pointer {
	points := GetServer().S().AI.Paths.Points()
	if len(points) == 0 {
		return nil
	}
	return unsafe.Pointer(&points[0])
}

func Nox_xxx_mobSearchEdible_544A00(a1 *server.Object, a2 float32) int {
	return int(C.nox_xxx_mobSearchEdible_544A00(asObjectC(a1), C.float(a2)))
}
func Nox_xxx_weaponGetStaminaByType_4F7E80(a1 int) int {
	return int(server.WeaponStaminaByType4F7E80(uint32(a1)))
}
func Nox_xxx_unitIsDangerous_547120(a1 *server.Object, a2 *server.Object) {
	C.nox_xxx_unitIsDangerous_547120(asObjectC(a1), asObjectC(a2))
}
func Nox_xxx_checkIsKillable_528190(a1 *server.Object) int {
	return bool2int(monsterSightKillable528190(a1))
}
func Nox_xxx_polygonIsPlayerInPolygon_4217B0(a1 unsafe.Pointer, a2 int) *Nox_player_polygon_check_data {
	return (*Nox_player_polygon_check_data)(unsafe.Pointer(C.nox_xxx_polygonIsPlayerInPolygon_4217B0((*C.int2)(a1), C.int(a2))))
}

func Sub_421F10(a1 unsafe.Pointer, a2 int) *Nox_player_polygon_check_data {
	return (*Nox_player_polygon_check_data)(unsafe.Pointer(C.sub_421F10((*C.int)(a1), C.int(a2))))
}
func Nox_xxx_mobAction_50A910(a1 *server.Object) {
	GetServer().S().MonsterActionRefresh50A910(a1)
}
func Nox_xxx_monsterGetSoundSet_424300(a1 *server.Object) unsafe.Pointer {
	if a1 == nil || !a1.Class().Has(object.ClassMonster) {
		return nil
	}
	return a1.UpdateDataMonster().SoundSet122
}
func Nox_xxx_monsterPlayHurtSound_532800(a1 *server.Object) {
	GetServer().S().MonsterHurtSound532800(a1)
}
func Nox_xxx_mobAction_5469B0(a1 *server.Object) {
	GetServer().S().MonsterIdleSound5469B0(a1)
}
func Nox_xxx_unitUpdateSightMB_5281F0(a1 *server.Object) {
	C.nox_xxx_unitUpdateSightMB_5281F0(asObjectC(a1))
}
func Nox_xxx_monsterMainAIFn_547210(a1 *server.Object) {
	if GetServer().S().MonsterMainNativeRuntime547210(a1, server.MonsterMainRuntime547210{
		AudioEvent: func(id uint32, unit *server.Object) {
			C.nox_xxx_aud_501960(C.int(id), asObjectC(unit), 0, 0)
		},
		ScriptCallback: func(block *server.ScriptCallback, caller, trigger *server.Object, event server.ScriptEventType) {
			GetServer().NoxScriptC().ScriptCallback(block, caller, trigger, event)
		},
		GUICursorActive: func() bool {
			return C.nox_xxx_guiCursor_477600() != 0
		},
		FindObjectAtCursor: Nox_xxx_findObjectAtCursor_54AF40,
		TileAt:             Nox_xxx_tileNFromPoint_411160,
	}) {
		return
	}
	C.nox_xxx_monsterMainAIFn_547210(asObjectC(a1))
}
func Nox_xxx_updateNPCAnimData_50A850(a1 *server.Object) {
	if GetServer().S().MonsterUpdateNPCAnim50A850(a1) {
		return
	}
	C.nox_xxx_updateNPCAnimData_50A850(asObjectC(a1))
}
func Nox_xxx_monsterPolygonEnter_421FF0(a1 *server.Object) {
	GetServer().MonsterPolygonEnterNative421FF0(a1)
}
func Nox_xxx_monsterMimicCheckMorph_534950(a1 *server.Object) {
	C.nox_xxx_monsterMimicCheckMorph_534950(asObjectC(a1))
}
func Sub_5466F0(a1 *server.Object) int {
	s := GetServer()
	return s.S().MonsterInterestingSound5466F0(a1, Nox_xxx_tileNFromPoint_411160, s.Sub_50B810)
}
func Nox_xxx_mobHealSomeone_5411A0(a1 *server.Object) {
	GetServer().S().MonsterHealSomeone5411A0(a1)
}
func Nox_xxx_mobActionCast_5413B0(a1 *server.Object, a2 int) {
	C.nox_xxx_mobActionCast_5413B0(asObjectC(a1), C.int(a2))
}
