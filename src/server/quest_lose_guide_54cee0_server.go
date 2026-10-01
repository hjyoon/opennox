package server

type questLoseGuideNativeDeps54CEE0 struct {
	eligible   func(int32) int32
	randomInt  func(int32, int32) int32
	sendPacket func(uint8, [4]byte) int32
}

func questLoseGuideNative54CEE0(unit *Object, deps questLoseGuideNativeDeps54CEE0) {
	questLoseGuide54CEE0(unit, questLoseGuideHooks54CEE0[*Object, *PlayerUpdateData, *Player]{
		loadUpdate: func(unit *Object) *PlayerUpdateData { return (*PlayerUpdateData)(unit.UpdateData) },
		loadPlayer: func(update *PlayerUpdateData) *Player { return update.Player },
		loadClass:  func(player *Player) uint8 { return player.info[66] },
		loadLevel:  func(player *Player, id int32) uint32 { return player.BeastScrollLvl[id] },
		storeLevel: func(player *Player, id int32, level uint32) { player.BeastScrollLvl[id] = level },
		eligible:   deps.eligible,
		randomInt:  deps.randomInt,
		loadIndex:  func(player *Player) uint8 { return player.PlayerInd },
		sendPacket: deps.sendPacket,
	})
}

// QuestLoseGuide54CEE0 binds the Conjurer-only helper to native pointers,
// the original field-guide predicate, logic RNG and unsequenced F0/19 packet.
func (s *Server) QuestLoseGuide54CEE0(unit *Object) {
	questLoseGuideNative54CEE0(unit, questLoseGuideNativeDeps54CEE0{
		eligible: RandomFieldGuideLossEligible4F2530,
		randomInt: func(minimum, maximum int32) int32 {
			return int32(s.Rand.Logic.IntClamp(int(minimum), int(maximum)))
		},
		sendPacket: func(recipient uint8, packet [4]byte) int32 {
			return int32(s.NetSendPacketXxx0(int(recipient), packet[:], nil, 1))
		},
	})
}
