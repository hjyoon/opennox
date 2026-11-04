package server

import (
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
)

//go:noinline
func skeletonOriginGrowStack54A310(n int) uint64 {
	var frame [1024]byte
	frame[n%len(frame)] = byte(n)
	var value uint64
	if n != 0 {
		value = skeletonOriginGrowStack54A310(n - 1)
	}
	return value + uint64(frame[n%len(frame)])
}

func TestSkeletonDeathOriginReference54A310NilAllocationAfterStackGrowth(t *testing.T) {
	var events []string
	var unit *Object
	rt := SkeletonDeathRuntime54A310{
		Audio: func(id uint32, source *Object, kind, arg uint32) uint32 {
			if id != 356 || source != nil || kind != 0 || arg != 0 {
				t.Fatal("audio arguments")
			}
			events = append(events, "audio")
			return 0
		},
		PointFX: func(id uint32, origin SkeletonOrigin54A310) uint32 {
			runtime.GC()
			_ = skeletonOriginGrowStack54A310(128)
			if id != 0x8a || origin.Source() != nil || origin.NativePoint() != nil || origin.Address() != unsafe.Offsetof(Object{}.PosVec) {
				t.Fatal("nil-base address changed before its consuming service")
			}
			events = append(events, "pointfx")
			return 0
		},
		New: func(unsafe.Pointer) *Object {
			events = append(events, "new-nil")
			return nil
		},
	}
	if got := unit.SkeletonDebrisNative54A4C0(nil, rt); got != 1 || !reflect.DeepEqual(events, []string{"audio", "pointfx", "new-nil"}) {
		t.Fatalf("result=%d events=%v", got, events)
	}
}

func TestSkeletonDeathOriginReference54A310FaultAtConsumingRead(t *testing.T) {
	var events []string
	var recovered any
	rt := SkeletonDeathRuntime54A310{
		Audio: func(uint32, *Object, uint32, uint32) uint32 { events = append(events, "audio"); return 0 },
		PointFX: func(uint32, SkeletonOrigin54A310) uint32 {
			events = append(events, "pointfx")
			return 0
		},
		New: func(unsafe.Pointer) *Object { events = append(events, "new"); return new(Object) },
		Point: func(_ float32, origin SkeletonOrigin54A310, _ *types.Pointf) {
			runtime.GC()
			_ = skeletonOriginGrowStack54A310(128)
			events = append(events, "point")
			_ = origin.Read()
		},
	}
	func() {
		defer func() { recovered = recover() }()
		(*Object)(nil).SkeletonDebrisNative54A4C0(nil, rt)
	}()
	if _, ok := recovered.(runtime.Error); !ok || !reflect.DeepEqual(events, []string{"audio", "pointfx", "new", "point"}) {
		t.Fatalf("fault=%v events=%v", recovered, events)
	}
}

func TestSkeletonDeathOriginReference54A310LiveSourceSurvivesReentryAndGC(t *testing.T) {
	first := &Object{PosVec: types.Pointf{X: 3, Y: 4}}
	source := first
	var saved SkeletonOrigin54A310
	rt := SkeletonDeathRuntime54A310{
		Audio: func(uint32, *Object, uint32, uint32) uint32 { return 0 },
		PointFX: func(_ uint32, origin SkeletonOrigin54A310) uint32 {
			saved = origin
			first.PosVec = types.Pointf{X: 17, Y: 19}
			source = &Object{PosVec: types.Pointf{X: 99, Y: 100}}
			runtime.GC()
			_ = skeletonOriginGrowStack54A310(128)
			return 0
		},
		New: func(unsafe.Pointer) *Object { return nil },
	}
	if source.SkeletonDebrisNative54A4C0(nil, rt) != 1 {
		t.Fatal("debris allocation-failure return")
	}
	runtime.GC()
	_ = skeletonOriginGrowStack54A310(128)
	if saved.Source() != first || saved.Source() == source || saved.NativePoint() != &first.PosVec || saved.Address() != uintptr(unsafe.Pointer(&first.PosVec)) || saved.Read() != (types.Pointf{X: 17, Y: 19}) {
		t.Fatal("cached native position source or live value changed")
	}
}
