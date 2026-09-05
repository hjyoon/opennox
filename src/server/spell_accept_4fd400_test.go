package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

func spellAcceptTestHooks4FD400(t *testing.T) spellAcceptHooks4FD400[int, int] {
	t.Helper()
	return spellAcceptHooks4FD400[int, int]{
		loadSpellArg:  func() int32 { return 1 },
		loadThirdArg:  func() int { return 3 },
		loadSecondArg: func() int { return 2 },
		loadAcceptArg: func() int { return 5 },
		spellHasFlags: func(int32, uint32) int32 { return 0 },
		loadTarget:    func(int) int { return 0 },
		loadClassLow: func(int) uint8 {
			t.Fatal("unexpected class load")
			return 0
		},
		captureMagic: func(int32, int) int32 {
			t.Fatal("unexpected capture call")
			return 0
		},
		audio: func(int32, int, int32, int32) {
			t.Fatal("unexpected audio call")
		},
		loadLevelArg:  func() int32 { return 6 },
		loadFourthArg: func() int { return 4 },
		tickRate: func() uint32 {
			t.Fatal("unexpected tick-rate load")
			return 0
		},
		plasmaTime: func() float64 {
			t.Fatal("unexpected plasma-time load")
			return 0
		},
		instant: func(int32, int, int, int, int, int32) int32 { return 1 },
		duration: func(spellAcceptDispatch4FD400, int32, int, int, int, int, int32, uint32) int32 {
			t.Fatal("unexpected duration call")
			return 0
		},
	}
}

func TestSpellAcceptDispatch4FD400ExactSelector(t *testing.T) {
	expected := make(map[int32]spellAcceptDispatch4FD400, 133)
	instant := []int32{
		1, 2, 3, 5, 6,
		10, 11, 12, 13, 14, 15, 16, 17, 18, 19,
		23, 26, 27, 29,
		32, 33, 34, 36, 37, 38, 39, 40, 41, 42,
		44, 45, 46, 47, 48, 49, 50, 52, 53, 55, 57, 58,
		60, 61, 62, 63, 64, 65, 66, 68, 69, 70, 71, 72, 74,
		127, 128, 129, 130, 131, 133,
	}
	for _, id := range instant {
		expected[id] = spellAcceptInstant4FD400
	}
	durations := map[spellAcceptDispatch4FD400][]int32{
		spellAcceptDurationBlink4FD400:          {4},
		spellAcceptDurationChannel4FD400:        {8},
		spellAcceptDurationCharm4FD400:          {9},
		spellAcceptDurationTurnUndead4FD400:     {21},
		spellAcceptDurationDrainMana4FD400:      {22},
		spellAcceptDurationLightning4FD400:      {24},
		spellAcceptDurationFirewalk4FD400:       {28},
		spellAcceptDurationForceNature4FD400:    {31},
		spellAcceptDurationGreaterHeal4FD400:    {35},
		spellAcceptDurationChainLightning4FD400: {43},
		spellAcceptDurationShield4FD400:         {51},
		spellAcceptDurationMoonglow4FD400:       {54},
		spellAcceptDurationManaBomb4FD400:       {56},
		spellAcceptDurationPlasma4FD400:         {59},
		spellAcceptDurationOvalShield4FD400:     {67},
		spellAcceptDurationSummon4FD400: {
			75, 76, 77, 78,
			80, 81, 82, 83, 84, 85, 86, 87, 88, 89,
			90, 91, 92, 93, 94, 95, 96, 97, 98, 99,
			100, 101, 102, 103, 104, 105, 106, 107, 108, 109,
			110, 111, 112, 113, 114,
		},
		spellAcceptDurationSwap4FD400:           {115},
		spellAcceptDurationTag4FD400:            {116},
		spellAcceptDurationTeleportMark4FD400:   {117, 118, 119, 120, 122, 123, 124, 125},
		spellAcceptDurationTeleportPop4FD400:    {121},
		spellAcceptDurationTeleportTarget4FD400: {126},
		spellAcceptDurationWall4FD400:           {132},
	}
	for dispatch, ids := range durations {
		for _, id := range ids {
			if old, ok := expected[id]; ok {
				t.Fatalf("spell %d assigned twice: %d and %d", id, old, dispatch)
			}
			expected[id] = dispatch
		}
	}

	counts := map[spellAcceptDispatch4FD400]int{}
	for id := int32(1); id <= 133; id++ {
		want := expected[id]
		got := spellAcceptDispatchFor4FD400(id)
		if got != want {
			t.Errorf("spell %d dispatch = %d, want %d", id, got, want)
		}
		counts[got]++
	}
	if got := counts[spellAcceptInstant4FD400]; got != 60 {
		t.Errorf("instant selector count = %d, want 60", got)
	}
	if got := counts[spellAcceptDefault4FD400]; got != 6 {
		t.Errorf("default selector count = %d, want 6", got)
	}
	for _, id := range []int32{7, 20, 25, 30, 73, 79} {
		if got := spellAcceptDispatchFor4FD400(id); got != spellAcceptDefault4FD400 {
			t.Errorf("in-range hole %d dispatch = %d, want default", id, got)
		}
	}
	for _, id := range []int32{math.MinInt32, -1, 0, 134, math.MaxInt32} {
		if got := spellAcceptDispatchFor4FD400(id); got != spellAcceptDefault4FD400 {
			t.Errorf("out-of-range spell %d dispatch = %d, want default", id, got)
		}
	}
}

func TestSpellAccept4FD400EntryGuardOrder(t *testing.T) {
	tests := []struct {
		name      string
		spellID   int32
		third     int
		second    int
		arg       int
		wantEvent []string
	}{
		{name: "zero spell", third: 3, second: 2, arg: 5, wantEvent: []string{"spell"}},
		{name: "nil third", spellID: 1, second: 2, arg: 5, wantEvent: []string{"spell", "third"}},
		{name: "nil second", spellID: 1, third: 3, arg: 5, wantEvent: []string{"spell", "third", "second"}},
		{name: "nil argument", spellID: 1, third: 3, second: 2, wantEvent: []string{"spell", "third", "second", "arg"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var events []string
			hooks := spellAcceptTestHooks4FD400(t)
			hooks.loadSpellArg = func() int32 { events = append(events, "spell"); return test.spellID }
			hooks.loadThirdArg = func() int { events = append(events, "third"); return test.third }
			hooks.loadSecondArg = func() int { events = append(events, "second"); return test.second }
			hooks.loadAcceptArg = func() int { events = append(events, "arg"); return test.arg }
			hooks.spellHasFlags = func(int32, uint32) int32 {
				t.Fatal("entry rejection reached flags")
				return 0
			}
			if got := spellAccept4FD400(hooks); got != 0 {
				t.Fatalf("result = %d, want 0", got)
			}
			if !reflect.DeepEqual(events, test.wantEvent) {
				t.Fatalf("events = %v, want %v", events, test.wantEvent)
			}
		})
	}
}

func TestSpellAccept4FD400FlagsRequireExactOne(t *testing.T) {
	for _, flagResult := range []int32{0, -1, 2, math.MinInt32, math.MaxInt32} {
		t.Run(fmt.Sprintf("result_%d", flagResult), func(t *testing.T) {
			var events []string
			hooks := spellAcceptTestHooks4FD400(t)
			hooks.loadSpellArg = func() int32 { events = append(events, "spell"); return 7 }
			hooks.loadThirdArg = func() int { events = append(events, "third"); return 3 }
			hooks.loadSecondArg = func() int { events = append(events, "second"); return 2 }
			hooks.loadAcceptArg = func() int { events = append(events, "arg"); return 5 }
			hooks.spellHasFlags = func(id int32, mask uint32) int32 {
				events = append(events, "flags")
				if id != 7 || mask != spellAcceptTargetFlag4FD400 {
					t.Fatalf("flags args = %d/%#x", id, mask)
				}
				return flagResult
			}
			hooks.loadTarget = func(arg int) int {
				events = append(events, "target")
				if arg != 5 {
					t.Fatalf("target arg = %d, want cached 5", arg)
				}
				return 9
			}
			hooks.loadClassLow = func(int) uint8 {
				t.Fatal("non-exact flags result loaded target class")
				return 0
			}
			hooks.captureMagic = func(id int32, target int) int32 {
				events = append(events, "capture")
				if id != 7 || target != 9 {
					t.Fatalf("capture args = %d/%d", id, target)
				}
				return 1
			}
			if got := spellAccept4FD400(hooks); got != 1 {
				t.Fatalf("result = %d, want default success 1", got)
			}
			want := []string{"spell", "third", "second", "arg", "flags", "target", "capture"}
			if !reflect.DeepEqual(events, want) {
				t.Fatalf("events = %v, want %v", events, want)
			}
		})
	}

	var events []string
	hooks := spellAcceptTestHooks4FD400(t)
	hooks.loadSpellArg = func() int32 { events = append(events, "spell"); return 7 }
	hooks.loadThirdArg = func() int { events = append(events, "third"); return 3 }
	hooks.loadSecondArg = func() int { events = append(events, "second"); return 2 }
	hooks.loadAcceptArg = func() int { events = append(events, "arg"); return 5 }
	hooks.spellHasFlags = func(int32, uint32) int32 { events = append(events, "flags"); return 1 }
	hooks.loadTarget = func(int) int { events = append(events, "target"); return 9 }
	hooks.loadClassLow = func(target int) uint8 {
		events = append(events, "class")
		if target != 9 {
			t.Fatalf("class target = %d, want 9", target)
		}
		return 0xf8
	}
	if got := spellAccept4FD400(hooks); got != 0 {
		t.Fatalf("non-unit target result = %d, want 0", got)
	}
	want := []string{"spell", "third", "second", "arg", "flags", "target", "class"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestSpellAccept4FD400TargetReloadsAndCaptureFailureSound(t *testing.T) {
	const (
		firstTarget  = 0x1000
		secondTarget = 0x2000
		thirdTarget  = 0x3000
	)
	target := firstTarget
	loads := 0
	var events []string
	hooks := spellAcceptTestHooks4FD400(t)
	hooks.spellHasFlags = func(int32, uint32) int32 { events = append(events, "flags"); return 1 }
	hooks.loadTarget = func(arg int) int {
		events = append(events, fmt.Sprintf("target:%#x", target))
		if arg != 5 {
			t.Fatalf("target arg = %d, want 5", arg)
		}
		loads++
		got := target
		if loads == 1 {
			target = secondTarget
		}
		return got
	}
	hooks.loadClassLow = func(got int) uint8 {
		events = append(events, "class")
		if got != firstTarget {
			t.Fatalf("class target = %#x, want first %#x", got, firstTarget)
		}
		return spellAcceptUnitMask4FD400
	}
	hooks.captureMagic = func(id int32, got int) int32 {
		events = append(events, "capture")
		if id != 1 || got != secondTarget {
			t.Fatalf("capture args = %d/%#x, want 1/%#x", id, got, secondTarget)
		}
		target = thirdTarget
		return 0
	}
	hooks.audio = func(sound int32, got int, a3, a4 int32) {
		events = append(events, "audio")
		if sound != spellAcceptFizzle4FD400 || got != thirdTarget || a3 != 0 || a4 != 0 {
			t.Fatalf("audio args = %d/%#x/%d/%d", sound, got, a3, a4)
		}
	}
	if got := spellAccept4FD400(hooks); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	want := []string{
		"flags", fmt.Sprintf("target:%#x", firstTarget), "class",
		fmt.Sprintf("target:%#x", secondTarget), "capture",
		fmt.Sprintf("target:%#x", thirdTarget), "audio",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestSpellAccept4FD400InstantCachesObjectsAndReturnsDword(t *testing.T) {
	const (
		oldSecond = 0x1002
		oldThird  = 0x1003
		oldArg    = 0x1005
		newFourth = 0x2004
		level     = int32(0x70000006)
		result    = int32(math.MinInt32)
	)
	second, third, arg := oldSecond, oldThird, oldArg
	fourth := int(4)
	levelArg := int32(6)
	var events []string
	hooks := spellAcceptTestHooks4FD400(t)
	hooks.loadSecondArg = func() int { events = append(events, "second"); return second }
	hooks.loadThirdArg = func() int { events = append(events, "third"); return third }
	hooks.loadAcceptArg = func() int { events = append(events, "arg"); return arg }
	hooks.spellHasFlags = func(int32, uint32) int32 {
		events = append(events, "flags")
		second, third, arg = 12, 13, 15
		fourth, levelArg = newFourth, level
		return 0
	}
	hooks.loadTarget = func(gotArg int) int {
		events = append(events, "target")
		if gotArg != oldArg {
			t.Fatalf("target arg = %#x, want cached %#x", gotArg, oldArg)
		}
		return 0
	}
	hooks.loadLevelArg = func() int32 { events = append(events, "level"); return levelArg }
	hooks.loadFourthArg = func() int { events = append(events, "fourth"); return fourth }
	hooks.instant = func(id int32, gotSecond, gotThird, gotFourth, gotArg int, gotLevel int32) int32 {
		events = append(events, "instant")
		if id != 1 || gotSecond != oldSecond || gotThird != oldThird || gotFourth != newFourth || gotArg != oldArg || gotLevel != level {
			t.Fatalf("instant args = %d/%#x/%#x/%#x/%#x/%#x", id, gotSecond, gotThird, gotFourth, gotArg, gotLevel)
		}
		return result
	}
	if got := spellAccept4FD400(hooks); got != result {
		t.Fatalf("result = %d, want verbatim %d", got, result)
	}
	want := []string{"third", "second", "arg", "flags", "target", "level", "fourth", "instant"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestSpellAccept4FD400InstantZeroSoundsOnCachedFourth(t *testing.T) {
	fourth := 40
	var events []string
	hooks := spellAcceptTestHooks4FD400(t)
	hooks.loadFourthArg = func() int { events = append(events, "fourth"); return fourth }
	hooks.instant = func(int32, int, int, int, int, int32) int32 {
		events = append(events, "instant")
		fourth = 99
		return 0
	}
	hooks.audio = func(sound int32, gotFourth int, a3, a4 int32) {
		events = append(events, "audio")
		if sound != 231 || gotFourth != 40 || a3 != 0 || a4 != 0 {
			t.Fatalf("audio args = %d/%d/%d/%d", sound, gotFourth, a3, a4)
		}
	}
	if got := spellAccept4FD400(hooks); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	if want := []string{"fourth", "instant", "audio"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want suffix %v", events, want)
	}
}

func TestSpellAccept4FD400DurationRoutesAndTiming(t *testing.T) {
	tests := []struct {
		name       string
		spellID    int32
		dispatch   spellAcceptDispatch4FD400
		tickRate   uint32
		plasmaTime float64
		wantTime   uint32
		wantOrder  []string
	}{
		{name: "ordinary", spellID: 4, dispatch: spellAcceptDurationBlink4FD400, wantOrder: []string{"level", "fourth", "duration"}},
		{name: "turn undead fixed", spellID: 21, dispatch: spellAcceptDurationTurnUndead4FD400, wantTime: 70, wantOrder: []string{"level", "fourth", "duration"}},
		{name: "lightning fixed", spellID: 24, dispatch: spellAcceptDurationLightning4FD400, wantTime: 30, wantOrder: []string{"level", "fourth", "duration"}},
		{name: "chain lightning fixed", spellID: 43, dispatch: spellAcceptDurationChainLightning4FD400, wantTime: 30, wantOrder: []string{"level", "fourth", "duration"}},
		{name: "firewalk wrap", spellID: 28, dispatch: spellAcceptDurationFirewalk4FD400, tickRate: 0x80000001, wantTime: 0x80000003, wantOrder: []string{"tick", "fourth", "level", "duration"}},
		{name: "force unsigned wrapped division", spellID: 31, dispatch: spellAcceptDurationForceNature4FD400, tickRate: 0x80000001, wantTime: 0, wantOrder: []string{"tick", "fourth", "level", "duration"}},
		{name: "plasma truncation", spellID: 59, dispatch: spellAcceptDurationPlasma4FD400, plasmaTime: -1.75, wantTime: math.MaxUint32, wantOrder: []string{"plasma", "fourth", "level", "duration"}},
		{name: "plasma invalid", spellID: 59, dispatch: spellAcceptDurationPlasma4FD400, plasmaTime: math.NaN(), wantTime: 0, wantOrder: []string{"plasma", "fourth", "level", "duration"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			const result = int32(math.MinInt32)
			var events []string
			hooks := spellAcceptTestHooks4FD400(t)
			hooks.loadSpellArg = func() int32 { return test.spellID }
			hooks.loadLevelArg = func() int32 { events = append(events, "level"); return math.MaxInt32 }
			hooks.loadFourthArg = func() int { events = append(events, "fourth"); return 4 }
			hooks.tickRate = func() uint32 { events = append(events, "tick"); return test.tickRate }
			hooks.plasmaTime = func() float64 { events = append(events, "plasma"); return test.plasmaTime }
			hooks.duration = func(dispatch spellAcceptDispatch4FD400, id int32, second, third, fourth, arg int, level int32, timeout uint32) int32 {
				events = append(events, "duration")
				if dispatch != test.dispatch || id != test.spellID || second != 2 || third != 3 || fourth != 4 || arg != 5 || level != math.MaxInt32 || timeout != test.wantTime {
					t.Fatalf("duration args = %d/%d/%d/%d/%d/%d/%d/%#x", dispatch, id, second, third, fourth, arg, level, timeout)
				}
				return result
			}
			if got := spellAccept4FD400(hooks); got != result {
				t.Fatalf("result = %d, want verbatim %d", got, result)
			}
			if !reflect.DeepEqual(events, test.wantOrder) {
				t.Fatalf("events = %v, want %v", events, test.wantOrder)
			}
		})
	}
}

func TestSpellAccept4FD400SpecialDurationsReloadTarget(t *testing.T) {
	for _, test := range []struct {
		spellID  int32
		dispatch spellAcceptDispatch4FD400
	}{
		{51, spellAcceptDurationShield4FD400},
		{54, spellAcceptDurationMoonglow4FD400},
		{67, spellAcceptDurationOvalShield4FD400},
	} {
		t.Run(fmt.Sprintf("spell_%d", test.spellID), func(t *testing.T) {
			targets := []int{10, 11, 12}
			var events []string
			hooks := spellAcceptTestHooks4FD400(t)
			hooks.loadSpellArg = func() int32 { return test.spellID }
			hooks.spellHasFlags = func(int32, uint32) int32 { return 1 }
			hooks.loadTarget = func(int) int {
				got := targets[0]
				targets = targets[1:]
				events = append(events, fmt.Sprintf("target:%d", got))
				return got
			}
			hooks.loadClassLow = func(int) uint8 { events = append(events, "class"); return 6 }
			hooks.captureMagic = func(int32, int) int32 { events = append(events, "capture"); return 1 }
			hooks.loadLevelArg = func() int32 { events = append(events, "level"); return 6 }
			hooks.loadFourthArg = func() int { events = append(events, "fourth"); return 4 }
			hooks.duration = func(dispatch spellAcceptDispatch4FD400, _ int32, _, third, _ int, _ int, _ int32, _ uint32) int32 {
				events = append(events, "duration")
				if dispatch != test.dispatch || third != 12 {
					t.Fatalf("duration dispatch/third = %d/%d, want %d/12", dispatch, third, test.dispatch)
				}
				return 0
			}
			if got := spellAccept4FD400(hooks); got != 0 {
				t.Fatalf("result = %d, want duration zero", got)
			}
			want := []string{"target:10", "class", "target:11", "capture", "level", "fourth", "target:12", "duration"}
			if !reflect.DeepEqual(events, want) {
				t.Fatalf("events = %v, want %v", events, want)
			}
		})
	}
}

func TestSpellAccept4FD400DefaultDoesNotLoadCastArguments(t *testing.T) {
	for _, id := range []int32{-1, 7, 20, 25, 30, 73, 79, 134, math.MaxInt32} {
		t.Run(fmt.Sprintf("spell_%d", id), func(t *testing.T) {
			hooks := spellAcceptTestHooks4FD400(t)
			hooks.loadSpellArg = func() int32 { return id }
			hooks.loadLevelArg = func() int32 { t.Fatal("default loaded level"); return 0 }
			hooks.loadFourthArg = func() int { t.Fatal("default loaded fourth object"); return 0 }
			hooks.instant = func(int32, int, int, int, int, int32) int32 { t.Fatal("default called instant"); return 0 }
			hooks.duration = func(spellAcceptDispatch4FD400, int32, int, int, int, int, int32, uint32) int32 {
				t.Fatal("default called duration")
				return 0
			}
			if got := spellAccept4FD400(hooks); got != 1 {
				t.Fatalf("result = %d, want canonical default 1", got)
			}
		})
	}
}

func TestSpellAccept4FD400InstantFaultPrefixes(t *testing.T) {
	want := []string{"spell", "third", "second", "arg", "flags", "target", "capture", "level", "fourth", "instant"}
	for failAt := 1; failAt <= len(want); failAt++ {
		t.Run(fmt.Sprintf("step_%02d", failAt), func(t *testing.T) {
			var events []string
			emit := func(event string) {
				events = append(events, event)
				if len(events) == failAt {
					panic(event)
				}
			}
			hooks := spellAcceptHooks4FD400[int, int]{
				loadSpellArg:  func() int32 { emit("spell"); return 1 },
				loadThirdArg:  func() int { emit("third"); return 3 },
				loadSecondArg: func() int { emit("second"); return 2 },
				loadAcceptArg: func() int { emit("arg"); return 5 },
				spellHasFlags: func(int32, uint32) int32 { emit("flags"); return 0 },
				loadTarget:    func(int) int { emit("target"); return 9 },
				loadClassLow:  func(int) uint8 { t.Fatal("flags-zero path loaded class"); return 0 },
				captureMagic:  func(int32, int) int32 { emit("capture"); return 1 },
				audio: func(int32, int, int32, int32) {
					t.Fatal("nonzero instant path called audio")
				},
				loadLevelArg:  func() int32 { emit("level"); return 6 },
				loadFourthArg: func() int { emit("fourth"); return 4 },
				tickRate:      func() uint32 { t.Fatal("instant loaded tick rate"); return 0 },
				plasmaTime:    func() float64 { t.Fatal("instant loaded plasma time"); return 0 },
				instant: func(int32, int, int, int, int, int32) int32 {
					emit("instant")
					return 1
				},
				duration: func(spellAcceptDispatch4FD400, int32, int, int, int, int, int32, uint32) int32 {
					t.Fatal("instant route called duration")
					return 0
				},
			}
			func() {
				defer func() {
					if recover() == nil {
						t.Fatalf("step %d did not panic", failAt)
					}
				}()
				_ = spellAccept4FD400(hooks)
			}()
			if expected := want[:failAt]; !reflect.DeepEqual(events, expected) {
				t.Fatalf("events = %v, want prefix %v", events, expected)
			}
		})
	}
}
