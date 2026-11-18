package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

const shopQuestUsedClassMask50E3D0 = object.ClassWand | object.ClassWeapon | object.ClassArmor

// shopItemCostRuntime50E3D0 isolates the services read by GAME.EXE 0050E3D0
// from its binary32 arithmetic. Tests can therefore cover every pricing
// branch without constructing a complete game-data server.
type shopItemCostRuntime50E3D0 struct {
	quest       bool
	balance     func(string) float32
	spellPrice  func(uint8) uint16
	objectWorth func(string) (uint32, bool)
	objectType  func(string) (uint16, bool)
}

func (rt shopItemCostRuntime50E3D0) balanceValue(key string) float32 {
	if rt.balance == nil {
		return 0
	}
	return rt.balance(key)
}

func (rt shopItemCostRuntime50E3D0) spellPriceValue(ind uint8) uint16 {
	if rt.spellPrice == nil {
		return 0
	}
	return rt.spellPrice(ind)
}

func (rt shopItemCostRuntime50E3D0) objectWorthValue(id string) (uint32, bool) {
	if rt.objectWorth == nil {
		return 0, false
	}
	return rt.objectWorth(id)
}

func (rt shopItemCostRuntime50E3D0) isGem(typeInd uint16) bool {
	if rt.objectType == nil {
		return false
	}
	for _, id := range [...]string{"Diamond", "Emerald", "Ruby"} {
		if ind, ok := rt.objectType(id); ok && typeInd == ind {
			return true
		}
	}
	return false
}

func (s *Server) shopItemCostRuntime50E3D0() shopItemCostRuntime50E3D0 {
	return shopItemCostRuntime50E3D0{
		quest: noxflags.HasGame(noxflags.GameModeQuest),
		balance: func(key string) float32 {
			return float32(s.Balance.Float(key))
		},
		spellPrice: func(ind uint8) uint16 {
			return uint16(s.Spells.Price(spell.ID(ind)))
		},
		objectWorth: func(id string) (uint32, bool) {
			typ := s.Types.ByID(id)
			if typ == nil {
				return 0, false
			}
			return uint32(typ.Worth), true
		},
		objectType: func(id string) (uint16, bool) {
			typ := s.Types.ByID(id)
			if typ == nil {
				return 0, false
			}
			return uint16(typ.Ind()), true
		},
	}
}

// shopItemCostNative50E3D0 is the pointer-width-safe implementation of
// GAME.EXE 0050E3D0. The startup x87 mode rounds both the 53-bit operations
// and stores to binary32 toward zero.
func shopItemCostNative50E3D0(
	session *TradeSession,
	item *Object,
	mode shopPriceMode50E3D0,
	repairCoefficient float32,
	rt shopItemCostRuntime50E3D0,
) (int32, bool) {
	if item == nil {
		return 0, false
	}

	var shop *ShopkeeperInitData
	if session != nil && session.Field16 != 0 {
		merchant := session.Field8
		if merchant != nil && merchant.Class().Has(object.ClassPlayer) {
			merchant = session.Field12
		}
		if merchant == nil || merchant.InitData == nil {
			return 0, false
		}
		shop = merchant.InitDataShopkeeper()
	}

	store32 := experienceSpill32TowardZero
	add53, mul53 := experienceAdd53TowardZero, experienceMul53TowardZero
	div53 := monsterMoveForceDivChop53_50D581
	price := store32(float64(item.Worth))
	class := item.Class()
	subclass := uint32(item.SubClass())

	if class.Has(object.ClassInfoBook) {
		if subclass&1 != 0 && item.UseData.Ptr != nil {
			price = float32(rt.spellPriceValue(item.UseData.AsSpellReward().Spell))
		} else if subclass&2 != 0 && item.UseData.Ptr != nil {
			if worth, ok := rt.objectWorthValue(item.UseData.AsFieldGuide().Creature()); ok {
				price = store32(float64(worth))
			}
			if rt.quest {
				price = store32(mul53(float64(rt.balanceValue("QuestGuideWorthMultiplier")), float64(price)))
			}
		}
	}

	if class.HasAny(shopModifierClassMask50E3D0) && item.InitData != nil {
		for _, modifier := range item.InitDataModifier().Modifiers {
			if modifier == nil {
				continue
			}
			value := float64(store32(float64(modifier.Price20)))
			if rt.quest {
				value = mul53(float64(rt.balanceValue("QuestModifierWorthMultiplier")), value)
			}
			price = store32(add53(float64(price), value))
		}
	}

	// Full price is captured before ammunition or wand-charge scaling. Repair
	// compares it with the charge- and durability-adjusted current value.
	fullPrice := price
	if class.Has(object.ClassWeapon) && subclass&0x82 != 0 && item.UseData.Ptr != nil {
		ammo := item.UseData.AsAmmo()
		if ammo.Field2 == 0 && ammo.Charge0 != 0 {
			key := "DefaultAmmoAmount"
			if rt.quest {
				key = "DefaultAmmoAmountQuest"
			}
			defaultAmount := netFXTruncateSignedDWORD419A70(rt.balanceValue(key))
			if int32(ammo.Charge1) > defaultAmount {
				perShot := store32(div53(float64(price), float64(defaultAmount)))
				extra := mul53(float64(int32(ammo.Charge1)-defaultAmount), float64(perShot))
				price = store32(add53(extra, float64(price)))
			} else {
				ratio := div53(float64(ammo.Charge1), float64(defaultAmount))
				price = store32(mul53(ratio, float64(price)))
			}
		}
	} else if class.Has(object.ClassWand) && subclass&0x047f0000 != 0 && item.UseData.Ptr != nil {
		wand := item.UseData.AsWand()
		if wand.MaxCharge != 0 {
			ratio := div53(float64(wand.Charge), float64(wand.MaxCharge))
			price = store32(mul53(ratio, float64(price)))
		}
	}

	if shop != nil && mode != shopPriceSell50E3D0 {
		price = store32(mul53(float64(price), float64(shop.BuyMultiplier)))
		fullPrice = store32(mul53(float64(fullPrice), float64(shop.BuyMultiplier)))
	}
	if health := item.HealthData; health != nil && health.Max != 0 {
		ratio := div53(float64(health.Cur), float64(health.Max))
		price = store32(mul53(ratio, float64(price)))
	}
	if shop != nil && mode == shopPriceSell50E3D0 && !rt.isGem(item.TypeInd) {
		multiplier := shop.SellMultiplier
		if rt.quest {
			multiplier = rt.balanceValue("QuestSellMultiplier")
		}
		price = store32(mul53(float64(multiplier), float64(price)))
		fullPrice = store32(mul53(float64(multiplier), float64(fullPrice)))
	}

	if rt.quest && mode == shopPriceSell50E3D0 && class.HasAny(shopQuestUsedClassMask50E3D0) &&
		item.UpdateData != nil && item.UpdateDataWeaponArmor().Field4&1 != 0 {
		fullPrice = 0
		price = 1
	} else if !(price >= 1) {
		// The negated comparison deliberately turns NaN into the minimum price.
		price = 1
	}
	if !(fullPrice >= 1) {
		fullPrice = 1
	}
	if mode == shopPriceRepair50E3D0 {
		difference := add53(float64(fullPrice), -float64(price))
		calculated := mul53(float64(repairCoefficient), difference)
		price = store32(calculated)
		if !(calculated >= 1) {
			price = 1
		}
	}
	return netFXTruncateSignedDWORD419A70(price), true
}

// ShopItemCostByModeNative50E3D0 prices the item for the public C entry and
// the retained PE32 pool path.
func (s *Server) ShopItemCostByModeNative50E3D0(mode int32, session *TradeSession, item *Object) int32 {
	priceMode := shopPriceBuy50E3D0
	var coefficient float32
	switch mode {
	case 0:
		priceMode = shopPriceSell50E3D0
	case 2:
		priceMode = shopPriceRepair50E3D0
		coefficient = float32(s.Balance.Float("RepairCoefficient"))
	}
	cost, ok := shopItemCostNative50E3D0(session, item, priceMode, coefficient, s.shopItemCostRuntime50E3D0())
	if !ok {
		return 1
	}
	return cost
}

// ShopItemCostNoMerchant50E3D0 is the mode=1, nil-session path called by
// 00513B00. Unlike the PE32 C routine, it never reinterprets an Object pointer
// as a float32 argument.
func (s *Server) ShopItemCostNoMerchant50E3D0(item *Object) int32 {
	return s.ShopItemCostByModeNative50E3D0(1, nil, item)
}
