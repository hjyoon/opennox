package opennox

import (
	"fmt"

	"github.com/opennox/libs/client/seat"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/legacy"
)

// Resolve the actual remote drawable and use normal mouse input. The owned
// server prepares its merchant separately; this client never creates objects,
// funds the player, starts a trade session or injects a trade/audio packet.
func (sc *e2eScenario) OpenPeerShop(typeID, name string) {
	var merchant *client.Drawable
	sc.addWhen(0, name, 1200, func() bool {
		if noxClient == nil || !nox_client_isConnected() || noxClient.ClientPlayerUnit() == nil {
			return false
		}
		typ := noxClient.Things.TypeByID(typeID)
		if typ == nil {
			return false
		}
		player := noxClient.ClientPlayerUnit().Pos()
		merchant = nil
		best := int64(1 << 62)
		for dr := noxClient.Objs.FirstList1(); dr != nil; dr = dr.NextPtr {
			if dr.TypeIDVal != uint32(typ.Index()) || !noxClient.Viewport().ToScreenPos(dr.Pos()).In(noxClient.Viewport().Screen) {
				continue
			}
			d := dr.Pos().Sub(player)
			dist := int64(d.X)*int64(d.X) + int64(d.Y)*int64(d.Y)
			if dist < best {
				merchant, best = dr, dist
			}
		}
		return merchant != nil
	}, func() {
		if noxServer.Players.HostUnit() != nil {
			e2eError(fmt.Errorf("peer shop requires an independent client without a local host unit"))
			return
		}
		pos := noxClient.Viewport().ToScreenPos(merchant.Pos())
		e2eLog.Printf("PEER SHOP MOUSE: type=%s wire=%#x world=%v screen=%v player=%d local-host-unit=nil", typeID, merchant.NetCode32, merchant.Pos(), pos, legacy.ClientPlayerNetCode())
		e2eQueueInput(&seat.MouseMoveEvent{Pos: pos, Relative: false})
	})
	sc.addWhen(1, name+" cursor", 600, func() bool {
		return noxClient.Nox_client_getCursorType() == gui.CursorShop
	}, func() {
		e2eLog.Printf("PEER SHOP CURSOR: cursor=%d", noxClient.Nox_client_getCursorType())
		e2eQueueInput(&seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true})
	})
	sc.Input(1, "", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false})
}

func (sc *e2eScenario) AssertPeerPlayerGold(gold int, name string) {
	sc.addWhen(0, name, 600, func() bool {
		return nox_client_isConnected() && legacy.Nox_client_gold_4674A0() == uint32(gold)
	}, func() {
		if noxServer.Players.HostUnit() != nil {
			e2eError(fmt.Errorf("peer gold assertion requires an independent client"))
			return
		}
		e2eLog.Printf("PEER PLAYER GOLD: client=%d local-host-unit=nil", legacy.Nox_client_gold_4674A0())
	})
}
