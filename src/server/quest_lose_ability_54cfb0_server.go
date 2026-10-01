package server

type questLoseAbilityNativeDeps54CFB0 struct {
	eligible   func(int32) int32
	randomInt  func(int32, int32) int32
	sendPacket func(uint8, [4]byte) int32
}

func questLoseAbilityNative54CFB0(unit *Object, deps questLoseAbilityNativeDeps54CFB0) int8 {
	return questLoseAbility54CFB0(unit, questLoseAbilityHooks54CFB0[*Object, *PlayerUpdateData, *Player]{
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

// QuestLoseAbility54CFB0 binds the original Warrior loss helper to native
// pointers, the logic RNG, and the four-byte unsequenced F0/18 packet.
// No class, nil-binding, or zero-candidate guards are added to the original.
func (s *Server) QuestLoseAbility54CFB0(unit *Object) int8 {
	return questLoseAbilityNative54CFB0(unit, questLoseAbilityNativeDeps54CFB0{
		eligible: RandomAbilityLossEligible4F2570,
		randomInt: func(minimum, maximum int32) int32 {
			return int32(s.Rand.Logic.IntClamp(int(minimum), int(maximum)))
		},
		sendPacket: func(recipient uint8, packet [4]byte) int32 {
			return int32(s.NetSendPacketXxx0(int(recipient), packet[:], nil, 1))
		},
	})
}
