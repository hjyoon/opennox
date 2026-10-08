package opennox

import (
	"fmt"

	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

// Fixture awards a stock guide and casts through the normal spell selector.
// The network ray, random orb emissions, rendering and ownership completion
// are observed; no ray packet, client sprite, RNG result or pixel is supplied.
func (sc *e2eScenario) CheckCharmEffect(name string) {
	var host, target *server.Object
	var original types.Pointf
	var frame uint32
	var observed bool
	sc.addWhen(0, name+" initialize unowned creature", 1200, func() bool {
		return nox_client_isConnected() && noxServer.Players.HostUnit() != nil && noxClient.ClientPlayerUnit() != nil
	}, func() {
		host = noxServer.Players.HostUnit()
		original = host.PosVec
		origin, direction, err := e2eWarriorAbilityArena(original, host.Shape.Circle.R+4, func(from, to types.Pointf) bool { return e2eWarriorLaneClear(host, from, to) })
		if err != nil {
			e2eError(err)
			return
		}
		asObjectS(host).SetPos(origin)
		host.VelVec, host.ForceVec, host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
		target = noxServer.NewObjectByTypeID("Urchin")
		if target == nil {
			e2eError(fmt.Errorf("stock Urchin missing"))
			return
		}
		noxServer.CreateObjectAt(target, nil, origin.Add(direction.Mul(80)))
		noxServer.ObjectsAddPending()
		target.UpdateDataMonster().SetAggression(0)
		target.ClearActionStack()
		target.MonsterPushAction(ai.ACTION_WAIT, noxServer.Frame()+250000)
		guide := int32(spell.SPELL_SUMMON_URCHIN) - 74
		noxServer.AwardBeastGuide4FAE80(host, guide, 1)
		if target.ObjOwner != nil || host.UpdateDataPlayer().Player.BeastScrollLvl[guide] == 0 {
			e2eError(fmt.Errorf("Charm stock guide/unowned setup failed"))
		}
	})
	sc.Wait(12, name+" publish endpoints")
	sc.add(0, name+" normal spell cast", func() {
		frame = noxServer.Frame()
		if !noxServer.castSpellBy(spell.SPELL_CHARM, 1, host, target, target.PosVec) {
			e2eError(fmt.Errorf("Charm selector rejected cast"))
			return
		}
		e2eLog.Printf("CHARM CAST: caster=%p target=%p frame=%d owner=%p", host, target, frame, target.ObjOwner)
	})
	sc.addWhen(1, name+" natural ray and green orb", 180, func() bool {
		var ray bool
		for _, fx := range noxClient.fxDurationRays {
			if fx.kind == 2 && fx.drawable != nil && fx.drawable.TypeIDVal == uint32(noxClient.Things.IndByID("CharmRay")) &&
				fx.source == uint16(noxServer.GetUnitNetCode(host)) && fx.target == uint16(noxServer.GetUnitNetCode(target)) {
				ray = true
			}
		}
		if !ray {
			return false
		}
		orbType := uint32(noxClient.Things.IndByID("CharmOrb"))
		for dr := noxClient.Objs.FirstList2(); dr != nil; dr = dr.Next() {
			if dr.TypeIDVal == orbType {
				observed = true
				e2eLog.Printf("CHARM VISIBLE: ray-kind=2 CharmRay=true CharmOrb=%p position=%v elapsed=%d", dr, dr.PosVec, noxServer.Frame()-frame)
				return true
			}
		}
		for _, dr := range noxClient.Objs.AllList1() {
			if dr.TypeIDVal == orbType {
				observed = true
				e2eLog.Printf("CHARM VISIBLE: ray-kind=2 CharmRay=true CharmOrb=%p position=%v elapsed=%d", dr, dr.PosVec, noxServer.Frame()-frame)
				return true
			}
		}
		return false
	}, func() {})
	sc.CaptureMagicFrame(name + " actual green orbs")
	sc.addWhen(1, name+" natural completion and removal", 1200, func() bool {
		if !observed || target.ObjOwner != host || target.HasEnchant(server.EnchantID(28)) {
			return false
		}
		for _, fx := range noxClient.fxDurationRays {
			if fx.kind == 2 && fx.drawable != nil {
				return false
			}
		}
		return true
	}, func() {
		e2eLog.Printf("CHARM COMPLETE: target=%p owner=%p elapsed=%d visual=true ray-retired=true", target, target.ObjOwner, noxServer.Frame()-frame)
		noxServer.DelayedDelete(target)
		asObjectS(host).SetPos(original)
	})
}
