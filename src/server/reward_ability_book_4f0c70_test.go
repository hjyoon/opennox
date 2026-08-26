package server

import (
	"math"
	"math/bits"
	"slices"
	"testing"
)

type rewardAbilityBookTestData4F0C70 struct {
	flags     uint8
	abilities [rewardAbilityBookCount4F0C70]uint8
}

type rewardAbilityBookTestMarker4F0C70 struct {
	data *rewardAbilityBookTestData4F0C70
}

type rewardAbilityBookTestObject4F0C70 struct {
	typeName string
	ability  uint8
}

func rewardAbilityBookTestHooks4F0C70() rewardAbilityBookHooks4F0C70[
	*rewardAbilityBookTestMarker4F0C70,
	*rewardAbilityBookTestData4F0C70,
	*rewardAbilityBookTestObject4F0C70,
] {
	return rewardAbilityBookHooks4F0C70[
		*rewardAbilityBookTestMarker4F0C70,
		*rewardAbilityBookTestData4F0C70,
		*rewardAbilityBookTestObject4F0C70,
	]{
		loadInitData: func(marker *rewardAbilityBookTestMarker4F0C70) *rewardAbilityBookTestData4F0C70 {
			return marker.data
		},
		loadFlags: func(data *rewardAbilityBookTestData4F0C70) uint8 {
			return data.flags
		},
		loadExplicitAbility: func(data *rewardAbilityBookTestData4F0C70, index int) uint8 {
			return data.abilities[index]
		},
		randomInt: func(int32, int32) int32 { return 1 },
		createObjectByType: func(typeName string) *rewardAbilityBookTestObject4F0C70 {
			return &rewardAbilityBookTestObject4F0C70{typeName: typeName}
		},
		isNilObject: func(object *rewardAbilityBookTestObject4F0C70) bool {
			return object == nil
		},
		storeAbility: func(object *rewardAbilityBookTestObject4F0C70, ability uint8) {
			object.ability = ability
		},
	}
}

func TestRewardAbilityBookExplicitExhaustive4F0C70(t *testing.T) {
	for mask := 0; mask < 1<<rewardAbilityBookCount4F0C70; mask++ {
		count := bits.OnesCount(uint(mask))
		draws := count
		if draws == 0 {
			draws = 1
		}
		for draw := 0; draw < draws; draw++ {
			data := &rewardAbilityBookTestData4F0C70{flags: rewardAbilityBookExplicitFlag4F0C70}
			for index := range data.abilities {
				if mask&(1<<index) != 0 {
					data.abilities[index] = 1
				}
			}
			hooks := rewardAbilityBookTestHooks4F0C70()
			rngCalls := 0
			hooks.randomInt = func(minimum, maximum int32) int32 {
				rngCalls++
				if minimum != 0 || maximum != int32(count-1) {
					t.Fatalf("mask %#x RNG bounds = %d..%d, want 0..%d", mask, minimum, maximum, count-1)
				}
				return int32(draw)
			}
			got := rewardAbilityBook4F0C70(&rewardAbilityBookTestMarker4F0C70{data: data}, hooks)
			if count == 0 {
				if got != nil || rngCalls != 0 {
					t.Fatalf("empty mask result/RNG = %#v/%d, want nil/0", got, rngCalls)
				}
				continue
			}
			wantID := -1
			ordinal := 0
			for index := 0; index < rewardAbilityBookCount4F0C70; index++ {
				if mask&(1<<index) == 0 {
					continue
				}
				if ordinal == draw {
					wantID = index
					break
				}
				ordinal++
			}
			if wantID == 0 {
				if got != nil || rngCalls != 1 {
					t.Fatalf("mask %#x draw %d index-zero result/RNG = %#v/%d", mask, draw, got, rngCalls)
				}
			} else if got == nil || got.typeName != rewardAbilityBookType4F0C70 || got.ability != uint8(wantID) || rngCalls != 1 {
				t.Fatalf("mask %#x draw %d result/RNG = %#v/%d, want AbilityBook/%d/1", mask, draw, got, rngCalls, wantID)
			}
		}
	}
}

func TestRewardAbilityBookExplicitUsesCachedDataAndLiveSecondPass4F0C70(t *testing.T) {
	entry := &rewardAbilityBookTestData4F0C70{flags: rewardAbilityBookExplicitFlag4F0C70}
	entry.abilities = [rewardAbilityBookCount4F0C70]uint8{2, 1, 0, 1, 0, 1}
	replacement := &rewardAbilityBookTestData4F0C70{}
	marker := &rewardAbilityBookTestMarker4F0C70{data: entry}
	hooks := rewardAbilityBookTestHooks4F0C70()
	reads := 0
	hooks.loadFlags = func(data *rewardAbilityBookTestData4F0C70) uint8 {
		if data != entry {
			t.Fatal("flags did not use entry-cached InitData")
		}
		marker.data = replacement
		return data.flags
	}
	hooks.loadExplicitAbility = func(data *rewardAbilityBookTestData4F0C70, index int) uint8 {
		if data != entry {
			t.Fatal("ability read did not use entry-cached InitData")
		}
		reads++
		return data.abilities[index]
	}
	hooks.randomInt = func(minimum, maximum int32) int32 {
		if minimum != 0 || maximum != 2 || reads != 6 {
			t.Fatalf("RNG bounds/read prefix = %d..%d/%d, want 0..2/6", minimum, maximum, reads)
		}
		entry.abilities[3] = 0
		entry.abilities[4] = 1
		return 1
	}
	got := rewardAbilityBook4F0C70(marker, hooks)
	if got == nil || got.typeName != rewardAbilityBookType4F0C70 || got.ability != 4 {
		t.Fatalf("result = %#v, want AbilityBook ability 4", got)
	}
	if reads != 11 || marker.data != replacement {
		t.Fatalf("reads/live marker = %d/%p, want 11/replacement", reads, marker.data)
	}
}

func TestRewardAbilityBookExplicitExactOneAndExhaustion4F0C70(t *testing.T) {
	t.Run("non-one bytes are disabled", func(t *testing.T) {
		data := &rewardAbilityBookTestData4F0C70{
			flags:     rewardAbilityBookExplicitFlag4F0C70,
			abilities: [rewardAbilityBookCount4F0C70]uint8{2, 3, 0xff, 0, 2, 0xff},
		}
		hooks := rewardAbilityBookTestHooks4F0C70()
		hooks.randomInt = func(int32, int32) int32 {
			t.Fatal("disabled bytes reached RNG")
			return 0
		}
		if got := rewardAbilityBook4F0C70(&rewardAbilityBookTestMarker4F0C70{data: data}, hooks); got != nil {
			t.Fatalf("disabled result = %#v, want nil", got)
		}
	})

	for _, draw := range []int32{-1, 2} {
		t.Run("out of contract draw", func(t *testing.T) {
			data := &rewardAbilityBookTestData4F0C70{flags: rewardAbilityBookExplicitFlag4F0C70}
			data.abilities[2], data.abilities[5] = 1, 1
			hooks := rewardAbilityBookTestHooks4F0C70()
			hooks.randomInt = func(int32, int32) int32 { return draw }
			if got := rewardAbilityBook4F0C70(&rewardAbilityBookTestMarker4F0C70{data: data}, hooks); got != nil {
				t.Fatalf("draw %d result = %#v, want nil", draw, got)
			}
		})
	}

	t.Run("second pass can exhaust", func(t *testing.T) {
		data := &rewardAbilityBookTestData4F0C70{flags: rewardAbilityBookExplicitFlag4F0C70}
		data.abilities[2], data.abilities[5] = 1, 1
		hooks := rewardAbilityBookTestHooks4F0C70()
		reads := 0
		hooks.loadExplicitAbility = func(data *rewardAbilityBookTestData4F0C70, index int) uint8 {
			reads++
			return data.abilities[index]
		}
		hooks.randomInt = func(int32, int32) int32 {
			data.abilities[5] = 0
			return 1
		}
		if got := rewardAbilityBook4F0C70(&rewardAbilityBookTestMarker4F0C70{data: data}, hooks); got != nil || reads != 12 {
			t.Fatalf("exhausted result/reads = %#v/%d, want nil/12", got, reads)
		}
	})
}

func TestRewardAbilityBookAutomaticPreservesSignedResultLowByte4F0C70(t *testing.T) {
	tests := []int32{0, 1, 5, 6, -1, math.MinInt32, math.MaxInt32}
	for _, draw := range tests {
		hooks := rewardAbilityBookTestHooks4F0C70()
		var events []string
		hooks.loadExplicitAbility = func(*rewardAbilityBookTestData4F0C70, int) uint8 {
			t.Fatal("automatic path read explicit ability")
			return 0
		}
		hooks.randomInt = func(minimum, maximum int32) int32 {
			events = append(events, "rng")
			if minimum != 1 || maximum != 5 {
				t.Fatalf("automatic bounds = %d..%d, want 1..5", minimum, maximum)
			}
			return draw
		}
		hooks.createObjectByType = func(typeName string) *rewardAbilityBookTestObject4F0C70 {
			events = append(events, "create:"+typeName)
			return &rewardAbilityBookTestObject4F0C70{typeName: typeName}
		}
		hooks.storeAbility = func(object *rewardAbilityBookTestObject4F0C70, ability uint8) {
			events = append(events, "store")
			object.ability = ability
		}
		got := rewardAbilityBook4F0C70(
			&rewardAbilityBookTestMarker4F0C70{data: &rewardAbilityBookTestData4F0C70{}},
			hooks,
		)
		if draw == 0 {
			if got != nil || !slices.Equal(events, []string{"rng"}) {
				t.Fatalf("draw zero result/events = %#v/%v, want nil/[rng]", got, events)
			}
		} else if got == nil || got.ability != uint8(draw) || !slices.Equal(events, []string{"rng", "create:AbilityBook", "store"}) {
			t.Fatalf("draw %d result/events = %#v/%v", draw, got, events)
		}
	}
}

func TestRewardAbilityBookNilCreatedObjectSkipsStore4F0C70(t *testing.T) {
	hooks := rewardAbilityBookTestHooks4F0C70()
	hooks.createObjectByType = func(string) *rewardAbilityBookTestObject4F0C70 { return nil }
	hooks.storeAbility = func(*rewardAbilityBookTestObject4F0C70, uint8) {
		t.Fatal("nil object reached store")
	}
	got := rewardAbilityBook4F0C70(
		&rewardAbilityBookTestMarker4F0C70{data: &rewardAbilityBookTestData4F0C70{}},
		hooks,
	)
	if got != nil {
		t.Fatalf("nil creation result = %#v", got)
	}
}

func TestRewardAbilityBookFaultPrefixes4F0C70(t *testing.T) {
	fault := new(int)
	tests := []struct {
		name string
		run  func(*[]string)
		want []string
	}{
		{
			name: "nil marker",
			run: func(events *[]string) {
				hooks := rewardAbilityBookTestHooks4F0C70()
				hooks.loadInitData = func(marker *rewardAbilityBookTestMarker4F0C70) *rewardAbilityBookTestData4F0C70 {
					*events = append(*events, "init")
					return marker.data
				}
				rewardAbilityBook4F0C70(nil, hooks)
			},
			want: []string{"init"},
		},
		{
			name: "nil InitData",
			run: func(events *[]string) {
				hooks := rewardAbilityBookTestHooks4F0C70()
				hooks.loadInitData = func(*rewardAbilityBookTestMarker4F0C70) *rewardAbilityBookTestData4F0C70 {
					*events = append(*events, "init")
					return nil
				}
				hooks.loadFlags = func(data *rewardAbilityBookTestData4F0C70) uint8 {
					*events = append(*events, "flags")
					return data.flags
				}
				rewardAbilityBook4F0C70(&rewardAbilityBookTestMarker4F0C70{}, hooks)
			},
			want: []string{"init", "flags"},
		},
		{
			name: "automatic RNG",
			run: func(events *[]string) {
				hooks := rewardAbilityBookTestHooks4F0C70()
				hooks.loadInitData = func(marker *rewardAbilityBookTestMarker4F0C70) *rewardAbilityBookTestData4F0C70 {
					*events = append(*events, "init")
					return marker.data
				}
				hooks.loadFlags = func(data *rewardAbilityBookTestData4F0C70) uint8 {
					*events = append(*events, "flags")
					return data.flags
				}
				hooks.randomInt = func(int32, int32) int32 {
					*events = append(*events, "rng")
					panic(fault)
				}
				rewardAbilityBook4F0C70(&rewardAbilityBookTestMarker4F0C70{data: &rewardAbilityBookTestData4F0C70{}}, hooks)
			},
			want: []string{"init", "flags", "rng"},
		},
		{
			name: "explicit first load",
			run: func(events *[]string) {
				data := &rewardAbilityBookTestData4F0C70{flags: rewardAbilityBookExplicitFlag4F0C70}
				hooks := rewardAbilityBookTestHooks4F0C70()
				hooks.loadExplicitAbility = func(*rewardAbilityBookTestData4F0C70, int) uint8 {
					*events = append(*events, "ability")
					panic(fault)
				}
				rewardAbilityBook4F0C70(&rewardAbilityBookTestMarker4F0C70{data: data}, hooks)
			},
			want: []string{"ability"},
		},
		{
			name: "explicit RNG after first pass",
			run: func(events *[]string) {
				data := &rewardAbilityBookTestData4F0C70{flags: rewardAbilityBookExplicitFlag4F0C70}
				data.abilities[5] = 1
				hooks := rewardAbilityBookTestHooks4F0C70()
				hooks.loadExplicitAbility = func(data *rewardAbilityBookTestData4F0C70, index int) uint8 {
					*events = append(*events, "ability")
					return data.abilities[index]
				}
				hooks.randomInt = func(int32, int32) int32 {
					*events = append(*events, "rng")
					panic(fault)
				}
				rewardAbilityBook4F0C70(&rewardAbilityBookTestMarker4F0C70{data: data}, hooks)
			},
			want: []string{"ability", "ability", "ability", "ability", "ability", "ability", "rng"},
		},
		{
			name: "explicit second pass",
			run: func(events *[]string) {
				data := &rewardAbilityBookTestData4F0C70{flags: rewardAbilityBookExplicitFlag4F0C70}
				data.abilities[5] = 1
				hooks := rewardAbilityBookTestHooks4F0C70()
				reads := 0
				hooks.loadExplicitAbility = func(data *rewardAbilityBookTestData4F0C70, index int) uint8 {
					reads++
					*events = append(*events, "ability")
					if reads == 7 {
						panic(fault)
					}
					return data.abilities[index]
				}
				hooks.randomInt = func(int32, int32) int32 {
					*events = append(*events, "rng")
					return 0
				}
				rewardAbilityBook4F0C70(&rewardAbilityBookTestMarker4F0C70{data: data}, hooks)
			},
			want: []string{"ability", "ability", "ability", "ability", "ability", "ability", "rng", "ability"},
		},
		{
			name: "create",
			run: func(events *[]string) {
				hooks := rewardAbilityBookTestHooks4F0C70()
				hooks.createObjectByType = func(string) *rewardAbilityBookTestObject4F0C70 {
					*events = append(*events, "create")
					panic(fault)
				}
				rewardAbilityBook4F0C70(&rewardAbilityBookTestMarker4F0C70{data: &rewardAbilityBookTestData4F0C70{}}, hooks)
			},
			want: []string{"create"},
		},
		{
			name: "nil-object predicate",
			run: func(events *[]string) {
				hooks := rewardAbilityBookTestHooks4F0C70()
				hooks.isNilObject = func(*rewardAbilityBookTestObject4F0C70) bool {
					*events = append(*events, "is-nil")
					panic(fault)
				}
				rewardAbilityBook4F0C70(&rewardAbilityBookTestMarker4F0C70{data: &rewardAbilityBookTestData4F0C70{}}, hooks)
			},
			want: []string{"is-nil"},
		},
		{
			name: "store",
			run: func(events *[]string) {
				hooks := rewardAbilityBookTestHooks4F0C70()
				hooks.storeAbility = func(*rewardAbilityBookTestObject4F0C70, uint8) {
					*events = append(*events, "store")
					panic(fault)
				}
				rewardAbilityBook4F0C70(&rewardAbilityBookTestMarker4F0C70{data: &rewardAbilityBookTestData4F0C70{}}, hooks)
			},
			want: []string{"store"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var events []string
			defer func() {
				if got := recover(); got == nil || (got != fault && test.name != "nil marker" && test.name != "nil InitData") {
					t.Fatalf("recover = %v, want fault", got)
				}
				if !slices.Equal(events, test.want) {
					t.Fatalf("events = %v, want %v", events, test.want)
				}
			}()
			test.run(&events)
		})
	}
}
