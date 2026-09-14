package server

import (
	"math"
	"testing"
	"unsafe"
)

func TestSparkUpdate53ADC0NativeDataAndLifetime(t *testing.T) {
	for _, tc := range []struct {
		name      string
		ticks     uint32
		mode      uint32
		wantTicks uint32
		wantFloat uint32
		wantDel   bool
	}{
		{"full damping", 2, 4, 1, math.Float32bits(1), false},
		{"partial damping", 1, 3, 0, 1064514355, false},
		{"expired", 0, 4, 0, math.Float32bits(0.5), true},
		{"negative lifetime", ^uint32(0), 0, ^uint32(0), math.Float32bits(0.5), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := &SparkUpdateData{LifetimeInitial: 0x1234, LifetimeRemaining: tc.ticks, Field8: 0x5678, Kind: tc.mode}
			obj := &Object{Float28: 0.5, UpdateData: unsafe.Pointer(data)}
			var deleted *Object
			new(Server).SparkUpdate53ADC0(obj, SparkUpdateRuntime53ADC0{DelayedDelete: func(got *Object) {
				deleted = got
			}})
			if data.LifetimeRemaining != tc.wantTicks || math.Float32bits(obj.Float28) != tc.wantFloat || (deleted == obj) != tc.wantDel {
				t.Fatalf("state = ticks %d, float %#x, deleted %t", data.LifetimeRemaining, math.Float32bits(obj.Float28), deleted == obj)
			}
			if data.LifetimeInitial != 0x1234 || data.Field8 != 0x5678 || data.Kind != tc.mode {
				t.Fatalf("adjacent update fields changed: %#v", data)
			}
		})
	}
}
