package server

import (
	"reflect"
	"testing"
)

func TestQuestLoseGems54D080NativeSeparateCachePublicationBeforeUnitRead(t *testing.T) {
	// Diamond is a standalone C variable, whereas Emerald/Ruby are blob
	// locations. The wrapper must not assume the three addresses are adjacent.
	cache := []uint32{0, 0xaabbccdd, 91, 0x11223344, 92, 0x55667788}
	types := [3]*uint32{&cache[0], &cache[2], &cache[4]}
	var names []string
	defer func() {
		if recover() == nil {
			t.Fatal("nil unit was normalized instead of faulting after cache publication")
		}
		if !reflect.DeepEqual(names, []string{"Diamond", "Emerald", "Ruby"}) ||
			!reflect.DeepEqual(cache, []uint32{0x1000b, 0xaabbccdd, 0x8000, 0x11223344, 0xffff, 0x55667788}) {
			t.Fatalf("names=%v cache=%x", names, cache)
		}
	}()
	questLoseGemsNative54D080(nil, types, questLoseGemsNativeDeps54D080{lookupType: func(name string) uint32 {
		if name == "Emerald" && cache[0] != 0x1000b || name == "Ruby" && cache[2] != 0x8000 {
			t.Fatal("type was not published before the next lookup")
		}
		names = append(names, name)
		return [...]uint32{0x1000b, 0x8000, 0xffff}[len(names)-1]
	}})
}

func TestQuestLoseGems54D080NativeConditionalCacheReads(t *testing.T) {
	diamond := uint32(11)
	types := [3]*uint32{&diamond, nil, nil}
	b := &Object{TypeInd: 11}
	a := &Object{TypeInd: 11, InvNextItem: b}
	u := &Object{InvFirstItem: a}
	var deleted []*Object
	deps := questLoseGemsNativeDeps54D080{delayedDelete: func(item *Object) { deleted = append(deleted, item) }}
	questLoseGemsNative54D080(u, types, deps)
	if !reflect.DeepEqual(deleted, []*Object{a}) {
		t.Fatalf("deleted=%v", deleted)
	}
	a.TypeInd = 12
	defer func() {
		if recover() == nil {
			t.Fatal("required Emerald cache read was normalized")
		}
		if len(deleted) != 1 {
			t.Fatal("faulting counting walk deleted an item")
		}
	}()
	questLoseGemsNative54D080(u, types, deps)
}

func TestQuestLoseGems54D080NativeModeOnePriceAndCachedNext(t *testing.T) {
	var s Server // No RNG, unit class or player binding enters this helper.
	cache := [3]uint32{11, 12, 13}
	types := [3]*uint32{&cache[0], &cache[1], &cache[2]}
	ruby := &Object{TypeInd: 13, Worth: 0xffffffff}
	emerald := &Object{TypeInd: 12, Worth: 7, InvNextItem: ruby}
	diamond := &Object{TypeInd: 11, Worth: 1001, InvNextItem: emerald}
	u := &Object{InvFirstItem: diamond, Worth: 0xfedcba98}
	beforeD, beforeE, beforeR := *diamond, *emerald, *ruby
	var deleted []*Object
	var credits []int32
	s.QuestLoseGems54D080(u, types, func(item *Object) {
		deleted = append(deleted, item)
		item.InvNextItem = nil
		u.InvFirstItem = nil // Neither mutation may replace the cached next.
	}, func(unit *Object, amount int32) {
		if unit != u || len(deleted) != len(credits)+1 {
			t.Fatal("gold credit did not follow its deletion on the original unit")
		}
		credits = append(credits, amount)
	})
	if !reflect.DeepEqual(deleted, []*Object{diamond, emerald, ruby}) ||
		!reflect.DeepEqual(credits, []int32{500, 3, -1073741824}) || cache != [3]uint32{11, 12, 13} {
		t.Fatalf("deleted=%v credits=%v cache=%v", deleted, credits, cache)
	}
	beforeD.InvNextItem, beforeE.InvNextItem = nil, nil
	if *diamond != beforeD || *emerald != beforeE || *ruby != beforeR || u.Worth != 0xfedcba98 || u.UpdateData != nil {
		t.Fatal("helper changed native fields outside explicit callback mutations")
	}
}

func TestQuestLoseGems54D080NativeMissingCallbacks(t *testing.T) {
	for _, name := range []string{"price", "delete", "gold"} {
		t.Run(name, func(t *testing.T) {
			cache := [3]uint32{11, 12, 13}
			types := [3]*uint32{&cache[0], &cache[1], &cache[2]}
			a := &Object{TypeInd: 11}
			u := &Object{InvFirstItem: a}
			var trace []string
			deps := questLoseGemsNativeDeps54D080{
				price: func(item *Object) int32 {
					if item != a {
						t.Fatal("price received a different native item")
					}
					trace = append(trace, "price")
					return -3
				},
				delayedDelete: func(item *Object) {
					if item != a {
						t.Fatal("deletion received a different native item")
					}
					trace = append(trace, "delete")
				},
				addGold: func(*Object, int32) { trace = append(trace, "gold") },
			}
			var want []string
			switch name {
			case "price":
				deps.price = nil
			case "delete":
				deps.delayedDelete, want = nil, []string{"price"}
			case "gold":
				deps.addGold, want = nil, []string{"price", "delete"}
			}
			defer func() {
				if recover() == nil {
					t.Fatal("required missing callback was normalized")
				}
				if !reflect.DeepEqual(trace, want) {
					t.Fatalf("trace=%v, want %v", trace, want)
				}
			}()
			questLoseGemsNative54D080(u, types, deps)
		})
	}
}
