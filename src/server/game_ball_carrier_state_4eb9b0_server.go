package server

func gameBallCarrierStateNative4EB9B0(
	ball, target *Object,
	loadFrame func() uint32,
) *Object {
	return gameBallCarrierState4EB9B0(ball, target, gameBallCarrierStateHooks4EB9B0[
		*Object,
		*GameBallUpdateData4EA800,
	]{
		loadUpdateData: func(obj *Object) *GameBallUpdateData4EA800 {
			return (*GameBallUpdateData4EA800)(obj.UpdateData)
		},
		findPlayer: (*Object).FindOwnerChainPlayer,
		loadClassLow: func(obj *Object) uint8 {
			return uint8(obj.ObjClass)
		},
		storeCarrier: func(data *GameBallUpdateData4EA800, carrier *Object) {
			data.Carrier = carrier
		},
		loadTeamID: func(obj *Object) uint8 {
			return uint8(obj.TeamVal.ID)
		},
		storeTeamID: func(data *GameBallUpdateData4EA800, teamID uint32) {
			data.TeamID = teamID
		},
		loadFrame: loadFrame,
		storeFrame: func(data *GameBallUpdateData4EA800, frame uint32) {
			data.CarrierFrame = frame
		},
	})
}

// GameBallCarrierState4EB9B0 binds GAME.EXE 004EB9B0 to native-width Object
// pointers and the full GameBall update record. Frame is loaded only after a
// Player carrier and its live team byte have been stored.
func (s *Server) GameBallCarrierState4EB9B0(ball, target *Object) *Object {
	return gameBallCarrierStateNative4EB9B0(ball, target, s.Frame)
}
