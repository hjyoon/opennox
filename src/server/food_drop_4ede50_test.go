package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type foodDropTestWorld4EDE50 struct {
	ownerArg     string
	foodArg      string
	pointArg     string
	defaultValue int32
	gameFlag     int32
	fps          uint32
	subClass     map[string]uint32
	flagsLow     map[string]uint16
	rules        []foodDropSoundRule4EDE50
	events       []string
	faultAt      int

	afterDefault  func(*foodDropTestWorld4EDE50)
	afterGameFlag func(*foodDropTestWorld4EDE50)
	afterFPS      func(*foodDropTestWorld4EDE50)
	afterDecay    func(*foodDropTestWorld4EDE50)
}

func newFoodDropTestWorld4EDE50() *foodDropTestWorld4EDE50 {
	return &foodDropTestWorld4EDE50{
		ownerArg:     "owner-a",
		foodArg:      "food-a",
		pointArg:     "point-a",
		defaultValue: 1,
		fps:          30,
		subClass:     map[string]uint32{"food-a": 0},
		flagsLow:     map[string]uint16{"food-a": 0},
		rules:        append([]foodDropSoundRule4EDE50(nil), foodDropSoundRules4EDE50[:]...),
	}
}

func (w *foodDropTestWorld4EDE50) event(value string) {
	w.events = append(w.events, value)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(value)
	}
}

func (w *foodDropTestWorld4EDE50) hooks() foodDropHooks4EDE50[string, string] {
	return foodDropHooks4EDE50[string, string]{
		loadOwnerArg: func() string {
			w.event("owner-arg:" + w.ownerArg)
			return w.ownerArg
		},
		loadFoodArg: func() string {
			w.event("food-arg:" + w.foodArg)
			return w.foodArg
		},
		loadPointArg: func() string {
			w.event("point-arg:" + w.pointArg)
			return w.pointArg
		},
		defaultDrop: func(owner, food, point string) int32 {
			w.event("default:" + owner + ":" + food + ":" + point)
			value := w.defaultValue
			if w.afterDefault != nil {
				w.afterDefault(w)
			}
			return value
		},
		gameFlag: func(flag uint32) int32 {
			value := w.gameFlag
			w.event(fmt.Sprintf("game-flag:%08x=%08x", flag, uint32(value)))
			if w.afterGameFlag != nil {
				w.afterGameFlag(w)
			}
			return value
		},
		loadGameFPS: func() uint32 {
			value := w.fps
			w.event(fmt.Sprintf("fps:%08x", value))
			if w.afterFPS != nil {
				w.afterFPS(w)
			}
			return value
		},
		setDecay: func(food string, delay uint32) {
			w.event(fmt.Sprintf("decay:%s:%08x", food, delay))
			if w.afterDecay != nil {
				w.afterDecay(w)
			}
		},
		loadRuleSound: func(row int) uint16 {
			value := w.rules[row].sound
			w.event(fmt.Sprintf("rule-sound:%d=%04x", row, value))
			return value
		},
		loadSubClass: func(food string) uint32 {
			value := w.subClass[food]
			w.event(fmt.Sprintf("subclass:%s=%08x", food, value))
			return value
		},
		loadRuleSubClassMask: func(row int) uint32 {
			value := w.rules[row].subClassMask
			w.event(fmt.Sprintf("rule-subclass:%d=%08x", row, value))
			return value
		},
		loadRuleFlagsLowMask: func(row int) uint16 {
			value := w.rules[row].flagsLowMask
			w.event(fmt.Sprintf("rule-flags:%d=%04x", row, value))
			return value
		},
		loadFlagsLow: func(food string) uint16 {
			value := w.flagsLow[food]
			w.event(fmt.Sprintf("flags:%s=%04x", food, value))
			return value
		},
		audio: func(sound uint32, owner string, kind int32, code uint32) {
			w.event(fmt.Sprintf("audio:%d:%s:%d:%08x", sound, owner, kind, code))
		},
	}
}

func foodDropEntryEvents4EDE50() []string {
	return []string{
		"owner-arg:owner-a",
		"food-arg:food-a",
		"point-arg:point-a",
		"default:owner-a:food-a:point-a",
	}
}

func appendFoodDropMissRowEvents4EDE50(events []string, row int, sound uint16) []string {
	return append(events,
		fmt.Sprintf("rule-sound:%d=%04x", row, sound),
		"subclass:food-a=00000000",
		fmt.Sprintf("rule-subclass:%d=%08x", row, foodDropSoundRules4EDE50[row].subClassMask),
		fmt.Sprintf("rule-flags:%d=%04x", row, foodDropSoundRules4EDE50[row].flagsLowMask),
		"flags:food-a=0000",
	)
}

func foodDropNoMatchEvents4EDE50(fps uint32) []string {
	events := append(foodDropEntryEvents4EDE50(),
		"game-flag:00000800=00000000",
		fmt.Sprintf("fps:%08x", fps),
		fmt.Sprintf("decay:food-a:%08x", fps*foodDropSeconds4EDE50),
	)
	for row := 0; row < len(foodDropSoundRules4EDE50)-1; row++ {
		events = appendFoodDropMissRowEvents4EDE50(events, row, foodDropSoundRules4EDE50[row].sound)
	}
	return append(events, "rule-sound:4=0000")
}

func foodDropAppleEvents4EDE50() []string {
	events := append(foodDropEntryEvents4EDE50(), "game-flag:00000800=ffffffff")
	events = append(events,
		"rule-sound:0=0343",
		"subclass:food-a=00000002",
		"rule-subclass:0=00000000",
		"rule-flags:0=0001",
		"flags:food-a=0000",
	)
	return append(events,
		"rule-sound:1=0345",
		"subclass:food-a=00000002",
		"rule-subclass:1=00000002",
		"rule-sound:1=0345",
		"audio:837:owner-a:0:00000000",
	)
}

func verifyFoodDropFaultPrefixes4EDE50(
	t *testing.T,
	want []string,
	build func() *foodDropTestWorld4EDE50,
) {
	t.Helper()
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := build()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if !reflect.DeepEqual(w.events, want[:faultAt]) {
					t.Fatalf("events = %v, want %v", w.events, want[:faultAt])
				}
			}()
			foodDrop4EDE50(w.hooks())
		})
	}
}

func TestFoodDrop4EDE50ExactNoMatchTraceFaultPrefixesAndUint32Wrap(t *testing.T) {
	build := func() *foodDropTestWorld4EDE50 {
		w := newFoodDropTestWorld4EDE50()
		w.defaultValue = math.MinInt32
		w.fps = math.MaxUint32
		return w
	}
	want := foodDropNoMatchEvents4EDE50(math.MaxUint32)
	w := build()
	if got := foodDrop4EDE50(w.hooks()); got != math.MinInt32 {
		t.Fatalf("result = %d, want %d", got, int32(math.MinInt32))
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	if want[6] != "decay:food-a:ffffffe7" {
		t.Fatalf("wrapped decay = %q", want[6])
	}
	verifyFoodDropFaultPrefixes4EDE50(t, want, build)
}

func TestFoodDrop4EDE50SoundRuleOrderAndTargetsOwner(t *testing.T) {
	tests := []struct {
		name     string
		subClass uint32
		flagsLow uint16
		sound    uint32
		row      int
	}{
		{name: "below-meat", flagsLow: 0x0001, sound: 835, row: 0},
		{name: "apple", subClass: 0x00000002, sound: 837, row: 1},
		{name: "jug", subClass: 0x00000004, sound: 833, row: 2},
		{name: "mushroom", subClass: 0x00000080, sound: 839, row: 3},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newFoodDropTestWorld4EDE50()
			w.gameFlag = -1
			w.subClass["food-a"] = tc.subClass
			w.flagsLow["food-a"] = tc.flagsLow
			if got := foodDrop4EDE50(w.hooks()); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			wantLast := fmt.Sprintf("audio:%d:owner-a:0:00000000", tc.sound)
			if got := w.events[len(w.events)-1]; got != wantLast {
				t.Fatalf("last event = %q, want %q; all = %v", got, wantLast, w.events)
			}
			matchSound := fmt.Sprintf("rule-sound:%d=%04x", tc.row, tc.sound)
			count := 0
			for _, event := range w.events {
				if event == matchSound {
					count++
				}
			}
			if count != 2 {
				t.Fatalf("matching sound load count = %d, want 2; events = %v", count, w.events)
			}
			if tc.subClass != 0 {
				for _, event := range w.events {
					if event == fmt.Sprintf("rule-flags:%d=0000", tc.row) {
						t.Fatalf("matching subclass unexpectedly read flags: %v", w.events)
					}
				}
			}
		})
	}
}

func TestFoodDrop4EDE50ExactSubclassMatchTraceAndFaultPrefixes(t *testing.T) {
	build := func() *foodDropTestWorld4EDE50 {
		w := newFoodDropTestWorld4EDE50()
		w.defaultValue = -1
		w.gameFlag = -1
		w.subClass["food-a"] = 0x00000002
		return w
	}
	want := foodDropAppleEvents4EDE50()
	w := build()
	if got := foodDrop4EDE50(w.hooks()); got != -1 {
		t.Fatalf("result = %d, want -1", got)
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	verifyFoodDropFaultPrefixes4EDE50(t, want, build)
}

func TestFoodDrop4EDE50NilGuardsUseOriginalLoadOrder(t *testing.T) {
	tests := []struct {
		name      string
		owner     string
		food      string
		point     string
		wantEvent []string
	}{
		{name: "owner", food: "food-a", point: "point-a", wantEvent: []string{"owner-arg:"}},
		{name: "food", owner: "owner-a", point: "point-a", wantEvent: []string{"owner-arg:owner-a", "food-arg:"}},
		{name: "point", owner: "owner-a", food: "food-a", wantEvent: []string{"owner-arg:owner-a", "food-arg:food-a", "point-arg:"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := newFoodDropTestWorld4EDE50()
			w.ownerArg, w.foodArg, w.pointArg = tc.owner, tc.food, tc.point
			if got := foodDrop4EDE50(w.hooks()); got != 0 {
				t.Fatalf("result = %d, want 0", got)
			}
			if !reflect.DeepEqual(w.events, tc.wantEvent) {
				t.Fatalf("events = %v, want %v", w.events, tc.wantEvent)
			}
		})
	}
}

func TestFoodDrop4EDE50DefaultGateAndResultPreservation(t *testing.T) {
	for _, value := range []int32{0, 1, -1, math.MinInt32, math.MaxInt32} {
		t.Run(fmt.Sprintf("%08x", uint32(value)), func(t *testing.T) {
			w := newFoodDropTestWorld4EDE50()
			w.defaultValue = value
			w.gameFlag = 1
			if got := foodDrop4EDE50(w.hooks()); got != value {
				t.Fatalf("result = %d, want %d", got, value)
			}
			if value == 0 && !reflect.DeepEqual(w.events, foodDropEntryEvents4EDE50()) {
				t.Fatalf("zero events = %v, want %v", w.events, foodDropEntryEvents4EDE50())
			}
		})
	}
}

func TestFoodDrop4EDE50GameFlagUsesWholeEAXAndSkipsDecay(t *testing.T) {
	for _, value := range []int32{1, -1, math.MinInt32} {
		t.Run(fmt.Sprintf("%08x", uint32(value)), func(t *testing.T) {
			w := newFoodDropTestWorld4EDE50()
			w.gameFlag = value
			if got := foodDrop4EDE50(w.hooks()); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			for _, event := range w.events {
				if len(event) >= 4 && (event[:4] == "fps:" || event[:4] == "deca") {
					t.Fatalf("decay path executed for nonzero flag: %v", w.events)
				}
			}
		})
	}
}

func TestFoodDrop4EDE50CachesArgumentsAndReadsPostCallbackState(t *testing.T) {
	w := newFoodDropTestWorld4EDE50()
	w.afterDefault = func(w *foodDropTestWorld4EDE50) {
		w.ownerArg = "owner-b"
		w.foodArg = "food-b"
		w.pointArg = "point-b"
	}
	w.afterGameFlag = func(w *foodDropTestWorld4EDE50) {
		w.fps = 0x80000001
	}
	w.afterFPS = func(w *foodDropTestWorld4EDE50) {
		w.fps = 7
	}
	w.afterDecay = func(w *foodDropTestWorld4EDE50) {
		w.subClass["food-a"] = 0x00000080
	}
	if got := foodDrop4EDE50(w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if got := w.events[len(w.events)-1]; got != "audio:839:owner-a:0:00000000" {
		t.Fatalf("last event = %q, want cached owner; all = %v", got, w.events)
	}
	wantDecay := "decay:food-a:80000019"
	found := false
	for _, event := range w.events {
		if event == wantDecay {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing live-FPS/cached-food decay %q: %v", wantDecay, w.events)
	}
}

func TestFoodDrop4EDE50FirstZeroSoundIsImmediateSentinel(t *testing.T) {
	w := newFoodDropTestWorld4EDE50()
	w.gameFlag = 1
	w.defaultValue = -1
	w.rules[0].sound = 0
	if got := foodDrop4EDE50(w.hooks()); got != -1 {
		t.Fatalf("result = %d, want -1", got)
	}
	want := append(foodDropEntryEvents4EDE50(),
		"game-flag:00000800=00000001",
		"rule-sound:0=0000",
	)
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}
