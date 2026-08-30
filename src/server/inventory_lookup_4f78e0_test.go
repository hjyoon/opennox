package server

import (
	"reflect"
	"testing"
)

type inventoryLookupTestObject4F78E0 struct {
	name   string
	holder *inventoryLookupTestObject4F78E0
	first  *inventoryLookupTestObject4F78E0
	next   *inventoryLookupTestObject4F78E0
	code   uint32
}

func inventoryLookupName4F78E0(obj *inventoryLookupTestObject4F78E0) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func TestInventoryContains4F78E0RejectsHolderBeforeHead(t *testing.T) {
	holder := &inventoryLookupTestObject4F78E0{name: "holder"}
	other := &inventoryLookupTestObject4F78E0{name: "other"}
	item := &inventoryLookupTestObject4F78E0{name: "item", holder: other}
	var events []string
	got := inventoryContains4F78E0(holder, item, inventoryContainsHooks4F78E0[*inventoryLookupTestObject4F78E0]{
		loadItemHolder: func(obj *inventoryLookupTestObject4F78E0) *inventoryLookupTestObject4F78E0 {
			events = append(events, "holder:"+obj.name)
			return obj.holder
		},
		loadHolderFirst: func(*inventoryLookupTestObject4F78E0) *inventoryLookupTestObject4F78E0 {
			t.Fatal("holder mismatch loaded inventory head")
			return nil
		},
		loadItemNext: func(*inventoryLookupTestObject4F78E0) *inventoryLookupTestObject4F78E0 {
			t.Fatal("holder mismatch loaded an item link")
			return nil
		},
	})
	if got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	if want := []string{"holder:item"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestInventoryContains4F78E0ExactIdentityAndLiveNext(t *testing.T) {
	holder := &inventoryLookupTestObject4F78E0{name: "holder"}
	first := &inventoryLookupTestObject4F78E0{name: "first"}
	equalLooking := &inventoryLookupTestObject4F78E0{name: "item", holder: holder}
	target := &inventoryLookupTestObject4F78E0{name: "item", holder: holder}
	holder.first = first
	first.next = equalLooking
	var events []string
	got := inventoryContains4F78E0(holder, target, inventoryContainsHooks4F78E0[*inventoryLookupTestObject4F78E0]{
		loadItemHolder: func(obj *inventoryLookupTestObject4F78E0) *inventoryLookupTestObject4F78E0 {
			events = append(events, "holder:"+obj.name)
			return obj.holder
		},
		loadHolderFirst: func(obj *inventoryLookupTestObject4F78E0) *inventoryLookupTestObject4F78E0 {
			events = append(events, "first:"+obj.name)
			return obj.first
		},
		loadItemNext: func(obj *inventoryLookupTestObject4F78E0) *inventoryLookupTestObject4F78E0 {
			events = append(events, "next:"+obj.name)
			if obj == first {
				// The original reloads the link after the identity comparison.
				return target
			}
			t.Fatal("matched identity loaded its next link")
			return nil
		},
	})
	if got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := []string{"holder:item", "first:holder", "next:first"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestInventoryContains4F78E0EmptyAndMiss(t *testing.T) {
	for _, tc := range []struct {
		name string
		head *inventoryLookupTestObject4F78E0
		want []string
	}{
		{name: "empty", want: []string{"holder:item", "first:holder"}},
		{
			name: "miss",
			head: &inventoryLookupTestObject4F78E0{name: "first", next: &inventoryLookupTestObject4F78E0{name: "second"}},
			want: []string{"holder:item", "first:holder", "next:first", "next:second"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			holder := &inventoryLookupTestObject4F78E0{name: "holder", first: tc.head}
			item := &inventoryLookupTestObject4F78E0{name: "item", holder: holder}
			var events []string
			got := inventoryContains4F78E0(holder, item, inventoryContainsHooks4F78E0[*inventoryLookupTestObject4F78E0]{
				loadItemHolder: func(obj *inventoryLookupTestObject4F78E0) *inventoryLookupTestObject4F78E0 {
					events = append(events, "holder:"+obj.name)
					return obj.holder
				},
				loadHolderFirst: func(obj *inventoryLookupTestObject4F78E0) *inventoryLookupTestObject4F78E0 {
					events = append(events, "first:"+obj.name)
					return obj.first
				},
				loadItemNext: func(obj *inventoryLookupTestObject4F78E0) *inventoryLookupTestObject4F78E0 {
					events = append(events, "next:"+obj.name)
					return obj.next
				},
			})
			if got != 0 {
				t.Fatalf("result = %d, want 0", got)
			}
			if !reflect.DeepEqual(events, tc.want) {
				t.Fatalf("events = %v, want %v", events, tc.want)
			}
		})
	}
}

func TestInventoryContains4F78E0FaultPrefix(t *testing.T) {
	const fault = "next fault"
	holder := &inventoryLookupTestObject4F78E0{name: "holder"}
	item := &inventoryLookupTestObject4F78E0{name: "item", holder: holder}
	holder.first = &inventoryLookupTestObject4F78E0{name: "first"}
	var events []string
	defer func() {
		if got := recover(); got != fault {
			t.Fatalf("panic = %v, want %q", got, fault)
		}
		want := []string{"holder", "first", "next"}
		if !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}()
	inventoryContains4F78E0(holder, item, inventoryContainsHooks4F78E0[*inventoryLookupTestObject4F78E0]{
		loadItemHolder: func(obj *inventoryLookupTestObject4F78E0) *inventoryLookupTestObject4F78E0 {
			events = append(events, "holder")
			return obj.holder
		},
		loadHolderFirst: func(obj *inventoryLookupTestObject4F78E0) *inventoryLookupTestObject4F78E0 {
			events = append(events, "first")
			return obj.first
		},
		loadItemNext: func(*inventoryLookupTestObject4F78E0) *inventoryLookupTestObject4F78E0 {
			events = append(events, "next")
			panic(fault)
		},
	})
}

func TestEquippedItemByCode4F7920FirstMatchAndLiveNext(t *testing.T) {
	holder := &inventoryLookupTestObject4F78E0{name: "holder"}
	first := &inventoryLookupTestObject4F78E0{name: "first", code: 1}
	stale := &inventoryLookupTestObject4F78E0{name: "stale", code: 7}
	matched := &inventoryLookupTestObject4F78E0{name: "matched", code: 7}
	duplicate := &inventoryLookupTestObject4F78E0{name: "duplicate", code: 7}
	holder.first = first
	first.next = stale
	matched.next = duplicate
	var events []string
	got := equippedItemByCode4F7920(holder, uint32(7), equippedItemByCodeHooks4F7920[*inventoryLookupTestObject4F78E0]{
		loadHolderFirst: func(obj *inventoryLookupTestObject4F78E0) *inventoryLookupTestObject4F78E0 {
			events = append(events, "first:"+obj.name)
			return obj.first
		},
		loadItemNetCode: func(obj *inventoryLookupTestObject4F78E0) uint32 {
			events = append(events, "code:"+obj.name)
			return obj.code
		},
		loadItemNext: func(obj *inventoryLookupTestObject4F78E0) *inventoryLookupTestObject4F78E0 {
			events = append(events, "next:"+obj.name)
			if obj == first {
				return matched
			}
			t.Fatal("first matching item loaded its next link")
			return nil
		},
	})
	if got != matched {
		t.Fatalf("result = %p (%s), want %p (%s)", got, inventoryLookupName4F78E0(got), matched, matched.name)
	}
	want := []string{"first:holder", "code:first", "next:first", "code:matched"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestEquippedItemByCode4F7920EmptyAndMiss(t *testing.T) {
	for _, tc := range []struct {
		name string
		head *inventoryLookupTestObject4F78E0
		want []string
	}{
		{name: "empty", want: []string{"first:holder"}},
		{
			name: "miss",
			head: &inventoryLookupTestObject4F78E0{name: "one", code: 1, next: &inventoryLookupTestObject4F78E0{name: "two", code: 2}},
			want: []string{"first:holder", "code:one", "next:one", "code:two", "next:two"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			holder := &inventoryLookupTestObject4F78E0{name: "holder", first: tc.head}
			var events []string
			got := equippedItemByCode4F7920(holder, uint32(9), equippedItemByCodeHooks4F7920[*inventoryLookupTestObject4F78E0]{
				loadHolderFirst: func(obj *inventoryLookupTestObject4F78E0) *inventoryLookupTestObject4F78E0 {
					events = append(events, "first:"+obj.name)
					return obj.first
				},
				loadItemNetCode: func(obj *inventoryLookupTestObject4F78E0) uint32 {
					events = append(events, "code:"+obj.name)
					return obj.code
				},
				loadItemNext: func(obj *inventoryLookupTestObject4F78E0) *inventoryLookupTestObject4F78E0 {
					events = append(events, "next:"+obj.name)
					return obj.next
				},
			})
			if got != nil {
				t.Fatalf("result = %p, want nil", got)
			}
			if !reflect.DeepEqual(events, tc.want) {
				t.Fatalf("events = %v, want %v", events, tc.want)
			}
		})
	}
}

func TestEquippedItemByCode4F7920FaultPrefix(t *testing.T) {
	const fault = "code fault"
	holder := &inventoryLookupTestObject4F78E0{name: "holder", first: &inventoryLookupTestObject4F78E0{name: "item"}}
	var events []string
	defer func() {
		if got := recover(); got != fault {
			t.Fatalf("panic = %v, want %q", got, fault)
		}
		want := []string{"first", "code"}
		if !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}()
	equippedItemByCode4F7920(holder, uint32(3), equippedItemByCodeHooks4F7920[*inventoryLookupTestObject4F78E0]{
		loadHolderFirst: func(obj *inventoryLookupTestObject4F78E0) *inventoryLookupTestObject4F78E0 {
			events = append(events, "first")
			return obj.first
		},
		loadItemNetCode: func(*inventoryLookupTestObject4F78E0) uint32 {
			events = append(events, "code")
			panic(fault)
		},
		loadItemNext: func(*inventoryLookupTestObject4F78E0) *inventoryLookupTestObject4F78E0 {
			t.Fatal("code fault loaded next")
			return nil
		},
	})
}
