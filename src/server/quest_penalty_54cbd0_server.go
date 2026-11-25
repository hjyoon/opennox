package server

// QuestPenaltyRuntime54CBD0 supplies the normal legacy gold protection,
// gem-cache/price, weapon-class and delayed-deletion services. All are
// required at their original call sites, including zero-gold subtraction.
type QuestPenaltyRuntime54CBD0 struct {
	SubGold                         func(*Object, uint32)
	LoseGems, LoseWeapon, LoseArmor func(*Object)
}

type questPenaltyNativeDeps54CBD0 struct {
	getGold                         func(*Object) uint32
	subGold                         func(*Object, uint32)
	loseGems, loseWeapon, loseArmor func(*Object)
	loseSpell, loseGuide            func(*Object)
	loseAbility                     func(*Object) int8
}

func questPenaltyNative54CBD0(unit *Object, deps questPenaltyNativeDeps54CBD0) {
	questPenalty54CBD0(unit, questPenaltyHooks54CBD0[*Object, *PlayerUpdateData, *Player]{
		loadUpdate:  func(unit *Object) *PlayerUpdateData { return (*PlayerUpdateData)(unit.UpdateData) },
		loadPlayer:  func(update *PlayerUpdateData) *Player { return update.Player },
		loadClass:   func(player *Player) uint8 { return player.info[66] },
		getGold:     deps.getGold,
		subGold:     deps.subGold,
		loseGems:    deps.loseGems,
		loseWeapon:  deps.loseWeapon,
		loseArmor:   deps.loseArmor,
		loseSpell:   deps.loseSpell,
		loseGuide:   deps.loseGuide,
		loseAbility: deps.loseAbility,
	})
}

func questPenaltyRequiredLoss(name string, lose func(*Object)) func(*Object) {
	if lose != nil {
		return lose
	}
	return func(*Object) { panic("missing Quest penalty " + name + " binding") }
}

// QuestPenalty54CBD0 binds the original loss dispatcher to the native loss
// services. A missing service faults at its call site, after preceding losses.
func (s *Server) QuestPenalty54CBD0(unit *Object, runtime QuestPenaltyRuntime54CBD0) {
	subGold := runtime.SubGold
	if subGold == nil {
		subGold = func(*Object, uint32) { panic("missing Quest penalty gold subtraction binding") }
	}
	questPenaltyNative54CBD0(unit, questPenaltyNativeDeps54CBD0{
		getGold: func(unit *Object) uint32 {
			return (*PlayerUpdateData)(unit.UpdateData).Player.GoldVal
		},
		subGold:     subGold,
		loseGems:    questPenaltyRequiredLoss("gem loss", runtime.LoseGems),
		loseWeapon:  questPenaltyRequiredLoss("weapon loss", runtime.LoseWeapon),
		loseArmor:   questPenaltyRequiredLoss("armor loss", runtime.LoseArmor),
		loseSpell:   s.QuestLoseSpell54CE00,
		loseGuide:   s.QuestLoseGuide54CEE0,
		loseAbility: s.QuestLoseAbility54CFB0,
	})
}
