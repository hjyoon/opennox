package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"
)

func TestQuestShopStartupCutoffOriginal50E970(t *testing.T) {
	s := NewHigh(nil, nil, strman.New())
	if !s.Objs.InitHigh(4) {
		t.Fatal("object pool initialization failed")
	}
	defer s.Objs.FreeObjects()
	s.FreeObjectTypes()
	defer s.FreeObjectTypes()
	defer s.Close()
	price := 40
	if err := s.Types.ReadObjectType(&things.Thing{Name: "AnkhTradable", Price: &price}); err != nil {
		t.Fatal(err)
	}
	init := nativeTradeTestValue(t, ShopkeeperInitData{BuyMultiplier: 1})
	merchant := nativeTradeTestValue(t, Object{ObjClass: object.ClassMonster, InitData: unsafe.Pointer(init)})
	player := nativeTradeTestValue(t, Object{ObjClass: object.ClassPlayer})
	session := s.NewShopSessionNative50E8F0(player, merchant)
	loaded, _ := s.LoadQuestShopItemsNative50E970(session, QuestShopItemLoadRuntime50E970{
		QuestStage:      3,
		AnkhCutoffStage: 3.5,
		DelayedDelete:   func(*Object) {},
	})
	if loaded != 0 || session.Field20 != nil {
		t.Fatalf("Quest shop at stage 3 loaded %d items with cutoff 3.5, original startup loads zero", loaded)
	}
}
