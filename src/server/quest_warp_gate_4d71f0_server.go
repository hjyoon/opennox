package server

import (
	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

const (
	questExitTimeoutFrames4D71F0 = uint32(0x2328)
	questWarpDelayFrames4D7600   = uint32(30)
	questPlayerClassByte4D7480   = uint8(0x04)
	questExitClassByte4D7520     = uint8(0x20)
	questExitSubclassByte4D7520  = uint8(0x02)
)

// QuestWarpGateRuntime4D71F0 supplies the Quest state and integration calls
// that remain owned by the legacy runtime. Object, update-data, player, and
// exit pointers stay native-width throughout all four restored routines.
type QuestWarpGateRuntime4D71F0 struct {
	ExitCountdownStart      func() uint32
	StoreExitCountdownStart func(uint32)
	WarpEnabled             func() uint32
	StoreWarpEnabled        func(uint32)
	WarpFrame               func() uint32
	LeaveObserver           func(*Player)
	CameraUnlock            func(*Object)
	Move                    func(*Object, types.Pointf)
	Audio                   func(sound.ID, *Object, int, uint32)
	PointFX                 func(netmsg.Op, types.Pointf)
	ObjectSetOff            func(*Object) uint32
	MaybeWarp               func() int32
	PriMessage              func(*Object, strman.ID, byte)
}

// QuestExitTimeout4D71F0 preserves GAME.EXE 004D71F0. Once the unsigned
// 9000-frame countdown expires, a multiplayer party occupying an exit opens
// the Quest doors and broadcasts their state.
func (s *Server) QuestExitTimeout4D71F0(runtime QuestWarpGateRuntime4D71F0) uint32 {
	start := runtime.ExitCountdownStart()
	if start == 0 || s.Frame()-start < questExitTimeoutFrames4D71F0 {
		return start
	}

	hasExit := false
	for unit := s.Players.FirstUnit(); unit != nil; unit = s.questNextPlayerUnit4DA7F0(unit) {
		update := (*PlayerUpdateData)(unit.UpdateData)
		if update.QuestExit != nil {
			hasExit = true
		}
	}
	if !hasExit {
		return 0
	}

	players := s.questPlayerCount4E3CE0()
	if players <= 1 {
		return uint32(players)
	}
	runtime.StoreExitCountdownStart(0)
	if s.Doors.Sub_4D72C0() {
		return 1
	}
	s.Doors.Sub_4D72B0(true)
	return uint32(s.Sub_4D7280(255, s.Doors.Sub_4D72C0()))
}

// QuestLeaveWarpGate4D7480 preserves GAME.EXE 004D7480. The destination is
// copied before observer callbacks run, matching the original collide-data
// load, and the gate pointer is cleared before the unit is moved.
func (s *Server) QuestLeaveWarpGate4D7480(unit *Object, runtime QuestWarpGateRuntime4D71F0) {
	if unit == nil || uint8(unit.ObjClass)&questPlayerClassByte4D7480 == 0 {
		return
	}
	update := (*PlayerUpdateData)(unit.UpdateData)
	gate := update.QuestWarpGate
	if gate == nil {
		return
	}
	data := (*ExitCollideData)(gate.CollideData)
	destination := types.Pointf{X: data.DestinationX, Y: data.DestinationY}

	runtime.LeaveObserver(update.Player)
	runtime.CameraUnlock(unit)
	update.QuestWarpGate = nil
	runtime.Move(unit, destination)
	runtime.Audio(sound.SoundPlayerEliminated, unit, 2, unit.NetCode)
	runtime.PointFX(netmsg.MSG_FX_BLUE_SPARKS, destination)
}

// QuestSetWarpEnabled4D7520 preserves GAME.EXE 004D7520. The expensive
// cleanup only runs on the 1 -> 0 transition. Exit successors are captured
// before ObjectSetOff because that callback may unlink the current object.
func (s *Server) QuestSetWarpEnabled4D7520(enabled int32, runtime QuestWarpGateRuntime4D71F0) uint8 {
	previous := runtime.WarpEnabled()
	result := uint8(previous)
	if previous != 1 || enabled != 0 {
		runtime.StoreWarpEnabled(uint32(enabled))
		return result
	}

	for unit := s.Players.FirstUnit(); unit != nil; unit = s.questNextPlayerUnit4DA7F0(unit) {
		update := (*PlayerUpdateData)(unit.UpdateData)
		if update.Player.Field4792 != 0 && update.QuestWarpGate != nil {
			s.QuestLeaveWarpGate4D7480(unit, runtime)
		}
	}

	obj := s.Objs.First()
	if obj == nil {
		runtime.StoreWarpEnabled(0)
		return 0
	}
	for obj != nil {
		next := obj.Next()
		result = uint8(obj.ObjClass)
		if result&questExitClassByte4D7520 != 0 &&
			uint8(obj.ObjSubClass)&questExitSubclassByte4D7520 != 0 {
			result = uint8(runtime.ObjectSetOff(obj))
		}
		obj = next
	}
	runtime.StoreWarpEnabled(0)
	return result
}

// QuestCheckWarpGate4D7600 preserves GAME.EXE 004D7600. A failed party warp
// returns every qualifying player to the gate destination and prints the
// original solo or multiplayer restriction message.
func (s *Server) QuestCheckWarpGate4D7600(runtime QuestWarpGateRuntime4D71F0) {
	expected := s.questPlayerCount4E3CE0()
	if expected == 0 || s.Frame()-runtime.WarpFrame() < questWarpDelayFrames4D7600 {
		return
	}

	inGate := 0
	for unit := s.Players.FirstUnit(); unit != nil; unit = s.questNextPlayerUnit4DA7F0(unit) {
		update := (*PlayerUpdateData)(unit.UpdateData)
		if update.Player.Field4792 != 0 && update.QuestWarpGate != nil {
			inGate++
		}
	}
	if expected != inGate || runtime.MaybeWarp() != 0 {
		return
	}

	message := strman.ID("Gauntlet.c:WarpRestrictedSolo")
	if expected > 1 {
		message = "Gauntlet.c:WarpRestrictedMulti"
	}
	for unit := s.Players.FirstUnit(); unit != nil; unit = s.questNextPlayerUnit4DA7F0(unit) {
		update := (*PlayerUpdateData)(unit.UpdateData)
		if update.Player.Field4792 != 0 && update.QuestWarpGate != nil {
			s.QuestLeaveWarpGate4D7480(unit, runtime)
			runtime.PriMessage(unit, message, 0)
		}
	}
}
