package opennox

import (
	"fmt"
	"image"

	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

// Exercise actual death dispatch, not a test drop predicate. NPC clothing
// must come from the stock map and is never re-equipped by this fixture.
// The player normally equips stock MedievalPants/Shirt. Ordinary armor and
// food are positive drop controls; no equipment/drop flags are written here.
func (sc *e2eScenario) CheckClothingDeath(player bool, id, name string) {
	var owner, host *server.Object
	var original types.Pointf
	var clothing, loot []*server.Object
	sc.addWhen(0, name+" prepare stock inventory", 1200, func() bool {
		return noxServer.Players.HostUnit() != nil && (player || noxServer.Objs.GetObjectByID(id) != nil)
	}, func() {
		host = noxServer.Players.HostUnit()
		original = host.PosVec
		owner = host
		if !player {
			owner = noxServer.Objs.GetObjectByID(id)
		}
		if owner.HealthData == nil || owner.HealthData.Cur == 0 || owner.Flags().HasAny(object.FlagDead|object.FlagDestroyed) ||
			!player && (!owner.Class().Has(object.ClassMonster) || uint32(owner.SubClass())&0x2000 != 0) {
			e2eError(fmt.Errorf("clothing death owner is unavailable: %v subclass=%#x flags=%#x health=%v", owner, uint32(owner.SubClass()), uint32(owner.Flags()), owner.HealthData))
			return
		}
		if player {
			for _, typeID := range []string{"MedievalPants", "MedievalShirt"} {
				item := legacy.Nox_xxx_playerRespawnItem_4EF750(owner, typeID, nil, 1, 0)
				if item == nil || !owner.HasItem(item) || !item.Flags().Has(object.FlagEquipped) && !asObjectS(owner).Equip(item) {
					e2eError(fmt.Errorf("normal player clothing equip failed: %s", typeID))
					return
				}
			}
		}
		for item := owner.InvFirstItem; item != nil; item = item.InvNextItem {
			if item.Flags().Has(object.FlagEquipped) && e2eBoundClothing(item) {
				if !item.Flags().Has(object.FlagNoAutoDrop) {
					e2eError(fmt.Errorf("death setup has unprotected clothing: %v flags=%#x", item, uint32(item.Flags())))
					return
				}
				clothing = append(clothing, item)
			}
		}
		if len(clothing) < 2 {
			e2eError(fmt.Errorf("death setup has fewer than two worn clothing items: %v", owner))
			return
		}
		for _, typeID := range []string{"LeatherArmor", "RedApple"} {
			item := legacy.Nox_xxx_playerRespawnItem_4EF750(owner, typeID, nil, 1, 0)
			if item == nil || !owner.HasItem(item) || item.Flags().Has(object.FlagNoAutoDrop) ||
				typeID == "LeatherArmor" && !item.Flags().Has(object.FlagEquipped) && !asObjectS(owner).Equip(item) {
				e2eError(fmt.Errorf("death positive drop control setup failed: %s item=%p", typeID, item))
				return
			}
			loot = append(loot, item)
		}
		if !player {
			asObjectS(host).SetPos(owner.PosVec.Sub(types.Ptf(48, 0)))
		}
		e2eQueueInput(&seat.MouseMoveEvent{Pos: image.Pt(10, 10)})
		e2eLog.Printf("CLOTHING DEATH PREPARED: player=%t owner=%p id=%s subclass=%#x clothing=%d regular-loot=%d map-clothing-not-repaired=%t", player, owner, owner.ID(), uint32(owner.SubClass()), len(clothing), len(loot), !player)
	})
	sc.Wait(60, name+" settle ordinary equipment and network")
	sc.add(0, name+" script lethal damage through real death dispatch", func() {
		if owner == nil || owner.HealthData == nil || owner.HealthData.Cur == 0 {
			e2eError(fmt.Errorf("clothing owner died before dispatch: %p", owner))
			return
		}
		before := owner.HealthData.Cur
		asObjectS(owner).DoDamage(nil, int(before), object.DamageTrue)
		if owner.HealthData.Cur != 0 || !owner.Flags().Has(object.FlagDead) {
			e2eError(fmt.Errorf("clothing real death dispatch failed: %v flags=%#x health=%v", owner, uint32(owner.Flags()), owner.HealthData))
			return
		}
		e2eLog.Printf("CLOTHING DEATH DISPATCHED: player=%t owner=%p hp=%d->%d flags=%#x", player, owner, before, owner.HealthData.Cur, uint32(owner.Flags()))
	})
	sc.Wait(60, name+" observe corpse and ground loot")
	sc.add(0, name+" verify worn clothing retained and ordinary loot dropped", func() {
		if owner.HealthData.Cur != 0 || !owner.Flags().Has(object.FlagDead) {
			e2eError(fmt.Errorf("clothing corpse is not dead: %v", owner))
			return
		}
		for _, item := range clothing {
			const wornBound = object.FlagEquipped | object.FlagNoAutoDrop
			if !owner.HasItem(item) || item.Flags()&wornBound != wornBound || item.Flags().Has(object.FlagDestroyed) {
				e2eError(fmt.Errorf("worn clothing dropped or changed after death: %v holder=%p flags=%#x", item, item.InvHolder, uint32(item.Flags())))
				return
			}
			e2eLog.Printf("DEATH CLOTHING RETAINED: player=%t owner=%p item=%s/%p flags=%#x", player, owner, item.ObjectTypeC().ID(), item, uint32(item.Flags()))
		}
		for _, item := range loot {
			inWorld := false
			for world := noxServer.Objs.First(); world != nil; world = world.Next() {
				if world == item {
					inWorld = true
					break
				}
			}
			if item.InvHolder != nil || owner.HasItem(item) || !inWorld || item.Flags().HasAny(object.FlagEquipped|object.FlagDestroyed) {
				e2eError(fmt.Errorf("ordinary death loot did not drop: %v holder=%p flags=%#x world=%t", item, item.InvHolder, uint32(item.Flags()), inWorld))
				return
			}
			e2eLog.Printf("DEATH REGULAR LOOT DROPPED: player=%t item=%s/%p pos=%v real-world=true", player, item.ObjectTypeC().ID(), item, item.PosVec)
		}
		e2eLog.Printf("CLOTHING DEATH VERIFIED: player=%t retained=%d dropped=%d actual-death-dispatch=true", player, len(clothing), len(loot))
		if !player {
			asObjectS(host).SetPos(original)
		}
	})
}
