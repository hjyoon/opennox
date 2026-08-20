package server

import noxflags "github.com/opennox/opennox/v1/common/flags"

// PlayerUnsetStatusRuntime417530 supplies the legacy session-timer services
// reached only when status bit one is removed. Poison uses status 0x400 and
// therefore does not enter this boundary.
type PlayerUnsetStatusRuntime417530 struct {
	AnyPlayers     func() int32
	TimerStatus    func() int32
	ModeEnabled    func(int16) uint8
	StartTimer     func()
	SetTimerStatus func(int32) int32
}

type playerStatusNativeDeps4174F0 struct {
	gameFlag       func(uint32) int32
	reportStatus   func(*Player) int32
	anyPlayers     func() int32
	timerStatus    func() int32
	gameFlagsValue func() uint32
	modeEnabled    func(int16) uint8
	startTimer     func()
	setTimerStatus func(int32) int32
}

func playerNeedStatusNative4174F0(player *Player, mask uint32, deps playerStatusNativeDeps4174F0) int32 {
	return playerNeedStatus4174F0(playerNeedStatusHooks4174F0[*Player]{
		loadPlayerArg: func() *Player {
			return player
		},
		loadMaskArg: func() uint32 {
			return mask
		},
		loadFlags: func(player *Player) uint32 {
			return player.Field3680
		},
		storeFlags: func(player *Player, flags uint32) {
			player.Field3680 = flags
		},
		gameFlag:     deps.gameFlag,
		reportStatus: deps.reportStatus,
	})
}

func playerUnsetStatusNative417530(player *Player, mask uint32, deps playerStatusNativeDeps4174F0) int32 {
	return playerUnsetStatus417530(playerUnsetStatusHooks417530[*Player]{
		loadMaskArg: func() uint32 {
			return mask
		},
		loadPlayerArg: func() *Player {
			return player
		},
		loadFlags: func(player *Player) uint32 {
			return player.Field3680
		},
		storeFlags: func(player *Player, flags uint32) {
			player.Field3680 = flags
		},
		gameFlag:       deps.gameFlag,
		anyPlayers:     deps.anyPlayers,
		timerStatus:    deps.timerStatus,
		gameFlagsValue: deps.gameFlagsValue,
		modeEnabled:    deps.modeEnabled,
		startTimer:     deps.startTimer,
		setTimerStatus: deps.setTimerStatus,
		reportStatus:   deps.reportStatus,
	})
}

func playerStatusReportNative417630(
	player *Player,
	send func(byte, []byte, int32) int32,
) int32 {
	return playerStatusReport417630(playerStatusReportHooks417630[*Player]{
		loadPlayerArg: func() *Player {
			return player
		},
		loadNetCode: func(player *Player) uint16 {
			return uint16(player.NetCodeVal)
		},
		loadFlags: func(player *Player) uint32 {
			return player.Field3680
		},
		send: send,
	})
}

func gameFlagNative4174F0(flag uint32) int32 {
	if noxflags.HasGame(noxflags.GameFlag(flag)) {
		return 1
	}
	return 0
}

func (s *Server) playerStatusNativeDeps4174F0(runtime PlayerUnsetStatusRuntime417530) playerStatusNativeDeps4174F0 {
	return playerStatusNativeDeps4174F0{
		gameFlag: gameFlagNative4174F0,
		reportStatus: func(player *Player) int32 {
			return s.NetReportPlayerStatus417630(player)
		},
		anyPlayers:     runtime.AnyPlayers,
		timerStatus:    runtime.TimerStatus,
		gameFlagsValue: func() uint32 { return uint32(noxflags.GetGame()) },
		modeEnabled:    runtime.ModeEnabled,
		startTimer:     runtime.StartTimer,
		setTimerStatus: runtime.SetTimerStatus,
	}
}

// NeedPlayerStatus4174F0 binds the original status setter to the native-width
// Player layout and the live server status reporter.
func (s *Server) NeedPlayerStatus4174F0(player *Player, mask uint32) int32 {
	return playerNeedStatusNative4174F0(player, mask, s.playerStatusNativeDeps4174F0(PlayerUnsetStatusRuntime417530{}))
}

// UnsetPlayerStatus417530 binds the original status remover to native Player
// fields while keeping its bit-one session services explicit.
func (s *Server) UnsetPlayerStatus417530(
	player *Player,
	mask uint32,
	runtime PlayerUnsetStatusRuntime417530,
) int32 {
	return playerUnsetStatusNative417530(player, mask, s.playerStatusNativeDeps4174F0(runtime))
}

// NetReportPlayerStatus417630 sends the original reliable seven-byte status
// packet to the broadcast recipient and preserves the send result.
func (s *Server) NetReportPlayerStatus417630(player *Player) int32 {
	return playerStatusReportNative417630(player, func(recipient byte, packet []byte, remove int32) int32 {
		return int32(s.NetSendPacketXxx1(int(recipient), packet, nil, int(remove)))
	})
}
