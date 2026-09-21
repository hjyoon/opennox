package opennox

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

const (
	mushroomCleanMessage53ECE0   = "Use.c:MushroomClean"
	mushroomConfuseMessage53ECE0 = "Use.c:MushroomConfuse"
)

type mushroomUseHooks53ECE0[O, I comparable] struct {
	loadPoison      func(O) uint8
	removePoison    func(O)
	priorityMessage func(O, string)
	cureAudio       func(O)
	applyConfusion  func(O)
	delayedDelete   func(I)
}

// useMushroom53ECE0 preserves GAME.EXE 0053ECE0 independently of the native
// object layout. A mushroom clears existing poison, always applies confusion,
// and is consumed in either case.
func useMushroom53ECE0[O, I comparable](owner O, item I, hooks mushroomUseHooks53ECE0[O, I]) bool {
	if hooks.loadPoison(owner) != 0 {
		hooks.removePoison(owner)
		hooks.priorityMessage(owner, mushroomCleanMessage53ECE0)
		hooks.cureAudio(owner)
	} else {
		hooks.priorityMessage(owner, mushroomConfuseMessage53ECE0)
	}
	hooks.applyConfusion(owner)
	hooks.delayedDelete(item)
	return true
}

type curePoisonPotionHooks53EF70[O comparable] struct {
	loadPoison   func(O) uint8
	removePoison func(O)
	cureAudio    func(O)
}

// tryUseCurePoisonPotion53EF70 is the antidote branch of GAME.EXE 0053EF70.
// It deliberately leaves an unpoisoned player's potion untouched so it can be
// used later, matching the original item's consumption rules.
func tryUseCurePoisonPotion53EF70[O comparable](
	owner O,
	isCurePoisonPotion, isPlayer bool,
	hooks curePoisonPotionHooks53EF70[O],
) bool {
	if !isCurePoisonPotion || !isPlayer || hooks.loadPoison(owner) == 0 {
		return false
	}
	hooks.removePoison(owner)
	hooks.cureAudio(owner)
	return true
}

func nox_xxx_useConsume_53EE10(obj, item *server.Object) bool {
	s := noxServer
	if !item.Class().Has(object.ClassFood) {
		return true
	}
	if item.SubClass().AsFood().Has(object.FoodPotion) {
		return true
	}
	if obj.HealthData == nil || item.UseData.Ptr == nil {
		return true
	}
	if obj.HealthData.Cur >= obj.HealthData.Max {
		return true
	}
	use := item.UseDataConsume()
	dhp := int(use.Value)
	legacy.Nox_xxx_unitAdjustHP_4EE460(obj, dhp)
	if obj.Class().Has(object.ClassPlayer) {
		ud := obj.UpdateDataPlayer()
		if ud.Player.Info().IsFemale() {
			if item.SubClass().AsFood().Has(object.FoodSimple) {
				s.Audio.EventObj(sound.SoundHumanFemaleEatFood, obj, 0, 0)
			} else if item.SubClass().AsFood().Has(object.FoodApple) {
				s.Audio.EventObj(sound.SoundHumanFemaleEatApple, obj, 0, 0)
			} else if item.SubClass().AsFood().Has(object.FoodJug) {
				s.Audio.EventObj(sound.SoundHumanFemaleDrinkJug, obj, 0, 0)
			}
		} else {
			if item.SubClass().AsFood().Has(object.FoodSimple) {
				s.Audio.EventObj(sound.SoundHumanMaleEatFood, obj, 0, 0)
			} else if item.SubClass().AsFood().Has(object.FoodApple) {
				s.Audio.EventObj(sound.SoundHumanMaleEatApple, obj, 0, 0)
			} else if item.SubClass().AsFood().Has(object.FoodJug) {
				s.Audio.EventObj(sound.SoundHumanMaleDrinkJug, obj, 0, 0)
			}
		}
	} else {
		s.Audio.EventObj(sound.SoundMonsterEatFood, obj, 0, 0)
	}
	s.DelayedDelete(item)
	return true
}

func nox_xxx_useMushroom_53ECE0(obj, item *server.Object) bool {
	s := noxServer
	return useMushroom53ECE0(obj, item, mushroomUseHooks53ECE0[*server.Object, *server.Object]{
		loadPoison: func(owner *server.Object) uint8 {
			return owner.Poison540
		},
		removePoison: func(owner *server.Object) {
			legacy.Nox_xxx_removePoison_4EE9D0(owner)
		},
		priorityMessage: func(owner *server.Object, message string) {
			s.NetPriMsgToPlayer(owner, strman.ID(message), 0)
		},
		cureAudio: func(owner *server.Object) {
			aud := s.Spells.DefByInd(spell.SPELL_CURE_POISON).GetOnSound()
			s.Audio.EventObj(aud, owner, 0, 0)
		},
		applyConfusion: func(owner *server.Object) {
			legacy.Nox_xxx_buffApplyTo_4FF380(owner, server.ENCHANT_CONFUSED, int(s.SecToFrames(10)), 5)
		},
		delayedDelete: s.DelayedDelete,
	})
}

func nox_xxx_useCiderConfuse_53EF00(obj, item *server.Object) bool {
	s := noxServer
	if obj == nil || item == nil || obj.HealthData == nil {
		return true
	}
	legacy.Nox_xxx_buffApplyTo_4FF380(obj, server.ENCHANT_CONFUSED, int(s.SecToFrames(5)), 4)
	s.NetPriMsgToPlayer(obj, "Use.c:CiderConfuse", 0)
	ok := nox_xxx_useConsume_53EE10(obj, item)
	if ok {
		s.DelayedDelete(item)
	}
	return ok
}

func nox_xxx_useEnchant_53ED60(obj, item *server.Object) bool {
	s := noxServer
	if use := item.UseDataEnchant(); use != nil {
		ench := server.EnchantID(use.Enchant)
		dur := int(use.Dur)
		legacy.Nox_xxx_buffApplyTo_4FF380(obj, ench, dur, 5)
	}
	s.DelayedDelete(item)
	return true
}

func nox_xxx_useCast_53ED90(obj, item *server.Object) bool {
	s := noxServer
	if use := item.UseDataCast(); use != nil {
		var pos types.Pointf
		if obj.Class().Has(object.ClassPlayer) {
			ud := obj.UpdateDataPlayer()
			pos = types.Point2f(ud.Player.CursorVec)
		} else {
			pos = obj.PosVec
		}
		s.Nox_xxx_spellAccept4FD400(spell.ID(use.Spell), item, item, item, &server.SpellAcceptArg{
			Obj: obj, Pos: pos,
		}, 4)
	}
	s.DelayedDelete(item)
	return true
}

func nox_xxx_usePotion_53EF70(obj, potion *server.Object) bool {
	s := noxServer
	if obj.Class().Has(object.ClassPlayer) && obj.Flags().Has(object.FlagDead) {
		return false
	}
	consumed := false
	use := potion.UseDataPotion()
	if use != nil && potion.SubClass().AsFood().Has(object.FoodHealthPotion) && obj.HealthData != nil && obj.HealthData.Cur < obj.HealthData.Max {
		dhp := int(use.Value)
		if obj.Class().Has(object.ClassPlayer) {
			ud := obj.UpdateDataPlayer()
			if mult := s.Players.ClassStatsMult(ud.Player.PlayerClass()); mult != nil {
				dhp = int(float64(dhp) * float64(mult.Health))
			}
		}
		legacy.Nox_xxx_unitAdjustHP_4EE460(obj, dhp)
		s.Audio.EventObj(sound.SoundRestoreHealth, obj, 0, 0)
		consumed = true
	}
	if use != nil && potion.SubClass().AsFood().Has(object.FoodManaPotion) && obj.Class().Has(object.ClassPlayer) {
		ud := obj.UpdateDataPlayer()
		if ud.ManaCur < ud.ManaMax {
			dmp := int(use.Value)
			if mult := s.Players.ClassStatsMult(ud.Player.PlayerClass()); mult != nil {
				dmp = int(float64(dmp) * float64(mult.Mana))
			}
			legacy.Nox_xxx_playerManaAdd_4EEB80(obj, dmp)
			s.Audio.EventObj(sound.SoundRestoreMana, obj, 0, 0)
			consumed = true
		}
	}
	if tryUseCurePoisonPotion53EF70(
		obj,
		potion.SubClass().AsFood().Has(object.FoodCurePoisonPotion),
		obj.Class().Has(object.ClassPlayer),
		curePoisonPotionHooks53EF70[*server.Object]{
			loadPoison: func(owner *server.Object) uint8 {
				return owner.Poison540
			},
			removePoison: func(owner *server.Object) {
				legacy.Nox_xxx_removePoison_4EE9D0(owner)
			},
			cureAudio: func(owner *server.Object) {
				aud := s.Spells.DefByInd(spell.SPELL_CURE_POISON).GetOnSound()
				s.Audio.EventObj(aud, owner, 0, 0)
			},
		},
	) {
		consumed = true
	}
	for _, t := range []struct {
		SubClass object.FoodClass
		Enchant  server.EnchantID
	}{
		{object.FoodHastePotion, server.ENCHANT_HASTED},
		{object.FoodInvisibilityPotion, server.ENCHANT_INVISIBLE},
		{object.FoodFireProtectPotion, server.ENCHANT_PROTECT_FROM_FIRE},
		{object.FoodShockProtectPotion, server.ENCHANT_PROTECT_FROM_ELECTRICITY},
		{object.FoodPoisonProtectPotion, server.ENCHANT_PROTECT_FROM_POISON},
		{object.FoodInvulnerabilityPotion, server.ENCHANT_INVULNERABLE},
		{object.FoodInfravisionPotion, server.ENCHANT_INFRAVISION},
		{object.FoodVampirismPotion, server.ENCHANT_VAMPIRISM},
	} {
		if potion.SubClass().AsFood().Has(t.SubClass) {
			legacy.Nox_xxx_buffApplyTo_4FF380(obj, t.Enchant, int(s.SecToFrames(120)), 3)
			consumed = true
		}
	}
	if use != nil && potion.SubClass().AsFood().Has(object.FoodShieldPotion) {
		lvl := int(use.Value)
		s.Nox_xxx_spellAccept4FD400(spell.SPELL_SHIELD, obj, obj, obj, &server.SpellAcceptArg{
			Obj: obj,
			Pos: obj.PosVec,
		}, lvl)
		consumed = true
	}
	if consumed {
		s.DelayedDelete(potion)
	}
	return true
}
