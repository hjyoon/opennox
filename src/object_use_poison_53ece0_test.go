package opennox

import (
	"reflect"
	"testing"
)

type poisonUseTestObject53ECE0 struct {
	name     string
	poison   uint8
	confused bool
	deleted  bool
}

func TestUseMushroom53ECE0PoisonedAndCleanCases(t *testing.T) {
	for _, test := range []struct {
		name       string
		poison     uint8
		wantEvents []string
	}{
		{
			name:   "poisoned owner is cured before confusion",
			poison: 4,
			wantEvents: []string{
				"load:owner", "remove:owner", "message:owner:Use.c:MushroomClean",
				"audio:owner", "confuse:owner", "delete:mushroom",
			},
		},
		{
			name:   "clean owner is still confused and mushroom is consumed",
			poison: 0,
			wantEvents: []string{
				"load:owner", "message:owner:Use.c:MushroomConfuse",
				"confuse:owner", "delete:mushroom",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			owner := &poisonUseTestObject53ECE0{name: "owner", poison: test.poison}
			item := &poisonUseTestObject53ECE0{name: "mushroom"}
			var events []string
			hooks := mushroomUseHooks53ECE0[*poisonUseTestObject53ECE0, *poisonUseTestObject53ECE0]{
				loadPoison: func(obj *poisonUseTestObject53ECE0) uint8 {
					events = append(events, "load:"+obj.name)
					return obj.poison
				},
				removePoison: func(obj *poisonUseTestObject53ECE0) {
					events = append(events, "remove:"+obj.name)
					obj.poison = 0
				},
				priorityMessage: func(obj *poisonUseTestObject53ECE0, message string) {
					events = append(events, "message:"+obj.name+":"+message)
				},
				cureAudio: func(obj *poisonUseTestObject53ECE0) {
					events = append(events, "audio:"+obj.name)
				},
				applyConfusion: func(obj *poisonUseTestObject53ECE0) {
					events = append(events, "confuse:"+obj.name)
					obj.confused = true
				},
				delayedDelete: func(obj *poisonUseTestObject53ECE0) {
					events = append(events, "delete:"+obj.name)
					obj.deleted = true
				},
			}
			if !useMushroom53ECE0(owner, item, hooks) {
				t.Fatal("mushroom use returned false")
			}
			if owner.poison != 0 || !owner.confused || !item.deleted {
				t.Fatalf("state = poison:%d confused:%t deleted:%t, want 0/true/true",
					owner.poison, owner.confused, item.deleted)
			}
			if !reflect.DeepEqual(events, test.wantEvents) {
				t.Fatalf("events = %q, want %q", events, test.wantEvents)
			}
		})
	}
}

func TestTryUseCurePoisonPotion53EF70GatesAndEffects(t *testing.T) {
	for _, test := range []struct {
		name       string
		isAntidote bool
		isPlayer   bool
		poison     uint8
		wantUsed   bool
		wantEvents []string
	}{
		{name: "different potion", isPlayer: true, poison: 3},
		{name: "non-player", isAntidote: true, poison: 3},
		{name: "unpoisoned player keeps antidote", isAntidote: true, isPlayer: true, wantEvents: []string{"load:owner"}},
		{
			name:       "poisoned player consumes antidote",
			isAntidote: true,
			isPlayer:   true,
			poison:     3,
			wantUsed:   true,
			wantEvents: []string{"load:owner", "remove:owner", "audio:owner"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			owner := &poisonUseTestObject53ECE0{name: "owner", poison: test.poison}
			var events []string
			hooks := curePoisonPotionHooks53EF70[*poisonUseTestObject53ECE0]{
				loadPoison: func(obj *poisonUseTestObject53ECE0) uint8 {
					events = append(events, "load:"+obj.name)
					return obj.poison
				},
				removePoison: func(obj *poisonUseTestObject53ECE0) {
					events = append(events, "remove:"+obj.name)
					obj.poison = 0
				},
				cureAudio: func(obj *poisonUseTestObject53ECE0) {
					events = append(events, "audio:"+obj.name)
				},
			}
			if got := tryUseCurePoisonPotion53EF70(owner, test.isAntidote, test.isPlayer, hooks); got != test.wantUsed {
				t.Fatalf("used = %t, want %t", got, test.wantUsed)
			}
			wantPoison := test.poison
			if test.wantUsed {
				wantPoison = 0
			}
			if owner.poison != wantPoison {
				t.Fatalf("poison = %d, want %d", owner.poison, wantPoison)
			}
			if !reflect.DeepEqual(events, test.wantEvents) {
				t.Fatalf("events = %q, want %q", events, test.wantEvents)
			}
		})
	}
}
