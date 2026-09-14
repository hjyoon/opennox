package server

import (
	"math"
	"reflect"
	"testing"
)

type channelLifeTestObject52F460 struct {
	flags, class  uint32
	mana, maximum uint16
	hp            uint16
	buffed, moved bool
}

type channelLifeTestRecord52F460 struct {
	target, caster *channelLifeTestObject52F460
	mode, level    uint32
	fraction       float32
}

func channelLifeTestHooks52F460(events *[]string) channelLifeHooks52F460[*channelLifeTestRecord52F460, *channelLifeTestObject52F460] {
	log := func(s string) { *events = append(*events, s) }
	return channelLifeHooks52F460[*channelLifeTestRecord52F460, *channelLifeTestObject52F460]{
		loadTarget: func(r *channelLifeTestRecord52F460) *channelLifeTestObject52F460 {
			log("target")
			return r.target
		},
		loadFlags: func(o *channelLifeTestObject52F460) uint32 { log("flags"); return o.flags },
		loadMode:  func(r *channelLifeTestRecord52F460) uint32 { log("mode"); return r.mode },
		addMana: func(o *channelLifeTestObject52F460, n int16) {
			log("add")
			o.mana += uint16(n)
		},
		clearDamage: func(*channelLifeTestObject52F460, int32) { log("clear") },
		loadCaster: func(r *channelLifeTestRecord52F460) *channelLifeTestObject52F460 {
			log("caster")
			return r.caster
		},
		testBuff: func(o *channelLifeTestObject52F460, buff int32) int32 {
			log("buff")
			if buff != 8 {
				panic("wrong buff")
			}
			if o.buffed {
				return 1
			}
			return 0
		},
		loadClass: func(o *channelLifeTestObject52F460) uint8 { log("class"); return uint8(o.class) },
		positionDelta: func(o *channelLifeTestObject52F460, _ *channelLifeTestRecord52F460) int32 {
			log("position")
			if o.moved {
				return 1
			}
			return 0
		},
		maxMana:     func(o *channelLifeTestObject52F460) uint16 { log("max"); return o.maximum },
		currentMana: func(o *channelLifeTestObject52F460) uint16 { log("mana"); return o.mana },
		getHP: func(o *channelLifeTestObject52F460) uint16 {
			log("hp")
			if o == nil {
				return 0
			}
			return o.hp
		},
		loadFraction: func(r *channelLifeTestRecord52F460) float32 { log("fraction"); return r.fraction },
		loadLevel:    func(r *channelLifeTestRecord52F460) uint32 { log("level"); return r.level },
		coefficient: func(i uint32) float64 {
			log("coefficient")
			if i != 1 {
				panic("wrong level index")
			}
			return 0.5
		},
		storeFraction: func(r *channelLifeTestRecord52F460, v float32) {
			log("store")
			r.fraction = v
		},
	}
}

func TestChannelLife52F460FullPathAndAccumulation(t *testing.T) {
	var events []string
	target := &channelLifeTestObject52F460{class: 4, mana: 4, maximum: 10}
	caster := &channelLifeTestObject52F460{class: 4, hp: 3}
	record := &channelLifeTestRecord52F460{target: target, caster: caster, level: 2, fraction: 0.25}
	if got := spellChannelLifeUpdate52F460(record, channelLifeTestHooks52F460(&events)); got != 0 {
		t.Fatalf("update = %d, want 0", got)
	}
	if target.mana != 5 || record.fraction != -0.25 {
		t.Fatalf("mana = %d, fraction = %g; want 5, -0.25", target.mana, record.fraction)
	}
	want := []string{"target", "flags", "mode", "caster", "buff", "target", "class", "target", "max", "target", "mana", "caster", "hp", "caster", "hp", "fraction", "level", "coefficient", "store", "target", "add", "caster", "clear"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestChannelLife52F460TerminationBranches(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*channelLifeTestRecord52F460)
		want   int32
	}{
		{"no target", func(r *channelLifeTestRecord52F460) { r.target = nil }, 1},
		{"dead target", func(r *channelLifeTestRecord52F460) { r.target.flags = 0x8000 }, 1},
		{"destroyed target", func(r *channelLifeTestRecord52F460) { r.target.flags = 0x20 }, 1},
		{"buffed caster", func(r *channelLifeTestRecord52F460) { r.caster.buffed = true }, 1},
		{"moved monster", func(r *channelLifeTestRecord52F460) { r.target.class = 2; r.target.moved = true }, 1},
		{"full mana", func(r *channelLifeTestRecord52F460) { r.target.mana = r.target.maximum }, 1},
		{"low HP", func(r *channelLifeTestRecord52F460) { r.caster.hp = 1 }, 1},
		{"no caster", func(r *channelLifeTestRecord52F460) { r.caster = nil }, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			r := &channelLifeTestRecord52F460{
				target: &channelLifeTestObject52F460{class: 4, mana: 2, maximum: 10},
				caster: &channelLifeTestObject52F460{class: 4, hp: 3}, level: 2,
			}
			tc.mutate(r)
			if got := spellChannelLifeUpdate52F460(r, channelLifeTestHooks52F460(&events)); got != tc.want {
				t.Fatalf("update = %d, want %d", got, tc.want)
			}
			for _, event := range events {
				if event == "add" || event == "clear" || event == "store" {
					t.Fatalf("unexpected effect %q in %v", event, events)
				}
			}
		})
	}
}

func TestChannelLife52F460ModeReloadsTargetAfterMana(t *testing.T) {
	var events []string
	first, second := &channelLifeTestObject52F460{}, &channelLifeTestObject52F460{}
	r := &channelLifeTestRecord52F460{target: first, mode: 1}
	h := channelLifeTestHooks52F460(&events)
	h.addMana = func(got *channelLifeTestObject52F460, n int16) {
		if got != first || n != 20 {
			t.Fatalf("add = %p/%d", got, n)
		}
		r.target = second
	}
	h.clearDamage = func(got *channelLifeTestObject52F460, n int32) {
		if got != second || n != 20 {
			t.Fatalf("clear = %p/%d", got, n)
		}
	}
	if got := spellChannelLifeUpdate52F460(r, h); got != 1 {
		t.Fatalf("update = %d, want 1", got)
	}
	if want := []string{"target", "flags", "mode", "target"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestChannelLife52F460Rounding(t *testing.T) {
	for _, tc := range []struct {
		in   float32
		want int32
	}{
		{0.5, 0}, {1.5, 2}, {2.5, 2}, {-1.5, -2},
		{float32(math.NaN()), math.MinInt32}, {float32(math.Inf(1)), math.MinInt32},
	} {
		if got := channelLifeRound52F460(tc.in); got != tc.want {
			t.Errorf("round(%g) = %d, want %d", tc.in, got, tc.want)
		}
	}
}
