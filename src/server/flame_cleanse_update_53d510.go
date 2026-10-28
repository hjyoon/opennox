package server

import (
	"math"

	"github.com/opennox/libs/types"
)

type flameCleanseUpdateDeps53D510 struct {
	frame         func() uint32
	traceRay      func(types.Pointf, types.Pointf, MapTraceFlags) bool
	delayedDelete func(*Object)
}

// flameCleanseEqual53D510 models the C3-only x87 FCOMPS tests at
// 0053D573/0053D580: both equality and unordered comparisons set C3.
func flameCleanseEqual53D510(a, b float32) bool {
	return a == b || math.IsNaN(float64(a)) || math.IsNaN(float64(b))
}

// flameCleanseUpdate53D510 restores GAME.EXE 0053D510..0053D597. The two
// frame reads bracket the trace callback; age and deadline comparisons are
// unsigned DWORD operations. None of the native Object pointers are narrowed.
func flameCleanseUpdate53D510(obj *Object, h flameCleanseUpdateDeps53D510) {
	LifetimeEffectUpdateRaw53CC90(LifetimeEffectFlameCleanse53D510, obj, LifetimeEffectRuntime53CC90{
		Frame: h.frame, DelayedDelete: h.delayedDelete, TraceRay: func(r *types.Rectf, f MapTraceFlags) bool { return h.traceRay(r.Min, r.Max, f) },
	})
}

type FlameCleanseUpdateRuntime53D510 struct {
	DelayedDelete func(*Object)
}

func (s *Server) FlameCleanseUpdate53D510(obj *Object, runtime FlameCleanseUpdateRuntime53D510) {
	flameCleanseUpdate53D510(obj, flameCleanseUpdateDeps53D510{
		frame:         s.Frame,
		traceRay:      s.MapTraceRay,
		delayedDelete: runtime.DelayedDelete,
	})
}
