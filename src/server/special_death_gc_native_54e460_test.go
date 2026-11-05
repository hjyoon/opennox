package server

import (
	"reflect"
	"runtime"
	"testing"
	"unsafe"
)

func TestSpecialDeathCachedCallbackGC54E460(t *testing.T) {
	for _, nilData := range []bool{true, false} {
		t.Run(map[bool]string{true: "nil-data", false: "live-data"}[nilData], func(t *testing.T) {
			unit := new(Object)
			var data unsafe.Pointer
			if !nilData {
				data = unsafe.Pointer(new(MonsterGenUpdateData))
			}
			unit.UpdateData = data
			replacement := new(MonsterGenUpdateData)
			var rt SpecialDeathRuntime54E460
			events := []string{}
			grow := func() { runtime.GC(); _ = skeletonOriginGrowStack54A310(128) }
			rt.Frame = func() uint32 { grow(); return 123 }
			rt.Timer = func(frame uint32) uint32 {
				if frame != 123 {
					t.Fatal("entry frame")
				}
				unit.UpdateData = unsafe.Pointer(replacement)
				grow()
				events = append(events, "timer")
				return 0
			}
			rt.Mode = func(value int32) uint32 { grow(); events = append(events, "mode"); return 0 }
			rt.Script = func(callback SpecialDeathCallback54E460, caller, trigger *Object, event uint32) uint32 {
				grow()
				if callback.Data != data || caller != nil || trigger != unit || event != 3 || callback.Address() != uintptr(data)+unsafe.Offsetof(MonsterGenUpdateData{}.Field56) {
					t.Fatal("original entry-cached callback across replacement/GC")
				}
				if nilData {
					if callback.NativeCallback() != nil {
						t.Fatal("nil base must use scalar foreign address")
					}
				} else if unsafe.Pointer(callback.NativeCallback()) != unsafe.Add(data, unsafe.Offsetof(MonsterGenUpdateData{}.Field56)) {
					t.Fatal("typed callback offset")
				}
				events = append(events, "script")
				return 0
			}
			rt.Audio = func(uint32, *Object, uint32, uint32) uint32 { grow(); events = append(events, "audio"); return 0 }
			rt.GeneratorFX = func(origin SkeletonOrigin54A310, value uint32) uint32 {
				grow()
				if origin.Source() != unit || value != 200 {
					t.Fatal("source-relative generator FX")
				}
				events = append(events, "fx")
				return 0
			}
			rt.Quest = func() int32 { grow(); events = append(events, "quest"); return 0 }
			rt.New = func(unsafe.Pointer) *Object { grow(); events = append(events, "new-nil"); return nil }
			rt.Delete = func(source *Object) uint32 {
				grow()
				if source != unit {
					t.Fatal("source identity")
				}
				events = append(events, "delete")
				return 0
			}
			unit.SpecialDeathNative54E460(2, nil, rt)
			if !reflect.DeepEqual(events, []string{"timer", "mode", "script", "audio", "fx", "quest", "new-nil", "delete"}) {
				t.Fatal("original order", events)
			}
		})
	}
}
