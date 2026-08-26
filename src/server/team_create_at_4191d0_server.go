package server

// TeamCreateAtRuntime4191D0 supplies client/online effects that sit between
// the native team link and the original member-count publication point.
type TeamCreateAtRuntime4191D0 struct {
	ClientNetCode   uint32
	SelectLocalTeam func(TeamID)
	AfterAttach     func(*Team, *ObjectTeam, int32, uint32, int32, uint16)
}

// TeamCreateAtImpl4191D0 is the native-width binding for GAME.EXE 004191D0.
func (s *Server) TeamCreateAtImpl4191D0(
	teamID TeamID,
	value *ObjectTeam,
	active int32,
	netCode uint32,
	flags int32,
	runtime TeamCreateAtRuntime4191D0,
) {
	teamCreateAt4191D0(uint8(teamID), value, active, netCode, flags, teamCreateAtHooks4191D0[
		*ObjectTeam,
		*Object,
		*Team,
		*Player,
	]{
		loadGameBallType: func() uint16 {
			return uint16(s.Types.GameBallID())
		},
		findTeam: func(id uint8) *Team {
			return s.Teams.ByID(TeamID(id))
		},
		containsTeam: func(value *ObjectTeam, id uint8) bool {
			return s.Teams.ContainsObject(value, TeamID(id))
		},
		createTeam: func(id uint8) *Team {
			return s.Teams.Create(TeamID(id))
		},
		linkTeam: func(value *ObjectTeam, team *Team) bool {
			return s.Teams.LinkObject(value, team.ID()) != nil
		},
		loadTeamID: func(team *Team) uint8 {
			return uint8(team.ID())
		},
		clientNetCode: func() uint32 {
			return runtime.ClientNetCode
		},
		selectLocalTeam: func(id uint8) {
			if runtime.SelectLocalTeam != nil {
				runtime.SelectLocalTeam(TeamID(id))
			}
		},
		afterAttach: func(team *Team, value *ObjectTeam, active int32, netCode uint32, flags int32, gameBallType uint16) {
			if runtime.AfterAttach != nil {
				runtime.AfterAttach(team, value, active, netCode, flags, gameBallType)
			}
		},
		commitMemberCount: s.Teams.CommitObjectAttach,
		firstPlayerUnit:   s.Players.FirstUnit,
		nextPlayerUnit:    s.Players.NextUnit,
		loadNetCode: func(unit *Object) uint32 {
			return unit.NetCode
		},
		loadPlayer: func(unit *Object) *Player {
			data := unit.UpdateDataPlayer()
			if data == nil {
				return nil
			}
			return data.Player
		},
		loadPlayerIndex: func(player *Player) uint8 {
			return player.PlayerInd
		},
		resetPlayer: func(playerIndex uint8) {
			s.NetSendPacketXxx1(int(playerIndex), []byte{54}, nil, 1)
		},
		rebuildMasks: func(playerIndex uint8) {
			_ = s.RebuildObjectPlayerMasks4E8110(int32(playerIndex))
		},
		markUpdate: s.Nox_xxx_monsterMarkUpdate_4E8020,
		firstOwned: func(unit *Object) *Object {
			return unit.Field129
		},
		nextOwned: func(unit *Object) *Object {
			return unit.Field128
		},
		loadClassLow: func(unit *Object) uint8 {
			return uint8(unit.ObjClass)
		},
	})
}
