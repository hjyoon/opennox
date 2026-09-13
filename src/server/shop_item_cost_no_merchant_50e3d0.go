package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

// ShopItemCostNoMerchant50E3D0 is the GAME.EXE 0050E3D0 pricing path with
// mode=1 and a nil trade session, as called by 00513B00. Unlike the old C
// function, it never reinterprets an Object pointer as a float32 argument.
func (s *Server) ShopItemCostNoMerchant50E3D0(item *Object) int32 {
	quest := noxflags.HasGame(noxflags.GameModeQuest)
	// The PE32 routine spills v31 to binary32 after each assignment. In
	// particular, the initial unsigned worth loses low bits above 2^24.
	price := float32(item.Worth)
	class := uint32(item.ObjClass)
	subclass := uint32(item.ObjSubClass)
	if class&uint32(object.ClassInfoBook) != 0 {
		if subclass&1 != 0 && item.UseData.Ptr != nil {
			price = float32(uint16(s.Spells.Price(spell.ID(item.UseData.AsSpellReward().Spell))))
		} else if subclass&2 != 0 && item.UseData.Ptr != nil {
			if typ := s.Types.ByID(item.UseData.AsFieldGuide().Creature()); typ != nil && typ.Worth >= 0 {
				price = float32(typ.Worth)
			}
			if quest {
				price = float32(float64(float32(s.Balance.Float("QuestGuideWorthMultiplier"))) * float64(price))
			}
		}
	}
	if class&0x13001000 != 0 {
		if attrs := item.InitDataModifier(); attrs != nil {
			for _, mod := range attrs.Modifiers {
				if mod == nil {
					continue
				}
				value := float32(mod.Price20)
				if quest {
					value = float32(float64(float32(s.Balance.Float("QuestModifierWorthMultiplier"))) * float64(value))
				}
				price = float32(float64(price) + float64(value))
			}
		}
	}
	if class&uint32(object.ClassWeapon) != 0 && subclass&0x82 != 0 && item.UseData.Ptr != nil {
		ammo := item.UseData.AsAmmo()
		if ammo.Field2 == 0 && ammo.Charge0 != 0 {
			key := "DefaultAmmoAmount"
			if quest {
				key = "DefaultAmmoAmountQuest"
			}
			defaultAmount := questInventoryRoundFloat32ToInt32_4F2C30(float32(s.Balance.Float(key)))
			if int32(ammo.Charge1) > defaultAmount {
				perShot := float32(float64(price) / float64(defaultAmount))
				price = float32(float64(int32(ammo.Charge1)-defaultAmount)*float64(perShot) + float64(price))
			} else {
				price = float32(float64(ammo.Charge1) / float64(defaultAmount) * float64(price))
			}
		}
	} else if class&uint32(object.ClassWand) != 0 && subclass&0x47f0000 != 0 && item.UseData.Ptr != nil {
		wand := item.UseData.AsWand()
		if wand.MaxCharge != 0 {
			price = float32(float64(wand.Charge) / float64(wand.MaxCharge) * float64(price))
		}
	}
	if health := item.HealthData; health != nil && health.Max != 0 {
		price = float32(float64(health.Cur) / float64(health.Max) * float64(price))
	}
	if !(price >= 1) {
		price = 1
	}
	return questInventoryRoundFloat32ToInt32_4F2C30(price)
}
