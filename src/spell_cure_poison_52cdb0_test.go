package opennox

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/opennox/libs/spell"
)

type curePoisonTestObject52CDB0 struct {
	name   string
	poison uint8
}

func TestCurePoison52CDB0ClearsOrReducesPoison(t *testing.T) {
	for _, tc := range []struct {
		name       string
		poison     uint8
		level      int32
		wantPoison uint8
		wantEvents []string
	}{
		{
			name:       "weaker cast reduces poison",
			poison:     5,
			level:      2,
			wantPoison: 3,
			wantEvents: []string{"load:target=5", "update:target:2", "message:target:ExecSpel.c:PoisonCure", "audio:target:SPELL_CURE_POISON"},
		},
		{
			name:       "equal cast clears poison",
			poison:     3,
			level:      3,
			wantPoison: 0,
			wantEvents: []string{"load:target=3", "remove:target", "message:target:ExecSpel.c:PoisonClean", "audio:target:SPELL_CURE_POISON"},
		},
		{
			name:       "stronger cast clears poison",
			poison:     1,
			level:      5,
			wantPoison: 0,
			wantEvents: []string{"load:target=1", "remove:target", "message:target:ExecSpel.c:PoisonClean", "audio:target:SPELL_CURE_POISON"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			caster := &curePoisonTestObject52CDB0{name: "caster"}
			target := &curePoisonTestObject52CDB0{name: "target", poison: tc.poison}
			var events []string
			hooks := curePoisonHooks52CDB0[*curePoisonTestObject52CDB0]{
				loadPoison: func(obj *curePoisonTestObject52CDB0) uint8 {
					events = append(events, fmt.Sprintf("load:%s=%d", obj.name, obj.poison))
					return obj.poison
				},
				update: func(obj *curePoisonTestObject52CDB0, amount int32) {
					events = append(events, fmt.Sprintf("update:%s:%d", obj.name, amount))
					obj.poison -= uint8(amount)
				},
				remove: func(obj *curePoisonTestObject52CDB0) {
					events = append(events, "remove:"+obj.name)
					obj.poison = 0
				},
				message: func(obj *curePoisonTestObject52CDB0, message string) {
					events = append(events, "message:"+obj.name+":"+message)
				},
				audio: func(id spell.ID, obj *curePoisonTestObject52CDB0) {
					events = append(events, "audio:"+obj.name+":"+id.String())
				},
				manaCost: func(spell.ID, int) int {
					t.Fatal("poisoned target requested a mana refund")
					return 0
				},
				refundMana: func(*curePoisonTestObject52CDB0, int) {
					t.Fatal("poisoned target refunded mana")
				},
			}
			if got := curePoison52CDB0(spell.SPELL_CURE_POISON, caster, target, tc.level, hooks); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			if target.poison != tc.wantPoison {
				t.Fatalf("poison = %d, want %d", target.poison, tc.wantPoison)
			}
			if !reflect.DeepEqual(events, tc.wantEvents) {
				t.Fatalf("events = %v, want %v", events, tc.wantEvents)
			}
		})
	}
}

func TestCurePoison52CDB0UnpoisonedTargetSemantics(t *testing.T) {
	caster := &curePoisonTestObject52CDB0{name: "caster"}
	other := &curePoisonTestObject52CDB0{name: "other"}
	var events []string
	hooks := curePoisonHooks52CDB0[*curePoisonTestObject52CDB0]{
		loadPoison: func(obj *curePoisonTestObject52CDB0) uint8 {
			events = append(events, "load:"+obj.name)
			return obj.poison
		},
		update:  func(*curePoisonTestObject52CDB0, int32) { t.Fatal("unexpected update") },
		remove:  func(*curePoisonTestObject52CDB0) { t.Fatal("unexpected remove") },
		message: func(*curePoisonTestObject52CDB0, string) { t.Fatal("unexpected message") },
		audio: func(_ spell.ID, obj *curePoisonTestObject52CDB0) {
			events = append(events, "audio:"+obj.name)
		},
		manaCost: func(_ spell.ID, level int) int {
			events = append(events, fmt.Sprintf("cost:%d", level))
			return 17
		},
		refundMana: func(obj *curePoisonTestObject52CDB0, amount int) {
			events = append(events, fmt.Sprintf("refund:%s:%d", obj.name, amount))
		},
	}

	if got := curePoison52CDB0(spell.SPELL_CURE_POISON, caster, other, 3, hooks); got != 1 {
		t.Fatalf("other-target result = %d, want 1", got)
	}
	if got := curePoison52CDB0(spell.SPELL_CURE_POISON, caster, caster, 3, hooks); got != 1 {
		t.Fatalf("self-target result = %d, want 1", got)
	}
	if got := curePoison52CDB0[*curePoisonTestObject52CDB0](spell.SPELL_CURE_POISON, caster, nil, 3, hooks); got != 0 {
		t.Fatalf("nil-target result = %d, want 0", got)
	}
	want := []string{"load:other", "audio:other", "load:caster", "cost:1", "refund:caster:17"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}
