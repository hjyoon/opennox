package opennox

import (
	"fmt"
	"image"
	"math"

	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

// The encounter uses a minion created by the real Quest stage transition.
// Only player placement is a fixture: never create the monster, inject its
// enemies/actions, change aggression/HP, call damage, or force a death/drop.
type e2eQuestCombatFixture struct {
	unit, target                 *server.Object
	typeID                       string
	playerHP, targetHP           uint16
	incoming, outgoing, dead     bool
	clientIncoming, clientDamage bool
	items                        []*server.Object
	ticks                        uint32
	startFrame                   uint32
}

// A last-damage source can already have been removed by a subsequent tick.
// Compare identities and membership before inspecting any native pointer.
func e2eQuestCombatAttributedTo(source, unit *server.Object) bool {
	if source == nil || unit == nil {
		return false
	}
	if source == unit {
		return true
	}
	for item := unit.FirstItem(); item != nil; item = item.NextItem() {
		if item == source {
			return true
		}
	}
	return e2eObjectInWorld(source) && source.ObjOwner == unit
}

func (f *e2eQuestCombatFixture) prepare() {
	state, err := e2eReadQuestPlayer()
	if err == nil {
		err = state.validate(5, "g_forest")
	}
	if err != nil || f.typeID != "Necromancer" && f.typeID != "Hecubah" {
		e2eError(fmt.Errorf("Quest combat setup: type=%q state=%+v error=%v", f.typeID, state, err))
		return
	}
	f.unit = noxServer.Players.HostUnit()
	var pos types.Pointf
	found := false
	for obj := noxServer.Objs.First(); obj != nil; obj = obj.Next() {
		if obj.ObjectTypeC().ID() != f.typeID || obj.UpdateData == nil || obj.HealthData == nil ||
			obj.HealthData.Cur == 0 || obj.Flags().HasAny(object.FlagDead|object.FlagDestroyed) {
			continue
		}
		for _, distance := range []float32{64, 96, 128} {
			for direction := 0; direction < 32; direction++ {
				angle := 2 * math.Pi * float64(direction) / 32
				candidate := types.Pointf{X: obj.PosVec.X + distance*float32(math.Cos(angle)), Y: obj.PosVec.Y + distance*float32(math.Sin(angle))}
				// Exclude the target's own collision shape at the ray endpoint.
				// All other live objects and map walls still obstruct the approach.
				if e2eWarriorLaneClear(obj, candidate, obj.PosVec) {
					pos, found, f.target = candidate, true, obj
					break
				}
			}
			if found {
				break
			}
		}
		if found {
			break
		}
		e2eLog.Printf("QUEST COMBAT APPROACH BLOCKED: type=%s object=%p pos=%v", f.typeID, obj, obj.PosVec)
	}
	if !found {
		e2eError(fmt.Errorf("Quest combat: no reachable player approach to any stock %s", f.typeID))
		return
	}
	for item := f.target.FirstItem(); item != nil; item = item.NextItem() {
		f.items = append(f.items, item)
	}
	asObjectS(f.unit).SetPos(pos)
	f.unit.NewPos, f.unit.PrevPos = pos, pos
	f.unit.VelVec, f.unit.ForceVec, f.unit.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	legacy.Nox_xxx_unitHasCollideOrUpdateFn_537610(f.unit)
	f.playerHP, f.targetHP, f.startFrame = f.unit.HealthData.Cur, f.target.HealthData.Cur, noxServer.Frame()
	e2eLog.Printf("QUEST COMBAT PREPARED: type=%s stage=%d frame=%d player=%p hp=%d pos=%v target=%p hp=%d pos=%v inventory=%d AI-unchanged=true",
		f.typeID, state.stage, f.startFrame, f.unit, f.playerHP, pos, f.target, f.targetHP, f.target.PosVec, len(f.items))
	f.logState()
}

func (f *e2eQuestCombatFixture) logState() {
	if !e2eObjectInWorld(f.target) {
		e2eLog.Printf("QUEST COMBAT STATE: type=%s frame=%d target-removed=true dead-seen=%t", f.typeID, noxServer.Frame(), f.dead)
		return
	}
	ud := f.target.UpdateDataMonster()
	action := "empty"
	if ud.AIStackInd >= 0 && int(ud.AIStackInd) < len(ud.AIStack) {
		action = ud.AIStack[ud.AIStackInd].Type().String()
	}
	e2eLog.Printf("QUEST COMBAT STATE: type=%s frame=%d player-hp=%d target-hp=%d flags=%#x aggression=%g sight=%g stack=%d action=%s enemy=%p preferred=%p seen=%d can-see=%t distance=%.3f damage-source=%p damage-type=%d damage-frame=%d",
		f.typeID, noxServer.Frame(), f.unit.HealthData.Cur, f.target.HealthData.Cur, uint32(f.target.Flags()), ud.Aggression, ud.SightRange,
		ud.AIStackInd, action, ud.CurrentEnemy, ud.PreferredEnemy, ud.Field282_1, f.target.CanSee(f.unit),
		math.Hypot(float64(f.unit.PosVec.X-f.target.PosVec.X), float64(f.unit.PosVec.Y-f.target.PosVec.Y)), f.target.Obj130, f.target.Field131, f.target.Frame134)
}

func (f *e2eQuestCombatFixture) observe() {
	f.ticks++
	if !e2eObjectInWorld(f.unit) || f.unit.HealthData.Cur == 0 {
		e2eError(fmt.Errorf("Quest combat: player died before minion verification"))
		return
	}
	if f.unit.HealthData.Cur < f.playerHP {
		source := f.unit.Obj130
		if e2eObjectInWorld(f.target) && e2eQuestCombatAttributedTo(source, f.target) {
			f.incoming = true
		}
	}
	if drawable := noxClient.ClientPlayerUnit(); drawable != nil {
		if delta, ok := legacy.HealthChangeForDrawable(drawable.NetCode32); ok && delta < 0 {
			f.clientIncoming = true
		}
	}
	if e2eObjectInWorld(f.target) {
		if f.target.HealthData.Cur < f.targetHP && f.target.Frame134 >= f.startFrame {
			source := f.target.Obj130
			if e2eQuestCombatAttributedTo(source, f.unit) {
				f.outgoing = true
			}
		}
		f.dead = f.dead || f.target.HealthData.Cur == 0 && f.target.Flags().Has(object.FlagDead)
		if drawable := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.target))); drawable != nil {
			if delta, ok := legacy.HealthChangeForDrawable(drawable.NetCode32); ok && delta < 0 {
				f.clientDamage = true
			}
		}
	}
	if f.ticks%120 == 0 {
		f.logState()
	}
}

func (f *e2eQuestCombatFixture) acquired() bool {
	f.observe()
	if !e2eObjectInWorld(f.target) {
		return false
	}
	ud := f.target.UpdateDataMonster()
	if ud.CurrentEnemy == f.unit || ud.PreferredEnemy == f.unit {
		return true
	}
	for i := 0; i < int(ud.Field282_1) && i < len(ud.SeenEnemies); i++ {
		if ud.SeenEnemies[i] == f.unit {
			return true
		}
	}
	return false
}

func (f *e2eQuestCombatFixture) fight() bool {
	f.observe()
	if f.dead || !e2eObjectInWorld(f.target) {
		return true
	}
	if f.ticks%3 == 0 {
		mouse := noxClient.Viewport().ToScreenPos(image.Pt(int(f.target.PosVec.X), int(f.target.PosVec.Y)))
		noxClient.ChangeMousePos(mouse, true)
		e2eQueueInput(&seat.MouseMoveEvent{Pos: mouse, Relative: false})
		distance := math.Hypot(float64(f.unit.PosVec.X-f.target.PosVec.X), float64(f.unit.PosVec.Y-f.target.PosVec.Y))
		e2eQueueInput(&seat.MouseButtonEvent{Button: seat.MouseButtonRight, Pressed: distance > 32})
	}
	if f.ticks%30 == 0 {
		e2eQueueInput(&seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true})
	} else if f.ticks%30 == 1 {
		e2eQueueInput(&seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false})
	}
	return false
}

func (sc *e2eScenario) CheckQuestMinionCombat(typeID, name string) {
	f := &e2eQuestCombatFixture{typeID: typeID}
	sc.add(0, name+" place only player beside stock minion", f.prepare)
	sc.addWhen(0, name+" wait for natural AI acquisition", 600, f.acquired, func() {
		e2eLog.Printf("QUEST COMBAT ACQUIRED: type=%s frame=%d elapsed=%d", f.typeID, noxServer.Frame(), noxServer.Frame()-f.startFrame)
		f.logState()
	})
	sc.addWhen(0, name+" wait for minion attack and client damage", 900, func() bool {
		f.observe()
		return f.incoming && f.clientIncoming
	}, func() {
		e2eLog.Printf("QUEST COMBAT INCOMING: type=%s frame=%d player-hp=%d->%d damage-source=%p damage-type=%d client-damage=true", f.typeID, noxServer.Frame(), f.playerHP, f.unit.HealthData.Cur, f.unit.Obj130, f.unit.Field131)
	})
	sc.Screen(name + " enemy attack")
	sc.addWhen(0, name+" defeat minion with real mouse input", 2400, f.fight, func() {
		e2eQueueInput(&seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false}, &seat.MouseButtonEvent{Button: seat.MouseButtonRight, Pressed: false})
		if !f.outgoing || !f.clientDamage || !f.dead {
			e2eError(fmt.Errorf("Quest combat death unproven: type=%s player-damage=%t client-damage=%t dead=%t", f.typeID, f.outgoing, f.clientDamage, f.dead))
			return
		}
		e2eLog.Printf("QUEST COMBAT DEAD: type=%s frame=%d initial-hp=%d hp=0 damage-by-player=true client-damage=true player-hp=%d", f.typeID, noxServer.Frame(), f.targetHP, f.unit.HealthData.Cur)
	})
	sc.Wait(90, name+" let stock death and reward inventory update")
	sc.add(0, name+" observe original reward items", func() {
		for _, item := range f.items {
			inInventory := false
			for owned := f.unit.FirstItem(); owned != nil; owned = owned.NextItem() {
				if owned == item {
					inInventory = true
					break
				}
			}
			if !e2eObjectInWorld(item) && !inInventory {
				e2eError(fmt.Errorf("Quest minion reward disappeared: type=%s item=%p", f.typeID, item))
				return
			}
			if item.InvHolder == f.target || item.Flags().HasAny(object.FlagDead|object.FlagDestroyed) {
				e2eError(fmt.Errorf("Quest minion reward not released by death: type=%s item=%s holder=%p flags=%#x", f.typeID, item.ObjectTypeC().ID(), item.InvHolder, uint32(item.Flags())))
				return
			}
			e2eLog.Printf("QUEST COMBAT REWARD: minion=%s item=%s object=%p holder=%p pos=%v", f.typeID, item.ObjectTypeC().ID(), item, item.InvHolder, item.PosVec)
		}
	})
	sc.Screen(name + " defeated and rewards")
}
