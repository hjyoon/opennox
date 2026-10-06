package legacy

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

type questMapFileLegacyServer4D0F60 struct {
	Server
	native *server.Server
}

func TestQuestMapFile4D0F60RecordLayout(t *testing.T) {
	var entry questMapEntry4D0F60
	if unsafe.Sizeof(entry) != 32 || unsafe.Offsetof(entry.Group) != 0 ||
		unsafe.Offsetof(entry.Name) != 4 || unsafe.Offsetof(entry.Uses) != 24 ||
		unsafe.Offsetof(entry.LastUsed) != 28 {
		t.Fatalf("Quest map PE32 record has wrong native layout: size=%d offsets=%d/%d/%d/%d",
			unsafe.Sizeof(entry), unsafe.Offsetof(entry.Group), unsafe.Offsetof(entry.Name), unsafe.Offsetof(entry.Uses), unsafe.Offsetof(entry.LastUsed))
	}
}

func TestQuestMapFile4D0F60CEntryFullPointerAndNil(t *testing.T) {
	old := questMapFileCall4D0F60
	t.Cleanup(func() { questMapFileCall4D0F60 = old })
	for _, tc := range []struct {
		name string
		want unsafe.Pointer
	}{
		{name: "native-name", want: memmap.PtrOff(0x5D4594, 1525136+32*127)},
		{name: "nil"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			questMapFileCall4D0F60 = func() unsafe.Pointer { calls++; return tc.want }
			if got := questMapFileCEntry4D0F60(); calls != 1 || got != tc.want {
				t.Fatalf("C entry returned %p after %d calls, want %p after one call", got, calls, tc.want)
			}
		})
	}
}

func TestQuestMapFile4D0F60EarlyReturnsAndSignedCount(t *testing.T) {
	for _, tc := range []struct {
		name   string
		count  uint32
		max    int32
		choose bool
		index  int32
		isNil  bool
	}{
		{name: "zero", isNil: true},
		{name: "one", count: 1},
		{name: "negative-one", count: 0xffffffff, max: -2, choose: true, index: 17},
		{name: "minimum-signed-count-wraps-upper-bound", count: 0x80000000, max: 0x7fffffff, choose: true, index: 17},
		{name: "negative-two", count: 0xfffffffe, max: -3, choose: true, index: 17},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var result byte
			var names []int32
			randomCalls, countCalls := 0, 0
			got := questMapFile4D0F60(questMapFileHooks4D0F60{
				Count: func() uint32 { countCalls++; return tc.count },
				Name: func(index int32) unsafe.Pointer {
					names = append(names, index)
					return unsafe.Pointer(&result)
				},
				Random: func(min, max int32) int32 {
					randomCalls++
					if !tc.choose || min != 0 || max != tc.max {
						t.Fatalf("random range [%d,%d], want [0,%d], allowed=%t", min, max, tc.max, tc.choose)
					}
					return tc.index
				},
				// History accessors deliberately nil: these branches cannot read
				// entries, last-map group or cooldown state, even for bad counts.
			})
			wantRandom := 0
			if tc.choose {
				wantRandom = 1
			}
			if countCalls != 1 || randomCalls != wantRandom {
				t.Fatalf("count/random reads=%d/%d, want 1/%d", countCalls, randomCalls, wantRandom)
			}
			if tc.isNil {
				if got != nil || len(names) != 0 {
					t.Fatalf("empty table returned %p / names %v", got, names)
				}
			} else if got != unsafe.Pointer(&result) || !reflect.DeepEqual(names, []int32{tc.index}) {
				t.Fatalf("return/name=%p/%v, want %p/[%d]", got, names, &result, tc.index)
			}
		})
	}
}

func TestQuestMapFile4D0F60ReloadsLiveHistoryAfterRandom(t *testing.T) {
	for _, tc := range []struct {
		name      string
		choice    int32
		want      int32
		lastReads int
		mutate    func(*uint32, *uint32, *uint32, []questMapEntry4D0F60)
	}{
		{name: "stable-second-ordinal", choice: 1, want: 2, lastReads: 2},
		{name: "count-shrinks-fallback-is-ordinal", choice: 1, want: 1, lastReads: 2, mutate: func(count, _, _ *uint32, _ []questMapEntry4D0F60) { *count = 2 }},
		{name: "zero-count-fallback", choice: 1, want: 1, lastReads: 1, mutate: func(count, _, _ *uint32, _ []questMapEntry4D0F60) { *count = 0 }},
		{name: "negative-count-fallback", choice: 1, want: 1, lastReads: 1, mutate: func(count, _, _ *uint32, _ []questMapEntry4D0F60) { *count = 0x80000000 }},
		{name: "last-index-reloaded", want: 2, lastReads: 2, mutate: func(_, last, _ *uint32, _ []questMapEntry4D0F60) { *last = 1 }},
		{name: "uses-reloaded", want: 2, lastReads: 2, mutate: func(_, _, _ *uint32, entries []questMapEntry4D0F60) { entries[1].Uses = 3 }},
		{name: "threshold-must-not-be-recomputed", want: 2, lastReads: 2, mutate: func(_, _, _ *uint32, entries []questMapEntry4D0F60) { entries[0].Uses, entries[1].Uses = 10, 5 }},
		{name: "clock-reloaded", want: 2, lastReads: 2, mutate: func(_, _, clock *uint32, entries []questMapEntry4D0F60) { *clock, entries[1].LastUsed = 8, 5 }},
		{name: "family-reloaded", want: 2, lastReads: 2, mutate: func(_, _, _ *uint32, entries []questMapEntry4D0F60) { entries[1].Group = 0 }},
		{name: "negative-choice-unclamped", choice: -1, want: -1, lastReads: 2},
		{name: "large-choice-unclamped", choice: 5, want: 5, lastReads: 2},
		{name: "count-growth-in-second-scan", choice: 1, want: 3, lastReads: 2, mutate: func(count, _, _ *uint32, entries []questMapEntry4D0F60) { *count, entries[1].Uses = 4, 3 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			count, last, clock := uint32(3), uint32(0), uint32(1000)
			entries := []questMapEntry4D0F60{{Group: 0, Uses: 3}, {Group: 1}, {Group: 2}, {Group: 3}}
			var result byte
			var names []int32
			countReads, lastReads, randomCalls := 0, 0, 0
			got := questMapFile4D0F60(questMapFileHooks4D0F60{
				Count:     func() uint32 { countReads++; return count },
				LastIndex: func() uint32 { lastReads++; return last },
				Clock:     func() uint32 { return clock },
				Entry:     func(index int32) *questMapEntry4D0F60 { return &entries[index] },
				Name: func(index int32) unsafe.Pointer {
					names = append(names, index)
					return unsafe.Pointer(&result)
				},
				Random: func(min, max int32) int32 {
					randomCalls++
					if min != 0 || max != 1 {
						t.Fatalf("first-pass range [%d,%d], want [0,1]", min, max)
					}
					if tc.mutate != nil {
						tc.mutate(&count, &last, &clock, entries)
					}
					return tc.choice
				},
			})
			if got != unsafe.Pointer(&result) || !reflect.DeepEqual(names, []int32{tc.want}) ||
				countReads != 2 || lastReads != tc.lastReads || randomCalls != 1 {
				t.Fatalf("return/names/count reads/last reads/random calls=%p/%v/%d/%d/%d, want %p/[%d]/2/%d/1",
					got, names, countReads, lastReads, randomCalls, &result, tc.want, tc.lastReads)
			}
		})
	}
}

func TestQuestMapFile4D0F60IndependentBoundaryMatrix(t *testing.T) {
	words := []uint32{0, 1, 2, 3, 4, 5, 7, 1000, 0x7ffffffe, 0x7fffffff, 0x80000000, 0x80000001, 0xfffffffe, 0xffffffff}
	data := rand.New(rand.NewSource(0x4d0f60))
	for fixture := 0; fixture < 4096; fixture++ {
		count := data.Intn(32) + 2
		clock := words[data.Intn(len(words))]
		last := uint32(data.Intn(count))
		seed := data.Intn(4096)
		entries := make([]questMapEntry4D0F60, count)
		reference := make([]questMapFileTestEntry4D0F60, count)
		for index := range entries {
			group := uint32(data.Intn(8))
			uses, lastUsed := words[data.Intn(len(words))], words[data.Intn(len(words))]
			entries[index] = questMapEntry4D0F60{Group: group, Uses: uses, LastUsed: lastUsed}
			reference[index] = questMapFileTestEntry4D0F60{group: group, uses: uses, last: lastUsed}
		}
		before := append([]questMapEntry4D0F60(nil), entries...)
		actualRNG, wantRNG := prand.New(seed), prand.New(seed)
		want := questMapFileReference4D0F60(reference, last, clock, wantRNG)
		var result byte
		var names []int32
		got := questMapFile4D0F60(questMapFileHooks4D0F60{
			Count:     func() uint32 { return uint32(count) },
			LastIndex: func() uint32 { return last },
			Clock:     func() uint32 { return clock },
			Entry:     func(index int32) *questMapEntry4D0F60 { return &entries[index] },
			Name:      func(index int32) unsafe.Pointer { names = append(names, index); return unsafe.Pointer(&result) },
			Random:    func(min, max int32) int32 { return int32(actualRNG.IntClamp(int(min), int(max))) },
		})
		if got != unsafe.Pointer(&result) || !reflect.DeepEqual(names, []int32{int32(want)}) ||
			actualRNG.Index() != wantRNG.Index() || !reflect.DeepEqual(entries, before) {
			t.Fatalf("fixture %d seed %d count %d last %d clock %#x: map %v / RNG %d, want [%d] / RNG %d; history unchanged=%t",
				fixture, seed, count, last, clock, names, actualRNG.Index(), want, wantRNG.Index(), reflect.DeepEqual(entries, before))
		}
	}
}

func (s *questMapFileLegacyServer4D0F60) S() *server.Server { return s.native }

type questMapFileTestEntry4D0F60 struct {
	group, uses, last uint32
}

// This independent reference follows GAME.EXE's signed JLE/JGE branches at
// 004D0F9C, 004D1022/004D108C and 004D1045/004D10AE. Decode DWORDs through
// int64 arithmetic, rather than sharing the production comparison helpers.
func questMapFileReference4D0F60(entries []questMapFileTestEntry4D0F60, last, clock uint32, rng *prand.Rand) int {
	signed := func(word uint32) int64 {
		value := int64(word)
		if word&0x80000000 != 0 {
			value -= 1 << 32
		}
		return value
	}
	if len(entries) == 0 {
		return -2 // Distinguish the null return from the legacy index -1 fallback.
	}
	if len(entries) == 1 {
		return 0
	}
	var maximum uint32
	for _, entry := range entries {
		if signed(entry.uses) > signed(maximum) {
			maximum = entry.uses
		}
	}
	if maximum == 0 {
		return rng.IntClamp(0, len(entries)-1)
	}
	equal := true
	for _, entry := range entries[1:] {
		if entry.uses != entries[0].uses {
			equal = false
		}
	}
	threshold := maximum
	if equal {
		threshold = uint32((uint64(maximum) + 1) & 0xffffffff)
	}
	var eligible []int
	for index, entry := range entries {
		age := uint32((uint64(clock) + (1 << 32) - uint64(entry.last)) & 0xffffffff)
		if signed(entry.uses) < signed(threshold) && uint32(index) != last &&
			entry.group != entries[last].group && signed(age) > 4 {
			eligible = append(eligible, index)
		}
	}
	choice := rng.IntClamp(0, len(eligible)-1)
	if choice >= 0 && choice < len(eligible) {
		return eligible[choice]
	}
	return choice
}

func TestQuestMapFile4D0F60NativeSignedHistoryAndPointer(t *testing.T) {
	capacity := make([]questMapFileTestEntry4D0F60, 128)
	for index := range capacity {
		capacity[index] = questMapFileTestEntry4D0F60{group: uint32(index), uses: 1}
	}
	capacity[0].uses = 2
	for _, tc := range []struct {
		name    string
		entries []questMapFileTestEntry4D0F60
		last    uint32
		clock   uint32
	}{
		{name: "empty-no-rng", last: 0xffffffff},
		{name: "singleton-no-rng", entries: []questMapFileTestEntry4D0F60{{group: 7, uses: 0xffffffff, last: 0xffffffff}}, last: 0xffffffff},
		{name: "unused-direct-choice", entries: []questMapFileTestEntry4D0F60{{group: 7}, {group: 7}, {group: 7}}, last: 0xffffffff},
		{name: "negative-uses-direct-choice", entries: []questMapFileTestEntry4D0F60{{uses: 0xffffffff}, {uses: 0x80000000}, {uses: 0xfffffffe}}, clock: 1000},
		{name: "negative-uses-eligible", entries: []questMapFileTestEntry4D0F60{{group: 0, uses: 2}, {group: 1, uses: 0xffffffff}, {group: 2}}, clock: 1000},
		{name: "unsigned-max-must-not-replace-positive-max", entries: []questMapFileTestEntry4D0F60{{group: 0, uses: 2}, {group: 1, uses: 0x80000000}, {group: 2}}, clock: 1000},
		{name: "future-clock-excluded", entries: []questMapFileTestEntry4D0F60{{group: 0, uses: 2}, {group: 1, last: 1001}, {group: 2}}, clock: 1000},
		{name: "high-bit-age-excluded", entries: []questMapFileTestEntry4D0F60{{group: 0, uses: 2}, {group: 1}, {group: 2, last: 0x7ffffffb}}, clock: 0x80000000},
		{name: "wrapped-five-ticks-eligible", entries: []questMapFileTestEntry4D0F60{{group: 0, uses: 2}, {group: 1, last: 0xfffffffd}, {group: 2, last: 0xfffffffe}}, clock: 2},
		{name: "four-tick-boundary", entries: []questMapFileTestEntry4D0F60{{group: 0, uses: 2}, {group: 1, last: 996}, {group: 2, last: 995}}, clock: 1000},
		{name: "uniform-threshold-overflow", entries: []questMapFileTestEntry4D0F60{{group: 0, uses: 0x7fffffff}, {group: 1, uses: 0x7fffffff}, {group: 2, uses: 0x7fffffff}}, clock: 1000},
		{name: "uniform-increments-threshold", entries: []questMapFileTestEntry4D0F60{{group: 0, uses: 2}, {group: 1, uses: 2}, {group: 2, uses: 2}}, clock: 1000},
		{name: "same-family-excluded", entries: []questMapFileTestEntry4D0F60{{group: 9, uses: 2}, {group: 9}, {group: 7}}, clock: 1000},
		{name: "empty-eligible-fallback-no-rng", entries: []questMapFileTestEntry4D0F60{{group: 9, uses: 2}, {group: 9}, {group: 9}}, clock: 1000},
		{name: "full-128-entry-table", entries: capacity, clock: 1000},
	} {
		for _, seed := range []int{0, 4095} {
			t.Run(fmt.Sprintf("%s/seed-%d", tc.name, seed), func(t *testing.T) {
				s := server.New(nil, nil, strman.New())
				t.Cleanup(s.Close)
				oldServer := GetServer
				GetServer = func() Server { return &questMapFileLegacyServer4D0F60{native: s} }
				t.Cleanup(func() { GetServer = oldServer })
				count, clock := questMapFileGlobals4D0F60()
				last := memmap.PtrUint32(0x587000, 191880)
				savedCount, savedClock, savedLast := *count, *clock, *last
				storage := memmap.Slice(0x5D4594, 1525132-32)[:32*(128+2)]
				savedStorage := append([]byte(nil), storage...)
				t.Cleanup(func() {
					*count, *clock, *last = savedCount, savedClock, savedLast
					copy(storage, savedStorage)
				})
				for index := range storage {
					storage[index] = 0xc7
				}
				for index, entry := range tc.entries {
					slot := storage[32*(index+1) : 32*(index+2)]
					binary.LittleEndian.PutUint32(slot[0:4], entry.group)
					copy(slot[4:24], fmt.Sprintf("g_map%03d.map\x00", index))
					binary.LittleEndian.PutUint32(slot[24:28], entry.uses)
					binary.LittleEndian.PutUint32(slot[28:32], entry.last)
				}
				*count, *clock, *last = uint32(len(tc.entries)), tc.clock, tc.last
				before := append([]byte(nil), storage...)
				s.Rand.Logic, s.Rand.Other = prand.New(seed), prand.New(93)
				wantRNG := prand.New(seed)
				index := questMapFileReference4D0F60(tc.entries, tc.last, tc.clock, wantRNG)
				var want unsafe.Pointer
				if index != -2 {
					want = memmap.PtrOff(0x5D4594, uintptr(1525136+32*index))
				}
				got := questMapFileCEntry4D0F60()
				if got != want || s.Rand.Logic.Index() != wantRNG.Index() {
					t.Errorf("C entry returned %p / Logic RNG %d, want native map index %d (%p) / RNG %d", got, s.Rand.Logic.Index(), index, want, wantRNG.Index())
				}
				if !bytes.Equal(storage, before) || *count != uint32(len(tc.entries)) || *clock != tc.clock || *last != tc.last || s.Rand.Other.Index() != 93 {
					t.Fatal("map selection modified history, guards, scalar globals or Other RNG")
				}
				if want != nil && unsafe.Sizeof(uintptr(0)) == 8 && uintptr(want) <= math.MaxUint32 {
					t.Fatalf("native map pointer %p must exceed 4 GiB", want)
				}
			})
		}
	}
}
