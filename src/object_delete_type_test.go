package opennox

import (
	"testing"

	"github.com/opennox/opennox/v1/server"
)

func TestDeleteAllObjectsOfType4E5DB0Empty(t *testing.T) {
	deleteAllObjectsOfType_4E5DB0(nil, 7, func(*server.Object) {
		t.Fatal("empty object list called delayed delete")
	})
}

func TestDeleteAllObjectsOfType4E5DB0OrderAndSelection(t *testing.T) {
	const target = 0x1234
	obj1 := &server.Object{TypeInd: target}
	obj2 := &server.Object{TypeInd: target + 1}
	obj3 := &server.Object{TypeInd: target}
	obj1.ObjNext = obj2
	obj2.ObjNext = obj3

	item11 := &server.Object{TypeInd: target + 1}
	item12 := &server.Object{TypeInd: target}
	item13 := &server.Object{TypeInd: target}
	obj1.InvFirstItem = item11
	item11.InvNextItem = item12
	item12.InvNextItem = item13

	item21 := &server.Object{TypeInd: target}
	item22 := &server.Object{TypeInd: target + 1}
	obj2.InvFirstItem = item21
	item21.InvNextItem = item22

	want := []*server.Object{item12, item13, obj1, item21, obj3}
	var got []*server.Object
	deleteAllObjectsOfType_4E5DB0(obj1, target, func(obj *server.Object) {
		got = append(got, obj)
	})
	if len(got) != len(want) {
		t.Fatalf("delayed delete count = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("delayed delete %d = %p, want %p", i, got[i], want[i])
		}
	}
}

func TestDeleteAllObjectsOfType4E5DB0SavesSuccessors(t *testing.T) {
	const target = 9
	obj1 := &server.Object{TypeInd: target}
	obj2 := &server.Object{TypeInd: target}
	obj1.ObjNext = obj2

	item1 := &server.Object{TypeInd: target}
	item2 := &server.Object{TypeInd: target}
	obj1.InvFirstItem = item1
	item1.InvNextItem = item2

	want := []*server.Object{item1, item2, obj1, obj2}
	var got []*server.Object
	deleteAllObjectsOfType_4E5DB0(obj1, target, func(obj *server.Object) {
		got = append(got, obj)
		obj.InvNextItem = nil
		obj.ObjNext = nil
		if obj == item1 {
			obj1.ObjNext = nil
		}
	})
	if len(got) != len(want) {
		t.Fatalf("delayed delete count after unlink = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("delayed delete after unlink %d = %p, want %p", i, got[i], want[i])
		}
	}
}

func TestDeleteAllObjectsOfType4E5DB0UsesUnsigned16BitType(t *testing.T) {
	obj := &server.Object{TypeInd: 0xffff}
	for _, tc := range []struct {
		name    string
		typeInd int
		want    int
	}{
		{name: "exact", typeInd: 0xffff, want: 1},
		{name: "wider", typeInd: 0x1ffff, want: 0},
		{name: "negative", typeInd: -1, want: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := 0
			deleteAllObjectsOfType_4E5DB0(obj, tc.typeInd, func(*server.Object) {
				got++
			})
			if got != tc.want {
				t.Fatalf("delayed delete count = %d, want %d", got, tc.want)
			}
		})
	}
}
