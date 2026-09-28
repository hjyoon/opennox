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
	objectWorth func(string) (int32, bool)
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

func (rt shopItemCostRuntime50E3D0) objectWorthValue(id string) (int32, bool) {
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
		objectWorth: func(id string) (int32, bool) {
			typ := s.Types.ByID(id)
			if typ == nil {
				return 0, false
			}
			return int32(typ.Worth), true
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
// GAME.EXE 0050E3D0. Each assignment to the original float locals is rounded
// to binary32, while expressions use float64 to model the x87 intermediates.
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

	// The source first converts the unsigned worth to an x87 value and spills
	// it to a float local. Values above 2^24 intentionally lose low bits here.
	price := float32(item.Worth)
	class := item.Class()
	subclass := uint32(item.SubClass())

	if class.Has(object.ClassInfoBook) {
		if subclass&1 != 0 && item.UseData.Ptr != nil {
			price = float32(rt.spellPriceValue(item.UseData.AsSpellReward().Spell))
		} else if subclass&2 != 0 && item.UseData.Ptr != nil {
			if worth, ok := rt.objectWorthValue(item.UseData.AsFieldGuide().Creature()); ok && worth >= 0 {
				price = float32(worth)
			}
			if rt.quest {
				price = float32(float64(rt.balanceValue("QuestGuideWorthMultiplier")) * float64(price))
			}
		}
	}

	if class.HasAny(shopModifierClassMask50E3D0) && item.InitData != nil {
		for _, modifier := range item.InitDataModifier().Modifiers {
			if modifier == nil {
				continue
			}
			value := float32(modifier.Price20)
			if rt.quest {
				value = float32(float64(rt.balanceValue("QuestModifierWorthMultiplier")) * float64(value))
			}
			price = float32(float64(price) + float64(value))
		}
	}

	// Full price is captured before ammunition or wand-charge scaling. Repair
	// therefore charges for restoring durability only, just like the source.
	fullPrice := price
	if class.Has(object.ClassWeapon) && subclass&0x82 != 0 && item.UseData.Ptr != nil {
		ammo := item.UseData.AsAmmo()
		if ammo.Field2 == 0 && ammo.Charge0 != 0 {
			key := "DefaultAmmoAmount"
			if rt.quest {
				key = "DefaultAmmoAmountQuest"
			}
			defaultAmount := questInventoryRoundFloat32ToInt32_4F2C30(rt.balanceValue(key))
			if int32(ammo.Charge1) > defaultAmount {
				perShot := float32(float64(price) / float64(defaultAmount))
				price = float32(float64(int32(ammo.Charge1)-defaultAmount)*float64(perShot) + float64(price))
			} else {
				price = float32(float64(ammo.Charge1) / float64(defaultAmount) * float64(price))
			}
		}
	} else if class.Has(object.ClassWand) && subclass&0x047f0000 != 0 && item.UseData.Ptr != nil {
		wand := item.UseData.AsWand()
		if wand.MaxCharge != 0 {
			price = float32(float64(wand.Charge) / float64(wand.MaxCharge) * float64(price))
		}
	}

	if shop != nil && mode != shopPriceSell50E3D0 {
		price = float32(float64(price) * float64(shop.BuyMultiplier))
		fullPrice = float32(float64(fullPrice) * float64(shop.BuyMultiplier))
	}
	if health := item.HealthData; health != nil && health.Max != 0 {
		price = float32(float64(health.Cur) / float64(health.Max) * float64(price))
	}
	if shop != nil && mode == shopPriceSell50E3D0 && !rt.isGem(item.TypeInd) {
		multiplier := shop.SellMultiplier
		if rt.quest {
			multiplier = rt.balanceValue("QuestSellMultiplier")
		}
		price = float32(float64(multiplier) * float64(price))
		fullPrice = float32(float64(multiplier) * float64(fullPrice))
	}

	if rt.quest && mode == shopPriceSell50E3D0 && class.HasAny(shopQuestUsedClassMask50E3D0) &&
		item.UpdateData != nil && item.UpdateDataWeaponArmor().Field4&1 != 0 {
		fullPrice = 0
		price = 1
	} else if !(price >= 1) {
		// The negated comparison deliberately turns NaN into the minimum price.
		price = 1
	}
	if fullPrice < 1 {
		fullPrice = 1
	}
	if mode == shopPriceRepair50E3D0 {
		price = float32(float64(repairCoefficient) * (float64(fullPrice) - float64(price)))
		if price < 1 {
			price = 1
		}
	}
	return questInventoryRoundFloat32ToInt32_4F2C30(price), true
}

// ShopItemCostNoMerchant50E3D0 is the mode=1, nil-session path called by
// 00513B00. Unlike the PE32 C routine, it never reinterprets an Object pointer
// as a float32 argument.
func (s *Server) ShopItemCostNoMerchant50E3D0(item *Object) int32 {
	cost, ok := shopItemCostNative50E3D0(nil, item, shopPriceBuy50E3D0, 0, s.shopItemCostRuntime50E3D0())
	if !ok {
		return 1
	}
	return cost
}
