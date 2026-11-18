package legacy

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"

	"github.com/opennox/opennox/v1/server"
)

func TestShopConstructorsNativeCABI50E87050E8F050E970(t *testing.T) {
	for route := range 2 {
		t.Run(fmt.Sprintf("route%d", route), func(t *testing.T) {
			world := server.NewHigh(nil, nil, strman.New())
			defer world.Close()
			if !world.Objs.InitHigh(32) {
				t.Fatal("fixture object pool")
			}
			defer world.Objs.FreeObjects()
			defer world.FreeObjectTypes()
			defer world.TradeFree50E300()
			world.FreeObjectTypes()
			price := 27
			for _, name := range []string{"Gold", "CShopStock", "Diamond"} {
				if err := world.Types.ReadObjectType(&things.Thing{Name: name, Price: &price}); err != nil {
					t.Fatal(err)
				}
			}
			player, info := game4ConsumerPlayer(t, 0)
			info.PlayerInd = 31
			regularType := uint16(world.Types.ByID("CShopStock").Ind())
			diamondType := uint16(world.Types.ByID("Diamond").Ind())
			init := game4ConsumerValue(t, server.ShopkeeperInitData{Count: 1, BuyMultiplier: 1, SellMultiplier: 1})
			init.Items[0] = server.ShopkeeperItemDefinition{TypeInd: uint32(regularType), Count: 3}
			merchant := game4ConsumerValue(t, server.Object{ObjClass: object.ClassMonster, InitData: unsafe.Pointer(init)})
			other := game4ConsumerValue(t, server.Object{ObjClass: object.ClassMonster, InitData: unsafe.Pointer(init)})

			quest, loads := false, 0
			load := func(session *server.TradeSession) {
				loads++
				if quest {
					count, complete := world.LoadQuestShopItemsNative50E970(session, server.QuestShopItemLoadRuntime50E970{
						AnkhCutoffStage: 3, QuestStage: 3, DelayedDelete: func(*server.Object) {},
					})
					if count != 5 || complete {
						t.Fatalf("partial Quest fixture stock count = %d, complete = %t", count, complete)
					}
					return
				}
				count, complete := world.LoadRegularShopItemsNative50E970(session, server.ShopItemLoadRuntime50E970{})
				if count != 3 || !complete {
					t.Fatalf("regular fixture stock count = %d, complete = %t", count, complete)
				}
			}
			game4ConsumerServices(t, game4TradeServices{world: world, load: load, open: func(unit, keeper *server.Object) *server.TradeSession {
				if unit != player || keeper != merchant && keeper != other {
					t.Fatalf("open received unexpected participants %p and %p", unit, keeper)
				}
				session, reused := world.OpenShopSessionNative50E8F0(unit, keeper, quest, info.Index())
				if !reused {
					load(session)
				}
				return session
			}})

			first := shopConstructorAllocateFixture(route)
			second := shopConstructorAllocateFixture(route)
			checkShopConstructorSession(t, world, first)
			checkShopConstructorSession(t, world, second)
			if second == first || second.Field56 != first || first.Field60 != second || second.Field60 != nil ||
				first.Field8 != nil || first.Field12 != nil || first.Field16 != 0 ||
				second.Field48 == first.Field48 || second.Field52 == first.Field52 {
				t.Fatal("empty session allocation order or Gold ownership")
			}
			first.Field8, first.Field12, first.Field16 = merchant, player, 1
			shopConstructorLoadFixture(route, first)
			checkShopConstructorStock(t, first, regularType, 3)

			regular := shopConstructorOpenFixture(route, player, merchant)
			checkShopConstructorSession(t, world, regular)
			if regular.Field8 != player || regular.Field12 != merchant || regular.Field16 != 1 || loads != 2 {
				t.Fatal("fresh shop participants or initial stock loading")
			}
			checkShopConstructorStock(t, regular, regularType, 3)
			stock := regular.Field20
			if !world.CacheQuestShopSessionNative50F4C0(info.Index(), regular) {
				t.Fatal("Quest fixture cache")
			}
			quest = true
			reused := shopConstructorOpenFixture(route, player, other)
			if reused != regular || reused.Field8 != player || reused.Field12 != other || reused.Field16 != 1 ||
				reused.Field20 != stock || loads != 2 || world.ClearQuestShopSessionNative510E20(info.Index()) {
				t.Fatal("cached session or stock changed during reopening")
			}
			second.Field8, second.Field12, second.Field16 = player, merchant, 1
			shopConstructorLoadFixture(route, second)
			checkShopConstructorStock(t, second, diamondType, 5)
		})
	}
}

func checkShopConstructorSession(t *testing.T, world *server.Server, session *server.TradeSession) {
	t.Helper()
	if !world.IsTradeSessionNative(session) || session.Field48 == nil || session.Field52 == nil || session.Field48 == session.Field52 {
		t.Fatal("session and its two private Gold objects")
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(session), unsafe.Pointer(session.Field48), unsafe.Pointer(session.Field52)} {
			if uintptr(ptr) <= 0xffffffff {
				t.Fatalf("fixture did not exercise the full address of %p", ptr)
			}
		}
	}
}

func checkShopConstructorStock(t *testing.T, session *server.TradeSession, typeIndex uint16, count int) {
	t.Helper()
	node := session.Field20
	var previous *server.TradeItem
	for i := 0; i < count; i++ {
		if node == nil || node.Item0.TypeInd != typeIndex || node.Cost4 != 27 || node.Field12 != previous {
			t.Fatalf("invalid stock node %d of %d", i, count)
		}
		previous, node = node, node.Field8
	}
	if node != nil {
		t.Fatalf("stock contains more than %d nodes", count)
	}
}
