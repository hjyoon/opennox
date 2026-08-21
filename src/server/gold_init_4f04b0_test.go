package server

import (
	"math"
	"reflect"
	"testing"
)

type goldInitTestData4F04B0 struct {
	amount uint32
	guard  uint32
}

type goldInitTestObject4F04B0 struct {
	initData   *goldInitTestData4F04B0
	experience float32
}

type goldInitTestPlayer4F04B0 struct {
	name string
	unit *goldInitTestObject4F04B0
	next *goldInitTestPlayer4F04B0
}

func TestGoldInit4F04B0NonzeroAmountReturnsEntryResultAndStops(t *testing.T) {
	entryData := &goldInitTestData4F04B0{amount: 0x01020304, guard: 0xa5a5a5a5}
	replacement := &goldInitTestData4F04B0{amount: 0, guard: 0x5a5a5a5a}
	unit := &goldInitTestObject4F04B0{initData: entryData}
	events := make([]string, 0, 3)
	const entryResult = int32(-0x1234567)

	got := goldInit4F04B0(goldInitHooks4F04B0[
		*goldInitTestObject4F04B0,
		*goldInitTestData4F04B0,
		*goldInitTestPlayer4F04B0,
	]{
		loadUnitArg: func() (*goldInitTestObject4F04B0, int32) {
			events = append(events, "load-unit")
			return unit, entryResult
		},
		loadInitData: func(got *goldInitTestObject4F04B0) *goldInitTestData4F04B0 {
			events = append(events, "load-init")
			return got.initData
		},
		loadAmount: func(data *goldInitTestData4F04B0) uint32 {
			events = append(events, "load-amount")
			unit.initData = replacement
			return data.amount
		},
		firstPlayer: func() *goldInitTestPlayer4F04B0 {
			t.Fatal("nonzero Amount reached player traversal")
			return nil
		},
		truncQwordLow: func(float64) int32 {
			t.Fatal("nonzero Amount reached x87 truncation")
			return 0
		},
		randomInt: func(int32, int32, string, int32) int32 {
			t.Fatal("nonzero Amount reached RNG")
			return 0
		},
		storeAmount: func(*goldInitTestData4F04B0, uint32) {
			t.Fatal("nonzero Amount reached store")
		},
	})

	if got != entryResult {
		t.Fatalf("return = %#x, want entry result %#x", got, entryResult)
	}
	if want := []string{"load-unit", "load-init", "load-amount"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if *entryData != (goldInitTestData4F04B0{amount: 0x01020304, guard: 0xa5a5a5a5}) {
		t.Fatalf("cached entry data changed: %+v", *entryData)
	}
	if unit.initData != replacement {
		t.Fatal("live InitData mutation was lost")
	}
}

func TestGoldInit4F04B0CountsNilUnitsCachesDataAndPreservesOrder(t *testing.T) {
	entryData := &goldInitTestData4F04B0{guard: 0xa5a5a5a5}
	replacement := &goldInitTestData4F04B0{amount: 0x11111111, guard: 0x5a5a5a5a}
	gold := &goldInitTestObject4F04B0{initData: entryData}
	firstUnit := &goldInitTestObject4F04B0{experience: 1 << 24}
	thirdUnit := &goldInitTestObject4F04B0{experience: 1}
	third := &goldInitTestPlayer4F04B0{name: "p3", unit: thirdUnit}
	second := &goldInitTestPlayer4F04B0{name: "p2", next: third}
	first := &goldInitTestPlayer4F04B0{name: "p1", unit: firstUnit, next: second}
	events := make([]string, 0, 20)
	wantAverage := goldInitAverage4F04B0(1<<24, 3)
	wantScales := []float64{
		goldInitScale4F04B0(wantAverage, goldInitUpperScaleBits4F04B0),
		goldInitScale4F04B0(wantAverage, goldInitLowerScaleBits4F04B0),
		goldInitScale4F04B0(wantAverage, goldInitNegativeScaleBits4F04B0),
	}
	truncCall := 0
	randomCall := 0

	got := goldInit4F04B0(goldInitHooks4F04B0[
		*goldInitTestObject4F04B0,
		*goldInitTestData4F04B0,
		*goldInitTestPlayer4F04B0,
	]{
		loadUnitArg: func() (*goldInitTestObject4F04B0, int32) {
			events = append(events, "load-unit")
			return gold, -1
		},
		loadInitData: func(unit *goldInitTestObject4F04B0) *goldInitTestData4F04B0 {
			events = append(events, "load-init")
			return unit.initData
		},
		loadAmount: func(data *goldInitTestData4F04B0) uint32 {
			events = append(events, "load-amount")
			return data.amount
		},
		firstPlayer: func() *goldInitTestPlayer4F04B0 {
			events = append(events, "first-player")
			return first
		},
		loadPlayerUnit: func(player *goldInitTestPlayer4F04B0) *goldInitTestObject4F04B0 {
			events = append(events, "load-player-unit:"+player.name)
			return player.unit
		},
		loadExperience: func(unit *goldInitTestObject4F04B0) float32 {
			name := "u3"
			if unit == firstUnit {
				name = "u1"
			}
			events = append(events, "load-experience:"+name)
			return unit.experience
		},
		nextPlayer: func(player *goldInitTestPlayer4F04B0) *goldInitTestPlayer4F04B0 {
			events = append(events, "next-player:"+player.name)
			return player.next
		},
		truncQwordLow: func(value float64) int32 {
			name := []string{"upper", "lower", "negative"}[truncCall]
			events = append(events, "trunc-"+name)
			if math.Float64bits(value) != math.Float64bits(wantScales[truncCall]) {
				t.Fatalf("%s trunc input = %016x, want %016x", name, math.Float64bits(value), math.Float64bits(wantScales[truncCall]))
			}
			values := []int32{40, 20, -40}
			result := values[truncCall]
			truncCall++
			return result
		},
		randomInt: func(minimum, maximum int32, path string, line int32) int32 {
			randomCall++
			switch randomCall {
			case 1:
				events = append(events, "random-scaled")
				if minimum != 20 || maximum != 40 || path != goldInitScaledRandomPath4F04B0 || line != 1017 {
					t.Fatalf("scaled RNG args = %d/%d/%q/%d", minimum, maximum, path, line)
				}
				gold.initData = replacement
				return math.MaxInt32
			case 2:
				events = append(events, "random-base")
				if minimum != 15 || maximum != 30 || path != goldInitBaseRandomPath4F04B0 || line != 1018 {
					t.Fatalf("base RNG args = %d/%d/%q/%d", minimum, maximum, path, line)
				}
				return 30
			default:
				t.Fatalf("unexpected RNG call %d", randomCall)
				return 0
			}
		},
		storeAmount: func(data *goldInitTestData4F04B0, amount uint32) {
			events = append(events, "store-amount")
			data.amount = amount
		},
	})

	if got != 30 {
		t.Fatalf("return = %d, want full base RNG result 30", got)
	}
	wantEvents := []string{
		"load-unit", "load-init", "load-amount", "first-player",
		"load-player-unit:p1", "load-experience:u1", "next-player:p1",
		"load-player-unit:p2", "next-player:p2",
		"load-player-unit:p3", "load-experience:u3", "next-player:p3",
		"trunc-upper", "trunc-lower", "random-scaled", "trunc-negative",
		"random-base", "store-amount",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	scaledRandom := int32(math.MaxInt32)
	negative := int32(-40)
	wantAmount := uint32(scaledRandom) - uint32(negative) + 30
	if entryData.amount != wantAmount {
		t.Fatalf("cached Amount = %#x", entryData.amount)
	}
	if entryData.guard != 0xa5a5a5a5 || replacement.amount != 0x11111111 || replacement.guard != 0x5a5a5a5a {
		t.Fatalf("records changed unexpectedly: entry=%+v replacement=%+v", *entryData, *replacement)
	}
	if gold.initData != replacement {
		t.Fatal("RNG mutation of live InitData was lost")
	}
}

func TestGoldInit4F04B0EmptyPlayerListUsesNaNConversionsAsZero(t *testing.T) {
	data := &goldInitTestData4F04B0{}
	unit := &goldInitTestObject4F04B0{initData: data}
	truncCalls := 0
	randomCalls := 0
	got := goldInit4F04B0(goldInitHooks4F04B0[
		*goldInitTestObject4F04B0,
		*goldInitTestData4F04B0,
		*goldInitTestPlayer4F04B0,
	]{
		loadUnitArg:  func() (*goldInitTestObject4F04B0, int32) { return unit, -1 },
		loadInitData: func(unit *goldInitTestObject4F04B0) *goldInitTestData4F04B0 { return unit.initData },
		loadAmount:   func(data *goldInitTestData4F04B0) uint32 { return data.amount },
		firstPlayer:  func() *goldInitTestPlayer4F04B0 { return nil },
		truncQwordLow: func(value float64) int32 {
			truncCalls++
			if !math.IsNaN(value) {
				t.Fatalf("truncation %d input = %v, want NaN", truncCalls, value)
			}
			return goldInitTruncQwordLow4F04B0(value)
		},
		randomInt: func(minimum, maximum int32, path string, line int32) int32 {
			randomCalls++
			if randomCalls == 1 {
				if minimum != 0 || maximum != 0 || path != goldInitScaledRandomPath4F04B0 || line != 1017 {
					t.Fatalf("scaled RNG args = %d/%d/%q/%d", minimum, maximum, path, line)
				}
				return 0
			}
			if minimum != 15 || maximum != 30 || path != goldInitBaseRandomPath4F04B0 || line != 1018 {
				t.Fatalf("base RNG args = %d/%d/%q/%d", minimum, maximum, path, line)
			}
			return 23
		},
		storeAmount: func(data *goldInitTestData4F04B0, amount uint32) { data.amount = amount },
	})
	if got != 23 || data.amount != 23 || truncCalls != 3 || randomCalls != 2 {
		t.Fatalf("return/amount/trunc/RNG = %d/%d/%d/%d", got, data.amount, truncCalls, randomCalls)
	}
}

func TestGoldInit4F04B0Binary32SpillsAndTruncation(t *testing.T) {
	sum := goldInitAddExperience4F04B0(0, 1<<24)
	sum = goldInitAddExperience4F04B0(sum, 1)
	sum = goldInitAddExperience4F04B0(sum, -(1 << 24))
	if math.Float32bits(sum) != 0 {
		t.Fatalf("sequentially spilled sum bits = %08x, want +0", math.Float32bits(sum))
	}
	if !math.IsNaN(float64(goldInitAverage4F04B0(0, 0))) {
		t.Fatal("zero-player average is not NaN")
	}

	tests := []struct {
		name  string
		value float64
		want  int32
	}{
		{"positive", 123.999, 123},
		{"negative", -123.999, -123},
		{"positive-low-dword", 0x100000001, 1},
		{"negative-low-dword", -0x100000001, -1},
		{"largest-finite-qword-side", math.Nextafter(0x1p63, 0), -1024},
		{"positive-overflow", 0x1p63, 0},
		{"negative-limit", -0x1p63, 0},
		{"negative-overflow", math.Nextafter(-0x1p63, math.Inf(-1)), 0},
		{"positive-infinity", math.Inf(1), 0},
		{"negative-infinity", math.Inf(-1), 0},
		{"nan", math.NaN(), 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := goldInitTruncQwordLow4F04B0(tc.value); got != tc.want {
				t.Fatalf("trunc(%v) = %d (%#x), want %d (%#x)", tc.value, got, uint32(got), tc.want, uint32(tc.want))
			}
		})
	}
	if math.Float64bits(math.Float64frombits(goldInitLowerScaleBits4F04B0)) != goldInitLowerScaleBits4F04B0 ||
		math.Float64bits(math.Float64frombits(goldInitUpperScaleBits4F04B0)) != goldInitUpperScaleBits4F04B0 ||
		math.Float64bits(math.Float64frombits(goldInitNegativeScaleBits4F04B0)) != goldInitNegativeScaleBits4F04B0 {
		t.Fatal("sealed scale bit patterns changed")
	}
}

func TestGoldInit4F04B0EveryObservableFaultPrefix(t *testing.T) {
	allEvents := []string{
		"load-unit", "load-init", "load-amount", "first-player",
		"load-player-unit:p1", "load-experience:u1", "next-player:p1",
		"load-player-unit:p2", "next-player:p2",
		"trunc-upper", "trunc-lower", "random-scaled", "trunc-negative",
		"random-base", "store-amount",
	}
	for faultIndex, fault := range allEvents {
		t.Run(fault, func(t *testing.T) {
			data := &goldInitTestData4F04B0{guard: 0xa5a5a5a5}
			gold := &goldInitTestObject4F04B0{initData: data}
			playerUnit := &goldInitTestObject4F04B0{experience: 1000}
			second := &goldInitTestPlayer4F04B0{name: "p2"}
			first := &goldInitTestPlayer4F04B0{name: "p1", unit: playerUnit, next: second}
			events := make([]string, 0, len(allEvents))
			record := func(event string) {
				events = append(events, event)
				if event == fault {
					panic(fault)
				}
			}
			truncCall := 0
			randomCall := 0
			hooks := goldInitHooks4F04B0[
				*goldInitTestObject4F04B0,
				*goldInitTestData4F04B0,
				*goldInitTestPlayer4F04B0,
			]{
				loadUnitArg: func() (*goldInitTestObject4F04B0, int32) {
					record("load-unit")
					return gold, -1
				},
				loadInitData: func(unit *goldInitTestObject4F04B0) *goldInitTestData4F04B0 {
					record("load-init")
					return unit.initData
				},
				loadAmount: func(data *goldInitTestData4F04B0) uint32 {
					record("load-amount")
					return data.amount
				},
				firstPlayer: func() *goldInitTestPlayer4F04B0 {
					record("first-player")
					return first
				},
				loadPlayerUnit: func(player *goldInitTestPlayer4F04B0) *goldInitTestObject4F04B0 {
					record("load-player-unit:" + player.name)
					return player.unit
				},
				loadExperience: func(unit *goldInitTestObject4F04B0) float32 {
					record("load-experience:u1")
					return unit.experience
				},
				nextPlayer: func(player *goldInitTestPlayer4F04B0) *goldInitTestPlayer4F04B0 {
					record("next-player:" + player.name)
					return player.next
				},
				truncQwordLow: func(float64) int32 {
					name := []string{"upper", "lower", "negative"}[truncCall]
					truncCall++
					record("trunc-" + name)
					return 0
				},
				randomInt: func(minimum, maximum int32, _ string, _ int32) int32 {
					name := "scaled"
					if randomCall != 0 {
						name = "base"
					}
					randomCall++
					record("random-" + name)
					return minimum
				},
				storeAmount: func(data *goldInitTestData4F04B0, amount uint32) {
					record("store-amount")
					data.amount = amount
				},
			}

			func() {
				defer func() {
					if got := recover(); got != fault {
						t.Fatalf("panic = %v, want %q", got, fault)
					}
				}()
				goldInit4F04B0(hooks)
			}()
			if want := allEvents[:faultIndex+1]; !reflect.DeepEqual(events, want) {
				t.Fatalf("events = %v, want %v", events, want)
			}
			if data.amount != 0 || data.guard != 0xa5a5a5a5 {
				t.Fatalf("record changed before completed store: %+v", *data)
			}
		})
	}
}
