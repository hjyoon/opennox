package server

type monsterMarkUpdateNativeDeps4E8020 struct {
	firstPlayer func() *Player
	nextPlayer  func(*Player) *Player
	isHostile   func(*Object, *Object) int32
}

func monsterMarkUpdateNative4E8020(
	obj *Object,
	deps monsterMarkUpdateNativeDeps4E8020,
) {
	monsterMarkUpdate4E8020(obj, monsterMarkUpdateHooks4E8020[*Object, *Player]{
		firstPlayer: deps.firstPlayer,
		loadPlayerInd: func(player *Player) uint8 {
			return player.PlayerInd
		},
		loadPlayerUnit: func(player *Player) *Object {
			return player.PlayerUnit
		},
		isHostile: deps.isHostile,
		loadField36: func(obj *Object) uint32 {
			return obj.Field36
		},
		loadField35: func(obj *Object) uint32 {
			return obj.Field35
		},
		storeField36: func(obj *Object, value uint32) {
			obj.Field36 = value
		},
		storeField35: func(obj *Object, value uint32) {
			obj.Field35 = value
		},
		nextPlayer: deps.nextPlayer,
	})
}

// Nox_xxx_monsterMarkUpdate_4E8020 is the faithful native-width entry point.
// In particular, it obtains the player-list head before reading obj.
func (s *Server) Nox_xxx_monsterMarkUpdate_4E8020(obj *Object) {
	monsterMarkUpdateNative4E8020(obj, monsterMarkUpdateNativeDeps4E8020{
		firstPlayer: s.Players.First,
		nextPlayer:  s.Players.Next,
		isHostile:   s.isHostileMimicResult4E7F90,
	})
}

// Nox_xxx_monsterMarkUpdate_4E8020 retains the historical object method for
// native Go callers. The Server entry point above is used at the CGo boundary.
func (obj *Object) Nox_xxx_monsterMarkUpdate_4E8020() {
	obj.Server().Nox_xxx_monsterMarkUpdate_4E8020(obj)
}
