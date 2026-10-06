package legacy

import (
	"bytes"
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func questMapSelectionObserverFixture4D0F60(t *testing.T, countValue uint32) (*server.Server, []byte) {
	t.Helper()
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	oldServer, oldSelector := GetServer, questMapFileCall4D0F60
	GetServer = func() Server { return &questMapFileLegacyServer4D0F60{native: s} }
	t.Cleanup(func() { GetServer, questMapFileCall4D0F60 = oldServer, oldSelector })
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
	for index := int32(0); index < 128; index++ {
		entry := questMapEntry4D0F60{Group: uint32(index), Uses: 0, LastUsed: 0}
		copy(entry.Name[:], fmt.Sprintf("g_map%03d.map", index))
		*questMapEntryNative4D0F60(index) = entry
	}
	*count, *clock, *last = countValue, 1000, 0xffffffff
	s.Rand.Logic, s.Rand.Other = prand.New(4095), prand.New(93)
	return s, storage
}

func TestQuestMapSelectionObserver4D0F60PreservesDelegateAndDetachedCopies(t *testing.T) {
	for _, tc := range []struct {
		name  string
		count uint32
		index int32
	}{
		{name: "null", count: 0, index: -2},
		{name: "last-full-native-slot", count: 128, index: 127},
		{name: "fallback-slot", count: 3, index: -1},
		{name: "unknown-pointer", count: 3, index: -3},
		{name: "invalid-count-bounded-diagnostic", count: 129, index: -2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, storage := questMapSelectionObserverFixture4D0F60(t, tc.count)
			before := append([]byte(nil), storage...)
			var result unsafe.Pointer
			switch tc.index {
			case -2:
			case -3:
				result = unsafe.Add(questMapNameNative4D0F60(127), 1)
			default:
				result = questMapNameNative4D0F60(tc.index)
			}
			if result != nil && unsafe.Sizeof(uintptr(0)) == 8 && uintptr(result) <= math.MaxUint32 {
				t.Fatalf("native pointer %p must exceed 4 GiB", result)
			}
			calls, observed := 0, 0
			questMapFileCall4D0F60 = func() unsafe.Pointer { calls++; return result }
			stop := ObserveQuestMapSelection4D0F60(func(selection QuestMapSelection4D0F60) {
				observed++
				if selection.ResultPointer != uintptr(result) || selection.ResultIndex != tc.index ||
					selection.Before.Count != tc.count || !reflect.DeepEqual(selection.Before, selection.After) ||
					selection.Before.LogicIndex != 4095 || selection.Before.OtherIndex != 93 {
					t.Fatalf("observer changed result/state: %+v", selection)
				}
				wantLen := int(tc.count)
				if tc.count > 128 {
					wantLen = 0
				}
				if len(selection.Before.Entries) != wantLen || len(selection.After.Entries) != wantLen {
					t.Fatalf("diagnostic records=%d/%d want %d", len(selection.Before.Entries), len(selection.After.Entries), wantLen)
				}
				if wantLen > 0 {
					selection.Before.Entries[0].Group = 42
					selection.Before.Entries[0].Name[0] = 'x'
					selection.Before.Entries[0].Uses = 43
					selection.Before.Entries[0].LastUsed = 44
					if selection.After.Entries[0].Group != 0 || selection.After.Entries[0].Name[0] != 'g' ||
						selection.After.Entries[0].Uses != 0 || selection.After.Entries[0].LastUsed != 0 {
						t.Fatal("before and after snapshot records alias each other")
					}
					selection.After.Entries[wantLen-1] = QuestMapHistory4D0F60{}
				}
				selection.Before.Count, selection.After.Clock = 99, 42
			})
			t.Cleanup(stop)
			if got := questMapFileCEntry4D0F60(); got != result || calls != 1 || observed != 1 {
				t.Fatalf("C delegate=%p calls=%d observed=%d, want %p/1/1", got, calls, observed, result)
			}
			count, clock := questMapFileGlobals4D0F60()
			if !bytes.Equal(storage, before) || *count != tc.count || *clock != 1000 ||
				memmap.Uint32(0x587000, 191880) != 0xffffffff || s.Rand.Logic.Index() != 4095 || s.Rand.Other.Index() != 93 {
				t.Fatal("observer/callback edited native records, guards, globals or RNG")
			}
			stop()
			stop()
			if got := questMapFileCEntry4D0F60(); got != result || calls != 2 || observed != 1 {
				t.Fatalf("stopped delegate=%p calls=%d observed=%d, want %p/2/1", got, calls, observed, result)
			}
		})
	}
}

func TestQuestMapSelectionObserver4D0F60RealSelectorConsumesOnlyOriginalRNG(t *testing.T) {
	s, storage := questMapSelectionObserverFixture4D0F60(t, 13)
	before := append([]byte(nil), storage...)
	random := prand.New(4095)
	wantIndex := random.Int(0, 12)
	want := questMapNameNative4D0F60(int32(wantIndex))
	delegate, calls, observed := questMapFileCall4D0F60, 0, 0
	questMapFileCall4D0F60 = func() unsafe.Pointer { calls++; return delegate() }
	stop := ObserveQuestMapSelection4D0F60(func(selection QuestMapSelection4D0F60) {
		observed++
		after := selection.After
		after.LogicIndex = selection.Before.LogicIndex
		if selection.ResultPointer != uintptr(want) || selection.ResultIndex != int32(wantIndex) ||
			selection.Before.LogicIndex != 4095 || selection.After.LogicIndex != 0 ||
			!reflect.DeepEqual(after, selection.Before) {
			t.Fatalf("observed original selector state/result differs: %+v", selection)
		}
	})
	t.Cleanup(stop)
	if got := questMapFileCEntry4D0F60(); got != want || calls != 1 || observed != 1 ||
		s.Rand.Logic.Index() != 0 || s.Rand.Other.Index() != 93 || !bytes.Equal(storage, before) {
		t.Fatalf("observer altered original call/result/RNG/history: result=%p/%p calls=%d observed=%d logic=%d other=%d", got, want, calls, observed, s.Rand.Logic.Index(), s.Rand.Other.Index())
	}
}

func TestQuestMapSelectionObserver4D0F60NestedStopOrder(t *testing.T) {
	questMapSelectionObserverFixture4D0F60(t, 1)
	var events []string
	want := questMapNameNative4D0F60(0)
	questMapFileCall4D0F60 = func() unsafe.Pointer { events = append(events, "delegate"); return want }
	stopFirst := ObserveQuestMapSelection4D0F60(func(QuestMapSelection4D0F60) { events = append(events, "first") })
	t.Cleanup(stopFirst)
	stopSecond := ObserveQuestMapSelection4D0F60(func(QuestMapSelection4D0F60) { events = append(events, "second") })
	t.Cleanup(stopSecond)
	if got := questMapFileCEntry4D0F60(); got != want || !reflect.DeepEqual(events, []string{"delegate", "first", "second"}) {
		t.Fatalf("nested result/events=%p/%v", got, events)
	}
	stopSecond()
	stopSecond()
	events = nil
	questMapFileCEntry4D0F60()
	if !reflect.DeepEqual(events, []string{"delegate", "first"}) {
		t.Fatalf("second observer did not restore first: %v", events)
	}
	stopFirst()
	events = nil
	questMapFileCEntry4D0F60()
	if !reflect.DeepEqual(events, []string{"delegate"}) {
		t.Fatalf("first observer did not restore delegate: %v", events)
	}
}

func TestQuestMapSelectionObserver4D0F60NilDoesNotReadServer(t *testing.T) {
	oldServer, oldSelector := GetServer, questMapFileCall4D0F60
	t.Cleanup(func() { GetServer, questMapFileCall4D0F60 = oldServer, oldSelector })
	GetServer = func() Server { t.Fatal("nil observer read server"); return nil }
	calls := 0
	questMapFileCall4D0F60 = func() unsafe.Pointer { calls++; return nil }
	stop := ObserveQuestMapSelection4D0F60(nil)
	stop()
	stop()
	if got := questMapFileCEntry4D0F60(); got != nil || calls != 1 {
		t.Fatalf("nil observer changed delegate: %p/%d", got, calls)
	}
}
