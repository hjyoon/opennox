package server

type teamAutoAssignHooks4181F0[P, O, T comparable] struct {
	firstPlayer          func() P
	nextPlayer           func(P) P
	loadUnit             func(P) O
	loadNetCode          func(P) uint32
	loadStatus           func(P) uint32
	hasTeam              func(O) bool
	clientNetCode        func() uint32
	noRendering          func() bool
	randomInt            func(int, int) int
	preferConfiguredTeam func() bool
	configuredTeam       func(P) T
	firstTeam            func() T
	nextTeam             func(T) T
	teamMemberCount      func(T) int
	teamID               func(T) TeamID
	attach               func(TeamID, O, uint32, int32)
}

// teamAutoAssign4181F0 preserves the player selection, shuffle, and team
// choice performed by GAME.EXE 004181F0. Players and teams remain typed so
// the original PE32 scratch array cannot truncate their addresses.
func teamAutoAssign4181F0[P, O, T comparable](hooks teamAutoAssignHooks4181F0[P, O, T]) {
	var players [32]P
	count := 0
	var zeroPlayer P
	var zeroObject O
	var zeroTeam T
	for player := hooks.firstPlayer(); player != zeroPlayer; player = hooks.nextPlayer(player) {
		unit := hooks.loadUnit(player)
		if unit == zeroObject {
			continue
		}
		if hooks.loadNetCode(player) == hooks.clientNetCode() && hooks.noRendering() {
			continue
		}
		status := hooks.loadStatus(player)
		if (status&1 != 0 && status&0x20 == 0) || hooks.hasTeam(unit) {
			continue
		}
		if count == len(players) {
			break
		}
		players[count] = player
		count++
	}
	if count == 0 {
		return
	}
	if count > 1 {
		for i := 0; i < 50; i++ {
			first := hooks.randomInt(0, count-1)
			second := hooks.randomInt(0, count-1)
			players[first], players[second] = players[second], players[first]
		}
	}

	for _, player := range players[:count] {
		unit := hooks.loadUnit(player)
		var team T
		flags := int32(1)
		if hooks.preferConfiguredTeam() {
			team = hooks.configuredTeam(player)
			flags = 0
		} else {
			minimum := 32
			for candidate := hooks.firstTeam(); candidate != zeroTeam; candidate = hooks.nextTeam(candidate) {
				members := hooks.teamMemberCount(candidate)
				if members < minimum {
					minimum = members
					team = candidate
				}
			}
		}
		if team != zeroTeam {
			hooks.attach(hooks.teamID(team), unit, hooks.loadNetCode(player), flags)
		}
	}
}

// TeamAutoAssignRuntime4181F0 supplies the client and network operations that
// remain outside the server package.
type TeamAutoAssignRuntime4181F0 struct {
	ClientNetCode        func() uint32
	NoRendering          func() bool
	PreferConfiguredTeam func() bool
	Attach               func(TeamID, *ObjectTeam, int32, uint32, int32)
}

// TeamAutoAssign4181F0 assigns every eligible teamless player without routing
// Object, Player, or Team pointers through the legacy integer ABI.
func (s *Server) TeamAutoAssign4181F0(runtime TeamAutoAssignRuntime4181F0) {
	teamAutoAssign4181F0(teamAutoAssignHooks4181F0[*Player, *Object, *Team]{
		firstPlayer: s.Players.First,
		nextPlayer:  s.Players.Next,
		loadUnit: func(player *Player) *Object {
			return player.PlayerUnit
		},
		loadNetCode: func(player *Player) uint32 {
			return player.NetCodeVal
		},
		loadStatus: func(player *Player) uint32 {
			return player.Field3680
		},
		hasTeam: func(unit *Object) bool {
			return unit.TeamVal.Has()
		},
		clientNetCode:        runtime.ClientNetCode,
		noRendering:          runtime.NoRendering,
		randomInt:            s.Rand.Logic.IntClamp,
		preferConfiguredTeam: runtime.PreferConfiguredTeam,
		configuredTeam: func(player *Player) *Team {
			return s.Teams.ByXxx(int(player.Field2068))
		},
		firstTeam:       s.Teams.First,
		nextTeam:        s.Teams.Next,
		teamMemberCount: s.Teams.MemberCount,
		teamID: func(team *Team) TeamID {
			return team.ID()
		},
		attach: func(teamID TeamID, unit *Object, netCode uint32, flags int32) {
			runtime.Attach(teamID, &unit.TeamVal, 1, netCode, flags)
		},
	})
}
