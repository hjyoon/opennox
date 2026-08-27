package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type pickupFoodTestObject4F3350 struct {
	name        string
	subClass    uint32
	materialLow uint16
	flagsLow    uint8
	use         string
}

type pickupFoodTestWorld4F3350 struct {
	playerState  int32
	defaultValue int32
	rules        []pickupFoodSoundRule4F3350
	events       []string
	faultAt      int
	afterUse     func(*pickupFoodTestWorld4F3350)
	afterDefault func(*pickupFoodTestWorld4F3350)
}

func newPickupFoodTestWorld4F3350() *pickupFoodTestWorld4F3350 {
	return &pickupFoodTestWorld4F3350{
		defaultValue: 1,
		rules:        append([]pickupFoodSoundRule4F3350(nil), pickupFoodSoundRules4F3350[:]...),
	}
}

func pickupFoodTestName4F3350(obj *pickupFoodTestObject4F3350) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func (w *pickupFoodTestWorld4F3350) event(value string) {
	w.events = append(w.events, value)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(value)
	}
}

func (w *pickupFoodTestWorld4F3350) hooks() pickupFoodHooks4F3350[*pickupFoodTestObject4F3350, string] {
	return pickupFoodHooks4F3350[*pickupFoodTestObject4F3350, string]{
		playerState: func(owner *pickupFoodTestObject4F3350) int32 {
			w.event(fmt.Sprintf("player-state:%s=%08x", pickupFoodTestName4F3350(owner), uint32(w.playerState)))
			return w.playerState
		},
		loadSubClassLow: func(item *pickupFoodTestObject4F3350) uint8 {
			value := uint8(item.subClass)
			w.event(fmt.Sprintf("subclass-low:%s=%02x", pickupFoodTestName4F3350(item), value))
			return value
		},
		loadUse: func(item *pickupFoodTestObject4F3350) string {
			w.event("load-use:" + pickupFoodTestName4F3350(item) + "=" + item.use)
			return item.use
		},
		callUse: func(use string, owner, item *pickupFoodTestObject4F3350) int32 {
			w.event("call-use:" + use + ":" + pickupFoodTestName4F3350(owner) + ":" + pickupFoodTestName4F3350(item))
			if w.afterUse != nil {
				w.afterUse(w)
			}
			return math.MinInt32
		},
		loadFlagsLow: func(item *pickupFoodTestObject4F3350) uint8 {
			value := item.flagsLow
			w.event(fmt.Sprintf("flags-low:%s=%02x", pickupFoodTestName4F3350(item), value))
			return value
		},
		defaultPickup: func(owner, item *pickupFoodTestObject4F3350, arg3, arg4 int32) int32 {
			w.event(fmt.Sprintf("default:%s:%s:%08x:%08x", pickupFoodTestName4F3350(owner), pickupFoodTestName4F3350(item), uint32(arg3), uint32(arg4)))
			value := w.defaultValue
			if w.afterDefault != nil {
				w.afterDefault(w)
			}
			return value
		},
		loadRuleSound: func(row int) uint16 {
			value := w.rules[row].sound
			w.event(fmt.Sprintf("rule-sound:%d=%04x", row, value))
			return value
		},
		loadSubClass: func(item *pickupFoodTestObject4F3350) uint32 {
			value := item.subClass
			w.event(fmt.Sprintf("subclass:%s=%08x", pickupFoodTestName4F3350(item), value))
			return value
		},
		loadRuleSubClassMask: func(row int) uint32 {
			value := w.rules[row].subClassMask
			w.event(fmt.Sprintf("rule-subclass:%d=%08x", row, value))
			return value
		},
		loadRuleMaterialMask: func(row int) uint16 {
			value := w.rules[row].materialMask
			w.event(fmt.Sprintf("rule-material:%d=%04x", row, value))
			return value
		},
		loadMaterialLow: func(item *pickupFoodTestObject4F3350) uint16 {
			value := item.materialLow
			w.event(fmt.Sprintf("material:%s=%04x", pickupFoodTestName4F3350(item), value))
			return value
		},
		audio: func(sound uint32, owner *pickupFoodTestObject4F3350, kind int32, code uint32) {
			w.event(fmt.Sprintf("audio:%d:%s:%d:%08x", sound, pickupFoodTestName4F3350(owner), kind, code))
		},
	}
}

func pickupFoodAppleTrace4F3350() []string {
	return []string{
		"player-state:owner=ffffffff",
		"flags-low:item=00",
		"default:owner:item:80000000:7fffffff",
		"rule-sound:0=0342",
		"subclass:item=00000002",
		"rule-subclass:0=00000000",
		"rule-material:0=0001",
		"material:item=0000",
		"rule-sound:1=0344",
		"subclass:item=00000002",
		"rule-subclass:1=00000002",
		"rule-sound:1=0344",
		"audio:836:owner:0:00000000",
	}
}

func verifyPickupFoodFaultPrefixes4F3350(t *testing.T, want []string, build func() (*pickupFoodTestWorld4F3350, *pickupFoodTestObject4F3350, *pickupFoodTestObject4F3350)) {
	t.Helper()
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w, owner, item := build()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if !reflect.DeepEqual(w.events, want[:faultAt]) {
					t.Fatalf("events = %v, want %v", w.events, want[:faultAt])
				}
			}()
			pickupFood4F3350(owner, item, math.MinInt32, math.MaxInt32, w.hooks())
		})
	}
}

func TestPickupFood4F3350ExactAppleTraceResultAndFaultPrefixes(t *testing.T) {
	build := func() (*pickupFoodTestWorld4F3350, *pickupFoodTestObject4F3350, *pickupFoodTestObject4F3350) {
		w := newPickupFoodTestWorld4F3350()
		w.playerState = -1
		w.defaultValue = math.MinInt32
		return w,
			&pickupFoodTestObject4F3350{name: "owner"},
			&pickupFoodTestObject4F3350{name: "item", subClass: 2, use: "eat"}
	}
	w, owner, item := build()
	if got := pickupFood4F3350(owner, item, math.MinInt32, math.MaxInt32, w.hooks()); got != math.MinInt32 {
		t.Fatalf("result = %d, want %d", got, int32(math.MinInt32))
	}
	want := pickupFoodAppleTrace4F3350()
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	verifyPickupFoodFaultPrefixes4F3350(t, want, build)
}

func TestPickupFood4F3350NilGuardsHaveNoObservableReads(t *testing.T) {
	owner := &pickupFoodTestObject4F3350{name: "owner"}
	item := &pickupFoodTestObject4F3350{name: "item"}
	for _, tc := range []struct {
		name        string
		owner, item *pickupFoodTestObject4F3350
	}{
		{name: "owner", item: item},
		{name: "item", owner: owner},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := newPickupFoodTestWorld4F3350()
			if got := pickupFood4F3350(tc.owner, tc.item, -1, -2, w.hooks()); got != 0 {
				t.Fatalf("result = %d, want 0", got)
			}
			if len(w.events) != 0 {
				t.Fatalf("events = %v, want none", w.events)
			}
		})
	}
}

func TestPickupFood4F3350UseGateAndDestroyedShortCircuit(t *testing.T) {
	owner := &pickupFoodTestObject4F3350{name: "owner"}
	item := &pickupFoodTestObject4F3350{name: "item", subClass: 0x80000100, use: "eat"}
	w := newPickupFoodTestWorld4F3350()
	w.defaultValue = 0
	w.afterUse = func(*pickupFoodTestWorld4F3350) {
		item.flagsLow = pickupFoodDestroyedFlagsLow4F3350 | 0x80
	}
	if got := pickupFood4F3350(owner, item, 7, 9, w.hooks()); got != 1 {
		t.Fatalf("result = %d, want canonical destroyed success", got)
	}
	want := []string{
		"player-state:owner=00000000",
		"subclass-low:item=00",
		"load-use:item=eat",
		"call-use:eat:owner:item",
		"flags-low:item=a0",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestPickupFood4F3350UseBypassReadsOnlyLowSubclass(t *testing.T) {
	for _, subClass := range []uint32{0x04, 0x80, 0x84, 0x80000104} {
		t.Run(fmt.Sprintf("%08x", subClass), func(t *testing.T) {
			owner := &pickupFoodTestObject4F3350{name: "owner"}
			item := &pickupFoodTestObject4F3350{name: "item", subClass: subClass, use: "must-not-run"}
			w := newPickupFoodTestWorld4F3350()
			w.defaultValue = 0
			if got := pickupFood4F3350(owner, item, 1, 2, w.hooks()); got != 0 {
				t.Fatalf("result = %d, want 0", got)
			}
			for _, event := range w.events {
				if event == "load-use:item=must-not-run" {
					t.Fatalf("bypass loaded Use: %v", w.events)
				}
			}
		})
	}
}

func TestPickupFood4F3350NonzeroPlayerStateSkipsSubclassAndUse(t *testing.T) {
	owner := &pickupFoodTestObject4F3350{name: "owner"}
	item := &pickupFoodTestObject4F3350{name: "item", use: "must-not-run"}
	w := newPickupFoodTestWorld4F3350()
	w.playerState = math.MinInt32
	w.defaultValue = 0
	if got := pickupFood4F3350(owner, item, 1, 2, w.hooks()); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	want := []string{
		"player-state:owner=80000000",
		"flags-low:item=00",
		"default:owner:item:00000001:00000002",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestPickupFood4F3350DefaultGateFourArgsAndExactResult(t *testing.T) {
	for _, value := range []int32{0, 1, -1, math.MinInt32, math.MaxInt32} {
		t.Run(fmt.Sprintf("%08x", uint32(value)), func(t *testing.T) {
			owner := &pickupFoodTestObject4F3350{name: "owner"}
			item := &pickupFoodTestObject4F3350{name: "item"}
			w := newPickupFoodTestWorld4F3350()
			w.playerState = 1
			w.defaultValue = value
			w.rules[0].sound = 0
			if got := pickupFood4F3350(owner, item, math.MinInt32, math.MaxInt32, w.hooks()); got != value {
				t.Fatalf("result = %d, want %d", got, value)
			}
			wantCount := 3
			if value != 0 {
				wantCount = 4
			}
			if len(w.events) != wantCount {
				t.Fatalf("events = %v, want %d events", w.events, wantCount)
			}
		})
	}
}

func TestPickupFood4F3350SoundRuleOrderSubclassMaterialAndSentinel(t *testing.T) {
	for _, tc := range []struct {
		name        string
		subClass    uint32
		materialLow uint16
		sound       uint32
		row         int
	}{
		{name: "flesh", materialLow: 1, sound: 834, row: 0},
		{name: "apple", subClass: 2, sound: 836, row: 1},
		{name: "jug", subClass: 4, sound: 832, row: 2},
		{name: "mushroom", subClass: 0x80, sound: 838, row: 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			owner := &pickupFoodTestObject4F3350{name: "owner"}
			item := &pickupFoodTestObject4F3350{name: "item", subClass: tc.subClass, materialLow: tc.materialLow}
			w := newPickupFoodTestWorld4F3350()
			w.playerState = 1
			w.defaultValue = -1
			if got := pickupFood4F3350(owner, item, 3, 4, w.hooks()); got != -1 {
				t.Fatalf("result = %d, want -1", got)
			}
			wantLast := fmt.Sprintf("audio:%d:owner:0:00000000", tc.sound)
			if got := w.events[len(w.events)-1]; got != wantLast {
				t.Fatalf("last event = %q, want %q; all = %v", got, wantLast, w.events)
			}
			matchedSound := fmt.Sprintf("rule-sound:%d=%04x", tc.row, tc.sound)
			count := 0
			for _, event := range w.events {
				if event == matchedSound {
					count++
				}
			}
			if count != 2 {
				t.Fatalf("matching sound loads = %d, want 2; events = %v", count, w.events)
			}
			if tc.subClass != 0 {
				for _, event := range w.events {
					if event == fmt.Sprintf("rule-material:%d=0000", tc.row) {
						t.Fatalf("subclass match read material: %v", w.events)
					}
				}
			}
		})
	}

	owner := &pickupFoodTestObject4F3350{name: "owner"}
	item := &pickupFoodTestObject4F3350{name: "item"}
	w := newPickupFoodTestWorld4F3350()
	w.playerState = 1
	w.defaultValue = math.MaxInt32
	if got := pickupFood4F3350(owner, item, 0, 0, w.hooks()); got != math.MaxInt32 {
		t.Fatalf("no-match result = %d", got)
	}
	if last := w.events[len(w.events)-1]; last != "rule-sound:4=0000" {
		t.Fatalf("sentinel event = %q; all = %v", last, w.events)
	}
	for _, event := range w.events {
		if event == "rule-subclass:4=00000000" || event == "rule-material:4=0000" {
			t.Fatalf("sentinel masks were read: %v", w.events)
		}
	}
}

func TestPickupFood4F3350ReadsLiveFieldsAndReloadsMatchedSound(t *testing.T) {
	owner := &pickupFoodTestObject4F3350{name: "owner"}
	item := &pickupFoodTestObject4F3350{name: "item"}
	w := newPickupFoodTestWorld4F3350()
	w.playerState = 1
	w.defaultValue = -7
	w.afterDefault = func(w *pickupFoodTestWorld4F3350) {
		item.subClass = 2
		w.rules[1].sound = 900
	}
	hooks := w.hooks()
	originalLoadSound := hooks.loadRuleSound
	rowOneLoads := 0
	hooks.loadRuleSound = func(row int) uint16 {
		value := originalLoadSound(row)
		if row == 1 {
			rowOneLoads++
			if rowOneLoads == 1 {
				w.rules[1].sound = 901
			}
		}
		return value
	}
	if got := pickupFood4F3350(owner, item, 5, 6, hooks); got != -7 {
		t.Fatalf("result = %d, want -7", got)
	}
	if last := w.events[len(w.events)-1]; last != "audio:901:owner:0:00000000" {
		t.Fatalf("last event = %q, want live reloaded sound; all = %v", last, w.events)
	}
}
