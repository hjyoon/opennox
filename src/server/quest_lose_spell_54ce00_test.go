package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/prand"
)

func TestQuestLoseSpell54CE00ExactOrderAndCachedPlayer(t *testing.T) {
	type player struct {
		levels [137]uint32
		index  uint8
	}
	type update struct{ player *player }
	type unit struct{ update *update }
	p := &player{index: 9}
	for _, id := range []int{0, 1, 27, 128, 136} {
		p.levels[id] = uint32(id) | 0x80000000
	}
	u := &unit{update: &update{player: p}}
	replacement := &player{index: 3}
	replacement.levels[128] = 111
	var trace []string
	questLoseSpell54CE00(u, questLoseSpellHooks54CE00[*unit, *update, *player]{
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
			return RandomSpellLossEligible4F24E0(id)
		},
		randomInt: func(minimum, maximum int32) int32 {
			trace = append(trace, "rng")
			if minimum != 0 || maximum != 2 {
				t.Fatalf("RNG bounds = %d..%d", minimum, maximum)
			}
			p.levels[1], p.levels[4] = 0, 9
			u.update.player = replacement
			u.update = &update{player: replacement}
			return 1
		},
		storeLevel: func(got *player, id int32, value uint32) {
			trace = append(trace, "clear")
			if got != p || id != 128 || value != 0 {
				t.Fatalf("clear = %p/%d/%d", got, id, value)
			}
			got.levels[id], got.index = value, 255
		},
		loadIndex: func(got *player) uint8 { trace = append(trace, "index"); return got.index },
		sendPacket: func(recipient uint8, packet [4]byte) int32 {
			trace = append(trace, "send")
			if p.levels[128] != 0 || recipient != 255 || packet != ([4]byte{0xf0, 0x11, 128, 0}) {
				t.Fatalf("packet = %d/%x", recipient, packet)
			}
			return -257 // The original void helper ignores this result.
		},
	})
	want := []string{"update", "player", "class"}
	for id := 0; id < 137; id++ {
		want = append(want, fmt.Sprintf("level:%d", id))
		if id == 0 || id == 1 || id == 27 || id == 128 || id == 136 {
			want = append(want, fmt.Sprintf("eligible:%d", id))
		}
	}
	want = append(want, "rng")
	for id := 0; id <= 128; id++ {
		want = append(want, fmt.Sprintf("level:%d", id))
		if id == 0 || id == 4 || id == 27 || id == 128 {
			want = append(want, fmt.Sprintf("eligible:%d", id))
		}
	}
	want = append(want, "clear", "index", "send")
	if !reflect.DeepEqual(trace, want) || replacement.levels[128] != 111 {
		t.Fatalf("trace = %v, want %v", trace, want)
	}
}

func TestQuestLoseSpell54CE00ClassEmptyNoMatchAndScanEndpoints(t *testing.T) {
	for _, test := range []struct {
		name  string
		class uint8
		id    int32
		draw  int32
		send  bool
	}{
		{"warrior", 0, -1, 0, false}, {"class-128", 128, -1, 0, false},
		{"class-255", 255, -1, 0, false}, {"empty-wizard", 1, -1, -1, false},
		{"empty-conjurer", 2, -1, -1, false}, {"first-index", 1, 0, 0, true},
		{"last-index", 2, 136, 0, true}, {"no-match", 1, 5, 17, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			rngCalls, sends, clears, levelCalls := 0, 0, 0, 0
			questLoseSpell54CE00(1, questLoseSpellHooks54CE00[int, int, int]{
				loadUpdate: func(int) int { return 2 }, loadPlayer: func(int) int { return 3 },
				loadClass: func(int) uint8 { return test.class },
				loadLevel: func(_ int, id int32) uint32 {
					levelCalls++
					if id == test.id {
						return math.MaxUint32
					}
					return 0
				},
				eligible: func(int32) int32 { return -1 },
				randomInt: func(minimum, maximum int32) int32 {
					rngCalls++
					wantMax := int32(-1)
					if test.id >= 0 {
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
					if clears != 1 || index != 255 || packet != ([4]byte{0xf0, 0x11, byte(test.id), 0}) {
						t.Fatal("wrong packet")
					}
					return 0
				},
			})
			wantRNG, wantLevels, wantSend := 1, 274, 0
			if test.class != 1 && test.class != 2 {
				wantRNG, wantLevels = 0, 0
			}
			if test.send {
				wantSend, wantLevels = 1, 137+int(test.id)+1
			}
			if rngCalls != wantRNG || levelCalls != wantLevels || sends != wantSend || clears != wantSend {
				t.Fatalf("calls = RNG:%d levels:%d clears:%d sends:%d", rngCalls, levelCalls, clears, sends)
			}
		})
	}
}

func TestQuestLoseSpell54CE00NativeRNGPacketAndUntouchedFields(t *testing.T) {
	// Independent sorted candidates from the sealed GAME.EXE reward table.
	eligible := []int{1, 4, 5, 8, 10, 12, 13, 14, 16, 21, 22, 23, 24, 26, 29, 35, 36, 37, 38, 39, 42, 43, 50, 51, 52, 54, 58, 60, 61, 62, 64, 67, 71, 72, 74, 128, 129, 130, 132, 134, 135, 136}
	for _, class := range []uint8{1, 2} {
		for seed := 0; seed < 48; seed++ {
			s := new(Server)
			s.Rand.Logic, s.Rand.Other = prand.New(seed), prand.New(seed+11)
			otherIndex := s.Rand.Other.Index()
			p := &Player{PlayerInd: 255, GoldVal: 0xaabbccdd, Level: 9}
			p.info[66] = class
			for i := range p.SpellLvl {
				p.SpellLvl[i] = uint32(i) | 0x80000000
			}
			u := &Object{ObjClass: 0xffffffff, UpdateData: unsafe.Pointer(&PlayerUpdateData{Player: p})}
			beforeUnit, beforeUpdate, wantPlayer := *u, *(*PlayerUpdateData)(u.UpdateData), *p
			wantRNG := prand.New(seed)
			selected := eligible[wantRNG.IntClamp(0, len(eligible)-1)]
			wantPlayer.SpellLvl[selected] = 0
			packets := 0
			s.NetSendPacketXxx = func(recipient int, packet []byte, related *Object, remove, sequence int) int {
				packets++
				if *p != wantPlayer || recipient != 255 || !reflect.DeepEqual(packet, []byte{0xf0, 0x11, byte(selected), 0}) || related != nil || remove != 1 || sequence != 0 {
					t.Fatalf("packet/fields = %d/%x/%p/%d/%d", recipient, packet, related, remove, sequence)
				}
				return -257
			}
			s.QuestLoseSpell54CE00(u)
			if packets != 1 || *u != beforeUnit || *(*PlayerUpdateData)(u.UpdateData) != beforeUpdate || *p != wantPlayer {
				t.Fatalf("class %d seed %d packets/fields = %d", class, seed, packets)
			}
			if s.Rand.Logic.Index() != wantRNG.Index() || s.Rand.Other.Index() != otherIndex {
				t.Fatal("wrong RNG stream or number of draws")
			}
			if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(u)) <= math.MaxUint32 || uintptr(u.UpdateData) <= math.MaxUint32 || uintptr(unsafe.Pointer(p)) <= math.MaxUint32) {
				t.Fatal("fixture pointers must exceed 4 GiB")
			}
		}
	}
}

func TestQuestLoseSpell54CE00NativeProtectedEmptyAndOtherClasses(t *testing.T) {
	s := new(Server)
	s.Rand.Logic = prand.New(12)
	p := &Player{}
	p.info[66] = 1
	for _, id := range []int{0, 9, 27, 34, 41} {
		p.SpellLvl[id] = math.MaxUint32
	}
	u := &Object{UpdateData: unsafe.Pointer(&PlayerUpdateData{Player: p})}
	before := *p
	s.QuestLoseSpell54CE00(u)
	if *p != before || s.Rand.Logic.Index() != 12 {
		t.Fatal("empty candidates changed fields or advanced RNG")
	}
	s.Rand.Logic = nil
	for _, class := range []uint8{0, 128, 255} {
		p.info[66] = class
		before = *p
		s.QuestLoseSpell54CE00(u)
		if *p != before {
			t.Fatal("non-Wizard/Conjurer lost a spell")
		}
	}
}

func TestQuestLoseSpell54CE00NativeMissingBindingFaults(t *testing.T) {
	for _, u := range []*Object{nil, {}, {UpdateData: unsafe.Pointer(&PlayerUpdateData{})}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("missing binding was silently ignored")
				}
			}()
			new(Server).QuestLoseSpell54CE00(u)
		}()
	}
}

func TestQuestLoseSpell54CE00NativeLayout(t *testing.T) {
	wantUpdate, wantPlayer, wantClass, wantIndex, wantLevels := uintptr(748), uintptr(276), uintptr(2251), uintptr(2064), uintptr(3696)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantUpdate, wantPlayer, wantClass, wantIndex, wantLevels = 872, 336, 2255, 2068, 4992
	}
	for _, check := range []struct{ got, want uintptr }{
		{unsafe.Offsetof(Object{}.UpdateData), wantUpdate}, {unsafe.Offsetof(PlayerUpdateData{}.Player), wantPlayer},
		{unsafe.Offsetof(Player{}.info) + 66, wantClass}, {unsafe.Offsetof(Player{}.PlayerInd), wantIndex},
		{unsafe.Offsetof(Player{}.SpellLvl), wantLevels}, {unsafe.Sizeof(Player{}.SpellLvl[0]), 4},
		{uintptr(len(Player{}.SpellLvl)), 137},
	} {
		if check.got != check.want {
			t.Errorf("native offset/size = %d, want %d", check.got, check.want)
		}
	}
}
