package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/prand"
)

func TestQuestLoseAbility54CFB0ExactOrderAndCachedPlayer(t *testing.T) {
	type player struct {
		levels [6]uint32
		index  uint8
	}
	type update struct{ player *player }
	type unit struct{ update *update }
	p := &player{levels: [6]uint32{7, 9, 0, 0x80000000, 0, math.MaxUint32}, index: 9}
	u := &unit{update: &update{player: p}}
	replacement := &player{levels: [6]uint32{1, 1, 1, 1, 1, 1}, index: 3}
	var trace []string
	got := questLoseAbility54CFB0(u, questLoseAbilityHooks54CFB0[*unit, *update, *player]{
		loadUpdate: func(got *unit) *update { trace = append(trace, "update"); return got.update },
		loadPlayer: func(got *update) *player { trace = append(trace, "player"); return got.player },
		loadClass:  func(got *player) uint8 { trace = append(trace, "class"); return 0 },
		loadLevel: func(got *player, id int32) uint32 {
			if got != p {
				t.Fatal("cached Player was replaced")
			}
			trace = append(trace, fmt.Sprintf("level:%d", id))
			return got.levels[id]
		},
		eligible: func(id int32) int32 {
			trace = append(trace, fmt.Sprintf("eligible:%d", id))
			return RandomAbilityLossEligible4F2570(id)
		},
		randomInt: func(minimum, maximum int32) int32 {
			trace = append(trace, "rng")
			if minimum != 1 || maximum != 3 {
				t.Fatalf("RNG bounds = %d..%d", minimum, maximum)
			}
			// The second scan must see this edit, but not the replacement binding.
			p.levels[1], p.levels[2] = 0, 5
			u.update.player = replacement
			u.update = &update{player: replacement}
			return 2
		},
		storeLevel: func(got *player, id int32, value uint32) {
			trace = append(trace, "clear")
			if got != p || id != 3 || value != 0 {
				t.Fatalf("clear = %p/%d/%d", got, id, value)
			}
			got.levels[id], got.index = value, 255
		},
		loadIndex: func(got *player) uint8 { trace = append(trace, "index"); return got.index },
		sendPacket: func(recipient uint8, packet [4]byte) int32 {
			trace = append(trace, "send")
			if p.levels[3] != 0 || recipient != 255 || packet != ([4]byte{0xf0, 0x12, 3, 0}) {
				t.Fatalf("packet = %d/%x", recipient, packet)
			}
			return 0x123456e1
		},
	})
	want := []string{"update", "player", "class", "level:0", "eligible:0", "level:1", "eligible:1", "level:2", "level:3", "eligible:3", "level:4", "level:5", "eligible:5", "rng", "level:1", "level:2", "eligible:2", "level:3", "eligible:3", "clear", "index", "send"}
	if got != -31 || !reflect.DeepEqual(trace, want) || replacement.levels != ([6]uint32{1, 1, 1, 1, 1, 1}) {
		t.Fatalf("result/trace = %d/%v, want -31/%v", got, trace, want)
	}
}

func TestQuestLoseAbility54CFB0OriginalLowByteFallbacks(t *testing.T) {
	for _, test := range []struct {
		name           string
		class          uint8
		levels         [6]uint32
		draw, eligible int32
		want           int8
	}{
		{"wizard", 1, [6]uint32{}, 0, 0, 1},
		{"conjurer", 2, [6]uint32{}, 0, 0, 2},
		{"class-low-byte", 255, [6]uint32{}, 0, 0, -1},
		{"empty-rng-low-byte", 0, [6]uint32{}, 0x12345680, 0, -128},
		{"invalid-ordinal-last-eligibility", 0, [6]uint32{0, 1, 0, 0, 0, 1}, -1, -257, -1},
		{"rejected-eligibility", 0, [6]uint32{0, 1}, 0x1234567f, 0, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			rngCalls, levelCalls := 0, 0
			got := questLoseAbility54CFB0(1, questLoseAbilityHooks54CFB0[int, int, int]{
				loadUpdate: func(int) int { return 2 }, loadPlayer: func(int) int { return 3 },
				loadClass: func(int) uint8 { return test.class },
				loadLevel: func(_ int, id int32) uint32 { levelCalls++; return test.levels[id] },
				eligible:  func(int32) int32 { return test.eligible },
				randomInt: func(minimum, maximum int32) int32 {
					rngCalls++
					var count int32
					for _, level := range test.levels {
						if level != 0 && test.eligible != 0 {
							count++
						}
					}
					if minimum != 1 || maximum != count {
						t.Fatalf("bounds = %d..%d, want 1..%d", minimum, maximum, count)
					}
					return test.draw
				},
			})
			wantCalls := 1
			if test.class != 0 {
				wantCalls = 0
				if levelCalls != 0 {
					t.Fatal("non-Warrior scanned abilities")
				}
			}
			if got != test.want || rngCalls != wantCalls {
				t.Fatalf("result/RNG calls = %d/%d", got, rngCalls)
			}
		})
	}
}

func TestQuestLoseAbility54CFB0NativeRNGPacketAndUntouchedFields(t *testing.T) {
	for seed := 0; seed < 48; seed++ {
		s := new(Server)
		s.Rand.Logic = prand.New(seed)
		s.Rand.Other = prand.New(seed + 11)
		otherIndex := s.Rand.Other.Index()
		p := &Player{PlayerInd: 255, GoldVal: 0xaabbccdd, Level: 9}
		for i := range p.SpellLvl {
			p.SpellLvl[i] = uint32(i) | 0x80000000
		}
		u := &Object{ObjClass: 0xffffffff, UpdateData: unsafe.Pointer(&PlayerUpdateData{Player: p})}
		beforeUnit, beforeUpdate, wantPlayer := *u, *(*PlayerUpdateData)(u.UpdateData), *p
		wantRNG := prand.New(seed)
		selected := wantRNG.IntClamp(1, 5)
		wantPlayer.SpellLvl[selected] = 0
		packets := 0
		s.NetSendPacketXxx = func(recipient int, packet []byte, related *Object, remove, sequence int) int {
			packets++
			if *p != wantPlayer || recipient != 255 || !reflect.DeepEqual(packet, []byte{0xf0, 0x12, byte(selected), 0}) || related != nil || remove != 1 || sequence != 0 {
				t.Fatalf("packet/fields = %d/%x/%p/%d/%d", recipient, packet, related, remove, sequence)
			}
			return 0x123456fe
		}
		if got := s.QuestLoseAbility54CFB0(u); got != -2 || packets != 1 || *u != beforeUnit || *(*PlayerUpdateData)(u.UpdateData) != beforeUpdate {
			t.Fatalf("result/packets = %d/%d", got, packets)
		}
		if s.Rand.Logic.Index() != wantRNG.Index() || s.Rand.Other.Index() != otherIndex {
			t.Fatal("wrong RNG stream or number of draws")
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(u)) <= math.MaxUint32 || uintptr(u.UpdateData) <= math.MaxUint32 || uintptr(unsafe.Pointer(p)) <= math.MaxUint32) {
			t.Fatal("fixture pointers must exceed 4 GiB")
		}
	}
}

func TestQuestLoseAbility54CFB0NativeEmptyAndNonWarrior(t *testing.T) {
	s := new(Server)
	s.Rand.Logic = prand.New(12)
	p := &Player{}
	p.SpellLvl[0], p.SpellLvl[6] = math.MaxUint32, math.MaxUint32
	u := &Object{UpdateData: unsafe.Pointer(&PlayerUpdateData{Player: p})}
	before := *p
	if got := s.QuestLoseAbility54CFB0(u); got != 0 || *p != before || s.Rand.Logic.Index() != 12 {
		t.Fatalf("empty result/index = %d/%d", got, s.Rand.Logic.Index())
	}
	s.Rand.Logic = nil
	for _, class := range []uint8{1, 2, 128, 255} {
		p.info[66] = class
		if got := s.QuestLoseAbility54CFB0(u); got != int8(class) {
			t.Fatalf("class %d result = %d", class, got)
		}
	}
}

func TestQuestLoseAbility54CFB0NativeMissingBindingFaults(t *testing.T) {
	for _, u := range []*Object{nil, {}, {UpdateData: unsafe.Pointer(&PlayerUpdateData{})}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("missing binding was silently ignored")
				}
			}()
			new(Server).QuestLoseAbility54CFB0(u)
		}()
	}
}

func TestQuestLoseAbility54CFB0NativeLayout(t *testing.T) {
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
