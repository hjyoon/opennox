package server

import (
	"testing"
	"unsafe"
)

func TestObjectFromNetCodeNativeLayout4ECCB0(t *testing.T) {
	wantFlags := uintptr(16)
	wantNetCode := uintptr(36)
	wantNextItem := uintptr(496)
	wantFirstItem := uintptr(504)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantFlags = 20
		wantNetCode = 40
		wantNextItem = 528
		wantFirstItem = 544
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantFlags},
		{"Object.NetCode", unsafe.Offsetof(Object{}.NetCode), wantNetCode},
		{"Object.InvNextItem", unsafe.Offsetof(Object{}.InvNextItem), wantNextItem},
		{"Object.InvFirstItem", unsafe.Offsetof(Object{}.InvFirstItem), wantFirstItem},
		{"Player.PlayerUnit", unsafe.Offsetof(Player{}.PlayerUnit), 2056},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s offset = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestObjectFromNetCodeNativeSearchAndCacheLifecycle4ECCB0(t *testing.T) {
	s := &Server{}
	owner := &Object{NetCode: 1}
	item := &Object{NetCode: 0x80000009}
	owner.InvFirstItem = item
	s.Objs.List = owner

	if got := s.ObjectFromNetCode4ECCB0(0x80000009); got != item {
		t.Fatalf("inventory result = %p, want %p", got, item)
	}
	item.ObjFlags = 0x20
	s.Objs.List = nil
	if got := s.ObjectFromNetCode4ECCB0(0x80000009); got != item {
		t.Fatalf("cache hit = %p, want dead cached object %p", got, item)
	}

	s.ObjectNetCodeCacheRemove4ECFA0(item)
	pending := &Object{NetCode: 0x80000009}
	s.Objs.Pending = pending
	if got := s.ObjectFromNetCode4ECCB0(0x80000009); got != pending {
		t.Fatalf("pending result = %p, want %p", got, pending)
	}

	s.ObjectNetCodeCacheClear4ECFE0()
	pending.ObjFlags = 0x20
	s.Objs.Pending = nil
	unit := &Object{NetCode: 0x80000009}
	s.Players.list = []Player{{Active: 1, PlayerUnit: unit}}
	if got := s.ObjectFromNetCode4ECCB0(0x80000009); got != unit {
		t.Fatalf("player-unit result = %p, want %p", got, unit)
	}
	if s.Objs.netCodeCache.len != 0 {
		t.Fatalf("player-unit match cache length = %d, want 0", s.Objs.netCodeCache.len)
	}
}

func TestObjectNetCodeCacheCapacityAndRecency4ECCB0(t *testing.T) {
	var cache objectNetCodeCache4ECCB0
	objects := make([]Object, objectNetCodeCacheCapacity4ECCB0+2)
	for i := range objects {
		objects[i].NetCode = uint32(i + 1)
	}
	for i := 0; i < objectNetCodeCacheCapacity4ECCB0+1; i++ {
		cache.add(&objects[i])
	}
	if cache.len != objectNetCodeCacheCapacity4ECCB0 {
		t.Fatalf("cache length = %d, want %d", cache.len, objectNetCodeCacheCapacity4ECCB0)
	}
	if got := cache.lookup(1); got != nil {
		t.Fatalf("evicted code 1 returned %p", got)
	}
	if got := cache.lookup(2); got != &objects[1] {
		t.Fatalf("code 2 lookup = %p, want %p", got, &objects[1])
	}
	cache.add(&objects[objectNetCodeCacheCapacity4ECCB0+1])
	if got := cache.lookup(3); got != nil {
		t.Fatalf("least-recent code 3 returned %p", got)
	}
	if got := cache.lookup(2); got != &objects[1] {
		t.Fatalf("recent code 2 was evicted: %p", got)
	}
	cache.remove(&objects[1])
	if got := cache.lookup(2); got != nil {
		t.Fatalf("removed code 2 returned %p", got)
	}
	cache.clear()
	if cache.len != 0 {
		t.Fatalf("cleared cache length = %d", cache.len)
	}
	for i, obj := range cache.objects {
		if obj != nil {
			t.Fatalf("cleared cache entry %d = %p", i, obj)
		}
	}
}
