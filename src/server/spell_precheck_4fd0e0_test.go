package server

import (
	"fmt"
	"reflect"
	"testing"
)

func TestSpellPrecheck4FD0E0PlayerExactTraceAndLiveLoads(t *testing.T) {
	const (
		originalSpell = int32(0x12345678)
		mutatedSpell  = int32(-77)
		originalUnit  = uint64(0x1_0000_1000)
		mutatedUnit   = uint64(0x2_0000_2000)
		cachedOwner   = uint64(0x3_0000_3000)
		updateData    = uint64(0x4_0000_4000)
		playerData    = uint64(0x5_0000_5000)
		class         = uint8(2)
		classResult   = int32(-0x1234567)
	)
	spellArg := originalSpell
	unitArg := originalUnit
	unitClass := uint8(0)
	var events []string

	got := spellPrecheck4FD0E0(spellPrecheckHooks4FD0E0[uint64, uint64, uint64]{
		loadSpellArg: func() int32 {
			events = append(events, "spell-arg")
			return spellArg
		},
		spellFlags: func(gotSpell int32) uint32 {
			events = append(events, fmt.Sprintf("flags:%d", gotSpell))
			if gotSpell != originalSpell {
				t.Fatalf("flags spell = %d, want cached %d", gotSpell, originalSpell)
			}
			spellArg = mutatedSpell
			unitArg = mutatedUnit
			return ^uint32(0)
		},
		loadUnitArg: func() uint64 {
			events = append(events, "unit-arg")
			return unitArg
		},
		findParentPlayer: func(gotUnit uint64) uint64 {
			events = append(events, fmt.Sprintf("parent:%#x", gotUnit))
			if gotUnit != mutatedUnit {
				t.Fatalf("parent unit = %#x, want post-flags unit %#x", gotUnit, mutatedUnit)
			}
			return cachedOwner
		},
		spellEnabled: func(gotSpell int32) int32 {
			events = append(events, fmt.Sprintf("enabled:%d", gotSpell))
			if gotSpell != originalSpell {
				t.Fatalf("enabled spell = %d, want cached %d", gotSpell, originalSpell)
			}
			unitClass = 0x84
			return -1
		},
		loadUnitClassLow: func(gotUnit uint64) uint8 {
			events = append(events, fmt.Sprintf("class-low:%#x", gotUnit))
			return unitClass
		},
		loadUpdateData: func(gotUnit uint64) uint64 {
			events = append(events, fmt.Sprintf("update:%#x", gotUnit))
			return updateData
		},
		loadPlayer: func(gotUpdate uint64) uint64 {
			events = append(events, fmt.Sprintf("player:%#x", gotUpdate))
			return playerData
		},
		loadPlayerClass: func(gotPlayer uint64) uint8 {
			events = append(events, fmt.Sprintf("player-class:%#x", gotPlayer))
			return class
		},
		checkPlayerSpellClass: func(gotClass uint8, gotSpell int32) int32 {
			events = append(events, fmt.Sprintf("check:%d:%d", gotClass, gotSpell))
			if gotClass != class || gotSpell != originalSpell {
				t.Fatalf("class check = (%d, %d), want (%d, %d)", gotClass, gotSpell, class, originalSpell)
			}
			return classResult
		},
		summonAllowed: func(int32, uint64) int32 {
			t.Fatal("Player path reached summon check")
			return 0
		},
	})

	wantEvents := []string{
		"spell-arg",
		fmt.Sprintf("flags:%d", originalSpell),
		"unit-arg",
		fmt.Sprintf("parent:%#x", mutatedUnit),
		fmt.Sprintf("enabled:%d", originalSpell),
		fmt.Sprintf("class-low:%#x", mutatedUnit),
		fmt.Sprintf("update:%#x", mutatedUnit),
		fmt.Sprintf("player:%#x", updateData),
		fmt.Sprintf("player-class:%#x", playerData),
		fmt.Sprintf("check:%d:%d", class, originalSpell),
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	if got != classResult {
		t.Fatalf("result = %d, want verbatim class result %d", got, classResult)
	}
	for _, token := range []uint64{mutatedUnit, cachedOwner, updateData, playerData} {
		if token <= uint64(^uint32(0)) {
			t.Fatalf("token %#x did not retain native-width high bits", token)
		}
	}
}

func TestSpellPrecheck4FD0E0DisabledStopsBeforeClassLoad(t *testing.T) {
	var events []string
	got := spellPrecheck4FD0E0(spellPrecheckHooks4FD0E0[int, int, int]{
		loadSpellArg: func() int32 {
			events = append(events, "spell")
			return 41
		},
		spellFlags: func(int32) uint32 {
			events = append(events, "flags")
			return 0
		},
		loadUnitArg: func() int {
			events = append(events, "unit")
			return 0
		},
		findParentPlayer: func(int) int {
			events = append(events, "parent")
			return 0
		},
		spellEnabled: func(int32) int32 {
			events = append(events, "enabled")
			return 0
		},
		loadUnitClassLow: func(int) uint8 { t.Fatal("disabled spell loaded class"); return 0 },
		loadUpdateData:   func(int) int { t.Fatal("disabled spell loaded update data"); return 0 },
		loadPlayer:       func(int) int { t.Fatal("disabled spell loaded Player"); return 0 },
		loadPlayerClass:  func(int) uint8 { t.Fatal("disabled spell loaded Player class"); return 0 },
		checkPlayerSpellClass: func(uint8, int32) int32 {
			t.Fatal("disabled spell checked Player class")
			return 0
		},
		summonAllowed: func(int32, int) int32 {
			t.Fatal("disabled spell checked summon capacity")
			return 0
		},
	})
	if got != spellPrecheckIllegal4FD0E0 {
		t.Fatalf("result = %d, want %d", got, spellPrecheckIllegal4FD0E0)
	}
	if want := []string{"spell", "flags", "unit", "parent", "enabled"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestSpellPrecheck4FD0E0NonPlayerCanonicalResultAndCachedOwner(t *testing.T) {
	const (
		unit        = uint64(0x1_1111_1111)
		cachedOwner = uint64(0x2_2222_2222)
		newOwner    = uint64(0x3_3333_3333)
	)
	tests := []struct {
		name   string
		helper int32
		want   int32
	}{
		{name: "zero", helper: 0, want: spellPrecheckIllegal4FD0E0},
		{name: "one", helper: 1, want: 0},
		{name: "minus_one", helper: -1, want: 0},
		{name: "minimum", helper: -2147483648, want: 0},
		{name: "maximum", helper: 2147483647, want: 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ownerSource := cachedOwner
			var events []string
			got := spellPrecheck4FD0E0(spellPrecheckHooks4FD0E0[uint64, uint64, uint64]{
				loadSpellArg: func() int32 { events = append(events, "spell"); return 29 },
				spellFlags:   func(int32) uint32 { events = append(events, "flags"); return 0 },
				loadUnitArg:  func() uint64 { events = append(events, "unit"); return unit },
				findParentPlayer: func(uint64) uint64 {
					events = append(events, "parent")
					return ownerSource
				},
				spellEnabled: func(int32) int32 {
					events = append(events, "enabled")
					ownerSource = newOwner
					return 7
				},
				loadUnitClassLow: func(uint64) uint8 {
					events = append(events, "class")
					return 0xf8
				},
				loadUpdateData:  func(uint64) uint64 { t.Fatal("non-Player loaded update data"); return 0 },
				loadPlayer:      func(uint64) uint64 { t.Fatal("non-Player loaded Player"); return 0 },
				loadPlayerClass: func(uint64) uint8 { t.Fatal("non-Player loaded Player class"); return 0 },
				checkPlayerSpellClass: func(uint8, int32) int32 {
					t.Fatal("non-Player checked Player class")
					return 0
				},
				summonAllowed: func(gotSpell int32, gotOwner uint64) int32 {
					events = append(events, "summon")
					if gotSpell != 29 || gotOwner != cachedOwner {
						t.Fatalf("summon args = (%d, %#x), want (29, %#x)", gotSpell, gotOwner, cachedOwner)
					}
					return test.helper
				},
			})
			if got != test.want {
				t.Fatalf("result = %d, want %d", got, test.want)
			}
			if want := []string{"spell", "flags", "unit", "parent", "enabled", "class", "summon"}; !reflect.DeepEqual(events, want) {
				t.Fatalf("events = %v, want %v", events, want)
			}
		})
	}
}

func TestSpellPrecheck4FD0E0PlayerFaultPrefixes(t *testing.T) {
	want := []string{
		"spell-arg", "flags", "unit-arg", "parent", "enabled",
		"class-low", "update-data", "player", "player-class", "class-check",
	}
	for failAt := 1; failAt <= len(want); failAt++ {
		t.Run(fmt.Sprintf("step_%02d", failAt), func(t *testing.T) {
			var events []string
			emit := func(event string) {
				events = append(events, event)
				if len(events) == failAt {
					panic(event)
				}
			}
			hooks := spellPrecheckHooks4FD0E0[uint64, uint64, uint64]{
				loadSpellArg:          func() int32 { emit("spell-arg"); return 1 },
				spellFlags:            func(int32) uint32 { emit("flags"); return 0 },
				loadUnitArg:           func() uint64 { emit("unit-arg"); return 0x1_0000_0001 },
				findParentPlayer:      func(uint64) uint64 { emit("parent"); return 0x2_0000_0002 },
				spellEnabled:          func(int32) int32 { emit("enabled"); return 1 },
				loadUnitClassLow:      func(uint64) uint8 { emit("class-low"); return 4 },
				loadUpdateData:        func(uint64) uint64 { emit("update-data"); return 0x3_0000_0003 },
				loadPlayer:            func(uint64) uint64 { emit("player"); return 0x4_0000_0004 },
				loadPlayerClass:       func(uint64) uint8 { emit("player-class"); return 1 },
				checkPlayerSpellClass: func(uint8, int32) int32 { emit("class-check"); return 0 },
				summonAllowed: func(int32, uint64) int32 {
					t.Fatal("Player fault path reached summon check")
					return 0
				},
			}
			func() {
				defer func() {
					if recover() == nil {
						t.Fatalf("step %d did not panic", failAt)
					}
				}()
				_ = spellPrecheck4FD0E0(hooks)
			}()
			if expected := want[:failAt]; !reflect.DeepEqual(events, expected) {
				t.Fatalf("events = %v, want prefix %v", events, expected)
			}
		})
	}
}

func TestSpellPrecheck4FD0E0NonPlayerSummonFaultPrefix(t *testing.T) {
	var events []string
	emit := func(event string) { events = append(events, event) }
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("summon callback did not panic")
			}
		}()
		_ = spellPrecheck4FD0E0(spellPrecheckHooks4FD0E0[int, int, int]{
			loadSpellArg:     func() int32 { emit("spell-arg"); return 1 },
			spellFlags:       func(int32) uint32 { emit("flags"); return 0 },
			loadUnitArg:      func() int { emit("unit-arg"); return 1 },
			findParentPlayer: func(int) int { emit("parent"); return 2 },
			spellEnabled:     func(int32) int32 { emit("enabled"); return 1 },
			loadUnitClassLow: func(int) uint8 { emit("class-low"); return 0 },
			loadUpdateData:   func(int) int { t.Fatal("non-Player fault path loaded update data"); return 0 },
			loadPlayer:       func(int) int { t.Fatal("non-Player fault path loaded Player"); return 0 },
			loadPlayerClass:  func(int) uint8 { t.Fatal("non-Player fault path loaded class"); return 0 },
			checkPlayerSpellClass: func(uint8, int32) int32 {
				t.Fatal("non-Player fault path checked class")
				return 0
			},
			summonAllowed: func(int32, int) int32 {
				emit("summon")
				panic("summon")
			},
		})
	}()
	want := []string{"spell-arg", "flags", "unit-arg", "parent", "enabled", "class-low", "summon"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}
