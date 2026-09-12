package noxgroupwalk

import (
	"reflect"
	"testing"
)

type testGroup struct {
	kind uint8
	list *testItem
}

type testItem struct {
	first, second uint32
	next          *testItem
}

func testHooks(groups map[uint32]*testGroup, visits *[]uint32) Hooks[*testGroup, *testItem] {
	return Hooks[*testGroup, *testItem]{
		Kind:         func(g *testGroup) uint8 { return g.kind },
		First:        func(g *testGroup) *testItem { return g.list },
		Next:         func(it *testItem) *testItem { return it.next },
		IDs:          func(it *testItem) (uint32, uint32) { return it.first, it.second },
		ResolveGroup: func(id uint32) *testGroup { return groups[id] },
		Visit: func(kind uint8, first, second uint32) {
			*visits = append(*visits, uint32(kind)<<24|first)
		},
	}
}

func TestWalkKindsAndRecursion(t *testing.T) {
	obj := &testGroup{kind: 0, list: &testItem{first: 11, next: &testItem{first: 12}}}
	wp := &testGroup{kind: 1, list: &testItem{first: 21}}
	wall := &testGroup{kind: 2, list: &testItem{first: 31, second: 32}}
	root := &testGroup{kind: 3, list: &testItem{first: 1, next: &testItem{first: 2, next: &testItem{first: 3}}}}
	groups := map[uint32]*testGroup{1: obj, 2: wp, 3: wall}
	for _, tc := range []struct {
		expected int32
		want     []uint32
	}{
		{0, []uint32{11, 12}},
		{1, []uint32{1<<24 | 21}},
		{2, []uint32{2<<24 | 31}},
		{3, nil},
	} {
		var got []uint32
		Walk(root, tc.expected, testHooks(groups, &got))
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("expected %d: got %v, want %v", tc.expected, got, tc.want)
		}
	}
}

func TestWalkWallFallthroughAndLiveList(t *testing.T) {
	child := &testGroup{kind: 2, list: &testItem{first: 43}}
	first := &testItem{first: 42}
	group := &testGroup{kind: 2, list: first}
	var got []uint32
	h := testHooks(map[uint32]*testGroup{42: child}, &got)
	h.Visit = func(kind uint8, id, _ uint32) {
		got = append(got, id)
		if id == 42 {
			first.next = &testItem{first: 99}
		}
	}
	Walk(group, 2, h)
	if want := []uint32{42, 99, 43}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestWalkSkipsNilUnknownAndMismatched(t *testing.T) {
	var got []uint32
	h := testHooks(nil, &got)
	Walk((*testGroup)(nil), 0, h)
	Walk(&testGroup{kind: 4, list: &testItem{first: 1}}, 0, h)
	Walk(&testGroup{kind: 0, list: &testItem{first: 2}}, 1, h)
	Walk(&testGroup{kind: 2, list: &testItem{first: 3}}, 0, h)
	Walk(&testGroup{kind: 3, list: &testItem{first: 5}}, 0, h)
	if len(got) != 0 {
		t.Fatalf("unexpected visits: %v", got)
	}
}
