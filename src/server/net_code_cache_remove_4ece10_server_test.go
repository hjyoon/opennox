package server

import "testing"

func TestObjectNetCodeCacheListRemove4ECE10(t *testing.T) {
	t.Run("middle", func(t *testing.T) {
		first := &objectNetCodeCacheEntry4ECD90{}
		entry := &objectNetCodeCacheEntry4ECD90{prev: first}
		last := &objectNetCodeCacheEntry4ECD90{prev: entry}
		first.next = entry
		entry.next = last
		list := objectNetCodeCacheList4ECD90{first: first, last: last}

		if got := list.remove(entry); got != entry {
			t.Fatalf("result = %p, want removed entry %p", got, entry)
		}
		if list.first != first || list.last != last {
			t.Fatalf("list endpoints = %p/%p, want %p/%p", list.first, list.last, first, last)
		}
		if first.next != last || last.prev != first {
			t.Fatalf("survivor links = %p/%p, want %p/%p", first.next, last.prev, last, first)
		}
		if entry.prev != first || entry.next != last {
			t.Fatalf("removed links were cleared or changed: %p/%p", entry.prev, entry.next)
		}
	})

	t.Run("head", func(t *testing.T) {
		entry := &objectNetCodeCacheEntry4ECD90{}
		last := &objectNetCodeCacheEntry4ECD90{prev: entry}
		entry.next = last
		list := objectNetCodeCacheList4ECD90{first: entry, last: last}

		if got := list.remove(entry); got != last {
			t.Fatalf("result = %p, want successor %p", got, last)
		}
		if list.first != last || list.last != last || last.prev != nil {
			t.Fatalf("list = first %p last %p last.prev %p", list.first, list.last, last.prev)
		}
		if entry.next != last || entry.prev != nil {
			t.Fatalf("removed links = %p/%p, want successor/nil", entry.next, entry.prev)
		}
	})

	t.Run("tail", func(t *testing.T) {
		first := &objectNetCodeCacheEntry4ECD90{}
		entry := &objectNetCodeCacheEntry4ECD90{prev: first}
		first.next = entry
		list := objectNetCodeCacheList4ECD90{first: first, last: entry}

		if got := list.remove(entry); got != entry {
			t.Fatalf("result = %p, want removed entry %p", got, entry)
		}
		if list.first != first || list.last != first || first.next != nil {
			t.Fatalf("list = first %p last %p first.next %p", list.first, list.last, first.next)
		}
		if entry.prev != first || entry.next != nil {
			t.Fatalf("removed links = %p/%p, want predecessor/nil", entry.prev, entry.next)
		}
	})

	t.Run("sole", func(t *testing.T) {
		entry := &objectNetCodeCacheEntry4ECD90{}
		list := objectNetCodeCacheList4ECD90{first: entry, last: entry}

		if got := list.remove(entry); got != nil {
			t.Fatalf("result = %p, want nil successor", got)
		}
		if list.first != nil || list.last != nil {
			t.Fatalf("list endpoints = %p/%p, want nil/nil", list.first, list.last)
		}
		if entry.prev != nil || entry.next != nil {
			t.Fatalf("removed links = %p/%p, want nil/nil", entry.prev, entry.next)
		}
	})
}
