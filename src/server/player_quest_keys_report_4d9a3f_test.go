package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type questKeyReportState4D9A3F struct {
	cache   [2]uint32
	markers [2][32]uint8
	players [32]bool
	first   int
	types   map[int]uint16
	next    map[int]int
	events  []string
	lookup  func(int) uint32
	send    func(int32, int, uint8)
}

func (f *questKeyReportState4D9A3F) hooks() playerQuestKeysReportHooks4D9A3F[int, int, int] {
	record := func(format string, args ...any) { f.events = append(f.events, fmt.Sprintf(format, args...)) }
	return playerQuestKeysReportHooks4D9A3F[int, int, int]{
		loadCachedType: func(kind int) uint32 { record("cache:%d", kind); return f.cache[kind] },
		lookupType: func(kind int) uint32 {
			record("lookup:%d", kind)
			return f.lookup(kind)
		},
		storeCachedType: func(kind int, value uint32) { record("publish:%d:%d", kind, value); f.cache[kind] = value },
		playerByIndex: func(index int32) int {
			record("player:%d", index)
			if f.players[index] {
				return 100 + int(index)
			}
			return 0
		},
		firstItem: func(unit int) int { record("first:%d", unit); return f.first },
		nextItem:  func(item int) int { record("next:%d", item); return f.next[item] },
		typeInd:   func(item int) uint16 { record("type:%d", item); return f.types[item] },
		marker: func(update, kind int, index int32) uint8 {
			record("marker:%d:%d:%d", update, kind, index)
			return f.markers[kind][index]
		},
		storeMarker: func(update, kind int, index int32, value uint8) {
			record("store:%d:%d:%d:%d", update, kind, index, value)
			f.markers[kind][index] = value
		},
		report: func(index int32, unit, kind int, presence uint8) int32 {
			record("send:%d:%d:%d:%d", index, unit, kind, presence)
			if f.send != nil {
				f.send(index, kind, presence)
			}
			return math.MinInt32
		},
	}
}

func TestPlayerQuestKeysReport4D9A3FOriginalReadOrderAndFirstMatch(t *testing.T) {
	f := &questKeyReportState4D9A3F{first: 10, types: map[int]uint16{10: 0x9234, 11: 0xb678}, next: map[int]int{10: 11}}
	f.players[1] = true
	f.lookup = func(kind int) uint32 { return [2]uint32{0x9234, 0xb678}[kind] }
	playerQuestKeysReport4D9A3F(77, 88, f.hooks())
	var want []string
	for kind := 0; kind < 2; kind++ {
		for index := 0; index < 32; index++ {
			want = append(want, fmt.Sprintf("cache:%d", kind))
			if index == 0 {
				want = append(want, fmt.Sprintf("lookup:%d", kind), fmt.Sprintf("publish:%d:%d", kind, [2]int{0x9234, 0xb678}[kind]))
			}
			want = append(want, fmt.Sprintf("player:%d", index))
			if index == 1 {
				want = append(want, "first:77", fmt.Sprintf("cache:%d", kind), "type:10")
				if kind == 1 {
					want = append(want, "next:10", "cache:1", "type:11")
				}
				want = append(want, fmt.Sprintf("marker:88:%d:1", kind), fmt.Sprintf("send:1:77:%d:1", kind), fmt.Sprintf("store:88:%d:1:1", kind))
			}
		}
	}
	if !reflect.DeepEqual(f.events, want) {
		t.Fatalf("events=%v\nwant=%v", f.events, want)
	}
}

func TestPlayerQuestKeysReport4D9A3FPublishesZeroAndRetriesEvenWithoutRecipients(t *testing.T) {
	f := &questKeyReportState4D9A3F{lookup: func(int) uint32 { return 0 }}
	playerQuestKeysReport4D9A3F(0, 0, f.hooks())
	var want []string
	for kind := 0; kind < 2; kind++ {
		for index := 0; index < 32; index++ {
			want = append(want, fmt.Sprintf("cache:%d", kind), fmt.Sprintf("lookup:%d", kind), fmt.Sprintf("publish:%d:0", kind), fmt.Sprintf("player:%d", index))
		}
	}
	if !reflect.DeepEqual(f.events, want) {
		t.Fatalf("missing zero publication or per-slot cache retry: %v", f.events)
	}
}

func TestPlayerQuestKeysReport4D9A3FReloadsFullCacheBeforeEachTypeWORD(t *testing.T) {
	f := &questKeyReportState4D9A3F{cache: [2]uint32{0x19234, 0x1b678}, first: 10,
		types: map[int]uint16{10: 0x9234, 11: 0xb678}, next: map[int]int{10: 11}}
	f.players[31] = true
	f.markers[0][31], f.markers[1][31] = 1, 1
	playerQuestKeysReport4D9A3F(77, 88, f.hooks())
	if f.markers[0][31] != 0 || f.markers[1][31] != 0 {
		t.Fatal("cache DWORD was narrowed to the item WORD")
	}
	f.cache = [2]uint32{123, 456}
	f.events = nil
	h := f.hooks()
	h.nextItem = func(item int) int {
		// First mismatch changes the shared cache before the second candidate.
		f.cache[0], f.cache[1] = 0xb678, 0xb678
		return f.next[item]
	}
	playerQuestKeysReport4D9A3F(77, 88, h)
	if f.markers[0][31] != 1 || f.markers[1][31] != 1 {
		t.Fatal("cache or item type was snapshotted before live inventory traversal")
	}
}

func TestPlayerQuestKeysReport4D9A3FSendStoresPrecomputedPresenceAndLaterSlotsReadLiveInventory(t *testing.T) {
	f := &questKeyReportState4D9A3F{cache: [2]uint32{11, 22}, first: 10, types: map[int]uint16{10: 11}}
	f.players[1], f.players[3] = true, true
	f.markers[0][3] = 1
	f.send = func(index int32, kind int, presence uint8) {
		if kind == 0 && index == 1 {
			f.first = 0
			f.markers[0][1] = 0xff // The old computed presence must overwrite this.
			f.players[2] = true    // Later slots must perform a fresh player lookup.
		}
	}
	playerQuestKeysReport4D9A3F(77, 88, f.hooks())
	if f.markers[0][1] != 1 || f.markers[0][3] != 0 || f.markers[0][2] != 0 {
		t.Fatal("send callback changed acknowledgment or later inventory snapshot")
	}
	var sends []string
	for _, event := range f.events {
		if len(event) >= 5 && event[:5] == "send:" {
			sends = append(sends, event)
		}
	}
	if !reflect.DeepEqual(sends, []string{"send:1:77:0:1", "send:3:77:0:0"}) {
		t.Fatalf("sends=%v", sends)
	}
}

func TestPlayerQuestKeysReport4D9A3FFaultPrefixes(t *testing.T) {
	for _, failure := range []string{"cache:0", "lookup:0", "publish:0", "player:0", "first", "candidate-cache", "type", "marker", "send", "store"} {
		t.Run(failure, func(t *testing.T) {
			var events []string
			step := func(name string) {
				events = append(events, name)
				if name == failure {
					panic(name)
				}
			}
			cacheReads := 0
			h := playerQuestKeysReportHooks4D9A3F[int, int, int]{
				loadCachedType: func(int) uint32 {
					cacheReads++
					if cacheReads == 1 {
						step("cache:0")
						return 0
					}
					step("candidate-cache")
					return 11
				},
				lookupType:      func(int) uint32 { step("lookup:0"); return 11 },
				storeCachedType: func(int, uint32) { step("publish:0") },
				playerByIndex:   func(int32) int { step("player:0"); return 1 },
				firstItem:       func(int) int { step("first"); return 10 },
				nextItem:        func(int) int { t.Fatal("first match must not read next"); return 0 },
				typeInd:         func(int) uint16 { step("type"); return 11 },
				marker:          func(int, int, int32) uint8 { step("marker"); return 0 },
				report:          func(int32, int, int, uint8) int32 { step("send"); return -1 },
				storeMarker:     func(int, int, int32, uint8) { step("store") },
			}
			order := []string{"cache:0", "lookup:0", "publish:0", "player:0", "first", "candidate-cache", "type", "marker", "send", "store"}
			defer func() {
				if got := recover(); got != failure {
					t.Fatalf("fault=%v want=%s", got, failure)
				}
				for index, name := range order {
					if name == failure && !reflect.DeepEqual(events, order[:index+1]) {
						t.Fatalf("fault prefix=%v want=%v", events, order[:index+1])
					}
				}
			}()
			playerQuestKeysReport4D9A3F(77, 88, h)
		})
	}
}

func questKeyNativeFixture4D9A3F(t *testing.T) (*Server, *Object, *PlayerUpdateData) {
	t.Helper()
	// Standalone server tests do not load the legacy C data blobs. Preserve
	// previously registered data when this fixture needs a larger test view.
	const base, need = 0x5D4594, 1556328 + 4
	if blob := memmap.BlobByAddr(base); blob == nil {
		memmap.RegisterBlobData(base, "quest_key_report_test", make([]byte, need))
	} else if len(blob.Data) < need {
		previous := *blob
		data := make([]byte, need)
		copy(data, blob.Data)
		blob.Data, blob.Size = data, uintptr(len(data))
		t.Cleanup(func() { *blob = previous })
	}
	s := New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	unit, freeUnit := alloc.New(Object{})
	update, freeUpdate := alloc.New(PlayerUpdateData{})
	t.Cleanup(freeUnit)
	t.Cleanup(freeUpdate)
	unit.NetCode, unit.UpdateData = 0xffff9234, unsafe.Pointer(update)
	for kind, value := range []uint32{0x9234, 0xb678} {
		cache := memmap.PtrT[uint32](0x5D4594, uintptr(1556324+4*kind))
		old := *cache
		*cache = value
		t.Cleanup(func() { *cache = old })
	}
	for index := 0; index < 32; index++ {
		s.Players.list[index].Active = 1
		// No unit: unlike the life report, the original key loop includes these.
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(unit), unsafe.Pointer(update), unsafe.Pointer(&s.Players.list[31])} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatal("native key report fixture must exceed 4 GiB")
			}
		}
	}
	return s, unit, update
}

func TestPlayerQuestKeysReportNative4D9A3FAllMarkerBytesPresenceAndSlots(t *testing.T) {
	s, unit, update := questKeyNativeFixture4D9A3F(t)
	silver, freeSilver := alloc.New(Object{})
	gold, freeGold := alloc.New(Object{})
	t.Cleanup(freeSilver)
	t.Cleanup(freeGold)
	silver.TypeInd, gold.TypeInd = 0x9234, 0xb678
	for presence := 0; presence < 4; presence++ {
		unit.InvFirstItem, silver.InvNextItem = nil, nil
		if presence&1 != 0 {
			unit.InvFirstItem = silver
		}
		if presence&2 != 0 {
			if unit.InvFirstItem != nil {
				silver.InvNextItem = gold
			} else {
				unit.InvFirstItem = gold
			}
		}
		for marker := 0; marker < 256; marker++ {
			for index := 0; index < 32; index++ {
				update.QuestPlayerFlagsA[index], update.QuestPlayerFlagsB[index] = byte(marker), byte(marker)
			}
			calls := 0
			var seen [2][32]bool
			s.NetSendPacketXxx = func(index int, packet []byte, related *Object, remove, sequence int) int {
				if len(packet) != 5 || index < 0 || index >= 32 || packet[1] < 22 || packet[1] > 23 {
					t.Fatalf("unexpected key packet: recipient=%d packet=%x", index, packet)
				}
				kind := int(packet[1]) - 22
				want := byte((presence >> kind) & 1)
				if seen[kind][index] ||
					!reflect.DeepEqual(packet, []byte{0xf0, byte(22 + kind), want, 0x34, 0x92}) || related != nil || remove != 1 || sequence != 0 {
					t.Fatalf("presence=%d marker=%d send=%d/%x", presence, marker, index, packet)
				}
				seen[kind][index] = true
				calls++
				return math.MinInt32
			}
			s.PlayerQuestKeysReport4D9A3F(unit, update)
			wantCalls := 0
			for kind := 0; kind < 2; kind++ {
				want := byte((presence >> kind) & 1)
				if want != byte(marker) {
					wantCalls += 32
				}
				for index := 0; index < 32; index++ {
					actual := update.QuestPlayerFlagsA[index]
					if kind == 1 {
						actual = update.QuestPlayerFlagsB[index]
					}
					if actual != want || seen[kind][index] != (want != byte(marker)) {
						t.Fatalf("presence=%d marker=%d kind=%d slot=%d cache=%d sent=%v", presence, marker, kind, index, actual, seen[kind][index])
					}
				}
			}
			if calls != wantCalls || update.ExtraLives != 0 || update.RespawnMarkers != [32]byte{} || update.Player != nil || unit.UpdateData != unsafe.Pointer(update) {
				t.Fatalf("presence=%d marker=%d calls=%d want=%d or unrelated update mutation", presence, marker, calls, wantCalls)
			}
			s.PlayerQuestKeysReport4D9A3F(unit, update)
			if calls != wantCalls {
				t.Fatal("unchanged keys resent")
			}
		}
	}
}

func TestQuestKeyReport4D9DF0WireAndSignedResult(t *testing.T) {
	s, unit, _ := questKeyNativeFixture4D9A3F(t)
	for _, recipient := range []int32{math.MinInt32, -1, 0, 31, 32, 255, math.MaxInt32} {
		for _, result := range []int32{math.MinInt32, -1, 0, 1, math.MaxInt32} {
			for kind := 0; kind < 2; kind++ {
				for presence := 0; presence < 256; presence++ {
					unit.NetCode = uint32(presence)<<24 | 0x80ff
					calls := 0
					s.NetSendPacketXxx = func(index int, packet []byte, related *Object, remove, sequence int) int {
						calls++
						if index != int(recipient) || !reflect.DeepEqual(packet, []byte{0xf0, byte(22 + kind), byte(presence), 0xff, 0x80}) || related != nil || remove != 1 || sequence != 0 {
							t.Fatalf("sender=%d/%x/%p/%d/%d", index, packet, related, remove, sequence)
						}
						return int(result)
					}
					if got := s.questKeyReport4D9DF0(recipient, unit, kind, byte(presence)); got != result || calls != 1 {
						t.Fatalf("return=%d want=%d calls=%d", got, result, calls)
					}
				}
			}
		}
	}
}

func TestPlayerQuestKeysReportNative4D9A3FStoresEntryUpdateAcrossCallbackAndIncludesUnitlessPlayer(t *testing.T) {
	s, unit, entry := questKeyNativeFixture4D9A3F(t)
	live, freeLive := alloc.New(PlayerUpdateData{})
	key, freeKey := alloc.New(Object{})
	t.Cleanup(freeLive)
	t.Cleanup(freeKey)
	key.TypeInd = 0x9234
	for index := range s.Players.list {
		s.Players.list[index].Active = 0
	}
	s.Players.list[1].Active, s.Players.list[3].Active = 1, 1
	entry.QuestPlayerFlagsA[3], entry.QuestPlayerFlagsA[4] = 1, 0x77
	unit.InvFirstItem = key
	var recipients []int
	s.NetSendPacketXxx = func(index int, packet []byte, _ *Object, _, _ int) int {
		recipients = append(recipients, index)
		if index == 1 {
			unit.UpdateData, unit.InvFirstItem = unsafe.Pointer(live), nil
			entry.QuestPlayerFlagsA[1] = 0xaa
			s.Players.list[2].Active = 1
		}
		return -1
	}
	s.PlayerQuestKeysReport4D9A3F(unit, entry)
	if !reflect.DeepEqual(recipients, []int{1, 3}) || entry.QuestPlayerFlagsA[1] != 1 || entry.QuestPlayerFlagsA[3] != 0 ||
		entry.QuestPlayerFlagsA[4] != 0x77 || *live != (PlayerUpdateData{}) || unit.UpdateData != unsafe.Pointer(live) {
		t.Fatalf("recipients=%v entry=%v live=%v", recipients, entry.QuestPlayerFlagsA, live.QuestPlayerFlagsA)
	}
}

func TestPlayerQuestKeysReportNative4D9A3FInitializesLiteralTypesWithoutRecipients(t *testing.T) {
	s, unit, update := questKeyNativeFixture4D9A3F(t)
	s.Types.byID = map[string]*ObjectType{
		"silverkey": {ind: 0x9234},
		"goldkey":   {ind: 0xb678},
	}
	for index := range s.Players.list {
		s.Players.list[index].Active = 0
	}
	for kind := 0; kind < 2; kind++ {
		*memmap.PtrT[uint32](0x5D4594, uintptr(1556324+4*kind)) = 0
	}
	s.NetSendPacketXxx = func(int, []byte, *Object, int, int) int {
		t.Fatal("absent recipients must not send")
		return 0
	}
	// No unit/update access is allowed when every player lookup misses, but
	// both literal type lookups must still publish their original DWORDs.
	s.PlayerQuestKeysReport4D9A3F(nil, nil)
	if silver, gold := *memmap.PtrT[uint32](0x5D4594, 1556324), *memmap.PtrT[uint32](0x5D4594, 1556328); silver != 0x9234 || gold != 0xb678 ||
		*update != (PlayerUpdateData{}) || unit.UpdateData != unsafe.Pointer(update) {
		t.Fatalf("literal type caches=%x/%x or unrelated state mutation", silver, gold)
	}
}
