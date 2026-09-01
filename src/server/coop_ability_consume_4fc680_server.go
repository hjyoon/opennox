package server

const (
	coopAbilityModeFlag4FC680 = uint32(0x00000800)
	coopAbilityStopFlag4FC680 = uint32(0x00080000)
)

type coopAbilityConsumeNativeDeps4FC680 struct {
	gameFlag        func(uint32) int32
	firstPlayerUnit func() *Object
	executeAbility  func(*Object, int32)
}

// coopAbilityConsumeNative4FC680 binds the fixed-width queued state to the
// native-width Object pointer returned by the server player list.
func coopAbilityConsumeNative4FC680(s *Server, deps coopAbilityConsumeNativeDeps4FC680) {
	coopAbilityConsume4FC680(coopAbilityConsumeHooks4FC680[*Object]{
		loadCoopFlag: func() int32 {
			return deps.gameFlag(coopAbilityModeFlag4FC680)
		},
		loadFlag20: func() int32 {
			return deps.gameFlag(coopAbilityStopFlag4FC680)
		},
		loadState:       s.CoopAbilityState4FC670,
		firstPlayerUnit: deps.firstPlayerUnit,
		executeAbility:  deps.executeAbility,
		storeState: func(value int32) {
			s.coopAbilityState4FC670 = value
		},
	})
}

// CoopAbilityConsumeRuntime4FC680 supplies the two services owned by the
// outer game runtime: game-flag lookup and ability execution.
type CoopAbilityConsumeRuntime4FC680 struct {
	GameFlag       func(uint32) int32
	ExecuteAbility func(*Object, int32)
}

// CoopAbilityConsume4FC680 restores GAME.EXE 004FC680 using the native server
// state and first available PlayerUnit. No PE32 pointer slot crosses this
// boundary.
func (s *Server) CoopAbilityConsume4FC680(runtime CoopAbilityConsumeRuntime4FC680) {
	coopAbilityConsumeNative4FC680(s, coopAbilityConsumeNativeDeps4FC680{
		gameFlag:        runtime.GameFlag,
		firstPlayerUnit: s.Players.FirstUnit,
		executeAbility:  runtime.ExecuteAbility,
	})
}
