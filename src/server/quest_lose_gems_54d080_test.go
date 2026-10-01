package server

import (
	"fmt"
	"reflect"
	"testing"
)

type questGemItem54D080 struct {
	name  string
	kind  uint16
	price int32
	next  *questGemItem54D080
}

type questGemState54D080 struct {
	types, resolved [3]uint32
	head            *questGemItem54D080
	trace           []string
	deleted         []*questGemItem54D080
	credits         []int32
}

func (s *questGemState54D080) hooks() questLoseGemsHooks54D080[*questGemItem54D080] {
	return questLoseGemsHooks54D080[*questGemItem54D080]{
		loadType: func(slot int) uint32 {
			s.trace = append(s.trace, fmt.Sprintf("cache:%d", slot))
			return s.types[slot]
		},
		storeType: func(slot int, value uint32) {
			s.trace = append(s.trace, fmt.Sprintf("store:%d=%d", slot, value))
			s.types[slot] = value
		},
		lookupType: func(name string) uint32 {
			s.trace = append(s.trace, "lookup:"+name)
			switch name {
			case "Diamond":
				return s.resolved[0]
			case "Emerald":
				return s.resolved[1]
			case "Ruby":
				return s.resolved[2]
			default:
				panic("unexpected gem name")
			}
		},
		first: func(*questGemItem54D080) *questGemItem54D080 {
			s.trace = append(s.trace, "first")
			return s.head
		},
		next: func(item *questGemItem54D080) *questGemItem54D080 {
			s.trace = append(s.trace, "next:"+item.name)
			return item.next
		},
		typeIndex: func(item *questGemItem54D080) uint16 {
			s.trace = append(s.trace, "type:"+item.name)
			return item.kind
		},
		price: func(item *questGemItem54D080) int32 {
			s.trace = append(s.trace, "price:"+item.name)
			return item.price
		},
		delayedDelete: func(item *questGemItem54D080) {
			s.trace = append(s.trace, "delete:"+item.name)
			s.deleted = append(s.deleted, item)
		},
		addGold: func(_ *questGemItem54D080, amount int32) {
			s.trace = append(s.trace, fmt.Sprintf("gold:%d", amount))
			s.credits = append(s.credits, amount)
		},
	}
}

func TestQuestLoseGems54D080LazyCachePublicationAndEmptyWalks(t *testing.T) {
	s := &questGemState54D080{types: [3]uint32{0, 91, 92}, resolved: [3]uint32{11, 12, 13}}
	h := s.hooks()
	lookup := h.lookupType
	h.lookupType = func(name string) uint32 {
		if name == "Emerald" && s.types[0] != 11 || name == "Ruby" && s.types[1] != 12 {
			t.Fatal("previous type was not published before the next lookup")
		}
		return lookup(name)
	}
	questLoseGems54D080[*questGemItem54D080](nil, h)
	want := []string{"cache:0", "lookup:Diamond", "store:0=11", "lookup:Emerald", "store:1=12", "lookup:Ruby", "store:2=13", "first", "first"}
	if !reflect.DeepEqual(s.trace, want) || s.types != [3]uint32{11, 12, 13} {
		t.Fatalf("trace=%v cache=%v", s.trace, s.types)
	}
	s.trace, s.types[1], s.types[2] = nil, 0, 0
	h.lookupType = nil
	questLoseGems54D080[*questGemItem54D080](nil, h)
	if !reflect.DeepEqual(s.trace, []string{"cache:0", "first", "first"}) || s.types != [3]uint32{11, 0, 0} {
		t.Fatal("nonzero Diamond cache must skip all three lookups")
	}
	// A zero Diamond lookup remains the lazy sentinel, including on repetition.
	s = &questGemState54D080{resolved: [3]uint32{0, 12, 13}}
	for repetition := 0; repetition < 2; repetition++ {
		s.trace = nil
		questLoseGems54D080[*questGemItem54D080](nil, s.hooks())
		want := []string{"cache:0", "lookup:Diamond", "store:0=0", "lookup:Emerald", "store:1=12", "lookup:Ruby", "store:2=13", "first", "first"}
		if !reflect.DeepEqual(s.trace, want) {
			t.Fatalf("zero-cache repetition=%d trace=%v", repetition, s.trace)
		}
	}
}

func TestQuestLoseGems54D080AllSmallCountCombinations(t *testing.T) {
	for diamonds := 0; diamonds <= 5; diamonds++ {
		for emeralds := 0; emeralds <= 5; emeralds++ {
			for rubies := 0; rubies <= 5; rubies++ {
				counts := [3]int{diamonds, emeralds, rubies}
				s := &questGemState54D080{types: [3]uint32{11, 32768, 65535}}
				var items, wantDeleted []*questGemItem54D080
				var wantCredits []int32
				// Independent prefix-count oracle: lose ceil(n/2) of each kind,
				// refund price/2 truncated toward zero only for its odd first gem.
				for ordinal := 0; ordinal < 5; ordinal++ {
					items = append(items, &questGemItem54D080{name: "other", kind: 999})
					for slot, count := range counts {
						if ordinal >= count {
							continue
						}
						item := &questGemItem54D080{name: fmt.Sprintf("%d/%d", slot, ordinal), kind: uint16(s.types[slot]), price: int32(9 - 10*slot)}
						items = append(items, item)
						if ordinal < (count+1)/2 {
							wantDeleted = append(wantDeleted, item)
						}
						if ordinal == 0 && count%2 == 1 {
							wantCredits = append(wantCredits, item.price/2)
						}
					}
				}
				for i := 1; i < len(items); i++ {
					items[i-1].next = items[i]
				}
				s.head = items[0]
				questLoseGems54D080(s.head, s.hooks())
				if !reflect.DeepEqual(s.deleted, wantDeleted) || !reflect.DeepEqual(s.credits, wantCredits) {
					t.Fatalf("counts=%v deleted=%v want=%v credits=%v want=%v", counts, s.deleted, wantDeleted, s.credits, wantCredits)
				}
			}
		}
	}
}

func TestQuestLoseGems54D080ExactOddEvenTrace(t *testing.T) {
	c := &questGemItem54D080{name: "c", kind: 11}
	b := &questGemItem54D080{name: "b", kind: 11, next: c}
	a := &questGemItem54D080{name: "a", kind: 11, price: -3, next: b}
	s := &questGemState54D080{types: [3]uint32{11, 12, 13}, head: a}
	questLoseGems54D080(a, s.hooks())
	want := []string{
		"cache:0", "first", "cache:0", "type:a", "next:a", "cache:0", "type:b", "next:b", "cache:0", "type:c", "next:c",
		"first", "next:a", "cache:0", "type:a", "price:a", "delete:a", "gold:-1",
		"next:b", "cache:0", "type:b", "delete:b", "next:c", "cache:0", "type:c",
	}
	if !reflect.DeepEqual(s.trace, want) || !reflect.DeepEqual(s.deleted, []*questGemItem54D080{a, b}) || !reflect.DeepEqual(s.credits, []int32{-1}) {
		t.Fatalf("trace=%v, want %v", s.trace, want)
	}
}

func TestQuestLoseGems54D080FullDWORDTypesAndComparisonPriority(t *testing.T) {
	for _, types := range [][3]uint32{{0x1000b, 0x1000c, 0x1000d}, {11, 11, 11}} {
		b := &questGemItem54D080{name: "b", kind: 11, price: 17}
		a := &questGemItem54D080{name: "a", kind: 11, price: 9, next: b}
		s := &questGemState54D080{types: types, head: a}
		questLoseGems54D080(a, s.hooks())
		if types[0] > 65535 {
			if len(s.deleted) != 0 || len(s.credits) != 0 {
				t.Fatal("DWORD type cache was narrowed to a WORD")
			}
		} else if !reflect.DeepEqual(s.deleted, []*questGemItem54D080{a}) || len(s.credits) != 0 {
			t.Fatal("colliding type IDs did not use Diamond precedence")
		}
	}
}

func TestQuestLoseGems54D080SignedPriceHalves(t *testing.T) {
	for slot := 0; slot < 3; slot++ {
		for _, price := range []int32{-2147483648, -2147483647, -3, -1, 0, 1, 3, 2147483647} {
			a := &questGemItem54D080{name: "a", kind: uint16(11 + slot), price: price}
			s := &questGemState54D080{types: [3]uint32{11, 12, 13}, head: a}
			questLoseGems54D080(a, s.hooks())
			if !reflect.DeepEqual(s.deleted, []*questGemItem54D080{a}) || !reflect.DeepEqual(s.credits, []int32{price / 2}) {
				t.Fatalf("slot=%d price=%d deleted=%v credits=%v", slot, price, s.deleted, s.credits)
			}
		}
	}
}

func TestQuestLoseGems54D080FreshHeadCachedNextAndLiveTypeCache(t *testing.T) {
	c := &questGemItem54D080{name: "c", kind: 11}
	b := &questGemItem54D080{name: "b", kind: 11, next: c}
	a := &questGemItem54D080{name: "a", kind: 11, next: b}
	z := &questGemItem54D080{name: "z", kind: 99}
	y := &questGemItem54D080{name: "y", kind: 17, next: z}
	x := &questGemItem54D080{name: "x", kind: 11, price: 9, next: y}
	unit := &questGemItem54D080{name: "unit"}
	s := &questGemState54D080{types: [3]uint32{11, 12, 13}, head: a}
	h := s.hooks()
	first, next, loadType, price, deleteItem, addGold := h.first, h.next, h.loadType, h.price, h.delayedDelete, h.addGold
	walk := 0
	h.first = func(u *questGemItem54D080) *questGemItem54D080 {
		walk++
		if walk == 2 {
			s.head = x
		}
		return first(u)
	}
	h.next = func(item *questGemItem54D080) *questGemItem54D080 {
		cached := next(item)
		if item == x {
			x.next = a // The returned next is cached before all later callbacks.
		}
		return cached
	}
	h.loadType = func(slot int) uint32 {
		cached := loadType(slot)
		if walk == 2 && slot == 0 && cached == 11 {
			s.types[0] = 17 // Diamond is loaded before the current type WORD.
		}
		return cached
	}
	h.price = func(item *questGemItem54D080) int32 {
		cached := price(item)
		item.price, item.kind = 999, 13
		return cached
	}
	h.delayedDelete = func(item *questGemItem54D080) {
		deleteItem(item)
		item.next = nil
		s.head = nil
	}
	h.addGold = func(u *questGemItem54D080, amount int32) {
		if u != unit || len(s.deleted) != 1 || s.deleted[0] != x {
			t.Fatal("gold preceded deletion or used a different unit")
		}
		addGold(u, amount)
		s.types[1] = 99
	}
	questLoseGems54D080(unit, h)
	want := []string{
		"cache:0", "first", "cache:0", "type:a", "next:a", "cache:0", "type:b", "next:b", "cache:0", "type:c", "next:c",
		"first", "next:x", "cache:0", "type:x", "price:x", "delete:x", "gold:4",
		"next:y", "cache:0", "type:y", "delete:y", "next:z", "cache:0", "type:z", "cache:1",
	}
	if !reflect.DeepEqual(s.trace, want) || !reflect.DeepEqual(s.deleted, []*questGemItem54D080{x, y}) || !reflect.DeepEqual(s.credits, []int32{4}) {
		t.Fatalf("trace=%v deleted=%v credits=%v, want trace=%v", s.trace, s.deleted, s.credits, want)
	}
}

func TestQuestLoseGems54D080FirstWalkNextIsLive(t *testing.T) {
	skipped := &questGemItem54D080{name: "skipped", kind: 11}
	added := &questGemItem54D080{name: "added", kind: 11}
	a := &questGemItem54D080{name: "a", kind: 11, next: skipped}
	s := &questGemState54D080{types: [3]uint32{11, 12, 13}, head: a}
	h := s.hooks()
	first, index := h.first, h.typeIndex
	walk := 0
	h.first = func(u *questGemItem54D080) *questGemItem54D080 { walk++; return first(u) }
	h.typeIndex = func(item *questGemItem54D080) uint16 {
		if walk == 1 && item == a {
			a.next = added
		}
		return index(item)
	}
	questLoseGems54D080(a, h)
	if !reflect.DeepEqual(s.deleted, []*questGemItem54D080{a}) || len(s.credits) != 0 {
		t.Fatalf("deleted=%v credits=%v", s.deleted, s.credits)
	}
	for _, event := range s.trace {
		if event == "type:skipped" || event == "next:skipped" {
			t.Fatal("first walk next was read before the current type")
		}
	}
}

func TestQuestLoseGems54D080MissingBindingsKeepFaultPrefixes(t *testing.T) {
	countPrefix := []string{"cache:0", "first", "cache:0", "type:a", "next:a"}
	oddPrefix := append(append([]string{}, countPrefix...), "first", "next:a", "cache:0", "type:a")
	for _, name := range []string{"lookup", "store", "first", "type", "count-next", "second-next", "price", "delete", "gold"} {
		t.Run(name, func(t *testing.T) {
			a := &questGemItem54D080{name: "a", kind: 11, price: 9}
			s := &questGemState54D080{types: [3]uint32{11, 12, 13}, resolved: [3]uint32{11, 12, 13}, head: a}
			h := s.hooks()
			var want []string
			switch name {
			case "lookup":
				s.types[0], h.lookupType = 0, nil
				want = []string{"cache:0"}
			case "store":
				s.types[0], h.storeType = 0, nil
				want = []string{"cache:0", "lookup:Diamond"}
			case "first":
				h.first = nil
				want = []string{"cache:0"}
			case "type":
				h.typeIndex = nil
				want = []string{"cache:0", "first", "cache:0"}
			case "count-next":
				h.next = nil
				want = []string{"cache:0", "first", "cache:0", "type:a"}
			case "second-next":
				first, calls := h.first, 0
				h.first = func(u *questGemItem54D080) *questGemItem54D080 {
					calls++
					return first(u)
				}
				// The generic helper receives its hooks by value. Use a live
				// next closure to fault at the second walk, before any type read.
				next := h.next
				h.next = func(item *questGemItem54D080) *questGemItem54D080 {
					if calls == 2 {
						var missing func(*questGemItem54D080) *questGemItem54D080
						return missing(item)
					}
					return next(item)
				}
				want = append(append([]string{}, countPrefix...), "first")
			case "price":
				h.price = nil
				want = oddPrefix
			case "delete":
				h.delayedDelete = nil
				want = append(append([]string{}, oddPrefix...), "price:a")
			case "gold":
				h.addGold = nil
				want = append(append([]string{}, oddPrefix...), "price:a", "delete:a")
			}
			defer func() {
				if recover() == nil {
					t.Fatal("required missing binding did not fault")
				}
				if !reflect.DeepEqual(s.trace, want) {
					t.Fatalf("trace=%v, want fault prefix %v", s.trace, want)
				}
			}()
			questLoseGems54D080(a, h)
		})
	}
}

func TestQuestLoseGems54D080UnusedOddBindingsAreNotRead(t *testing.T) {
	b := &questGemItem54D080{name: "b", kind: 11}
	a := &questGemItem54D080{name: "a", kind: 11, next: b}
	s := &questGemState54D080{types: [3]uint32{11, 12, 13}, head: a}
	h := s.hooks()
	h.lookupType, h.price, h.addGold = nil, nil, nil
	questLoseGems54D080(a, h)
	if !reflect.DeepEqual(s.deleted, []*questGemItem54D080{a}) || len(s.credits) != 0 {
		t.Fatalf("even loss used odd-price bindings: deleted=%v credits=%v", s.deleted, s.credits)
	}
	a.kind, b.kind, s.deleted = 999, 999, nil
	h.delayedDelete = nil
	questLoseGems54D080(a, h)
	if len(s.deleted) != 0 {
		t.Fatal("non-gems used deletion binding")
	}
}
