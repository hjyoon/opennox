package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func testShopItemCostRuntime50E3D0(quest bool) shopItemCostRuntime50E3D0 {
	balances := map[string]float32{
		"DefaultAmmoAmount":            20,
		"DefaultAmmoAmountQuest":       10,
		"QuestGuideWorthMultiplier":    1.25,
		"QuestModifierWorthMultiplier": 1.5,
		"QuestSellMultiplier":          0.4,
	}
	objectTypes := map[string]uint16{
		"Diamond": 11,
		"Emerald": 12,
		"Ruby":    13,
	}
	return shopItemCostRuntime50E3D0{
		quest: quest,
		balance: func(key string) float32 {
			return balances[key]
		},
		spellPrice: func(ind uint8) uint16 {
			if ind != 7 {
				return 0
			}
			return 321
		},
		objectWorth: func(id string) (int32, bool) {
			if id == "Bear" {
				return 80, true
			}
			return 0, false
		},
		objectType: func(id string) (uint16, bool) {
			ind, ok := objectTypes[id]
			return ind, ok
		},
	}
}

func testShopItemCostSession50E3D0(buy, sell float32) *TradeSession {
	idata := &ShopkeeperInitData{BuyMultiplier: buy, SellMultiplier: sell}
	merchant := &Object{ObjClass: object.ClassMonster, InitData: unsafe.Pointer(idata)}
	return &TradeSession{
		Field8:  &Object{ObjClass: object.ClassPlayer},
		Field12: merchant,
		Field16: 1,
	}
}

func TestShopItemCostNoMerchant50E3D0BasicModifierAndHealth(t *testing.T) {
	s := &Server{}
	item := &Object{Worth: 76, ObjClass: object.ClassFood}
	if got := s.ShopItemCostNoMerchant50E3D0(item); got != 76 {
		t.Fatalf("basic cost = %d", got)
	}
	item.Worth = 0
	if got := s.ShopItemCostNoMerchant50E3D0(item); got != 1 {
		t.Fatalf("minimum cost = %d", got)
	}
	item.Worth = 1<<24 + 1
	if got := s.ShopItemCostNoMerchant50E3D0(item); got != 1<<24 {
		t.Fatalf("binary32 worth spill = %d", got)
	}
	mod := &ModifierEff{Price20: 31}
	attrs := &ModifierInitData{Modifiers: [4]*ModifierEff{mod}}
	item.ObjClass = object.ClassArmor
	item.Worth = 69
	item.InitData = unsafe.Pointer(attrs)
	item.HealthData = &HealthData{Cur: 1, Max: 2}
	if got := s.ShopItemCostNoMerchant50E3D0(item); got != 50 {
		t.Fatalf("modifier and health cost = %d", got)
	}
}

func TestShopItemCostNoMerchant50E3D0WandCharge(t *testing.T) {
	s := &Server{}
	item := &Object{Worth: 100, ObjClass: object.ClassWand, ObjSubClass: 0x10000}
	data := &WandUseData{Charge: 3, MaxCharge: 4}
	item.UseData.Ptr = unsafe.Pointer(data)
	item.InitData = unsafe.Pointer(&ModifierInitData{})
	if got := s.ShopItemCostNoMerchant50E3D0(item); got != 75 {
		t.Fatalf("wand charge cost = %d", got)
	}
}

func TestShopItemCostNoMerchant50E3D0MissingOptionalData(t *testing.T) {
	s := &Server{}
	for _, class := range []object.Class{object.ClassArmor, object.ClassWeapon, object.ClassWand} {
		item := &Object{Worth: 40, ObjClass: class, ObjSubClass: 0x10000}
		if got := s.ShopItemCostNoMerchant50E3D0(item); got != 40 {
			t.Fatalf("class %v without optional data: cost = %d", class, got)
		}
	}
}

func TestShopItemCostNative50E3D0InformationBooks(t *testing.T) {
	rt := testShopItemCostRuntime50E3D0(true)
	spellReward := &SpellRewardUseData{Spell: 7}
	spellBook := &Object{Worth: 40, ObjClass: object.ClassInfoBook, ObjSubClass: 1}
	spellBook.UseData.Ptr = unsafe.Pointer(spellReward)
	if got, ok := shopItemCostNative50E3D0(nil, spellBook, shopPriceBuy50E3D0, 0, rt); !ok || got != 321 {
		t.Fatalf("spell reward = %d, %t, want 321, true", got, ok)
	}

	guideData := &FieldGuideUseData{}
	guideData.SetCreature("Bear")
	guide := &Object{Worth: 40, ObjClass: object.ClassInfoBook, ObjSubClass: 2}
	guide.UseData.Ptr = unsafe.Pointer(guideData)
	if got, ok := shopItemCostNative50E3D0(nil, guide, shopPriceBuy50E3D0, 0, rt); !ok || got != 100 {
		t.Fatalf("quest field guide = %d, %t, want 100, true", got, ok)
	}
}

func TestShopItemCostNative50E3D0QuestModifier(t *testing.T) {
	modifier := &ModifierEff{Price20: 20}
	attrs := &ModifierInitData{Modifiers: [4]*ModifierEff{modifier}}
	item := &Object{Worth: 100, ObjClass: object.ClassArmor, InitData: unsafe.Pointer(attrs)}
	if got, ok := shopItemCostNative50E3D0(
		nil, item, shopPriceBuy50E3D0, 0, testShopItemCostRuntime50E3D0(true),
	); !ok || got != 130 {
		t.Fatalf("quest modifier = %d, %t, want 130, true", got, ok)
	}
}

func TestShopItemCostNative50E3D0AmmoScalingAndRepairBaseline(t *testing.T) {
	session := testShopItemCostSession50E3D0(2, 0.5)
	ammo := &AmmoUseData{Charge0: 1, Charge1: 30}
	item := &Object{
		Worth:       100,
		ObjClass:    object.ClassWeapon,
		ObjSubClass: 0x82,
		HealthData:  &HealthData{Cur: 50, Max: 100},
	}
	item.UseData.Ptr = unsafe.Pointer(ammo)
	rt := testShopItemCostRuntime50E3D0(false)
	if got, ok := shopItemCostNative50E3D0(session, item, shopPriceBuy50E3D0, 0, rt); !ok || got != 150 {
		t.Fatalf("above-default ammo buy = %d, %t, want 150, true", got, ok)
	}
	if got, ok := shopItemCostNative50E3D0(session, item, shopPriceRepair50E3D0, 0.5, rt); !ok || got != 25 {
		t.Fatalf("ammo durability repair = %d, %t, want 25, true", got, ok)
	}
	ammo.Charge1 = 10
	item.HealthData = nil
	if got, ok := shopItemCostNative50E3D0(nil, item, shopPriceBuy50E3D0, 0, rt); !ok || got != 50 {
		t.Fatalf("below-default ammo = %d, %t, want 50, true", got, ok)
	}
}

func TestShopItemCostNative50E3D0WandRepairBaseline(t *testing.T) {
	session := testShopItemCostSession50E3D0(2, 0.5)
	wand := &WandUseData{Charge: 3, MaxCharge: 4}
	item := &Object{
		Worth:       100,
		ObjClass:    object.ClassWand,
		ObjSubClass: 0x10000,
		InitData:    unsafe.Pointer(&ModifierInitData{}),
		HealthData:  &HealthData{Cur: 50, Max: 100},
	}
	item.UseData.Ptr = unsafe.Pointer(wand)
	if got, ok := shopItemCostNative50E3D0(
		session, item, shopPriceRepair50E3D0, 0.5, testShopItemCostRuntime50E3D0(false),
	); !ok || got != 62 {
		t.Fatalf("wand durability repair = %d, %t, want 62, true", got, ok)
	}
}

func TestShopItemCostNative50E3D0SellMultipliersAndGemExemption(t *testing.T) {
	session := testShopItemCostSession50E3D0(2, 0.25)
	item := &Object{Worth: 100, TypeInd: 14, ObjClass: object.ClassFood}
	regular := testShopItemCostRuntime50E3D0(false)
	if got, ok := shopItemCostNative50E3D0(session, item, shopPriceSell50E3D0, 0, regular); !ok || got != 25 {
		t.Fatalf("regular sale = %d, %t, want 25, true", got, ok)
	}
	item.TypeInd = 11
	if got, ok := shopItemCostNative50E3D0(session, item, shopPriceSell50E3D0, 0, regular); !ok || got != 100 {
		t.Fatalf("diamond sale = %d, %t, want 100, true", got, ok)
	}
	item.TypeInd = 14
	if got, ok := shopItemCostNative50E3D0(
		session, item, shopPriceSell50E3D0, 0, testShopItemCostRuntime50E3D0(true),
	); !ok || got != 40 {
		t.Fatalf("quest sale = %d, %t, want 40, true", got, ok)
	}
}

func TestShopItemCostNative50E3D0QuestUsedItemMinimum(t *testing.T) {
	item := &Object{
		Worth:      100,
		TypeInd:    14,
		ObjClass:   object.ClassWeapon,
		InitData:   unsafe.Pointer(&ModifierInitData{}),
		UpdateData: unsafe.Pointer(&WeaponArmorUpdateData{Field4: 1}),
	}
	if got, ok := shopItemCostNative50E3D0(
		testShopItemCostSession50E3D0(2, 0.25), item, shopPriceSell50E3D0, 0,
		testShopItemCostRuntime50E3D0(true),
	); !ok || got != 1 {
		t.Fatalf("quest used-item sale = %d, %t, want 1, true", got, ok)
	}
}

func TestShopItemCostNative50E3D0RejectsInvalidTransactionResult(t *testing.T) {
	s := &Server{}
	ammo := &AmmoUseData{Charge0: 1, Charge1: 1}
	item := &Object{Worth: 100, ObjClass: object.ClassWeapon, ObjSubClass: 0x82}
	item.UseData.Ptr = unsafe.Pointer(ammo)
	rt := testShopItemCostRuntime50E3D0(false)
	rt.balance = func(string) float32 { return 0 }
	if got, ok := shopItemCostNative50E3D0(nil, item, shopPriceBuy50E3D0, 0, rt); !ok || got != math.MinInt32 {
		t.Fatalf("integer-indefinite cost = %d, %t, want %d, true", got, ok, math.MinInt32)
	}
	if _, ok := s.shopInventoryItemCost50E3D0(nil, item, shopPriceSell50E3D0, 0); ok {
		t.Fatal("nil merchant session was accepted")
	}
	badSession := &TradeSession{Field16: 1}
	if _, ok := shopItemCostNative50E3D0(badSession, item, shopPriceBuy50E3D0, 0, rt); ok {
		t.Fatal("active session without a merchant was accepted")
	}
}
