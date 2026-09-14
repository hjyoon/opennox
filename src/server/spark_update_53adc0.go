package server

import (
	"math"
)

type SparkUpdateRuntime53ADC0 struct {
	DelayedDelete func(*Object)
}

// SparkUpdate53ADC0 restores GAME.EXE 0053ADC0 without dereferencing the
// PE32 Object+748 update-data slot, which is a different field on 64-bit hosts.
func (s *Server) SparkUpdate53ADC0(obj *Object, runtime SparkUpdateRuntime53ADC0) {
	if obj == nil || obj.UpdateData == nil {
		return
	}
	data := (*SparkUpdateData)(obj.UpdateData)
	if int32(data.LifetimeRemaining) <= 0 {
		if runtime.DelayedDelete != nil {
			runtime.DelayedDelete(obj)
		}
		return
	}
	data.LifetimeRemaining--
	if data.Kind == 4 {
		obj.Float28 = 1
	} else {
		obj.Float28 = math.Float32frombits(1064514355)
	}
}
