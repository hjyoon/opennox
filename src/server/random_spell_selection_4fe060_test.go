package server

import (
	"crypto/sha256"
	"fmt"
	"math"
	"reflect"
	"testing"
)

var randomSpellExcludedIDs4FE100 = map[int32]bool{
	1: true, 2: true, 6: true, 13: true, 15: true, 18: true, 19: true,
	20: true, 30: true, 32: true, 33: true, 34: true, 38: true, 51: true,
	57: true, 68: true, 69: true, 70: true, 73: true, 129: true, 133: true,
}

func TestRandomSpellExcluded4FE100ExactSelector(t *testing.T) {
	var selector [133]byte
	excluded := 0
	for spellID := int32(1); spellID <= int32(len(selector)); spellID++ {
		got := randomSpellExcluded4FE100(spellID)
		want := int32(0)
		selector[spellID-1] = 1
		if randomSpellExcludedIDs4FE100[spellID] {
			want = 1
			selector[spellID-1] = 0
			excluded++
		}
		if got != want {
			t.Fatalf("spell %d exclusion = %d, want %d", spellID, got, want)
		}
	}
	if excluded != 21 {
		t.Fatalf("excluded spell count = %d, want 21", excluded)
	}
	if got, want := fmt.Sprintf("%x", sha256.Sum256(selector[:])), "2acf6a879c7fa625c9eb3fd21231ceb43b98cdf6f09cfbdeaa648cd1ba50da4f"; got != want {
		t.Fatalf("selector SHA-256 = %s, want %s", got, want)
	}
	for _, spellID := range []int32{math.MinInt32, -1, 0, 134, 135, math.MaxInt32} {
		if got := randomSpellExcluded4FE100(spellID); got != 0 {
			t.Fatalf("out-of-range spell %d exclusion = %d, want 0", spellID, got)
		}
	}
}

func TestRandomSpellSelection4FE060ExactOrderAndCandidateOrder(t *testing.T) {
	valid := []int32{7, 1, 8, 34, 9}
	flagValues := map[int32]uint32{
		7: 0x00000006,
		8: randomSpellClassAny4FE060 | 0x00000004,
		9: 0x00000002,
	}
	index := 0
	var events []string
	hooks := randomSpellSelectionHooks4FE060{
		firstValid: func() int32 {
			events = append(events, "first")
			return valid[0]
		},
		excluded: func(spellID int32) int32 {
			events = append(events, fmt.Sprintf("excluded:%d", spellID))
			return randomSpellExcluded4FE100(spellID)
		},
		flags: func(spellID int32) uint32 {
			events = append(events, fmt.Sprintf("flags:%d", spellID))
			return flagValues[spellID]
		},
		nextValid: func(spellID int32) int32 {
			events = append(events, fmt.Sprintf("next:%d", spellID))
			if spellID != valid[index] {
				t.Fatalf("next current = %d, want cached %d", spellID, valid[index])
			}
			index++
			if index == len(valid) {
				return 0
			}
			return valid[index]
		},
		randomInt: func(minimum, maximum int32) int32 {
			events = append(events, fmt.Sprintf("random:%d:%d", minimum, maximum))
			if minimum != 0 || maximum != 1 {
				t.Fatalf("RNG bounds = %d..%d, want 0..1", minimum, maximum)
			}
			return 1
		},
	}
	if got := randomSpellSelection4FE060(0x2, 0x4, hooks); got != 8 {
		t.Fatalf("selected spell = %d, want second retained spell 8", got)
	}
	wantEvents := []string{
		"first",
		"excluded:7", "flags:7", "next:7",
		"excluded:1", "next:1",
		"excluded:8", "flags:8", "next:8",
		"excluded:34", "next:34",
		"excluded:9", "flags:9", "next:9",
		"random:0:1",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events:\n got  %#v\n want %#v", events, wantEvents)
	}
}

func TestRandomSpellSelection4FE060ExactMaskGates(t *testing.T) {
	tests := []struct {
		name       string
		firstMask  uint32
		secondMask uint32
		flags      uint32
		want       int32
	}{
		{"first and second", 0x2, 0x4, 0x6, 7},
		{"any and second", 0x2, 0x4, randomSpellClassAny4FE060 | 0x4, 7},
		{"full high dword masks", 0x80000000, 0x40000000, 0xc0000000, 7},
		{"second without first or any", 0x2, 0x4, 0x4, 0},
		{"first without second", 0x2, 0x4, 0x2, 0},
		{"any without second", 0x2, 0x4, randomSpellClassAny4FE060, 0},
		{"zero second mask", 0x2, 0, 0xffffffff, 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			flagsCalls := 0
			randomCalls := 0
			hooks := randomSpellSelectionHooks4FE060{
				firstValid: func() int32 { return 7 },
				excluded:   randomSpellExcluded4FE100,
				flags: func(spellID int32) uint32 {
					flagsCalls++
					if spellID != 7 {
						t.Fatalf("flags spell = %d, want 7", spellID)
					}
					return test.flags
				},
				nextValid: func(spellID int32) int32 {
					if spellID != 7 {
						t.Fatalf("next spell = %d, want 7", spellID)
					}
					return 0
				},
				randomInt: func(minimum, maximum int32) int32 {
					randomCalls++
					if minimum != 0 || maximum != 0 {
						t.Fatalf("RNG bounds = %d..%d, want 0..0", minimum, maximum)
					}
					return 0
				},
			}
			if got := randomSpellSelection4FE060(test.firstMask, test.secondMask, hooks); got != test.want {
				t.Fatalf("selected spell = %d, want %d", got, test.want)
			}
			wantRandomCalls := 0
			if test.want != 0 {
				wantRandomCalls = 1
			}
			if flagsCalls != 1 || randomCalls != wantRandomCalls {
				t.Fatalf("flags/RNG calls = %d/%d, want 1/%d", flagsCalls, randomCalls, wantRandomCalls)
			}
		})
	}
}

func TestRandomSpellSelection4FE060ZeroFirstAndEmptyResultSkipRNG(t *testing.T) {
	t.Run("zero first", func(t *testing.T) {
		hooks := randomSpellSelectionHooks4FE060{
			firstValid: func() int32 { return 0 },
			excluded:   func(int32) int32 { t.Fatal("zero first called exclusion helper"); return 0 },
			flags:      func(int32) uint32 { t.Fatal("zero first loaded flags"); return 0 },
			nextValid:  func(int32) int32 { t.Fatal("zero first called NextValid"); return 0 },
			randomInt:  func(int32, int32) int32 { t.Fatal("zero first called RNG"); return 0 },
		}
		if got := randomSpellSelection4FE060(0xffffffff, 0xffffffff, hooks); got != 0 {
			t.Fatalf("selected spell = %d, want 0", got)
		}
	})

	t.Run("no retained candidates", func(t *testing.T) {
		nextCalls := 0
		hooks := randomSpellSelectionHooks4FE060{
			firstValid: func() int32 { return 7 },
			excluded:   randomSpellExcluded4FE100,
			flags:      func(int32) uint32 { return 0 },
			nextValid: func(spellID int32) int32 {
				nextCalls++
				if spellID != 7 {
					t.Fatalf("next spell = %d, want 7", spellID)
				}
				return 0
			},
			randomInt: func(int32, int32) int32 { t.Fatal("empty candidates called RNG"); return 0 },
		}
		if got := randomSpellSelection4FE060(1, 1, hooks); got != 0 || nextCalls != 1 {
			t.Fatalf("selected spell/next calls = %d/%d, want 0/1", got, nextCalls)
		}
	})
}

func TestRandomSpellSelection4FE060TreatsAnyNonzeroHelperResultAsExcluded(t *testing.T) {
	for _, result := range []int32{1, -1, math.MinInt32, math.MaxInt32} {
		t.Run(fmt.Sprintf("result_%d", result), func(t *testing.T) {
			nextCalls := 0
			hooks := randomSpellSelectionHooks4FE060{
				firstValid: func() int32 { return 7 },
				excluded:   func(int32) int32 { return result },
				flags:      func(int32) uint32 { t.Fatal("excluded spell loaded flags"); return 0 },
				nextValid: func(int32) int32 {
					nextCalls++
					return 0
				},
				randomInt: func(int32, int32) int32 { t.Fatal("excluded spell called RNG"); return 0 },
			}
			if got := randomSpellSelection4FE060(1, 1, hooks); got != 0 || nextCalls != 1 {
				t.Fatalf("selected spell/next calls = %d/%d, want 0/1", got, nextCalls)
			}
		})
	}
}

func TestRandomSpellSelection4FE060PreservesSignedDwordIDsAndFixedCapacity(t *testing.T) {
	t.Run("signed IDs", func(t *testing.T) {
		valid := []int32{math.MinInt32, math.MaxInt32, -1}
		index := 0
		hooks := randomSpellSelectionHooks4FE060{
			firstValid: func() int32 { return valid[0] },
			excluded:   randomSpellExcluded4FE100,
			flags:      func(int32) uint32 { return 0x80000000 },
			nextValid: func(spellID int32) int32 {
				if spellID != valid[index] {
					t.Fatalf("next spell = %d, want %d", spellID, valid[index])
				}
				index++
				if index == len(valid) {
					return 0
				}
				return valid[index]
			},
			randomInt: func(minimum, maximum int32) int32 {
				if minimum != 0 || maximum != 2 {
					t.Fatalf("RNG bounds = %d..%d, want 0..2", minimum, maximum)
				}
				return 2
			},
		}
		if got := randomSpellSelection4FE060(0x80000000, 0x80000000, hooks); got != -1 {
			t.Fatalf("selected spell = %d, want -1", got)
		}
	})

	t.Run("all 137 candidate slots", func(t *testing.T) {
		visited := 1
		flagsCalls := 0
		nextCalls := 0
		hooks := randomSpellSelectionHooks4FE060{
			firstValid: func() int32 { return 7 },
			excluded:   randomSpellExcluded4FE100,
			flags: func(int32) uint32 {
				flagsCalls++
				return 1
			},
			nextValid: func(int32) int32 {
				nextCalls++
				if visited == randomSpellCandidateCapacity4FE060 {
					return 0
				}
				visited++
				return 7
			},
			randomInt: func(minimum, maximum int32) int32 {
				if minimum != 0 || maximum != randomSpellCandidateCapacity4FE060-1 {
					t.Fatalf("RNG bounds = %d..%d, want 0..136", minimum, maximum)
				}
				return maximum
			},
		}
		if got := randomSpellSelection4FE060(1, 1, hooks); got != 7 {
			t.Fatalf("selected spell = %d, want 7", got)
		}
		if flagsCalls != randomSpellCandidateCapacity4FE060 || nextCalls != randomSpellCandidateCapacity4FE060 {
			t.Fatalf("flags/next calls = %d/%d, want 137/137", flagsCalls, nextCalls)
		}
	})
}

type randomSpellFaultState4FE060 struct {
	events  []string
	faultAt int
	index   int
}

func (s *randomSpellFaultState4FE060) event(name string) {
	s.events = append(s.events, name)
	if len(s.events) == s.faultAt {
		panic(name)
	}
}

func (s *randomSpellFaultState4FE060) hooks() randomSpellSelectionHooks4FE060 {
	valid := [...]int32{7, 1, 8}
	return randomSpellSelectionHooks4FE060{
		firstValid: func() int32 {
			s.event("first")
			return valid[0]
		},
		excluded: func(spellID int32) int32 {
			s.event(fmt.Sprintf("excluded:%d", spellID))
			return randomSpellExcluded4FE100(spellID)
		},
		flags: func(spellID int32) uint32 {
			s.event(fmt.Sprintf("flags:%d", spellID))
			if spellID == 8 {
				return randomSpellClassAny4FE060 | 4
			}
			return 6
		},
		nextValid: func(spellID int32) int32 {
			s.event(fmt.Sprintf("next:%d", spellID))
			s.index++
			if s.index == len(valid) {
				return 0
			}
			return valid[s.index]
		},
		randomInt: func(minimum, maximum int32) int32 {
			s.event(fmt.Sprintf("random:%d:%d", minimum, maximum))
			return 1
		},
	}
}

func TestRandomSpellSelection4FE060ExactFaultPrefixes(t *testing.T) {
	want := []string{
		"first",
		"excluded:7", "flags:7", "next:7",
		"excluded:1", "next:1",
		"excluded:8", "flags:8", "next:8",
		"random:0:1",
	}
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("%02d_%s", faultAt, want[faultAt-1]), func(t *testing.T) {
			state := &randomSpellFaultState4FE060{faultAt: faultAt}
			panicked := false
			func() {
				defer func() { panicked = recover() != nil }()
				randomSpellSelection4FE060(2, 4, state.hooks())
			}()
			if !panicked {
				t.Fatal("expected injected fault")
			}
			if !reflect.DeepEqual(state.events, want[:faultAt]) {
				t.Fatalf("events:\n got  %#v\n want %#v", state.events, want[:faultAt])
			}
		})
	}
}
