package opennox

import (
	"testing"

	"github.com/opennox/opennox/v1/server"
)

func TestFinalizeDeletingObjects4E5EC0Empty(t *testing.T) {
	getCalls := 0
	setCalls := 0
	finalizeDeletingObjects_4E5EC0(finalizeDeletingObjects4E5EC0Hooks{
		deletedList: func() *server.Object {
			getCalls++
			return nil
		},
		setDeletedList: func(obj *server.Object) {
			setCalls++
			if obj != nil {
				t.Fatalf("empty finalize published %p, want nil", obj)
			}
		},
		finish: func(*server.Object) {
			t.Fatal("empty finalize deleted an object")
		},
	})
	if getCalls != 1 || setCalls != 1 {
		t.Fatalf("deleted-list accesses = get:%d set:%d, want 1 and 1", getCalls, setCalls)
	}
}

func TestFinalizeDeletingObjects4E5EC0OrderAndSavedSuccessors(t *testing.T) {
	first := &server.Object{}
	second := &server.Object{}
	third := &server.Object{}
	first.DeletedNext = second
	second.DeletedNext = third

	head := first
	setCalls := 0
	want := []*server.Object{first, second, third}
	var got []*server.Object
	finalizeDeletingObjects_4E5EC0(finalizeDeletingObjects4E5EC0Hooks{
		deletedList: func() *server.Object {
			return head
		},
		setDeletedList: func(obj *server.Object) {
			setCalls++
			if len(got) != len(want) {
				t.Fatalf("published deleted list after %d callbacks, want %d", len(got), len(want))
			}
			head = obj
		},
		finish: func(obj *server.Object) {
			if setCalls != 0 {
				t.Fatal("deleted-list head was published before traversal completed")
			}
			got = append(got, obj)
			obj.DeletedNext = nil
			if obj == first {
				head = &server.Object{}
			}
		},
	})
	if len(got) != len(want) {
		t.Fatalf("finish count = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("finish %d = %p, want %p", i, got[i], want[i])
		}
	}
	if setCalls != 1 || head != nil {
		t.Fatalf("final deleted list = %p after %d writes, want nil after 1", head, setCalls)
	}
}

func TestFinalizeDeletingObjects4E5EC0ReadsSuccessorPerObject(t *testing.T) {
	first := &server.Object{}
	second := &server.Object{}
	originalThird := &server.Object{}
	replacementThird := &server.Object{}
	first.DeletedNext = second
	second.DeletedNext = originalThird

	var got []*server.Object
	finalizeDeletingObjects_4E5EC0(finalizeDeletingObjects4E5EC0Hooks{
		deletedList: func() *server.Object {
			return first
		},
		setDeletedList: func(*server.Object) {},
		finish: func(obj *server.Object) {
			got = append(got, obj)
			if obj == first {
				second.DeletedNext = replacementThird
			}
		},
	})
	want := []*server.Object{first, second, replacementThird}
	if len(got) != len(want) {
		t.Fatalf("finish count = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("finish %d = %p, want %p", i, got[i], want[i])
		}
	}
	if originalThird.DeletedNext != nil {
		t.Fatal("unvisited original successor was modified")
	}
}
