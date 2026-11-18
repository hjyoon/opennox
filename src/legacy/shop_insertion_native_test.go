package legacy

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func TestShopInsertionNativeCABI50EE00(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.UnsetGame(noxflags.GameModeQuest)
	t.Cleanup(func() { noxflags.SetGame(oldFlags) })
	for route := range 2 {
		t.Run(fmt.Sprintf("route%d", route), func(t *testing.T) {
			world := server.NewHigh(nil, nil, strman.New())
			defer world.Close()
			if !world.Objs.InitHigh(8) {
				t.Fatal("object pool")
			}
			defer world.Objs.FreeObjects()
			defer world.FreeObjectTypes()
			world.FreeObjectTypes()
			price := 10
			if err := world.Types.ReadObjectType(&things.Thing{Name: "NativeShopInsertionItem", Price: &price}); err != nil {
				t.Fatal(err)
			}
			owner, _ := game4ConsumerPlayer(t, 0)
			init := game4ConsumerValue(t, server.ShopkeeperInitData{BuyMultiplier: 1, SellMultiplier: 1})
			merchant := game4ConsumerValue(t, server.Object{ObjClass: object.ClassMonster, InitData: unsafe.Pointer(init)})
			session := world.NewShopSessionNative50E8F0(owner, merchant)
			defer world.ReleaseTradeSessionNative510000(session)
			game4ConsumerServices(t, game4TradeServices{world: world})

			costs := []uint32{50, 5, 20, 20, 250}
			nodes := make([]*server.TradeItem, len(costs))
			for i, cost := range costs {
				item := world.NewObjectByTypeID("NativeShopInsertionItem")
				if item == nil {
					t.Fatal("fixture item")
				}
				item.Worth = cost
				node := shopInsertionFixtureCall(route, session, item)
				if node == nil {
					world.Objs.FreeObject(item)
					t.Fatalf("insert cost %d", cost)
				}
				if node.Item0 != item || node.Cost4 != cost {
					t.Fatalf("node item %p, cost %d, want %p %d", node.Item0, node.Cost4, item, cost)
				}
				if unsafe.Sizeof(uintptr(0)) == 8 {
					for _, p := range []unsafe.Pointer{unsafe.Pointer(session), unsafe.Pointer(node), unsafe.Pointer(item)} {
						if uintptr(p) <= 0xffffffff {
							t.Fatalf("fixture did not allocate above 4 GiB: %p", p)
						}
					}
				}
				nodes[i] = node
			}
			previous, head := (*server.TradeItem)(nil), session.Field20
			for _, i := range []int{1, 3, 2, 0, 4} {
				if head != nodes[i] || head.Field12 != previous {
					t.Fatalf("bad insertion order, linked node %p, want %p", head, nodes[i])
				}
				previous, head = head, head.Field8
			}
			if head != nil {
				t.Fatal("additional linked node")
			}
			if node := shopInsertionFixtureCall(route, session, nil); node != nil || session.Field20 != nodes[1] {
				t.Fatal("nil item changed stock")
			}
		})
	}
}

func TestShopInsertionPE32PoolOwnership50EE00(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 4 {
		t.Skip("fixed-width trade pool runs on 32-bit hosts")
	}
	oldFlags := noxflags.GetGame()
	noxflags.UnsetGame(noxflags.GameModeQuest)
	t.Cleanup(func() { noxflags.SetGame(oldFlags) })
	cache := memmap.PtrT[[3]uint32](0x5D4594, 2386504)
	previous := *cache
	*cache = [3]uint32{1, 2, 3}
	t.Cleanup(func() { *cache = previous })
	restore := shopInsertionFixtureUsePE32Pool(2)
	if restore == nil {
		t.Fatal("create legacy trade pool")
	}
	t.Cleanup(restore)
	world := &server.Server{}
	game4ConsumerServices(t, game4TradeServices{world: world})
	session := game4ConsumerValue(t, server.TradeSession{})
	objects := []*server.Object{
		game4ConsumerValue(t, server.Object{Worth: 30}),
		game4ConsumerValue(t, server.Object{Worth: 12}),
		game4ConsumerValue(t, server.Object{Worth: 5}),
	}
	first := shopInsertionFixtureCall(0, session, objects[0])
	second := shopInsertionFixtureCall(1, session, objects[1])
	if first == nil || second == nil || first.Cost4 != 30 || second.Cost4 != 12 ||
		first.Item0 != objects[0] || second.Item0 != objects[1] {
		t.Fatalf("legacy pool insertion returned %p and %p", first, second)
	}
	if session.Field20 != second || second.Field8 != first || first.Field12 != second || first.Field8 != nil {
		t.Fatal("incorrect legacy list links")
	}
	if last := shopInsertionFixtureCall(0, session, objects[2]); last != nil || session.Field20 != second ||
		second.Field8 != first || first.Field8 != nil {
		t.Fatal("exhausted legacy pool changed stock")
	}
	if world.IsTradeSessionNative(session) {
		t.Fatal("legacy session changed owners")
	}
}
