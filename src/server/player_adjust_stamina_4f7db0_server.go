package server

type playerAdjustStaminaNativeDeps4F7DB0 struct {
	reportStamina func(uint8, *Object)
}

func playerAdjustStaminaNative4F7DB0(
	unit *Object,
	amount uint8,
	deps playerAdjustStaminaNativeDeps4F7DB0,
) {
	playerAdjustStamina4F7DB0(unit, playerAdjustStaminaHooks4F7DB0[
		*Object,
		*PlayerUpdateData,
		*Player,
	]{
		loadClass: func(unit *Object) uint32 {
			return uint32(unit.ObjClass)
		},
		loadUpdate: func(unit *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(unit.UpdateData)
		},
		loadAmount: func() uint8 {
			return amount
		},
		loadStamina: func(update *PlayerUpdateData) uint8 {
			return update.Stamina
		},
		storeStamina: func(update *PlayerUpdateData, stamina uint8) {
			update.Stamina = stamina
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadPlayerIndex: func(player *Player) uint8 {
			return player.PlayerInd
		},
		reportStamina: deps.reportStamina,
	})
}

// PlayerAdjustStamina4F7DB0 binds GAME.EXE 004F7DB0 to native-width Object,
// PlayerUpdateData, and Player pointers while preserving low-byte arithmetic.
func (s *Server) PlayerAdjustStamina4F7DB0(unit *Object, amount uint8) {
	playerAdjustStaminaNative4F7DB0(unit, amount, playerAdjustStaminaNativeDeps4F7DB0{
		reportStamina: playerStaminaReportServer4D8800(s),
	})
}
