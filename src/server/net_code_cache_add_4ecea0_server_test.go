package server

import "testing"

func TestNetCodeCacheAddNativeFreeEntry4ECEA0(t *testing.T) {
	var cache objectNetCodeCache4ECCB0
	cache.init()
	obj := &Object{NetCode: 0x80000001}

	got := cache.add(obj)
	want := &cache.entries[objectNetCodeCacheCapacity4ECCB0-1]
	if got != want {
		t.Fatalf("result = %p, want free head %p", got, want)
	}
	if got.object != obj {
		t.Fatalf("entry object = %p, want %p", got.object, obj)
	}
	if cache.used.first != want || cache.used.last != want {
		t.Fatalf("used list = %p/%p, want %p/%p", cache.used.first, cache.used.last, want, want)
	}
	if got.prev != nil || got.next != nil {
		t.Fatalf("used entry links = %p/%p, want nil/nil", got.prev, got.next)
	}
	wantFreeFirst := &cache.entries[objectNetCodeCacheCapacity4ECCB0-2]
	if cache.free.first != wantFreeFirst {
		t.Fatalf("free first = %p, want %p", cache.free.first, wantFreeFirst)
	}
	if cache.free.last != &cache.entries[0] {
		t.Fatalf("free last = %p, want entry zero %p", cache.free.last, &cache.entries[0])
	}
}

func TestNetCodeCacheAddNativeReusesCachedTail4ECEA0(t *testing.T) {
	var cache objectNetCodeCache4ECCB0
	cache.init()
	objects := make([]Object, objectNetCodeCacheCapacity4ECCB0)
	for i := range objects {
		objects[i].NetCode = uint32(i + 1)
		cache.add(&objects[i])
	}

	oldTail := &cache.entries[objectNetCodeCacheCapacity4ECCB0-1]
	if cache.used.last != oldTail {
		t.Fatalf("used last = %p, want first inserted entry %p", cache.used.last, oldTail)
	}
	if cache.free.first != nil {
		t.Fatalf("free first = %p, want nil", cache.free.first)
	}
	staleFreeLast := cache.free.last
	replacement := &Object{NetCode: 0xffffffff}

	got := cache.add(replacement)
	if got != oldTail {
		t.Fatalf("result = %p, want cached old tail %p", got, oldTail)
	}
	if got.object != replacement {
		t.Fatalf("reused entry object = %p, want %p", got.object, replacement)
	}
	if cache.used.first != oldTail {
		t.Fatalf("used first = %p, want reused tail %p", cache.used.first, oldTail)
	}
	wantLast := &cache.entries[objectNetCodeCacheCapacity4ECCB0-2]
	if cache.used.last != wantLast {
		t.Fatalf("used last = %p, want predecessor %p", cache.used.last, wantLast)
	}
	if oldTail.prev != nil || oldTail.next != &cache.entries[0] {
		t.Fatalf("reused entry links = %p/%p, want nil/%p", oldTail.prev, oldTail.next, &cache.entries[0])
	}
	if cache.entries[0].prev != oldTail {
		t.Fatalf("old head prev = %p, want %p", cache.entries[0].prev, oldTail)
	}
	if cache.usedLen() != objectNetCodeCacheCapacity4ECCB0 {
		t.Fatalf("used length = %d, want %d", cache.usedLen(), objectNetCodeCacheCapacity4ECCB0)
	}
	if cache.free.first != nil || cache.free.last != staleFreeLast {
		t.Fatalf("free list = %p/%p, want nil/stale %p", cache.free.first, cache.free.last, staleFreeLast)
	}
}
