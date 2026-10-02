package opennox

import (
	"fmt"
	"image"
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"
	nsp "github.com/opennox/noxscript/ns/v4/spell"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

func e2ePushMode(level int, direction string) (fromNPC, ok bool) {
	switch direction {
	case "player-to-npc":
		return false, level >= 1 && level <= 5
	case "npc-to-player":
		return true, level == 0 // Select power through the real monster cast path.
	default:
		return false, false
	}
}

// Predict the radial attenuation and the stock ApplyForce conversions without
// calling either production service. The map radius uses binary32 squared
// distance; ApplyForce uses the separate double-precision vector length.
func e2ePushExpectedForce(origin, position types.Pointf, mass float32, coefficient float64, level int32) types.Pointf {
	delta := position.Sub(origin)
	distance := float32(math.Sqrt(float64(delta.X*delta.X+delta.Y*delta.Y)) + 0.1)
	if distance > 600 {
		return types.Pointf{}
	}
	force := float32(coefficient * float64(level))
	if distance > 10 {
		force *= 1 - (distance-10)/(600-10)
	}
	strength := 10 * force / mass
	length := float32(math.Hypot(float64(delta.X), float64(delta.Y)) + 0.1)
	return types.Ptf(float32(float64(delta.X)*float64(strength)/float64(length)),
		float32(float64(delta.Y)*float64(strength)/float64(length)))
}

type e2ePushFixture struct {
	level                  int
	fromNPC                bool
	direction              string
	host, npc              *server.Object
	caster, target         *server.Object
	original, targetOrigin types.Pointf
	clientOrigin           image.Point
	health                 uint16
	frame                  uint32
	active, castSeen       bool
	audioCount             int
}

func (f *e2ePushFixture) prepare() {
	f.host = noxServer.Players.HostUnit()
	if f.host == nil || f.host.HealthData == nil || f.host.HealthData.Cur == 0 ||
		f.host.Flags().HasAny(object.FlagDead|object.FlagDestroyed|object.FlagNoUpdate) {
		e2eError(fmt.Errorf("Push requires a live host"))
		return
	}
	f.original = f.host.PosVec
	origin, direction, err := e2eWarriorAbilityArena(f.original, f.host.Shape.Circle.R+16, func(from, to types.Pointf) bool {
		return e2eWarriorLaneClear(f.host, from, to)
	})
	if err != nil {
		e2eError(err)
		return
	}
	f.npc = noxServer.NewObjectByTypeID("NPC")
	if f.npc == nil {
		e2eError(fmt.Errorf("Push has no stock NPC type"))
		return
	}
	hostPos, npcPos := origin, origin.Add(direction.Mul(112))
	f.caster, f.target = f.host, f.npc
	if f.fromNPC {
		hostPos, npcPos = npcPos, hostPos
		f.caster, f.target = f.npc, f.host
	}
	asObjectS(f.host).SetPos(hostPos)
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	noxServer.CreateObjectAt(f.npc, nil, npcPos)
	noxServer.ObjectsAddPending()
	if !e2eObjectInWorld(f.npc) || f.npc.UpdateData == nil || f.npc.HealthData == nil ||
		f.npc.UpdateDataMonster().MonsterDef == nil || !f.npc.SubClass().AsMonster().Has(object.MonsterNPC) ||
		!f.target.IsMovable() || f.target.Mass <= 0 || noxServer.Balance.Float("PushPowerCoeff") <= 0 {
		e2eError(fmt.Errorf("Push stock NPC/target/balance was not initialized"))
		return
	}
	f.npc.UpdateDataMonster().SetAggression(0)
	f.npc.ClearActionStack()
	f.npc.MonsterPushAction(ai.ACTION_WAIT, noxServer.Frame()+250000)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, pointer := range []unsafe.Pointer{unsafe.Pointer(f.host), unsafe.Pointer(f.npc), f.npc.UpdateData} {
			if uintptr(pointer) <= math.MaxUint32 {
				e2eError(fmt.Errorf("Push native allocation is below 4 GiB: %p", pointer))
				return
			}
		}
	}
	// Observe actual server sound dispatch after the radial push, without
	// replacing the cast, force service, animation frame or packet consumer.
	noxServer.Audio.OnSound(func(id sound.ID, kind int, owner *server.Object, _ types.Pointf) {
		if !f.active || id != sound.SoundPushCast || owner != f.caster {
			return
		}
		f.audioCount++
		if kind != 0 || f.audioCount != 1 {
			e2eError(fmt.Errorf("Push cast audio kind/count=%d/%d", kind, f.audioCount))
			return
		}
		if f.fromNPC {
			ud := f.npc.UpdateDataMonster()
			head := ud.AIStackHead()
			if head.Type() != ai.ACTION_CAST_SPELL_ON_OBJECT || head.ArgObj(2) != f.target ||
				head.ArgU32(0) != uint32(spell.SPELL_PUSH) || ud.Field120_2 != 0 ||
				uint32(ud.Field120_1) != ud.MonsterDef.MissileAttackFrame216 {
				e2eError(fmt.Errorf("Push did not use the real NPC action/cast-frame gate"))
				return
			}
			f.verifyForce()
			e2eLog.Printf("PUSH NPC CAST FRAME: animation=%d duration=%d target=%p elapsed=%d", ud.Field120_1, ud.Field120_2, head.ArgObj(2), noxServer.Frame()-f.frame)
		}
	})
	e2eLog.Printf("PUSH PREPARED: direction=%s level=%d caster=%p target=%p update=%p mass=%g coefficient=%g", f.direction, f.level, f.caster, f.target, f.npc.UpdateData, f.target.Mass, noxServer.Balance.Float("PushPowerCoeff"))
}

func (f *e2ePushFixture) beginCast() {
	if !e2eObjectInWorld(f.target) || !e2eObjectInWorld(f.caster) || f.target.ForceVec != (types.Pointf{}) ||
		f.target.VelVec != (types.Pointf{}) || !noxServer.MapTraceRay(f.caster.PosVec, f.target.PosVec, 0) {
		e2eError(fmt.Errorf("Push placement did not settle without external force"))
		return
	}
	drawable := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.target)))
	if drawable == nil {
		e2eError(fmt.Errorf("Push target drawable has not been published"))
		return
	}
	f.targetOrigin, f.clientOrigin, f.health = f.target.PosVec, drawable.PosVec, f.target.HealthData.Cur
	f.frame, f.active = noxServer.Frame(), true
	if f.fromNPC {
		f.level = int(noxServer.Server.SpellPower4FE7B0(spell.SPELL_PUSH, f.caster))
		if f.level <= 0 {
			e2eError(fmt.Errorf("stock NPC has no Push spell power in this mode"))
			return
		}
		// Explicit spell/target selection is a fixture. Let ordinary monster
		// update, cast animation and AIActionCastOnObj execute it naturally.
		f.npc.MonsterPushAction(ai.ACTION_CAST_SPELL_ON_OBJECT, uint32(spell.SPELL_PUSH), 0, f.target)
		e2eLog.Printf("PUSH NPC QUEUED: level=%d cast-frame=%d", f.level, f.npc.UpdateDataMonster().MonsterDef.MissileAttackFrame216)
		return
	}
	api := noxServer.noxScriptP()
	api.CastSpellLvl(nsp.Spell("SPELL_PUSH"), f.level, api.toObj(f.caster), e2eFistPosition{f.targetOrigin})
	f.verifyForce()
}

func (f *e2ePushFixture) verifyForce() {
	want := e2ePushExpectedForce(f.caster.PosVec, f.target.PosVec, f.target.Mass, noxServer.Balance.Float("PushPowerCoeff"), int32(f.level))
	got := f.target.ForceVec
	if got != want || got == (types.Pointf{}) || f.target.HealthData.Cur != f.health || f.castSeen {
		e2eError(fmt.Errorf("Push force/health/count mismatch: direction=%s level=%d force=%v want=%v HP=%d/%d repeated=%t", f.direction, f.level, got, want, f.target.HealthData.Cur, f.health, f.castSeen))
		return
	}
	f.castSeen = true
	e2eLog.Printf("PUSH FORCE: direction=%s level=%d force=%v HP=%d unchanged elapsed=%d", f.direction, f.level, got, f.health, noxServer.Frame()-f.frame)
}

func (f *e2ePushFixture) moved() bool {
	if !e2eObjectInWorld(f.target) || !e2eObjectInWorld(f.caster) || f.target.HealthData.Cur != f.health {
		e2eError(fmt.Errorf("Push target/caster disappeared or took damage"))
		return true
	}
	if !f.castSeen || f.audioCount != 1 {
		return false
	}
	direction := f.targetOrigin.Sub(f.caster.PosVec).Normalize()
	delta := f.target.PosVec.Sub(f.targetOrigin)
	serverProgress := delta.X*direction.X + delta.Y*direction.Y
	drawable := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.target)))
	if drawable == nil {
		return false
	}
	clientDelta := drawable.PosVec.Sub(f.clientOrigin)
	clientProgress := float32(clientDelta.X)*direction.X + float32(clientDelta.Y)*direction.Y
	if serverProgress <= 1 || clientProgress <= 1 {
		return false
	}
	if f.fromNPC && f.npc.MonsterActionIsScheduled(ai.ACTION_CAST_SPELL_ON_OBJECT) {
		return false
	}
	e2eLog.Printf("PUSH MOVEMENT: direction=%s level=%d server=%g client=%g audio=1 HP=%d elapsed=%d", f.direction, f.level, serverProgress, clientProgress, f.health, noxServer.Frame()-f.frame)
	return true
}

// CheckPushSpell covers all script-cast player levels and the NPC's real
// AIActionCastOnObj / cast-frame / CastSpellByUser path. Positioning, waiting
// AI and explicit spell/target selection are fixtures. No force, animation,
// cast result, HP, movement packet or client position is injected. This is not
// an autonomous spell-choice, incantation/mana or campaign-trigger test.
func (sc *e2eScenario) CheckPushSpell(level int, direction, name string) {
	fromNPC, ok := e2ePushMode(level, direction)
	if !ok {
		e2eError(fmt.Errorf("invalid Push fixture: level=%d direction=%q", level, direction))
		return
	}
	f := &e2ePushFixture{level: level, direction: direction, fromNPC: fromNPC}
	sc.addWhen(0, name+" prepare", 1200, func() bool {
		return noxServer.Players.HostUnit() != nil && noxClient.ClientPlayerUnit() != nil && nox_client_isConnected()
	}, f.prepare)
	sc.Wait(12, name+" publish placement")
	sc.add(0, name+" cast", f.beginCast)
	sc.addWhen(1, name+" actual force and client movement", 300, f.moved, func() {
		f.active = false
		noxServer.DelayedDelete(f.npc)
		asObjectS(f.host).SetPos(f.original)
		f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	})
	sc.Wait(3, name+" cleanup")
}
