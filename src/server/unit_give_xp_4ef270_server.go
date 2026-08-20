package server

import "math"

// UnitGiveXPRuntime4EF270 supplies the fixed-width protection service and the
// next sequential level-update routine. Object-bearing arguments retain their
// native pointer width.
type UnitGiveXPRuntime4EF270 struct {
	ProtectExperience func(uint32, float32)
	SyncLevel         func(*Object)
}

type unitGiveXPNativeDeps4EF270 struct {
	protectExperience func(uint32, float32)
	reportExperience  func(*Object)
	syncLevel         func(*Object)
}

func unitGiveXPNative4EF270(
	unit *Object,
	target float32,
	deps unitGiveXPNativeDeps4EF270,
) float64 {
	return unitGiveXP4EF270(unitGiveXPHooks4EF270[
		*Object,
		*PlayerUpdateData,
		*Player,
	]{
		loadUnitArg: func() *Object {
			return unit
		},
		loadExperience: func(unit *Object) float32 {
			return unit.Experience
		},
		loadTargetArg: func() float32 {
			return target
		},
		loadUpdateData: func(unit *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(unit.UpdateData)
		},
		loadScale: func() float32 {
			return math.Float32frombits(unitGiveXPScaleBits4EF270)
		},
		loadOne: func() float32 {
			return math.Float32frombits(unitGiveXPOneBits4EF270)
		},
		loadZero: func() float32 {
			return math.Float32frombits(unitGiveXPZeroBits4EF270)
		},
		storeExperience: func(unit *Object, experience float32) {
			unit.Experience = experience
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadExperienceToken: func(player *Player) uint32 {
			return player.ProtUnitExperience
		},
		protectExperience: deps.protectExperience,
		reportExperience:  deps.reportExperience,
		syncLevel:         deps.syncLevel,
	})
}

// UnitGiveXP4EF270 binds GAME.EXE 004EF270 to native Object,
// PlayerUpdateData, and Player layouts. The opaque protection token stays
// uint32, while every object pointer passed to the remaining services keeps
// the host pointer width.
func (s *Server) UnitGiveXP4EF270(
	unit *Object,
	target float32,
	runtime UnitGiveXPRuntime4EF270,
) float64 {
	return unitGiveXPNative4EF270(unit, target, unitGiveXPNativeDeps4EF270{
		protectExperience: runtime.ProtectExperience,
		reportExperience:  s.NetReportExperience,
		syncLevel:         runtime.SyncLevel,
	})
}
