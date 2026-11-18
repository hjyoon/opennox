package legacy

import (
	"bytes"
	"fmt"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

func TestShopExitNativeCABI50F4C0(t *testing.T) {
	oldFlags := noxflags.GetGame()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})
	for _, quest := range []bool{false, true} {
		for route := range 2 {
			t.Run(fmt.Sprintf("quest%t-route%d", quest, route), func(t *testing.T) {
				if quest {
					noxflags.SetGame(noxflags.GameModeQuest)
				} else {
					noxflags.UnsetGame(noxflags.GameModeQuest)
				}
				world := server.NewHigh(nil, nil, strman.New())
				defer world.Close()
				game4ConsumerServices(t, game4TradeServices{world: world})
				unit, player := game4ConsumerPlayer(t, 0)
				player.PlayerInd = 31
				merchant := game4ConsumerValue(t, server.Object{ObjClass: object.ClassMonster})
				session := world.NewShopSessionNative50E8F0(unit, merchant)
				session.Field0 = 1
				update := unit.UpdateDataPlayer()
				update.Trade70 = session
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(session)) <= 0xffffffff {
					t.Fatal("session must exercise high addresses")
				}
				sends := 0
				world.NetSendPacketXxx = func(index int, packet []byte, related *server.Object, remove, sequence int) int {
					sends++
					if sends != 1 || index != 31 || related != nil || remove != 1 || sequence != 1 ||
						!bytes.Equal(packet, []byte{0xc9, 0x02}) || session.Field0 != 0 || update.Trade70 != nil {
						t.Fatal("shop close state or packet")
					}
					return 0
				}
				shopExitFixtureCall(route, session)
				if sends != 1 || world.IsTradeSessionNative(session) != quest || update.Trade70 != nil {
					t.Fatal("shop close did not establish the expected ownership")
				}
				if got := world.ClearQuestShopSessionNative510E20(31); got != quest {
					t.Fatalf("cached shop close = %t, want %t", got, quest)
				}
			})
		}
	}
}
