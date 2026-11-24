package opennox

import (
	"fmt"
	"image"
	"math"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func (sc *e2eScenario) CheckP2PTradePreview(name string) {
	var potion, second, sword server.Object
	deliver := func(packet []byte) {
		if got := legacy.Nox_xxx_netOnPacketRecvCli_48EA70_switch(server.HostPlayerIndex, netmsg.MSG_TRADE, packet); got != len(packet) {
			e2eError(fmt.Errorf("trade packet %d consumed %d bytes, want %d", packet[1], got, len(packet)))
		}
	}
	sc.addWhen(0, name+" open trade and post offers", 1200, func() bool {
		return nox_client_isConnected() && noxServer.Players.HostUnit() != nil
	}, func() {
		host := noxServer.Players.HostUnit()
		potionType, swordType := noxServer.Types.ByID("RedPotion"), noxServer.Types.ByID("GreatSword")
		if potionType == nil || swordType == nil {
			e2eError(fmt.Errorf("trade fixture needs stock potion and sword types"))
			return
		}
		potion.TypeInd, potion.NetCode = uint16(potionType.Ind()), 0x7600
		second.TypeInd, second.NetCode = potion.TypeInd, 0x7601
		sword.TypeInd, sword.NetCode = uint16(swordType.Ind()), 0x7602

		start := server.BuildTradeP2PStartPacket50F1A0("Trade Partner")
		deliver(start[:])
		firstOffer := server.BuildTradeP2PItemPacket50FAE0(host, host, &potion, 40)
		deliver(firstOffer[:])
		secondOffer := server.BuildTradeP2PItemPacket50FAE0(host, host, &second, 41)
		deliver(secondOffer[:])
		remoteOffer := server.BuildTradeP2PItemPacket50FAE0(host, nil, &sword, 200)
		deliver(remoteOffer[:])
	})
	sc.Wait(8, name+" draw native offers")
	sc.add(0, name+" verify drawables and trade quantities", func() {
		win, active := legacy.ClientP2PTradeWindow()
		if !active || win == nil {
			e2eError(fmt.Errorf("trade window did not open"))
			return
		}
		check := func(dr *client.Drawable, kind uint16, count, cost, wantCount, wantCost uint32, control uint) {
			cell := win.ChildByID(control)
			if dr == nil || cell == nil {
				e2eError(fmt.Errorf("trade preview missing drawable or control %d", control))
				return
			}
			wantPos := cell.GlobalPos().Add(image.Pt(25, 25))
			if dr.Pos() != wantPos || dr.Field27() != uint32(kind) || count != wantCount || cost != wantCost {
				e2eError(fmt.Errorf("trade offer %d at %v type=%d count=%d cost=%d, want %v type=%d count=%d cost=%d",
					control, dr.Pos(), dr.Field27(), count, cost, wantPos, kind, wantCount, wantCost))
			}
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= math.MaxUint32 {
				e2eError(fmt.Errorf("trade preview %d did not exercise a native-width address: %p", control, dr))
			}
		}
		local, count, cost := legacy.ClientP2PTradeOffer(false, 0)
		check(local, potion.TypeInd, count, cost, 2, 81, 3704)
		remote, count, cost := legacy.ClientP2PTradeOffer(true, 0)
		check(remote, sword.TypeInd, count, cost, 1, 200, 3705)
		e2eLog.Printf("P2P TRADE PREVIEWS: local=%p remote=%p", local, remote)
	})
	sc.CaptureMagicFrame(name + " visible trade window")
	sc.add(0, name+" remove one stacked item and close", func() {
		remove := server.BuildTradeP2PRemovePacket50FF90(&potion)
		deliver(remove[:])
		if dr, count, cost := legacy.ClientP2PTradeOffer(false, 0); dr == nil || count != 1 || cost != 41 {
			e2eError(fmt.Errorf("trade removal left drawable=%p count=%d cost=%d, want count=1 cost=41", dr, count, cost))
		}
		close := server.BuildTradeP2PClosePacket50F450()
		deliver(close[:])
		if _, active := legacy.ClientP2PTradeWindow(); active {
			e2eError(fmt.Errorf("trade window stayed active after closing"))
		}
		if dr, count, cost := legacy.ClientP2PTradeOffer(false, 0); dr != nil || count != 0 || cost != 0 {
			e2eError(fmt.Errorf("trade close left drawable=%p count=%d cost=%d", dr, count, cost))
		}
		e2eLog.Printf("P2P TRADE CLOSED: stacked offer removal and slot reset verified")
	})
}
