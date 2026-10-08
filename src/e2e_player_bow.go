package opennox

import (
	"fmt"
	"image"

	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

// Real inventory input equips the stock bow and quiver before this observer.
// Each shot must enter the input/network/player-attack path, consume exactly
// one arrow, hurt a durable waiting target, and finish its attack animation.
func (sc *e2eScenario) CheckPlayerBow(item, name string) {
	flag, projectileName := object.WeaponBow, "ArcherArrow"
	if item == "CrossBow" {
		flag, projectileName = object.WeaponCrossbow, "ArcherBolt"
	} else if item != "Bow" {
		e2eError(fmt.Errorf("unknown bow %q", item))
		return
	}
	var unit, weapon, quiver, target, projectile *server.Object
	var original types.Pointf
	var charge uint8
	var hp uint16
	var start uint32
	sc.addWhen(0, name+" stock equipment", 1200, func() bool {
		return noxServer.Players.HostUnit() != nil && noxClient.ClientPlayerUnit() != nil && nox_client_isConnected()
	}, func() {
		unit = noxServer.Players.HostUnit()
		weapon = unit.UpdateDataPlayer().EquippedWeapon
		if weapon == nil || weapon.ObjectTypeC().ID() != item || !weapon.Flags().Has(object.FlagEquipped) || unit.ControllingPlayer().WeaponEquip&uint32(flag) == 0 {
			e2eError(fmt.Errorf("%s not equipped through inventory", item))
			return
		}
		for obj := unit.InvFirstItem; obj != nil; obj = obj.InvNextItem {
			if obj.WeaponClass() == object.WeaponQuiver && obj.Flags().Has(object.FlagEquipped) {
				quiver = obj
				break
			}
		}
		if quiver == nil || quiver.UseData.Ptr == nil || (*server.AmmoUseData)(quiver.UseData.Ptr).Charge1 < 3 {
			e2eError(fmt.Errorf("%s has no equipped stock quiver", item))
			return
		}
		original = unit.PosVec
		origin, direction, err := e2eWarriorAbilityArena(original, unit.Shape.Circle.R+4, func(from, to types.Pointf) bool { return e2eWarriorLaneClear(unit, from, to) })
		if err != nil {
			e2eError(err)
			return
		}
		asObjectS(unit).SetPos(origin)
		unit.VelVec, unit.ForceVec, unit.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
		target = noxServer.NewObjectByTypeID("Troll")
		if target == nil {
			e2eError(fmt.Errorf("missing stock Troll"))
			return
		}
		noxServer.CreateObjectAt(target, nil, origin.Add(direction.Mul(112)))
		noxServer.ObjectsAddPending()
		asObjectS(target).SetMaxHealth(2000)
		target.UpdateDataMonster().SetAggression(0)
		target.ClearActionStack()
		target.MonsterPushAction(ai.ACTION_WAIT, noxServer.Frame()+250000)
		e2eLog.Printf("PLAYER BOW PREPARED: item=%s player=%p weapon=%p quiver=%p charges=%d target=%p", item, unit, weapon, quiver, (*server.AmmoUseData)(quiver.UseData.Ptr).Charge1, target)
	})
	sc.Wait(12, name+" publish placement")
	for shot := 1; shot <= 3; shot++ {
		label := fmt.Sprintf("%s shot %d", name, shot)
		sc.add(0, label+" actual attack input", func() {
			charge, hp, start = (*server.AmmoUseData)(quiver.UseData.Ptr).Charge1, target.HealthData.Cur, noxServer.Frame()
			projectile = nil
			mouse := noxClient.Viewport().ToScreenPos(image.Pt(int(target.PosVec.X), int(target.PosVec.Y)))
			e2eQueueInput(&seat.MouseMoveEvent{Pos: mouse}, &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true})
		})
		sc.Input(1, label+" release", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false})
		sc.addWhen(1, label+" natural launch and charge", 120, func() bool {
			for _, obj := range noxServer.Objs.AllMissiles() {
				if obj.ObjOwner == unit && obj.ObjectTypeC().ID() == projectileName {
					projectile = obj
					if (*server.AmmoUseData)(quiver.UseData.Ptr).Charge1 != charge-1 {
						e2eError(fmt.Errorf("bow charge not consumed exactly once"))
					}
					e2eLog.Printf("PLAYER BOW LAUNCH: item=%s shot=%d owner=%p projectile=%p charges=%d->%d velocity=%v", item, shot, unit, projectile, charge, charge-1, obj.VelVec)
					return true
				}
			}
			return false
		}, func() {})
		sc.addWhen(0, label+" natural hit", 120, func() bool { return target.HealthData.Cur < hp }, func() {
			e2eLog.Printf("PLAYER BOW HIT: item=%s shot=%d HP=%d->%d source=%p projectile=%p elapsed=%d", item, shot, hp, target.HealthData.Cur, target.Obj130, projectile, noxServer.Frame()-start)
		})
		sc.addWhen(0, label+" animation completion", 120, func() bool { return unit.UpdateDataPlayer().State != server.PlayerState1 }, func() {
			if unit.UpdateDataPlayer().EquippedWeapon != weapon || weapon.InvHolder != unit || quiver.InvHolder != unit || (*server.AmmoUseData)(quiver.UseData.Ptr).Charge1 != charge-1 {
				e2eError(fmt.Errorf("bow completion changed equipment/charges"))
			}
			e2eLog.Printf("PLAYER BOW COMPLETE: item=%s shot=%d state=%d frame=%d deadline=%d", item, shot, unit.UpdateDataPlayer().State, noxServer.Frame(), unit.UpdateDataPlayer().Field0)
		})
		sc.Wait(8, label+" settle")
	}
	sc.Screen(name + " three real hits")
	sc.add(0, name+" cleanup", func() { noxServer.DelayedDelete(target); asObjectS(unit).SetPos(original) })
}
