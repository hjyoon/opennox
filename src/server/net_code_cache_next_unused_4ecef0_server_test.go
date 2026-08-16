package server

import "testing"

func TestNetCodeCacheNextUnusedNativeEmptyPreservesTail4ECEF0(t *testing.T) {
	tail := &objectNetCodeCacheEntry4ECD90{}
	cache := objectNetCodeCache4ECCB0{
		free: objectNetCodeCacheList4ECD90{last: tail},
	}

	if got := cache.nextUnused(); got != nil {
		t.Fatalf("result = %p, want nil", got)
	}
	if cache.free.first != nil || cache.free.last != tail {
		t.Fatalf("free list = %p/%p, want nil/%p", cache.free.first, cache.free.last, tail)
	}
}

func TestNetCodeCacheNextUnusedNativePreservesEntryLinksAndTail4ECEF0(t *testing.T) {
	prev := &objectNetCodeCacheEntry4ECD90{}
	next := &objectNetCodeCacheEntry4ECD90{}
	tail := &objectNetCodeCacheEntry4ECD90{}
	obj := &Object{NetCode: 0xffffffff}
	head := &objectNetCodeCacheEntry4ECD90{object: obj, prev: prev, next: next}
	cache := objectNetCodeCache4ECCB0{
		free: objectNetCodeCacheList4ECD90{first: head, last: tail},
	}

	got := cache.nextUnused()
	if got != head {
		t.Fatalf("result = %p, want cached head %p", got, head)
	}
	if cache.free.first != next || cache.free.last != tail {
		t.Fatalf("free list = %p/%p, want %p/%p", cache.free.first, cache.free.last, next, tail)
	}
	if head.object != obj || head.prev != prev || head.next != next {
		t.Fatalf("returned entry = object %p links %p/%p, want %p %p/%p", head.object, head.prev, head.next, obj, prev, next)
	}
}
