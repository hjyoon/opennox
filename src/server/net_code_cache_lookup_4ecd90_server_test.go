package server

import (
	"testing"
	"unsafe"
)

func TestNetCodeCacheNativeEntryLayout4ECD90(t *testing.T) {
	wantSize := uintptr(12)
	wantPrev := uintptr(4)
	wantNext := uintptr(8)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantSize = 24
		wantPrev = 8
		wantNext = 16
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"entry size", unsafe.Sizeof(objectNetCodeCacheEntry4ECD90{}), wantSize},
		{"entry.object", unsafe.Offsetof(objectNetCodeCacheEntry4ECD90{}.object), 0},
		{"entry.prev", unsafe.Offsetof(objectNetCodeCacheEntry4ECD90{}.prev), wantPrev},
		{"entry.next", unsafe.Offsetof(objectNetCodeCacheEntry4ECD90{}.next), wantNext},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestNetCodeCacheNativeLookupLazyInitAndPromotion4ECD90(t *testing.T) {
	var cache objectNetCodeCache4ECCB0
	if got := cache.lookup(7); got != nil {
		t.Fatalf("empty lookup = %p, want nil", got)
	}
	if !cache.initialized {
		t.Fatal("first lookup did not initialize cache")
	}
	if cache.used.first != nil || cache.used.last != nil {
		t.Fatalf("used list after init = (%p, %p), want empty", cache.used.first, cache.used.last)
	}
	if cache.free.first != &cache.entries[objectNetCodeCacheCapacity4ECCB0-1] {
		t.Fatalf("free head = %p, want last array entry %p", cache.free.first, &cache.entries[objectNetCodeCacheCapacity4ECCB0-1])
	}
	if cache.free.last != &cache.entries[0] {
		t.Fatalf("free tail = %p, want first array entry %p", cache.free.last, &cache.entries[0])
	}

	first := &Object{NetCode: 1}
	matched := &Object{NetCode: 0xffffffff}
	newest := &Object{NetCode: 3}
	cache.add(first)
	cache.add(matched)
	cache.add(newest)
	matchedEntry := cache.used.first.next
	if matchedEntry == nil {
		t.Fatal("middle entry is nil")
	}
	if matchedEntry.object != matched {
		t.Fatalf("middle entry object = %p, want %p", matchedEntry.object, matched)
	}
	if got := cache.lookup(^uint32(0)); got != matched {
		t.Fatalf("lookup = %p, want %p", got, matched)
	}
	if cache.used.first != matchedEntry {
		t.Fatalf("used head = %p, want promoted entry %p", cache.used.first, matchedEntry)
	}
	if matchedEntry.prev != nil || matchedEntry.next == nil || matchedEntry.next.object != newest {
		t.Fatalf("promoted links = prev %p next %p", matchedEntry.prev, matchedEntry.next)
	}
	if cache.used.last == nil || cache.used.last.object != first {
		t.Fatalf("used tail = %p, want first object entry", cache.used.last)
	}
}
