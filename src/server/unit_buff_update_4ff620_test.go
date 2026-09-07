package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

const unitBuffUpdateObject4FF620 = uint64(0x1234567889abcdef)

type unitBuffUpdateWorld4FF620 struct {
	events        []string
	after         map[string]func()
	faultAt       int
	unit          uint64
	buffs         uint32
	fps           uint32
	durations     [32]uint16
	powers        [32]uint8
	flags         uint32
	obj130        uint64
	damageType    uint32
	classLow      uint8
	testResult    int32
	speed         float32
	buffOffResult int32
}

func newUnitBuffUpdateWorld4FF620() *unitBuffUpdateWorld4FF620 {
	return &unitBuffUpdateWorld4FF620{
		after:         make(map[string]func()),
		unit:          unitBuffUpdateObject4FF620,
		fps:           60,
		flags:         0xa5b6c7ff,
		obj130:        0xfedcba9876543210,
		damageType:    math.MaxUint32,
		classLow:      unitBuffUpdateObjectClass4FF620,
		speed:         -3.5,
		buffOffResult: math.MinInt32,
	}
}

func (w *unitBuffUpdateWorld4FF620) observe(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *unitBuffUpdateWorld4FF620) hooks() UnitBuffUpdateHooks4FF620[uint64] {
	return UnitBuffUpdateHooks4FF620[uint64]{
		LoadUnitArg: func() uint64 {
			value := w.unit
			w.observe(fmt.Sprintf("unit=%016x", value))
			return value
		},
		LoadBuffs: func(unit uint64) uint32 {
			value := w.buffs
			w.observe(fmt.Sprintf("buffs:%016x=%08x", unit, value))
			return value
		},
		LoadFPS: func() uint32 {
			value := w.fps
			w.observe(fmt.Sprintf("fps=%08x", value))
			return value
		},
		LoadDuration: func(unit uint64, buff int32) uint16 {
			value := w.durations[int(buff)]
			w.observe(fmt.Sprintf("duration:%016x:%d=%04x", unit, buff, value))
			return value
		},
		StoreDuration: func(unit uint64, buff int32, value uint16) {
			w.observe(fmt.Sprintf("duration:%016x:%d<=%04x", unit, buff, value))
			w.durations[int(buff)] = value
		},
		Audio: func(id int32, unit uint64, kind, code int32) {
			w.observe(fmt.Sprintf("audio:%08x:%016x:%d:%d", uint32(id), unit, kind, code))
		},
		LoadFlags: func(unit uint64) uint32 {
			value := w.flags
			w.observe(fmt.Sprintf("flags:%016x=%08x", unit, value))
			return value
		},
		StoreFlags: func(unit uint64, value uint32) {
			w.observe(fmt.Sprintf("flags:%016x<=%08x", unit, value))
			w.flags = value
		},
		StoreObj130: func(unit, value uint64) {
			w.observe(fmt.Sprintf("obj130:%016x<=%016x", unit, value))
			w.obj130 = value
		},
		StoreDamageType: func(unit uint64, value uint32) {
			w.observe(fmt.Sprintf("damage-type:%016x<=%08x", unit, value))
			w.damageType = value
		},
		DamageClear: func(unit uint64, damage int32) {
			w.observe(fmt.Sprintf("damage:%016x:%08x", unit, uint32(damage)))
		},
		LoadClassLow: func(unit uint64) uint8 {
			value := w.classLow
			w.observe(fmt.Sprintf("class:%016x=%02x", unit, value))
			return value
		},
		IncrementElimDeath: func(unit uint64) {
			w.observe(fmt.Sprintf("increment-death:%016x", unit))
		},
		ReportLesson: func(unit uint64) {
			w.observe(fmt.Sprintf("report-lesson:%016x", unit))
		},
		BuffOff: func(unit uint64, buff int32) int32 {
			value := w.buffOffResult
			w.observe(fmt.Sprintf("buff-off:%016x:%d=%08x", unit, buff, uint32(value)))
			return value
		},
		StorePower: func(unit uint64, buff int32, value uint8) {
			w.observe(fmt.Sprintf("power:%016x:%d<=%02x", unit, buff, value))
			w.powers[int(buff)] = value
		},
		TestBuff: func(unit uint64, buff int32) int32 {
			value := w.testResult
			w.observe(fmt.Sprintf("test-buff:%016x:%d=%08x", unit, buff, uint32(value)))
			return value
		},
		LoadSpeed: func(unit uint64) float32 {
			value := w.speed
			w.observe(fmt.Sprintf("speed:%016x=%08x", unit, math.Float32bits(value)))
			return value
		},
		StoreSpeed: func(unit uint64, value float32) {
			w.observe(fmt.Sprintf("speed:%016x<=%08x", unit, math.Float32bits(value)))
			w.speed = value
		},
	}
}

func unitBuffUpdateAppendMaskLoads4FF620(events []string, buffs uint32, first, last int32) []string {
	for buff := first; buff <= last; buff++ {
		events = append(events, fmt.Sprintf("buffs:%016x=%08x", unitBuffUpdateObject4FF620, buffs))
	}
	return events
}

func TestUnitBuffUpdate4FF620ZeroEntryMaskReturnsImmediately(t *testing.T) {
	w := newUnitBuffUpdateWorld4FF620()
	UnitBuffUpdate4FF620(w.hooks())
	want := []string{
		"unit=1234567889abcdef",
		"buffs:1234567889abcdef=00000000",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want exact entry-gate trace %q", w.events, want)
	}
}

func TestUnitBuffUpdate4FF620PeriodicWarningAndFinalSpeedTrace(t *testing.T) {
	w := newUnitBuffUpdateWorld4FF620()
	w.buffs = uint32(1)<<0 | uint32(1)<<16
	w.durations[0] = 2
	w.durations[16] = 59
	w.testResult = math.MinInt32

	UnitBuffUpdate4FF620(w.hooks())

	want := []string{
		"unit=1234567889abcdef",
		"buffs:1234567889abcdef=00010001",
		"buffs:1234567889abcdef=00010001",
		"duration:1234567889abcdef:0=0002",
		"duration:1234567889abcdef:0<=0001",
	}
	want = unitBuffUpdateAppendMaskLoads4FF620(want, 0x00010001, 1, 15)
	want = append(want,
		"buffs:1234567889abcdef=00010001",
		"fps=0000003c",
		"duration:1234567889abcdef:16=003b",
		"audio:0000001a:1234567889abcdef:0:0",
		"duration:1234567889abcdef:16=003b",
		"duration:1234567889abcdef:16<=003a",
	)
	want = unitBuffUpdateAppendMaskLoads4FF620(want, 0x00010001, 17, 31)
	want = append(want,
		"test-buff:1234567889abcdef:9=80000000",
		"speed:1234567889abcdef=c0600000",
		"speed:1234567889abcdef<=c08c0000",
	)
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want exact warning/speed trace %q", w.events, want)
	}
	if w.durations[0] != 1 || w.durations[16] != 58 || math.Float32bits(w.speed) != 0xc08c0000 {
		t.Fatalf("state = duration0:%d duration16:%d speed:%08x", w.durations[0], w.durations[16], math.Float32bits(w.speed))
	}
}

func TestUnitBuffUpdate4FF620ExpiredFlagClearsLowFlagThenPower(t *testing.T) {
	w := newUnitBuffUpdateWorld4FF620()
	w.buffs = uint32(1) << 7
	w.durations[7] = 1
	w.powers[7] = 0xee
	w.after["buff-off:1234567889abcdef:7=80000000"] = func() {
		w.powers[7] = 0xdd
	}

	UnitBuffUpdate4FF620(w.hooks())

	want := []string{"unit=1234567889abcdef", "buffs:1234567889abcdef=00000080"}
	want = unitBuffUpdateAppendMaskLoads4FF620(want, 0x80, 0, 6)
	want = append(want,
		"buffs:1234567889abcdef=00000080",
		"duration:1234567889abcdef:7=0001",
		"duration:1234567889abcdef:7<=0000",
		"flags:1234567889abcdef=a5b6c7ff",
		"flags:1234567889abcdef<=a5b6c7bf",
		"buff-off:1234567889abcdef:7=80000000",
		"power:1234567889abcdef:7<=00",
	)
	want = unitBuffUpdateAppendMaskLoads4FF620(want, 0x80, 8, 31)
	want = append(want, "test-buff:1234567889abcdef:9=00000000")
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want exact flag expiration trace %q", w.events, want)
	}
	if w.flags != 0xa5b6c7bf || w.durations[7] != 0 || w.powers[7] != 0 {
		t.Fatalf("state = flags:%08x duration:%04x power:%02x", w.flags, w.durations[7], w.powers[7])
	}
}

func TestUnitBuffUpdate4FF620ExpiredDeathUsesNativeTokenAndLiveClass(t *testing.T) {
	w := newUnitBuffUpdateWorld4FF620()
	w.buffs = uint32(1) << 16
	w.durations[16] = 1
	w.powers[16] = 0xff
	w.classLow = 0
	w.after["audio:0000030b:1234567889abcdef:0:0"] = func() {
		w.classLow = unitBuffUpdateObjectClass4FF620
	}

	UnitBuffUpdate4FF620(w.hooks())

	want := []string{"unit=1234567889abcdef", "buffs:1234567889abcdef=00010000"}
	want = unitBuffUpdateAppendMaskLoads4FF620(want, 0x10000, 0, 15)
	want = append(want,
		"buffs:1234567889abcdef=00010000",
		"fps=0000003c",
		"duration:1234567889abcdef:16=0001",
		"duration:1234567889abcdef:16=0001",
		"duration:1234567889abcdef:16<=0000",
		"obj130:1234567889abcdef<=0000000000000000",
		"damage-type:1234567889abcdef<=0000000d",
		"damage:1234567889abcdef:0098967f",
		"audio:0000030b:1234567889abcdef:0:0",
		"class:1234567889abcdef=04",
		"increment-death:1234567889abcdef",
		"report-lesson:1234567889abcdef",
		"buff-off:1234567889abcdef:16=80000000",
		"power:1234567889abcdef:16<=00",
	)
	want = unitBuffUpdateAppendMaskLoads4FF620(want, 0x10000, 17, 31)
	want = append(want, "test-buff:1234567889abcdef:9=00000000")
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want exact death expiration trace %q", w.events, want)
	}
	if w.obj130 != 0 || w.damageType != 13 || w.durations[16] != 0 || w.powers[16] != 0 {
		t.Fatalf("state = obj130:%016x type:%08x duration:%04x power:%02x", w.obj130, w.damageType, w.durations[16], w.powers[16])
	}
}

func TestUnitBuffUpdate4FF620WarningCallbackCanReplaceDuration(t *testing.T) {
	w := newUnitBuffUpdateWorld4FF620()
	w.buffs = uint32(1) << 16
	w.durations[16] = 59
	w.classLow = 0
	w.after["audio:0000001a:1234567889abcdef:0:0"] = func() {
		w.durations[16] = 1
	}

	UnitBuffUpdate4FF620(w.hooks())

	wantSubsequence := []string{
		"fps=0000003c",
		"duration:1234567889abcdef:16=003b",
		"audio:0000001a:1234567889abcdef:0:0",
		"duration:1234567889abcdef:16=0001",
		"duration:1234567889abcdef:16<=0000",
		"obj130:1234567889abcdef<=0000000000000000",
	}
	for i, event := range wantSubsequence {
		got := w.events[19+i]
		if got != event {
			t.Fatalf("event %d = %q, want live-duration subsequence %q", 19+i, got, wantSubsequence)
		}
	}
}

func TestUnitBuffUpdate4FF620ReloadsMaskForEveryLaterSlot(t *testing.T) {
	w := newUnitBuffUpdateWorld4FF620()
	w.buffs = uint32(1) << 7
	w.durations[7] = 1
	w.durations[16] = 2
	w.after["buff-off:1234567889abcdef:7=80000000"] = func() {
		w.buffs = uint32(1) << 16
	}

	UnitBuffUpdate4FF620(w.hooks())

	var gotDuration16 bool
	for _, event := range w.events {
		if event == "duration:1234567889abcdef:16=0002" {
			gotDuration16 = true
			break
		}
	}
	if !gotDuration16 || w.durations[16] != 1 {
		t.Fatalf("later live mask did not enable slot 16: events=%q duration=%d", w.events, w.durations[16])
	}
	if got := w.events[len(w.events)-1]; got != "test-buff:1234567889abcdef:9=00000000" {
		t.Fatalf("final event = %q, want unconditional post-loop buff test", got)
	}
}

func TestUnitBuffUpdate4FF620ZeroFPSFaultsAfterSingleFPSLoad(t *testing.T) {
	w := newUnitBuffUpdateWorld4FF620()
	w.buffs = uint32(1) << 16
	w.durations[16] = 59
	w.fps = 0

	var recovered any
	func() {
		defer func() { recovered = recover() }()
		UnitBuffUpdate4FF620(w.hooks())
	}()
	if recovered == nil {
		t.Fatal("zero GAME.EXE FPS divisor did not fault")
	}
	want := []string{"unit=1234567889abcdef", "buffs:1234567889abcdef=00010000"}
	want = unitBuffUpdateAppendMaskLoads4FF620(want, 0x10000, 0, 16)
	want = append(want, "fps=00000000", "duration:1234567889abcdef:16=003b")
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want exact divide-fault prefix %q", w.events, want)
	}
}

func TestUnitBuffUpdate4FF620EveryObservableFaultHasExactPrefix(t *testing.T) {
	baseline := newUnitBuffUpdateWorld4FF620()
	baseline.buffs = uint32(1) << 16
	baseline.durations[16] = 1
	baseline.testResult = 1
	UnitBuffUpdate4FF620(baseline.hooks())
	if len(baseline.events) < 40 {
		t.Fatalf("baseline trace unexpectedly short: %q", baseline.events)
	}

	for faultAt := 1; faultAt <= len(baseline.events); faultAt++ {
		w := newUnitBuffUpdateWorld4FF620()
		w.buffs = uint32(1) << 16
		w.durations[16] = 1
		w.testResult = 1
		w.faultAt = faultAt

		var recovered any
		func() {
			defer func() { recovered = recover() }()
			UnitBuffUpdate4FF620(w.hooks())
		}()
		if recovered == nil {
			t.Fatalf("faultAt %d did not panic", faultAt)
		}
		want := baseline.events[:faultAt]
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("faultAt %d events = %q, want exact prefix %q", faultAt, w.events, want)
		}
	}
}
