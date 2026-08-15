package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

func withRespawnRemove4EC6A0Class(t *testing.T, count int, fn func(*alloc.Class, []*respawnRecord4EC5E0, []*server.Object)) {
	t.Helper()
	handles.Init()
	t.Cleanup(handles.Release)

	oldAllocator := respawnAddLoadAllocator4EC5E0()
	oldHead := respawnAddLoadHead4EC5E0()
	class := alloc.NewClass("RespawnRemove4EC6A0Test", unsafe.Sizeof(respawnRecord4EC5E0{}), count+1)
	objects := make([]*server.Object, count)
	freeObjects := make([]func(), count)
	records := make([]*respawnRecord4EC5E0, count)
	for i := range records {
		objects[i], freeObjects[i] = alloc.New(server.Object{})
		records[i] = (*respawnRecord4EC5E0)(class.NewObject())
		records[i].Object = objects[i]
	}
	defer func() {
		respawnAddStoreHead4EC5E0(oldHead)
		respawnAddStoreAllocator4EC5E0(oldAllocator)
		for _, freeObject := range freeObjects {
			freeObject()
		}
		class.Free()
	}()
	respawnAddStoreAllocator4EC5E0(class.UPtr())
	fn(class, records, objects)
}

func TestNoxXxxRespawnRemove4EC6A0HeadFastPathIgnoresPrev(t *testing.T) {
	withRespawnRemove4EC6A0Class(t, 3, func(class *alloc.Class, records []*respawnRecord4EC5E0, objects []*server.Object) {
		head, next, sentinel := records[0], records[1], records[2]
		head.Next = next
		head.Prev = sentinel
		next.Prev = head
		sentinel.Next = head
		respawnAddStoreHead4EC5E0(head)

		Nox_xxx_respawnRemove_4EC6A0(objects[0])

		if got := respawnAddLoadHead4EC5E0(); got != next {
			t.Fatalf("head = %p, want %p", got, next)
		}
		if next.Prev != nil {
			t.Fatalf("new head prev = %p, want nil", next.Prev)
		}
		if sentinel.Next != head {
			t.Fatalf("ignored predecessor next = %p, want original head %p", sentinel.Next, head)
		}
		if reused := (*respawnRecord4EC5E0)(class.NewObject()); reused != head {
			t.Fatalf("freed record reuse = %p, want %p", reused, head)
		}
	})
}

func TestNoxXxxRespawnRemove4EC6A0InteriorRecord(t *testing.T) {
	withRespawnRemove4EC6A0Class(t, 3, func(class *alloc.Class, records []*respawnRecord4EC5E0, objects []*server.Object) {
		head, middle, tail := records[0], records[1], records[2]
		head.Next = middle
		middle.Prev = head
		middle.Next = tail
		tail.Prev = middle
		respawnAddStoreHead4EC5E0(head)

		Nox_xxx_respawnRemove_4EC6A0(objects[1])

		if got := respawnAddLoadHead4EC5E0(); got != head {
			t.Fatalf("head = %p, want %p", got, head)
		}
		if head.Next != tail || tail.Prev != head {
			t.Fatalf("repaired links = head.next:%p tail.prev:%p, want %p/%p", head.Next, tail.Prev, tail, head)
		}
		if reused := (*respawnRecord4EC5E0)(class.NewObject()); reused != middle {
			t.Fatalf("freed record reuse = %p, want %p", reused, middle)
		}
	})
}

func TestNoxXxxRespawnRemove4EC6A0MissingRecord(t *testing.T) {
	withRespawnRemove4EC6A0Class(t, 3, func(class *alloc.Class, records []*respawnRecord4EC5E0, objects []*server.Object) {
		head, middle := records[0], records[1]
		head.Next = middle
		middle.Prev = head
		respawnAddStoreHead4EC5E0(head)

		Nox_xxx_respawnRemove_4EC6A0(objects[2])

		if got := respawnAddLoadHead4EC5E0(); got != head {
			t.Fatalf("head = %p, want %p", got, head)
		}
		if head.Next != middle || middle.Prev != head {
			t.Fatalf("links changed = head.next:%p middle.prev:%p", head.Next, middle.Prev)
		}
	})
}

func TestNoxXxxRespawnRemove4EC6A0EmptyHeadFaults(t *testing.T) {
	withRespawnRemove4EC6A0Class(t, 1, func(_ *alloc.Class, _ []*respawnRecord4EC5E0, objects []*server.Object) {
		respawnAddStoreHead4EC5E0(nil)
		defer func() {
			if recover() == nil {
				t.Fatal("empty head did not fault before the zero check")
			}
		}()
		Nox_xxx_respawnRemove_4EC6A0(objects[0])
	})
}
