package server

import "testing"

func TestNetCodeCacheNativeInit4ECE50(t *testing.T) {
	dirtyObject := &Object{NetCode: 0xffffffff}
	dirtyEntry := &objectNetCodeCacheEntry4ECD90{object: dirtyObject}
	cache := objectNetCodeCache4ECCB0{
		free: objectNetCodeCacheList4ECD90{first: dirtyEntry, last: dirtyEntry},
		used: objectNetCodeCacheList4ECD90{first: dirtyEntry, last: dirtyEntry},
	}
	for i := range cache.entries {
		cache.entries[i] = objectNetCodeCacheEntry4ECD90{
			object: dirtyObject,
			prev:   dirtyEntry,
			next:   dirtyEntry,
		}
	}

	got := cache.init()
	wantHead := &cache.entries[objectNetCodeCacheCapacity4ECCB0-1]
	wantTail := &cache.entries[0]
	if got != wantHead {
		t.Fatalf("result = %p, want final prepend result %p", got, wantHead)
	}
	if cache.used.first != nil || cache.used.last != nil {
		t.Fatalf("used list = (%p, %p), want empty", cache.used.first, cache.used.last)
	}
	if cache.free.first != wantHead || cache.free.last != wantTail {
		t.Fatalf("free list = (%p, %p), want (%p, %p)", cache.free.first, cache.free.last, wantHead, wantTail)
	}
	if !cache.initialized {
		t.Fatal("needs-initialization flag was not cleared")
	}

	for i := range cache.entries {
		entry := &cache.entries[i]
		var wantPrev, wantNext *objectNetCodeCacheEntry4ECD90
		if i+1 < len(cache.entries) {
			wantPrev = &cache.entries[i+1]
		}
		if i > 0 {
			wantNext = &cache.entries[i-1]
		}
		if entry.prev != wantPrev || entry.next != wantNext {
			t.Errorf("entry[%d] links = (%p, %p), want (%p, %p)", i, entry.prev, entry.next, wantPrev, wantNext)
		}
		if entry.object != dirtyObject {
			t.Errorf("entry[%d] object = %p, want preserved %p", i, entry.object, dirtyObject)
		}
	}
}
