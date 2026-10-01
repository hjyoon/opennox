package server

import (
	"fmt"
	"reflect"
	"testing"
)

type questWeaponTestItem54CC40 struct {
	name                   string
	flags, class, subclass uint32
	next                   *questWeaponTestItem54CC40
	data                   *questWeaponTestData54CC40
	usable                 int32
}

type questWeaponTestData54CC40 struct{ mods [4]*int }
type questWeaponTestUpdate54CC40 struct{ class uint8 }

type questWeaponTestState54CC40 struct {
	head    *questWeaponTestItem54CC40
	update  *questWeaponTestUpdate54CC40
	trace   []string
	deleted *questWeaponTestItem54CC40
}

func (s *questWeaponTestState54CC40) hooks() questLoseWeaponHooks54CC40[*questWeaponTestItem54CC40, *questWeaponTestUpdate54CC40, *questWeaponTestData54CC40, *int] {
	return questLoseWeaponHooks54CC40[*questWeaponTestItem54CC40, *questWeaponTestUpdate54CC40, *questWeaponTestData54CC40, *int]{
		updateData: func(*questWeaponTestItem54CC40) *questWeaponTestUpdate54CC40 {
			s.trace = append(s.trace, "update")
			return s.update
		},
		first: func(*questWeaponTestItem54CC40) *questWeaponTestItem54CC40 {
			s.trace = append(s.trace, "first")
			return s.head
		},
		next: func(item *questWeaponTestItem54CC40) *questWeaponTestItem54CC40 {
			s.trace = append(s.trace, "next:"+item.name)
			return item.next
		},
		flags: func(item *questWeaponTestItem54CC40) uint32 {
			s.trace = append(s.trace, "flags:"+item.name)
			return item.flags
		},
		class: func(item *questWeaponTestItem54CC40) uint32 {
			s.trace = append(s.trace, "class:"+item.name)
			return item.class
		},
		subclass: func(item *questWeaponTestItem54CC40) uint32 {
			s.trace = append(s.trace, "sub:"+item.name)
			return item.subclass
		},
		modifierData: func(item *questWeaponTestItem54CC40) *questWeaponTestData54CC40 {
			s.trace = append(s.trace, "data:"+item.name)
			return item.data
		},
		modifier: func(data *questWeaponTestData54CC40, slot int) *int {
			s.trace = append(s.trace, fmt.Sprintf("mod:%d", slot))
			return data.mods[slot]
		},
		playerClass: func(update *questWeaponTestUpdate54CC40) uint8 {
			s.trace = append(s.trace, "player-class")
			return update.class
		},
		classCanUse: func(item *questWeaponTestItem54CC40, class uint8) int32 {
			s.trace = append(s.trace, fmt.Sprintf("usable:%s:%d", item.name, class))
			return item.usable
		},
		delayedDelete: func(item *questWeaponTestItem54CC40) {
			s.trace = append(s.trace, "delete:"+item.name)
			s.deleted = item
		},
	}
}

func TestQuestLoseWeapon54CC40ExactSearchAndProtection(t *testing.T) {
	for _, tc := range []struct {
		name                   string
		flags, class, subclass uint32
		mods                   [4]*int
		loads                  []string
		second                 bool
	}{
		{"unequipped", 0xfffffeff, 0x01001000, 0, [4]*int{}, []string{"flags:w", "next:w"}, false},
		{"not-weapon", 0x100, 0xfeffefff, 0, [4]*int{}, []string{"flags:w", "class:w", "next:w"}, false},
		{"subclass-byte-two", 0x100, 0x01001000, 0x80000002, [4]*int{}, []string{"flags:w", "class:w", "sub:w", "next:w"}, false},
		{"protected-subclass", 0x100, 0x01000000, 0x10000, [4]*int{}, []string{"flags:w", "class:w", "sub:w", "sub:w"}, false},
		{"protected-before-modifiers", 0x100, 0x1000, 0x10104, [4]*int{}, []string{"flags:w", "class:w", "sub:w", "sub:w"}, false},
		{"plain-bit-four", 0x100, 0x1000, 4, [4]*int{}, []string{"flags:w", "class:w", "sub:w", "sub:w", "data:w", "mod:0", "mod:1", "mod:2", "mod:3"}, false},
		{"plain-bit-256", 0x100, 0x01000000, 0x100, [4]*int{}, []string{"flags:w", "class:w", "sub:w", "sub:w", "data:w", "mod:0", "mod:1", "mod:2", "mod:3"}, false},
		{"non-modifier-subclass", 0x80000100, 0x81001000, 0x80000200, [4]*int{}, []string{"flags:w", "class:w", "sub:w", "sub:w"}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := &questWeaponTestItem54CC40{name: "w", flags: tc.flags, class: tc.class, subclass: tc.subclass, data: &questWeaponTestData54CC40{mods: tc.mods}}
			s := &questWeaponTestState54CC40{head: w}
			questLoseWeapon54CC40(w, s.hooks())
			want := append([]string{"update", "first"}, tc.loads...)
			if tc.second {
				want = append(want, "first", "class:w", "flags:w", "next:w")
			}
			if !reflect.DeepEqual(s.trace, want) || s.deleted != nil {
				t.Fatalf("trace=%v delete=%p, want %v/no deletion", s.trace, s.deleted, want)
			}
		})
	}
	// A protected first match returns; a later equipped weapon is not selected.
	later := &questWeaponTestItem54CC40{name: "later", flags: 0x100, class: 0x1000}
	first := &questWeaponTestItem54CC40{name: "first", flags: 0x100, class: 0x1000, subclass: 0x10000, next: later}
	s := &questWeaponTestState54CC40{head: first}
	questLoseWeapon54CC40(first, s.hooks())
	if want := []string{"update", "first", "flags:first", "class:first", "sub:first", "sub:first"}; !reflect.DeepEqual(s.trace, want) {
		t.Fatalf("protected first trace=%v, want %v", s.trace, want)
	}
	s = &questWeaponTestState54CC40{}
	questLoseWeapon54CC40(first, s.hooks())
	if !reflect.DeepEqual(s.trace, []string{"update", "first"}) {
		t.Fatalf("empty inventory trace=%v", s.trace)
	}
}

func TestQuestLoseWeapon54CC40EveryModifierSlotAndExactOne(t *testing.T) {
	mod := 1
	for slot := -1; slot < 4; slot++ {
		for _, result := range []int32{-2147483648, -1, 0, 1, 2, 2147483647} {
			t.Run(fmt.Sprintf("slot%d-result%d", slot, result), func(t *testing.T) {
				spare := &questWeaponTestItem54CC40{name: "spare", flags: 0x80000000, class: 0x01000000, subclass: 0x10002, usable: result}
				data := &questWeaponTestData54CC40{}
				if slot >= 0 {
					data.mods[slot] = &mod
				}
				w := &questWeaponTestItem54CC40{name: "w", flags: 0x100, class: 0x1000, subclass: 0x104, data: data, next: spare}
				s := &questWeaponTestState54CC40{head: w, update: &questWeaponTestUpdate54CC40{class: 255}}
				questLoseWeapon54CC40(w, s.hooks())
				want := []string{"update", "first", "flags:w", "class:w", "sub:w", "sub:w", "data:w", "mod:0", "mod:1", "mod:2", "mod:3"}
				if slot >= 0 {
					want = append(want, "first", "class:w", "flags:w", "next:w", "class:spare", "flags:spare", "player-class", "usable:spare:255", "next:spare")
					if result == 1 {
						want = append(want, "delete:w")
					}
				}
				if !reflect.DeepEqual(s.trace, want) || (s.deleted == w) != (slot >= 0 && result == 1) {
					t.Fatalf("trace=%v delete=%p, want %v", s.trace, s.deleted, want)
				}
			})
		}
	}
}

func TestQuestLoseWeapon54CC40CachedUpdateFreshHeadLiveSpareWalk(t *testing.T) {
	bypassed := &questWeaponTestItem54CC40{name: "bypassed", class: 0x1000, usable: 1}
	tail := &questWeaponTestItem54CC40{name: "tail", class: 0x1000, usable: 0}
	b := &questWeaponTestItem54CC40{name: "b", class: 0x1000, usable: 1, next: bypassed}
	a := &questWeaponTestItem54CC40{name: "a", class: 0x01000000, usable: 1, next: b}
	w := &questWeaponTestItem54CC40{name: "w", flags: 0x100, class: 0x1000}
	cached := &questWeaponTestUpdate54CC40{class: 2}
	s := &questWeaponTestState54CC40{head: w, update: cached}
	h := s.hooks()
	first := h.first
	firstCalls := 0
	h.first = func(unit *questWeaponTestItem54CC40) *questWeaponTestItem54CC40 {
		firstCalls++
		if firstCalls == 2 {
			s.head = a
		}
		return first(unit)
	}
	canUse := h.classCanUse
	h.classCanUse = func(item *questWeaponTestItem54CC40, class uint8) int32 {
		result := canUse(item, class)
		switch item {
		case a:
			cached.class = 255
			s.update = &questWeaponTestUpdate54CC40{class: 17}
			s.head = bypassed                // The second walk must not restart from this head.
			w.flags, w.subclass = 0, 0x10000 // The selected weapon stays cached.
		case b:
			b.next = tail // Next is read after the callback, even after a match.
		}
		return result
	}
	questLoseWeapon54CC40(w, h)
	want := []string{"update", "first", "flags:w", "class:w", "sub:w", "sub:w", "first", "class:a", "flags:a", "player-class", "usable:a:2", "next:a", "class:b", "flags:b", "player-class", "usable:b:255", "next:b", "class:tail", "flags:tail", "player-class", "usable:tail:255", "next:tail", "delete:w"}
	if !reflect.DeepEqual(s.trace, want) || s.deleted != w {
		t.Fatalf("trace=%v delete=%p, want %v/%p", s.trace, s.deleted, want, w)
	}
}

func TestQuestLoseWeapon54CC40ModifierDataCachedButAllSlotsLive(t *testing.T) {
	mod := 1
	old := &questWeaponTestData54CC40{mods: [4]*int{&mod}}
	replacement := &questWeaponTestData54CC40{}
	spare := &questWeaponTestItem54CC40{name: "spare", class: 0x1000, usable: 1}
	w := &questWeaponTestItem54CC40{name: "w", flags: 0x100, class: 0x1000, subclass: 4, data: old, next: spare}
	s := &questWeaponTestState54CC40{head: w, update: &questWeaponTestUpdate54CC40{}}
	h := s.hooks()
	reads := 0
	h.modifier = func(data *questWeaponTestData54CC40, slot int) *int {
		if data != old || slot != reads {
			t.Fatal("modifier array was reloaded or a slot was skipped")
		}
		reads++
		if slot == 0 {
			w.data = replacement
			old.mods[3] = &mod
		}
		if slot == 3 && data.mods[3] != &mod {
			t.Fatal("later slot was a snapshot")
		}
		return data.mods[slot]
	}
	questLoseWeapon54CC40(w, h)
	if reads != 4 || s.deleted != w {
		t.Fatalf("reads=%d deleted=%p, want 4/%p", reads, s.deleted, w)
	}
}

func TestQuestLoseWeapon54CC40SpareGatesAndFaultPrefix(t *testing.T) {
	nonWeapon := &questWeaponTestItem54CC40{name: "non-weapon", class: 0x02000000}
	equipped := &questWeaponTestItem54CC40{name: "equipped", flags: 0x100, class: 0x1000, next: nonWeapon}
	w := &questWeaponTestItem54CC40{name: "w", flags: 0x100, class: 0x1000, next: equipped}
	s := &questWeaponTestState54CC40{head: w}
	questLoseWeapon54CC40(w, s.hooks()) // Missing player binding is never read.
	want := []string{"update", "first", "flags:w", "class:w", "sub:w", "sub:w", "first", "class:w", "flags:w", "next:w", "class:equipped", "flags:equipped", "next:equipped", "class:non-weapon", "next:non-weapon"}
	if !reflect.DeepEqual(s.trace, want) || s.deleted != nil {
		t.Fatalf("trace=%v delete=%p, want %v", s.trace, s.deleted, want)
	}
	spare := &questWeaponTestItem54CC40{name: "spare", class: 0x1000, usable: 1}
	w.next = spare
	s = &questWeaponTestState54CC40{head: w}
	defer func() {
		if recover() == nil {
			t.Fatal("missing update was normalized")
		}
		want := []string{"update", "first", "flags:w", "class:w", "sub:w", "sub:w", "first", "class:w", "flags:w", "next:w", "class:spare", "flags:spare", "player-class"}
		if !reflect.DeepEqual(s.trace, want) {
			t.Fatalf("fault trace=%v, want %v", s.trace, want)
		}
	}()
	questLoseWeapon54CC40(w, s.hooks())
}
