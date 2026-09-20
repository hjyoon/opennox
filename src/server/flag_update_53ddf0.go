package server

import "github.com/opennox/libs/types"

const (
	flagUpdateReturnSeconds53DDF0 = uint32(3)
	flagUpdateResetSeconds53DDF0  = uint32(30)
	flagUpdateReturnAudio53DDF0   = uint32(305)
)

// FlagUpdateRuntime53DDF0 supplies the side effects that remain owned by the
// outer server. Object and update-data identities stay native-width.
type FlagUpdateRuntime53DDF0 struct {
	AudioEvent func(uint32, *Object)
	FlagStatus func(uint8, uint8, uint8, uint16) int32
	Move       func(*Object, types.Pointf)
	InformHome func(uint32) int32
}

// FlagUpdate53DDF0 preserves GAME.EXE 0053DDF0. A dropped team flag keeps the
// frame at which it was dropped in State. After thirty seconds it is moved to
// Home, its team status is cleared, and the four-byte flag material index is
// broadcast to every player.
//
// The original callback returns 3*FPS while a nonzero timer is active and the
// result of the inform call after expiry. The object update dispatcher ignores
// that value, but retaining it keeps direct C callers compatible.
func (s *Server) FlagUpdate53DDF0(flag *Object, runtime FlagUpdateRuntime53DDF0) int32 {
	if flag == nil || flag.UpdateData == nil {
		return 0
	}
	update := (*FlagUpdateData4EA490)(flag.UpdateData)
	if update.State == 0 {
		return 0
	}

	flagIndex := TeamMaterialObjectIndex4ECBD0(flag)
	teamID := uint8(flag.TeamVal.ID)
	result := int32(flagUpdateReturnSeconds53DDF0 * s.TickRate())
	if s.Frame()-update.State <= flagUpdateResetSeconds53DDF0*s.TickRate() {
		return result
	}

	if runtime.AudioEvent != nil {
		runtime.AudioEvent(flagUpdateReturnAudio53DDF0, flag)
	}
	update.State = 0
	if runtime.FlagStatus != nil {
		runtime.FlagStatus(teamID, 0, uint8(flagIndex), 0)
	}
	if runtime.Move != nil {
		runtime.Move(flag, update.Home)
	}
	if runtime.InformHome != nil {
		result = runtime.InformHome(flagIndex)
	}
	return result
}
