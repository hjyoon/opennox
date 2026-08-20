package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestGlyphInventoryCount4EF6F0NativeLayout(t *testing.T) {
	wantSize := uintptr(780)
	wantType := uintptr(4)
	wantNext := uintptr(496)
	wantFirst := uintptr(504)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantSize = 928
		wantType = 8
		wantNext = 528
		wantFirst = 544
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantSize},
		{"Object.TypeInd", unsafe.Offsetof(Object{}.TypeInd), wantType},
		{"Object.TypeInd size", unsafe.Sizeof(Object{}.TypeInd), 2},
		{"Object.InvNextItem", unsafe.Offsetof(Object{}.InvNextItem), wantNext},
		{"Object.InvFirstItem", unsafe.Offsetof(Object{}.InvFirstItem), wantFirst},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestGlyphInventoryCountNative4EF6F0CountsDestroyedMatches(t *testing.T) {
	fourth := &Object{TypeInd: 7}
	third := &Object{TypeInd: 6, InvNextItem: fourth}
	second := &Object{TypeInd: 7, ObjFlags: object.FlagDestroyed, InvNextItem: third}
	first := &Object{TypeInd: 7, InvNextItem: second}
	owner := &Object{InvFirstItem: first}
	cache := uint32(0)
	lookups := 0

	got := glyphInventoryCountNative4EF6F0(owner, glyphInventoryCountNativeDeps4EF6F0{
		loadCache: func() uint32 { return cache },
		lookupType: func(name string) uint32 {
			lookups++
			if name != glyphInventoryCountName4EF6F0 {
				t.Fatalf("lookup name = %q", name)
			}
			return 7
		},
		storeCache: func(value uint32) { cache = value },
	})
	if got != 3 || cache != 7 || lookups != 1 {
		t.Fatalf("result/cache/lookups = %d/%#x/%d, want 3/7/1", got, cache, lookups)
	}
	if first.InvNextItem != second || second.InvNextItem != third || third.InvNextItem != fourth || fourth.InvNextItem != nil {
		t.Fatal("native count mutated the inventory chain")
	}
}

func TestGlyphInventoryCountServerCache4EF6F0IsDedicatedAndFixedWidth(t *testing.T) {
	s := &Server{}
	s.Types.fast.glyph = 0x1234
	deps := glyphInventoryCountServerDeps4EF6F0(s)
	if got := deps.loadCache(); got != 0 {
		t.Fatalf("entry respawn cache = %#x, want 0", got)
	}
	deps.storeCache(0xfedcba98)
	if got := deps.loadCache(); got != 0xfedcba98 {
		t.Fatalf("respawn cache = %#x", got)
	}
	if s.Types.fast.glyph != 0x1234 {
		t.Fatalf("general Glyph cache changed = %#x", s.Types.fast.glyph)
	}
	if unsafe.Sizeof(s.Types.fast.playerRespawnGlyph) != 4 {
		t.Fatalf("respawn cache size = %d, want 4", unsafe.Sizeof(s.Types.fast.playerRespawnGlyph))
	}
}

func TestGlyphInventoryCountNative4EF6F0NilOwnerFaultsAfterCache(t *testing.T) {
	cacheLoads := 0
	defer func() {
		if recover() == nil {
			t.Fatal("nil owner did not fault")
		}
		if cacheLoads != 1 {
			t.Fatalf("cache loads before fault = %d, want 1", cacheLoads)
		}
	}()
	glyphInventoryCountNative4EF6F0(nil, glyphInventoryCountNativeDeps4EF6F0{
		loadCache: func() uint32 {
			cacheLoads++
			return 7
		},
		lookupType: func(string) uint32 { return 0 },
		storeCache: func(uint32) {},
	})
}
