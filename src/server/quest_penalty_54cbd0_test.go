package server

import (
	"fmt"
	"reflect"
	"testing"
)

type questPenaltyTestPlayer54CBD0 struct{ class uint8 }
type questPenaltyTestUpdate54CBD0 struct{ player *questPenaltyTestPlayer54CBD0 }
type questPenaltyTestUnit54CBD0 struct {
	update *questPenaltyTestUpdate54CBD0
	gold   uint32
}
type questPenaltyTestState54CBD0 struct {
	unit   *questPenaltyTestUnit54CBD0
	trace  []string
	amount uint32
	result int8
}

func questPenaltyTestFixture54CBD0(class uint8, gold uint32) *questPenaltyTestState54CBD0 {
	return &questPenaltyTestState54CBD0{unit: &questPenaltyTestUnit54CBD0{
		update: &questPenaltyTestUpdate54CBD0{player: &questPenaltyTestPlayer54CBD0{class: class}}, gold: gold,
	}}
}

func (s *questPenaltyTestState54CBD0) hooks(t *testing.T) questPenaltyHooks54CBD0[*questPenaltyTestUnit54CBD0, *questPenaltyTestUpdate54CBD0, *questPenaltyTestPlayer54CBD0] {
	t.Helper()
	checkUnit := func(unit *questPenaltyTestUnit54CBD0, event string) {
		t.Helper()
		if unit != s.unit {
			t.Fatalf("%s received a different unit", event)
		}
		s.trace = append(s.trace, event)
	}
	return questPenaltyHooks54CBD0[*questPenaltyTestUnit54CBD0, *questPenaltyTestUpdate54CBD0, *questPenaltyTestPlayer54CBD0]{
		loadUpdate: func(unit *questPenaltyTestUnit54CBD0) *questPenaltyTestUpdate54CBD0 {
			checkUnit(unit, "update")
			return unit.update
		},
		getGold: func(unit *questPenaltyTestUnit54CBD0) uint32 { checkUnit(unit, "gold"); return unit.gold },
		subGold: func(unit *questPenaltyTestUnit54CBD0, amount uint32) {
			checkUnit(unit, "sub")
			s.amount = amount
		},
		loseGems:   func(unit *questPenaltyTestUnit54CBD0) { checkUnit(unit, "gems") },
		loseWeapon: func(unit *questPenaltyTestUnit54CBD0) { checkUnit(unit, "weapon") },
		loseArmor:  func(unit *questPenaltyTestUnit54CBD0) { checkUnit(unit, "armor") },
		loadPlayer: func(update *questPenaltyTestUpdate54CBD0) *questPenaltyTestPlayer54CBD0 {
			s.trace = append(s.trace, "player")
			return update.player
		},
		loadClass: func(player *questPenaltyTestPlayer54CBD0) uint8 {
			s.trace = append(s.trace, "class")
			return player.class
		},
		loseSpell: func(unit *questPenaltyTestUnit54CBD0) { checkUnit(unit, "spell") },
		loseGuide: func(unit *questPenaltyTestUnit54CBD0) { checkUnit(unit, "guide") },
		loseAbility: func(unit *questPenaltyTestUnit54CBD0) int8 {
			checkUnit(unit, "ability")
			return s.result
		},
	}
}

func questPenaltyTestTrace54CBD0(class uint8) []string {
	trace := []string{"update", "gold", "sub", "gems", "weapon", "armor", "player", "class"}
	if class == 0 {
		trace = append(trace, "armor")
	}
	return append(trace, "spell", "spell", "guide", "guide", "ability")
}

func TestQuestPenalty54CBD0ExactOrderAllClassBytesAndUnsignedGold(t *testing.T) {
	for class := 0; class < 256; class++ {
		for _, gold := range []uint32{0, 1, 2, 3, 0x7fffffff, 0x80000000, 0xfffffffe, 0xffffffff} {
			t.Run(fmt.Sprintf("class-%d-gold-%08x", class, gold), func(t *testing.T) {
				s := questPenaltyTestFixture54CBD0(uint8(class), gold)
				questPenalty54CBD0(s.unit, s.hooks(t))
				if s.amount != gold>>1 || !reflect.DeepEqual(s.trace, questPenaltyTestTrace54CBD0(uint8(class))) {
					t.Fatalf("amount=%08x trace=%v", s.amount, s.trace)
				}
			})
		}
	}
}

func TestQuestPenalty54CBD0CachedUpdateLateLivePlayerAndSingleClassRead(t *testing.T) {
	s := questPenaltyTestFixture54CBD0(2, 0xffffffff)
	cached := s.unit.update
	replacement := &questPenaltyTestUpdate54CBD0{player: &questPenaltyTestPlayer54CBD0{class: 255}}
	h := s.hooks(t)
	getGold, subGold, loseWeapon, loseArmor := h.getGold, h.subGold, h.loseWeapon, h.loseArmor
	h.getGold = func(unit *questPenaltyTestUnit54CBD0) uint32 {
		gold := getGold(unit)
		unit.update = replacement
		return gold
	}
	h.subGold = func(unit *questPenaltyTestUnit54CBD0, amount uint32) {
		subGold(unit, amount)
		unit.gold = 37 // No later read or recalculation of the subtraction amount.
	}
	h.loseWeapon = func(unit *questPenaltyTestUnit54CBD0) {
		loseWeapon(unit)
		cached.player = nil // Not read until after the first armor callback.
	}
	armors := 0
	h.loseArmor = func(unit *questPenaltyTestUnit54CBD0) {
		loseArmor(unit)
		armors++
		if armors == 1 {
			cached.player = &questPenaltyTestPlayer54CBD0{class: 0}
		} else {
			cached.player = nil // The dispatcher must not inspect class again.
		}
	}
	questPenalty54CBD0(s.unit, h)
	if armors != 2 || s.amount != 0x7fffffff || !reflect.DeepEqual(s.trace, questPenaltyTestTrace54CBD0(0)) {
		t.Fatalf("armor=%d amount=%x trace=%v", armors, s.amount, s.trace)
	}
}

func TestQuestPenalty54CBD0AbilityReturnIgnored(t *testing.T) {
	for result := -128; result < 128; result++ {
		s := questPenaltyTestFixture54CBD0(1, 0)
		s.result = int8(result)
		questPenalty54CBD0(s.unit, s.hooks(t))
		if !reflect.DeepEqual(s.trace, questPenaltyTestTrace54CBD0(1)) {
			t.Fatalf("result %d trace=%v", result, s.trace)
		}
	}
}

func TestQuestPenalty54CBD0MissingBindingsFaultAtExactPrefix(t *testing.T) {
	for _, missing := range []string{"update", "gold", "sub", "gems", "weapon", "armor", "player", "class", "spell", "guide", "ability", "nil-update", "nil-player"} {
		t.Run(missing, func(t *testing.T) {
			s := questPenaltyTestFixture54CBD0(1, 0)
			h := s.hooks(t)
			faultAt := missing
			switch missing {
			case "update":
				h.loadUpdate = nil
			case "gold":
				h.getGold = nil
			case "sub":
				h.subGold = nil
			case "gems":
				h.loseGems = nil
			case "weapon":
				h.loseWeapon = nil
			case "armor":
				h.loseArmor = nil
			case "player":
				h.loadPlayer = nil
			case "class":
				h.loadClass = nil
			case "spell":
				h.loseSpell = nil
			case "guide":
				h.loseGuide = nil
			case "ability":
				h.loseAbility = nil
			case "nil-update":
				s.unit.update = nil
				faultAt = "class"
			case "nil-player":
				s.unit.update.player = nil
				faultAt = "spell"
			}
			want := questPenaltyTestTrace54CBD0(1)
			for i, event := range want {
				if event == faultAt {
					want = want[:i]
					break
				}
			}
			defer func() {
				if recover() == nil || len(s.trace) != len(want) || len(want) != 0 && !reflect.DeepEqual(s.trace, want) {
					t.Fatalf("missing %s trace=%v, want fault after %v", missing, s.trace, want)
				}
			}()
			questPenalty54CBD0(s.unit, h)
		})
	}
}
