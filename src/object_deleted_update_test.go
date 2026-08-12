package opennox

import (
	"testing"

	"github.com/opennox/opennox/v1/server"
)

func TestDeletedObjectsUpdate4E5E20Empty(t *testing.T) {
	getCalls := 0
	setCalls := 0
	deletedObjectsUpdate_4E5E20(deletedObjectsUpdate4E5E20Hooks{
		deletedList: func() *server.Object {
			getCalls++
			return nil
		},
		setDeletedList: func(obj *server.Object) {
			setCalls++
			if obj != nil {
				t.Fatalf("empty update published %p, want nil", obj)
			}
		},
		frame: func() uint32 {
			t.Fatal("empty update read frame")
			return 0
		},
		removeFromUpdatable: func(*server.Object) {
			t.Fatal("empty update removed an object from updatable list")
		},
		finish: func(*server.Object) {
			t.Fatal("empty update finalized an object")
		},
	})
	if getCalls != 1 || setCalls != 1 {
		t.Fatalf("deleted-list accesses = get:%d set:%d, want 1 and 1", getCalls, setCalls)
	}
}

func TestDeletedObjectsUpdate4E5E20OrderAndRetainedList(t *testing.T) {
	const frame = 50
	keep1 := &server.Object{DeletedAt: frame}
	finish1 := &server.Object{DeletedAt: frame - 1}
	keep2 := &server.Object{DeletedAt: frame}
	finish2 := &server.Object{DeletedAt: frame + 1}
	keep1.DeletedNext = finish1
	finish1.DeletedNext = keep2
	keep2.DeletedNext = finish2

	head := keep1
	wantOrder := []*server.Object{keep1, finish1, keep2, finish2}
	var gotOrder []*server.Object
	frameCalls := 0
	setCalls := 0
	deletedObjectsUpdate_4E5E20(deletedObjectsUpdate4E5E20Hooks{
		deletedList: func() *server.Object {
			return head
		},
		setDeletedList: func(obj *server.Object) {
			setCalls++
			if len(gotOrder) != len(wantOrder) {
				t.Fatalf("published deleted list after %d callbacks, want %d", len(gotOrder), len(wantOrder))
			}
			head = obj
		},
		frame: func() uint32 {
			frameCalls++
			return frame
		},
		removeFromUpdatable: func(obj *server.Object) {
			if head != keep1 {
				t.Fatal("deleted-list head was published before traversal completed")
			}
			gotOrder = append(gotOrder, obj)
		},
		finish: func(obj *server.Object) {
			if head != keep1 {
				t.Fatal("deleted-list head was published before traversal completed")
			}
			gotOrder = append(gotOrder, obj)
			obj.DeletedNext = nil
		},
	})
	if frameCalls != len(wantOrder) || setCalls != 1 {
		t.Fatalf("calls = frame:%d set:%d, want %d and 1", frameCalls, setCalls, len(wantOrder))
	}
	if len(gotOrder) != len(wantOrder) {
		t.Fatalf("callback count = %d, want %d", len(gotOrder), len(wantOrder))
	}
	for i := range wantOrder {
		if gotOrder[i] != wantOrder[i] {
			t.Fatalf("callback %d = %p, want %p", i, gotOrder[i], wantOrder[i])
		}
	}
	if head != keep2 || head.DeletedNext != keep1 || keep1.DeletedNext != nil {
		t.Fatalf("retained list = %p -> %p -> %p, want keep2 -> keep1 -> nil", head, head.DeletedNext, keep1.DeletedNext)
	}
}

func TestDeletedObjectsUpdate4E5E20ReadOrder(t *testing.T) {
	const frame = 7
	first := &server.Object{DeletedAt: frame}
	originalNext := &server.Object{DeletedAt: frame}
	replacementNext := &server.Object{DeletedAt: frame}
	first.DeletedNext = originalNext

	current := first
	var removed []*server.Object
	deletedObjectsUpdate_4E5E20(deletedObjectsUpdate4E5E20Hooks{
		deletedList: func() *server.Object {
			return first
		},
		setDeletedList: func(*server.Object) {},
		frame: func() uint32 {
			if current == first {
				first.DeletedAt = frame + 1
				first.DeletedNext = replacementNext
			}
			return frame
		},
		removeFromUpdatable: func(obj *server.Object) {
			removed = append(removed, obj)
			current = obj.DeletedNext
		},
		finish: func(obj *server.Object) {
			t.Fatalf("unexpected finalization of %p", obj)
		},
	})
	if len(removed) != 2 {
		t.Fatalf("removed count = %d, want 2", len(removed))
	}
	if removed[0] != first || removed[1] != replacementNext {
		t.Fatalf("removed order = %p then %p, want %p then %p", removed[0], removed[1], first, replacementNext)
	}
	if originalNext.DeletedNext != nil {
		t.Fatal("unvisited original successor was modified")
	}
}
