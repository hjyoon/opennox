package server

import (
	"github.com/opennox/libs/types"
	"reflect"
	"runtime"
	"testing"
	"unsafe"
)

func golemOriginFixture549E90() (*GolemDeathState549E90, GolemDeathRuntime549E90) {
	index, divisor := uint32(0), uint32(12)
	name := unsafe.Pointer(new(byte))
	return &GolemDeathState549E90{Index: &index, Divisor: &divisor}, GolemDeathRuntime549E90{
		Flags: func() uint32 { return 1 }, RandomInt: func(int32, int32, GolemRandomSite549E90) int32 { return 1 },
		MechName: func(slot int32) unsafe.Pointer {
			if slot == 0 {
				return name
			}
			return nil
		},
		RockName: func(int32) unsafe.Pointer { return name },
	}
}
func TestGolemDeathOriginReference549E90NilAllocationAfterStackGrowth(t *testing.T) {
	for _, mech := range []bool{false, true} {
		t.Run(map[bool]string{false: "golem", true: "mech"}[mech], func(t *testing.T) {
			state, rt := golemOriginFixture549E90()
			events := []string{}
			rt.Audio = func(id uint32, source *Object, kind, arg uint32) uint32 {
				want := uint32(487)
				if mech {
					want = 494
				}
				if id != want || source != nil || kind != 0 || arg != 0 {
					t.Fatal("audio arguments")
				}
				events = append(events, "audio")
				return 0
			}
			rt.PointFX = func(id uint32, origin GolemOrigin549E90) uint32 {
				runtime.GC()
				_ = skeletonOriginGrowStack54A310(128)
				if id != 0x8a || origin.Source() != nil || origin.NativePoint() != nil || origin.Address() != unsafe.Offsetof(Object{}.PosVec) {
					t.Fatal("nil origin identity")
				}
				events = append(events, "pointfx")
				return 0
			}
			rt.New = func(unsafe.Pointer) *Object {
				runtime.GC()
				_ = skeletonOriginGrowStack54A310(128)
				events = append(events, "new-nil")
				return nil
			}
			if got := (*Object)(nil).GolemDeathNative549E90(mech, state, rt); got != 1 || *state.Index != 0 || !reflect.DeepEqual(events, []string{"audio", "pointfx", "new-nil"}) {
				t.Fatalf("result=%d events=%v", got, events)
			}
		})
	}
}
func TestGolemDeathOriginReference549E90FaultAtConsumingRead(t *testing.T) {
	for _, mech := range []bool{false, true} {
		t.Run(map[bool]string{false: "golem", true: "mech"}[mech], func(t *testing.T) {
			state, rt := golemOriginFixture549E90()
			events := []string{}
			var recovered any
			rt.Audio = func(uint32, *Object, uint32, uint32) uint32 { events = append(events, "audio"); return 0 }
			rt.PointFX = func(uint32, GolemOrigin549E90) uint32 { events = append(events, "pointfx"); return 0 }
			rt.New = func(unsafe.Pointer) *Object { events = append(events, "new"); return new(Object) }
			rt.Point = func(_ float32, origin GolemOrigin549E90, _ *types.Pointf) {
				runtime.GC()
				_ = skeletonOriginGrowStack54A310(128)
				events = append(events, "point")
				_ = origin.Read()
			}
			func() {
				defer func() { recovered = recover() }()
				(*Object)(nil).GolemDeathNative549E90(mech, state, rt)
			}()
			if _, ok := recovered.(runtime.Error); !ok || !reflect.DeepEqual(events, []string{"audio", "pointfx", "new", "point"}) {
				t.Fatalf("fault=%v events=%v", recovered, events)
			}
		})
	}
}
func TestGolemDeathOriginReference549E90LiveSourceSurvivesNestedCallAndGC(t *testing.T) {
	for _, mech := range []bool{false, true} {
		t.Run(map[bool]string{false: "golem", true: "mech"}[mech], func(t *testing.T) {
			state, rt := golemOriginFixture549E90()
			first := &Object{PosVec: types.Ptf(3, 4)}
			replacement := &Object{PosVec: types.Ptf(99, 100)}
			source := first
			var saved GolemOrigin549E90
			nested := false
			points, forces := 0, 0
			rt.Audio = func(uint32, *Object, uint32, uint32) uint32 { return 0 }
			rt.PointFX = func(_ uint32, origin GolemOrigin549E90) uint32 {
				if !nested {
					saved = origin
				}
				return 0
			}
			rt.New = func(unsafe.Pointer) *Object {
				if nested {
					return nil
				}
				nested = true
				source = replacement
				first.PosVec = types.Ptf(17, 19)
				if replacement.GolemDeathNative549E90(mech, state, rt) != 1 {
					t.Fatal("nested allocation short circuit")
				}
				nested = false
				runtime.GC()
				_ = skeletonOriginGrowStack54A310(128)
				return new(Object)
			}
			rt.Point = func(_ float32, origin GolemOrigin549E90, out *types.Pointf) {
				runtime.GC()
				_ = skeletonOriginGrowStack54A310(128)
				if origin.Source() != first || origin.Read() != (types.Ptf(17, 19)) {
					t.Fatal("live cached source after nested call")
				}
				*out = origin.Read()
				points++
			}
			rt.Create = func(*Object, *Object, types.Pointf, uint32) uint32 { return 0 }
			rt.RandomFloat = func(float32, float32) float64 { return 1 }
			rt.Raise = func(*Object, float32) uint32 {
				first.PosVec = types.Ptf(23, 29)
				runtime.GC()
				_ = skeletonOriginGrowStack54A310(128)
				return 0
			}
			rt.Mass = func(int32) byte { return 1 }
			rt.Force = func(origin GolemOrigin549E90, _ *Object, _ float32) uint32 {
				if origin.Source() != first || origin.Read() != types.Ptf(23, 29) {
					t.Fatal("live force origin")
				}
				forces++
				return 0
			}
			rt.FPS = func() uint32 { return 30 }
			rt.Decay = func(*Object, uint32) uint32 { return 0 }
			if first.GolemDeathNative549E90(mech, state, rt) != 1 || points != 1 || (!mech && forces != 1) {
				t.Fatal("outer native completion", points, forces)
			}
			runtime.GC()
			_ = skeletonOriginGrowStack54A310(128)
			if saved.Source() != first || saved.Source() == source || saved.NativePoint() != &first.PosVec || saved.Read() != types.Ptf(23, 29) {
				t.Fatal("cached source identity across nested call/GC")
			}
		})
	}
}
