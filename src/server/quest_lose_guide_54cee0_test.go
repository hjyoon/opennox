package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/prand"
)

func TestQuestLoseGuide54CEE0ExactOrderAndCachedPlayer(t *testing.T) {
	type player struct {
		levels [41]uint32
		index  uint8
	}
	type update struct{ player *player }
	type unit struct{ update *update }
	p := &player{index: 9}
	p.levels[0], p.levels[2], p.levels[39], p.levels[40] = 1, 1, 1, 1
	p.levels[3], p.levels[4], p.levels[5] = 2, 0x80000001, math.MaxUint32
	u := &unit{update: &update{player: p}}
	replacement := &player{index: 3}
	replacement.levels[39] = 1
	var trace []string
	questLoseGuide54CEE0(u, questLoseGuideHooks54CEE0[*unit, *update, *player]{
		loadUpdate: func(got *unit) *update { trace = append(trace, "update"); return got.update },
		loadPlayer: func(got *update) *player { trace = append(trace, "player"); return got.player },
		loadClass:  func(got *player) uint8 { trace = append(trace, "class"); return 2 },
		loadLevel: func(got *player, id int32) uint32 {
			if got != p {
				t.Fatal("cached Player was replaced")
			}
			trace = append(trace, fmt.Sprintf("level:%d", id))
			return got.levels[id]
		},
		eligible: func(id int32) int32 {
			trace = append(trace, fmt.Sprintf("eligible:%d", id))
			return RandomFieldGuideLossEligible4F2530(id)
		},
		randomInt: func(minimum, maximum int32) int32 {
			trace = append(trace, "rng")
			if minimum != 0 || maximum != 2 {
				t.Fatalf("bounds = %d..%d", minimum, maximum)
			}
			p.levels[2], p.levels[3], p.levels[40] = 2, 1, 2
			u.update.player = replacement
			u.update = &update{player: replacement}
			return 1
		},
		storeLevel: func(got *player, id int32, value uint32) {
			trace = append(trace, "clear")
			if got != p || id != 39 || value != 0 {
				t.Fatal("wrong clear")
			}
			got.levels[id], got.index = value, 255
		},
		loadIndex: func(got *player) uint8 { trace = append(trace, "index"); return got.index },
		sendPacket: func(index uint8, packet [4]byte) int32 {
			trace = append(trace, "send")
			if p.levels[39] != 0 || index != 255 || packet != ([4]byte{0xf0, 0x13, 39, 0}) {
				t.Fatal("wrong packet")
			}
			return -257
		},
	})
	want := []string{"update", "player", "class"}
	for id := 0; id < 41; id++ {
		want = append(want, fmt.Sprintf("level:%d", id))
		if id == 0 || id == 2 || id == 39 || id == 40 {
			want = append(want, fmt.Sprintf("eligible:%d", id))
		}
	}
	want = append(want, "rng")
	for id := 0; id <= 39; id++ {
		want = append(want, fmt.Sprintf("level:%d", id))
		if id == 0 || id == 3 || id == 39 {
			want = append(want, fmt.Sprintf("eligible:%d", id))
		}
	}
	want = append(want, "clear", "index", "send")
	if !reflect.DeepEqual(trace, want) || replacement.levels[39] != 1 {
		t.Fatalf("trace = %v, want %v", trace, want)
	}
}

func TestQuestLoseGuide54CEE0ExactLevelClassEmptyNoMatchAndEndpoints(t *testing.T) {
	for _, test := range []struct {
		name  string
		class uint8
		id    int32
		level uint32
		draw  int32
		send  bool
	}{
		{"warrior", 0, 7, 1, 0, false}, {"wizard", 1, 7, 1, 0, false}, {"class-255", 255, 7, 1, 0, false},
		{"empty", 2, -1, 0, -1, false}, {"zero", 2, 7, 0, -1, false}, {"one", 2, 7, 1, 0, true},
		{"two", 2, 7, 2, -1, false}, {"high-bit-one", 2, 7, 0x80000001, -1, false},
		{"all-bits", 2, 7, math.MaxUint32, -1, false}, {"first-index", 2, 0, 1, 0, true},
		{"last-index", 2, 40, 1, 0, true}, {"no-match", 2, 7, 1, 17, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			rngCalls, sends, clears, levelCalls, eligibleCalls := 0, 0, 0, 0, 0
			questLoseGuide54CEE0(1, questLoseGuideHooks54CEE0[int, int, int]{
				loadUpdate: func(int) int { return 2 }, loadPlayer: func(int) int { return 3 }, loadClass: func(int) uint8 { return test.class },
				loadLevel: func(_ int, id int32) uint32 {
					levelCalls++
					if id == test.id {
						return test.level
					}
					return 0
				},
				eligible: func(int32) int32 { eligibleCalls++; return -1 },
				randomInt: func(minimum, maximum int32) int32 {
					rngCalls++
					wantMax := int32(-1)
					if test.id >= 0 && test.level == 1 {
						wantMax = 0
					}
					if minimum != 0 || maximum != wantMax {
						t.Fatalf("bounds = %d..%d", minimum, maximum)
					}
					return test.draw
				},
				storeLevel: func(_ int, id int32, level uint32) {
					clears++
					if id != test.id || level != 0 {
						t.Fatal("wrong clear")
					}
				},
				loadIndex: func(int) uint8 { return 255 },
				sendPacket: func(index uint8, packet [4]byte) int32 {
					sends++
					if clears != 1 || index != 255 || packet != ([4]byte{0xf0, 0x13, byte(test.id), 0}) {
						t.Fatal("wrong packet")
					}
					return 0
				},
			})
			wantRNG, wantLevels, wantSend, wantEligible := 1, 82, 0, 0
			if test.id >= 0 && test.level == 1 {
				wantEligible = 2
			}
			if test.class != 2 {
				wantRNG, wantLevels, wantEligible = 0, 0, 0
			}
			if test.send {
				wantSend, wantLevels = 1, 41+int(test.id)+1
			}
			if rngCalls != wantRNG || levelCalls != wantLevels || sends != wantSend || clears != wantSend || eligibleCalls != wantEligible {
				t.Fatalf("calls = RNG:%d levels:%d eligible:%d clears:%d sends:%d", rngCalls, levelCalls, eligibleCalls, clears, sends)
			}
		})
	}
}

func TestQuestLoseGuide54CEE0NativeRNGPacketAndUntouchedFields(t *testing.T) {
	// Independent sorted candidates from the sealed GAME.EXE reward table.
	eligible := []int{2, 3, 4, 9, 10, 11, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 26, 27, 29, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40}
	for seed := 0; seed < 48; seed++ {
		s := new(Server)
		s.Rand.Logic, s.Rand.Other = prand.New(seed), prand.New(seed+11)
		otherIndex := s.Rand.Other.Index()
		p := &Player{PlayerInd: 255, GoldVal: 0xaabbccdd, Level: 9}
		p.info[66] = 2
		for i := range p.BeastScrollLvl {
			p.BeastScrollLvl[i] = 1
		}
		for i := range p.SpellLvl {
			p.SpellLvl[i] = uint32(i) | 0x80000000
		}
		u := &Object{ObjClass: 0xffffffff, UpdateData: unsafe.Pointer(&PlayerUpdateData{Player: p})}
		beforeUnit, beforeUpdate, wantPlayer := *u, *(*PlayerUpdateData)(u.UpdateData), *p
		wantRNG := prand.New(seed)
		selected := eligible[wantRNG.IntClamp(0, len(eligible)-1)]
		wantPlayer.BeastScrollLvl[selected] = 0
		packets := 0
		s.NetSendPacketXxx = func(index int, packet []byte, related *Object, remove, sequence int) int {
			packets++
			if *p != wantPlayer || index != 255 || !reflect.DeepEqual(packet, []byte{0xf0, 0x13, byte(selected), 0}) || related != nil || remove != 1 || sequence != 0 {
				t.Fatal("wrong packet/fields")
			}
			return -257
		}
		s.QuestLoseGuide54CEE0(u)
		if packets != 1 || *p != wantPlayer || *u != beforeUnit || *(*PlayerUpdateData)(u.UpdateData) != beforeUpdate {
			t.Fatalf("seed %d packets/fields = %d", seed, packets)
		}
		if s.Rand.Logic.Index() != wantRNG.Index() || s.Rand.Other.Index() != otherIndex {
			t.Fatal("wrong RNG stream or draws")
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(u)) <= math.MaxUint32 || uintptr(u.UpdateData) <= math.MaxUint32 || uintptr(unsafe.Pointer(p)) <= math.MaxUint32) {
			t.Fatal("fixture pointers must exceed 4 GiB")
		}
	}
}

func TestQuestLoseGuide54CEE0NativeRejectedLevelsAndOtherClasses(t *testing.T) {
	s := new(Server)
	p := &Player{}
	p.info[66] = 2
	u := &Object{UpdateData: unsafe.Pointer(&PlayerUpdateData{Player: p})}
	for _, level := range []uint32{0, 2, 0x80000001, math.MaxUint32} {
		for i := range p.BeastScrollLvl {
			p.BeastScrollLvl[i] = level
		}
		s.Rand.Logic = prand.New(12)
		before := *p
		s.QuestLoseGuide54CEE0(u)
		if *p != before || s.Rand.Logic.Index() != 12 {
			t.Fatal("rejected levels changed fields or advanced RNG")
		}
	}
	s.Rand.Logic = nil
	for _, class := range []uint8{0, 1, 128, 255} {
		p.info[66] = class
		before := *p
		s.QuestLoseGuide54CEE0(u)
		if *p != before {
			t.Fatal("non-Conjurer lost a guide")
		}
	}
}

func TestQuestLoseGuide54CEE0NativeMissingBindingFaults(t *testing.T) {
	for _, u := range []*Object{nil, {}, {UpdateData: unsafe.Pointer(&PlayerUpdateData{})}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("missing binding was silently ignored")
				}
			}()
			new(Server).QuestLoseGuide54CEE0(u)
		}()
	}
}

func TestQuestLoseGuide54CEE0NativeLayout(t *testing.T) {
	wantUpdate, wantPlayer, wantClass, wantIndex, wantLevels := uintptr(748), uintptr(276), uintptr(2251), uintptr(2064), uintptr(4244)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantUpdate, wantPlayer, wantClass, wantIndex, wantLevels = 872, 336, 2255, 2068, 5540
	}
	for _, check := range []struct{ got, want uintptr }{
		{unsafe.Offsetof(Object{}.UpdateData), wantUpdate}, {unsafe.Offsetof(PlayerUpdateData{}.Player), wantPlayer},
		{unsafe.Offsetof(Player{}.info) + 66, wantClass}, {unsafe.Offsetof(Player{}.PlayerInd), wantIndex},
		{unsafe.Offsetof(Player{}.BeastScrollLvl), wantLevels}, {unsafe.Sizeof(Player{}.BeastScrollLvl[0]), 4}, {uintptr(len(Player{}.BeastScrollLvl)), 41},
	} {
		if check.got != check.want {
			t.Errorf("native offset/size = %d, want %d", check.got, check.want)
		}
	}
}
