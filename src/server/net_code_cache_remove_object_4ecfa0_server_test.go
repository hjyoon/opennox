package server

import "testing"

func TestNetCodeCacheRemoveObjectNative4ECFA0MovesExactEntry(t *testing.T) {
	var cache objectNetCodeCache4ECCB0
	obj1 := &Object{NetCode: 1}
	obj2 := &Object{NetCode: 2}
	used1 := &cache.entries[0]
	used2 := &cache.entries[1]
	free := &cache.entries[2]

	used1.object = obj1
	used1.next = used2
	used2.object = obj2
	used2.prev = used1
	cache.used.first = used1
	cache.used.last = used2
	cache.free.first = free
	cache.free.last = free
	cache.initialized = true

	got := cache.remove(obj2)
	if got.kind != netCodeCacheRemoveObjectEntry4ECFA0 || got.entry != used2 {
		t.Fatalf("result = %#v, want entry %p", got, used2)
	}
	if cache.used.first != used1 || cache.used.last != used1 || used1.next != nil {
		t.Fatalf("used list = first %p last %p next %p, want sole %p", cache.used.first, cache.used.last, used1.next, used1)
	}
	if cache.free.first != used2 || cache.free.last != free {
		t.Fatalf("free list = first %p last %p, want %p/%p", cache.free.first, cache.free.last, used2, free)
	}
	if used2.prev != nil || used2.next != free || free.prev != used2 {
		t.Fatalf("free links = used2 prev %p next %p, old head prev %p", used2.prev, used2.next, free.prev)
	}
	if used2.object != obj2 {
		t.Fatalf("moved entry object = %p, want preserved %p", used2.object, obj2)
	}
}

func TestNetCodeCacheRemoveObjectNative4ECFA0MissPreservesLists(t *testing.T) {
	var cache objectNetCodeCache4ECCB0
	obj := &Object{NetCode: 1}
	missing := &Object{NetCode: 2}
	entry := &cache.entries[0]
	entry.object = obj
	cache.used.first = entry
	cache.used.last = entry
	cache.initialized = true

	got := cache.remove(missing)
	if got.kind != netCodeCacheRemoveObjectArgument4ECFA0 || got.object != missing {
		t.Fatalf("result = %#v, want missing object %p", got, missing)
	}
	if cache.used.first != entry || cache.used.last != entry || cache.free.first != nil || cache.free.last != nil {
		t.Fatalf("lists changed on miss: used %p/%p free %p/%p", cache.used.first, cache.used.last, cache.free.first, cache.free.last)
	}
	if entry.object != obj || entry.prev != nil || entry.next != nil {
		t.Fatalf("entry changed on miss: %#v", entry)
	}
}

func TestNetCodeCacheRemoveObjectNative4ECFA0InitialAndEmpty(t *testing.T) {
	var cache objectNetCodeCache4ECCB0
	obj := &Object{NetCode: 1}
	entry := &cache.entries[0]
	entry.object = obj
	cache.used.first = entry
	cache.used.last = entry

	got := cache.remove(obj)
	if got.kind != netCodeCacheRemoveObjectInitial4ECFA0 || got.initialFlag != 1 {
		t.Fatalf("uninitialized result = %#v, want initial flag 1", got)
	}
	if cache.used.first != entry || cache.used.last != entry {
		t.Fatalf("uninitialized lists changed: %p/%p", cache.used.first, cache.used.last)
	}

	cache.initialized = true
	cache.used = objectNetCodeCacheList4ECD90{}
	got = cache.remove(obj)
	if got.kind != netCodeCacheRemoveObjectInitial4ECFA0 || got.initialFlag != 0 {
		t.Fatalf("empty result = %#v, want initial zero", got)
	}
}

func TestNetCodeCacheRemoveObjectNative4ECFA0MatchesNilObject(t *testing.T) {
	var cache objectNetCodeCache4ECCB0
	entry := &cache.entries[0]
	cache.used.first = entry
	cache.used.last = entry
	cache.initialized = true

	got := cache.remove(nil)
	if got.kind != netCodeCacheRemoveObjectEntry4ECFA0 || got.entry != entry {
		t.Fatalf("result = %#v, want nil-object entry %p", got, entry)
	}
	if cache.used.first != nil || cache.used.last != nil || cache.free.first != entry || cache.free.last != entry {
		t.Fatalf("lists after nil match: used %p/%p free %p/%p", cache.used.first, cache.used.last, cache.free.first, cache.free.last)
	}
	if entry.object != nil {
		t.Fatalf("nil-object entry changed to %p", entry.object)
	}
}
