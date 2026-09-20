package server

import (
	"math"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

// GameBallUpdateRuntime53DF40 supplies services that remain owned by the
// root/legacy runtime. Object and update-data pointers never cross an integer
// ABI boundary.
type GameBallUpdateRuntime53DF40 struct {
	Ticks      func() uint64
	ResetBall  func(*Object) int
	ChangeTeam func(*ObjectTeam, uint32)
	BallStatus func(uint8, uint16) int32
	Move       func(*Object, types.Pointf)
	ApplyForce func(*Object, types.Pointf, float64)
}

type gameBallUpdateNativeDeps53DF40 struct {
	ticks        func() uint64
	resetBall    func(*Object)
	carrierState func(*Object, *Object)
	changeTeam   func(*Object)
	ballStatus   func(uint8, uint16)
	move         func(*Object, types.Pointf)
	applyForce   func(*Object, types.Pointf, float32)
	clearOwner   func(*Object)
	trace        func(types.Pointf, types.Pointf, MapTraceFlags) bool
	randomInt    func(int32, int32) int32
	audio        func(uint32, *Object)
	frame        func() uint32
}

func gameBallUpdateNative53DF40(ball *Object, deps gameBallUpdateNativeDeps53DF40) {
	gameBallUpdate53DF40(ball, gameBallUpdateHooks53DF40[*Object, *GameBallUpdateData4EA800]{
		loadUpdate: func(obj *Object) *GameBallUpdateData4EA800 {
			return (*GameBallUpdateData4EA800)(obj.UpdateData)
		},
		storeDrag: func(obj *Object, bits uint32) {
			obj.Float28 = math.Float32frombits(bits)
		},
		loadCarrier: func(update *GameBallUpdateData4EA800) *Object {
			return update.Carrier
		},
		loadFlagsLow: func(obj *Object) uint8 {
			return uint8(obj.ObjFlags)
		},
		carrierState: deps.carrierState,
		changeTeam:   deps.changeTeam,
		ballStatus:   deps.ballStatus,
		loadTicks:    deps.ticks,
		loadUpdateTicks: func(update *GameBallUpdateData4EA800) uint64 {
			return update.Ticks
		},
		resetBall: deps.resetBall,
		loadOwner: func(obj *Object) *Object {
			return obj.ObjOwner
		},
		loadFrame: deps.frame,
		loadCarrierFrame: func(update *GameBallUpdateData4EA800) uint32 {
			return update.CarrierFrame
		},
		loadPossessionDuration: func(update *GameBallUpdateData4EA800) uint32 {
			return update.PossessionDuration
		},
		loadFlags: func(obj *Object) uint32 {
			return uint32(obj.ObjFlags)
		},
		storeFlags: func(obj *Object, flags uint32) {
			obj.ObjFlags = object.Flags(flags)
		},
		storeUpdateTicks: func(update *GameBallUpdateData4EA800, ticks uint64) {
			update.Ticks = ticks
		},
		clearOwner: deps.clearOwner,
		loadRadius: func(obj *Object) float32 {
			return obj.Shape.Circle.R
		},
		loadPosX: func(obj *Object) float32 {
			return obj.PosVec.X
		},
		loadPosY: func(obj *Object) float32 {
			return obj.PosVec.Y
		},
		loadDirection: func(obj *Object) int16 {
			return int16(obj.Direction1)
		},
		loadDirectionCos: func(direction uint8) float32 {
			cosine, _ := SinCosDir(direction)
			return cosine
		},
		loadDirectionSin: func(direction uint8) float32 {
			_, sine := SinCosDir(direction)
			return sine
		},
		trace: func(from, to types.Pointf, flags uint8) bool {
			return deps.trace(from, to, MapTraceFlags(flags))
		},
		move: deps.move,
		storeObj130: func(obj, value *Object) {
			obj.Obj130 = value
		},
		randomInt:  deps.randomInt,
		applyForce: deps.applyForce,
		audio:      deps.audio,
		loadVelocityX: func(obj *Object) float32 {
			return obj.VelVec.X
		},
		loadVelocityY: func(obj *Object) float32 {
			return obj.VelVec.Y
		},
		loadResetVelocity: func(update *GameBallUpdateData4EA800) float32 {
			return update.ResetVelocity
		},
		storeCarrier: func(update *GameBallUpdateData4EA800, carrier *Object) {
			update.Carrier = carrier
		},
	})
}

func gameBallUpdateServerDeps53DF40(
	s *Server,
	runtime GameBallUpdateRuntime53DF40,
) gameBallUpdateNativeDeps53DF40 {
	return gameBallUpdateNativeDeps53DF40{
		ticks: runtime.Ticks,
		resetBall: func(ball *Object) {
			runtime.ResetBall(ball)
		},
		carrierState: func(ball, carrier *Object) {
			s.GameBallCarrierState4EB9B0(ball, carrier)
		},
		changeTeam: func(ball *Object) {
			runtime.ChangeTeam(ball.TeamPtr(), ball.NetCode)
		},
		ballStatus: func(state uint8, netCode uint16) {
			runtime.BallStatus(state, netCode)
		},
		move: runtime.Move,
		applyForce: func(ball *Object, origin types.Pointf, force float32) {
			runtime.ApplyForce(ball, origin, float64(force))
		},
		clearOwner: s.ObjClearOwner,
		trace:      s.MapTraceRay,
		randomInt: func(minimum, maximum int32) int32 {
			return int32(s.Rand.Logic.IntClamp(int(minimum), int(maximum)))
		},
		audio: func(id uint32, obj *Object) {
			s.Audio.EventObj(sound.ID(id), obj, 0, 0)
		},
		frame: s.Frame,
	}
}

// GameBallUpdate53DF40 executes the BallUpdate callback with native-width
// carrier, owner and auxiliary object fields on every architecture.
func (s *Server) GameBallUpdate53DF40(ball *Object, runtime GameBallUpdateRuntime53DF40) {
	gameBallUpdateNative53DF40(ball, gameBallUpdateServerDeps53DF40(s, runtime))
}
