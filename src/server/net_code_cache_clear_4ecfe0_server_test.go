package server

import "testing"

func TestNetCodeCacheClearNative4ECFE0MovesAllEntriesAndPreservesObjects(t *testing.T) {
	var cache objectNetCodeCache4ECCB0
	obj1 := &Object{NetCode: 1}
	obj2 := &Object{NetCode: 2}
	obj3 := &Object{NetCode: 3}
	freeObj := &Object{NetCode: 4}
	used1 := &cache.entries[0]
	used2 := &cache.entries[1]
	used3 := &cache.entries[2]
	free := &cache.entries[3]

	used1.object = obj1
	used1.next = used2
	used2.object = obj2
	used2.prev = used1
	used2.next = used3
	used3.object = obj3
	used3.prev = used2
	free.object = freeObj
	cache.used.first = used1
	cache.used.last = used3
	cache.free.first = free
	cache.free.last = free
	cache.initialized = true

	got := cache.clear()
	if got.kind != netCodeCacheClearEntry4ECFE0 || got.entry != used3 {
		t.Fatalf("result = %#v, want final prepend entry %p", got, used3)
	}
	if cache.used.first != nil || cache.used.last != nil {
		t.Fatalf("used list = %p/%p, want empty", cache.used.first, cache.used.last)
	}
	if cache.free.first != used3 || cache.free.last != free {
		t.Fatalf("free list = %p/%p, want %p/%p", cache.free.first, cache.free.last, used3, free)
	}
	if used3.prev != nil || used3.next != used2 || used2.prev != used3 || used2.next != used1 || used1.prev != used2 || used1.next != free || free.prev != used1 {
		t.Fatalf("free links are not reverse used order followed by old free head")
	}
	if used1.object != obj1 || used2.object != obj2 || used3.object != obj3 || free.object != freeObj {
		t.Fatalf("object pointers changed: %p %p %p %p", used1.object, used2.object, used3.object, free.object)
	}
}

func TestNetCodeCacheClearNative4ECFE0InitialAndEmpty(t *testing.T) {
	var cache objectNetCodeCache4ECCB0
	obj := &Object{NetCode: 1}
	used := &cache.entries[0]
	free := &cache.entries[1]
	used.object = obj
	cache.used.first = used
	cache.used.last = used
	cache.free.first = free
	cache.free.last = free

	got := cache.clear()
	if got.kind != netCodeCacheClearInitial4ECFE0 || got.initialFlag != 1 {
		t.Fatalf("uninitialized result = %#v, want raw flag 1", got)
	}
	if cache.used.first != used || cache.used.last != used || cache.free.first != free || cache.free.last != free || used.object != obj {
		t.Fatalf("uninitialized cache changed")
	}

	cache.initialized = true
	cache.used = objectNetCodeCacheList4ECD90{}
	got = cache.clear()
	if got.kind != netCodeCacheClearInitial4ECFE0 || got.initialFlag != 0 {
		t.Fatalf("empty result = %#v, want initial zero", got)
	}
	if cache.free.first != free || cache.free.last != free {
		t.Fatalf("free list changed on empty clear: %p/%p", cache.free.first, cache.free.last)
	}
}
