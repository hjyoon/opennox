package opennox

import (
	"math"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

// GameBallReset417F50 recreates the FlagBall at a randomly selected
// GameBallStart marker without routing object or player pointers through the
// ABI32 legacy implementation.
func (s *Server) GameBallReset417F50(old *server.Object) int {
	srv := s.S()
	markerType := uint16(srv.Types.IndByID("GameBallStart"))
	count := 0
	for obj := srv.Objs.First(); obj != nil; obj = obj.Next() {
		if obj.TypeInd == markerType {
			count++
		}
	}
	if count == 0 {
		return 0
	}

	selected := srv.Rand.Logic.IntClamp(0, count-1)
	var marker *server.Object
	for obj := srv.Objs.First(); obj != nil; obj = obj.Next() {
		if obj.TypeInd != markerType {
			continue
		}
		if selected == 0 {
			marker = obj
			break
		}
		selected--
	}
	if marker == nil {
		return 0
	}

	ball := srv.NewObjectByTypeID("GameBall")
	if ball == nil {
		return 0
	}
	update := (*server.GameBallUpdateData4EA800)(ball.UpdateData)
	update.Ticks = legacy.PlatformTicks()
	update.PossessionDuration = uint32(int32(math.RoundToEven(float64(float32(
		srv.Balance.Float("FlagballPossDuration"),
	)))))
	update.ResetVelocity = float32(srv.Balance.Float("FlagballResetVel"))

	for player := srv.Players.First(); player != nil; player = srv.Players.Next(player) {
		srv.Players.Nox_xxx_netMarkMinimapObject_417190(player.PlayerIndex(), ball, 1)
	}
	s.CreateObjectAt(ball, nil, types.Pointf{})
	srv.ObjClearOwner(ball)
	srv.GameBallCarrierState4EB9B0(ball, nil)
	legacy.Sub_4E8290(0, 0)
	legacy.Nox_xxx_unitMove_4E7010(ball, marker.PosVec)
	ball.VelVec = types.Pointf{}
	ball.ForceVec.X = 0
	ball.Pos24.Y = 0

	if old != nil {
		for player := srv.Players.First(); player != nil; player = srv.Players.Next(player) {
			if player.CameraTarget() == old {
				player.CameraUnlock()
				player.CameraFollow(ball)
			}
		}
		s.DelayedDelete(old)
	}
	return 1
}
