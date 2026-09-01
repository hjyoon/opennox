package server

type mapTransitionPlayerInitNativeDeps4FC6D0 struct {
	runtime         MapTransitionPlayerInitRuntime4FC6D0
	firstPlayerUnit func() *Object
	nextPlayerUnit  func(*Object) *Object
}

// mapTransitionPlayerInitNative4FC6D0 binds the original fixed-width player
// fields to native Go pointers. UpdateData, Player, and PlayerInd remain live
// loads at every point where GAME.EXE reloaded them.
func mapTransitionPlayerInitNative4FC6D0(s *Server, deps mapTransitionPlayerInitNativeDeps4FC6D0) {
	runtime := deps.runtime
	mapTransitionPlayerInit4FC6D0(mapTransitionPlayerInitHooks4FC6D0[*Object, *PlayerUpdateData, *Player]{
		loadMapInitState:  s.MapInitState4FC570,
		loadMapEntryState: s.MapEntryState4FC580,
		firstPlayerUnit:   deps.firstPlayerUnit,
		nextPlayerUnit:    deps.nextPlayerUnit,
		hasGame: func(mask int32) int32 {
			return runtime.GameFlag(uint32(mask))
		},

		loadQuestStage:        runtime.QuestStage,
		loadRestorePredicate:  runtime.RestorePredicate,
		loadRestoreReady:      runtime.RestoreReady,
		loadQueuedRestore:     runtime.QueuedRestore,
		sendQuestStage:        func(index int32) { runtime.SendQuestStage(uint8(index)) },
		sendQuestRestore:      func(index, value int32) { runtime.SendQuestRestore(uint8(index), value) },
		storeQueuedRestore:    runtime.StoreQueuedRestore,
		markQuestReady:        runtime.MarkQuestReady,
		finishQuestTransition: runtime.FinishQuestTransition,
		fadeBegin:             runtime.FadeBegin,

		loadDataRoot:       runtime.DataRoot,
		formatTempSavePath: runtime.FormatTempSavePath,
		loadDeleteTempFile: func() func(string) { return runtime.DeleteTempFile },

		loadUpdateData: func(unit *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(unit.UpdateData)
		},
		loadPlayer: func(data *PlayerUpdateData) *Player {
			return data.Player
		},
		loadPlayerField4792: func(player *Player) int32 {
			return int32(player.Field4792)
		},
		loadUpdateField138: func(data *PlayerUpdateData) int32 {
			return int32(data.Field138)
		},
		loadPlayerIndex: func(player *Player) uint8 {
			return player.PlayerInd
		},
		loadPlayerField3680: func(player *Player) uint8 {
			return uint8(player.Field3680)
		},

		savePlayerData:    runtime.SavePlayerData,
		preparePlayerData: runtime.PreparePlayerData,
		sendGauntlet:      runtime.SendGauntlet,
		restorePlayerData: runtime.RestorePlayerData,
		finishPlayerData:  runtime.FinishPlayerData,
		applyEnchant: func(unit *Object, enchant, duration, power int32) {
			runtime.ApplyEnchant(unit, EnchantID(enchant), duration, power)
		},
	})
}

// MapTransitionPlayerInitRuntime4FC6D0 supplies services that remain owned by
// the outer game runtime. All PE32 scalar arguments use explicit widths; only
// native Object pointers cross this Go-only boundary.
type MapTransitionPlayerInitRuntime4FC6D0 struct {
	GameFlag func(uint32) int32

	QuestStage            func() int32
	RestorePredicate      func() int32
	RestoreReady          func() int32
	QueuedRestore         func() int32
	SendQuestStage        func(uint8)
	SendQuestRestore      func(uint8, int32)
	StoreQueuedRestore    func(int32)
	MarkQuestReady        func(int32)
	FinishQuestTransition func()
	FadeBegin             func(int32, int32)

	DataRoot           func() string
	FormatTempSavePath func(string) string
	DeleteTempFile     func(string)

	SavePlayerData    func(string, uint8) int32
	PreparePlayerData func(uint8) int32
	SendGauntlet      func(uint8, int32)
	RestorePlayerData func(string, uint8) int32
	FinishPlayerData  func(uint8)
	ApplyEnchant      func(*Object, EnchantID, int32, int32)
}

// MapTransitionPlayerInit4FC6D0 restores GAME.EXE 004FC6D0 against the native
// server state and live player-unit list. No legacy PE32 pointer slot or C
// function pointer crosses this boundary.
func (s *Server) MapTransitionPlayerInit4FC6D0(runtime MapTransitionPlayerInitRuntime4FC6D0) {
	mapTransitionPlayerInitNative4FC6D0(s, mapTransitionPlayerInitNativeDeps4FC6D0{
		runtime:         runtime,
		firstPlayerUnit: s.Players.FirstUnit,
		nextPlayerUnit:  s.Players.NextUnit,
	})
}
