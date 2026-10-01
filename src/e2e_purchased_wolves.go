package opennox

import (
	"fmt"
	"image"

	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/object"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

var e2ePurchasedWolves struct {
	wolves [2]*server.Object
	ids    [2]int32
	gold   int
}

// War03b's stock Henrick dialog charges 200 gold per wolf and calls
// ReleaseCharmedWolf -> BecomePet. The fixture awards only purchase funds;
// it never spawns/reowns a wolf or writes migration/AI/health state.
func (sc *e2eScenario) PrepareWar03bWolfPurchase(name string) {
	sc.add(0, name, func() {
		host := noxServer.Players.HostUnit()
		if e2eMapBaseName(legacy.Nox_xxx_mapGetMapName_409B40()) != "war03b" ||
			!noxflags.HasGame(noxflags.GameModeCoop) || host == nil {
			e2eError(fmt.Errorf("purchased wolves require a live War03b campaign player"))
			return
		}
		e2eLog.Printf("WOLF PURCHASE PLAYER: object=%p flags=%#x state=%d dialog=%p save-eligible=%d frame=%d", host, uint32(host.Flags()), host.UpdateDataPlayer().State, host.UpdateDataPlayer().DialogWith, sub_4DCC10(host), noxServer.Frame())
		for i, id := range []string{"Wolf1", "Wolf2"} {
			wolf := noxServer.noxScriptP().Object(id)
			if wolf == nil {
				e2eError(fmt.Errorf("War03b stock %s is missing", id))
				return
			}
			unit := noxServer.noxScriptP().asObj(wolf)
			if unit == nil || !e2eObjectInWorld(unit) || unit.ObjectTypeC().ID() != "Wolf" ||
				unit.HealthData == nil || unit.HealthData.Cur == 0 || unit.UpdateData == nil ||
				unit.Flags().HasAny(object.FlagDead|object.FlagDestroyed) ||
				unit.SubClass().AsMonster().Has(object.MonsterMonitor) ||
				unit.SubClass().AsMonster().Has(object.MonsterMigrate) {
				e2eError(fmt.Errorf("War03b %s is not an unpurchased live stock wolf: %p", id, unit))
				return
			}
			e2ePurchasedWolves.wolves[i], e2ePurchasedWolves.ids[i] = unit, unit.ScriptIDVal
			e2eLog.Printf("WOLF PURCHASE STOCK: name=%s object=%p script-id=%d owner=%p HP=%d/%d subclass=%#x monitor=false migrate=false", id, unit, unit.ScriptIDVal, unit.ObjOwner, unit.HealthData.Cur, unit.HealthData.Max, uint32(unit.ObjSubClass))
		}
		if e2ePurchasedWolves.wolves[0] == e2ePurchasedWolves.wolves[1] {
			e2eError(fmt.Errorf("War03b stock wolf identities are duplicated"))
			return
		}
		asObjectS(host).ChangeGold(400)
		e2ePurchasedWolves.gold = asObjectS(host).GetGold()
		e2eLog.Printf("WOLF PURCHASE FUNDS: gold=%d fixture-award=400 stock-price=200 each", e2ePurchasedWolves.gold)
	})
}

func (sc *e2eScenario) OpenHenrickDialog(name string) {
	sc.add(0, name, func() {
		npc := noxServer.noxScriptP().Object("Henrick")
		if npc == nil || e2eMapBaseName(legacy.Nox_xxx_mapGetMapName_409B40()) != "war03b" {
			e2eError(fmt.Errorf("War03b Henrick is missing"))
			return
		}
		noxServer.noxScriptP().StartDialog(npc, noxServer.noxScriptP().GetHost())
		e2eLog.Printf("WOLF PURCHASE DIALOG OPEN: via stock StartDialog service frame=%d", noxServer.Frame())
	})
	sc.addWhen(0, name+" wait for stock Yes button", 300, func() bool {
		dialog := legacy.Get_dword_5d4594_1123524()
		if dialog == nil || dialog.GetFlags().IsHidden() {
			return false
		}
		yes := dialog.ChildByID(3908)
		return yes != nil && !yes.GetFlags().IsHidden() && yes.GetFlags().IsEnabled()
	}, nil)
}

func (sc *e2eScenario) ClickNPCDialogYes(name string) {
	sc.add(0, name, func() {
		dialog := legacy.Get_dword_5d4594_1123524()
		if dialog == nil || dialog.GetFlags().IsHidden() {
			e2eError(fmt.Errorf("NPC Yes/No dialog is not active"))
			return
		}
		yes := dialog.ChildByID(3908)
		if yes == nil || yes.GetFlags().IsHidden() || !yes.GetFlags().IsEnabled() {
			e2eError(fmt.Errorf("NPC Yes control is unavailable"))
			return
		}
		pos := yes.GlobalPos().Add(image.Pt(yes.Size().X/2, yes.Size().Y/2))
		e2eQueueInput(&seat.MouseMoveEvent{Pos: pos, Relative: false})
		e2eLog.Printf("WOLF PURCHASE YES INPUT: point=%v frame=%d", pos, noxServer.Frame())
	})
	sc.Input(1, "", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true})
	sc.Input(1, "", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false})
}

func (sc *e2eScenario) AssertWar03bWolfPurchase(count int, name string) {
	sc.addWhen(0, name+" synchronize purchase", 300, func() bool {
		host := noxServer.Players.HostUnit()
		return host != nil && asObjectS(host).GetGold() == e2ePurchasedWolves.gold-200*count
	}, func() {
		host := noxServer.Players.HostUnit()
		if count < 1 || count > 2 {
			e2eError(fmt.Errorf("wolf purchase count=%d", count))
			return
		}
		for i, wolf := range e2ePurchasedWolves.wolves {
			if !e2eObjectInWorld(wolf) {
				e2eError(fmt.Errorf("stock wolf %d disappeared at purchase", i+1))
				return
			}
			// MakeFriendly (00516720) marks Migrate; BecomePet (004E7B00)
			// independently marks Monitor. The real purchase script calls both.
			monitor := wolf.SubClass().AsMonster().Has(object.MonsterMonitor)
			migrate := wolf.SubClass().AsMonster().Has(object.MonsterMigrate)
			if monitor != (i < count) || migrate != (i < count) || wolf.ObjOwner != host {
				e2eError(fmt.Errorf("stock wolf %d purchase: monitor=%t migrate=%t owner=%p/%p count=%d", i+1, monitor, migrate, wolf.ObjOwner, host, count))
				return
			}
			e2eLog.Printf("WOLF PURCHASE PET: index=%d subclass=%#x monitor=%t migrate=%t", i+1, uint32(wolf.ObjSubClass), wolf.SubClass().AsMonster().Has(object.MonsterMonitor), wolf.SubClass().AsMonster().Has(object.MonsterMigrate))
		}
		e2eLog.Printf("WOLF PURCHASE VERIFIED: count=%d gold=%d->%d stock-objects=true", count, e2ePurchasedWolves.gold, asObjectS(host).GetGold())
	})
	if count == 2 {
		sc.addWhen(0, name+" capture native client identities", 300, func() bool {
			for _, wolf := range e2ePurchasedWolves.wolves {
				if !e2eObjectInWorld(wolf) || noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(wolf))) == nil {
					return false
				}
			}
			return true
		}, func() {
			e2eTransitionPets.pets, e2eTransitionPets.pixies = nil, nil
			for _, wolf := range e2ePurchasedWolves.wolves {
				e2eTransitionPets.pets = append(e2eTransitionPets.pets, e2eTransitionPet{object: wolf, typeID: "Wolf", health: wolf.HealthData.Cur, wire: uint16(noxServer.GetUnitNetCode(wolf))})
			}
		})
	}
}

func (sc *e2eScenario) ContactPurchasedWolfExit(mapID, name string) {
	sc.enterPetTransitionExit(mapID, name, true, 2)
}

func (sc *e2eScenario) AssertPurchasedWolfTransition(name string) {
	sc.assertTransitionSummons(name, 2)
	sc.add(0, name+" check purchased identities and no duplicate wolves", func() {
		host := noxServer.Players.HostUnit()
		monitored := 0
		for obj := noxServer.Objs.First(); obj != nil; obj = obj.Next() {
			if obj.ObjectTypeC().ID() == "Wolf" && obj.ObjOwner == host && obj.SubClass().AsMonster().Has(object.MonsterMonitor) &&
				!obj.Flags().HasAny(object.FlagDead|object.FlagDestroyed) {
				monitored++
			}
		}
		if monitored != 2 {
			e2eError(fmt.Errorf("purchased wolf transition: live owned monitored wolves=%d want=2", monitored))
			return
		}
		for i, wolf := range e2ePurchasedWolves.wolves {
			if !e2eObjectInWorld(wolf) || wolf.ScriptIDVal != e2ePurchasedWolves.ids[i] ||
				uint16(noxServer.GetUnitNetCode(wolf)) != e2eTransitionPets.pets[i].wire {
				e2eError(fmt.Errorf("purchased wolf identity %d was replaced after transition", i+1))
				return
			}
		}
		e2eLog.Printf("PURCHASED WOLVES TRANSITION VERIFIED: map=%s same-two-objects=true no-duplicates=true gold=%d", legacy.Nox_xxx_mapGetMapName_409B40(), asObjectS(host).GetGold())
	})
}
