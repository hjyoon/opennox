package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

func TestQuestLivesSync4D99E1ExactReadOrderAndSkippedSlots(t *testing.T) {
	var trace, want []string
	hit := func(event string) { trace = append(trace, event) }
	questLivesSync4D99E1(77, 88, questLivesSyncHooks4D99E1[int, int, int]{
		gameFlag: func(flag uint32) int32 {
			hit("flag")
			if flag != 0x1000 {
				t.Fatalf("flag=%08x", flag)
			}
			return math.MinInt32
		},
		playerByIndex: func(index int32) int {
			hit(fmt.Sprintf("player:%d", index))
			if index < 3 {
				return int(index) + 1
			}
			return 0
		},
		playerUnit: func(player int) int {
			hit(fmt.Sprintf("unit:%d", player))
			if player == 2 {
				return 0
			}
			return 11
		},
		lives: func(update int) uint32 {
			hit("lives")
			if update != 88 {
				t.Fatal("lost entry update")
			}
			return 3
		},
		marker: func(update int, index int32) uint8 {
			hit(fmt.Sprintf("marker:%d", index))
			if update != 88 {
				t.Fatal("lost entry update")
			}
			if index == 2 {
				return 3
			}
			return 1
		},
		report: func(index int32, unit int) int32 {
			hit(fmt.Sprintf("report:%d", index))
			if index != 0 || unit != 77 {
				t.Fatal("report arguments")
			}
			return -1
		},
		lifeByte: func(update int) uint8 {
			hit("byte")
			if update != 88 {
				t.Fatal("lost entry update")
			}
			return 5
		},
		storeMarker: func(update int, index int32, value uint8) {
			hit(fmt.Sprintf("store:%d", index))
			if update != 88 || index != 0 || value != 5 {
				t.Fatal("marker arguments")
			}
		},
	})
	want = []string{"flag", "player:0", "unit:1", "lives", "marker:0", "report:0", "byte", "store:0", "player:1", "unit:2", "player:2", "unit:3", "lives", "marker:2"}
	for index := 3; index < 32; index++ {
		want = append(want, fmt.Sprintf("player:%d", index))
	}
	if !reflect.DeepEqual(trace, want) {
		t.Fatalf("trace=%v want=%v", trace, want)
	}
}

func TestQuestLivesSync4D99E1ZeroFlagAndFullSignedNonzeroFlags(t *testing.T) {
	for _, flag := range []int32{0, 1, -1, math.MinInt32, math.MaxInt32} {
		calls := 0
		questLivesSync4D99E1(0, 0, questLivesSyncHooks4D99E1[int, int, int]{
			gameFlag: func(uint32) int32 { return flag },
			playerByIndex: func(index int32) int {
				if index != int32(calls) {
					t.Fatal("lookup order")
				}
				calls++
				return 0
			},
		})
		want := 32
		if flag == 0 {
			want = 0
		}
		if calls != want {
			t.Fatalf("flag=%d lookups=%d want=%d", flag, calls, want)
		}
	}
}

func TestQuestLivesSync4D99E1FaultPrefixes(t *testing.T) {
	steps := []string{"flag", "player", "unit", "lives", "marker", "report", "byte", "store"}
	for fault := range steps {
		t.Run(steps[fault], func(t *testing.T) {
			var trace []string
			hit := func(step int) {
				if step == fault {
					panic(steps[step])
				}
				trace = append(trace, steps[step])
			}
			defer func() {
				if got := recover(); got != steps[fault] || len(trace) != fault || fault > 0 && !reflect.DeepEqual(trace, steps[:fault]) {
					t.Fatalf("fault=%v trace=%v want=%v", got, trace, steps[:fault])
				}
			}()
			questLivesSync4D99E1(1, 2, questLivesSyncHooks4D99E1[int, int, int]{
				gameFlag:      func(uint32) int32 { hit(0); return 1 },
				playerByIndex: func(int32) int { hit(1); return 3 },
				playerUnit:    func(int) int { hit(2); return 4 },
				lives:         func(int) uint32 { hit(3); return 2 },
				marker:        func(int, int32) uint8 { hit(4); return 1 },
				report:        func(int32, int) int32 { hit(5); return math.MinInt32 },
				lifeByte:      func(int) uint8 { hit(6); return 7 },
				storeMarker:   func(int, int32, uint8) { hit(7) },
			})
			t.Fatal("missing fault")
		})
	}
}

func questLivesSyncNativeFixture4D99E1(t *testing.T) *Server {
	t.Helper()
	old := noxflags.GetGame()
	noxflags.SetGame(noxflags.GameModeQuest)
	t.Cleanup(func() { noxflags.ResetGame(); noxflags.SetGame(old) })
	s := &Server{}
	s.Players.list = make([]Player, 32)
	return s
}

func TestQuestLivesSyncNative4D99E1AllSlotsAndDWORDVersusBYTE(t *testing.T) {
	s := questLivesSyncNativeFixture4D99E1(t)
	update := &PlayerUpdateData{}
	unit := &Object{NetCode: 0xabcd9234, UpdateData: unsafe.Pointer(update)} // No class gate inside the original slice.
	for i := range s.Players.list {
		s.Players.list[i] = Player{Active: 1, PlayerUnit: unit, PlayerInd: 255}
	}
	values := []uint32{0x100, 0x101, 0x80000000, math.MaxUint32}
	for value := uint32(0); value < 256; value++ {
		values = append(values, value)
	}
	for _, lives := range values {
		update.ExtraLives = lives
		for i := range update.RespawnMarkers {
			update.RespawnMarkers[i] = uint8((uint64(lives) + 1) % 256)
		}
		calls := 0
		s.NetSendPacketXxx = func(index int, packet []byte, related *Object, remove, sequence int) int {
			if index != calls%32 || !reflect.DeepEqual(packet, []byte{0xf0, 4, byte(uint64(lives) % 256), 0x34, 0x92}) || related != nil || remove != 1 || sequence != 0 {
				t.Fatalf("lives=%08x send=%d/%x/%p/%d/%d", lives, index, packet, related, remove, sequence)
			}
			calls++
			return math.MinInt32
		}
		s.PlayerQuestLivesReport4D99E1(unit, update)
		wantUpdate := PlayerUpdateData{ExtraLives: lives}
		for i := range wantUpdate.RespawnMarkers {
			wantUpdate.RespawnMarkers[i] = byte(uint64(lives) % 256)
		}
		if calls != 32 || *update != wantUpdate {
			t.Fatalf("lives=%08x first calls=%d or unexpected field mutation", lives, calls)
		}
		s.PlayerQuestLivesReport4D99E1(unit, update)
		want := 32
		if uint64(lives) >= 256 {
			want = 64
		} // Full DWORD still differs from every zero-extended marker.
		if calls != want || *update != wantUpdate {
			t.Fatalf("lives=%08x repeat calls=%d want=%d", lives, calls, want)
		}
		for i := range s.Players.list {
			if s.Players.list[i].PlayerInd != byte(i) {
				t.Fatalf("ByInd did not preserve lookup index %d", i)
			}
		}
	}
}

func TestQuestLivesSyncNative4D99E1RetainsEntryUpdateAndUsesFreshSenderAndByte(t *testing.T) {
	s := questLivesSyncNativeFixture4D99E1(t)
	entry := &PlayerUpdateData{ExtraLives: 2}
	live := &PlayerUpdateData{ExtraLives: 7}
	unit := &Object{NetCode: 0x1234, UpdateData: unsafe.Pointer(live)}
	s.Players.list[3] = Player{Active: 1, PlayerUnit: unit}
	s.Players.list[4] = Player{Active: 0, PlayerUnit: unit} // Becomes eligible during the preceding send.
	s.Players.list[5] = Player{Active: 1}                   // No unit, so retain marker.
	entry.RespawnMarkers[5] = 0xaa
	calls := 0
	s.NetSendPacketXxx = func(index int, packet []byte, related *Object, remove, sequence int) int {
		calls++
		if index != 2+calls || !reflect.DeepEqual(packet, []byte{0xf0, 4, 7, 0x34, 0x12}) || related != nil || remove != 1 || sequence != 0 {
			t.Fatalf("live sender=%d/%x", index, packet)
		}
		entry.ExtraLives = 0x105
		s.Players.list[4].Active = 1
		return -1
	}
	s.PlayerQuestLivesReport4D99E1(unit, entry)
	if calls != 2 || entry.ExtraLives != 0x105 || entry.RespawnMarkers[3] != 5 || entry.RespawnMarkers[4] != 5 || entry.RespawnMarkers[5] != 0xaa || *live != (PlayerUpdateData{ExtraLives: 7}) || unit.UpdateData != unsafe.Pointer(live) {
		t.Fatal("lost cached update, post-send life byte, fresh later lookup, or skipped marker")
	}
}

func TestQuestLivesSyncNative4D99E1KeepsFlagAndDependentNilFaults(t *testing.T) {
	s := questLivesSyncNativeFixture4D99E1(t)
	s.Players.list[31] = Player{Active: 1, PlayerUnit: &Object{}}
	noxflags.UnsetGame(noxflags.GameModeQuest)
	s.PlayerQuestLivesReport4D99E1(nil, nil) // No update reads outside Quest.
	noxflags.SetGame(noxflags.GameModeQuest)
	s.Players.list[31].PlayerUnit = nil
	s.PlayerQuestLivesReport4D99E1(nil, nil) // No dependent read without a recipient unit.
	s.Players.list[31].PlayerUnit = &Object{}
	defer func() {
		if recover() == nil {
			t.Fatal("missing original update read fault")
		}
	}()
	s.PlayerQuestLivesReport4D99E1(nil, nil)
}
