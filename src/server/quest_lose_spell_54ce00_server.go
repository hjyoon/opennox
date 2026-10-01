package server

type questLoseSpellNativeDeps54CE00 struct {
	eligible   func(int32) int32
	randomInt  func(int32, int32) int32
	sendPacket func(uint8, [4]byte) int32
}

func questLoseSpellNative54CE00(unit *Object, deps questLoseSpellNativeDeps54CE00) {
	questLoseSpell54CE00(unit, questLoseSpellHooks54CE00[*Object, *PlayerUpdateData, *Player]{
		loadUpdate: func(unit *Object) *PlayerUpdateData { return (*PlayerUpdateData)(unit.UpdateData) },
		loadPlayer: func(update *PlayerUpdateData) *Player { return update.Player },
		loadClass:  func(player *Player) uint8 { return player.info[66] },
		loadLevel:  func(player *Player, id int32) uint32 { return player.SpellLvl[id] },
		storeLevel: func(player *Player, id int32, level uint32) { player.SpellLvl[id] = level },
		eligible:   deps.eligible,
		randomInt:  deps.randomInt,
		loadIndex:  func(player *Player) uint8 { return player.PlayerInd },
		sendPacket: deps.sendPacket,
	})
}

// QuestLoseSpell54CE00 binds the Wizard/Conjurer helper to native pointers,
// the original loss eligibility, the logic RNG and the unsequenced F0/17
// packet. Empty candidate sets still call the original RNG API.
func (s *Server) QuestLoseSpell54CE00(unit *Object) {
	questLoseSpellNative54CE00(unit, questLoseSpellNativeDeps54CE00{
		eligible: RandomSpellLossEligible4F24E0,
		randomInt: func(minimum, maximum int32) int32 {
			return int32(s.Rand.Logic.IntClamp(int(minimum), int(maximum)))
		},
		sendPacket: func(recipient uint8, packet [4]byte) int32 {
			return int32(s.NetSendPacketXxx0(int(recipient), packet[:], nil, 1))
		},
	})
}
