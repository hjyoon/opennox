package server

import (
	"fmt"
	"math"
	"reflect"
	"runtime"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

const (
	dangerousUnitCandidate547120 = uint64(0x1234567889abcdef)
	dangerousUnitUnit547120      = uint64(0xfedcba9876543210)
)

type dangerousUnitTestObject547120 struct {
	typeInd  uint16
	class    uint32
	subClass uint32
}

type dangerousUnitTestWorld547120 struct {
	cache   dangerousUnitTypeCache547120
	objects map[uint64]*dangerousUnitTestObject547120
	lookups map[string]uint32
	events  []string
	cleared bool
	faultAt int
	after   map[string]func()
}

func (w *dangerousUnitTestWorld547120) record(event string) {
	if w.faultAt != 0 && len(w.events)+1 == w.faultAt {
		panic(event)
	}
	w.events = append(w.events, event)
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *dangerousUnitTestWorld547120) hooks() dangerousUnitHooks547120[uint64] {
	return dangerousUnitHooks547120[uint64]{
		loadToxicCloudCache: func() uint32 {
			value := w.cache.toxicCloud
			w.record("toxic-cache")
			return value
		},
		lookupType: func(name string) uint32 {
			value := w.lookups[name]
			w.record("lookup:" + name)
			return value
		},
		storeToxicCloudCache: func(value uint32) {
			w.record("store:ToxicCloud")
			w.cache.toxicCloud = value
		},
		storeSmallToxicCloudCache: func(value uint32) {
			w.record("store:SmallToxicCloud")
			w.cache.smallToxicCloud = value
		},
		loadClass: func(value uint64) uint32 {
			class := w.objects[value].class
			w.record("class")
			return class
		},
		loadType: func(value uint64) uint16 {
			typeInd := w.objects[value].typeInd
			w.record("type")
			return typeInd
		},
		loadSmallToxicCloudCache: func() uint32 {
			value := w.cache.smallToxicCloud
			w.record("small-cache")
			return value
		},
		loadSubClass: func(value uint64) uint32 {
			subClass := w.objects[value].subClass
			w.record("subclass")
			return subClass
		},
		clearLocationSafe: func() {
			w.record("clear-safe")
			w.cleared = true
		},
	}
}

func newDangerousUnitTestWorld547120(
	candidate dangerousUnitTestObject547120,
	unit dangerousUnitTestObject547120,
) *dangerousUnitTestWorld547120 {
	return &dangerousUnitTestWorld547120{
		cache: dangerousUnitTypeCache547120{
			toxicCloud:      0x21,
			smallToxicCloud: 0x22,
		},
		objects: map[uint64]*dangerousUnitTestObject547120{
			dangerousUnitCandidate547120: &candidate,
			dangerousUnitUnit547120:      &unit,
		},
		lookups: map[string]uint32{
			"ToxicCloud":      0x21,
			"SmallToxicCloud": 0x22,
		},
		after: make(map[string]func()),
	}
}

func TestDangerousUnit547120BranchMatrixAndObservationOrder(t *testing.T) {
	tests := []struct {
		name       string
		candidate  dangerousUnitTestObject547120
		unit       dangerousUnitTestObject547120
		wantResult uint32
		wantClear  bool
		wantEvents []string
	}{
		{
			name:       "fire immune",
			candidate:  dangerousUnitTestObject547120{class: uint32(object.ClassFire | object.ClassDangerous), typeInd: 0x21},
			unit:       dangerousUnitTestObject547120{subClass: 0x400},
			wantResult: 1,
			wantEvents: []string{"toxic-cache", "class", "subclass"},
		},
		{
			name:       "fire vulnerable",
			candidate:  dangerousUnitTestObject547120{class: uint32(object.ClassFire)},
			wantResult: 0,
			wantClear:  true,
			wantEvents: []string{"toxic-cache", "class", "subclass", "clear-safe"},
		},
		{
			name:       "toxic cloud immune",
			candidate:  dangerousUnitTestObject547120{class: uint32(object.ClassDangerous), typeInd: 0x21},
			unit:       dangerousUnitTestObject547120{subClass: 0xa5a50200},
			wantResult: 0xa5a50200,
			wantEvents: []string{"toxic-cache", "class", "type", "toxic-cache", "subclass"},
		},
		{
			name:       "toxic cloud vulnerable",
			candidate:  dangerousUnitTestObject547120{typeInd: 0x21},
			unit:       dangerousUnitTestObject547120{subClass: 0xfffffc00},
			wantResult: 0xfffffc00,
			wantClear:  true,
			wantEvents: []string{"toxic-cache", "class", "type", "toxic-cache", "subclass", "clear-safe"},
		},
		{
			name:       "small toxic cloud",
			candidate:  dangerousUnitTestObject547120{typeInd: 0x22},
			unit:       dangerousUnitTestObject547120{subClass: 0x200},
			wantResult: 0x200,
			wantEvents: []string{"toxic-cache", "class", "type", "toxic-cache", "small-cache", "subclass"},
		},
		{
			name:       "dangerous class",
			candidate:  dangerousUnitTestObject547120{class: uint32(object.ClassDangerous), typeInd: 0x31},
			wantResult: 0x31,
			wantClear:  true,
			wantEvents: []string{"toxic-cache", "class", "type", "toxic-cache", "small-cache", "clear-safe"},
		},
		{
			name:       "benign class",
			candidate:  dangerousUnitTestObject547120{typeInd: 0x31},
			wantResult: 0x31,
			wantEvents: []string{"toxic-cache", "class", "type", "toxic-cache", "small-cache"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newDangerousUnitTestWorld547120(tc.candidate, tc.unit)
			got := dangerousUnit547120(dangerousUnitCandidate547120, dangerousUnitUnit547120, w.hooks())
			if got != tc.wantResult {
				t.Fatalf("result = %#x, want %#x", got, tc.wantResult)
			}
			if w.cleared != tc.wantClear {
				t.Fatalf("location-safe clear = %t, want %t", w.cleared, tc.wantClear)
			}
			if !reflect.DeepEqual(w.events, tc.wantEvents) {
				t.Fatalf("events = %q, want %q", w.events, tc.wantEvents)
			}
		})
	}
}

func TestDangerousUnit547120LazyCacheSentinelAndStoreOrder(t *testing.T) {
	w := newDangerousUnitTestWorld547120(
		dangerousUnitTestObject547120{typeInd: 0x31},
		dangerousUnitTestObject547120{},
	)
	w.cache = dangerousUnitTypeCache547120{}

	if got := dangerousUnit547120(dangerousUnitCandidate547120, dangerousUnitUnit547120, w.hooks()); got != 0x31 {
		t.Fatalf("result = %#x, want %#x", got, uint32(0x31))
	}
	want := []string{
		"toxic-cache",
		"lookup:ToxicCloud", "store:ToxicCloud",
		"lookup:SmallToxicCloud", "store:SmallToxicCloud",
		"class", "type", "toxic-cache", "small-cache",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
	if w.cache.toxicCloud != 0x21 || w.cache.smallToxicCloud != 0x22 {
		t.Fatalf("cache = {%#x, %#x}, want {0x21, 0x22}", w.cache.toxicCloud, w.cache.smallToxicCloud)
	}
}

func TestDangerousUnit547120ZeroToxicLookupRepeatsBothLookups(t *testing.T) {
	w := newDangerousUnitTestWorld547120(
		dangerousUnitTestObject547120{typeInd: 0x31},
		dangerousUnitTestObject547120{},
	)
	w.cache = dangerousUnitTypeCache547120{}
	w.lookups["ToxicCloud"] = 0

	for call := 0; call < 2; call++ {
		if got := dangerousUnit547120(dangerousUnitCandidate547120, dangerousUnitUnit547120, w.hooks()); got != 0x31 {
			t.Fatalf("call %d result = %#x, want %#x", call+1, got, uint32(0x31))
		}
	}
	lookupCount := 0
	for _, event := range w.events {
		if event == "lookup:ToxicCloud" || event == "lookup:SmallToxicCloud" {
			lookupCount++
		}
	}
	if lookupCount != 4 {
		t.Fatalf("lookup count = %d, want 4; events = %q", lookupCount, w.events)
	}
	if w.cache.toxicCloud != 0 || w.cache.smallToxicCloud != 0x22 {
		t.Fatalf("cache = {%#x, %#x}, want {0, 0x22}", w.cache.toxicCloud, w.cache.smallToxicCloud)
	}
}

func TestDangerousUnit547120ZeroSmallLookupDoesNotReinitialize(t *testing.T) {
	w := newDangerousUnitTestWorld547120(
		dangerousUnitTestObject547120{typeInd: 0x31},
		dangerousUnitTestObject547120{},
	)
	w.cache = dangerousUnitTypeCache547120{}
	w.lookups["SmallToxicCloud"] = 0

	for call := 0; call < 2; call++ {
		dangerousUnit547120(dangerousUnitCandidate547120, dangerousUnitUnit547120, w.hooks())
	}
	lookupCount := 0
	for _, event := range w.events {
		if len(event) >= len("lookup:") && event[:len("lookup:")] == "lookup:" {
			lookupCount++
		}
	}
	if lookupCount != 2 {
		t.Fatalf("lookup count = %d, want 2; events = %q", lookupCount, w.events)
	}
	if w.cache.toxicCloud != 0x21 || w.cache.smallToxicCloud != 0 {
		t.Fatalf("cache = {%#x, %#x}, want {0x21, 0}", w.cache.toxicCloud, w.cache.smallToxicCloud)
	}
}

func TestDangerousUnit547120ReloadsCacheAndRetainsObservedClass(t *testing.T) {
	w := newDangerousUnitTestWorld547120(
		dangerousUnitTestObject547120{class: uint32(object.ClassDangerous), typeInd: 0x31},
		dangerousUnitTestObject547120{subClass: 0x200},
	)
	w.after["class"] = func() {
		w.objects[dangerousUnitCandidate547120].class = 0
	}
	w.after["type"] = func() {
		w.cache.toxicCloud = 0x31
	}

	got := dangerousUnit547120(dangerousUnitCandidate547120, dangerousUnitUnit547120, w.hooks())
	if got != 0x200 || w.cleared {
		t.Fatalf("reloaded cloud result/clear = %#x/%t, want 0x200/false", got, w.cleared)
	}
	want := []string{"toxic-cache", "class", "type", "toxic-cache", "subclass"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}

	w = newDangerousUnitTestWorld547120(
		dangerousUnitTestObject547120{class: uint32(object.ClassDangerous), typeInd: 0x31},
		dangerousUnitTestObject547120{},
	)
	w.after["class"] = func() {
		w.objects[dangerousUnitCandidate547120].class = 0
	}
	dangerousUnit547120(dangerousUnitCandidate547120, dangerousUnitUnit547120, w.hooks())
	if !w.cleared {
		t.Fatal("Dangerous bit observed before mutation did not clear location-safe")
	}
}

func runDangerousUnitTest547120(w *dangerousUnitTestWorld547120) (result uint32, panicValue any) {
	func() {
		defer func() {
			panicValue = recover()
		}()
		result = dangerousUnit547120(dangerousUnitCandidate547120, dangerousUnitUnit547120, w.hooks())
	}()
	return result, panicValue
}

func newDangerousUnitFaultWorld547120() *dangerousUnitTestWorld547120 {
	w := newDangerousUnitTestWorld547120(
		dangerousUnitTestObject547120{typeInd: 0x22},
		dangerousUnitTestObject547120{},
	)
	w.cache = dangerousUnitTypeCache547120{}
	return w
}

func TestDangerousUnit547120FaultPrefixes(t *testing.T) {
	baseline := newDangerousUnitFaultWorld547120()
	result, panicValue := runDangerousUnitTest547120(baseline)
	if panicValue != nil || result != 0 || !baseline.cleared {
		t.Fatalf("baseline result/panic/clear = %#x/%v/%t", result, panicValue, baseline.cleared)
	}

	for faultAt := 1; faultAt <= len(baseline.events); faultAt++ {
		t.Run(fmt.Sprintf("fault_%02d", faultAt), func(t *testing.T) {
			w := newDangerousUnitFaultWorld547120()
			w.faultAt = faultAt
			_, gotPanic := runDangerousUnitTest547120(w)
			if gotPanic != baseline.events[faultAt-1] {
				t.Fatalf("panic = %v, want %q", gotPanic, baseline.events[faultAt-1])
			}
			wantPrefix := baseline.events[:faultAt-1]
			if !slices.Equal(w.events, wantPrefix) {
				t.Fatalf("fault prefix = %q, want %q", w.events, wantPrefix)
			}
		})
	}
}

func TestDangerousUnitNative547120LayoutAndHighPointers(t *testing.T) {
	if got, want := uint32(object.ClassFire), uint32(0x2000); got != want {
		t.Fatalf("object.ClassFire = %#x, want %#x", got, want)
	}
	if got, want := uint32(object.ClassDangerous), uint32(0x10000); got != want {
		t.Fatalf("object.ClassDangerous = %#x, want %#x", got, want)
	}

	var layout Object
	wantTypeOffset := uintptr(4)
	wantClassOffset := uintptr(8)
	wantSubClassOffset := uintptr(12)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantTypeOffset = 8
		wantClassOffset = 12
		wantSubClassOffset = 16
	}
	if got := unsafe.Offsetof(layout.TypeInd); got != wantTypeOffset {
		t.Fatalf("Object.TypeInd offset = %d, want %d", got, wantTypeOffset)
	}
	if got := unsafe.Offsetof(layout.ObjClass); got != wantClassOffset {
		t.Fatalf("Object.ObjClass offset = %d, want %d", got, wantClassOffset)
	}
	if got := unsafe.Offsetof(layout.ObjSubClass); got != wantSubClassOffset {
		t.Fatalf("Object.ObjSubClass offset = %d, want %d", got, wantSubClassOffset)
	}

	candidate := &Object{TypeInd: 0x31, ObjClass: object.ClassDangerous}
	unit := &Object{}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]unsafe.Pointer{
			"candidate": unsafe.Pointer(candidate),
			"unit":      unsafe.Pointer(unit),
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("%s pointer = %p, want a native address above 4 GiB", name, pointer)
			}
		}
	}

	cache := dangerousUnitTypeCache547120{toxicCloud: 0x21, smallToxicCloud: 0x22}
	cleared := false
	got := dangerousUnitNative547120(candidate, unit, &cache, func(name string) uint32 {
		t.Fatalf("initialized cache unexpectedly looked up %q", name)
		return 0
	}, func() {
		cleared = true
	})
	if got != 0x31 || !cleared {
		t.Fatalf("native result/clear = %#x/%t, want 0x31/true", got, cleared)
	}
	runtime.KeepAlive(candidate)
	runtime.KeepAlive(unit)
}

func TestServerUnitIsDangerous547120ProductionBinding(t *testing.T) {
	s := &Server{
		dangerousUnitTypes547120: dangerousUnitTypeCache547120{
			toxicCloud:      0x21,
			smallToxicCloud: 0x22,
		},
	}
	unit := &Object{ObjSubClass: 0x400}
	if s.UnitIsDangerous547120(&Object{ObjClass: object.ClassFire}, unit) {
		t.Fatal("Fire-immune unit reported dangerous")
	}
	if !s.UnitIsDangerous547120(&Object{ObjClass: object.ClassDangerous, TypeInd: 0x31}, unit) {
		t.Fatal("Dangerous-class object did not report dangerous")
	}
}
