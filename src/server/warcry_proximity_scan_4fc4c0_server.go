package server

type warcryProximityScanNativeDeps4FC4C0 struct {
	firstPlayer     func() *Player
	nextPlayer      func(*Player) *Player
	isAbilityActive func(*Object, Ability) int32
	mapCheck        func(*Object, *Object) int32
}

// warcryProximityScanNative4FC4C0 binds the executable's Player and Object
// field observations to native-width Go pointers. It intentionally does not
// synthesize the PE32 pointer slots used by GAME.EXE.
func warcryProximityScanNative4FC4C0(
	target *Object,
	deps warcryProximityScanNativeDeps4FC4C0,
) int32 {
	return warcryProximityScan4FC4C0(warcryProximityScanHooks4FC4C0[*Player, *Object]{
		firstPlayer: deps.firstPlayer,
		loadTargetArg: func() *Object {
			return target
		},
		loadPlayerUnit: func(player *Player) *Object {
			return player.PlayerUnit
		},
		loadPlayerClass: func(player *Player) uint8 {
			return uint8(player.Info().PlayerClass())
		},
		isAbilityActive: deps.isAbilityActive,
		loadPosX: func(object *Object) float32 {
			return object.PosVec.X
		},
		loadPosY: func(object *Object) float32 {
			return object.PosVec.Y
		},
		mapCheck:   deps.mapCheck,
		nextPlayer: deps.nextPlayer,
	})
}

// WarcryProximityScan4FC4C0 restores the deleted GAME.EXE 004FC4C0 helper.
// It scans active Player records for a Warrior sustaining Warcry within the
// original radius and with an unobstructed vision trace to target. No decoded
// executable call, jump, or stored entrypoint references this helper, so the
// restoration remains a native Go method rather than adding a speculative C
// ABI export.
//
//go:noinline
func (a *serverAbilities) WarcryProximityScan4FC4C0(target *Object) int32 {
	return warcryProximityScanNative4FC4C0(target, warcryProximityScanNativeDeps4FC4C0{
		firstPlayer: a.s.Players.First,
		nextPlayer:  a.s.Players.Next,
		isAbilityActive: func(unit *Object, ability Ability) int32 {
			if a.IsActive(unit, ability) {
				return 1
			}
			return 0
		},
		mapCheck: func(unit, target *Object) int32 {
			if a.s.MapTraceVision(unit, target) {
				return 1
			}
			return 0
		},
	})
}
