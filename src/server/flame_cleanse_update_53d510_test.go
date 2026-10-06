package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
)

func TestFlameCleanseUpdate53D510LifetimeTraceAndAge(t *testing.T) {
	for _, tc := range []struct {
		name                             string
		first, second, deadline, created uint32
		clear, stationary, deleted       bool
		want                             []string
	}{
		{"deadline equal", 100, 0, 100, 0, true, true, true, []string{"frame", "delete"}},
		{"deadline passed", 101, 0, 100, 0, true, true, true, []string{"frame", "delete"}},
		{"unsigned high deadline", 100, 102, 0xfffffffe, 100, true, true, false, []string{"frame", "trace", "frame"}},
		{"unsigned high frame", 0xfffffffe, 0, 100, 0, true, true, true, []string{"frame", "delete"}},
		{"blocked", 100, 0, 101, 0, false, true, true, []string{"frame", "trace", "delete"}},
		{"age three", 100, 103, 1000, 100, true, true, false, []string{"frame", "trace", "frame"}},
		{"age four stationary", 100, 104, 1000, 100, true, true, true, []string{"frame", "trace", "frame", "delete"}},
		{"age four moving", 100, 104, 1000, 100, true, false, false, []string{"frame", "trace", "frame"}},
		{"wrapped age three", 1, 2, 100, 0xffffffff, true, true, false, []string{"frame", "trace", "frame"}},
		{"wrapped age four", 1, 3, 100, 0xffffffff, true, true, true, []string{"frame", "trace", "frame", "delete"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			obj := &Object{Field34: tc.deadline, Field32: tc.created, Pos39: types.Ptf(11, 22), PosVec: types.Ptf(33, 44), PrevPos: types.Ptf(33, 44)}
			if !tc.stationary {
				obj.PrevPos.X = 32
			}
			before := *obj
			var events []string
			deleted, reads := false, 0
			flameCleanseUpdate53D510(obj, flameCleanseUpdateDeps53D510{
				frame: func() uint32 {
					events = append(events, "frame")
					reads++
					if reads == 1 {
						return tc.first
					}
					return tc.second
				},
				traceRay: func(from, to types.Pointf, flags MapTraceFlags) bool {
					events = append(events, "trace")
					if from != obj.Pos39 || to != obj.PosVec || flags != 65 {
						t.Fatalf("trace=%v,%v,%d", from, to, flags)
					}
					return tc.clear
				},
				delayedDelete: func(got *Object) {
					events = append(events, "delete")
					if got != obj {
						t.Fatal("deletion lost native pointer")
					}
					deleted = true
				},
			})
			if deleted != tc.deleted || !reflect.DeepEqual(events, tc.want) || *obj != before {
				t.Fatalf("deleted/events/mutated=%t/%v/%t", deleted, events, *obj != before)
			}
		})
	}
}

func TestFlameCleanseUpdate53D510ReloadsAfterTrace(t *testing.T) {
	obj := &Object{Field34: 200, Field32: 100, Pos39: types.Ptf(11, 22), PosVec: types.Ptf(33, 44), PrevPos: types.Ptf(0, 0)}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(obj)) <= math.MaxUint32 {
		t.Fatalf("object=%p, want actual pointer above 4 GiB", obj)
	}
	frame, deleted, reads := uint32(100), false, 0
	flameCleanseUpdate53D510(obj, flameCleanseUpdateDeps53D510{
		frame: func() uint32 { reads++; return frame },
		traceRay: func(from, to types.Pointf, flags MapTraceFlags) bool {
			frame = 104
			obj.Field32 = 99
			obj.Field34 = 0 // The deadline is not tested again after this callback.
			obj.PosVec, obj.PrevPos = types.Ptf(88, 99), types.Ptf(88, 99)
			if from != types.Ptf(11, 22) || to != types.Ptf(33, 44) || flags != 65 {
				t.Fatal("trace did not snapshot coordinates")
			}
			return true
		},
		delayedDelete: func(got *Object) {
			if got != obj {
				t.Fatal("pointer narrowed")
			}
			deleted = true
		},
	})
	if !deleted || reads != 2 {
		t.Fatalf("deleted/frame reads=%t/%d", deleted, reads)
	}
}

func TestFlameCleanseUpdate53D510X87C3(t *testing.T) {
	values := []float32{0, math.Float32frombits(0x80000000), 1, -1, float32(math.Inf(1)), float32(math.Inf(-1)), math.Float32frombits(0x7fc01234), math.Float32frombits(0x7f800001)}
	for _, a := range values {
		for _, b := range values {
			want := a == b || math.IsNaN(float64(a)) || math.IsNaN(float64(b))
			for _, axis := range []string{"X", "Y"} {
				t.Run(fmt.Sprintf("%s/%08x/%08x", axis, math.Float32bits(a), math.Float32bits(b)), func(t *testing.T) {
					obj := &Object{Field34: 1000, PosVec: types.Ptf(7, 9), PrevPos: types.Ptf(7, 9)}
					if axis == "X" {
						obj.PrevPos.X, obj.PosVec.X = a, b
					} else {
						obj.PrevPos.Y, obj.PosVec.Y = a, b
					}
					deleted := false
					flameCleanseUpdate53D510(obj, flameCleanseUpdateDeps53D510{frame: func() uint32 { return 4 }, traceRay: func(types.Pointf, types.Pointf, MapTraceFlags) bool { return true }, delayedDelete: func(*Object) { deleted = true }})
					if deleted != want {
						t.Fatalf("C3 deletion=%t, want %t", deleted, want)
					}
				})
			}
		}
	}
}

func TestFlameCleanseUpdate53D510FaultPrefix(t *testing.T) {
	called := false
	defer func() {
		if recover() == nil || !called {
			t.Fatal("null object must fault after first frame read")
		}
	}()
	flameCleanseUpdate53D510(nil, flameCleanseUpdateDeps53D510{frame: func() uint32 { called = true; return 0 }})
}
