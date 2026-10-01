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

// QuestPenalty54CBD0 binds one original dispatcher body to native-width
// pointers and the existing loss helpers. It does not admit the separate
// Quest lives/death branch or add a Player/class/nil guard to this body.
func (s *Server) QuestPenalty54CBD0(unit *Object, runtime QuestPenaltyRuntime54CBD0) {
	questPenaltyNative54CBD0(unit, questPenaltyNativeDeps54CBD0{
		getGold: func(unit *Object) uint32 {
			return (*PlayerUpdateData)(unit.UpdateData).Player.GoldVal
		},
		subGold:     runtime.SubGold,
		loseGems:    runtime.LoseGems,
		loseWeapon:  runtime.LoseWeapon,
		loseArmor:   runtime.LoseArmor,
		loseSpell:   s.QuestLoseSpell54CE00,
		loseGuide:   s.QuestLoseGuide54CEE0,
		loseAbility: s.QuestLoseAbility54CFB0,
	})
}
