package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"
)

type questLivesReportTestUpdate4D9D60 struct{ lives uint32 }
type questLivesReportTestUnit4D9D60 struct {
	code   uint32
	update *questLivesReportTestUpdate4D9D60
}

func TestQuestLivesReport4D9D60ReadOrderAndLiveUpdate(t *testing.T) {
	first := &questLivesReportTestUpdate4D9D60{lives: 7}
	live := &questLivesReportTestUpdate4D9D60{lives: 0x89abcdef}
	unit := &questLivesReportTestUnit4D9D60{code: 0xabcd9234, update: first}
	var trace []string
	got := questLivesReport4D9D60(-77, unit, questLivesReportHooks4D9D60[*questLivesReportTestUnit4D9D60, *questLivesReportTestUpdate4D9D60]{
		netCode: func(u *questLivesReportTestUnit4D9D60) uint16 {
			trace = append(trace, "netcode")
			code := uint16(u.code)
			u.code, u.update = 0x5555, live
			return code
		},
		update: func(u *questLivesReportTestUnit4D9D60) *questLivesReportTestUpdate4D9D60 {
			trace = append(trace, "update")
			return u.update
		},
		lives: func(u *questLivesReportTestUpdate4D9D60) uint8 {
			trace = append(trace, "lives")
			if u != live {
				t.Fatal("UpdateData must be loaded after netcode")
			}
			return uint8(u.lives)
		},
		send: func(recipient int32, packet [5]byte) int32 {
			trace = append(trace, "send")
			if recipient != -77 || packet != [5]byte{0xf0, 4, 0xef, 0x34, 0x92} {
				t.Fatalf("send=%d/%x", recipient, packet)
			}
			return math.MinInt32
		},
	})
	if got != math.MinInt32 || !reflect.DeepEqual(trace, []string{"netcode", "update", "lives", "send"}) {
		t.Fatalf("return=%d trace=%v", got, trace)
	}
}

func TestQuestLivesReport4D9D60FaultPrefixes(t *testing.T) {
	steps := []string{"netcode", "update", "lives", "send"}
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
				if got := recover(); got != steps[fault] || !reflect.DeepEqual(trace, steps[:fault]) && !(fault == 0 && len(trace) == 0) {
					t.Fatalf("fault=%v trace=%v want=%v", got, trace, steps[:fault])
				}
			}()
			questLivesReport4D9D60(31, 1, questLivesReportHooks4D9D60[int, int]{
				netCode: func(int) uint16 { hit(0); return 0x1234 },
				update:  func(int) int { hit(1); return 2 },
				lives:   func(int) uint8 { hit(2); return 3 },
				send:    func(int32, [5]byte) int32 { hit(3); return 0 },
			})
			t.Fatal("missing fault")
		})
	}
}

func TestQuestLivesReportNative4D9D60WireWidthsAndReturn(t *testing.T) {
	update := &PlayerUpdateData{}
	unit := &Object{UpdateData: unsafe.Pointer(update)} // No Player-class gate in the sender.
	s := &Server{}
	lives := []uint32{0x100, 0x101, 0x12345678, 0x80000000, math.MaxUint32}
	for value := uint32(0); value < 256; value++ {
		lives = append(lives, value)
	}
	codes := []uint32{0, 1, 0x7fff, 0x8000, 0xffff, 0x10000, 0x12349234, math.MaxUint32}
	for index, value := range lives {
		unit.NetCode, update.ExtraLives = codes[index%len(codes)], value
		want := [5]byte{0xf0, 4, byte(uint64(value) % 256), byte(uint64(unit.NetCode) % 256), byte(uint64(unit.NetCode) / 256 % 256)}
		for _, recipient := range []int32{math.MinInt32, -1, 0, 31, 255, math.MaxInt32} {
			for _, result := range []int32{math.MinInt32, -1, 0, 1, math.MaxInt32} {
				calls := 0
				s.NetSendPacketXxx = func(ind int, packet []byte, related *Object, remove, sequence int) int {
					calls++
					if ind != int(recipient) || len(packet) != len(want) || !reflect.DeepEqual(packet, want[:]) || related != nil || remove != 1 || sequence != 0 {
						t.Fatalf("send=%d/%x/%p/%d/%d want=%d/%x/nil/1/0", ind, packet, related, remove, sequence, recipient, want)
					}
					return int(result)
				}
				beforeUnit, beforeUpdate := *unit, *update
				if got := s.QuestLivesReport4D9D60(recipient, unit); got != result || calls != 1 || *unit != beforeUnit || *update != beforeUpdate {
					t.Fatalf("lives=%08x return=%d want=%d calls=%d or sender changed state", value, got, result, calls)
				}
			}
		}
	}
}

func TestQuestLivesReportNative4D9D60KeepsOriginalNilFaults(t *testing.T) {
	for _, unit := range []*Object{nil, {NetCode: 0x1234}} {
		t.Run(fmt.Sprintf("%p", unit), func(t *testing.T) {
			s := &Server{NetSendPacketXxx: func(int, []byte, *Object, int, int) int { t.Fatal("send before dependent read fault"); return 0 }}
			defer func() {
				if recover() == nil {
					t.Fatal("original unit/update read must fault")
				}
			}()
			s.QuestLivesReport4D9D60(31, unit)
		})
	}
}
