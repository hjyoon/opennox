package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/prand"
)

func TestQuestLoseArmor54CD30ExactGatesAndOrder(t *testing.T) {
	for _, test := range []struct {
		name         string
		flags, class uint32
		typ          uint16
		armor        uint32
		loads        []string
		eligible     bool
	}{
		{"unequipped", 0xfffffeff, 0x02000000, 65535, 0, []string{"flags"}, false},
		{"non-armor", 0x100, 0xfdffffff, 65535, 0, []string{"flags", "class"}, false},
		{"sneakers", 0x100, 0x02000000, 31, 1, []string{"flags", "class", "type", "armor:31"}, false},
		{"pants", 0x100, 0x02000000, 32, 4, []string{"flags", "class", "type", "armor:32"}, false},
		{"shirt", 0x100, 0x02000000, 33, 0x400, []string{"flags", "class", "type", "armor:33"}, false},
		{"combined-protection", 0x100, 0x02000000, 34, 0xffffffff, []string{"flags", "class", "type", "armor:34"}, false},
		{"unknown-type-word", 0xffffffff, 0xffffffff, 65535, 0, []string{"flags", "class", "type", "armor:65535"}, true},
		{"other-armor-bits", 0x80000100, 0x82000000, 32768, 0x80008000, []string{"flags", "class", "type", "armor:32768"}, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			var trace []string
			questLoseArmor54CD30(1, questLoseArmorHooks54CD30[int]{
				first: func(unit int) int {
					if unit != 1 {
						t.Fatal("wrong unit")
					}
					trace = append(trace, "first")
					return 2
				},
				next:      func(int) int { trace = append(trace, "next"); return 0 },
				flags:     func(int) uint32 { trace = append(trace, "flags"); return test.flags },
				class:     func(int) uint32 { trace = append(trace, "class"); return test.class },
				typeIndex: func(int) uint16 { trace = append(trace, "type"); return test.typ },
				armorFlags: func(typ uint16) uint32 {
					trace = append(trace, fmt.Sprintf("armor:%d", typ))
					return test.armor
				},
				randomInt: func(minimum, maximum int32) int32 {
					trace = append(trace, "rng")
					if minimum != 0 || maximum != 0 {
						t.Fatalf("bounds = %d..%d", minimum, maximum)
					}
					return 0
				},
				delayedDelete: func(item int) {
					if item != 2 {
						t.Fatal("wrong deletion")
					}
					trace = append(trace, "delete")
				},
			})
			want := append([]string{"first"}, test.loads...)
			want = append(want, "next")
			if test.eligible {
				want = append(want, "rng", "first")
				want = append(want, test.loads...)
				want = append(want, "delete") // No next-link read after deletion.
			}
			if !reflect.DeepEqual(trace, want) {
				t.Fatalf("trace = %v, want %v", trace, want)
			}
		})
	}
	firstCalls := 0
	questLoseArmor54CD30(1, questLoseArmorHooks54CD30[int]{first: func(int) int { firstCalls++; return 0 }})
	if firstCalls != 1 {
		t.Fatalf("empty inventory reads = %d", firstCalls)
	}
}

func TestQuestLoseArmor54CD30LiveLinksFreshHeadAndCachedGates(t *testing.T) {
	type item struct {
		name  string
		typ   uint16
		flags uint32
		next  *item
	}
	bypassed := &item{name: "bypassed", typ: 9, flags: 0x100}
	d := &item{name: "d", typ: 2, flags: 0x100}
	a := &item{name: "a", typ: 1, flags: 0x100, next: bypassed}
	y := &item{name: "y", typ: 4, flags: 0x100}
	x := &item{name: "x", typ: 3, flags: 0x100, next: y}
	replacement := &item{name: "replacement", typ: 10, flags: 0x100}
	u := &item{name: "unit", next: a}
	var trace []string
	questLoseArmor54CD30(u, questLoseArmorHooks54CD30[*item]{
		first: func(unit *item) *item { trace = append(trace, "first"); return unit.next },
		next: func(it *item) *item {
			trace = append(trace, "next:"+it.name)
			return it.next
		},
		flags:     func(it *item) uint32 { trace = append(trace, "flags:"+it.name); return it.flags },
		class:     func(it *item) uint32 { trace = append(trace, "class:"+it.name); return 0x02000000 },
		typeIndex: func(it *item) uint16 { trace = append(trace, "type:"+it.name); return it.typ },
		armorFlags: func(typ uint16) uint32 {
			trace = append(trace, fmt.Sprintf("armor:%d", typ))
			switch typ {
			case 1:
				a.flags, a.next = 0, d // Already-read gates stay cached; next is live.
			case 3:
				x.flags, u.next = 0, replacement // Does not restart this traversal.
			case 4:
				y.flags, y.next = 0, replacement // Still deletes cached y, then returns.
			case 9, 10:
				t.Fatal("read a bypassed item")
			}
			return 0
		},
		randomInt: func(minimum, maximum int32) int32 {
			trace = append(trace, "rng")
			if minimum != 0 || maximum != 1 {
				t.Fatalf("bounds = %d..%d", minimum, maximum)
			}
			u.next = x
			return 1
		},
		delayedDelete: func(it *item) {
			trace = append(trace, "delete:"+it.name)
			if it != y {
				t.Fatal("deleted a snapshot candidate instead of the live ordinal")
			}
		},
	})
	want := []string{"first", "flags:a", "class:a", "type:a", "armor:1", "next:a", "flags:d", "class:d", "type:d", "armor:2", "next:d", "rng", "first", "flags:x", "class:x", "type:x", "armor:3", "next:x", "flags:y", "class:y", "type:y", "armor:4", "delete:y"}
	if !reflect.DeepEqual(trace, want) {
		t.Fatalf("trace = %v, want %v", trace, want)
	}
}

func TestQuestLoseArmor54CD30UnmatchedOrdinalAndChangedCandidates(t *testing.T) {
	for _, test := range []struct {
		name   string
		draw   int32
		head   int
		remove bool
		nexts  int
	}{
		{"negative-ordinal", -1, 1, false, 4},
		{"out-of-range-ordinal", 2, 1, false, 4},
		{"fresh-empty-head", 0, 0, false, 2},
		{"removed-candidates", 0, 1, true, 4},
	} {
		t.Run(test.name, func(t *testing.T) {
			head, firsts, nexts, rngs, deletes := 1, 0, 0, 0, 0
			removed := false
			questLoseArmor54CD30(3, questLoseArmorHooks54CD30[int]{
				first: func(int) int { firsts++; return head },
				next: func(it int) int {
					nexts++
					if it == 1 {
						return 2
					}
					return 0
				},
				flags: func(int) uint32 {
					if removed {
						return 0
					}
					return 0x100
				},
				class:      func(int) uint32 { return 0x02000000 },
				typeIndex:  func(it int) uint16 { return uint16(it) },
				armorFlags: func(uint16) uint32 { return 0 },
				randomInt: func(minimum, maximum int32) int32 {
					rngs++
					if minimum != 0 || maximum != 1 {
						t.Fatal("wrong initial count")
					}
					head, removed = test.head, test.remove
					return test.draw
				},
				delayedDelete: func(int) { deletes++ },
			})
			if firsts != 2 || nexts != test.nexts || rngs != 1 || deletes != 0 {
				t.Fatalf("calls = first:%d next:%d RNG:%d delete:%d", firsts, nexts, rngs, deletes)
			}
		})
	}
}

func TestQuestLoseArmor54CD30NativeRNGAndUntouchedFields(t *testing.T) {
	// Independent inventory ordinals: only StreetShirt/Pants/Sneakers are
	// protected by the GAME.EXE 0x405 mask, not every piece of clothing.
	eligible := []int{0, 1, 2, 3, 4, 5, 6, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25}
	for seed := 0; seed < 48; seed++ {
		s := new(Server)
		s.Armor.table = armorTable
		s.Rand.Logic, s.Rand.Other = prand.New(seed), prand.New(seed+11)
		otherIndex := s.Rand.Other.Index()
		items := make([]Object, PlayerArmorCnt)
		beforeItems := make([]Object, PlayerArmorCnt)
		for i := range items {
			s.Armor.table[i].TypeInd = 500 + i
			items[i] = Object{TypeInd: uint16(500 + i), ObjClass: 0x82000000, ObjFlags: 0x80000100, Worth: uint32(1000 + i), Field5: 0xaabbccdd}
			if i+1 < len(items) {
				items[i].InvNextItem = &items[i+1]
			}
		}
		copy(beforeItems, items)
		u := &Object{InvFirstItem: &items[0], ObjClass: 0, Worth: 0x12345678}
		beforeUnit := *u // No Player/UpdateData binding is required.
		wantRNG := prand.New(seed)
		selected := eligible[wantRNG.IntClamp(0, len(eligible)-1)]
		deletes := 0
		s.QuestLoseArmor54CD30(u, func(item *Object) {
			deletes++
			if item != &items[selected] {
				t.Fatalf("seed %d deleted %p, want %p", seed, item, &items[selected])
			}
		})
		if deletes != 1 || *u != beforeUnit || !reflect.DeepEqual(items, beforeItems) {
			t.Fatalf("seed %d changed unrelated fields or made %d deletions", seed, deletes)
		}
		if s.Rand.Logic.Index() != wantRNG.Index() || s.Rand.Other.Index() != otherIndex {
			t.Fatal("wrong RNG stream or draw count")
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(u)) <= math.MaxUint32 || uintptr(unsafe.Pointer(&items[selected])) <= math.MaxUint32) {
			t.Fatal("fixture pointers must exceed 4 GiB")
		}
	}
}

func TestQuestLoseArmor54CD30NativeEmptyProtectedAndMissingBindings(t *testing.T) {
	s := new(Server)
	s.Armor.table = armorTable
	for i := range s.Armor.table {
		s.Armor.table[i].TypeInd = 500 + i
	}
	s.Rand.Logic = nil // An empty/protected candidate set must not use it.
	u := new(Object)
	s.QuestLoseArmor54CD30(u, nil)
	for _, index := range []int{7, 8, 9} {
		u.InvFirstItem = &Object{TypeInd: uint16(500 + index), ObjFlags: object.FlagEquipped, ObjClass: object.ClassArmor}
		before := *u.InvFirstItem
		s.QuestLoseArmor54CD30(u, nil)
		if *u.InvFirstItem != before {
			t.Fatal("protected armor was mutated")
		}
	}
	for _, test := range []struct {
		name string
		u    *Object
		rng  *prand.Rand
	}{
		{"missing-unit", nil, nil},
		{"missing-rng", &Object{InvFirstItem: &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, TypeInd: 65535}}, nil},
		{"missing-delete", &Object{InvFirstItem: &Object{ObjClass: object.ClassArmor, ObjFlags: object.FlagEquipped, TypeInd: 65535}}, prand.New(7)},
	} {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("missing binding was silently ignored")
				}
			}()
			s.Rand.Logic = test.rng
			s.QuestLoseArmor54CD30(test.u, nil)
		})
	}
}

func TestQuestLoseArmor54CD30NativeLayout(t *testing.T) {
	wantType, wantClass, wantFlags, wantNext, wantFirst := uintptr(4), uintptr(8), uintptr(16), uintptr(496), uintptr(504)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantType, wantClass, wantFlags, wantNext, wantFirst = 8, 12, 20, 528, 544
	}
	for _, check := range []struct{ got, want uintptr }{
		{unsafe.Offsetof(Object{}.TypeInd), wantType}, {unsafe.Sizeof(Object{}.TypeInd), 2},
		{unsafe.Offsetof(Object{}.ObjClass), wantClass}, {unsafe.Offsetof(Object{}.ObjFlags), wantFlags},
		{unsafe.Offsetof(Object{}.InvNextItem), wantNext}, {unsafe.Offsetof(Object{}.InvFirstItem), wantFirst},
	} {
		if check.got != check.want {
			t.Errorf("native offset/size = %d, want %d", check.got, check.want)
		}
	}
}
