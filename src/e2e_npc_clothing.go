package opennox

import (
	"fmt"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

func e2eBoundClothing(item *server.Object) bool {
	return item != nil && item.Class().Has(object.ClassArmor) &&
		noxServer.Armor.Nox_xxx_unitArmorInventoryEquipFlags_415C70(item)&0xC0D != 0
}

// AuditNPCClothing reads stock map-loaded inventory. It neither grants items
// nor repairs their flags, and also records the actual NPC greeting sound.
func (sc *e2eScenario) AuditNPCClothing(name string) {
	sc.addWhen(0, name, 1200, func() bool {
		return noxServer.Players.HostUnit() != nil && noxClient.ClientPlayerUnit() != nil
	}, func() {
		players, npcs, protected := 0, 0, 0
		for owner := noxServer.Objs.First(); owner != nil; owner = owner.Next() {
			if !owner.Class().HasAny(object.ClassMonster|object.ClassPlayer) || owner.UpdateData == nil {
				continue
			}
			bound := 0
			for item := owner.InvFirstItem; item != nil; item = item.InvNextItem {
				if !item.Flags().Has(object.FlagEquipped) || !e2eBoundClothing(item) {
					continue
				}
				if !item.Flags().Has(object.FlagNoAutoDrop) {
					e2eError(fmt.Errorf("map-loaded %s %p is wearing unprotected %s %p flags=%#x", owner.ObjectTypeC().ID(), owner, item.ObjectTypeC().ID(), item, uint32(item.Flags())))
					return
				}
				bound++
				protected++
				e2eLog.Printf("BOUND CLOTHING: owner=%s/%p item=%s/%p flags=%#x", owner.ObjectTypeC().ID(), owner, item.ObjectTypeC().ID(), item, uint32(item.Flags()))
			}
			if owner.Class().Has(object.ClassPlayer) {
				players++
			} else if uint32(owner.SubClass())&0x10 != 0 {
				npcs++
				var greeting sound.ID
				if ptr := owner.UpdateDataMonster().SoundSet122; ptr != nil {
					greeting = sound.ID(*(*uint32)(unsafe.Add(ptr, 4)))
				}
				wire := noxServer.GetUnitNetCode(owner)
				dr := noxClient.Objs.ByNetCode(uint16(wire))
				visible := dr != nil && noxClient.Viewport().ToScreenPos(dr.Pos()).In(noxClient.Viewport().Screen)
				e2eLog.Printf("NPC STOCK STATE: type=%s object=%p script=%s subclass=%#x flags=%#x health=%v clothing=%d greeting=%s/%d field5=%#x visible=%t stack=%v", owner.ObjectTypeC().ID(), owner, owner.ID(), uint32(owner.SubClass()), uint32(owner.Flags()), owner.HealthData, bound, greeting, greeting, owner.Field5, visible, owner.UpdateDataMonster().GetAIStack())
			}
		}
		if players == 0 || npcs == 0 || protected < 3 {
			e2eError(fmt.Errorf("map clothing audit found too little real equipment: players=%d NPCs=%d protected=%d", players, npcs, protected))
			return
		}
		e2eLog.Printf("MAP CLOTHING VERIFIED: players=%d NPCs=%d protected=%d", players, npcs, protected)
	})
}
