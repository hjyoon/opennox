package server

import (
	"math"
	"reflect"
	"testing"

	"github.com/opennox/libs/types"
)

type greaterHealTestRecord52F220 struct {
	mode, spell, level uint32
	caster, target     uint64
	aim                types.Pointf
	fraction           float32
}

func greaterHealCreateTestHooks52F220(events *[]string, found uint64) greaterHealCreateHooks52F220[*greaterHealTestRecord52F220, uint64] {
	log := func(event string) { *events = append(*events, event) }
	return greaterHealCreateHooks52F220[*greaterHealTestRecord52F220, uint64]{
		loadMode:   func(r *greaterHealTestRecord52F220) uint32 { log("mode"); return r.mode },
		loadCaster: func(r *greaterHealTestRecord52F220) uint64 { log("caster"); return r.caster },
		loadSpell:  func(r *greaterHealTestRecord52F220) uint32 { log("spell"); return r.spell },
		loadAim:    func(r *greaterHealTestRecord52F220) *types.Pointf { log("aim"); return &r.aim },
		spellFlags: func(spell uint32) uint32 {
			log("flags")
			if spell != 35 {
				panic("wrong spell")
			}
			return 0x123
		},
		searchTarget: func(aim *types.Pointf, caster uint64, flags uint32, distance float32, mode int, self uint64) uint64 {
			log("search")
			if *aim != (types.Ptf(12, 34)) || flags != 0x123 || distance != 400 || mode != 1 || caster != self {
				panic("wrong search arguments")
			}
			return found
		},
		adjustHP: func(target uint64, amount int32) {
			log("heal")
			if target != found || amount != 20 {
				panic("wrong instant heal")
			}
		},
		storeTarget: func(r *greaterHealTestRecord52F220, target uint64) {
			log("store")
			r.target = target
		},
		isEnemy: func(caster, target uint64) bool {
			log("enemy")
			return caster == 0x100000123 && target == 0x200000456
		},
		startRay: func(*greaterHealTestRecord52F220) { log("ray") },
		noTarget: func(uint64) { log("no-target") },
	}
}

func TestGreaterHealCreate52F220Paths(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mode   uint32
		caster uint64
		found  uint64
		want   int32
		state  uint64
		events []string
	}{
		{"no caster or mode", 0, 0, 0, 1, 0, []string{"mode", "caster"}},
		{"glyph heals", 1, 0, 0x200000456, 1, 0, []string{"mode", "caster", "spell", "flags", "aim", "search", "heal"}},
		{"glyph no target", 1, 0, 0, 1, 0, []string{"mode", "caster", "spell", "flags", "aim", "search"}},
		{"mode with caster heals", 1, 0x100000123, 0x200000456, 1, 0, []string{"mode", "caster", "spell", "flags", "aim", "search", "heal"}},
		{"ordinary friendly", 0, 0x100000123, 0x300000456, 0, 0x300000456, []string{"mode", "caster", "caster", "spell", "flags", "aim", "search", "store", "caster", "enemy", "ray"}},
		{"ordinary enemy", 0, 0x100000123, 0x200000456, 1, 0x200000456, []string{"mode", "caster", "caster", "spell", "flags", "aim", "search", "store", "caster", "enemy"}},
		{"ordinary no target", 0, 0x100000123, 0, 1, 0, []string{"mode", "caster", "caster", "spell", "flags", "aim", "search", "store", "caster", "no-target"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			r := &greaterHealTestRecord52F220{mode: tc.mode, caster: tc.caster, spell: 35, aim: types.Ptf(12, 34)}
			h := greaterHealCreateTestHooks52F220(&events, tc.found)
			search := h.searchTarget
			h.searchTarget = func(aim *types.Pointf, caster uint64, flags uint32, distance float32, mode int, self uint64) uint64 {
				if tc.mode != 0 && (caster != 0 || self != 0) {
					t.Fatal("glyph search should have no caster")
				}
				if tc.mode == 0 && tc.caster != 0 && (caster != tc.caster || self != tc.caster) {
					t.Fatal("ordinary search should retain native caster")
				}
				return search(aim, caster, flags, distance, mode, self)
			}
			if got := spellGreaterHealCreate52F220(r, h); got != tc.want || r.target != tc.state || !reflect.DeepEqual(events, tc.events) {
				t.Fatalf("create=%d target=%#x events=%v, want %d/%#x/%v", got, r.target, events, tc.want, tc.state, tc.events)
			}
		})
	}
}

type greaterHealTestObject52F2E0 struct {
	flags, class           uint32
	mana, hp, max          uint16
	buffed, moved, damaged bool
}

type greaterHealTestRecord52F2E0 struct {
	target, caster *greaterHealTestObject52F2E0
	level          uint32
	fraction       float32
}

func greaterHealUpdateTestHooks52F2E0(events *[]string) greaterHealUpdateHooks52F2E0[*greaterHealTestRecord52F2E0, *greaterHealTestObject52F2E0] {
	log := func(event string) { *events = append(*events, event) }
	return greaterHealUpdateHooks52F2E0[*greaterHealTestRecord52F2E0, *greaterHealTestObject52F2E0]{
		loadTarget: func(r *greaterHealTestRecord52F2E0) *greaterHealTestObject52F2E0 { log("target"); return r.target },
		loadFlags:  func(o *greaterHealTestObject52F2E0) uint32 { log("flags"); return o.flags },
		loadCaster: func(r *greaterHealTestRecord52F2E0) *greaterHealTestObject52F2E0 { log("caster"); return r.caster },
		testBuff: func(o *greaterHealTestObject52F2E0, buff int32) int32 {
			log("buff")
			if buff != 8 {
				panic("wrong buff")
			}
			if o.buffed {
				return 1
			}
			return 0
		},
		canInteract: func(*greaterHealTestObject52F2E0, *greaterHealTestObject52F2E0) bool { log("interact"); return true },
		oldMana: func(o *greaterHealTestObject52F2E0) uint16 {
			log("mana")
			if o == nil {
				return 0
			}
			return o.mana
		},
		loadClass: func(o *greaterHealTestObject52F2E0) uint8 { log("class"); return uint8(o.class) },
		positionDelta: func(o *greaterHealTestObject52F2E0, _ *greaterHealTestRecord52F2E0) int32 {
			log("position")
			if o.moved {
				return 1
			}
			return 0
		},
		wasDamaged:   func(o *greaterHealTestObject52F2E0) bool { log("damage"); return o.damaged },
		maxHP:        func(o *greaterHealTestObject52F2E0) uint16 { log("max"); return o.max },
		getHP:        func(o *greaterHealTestObject52F2E0) uint16 { log("hp"); return o.hp },
		loadFraction: func(r *greaterHealTestRecord52F2E0) float32 { log("fraction"); return r.fraction },
		loadLevel:    func(r *greaterHealTestRecord52F2E0) uint32 { log("level"); return r.level },
		coefficient: func(level uint32) float32 {
			log("coefficient")
			if level != 2 {
				panic("wrong level")
			}
			return 0.5
		},
		playerClass: func(*greaterHealTestObject52F2E0) uint8 { log("player-class"); return 1 },
		classHealth: func(class uint8) float32 {
			log("health-factor")
			if class != 1 {
				panic("wrong class")
			}
			return 2
		},
		storeFraction: func(r *greaterHealTestRecord52F2E0, fraction float32) { log("store"); r.fraction = fraction },
		adjustHP:      func(o *greaterHealTestObject52F2E0, amount int32) { log("heal"); o.hp += uint16(amount) },
		manaSub:       func(o *greaterHealTestObject52F2E0, amount int32) { log("subtract"); o.mana -= uint16(amount) },
	}
}

func TestGreaterHealUpdate52F2E0AccumulationAndPlayerMultiplier(t *testing.T) {
	for _, tc := range []struct {
		name         string
		class        uint32
		hp           uint16
		fraction     float32
		wantHP       uint16
		wantFraction float32
	}{
		{"monster", 2, 5, 0.25, 6, -0.25},
		{"player", 4, 5, 0.25, 7, -0.5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			caster := &greaterHealTestObject52F2E0{class: tc.class, mana: 5}
			target := &greaterHealTestObject52F2E0{hp: tc.hp, max: 10}
			r := &greaterHealTestRecord52F2E0{target: target, caster: caster, level: 2, fraction: tc.fraction}
			if got := spellGreaterHealUpdate52F2E0(r, greaterHealUpdateTestHooks52F2E0(&events)); got != 0 || target.hp != tc.wantHP || caster.mana != 4 || r.fraction != tc.wantFraction {
				t.Fatalf("update=%d hp=%d mana=%d fraction=%g events=%v", got, target.hp, caster.mana, r.fraction, events)
			}
			if !reflect.DeepEqual(events[len(events)-5:], []string{"store", "target", "heal", "caster", "subtract"}) {
				t.Fatalf("unexpected effect order: %v", events)
			}
		})
	}
}

func TestGreaterHealUpdate52F2E0StopsBeforeEffects(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*greaterHealTestRecord52F2E0)
	}{
		{"nil target", func(r *greaterHealTestRecord52F2E0) { r.target = nil }},
		{"dead target", func(r *greaterHealTestRecord52F2E0) { r.target.flags = 0x8000 }},
		{"destroyed target", func(r *greaterHealTestRecord52F2E0) { r.target.flags = 0x20 }},
		{"buffed caster", func(r *greaterHealTestRecord52F2E0) { r.caster.buffed = true }},
		{"no mana", func(r *greaterHealTestRecord52F2E0) { r.caster.mana = 0 }},
		{"moved monster", func(r *greaterHealTestRecord52F2E0) { r.caster.moved = true }},
		{"recent damage", func(r *greaterHealTestRecord52F2E0) { r.caster.damaged = true }},
		{"full health", func(r *greaterHealTestRecord52F2E0) { r.target.hp = r.target.max }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			r := &greaterHealTestRecord52F2E0{
				target: &greaterHealTestObject52F2E0{hp: 5, max: 10},
				caster: &greaterHealTestObject52F2E0{class: 2, mana: 5}, level: 2,
			}
			tc.mutate(r)
			var hp uint16
			if r.target != nil {
				hp = r.target.hp
			}
			mana := r.caster.mana
			if got := spellGreaterHealUpdate52F2E0(r, greaterHealUpdateTestHooks52F2E0(&events)); got != 1 || r.target != nil && r.target.hp != hp || r.caster.mana != mana {
				t.Fatalf("update=%d target=%+v caster=%+v events=%v", got, r.target, r.caster, events)
			}
			for _, event := range events {
				if event == "store" || event == "heal" || event == "subtract" {
					t.Fatalf("unexpected effect %s: %v", event, events)
				}
			}
		})
	}
}

func TestGreaterHealUpdate52F2E0StopsWhenTargetNotVisible(t *testing.T) {
	var events []string
	r := &greaterHealTestRecord52F2E0{
		target: &greaterHealTestObject52F2E0{hp: 5, max: 10},
		caster: &greaterHealTestObject52F2E0{class: 4, mana: 5}, level: 2,
	}
	h := greaterHealUpdateTestHooks52F2E0(&events)
	h.canInteract = func(*greaterHealTestObject52F2E0, *greaterHealTestObject52F2E0) bool {
		events = append(events, "interact")
		return false
	}
	if got := spellGreaterHealUpdate52F2E0(r, h); got != 1 {
		t.Fatalf("update = %d", got)
	}
	want := []string{"target", "flags", "caster", "buff", "caster", "target", "interact"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestGreaterHealUpdate52F2E0UnknownPlayerClassSkipsMultiplier(t *testing.T) {
	var events []string
	r := &greaterHealTestRecord52F2E0{
		target: &greaterHealTestObject52F2E0{hp: 5, max: 10},
		caster: &greaterHealTestObject52F2E0{class: 4, mana: 5}, level: 2, fraction: 0.25,
	}
	h := greaterHealUpdateTestHooks52F2E0(&events)
	h.playerClass = func(*greaterHealTestObject52F2E0) uint8 { events = append(events, "player-class"); return 3 }
	if got := spellGreaterHealUpdate52F2E0(r, h); got != 0 || r.target.hp != 6 || r.fraction != -0.25 {
		t.Fatalf("update=%d hp=%d fraction=%g", got, r.target.hp, r.fraction)
	}
	for _, event := range events {
		if event == "health-factor" {
			t.Fatalf("unknown class read health multiplier: %v", events)
		}
	}
}

func TestSpellDurationRoundNearestEven(t *testing.T) {
	for _, tc := range []struct {
		value float32
		want  int32
	}{
		{0.5, 0}, {1.5, 2}, {-0.5, 0}, {-1.5, -2}, {float32(math.Inf(1)), math.MinInt32},
	} {
		if got := spellDurationRoundNearestEven(tc.value); got != tc.want {
			t.Fatalf("round(%g)=%d, want %d", tc.value, got, tc.want)
		}
	}
}
