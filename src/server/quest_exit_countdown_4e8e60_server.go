package server

type questExitCountdownNativeDeps4E8E60 struct {
	balanceFloat         func(string) float64
	timerActive          func() int32
	timerRemainingMillis func() int32
	firstUnit            func() *Object
	nextUnit             func(*Object) *Object
	stopTimer            func(int32) int32
	countdownStarted     func() int32
	startCountdown       func(int32, string)
	sendGauntlet         func(int32) int32
}

// QuestExitCountdownRuntime4E8E60 supplies the timer operations whose storage
// remains in the legacy runtime.
type QuestExitCountdownRuntime4E8E60 struct {
	TimerActive          func() int32
	TimerRemainingMillis func() int32
	StopTimer            func(int32) int32
	CountdownStarted     func() int32
	StartCountdown       func(int32, string)
}

func questExitCountdownNative4E8E60(deps questExitCountdownNativeDeps4E8E60) int32 {
	return questExitCountdown4E8E60(questExitCountdownHooks4E8E60[*Object, *PlayerUpdateData, *Player]{
		balanceFloat:         deps.balanceFloat,
		floatToInt:           questExitRound4E8E60,
		timerActive:          deps.timerActive,
		timerRemainingMillis: deps.timerRemainingMillis,
		firstUnit:            deps.firstUnit,
		nextUnit:             deps.nextUnit,
		loadUpdateData: func(obj *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(obj.UpdateData)
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadQuestState: func(player *Player) uint32 {
			return player.Field4792
		},
		loadQuestExit: func(update *PlayerUpdateData) *Object {
			return update.QuestExit
		},
		stopTimer:        deps.stopTimer,
		countdownStarted: deps.countdownStarted,
		startCountdown:   deps.startCountdown,
		sendGauntlet:     deps.sendGauntlet,
	})
}

func (s *Server) questExitSendGauntlet4E8E60(recipient int32) int32 {
	packet := [2]byte{0xf0, 0x14}
	return int32(s.NetSendPacketXxx0(int(recipient), packet[:], nil, 1))
}

// QuestExitCountdown4E8E60 binds the restored countdown contract to native
// Object and Player pointers while retaining the legacy timer boundary.
func (s *Server) QuestExitCountdown4E8E60(runtime QuestExitCountdownRuntime4E8E60) int32 {
	return questExitCountdownNative4E8E60(questExitCountdownNativeDeps4E8E60{
		balanceFloat:         s.Balance.Float,
		timerActive:          runtime.TimerActive,
		timerRemainingMillis: runtime.TimerRemainingMillis,
		firstUnit:            s.Players.FirstUnit,
		nextUnit:             s.questNextPlayerUnit4DA7F0,
		stopTimer:            runtime.StopTimer,
		countdownStarted:     runtime.CountdownStarted,
		startCountdown:       runtime.StartCountdown,
		sendGauntlet:         s.questExitSendGauntlet4E8E60,
	})
}
