package legacy

import "testing"

type questShopCacheLegacyServer510E20 struct {
	Server
	index   int
	calls   int
	handled bool
}

func (s *questShopCacheLegacyServer510E20) ClearQuestShopSessionNative510E20(playerIndex int) bool {
	s.index = playerIndex
	s.calls++
	return s.handled
}

func TestQuestShopCacheEntry510E20UsesNativeRegistry(t *testing.T) {
	fake := &questShopCacheLegacyServer510E20{handled: true}
	oldGetServer := GetServer
	GetServer = func() Server { return fake }
	t.Cleanup(func() { GetServer = oldGetServer })

	Sub_510E20(7)
	if fake.calls != 1 || fake.index != 7 {
		t.Fatalf("native Quest cache clear = calls:%d index:%d, want 1/7", fake.calls, fake.index)
	}
}

func TestQuestShopCacheEntry510E20RejectsMissingNativeSlotOn64Bit(t *testing.T) {
	if cgoABIPointerSize == 4 {
		t.Skip("the PE32 fallback is valid on 32-bit hosts")
	}
	fake := &questShopCacheLegacyServer510E20{}
	oldGetServer := GetServer
	GetServer = func() Server { return fake }
	t.Cleanup(func() { GetServer = oldGetServer })

	// A missing native entry must return before the C fallback interprets the
	// fixed uint32 slot as a host pointer.
	Sub_510E20(11)
	if fake.calls != 1 || fake.index != 11 {
		t.Fatalf("missing Quest cache clear = calls:%d index:%d, want 1/11", fake.calls, fake.index)
	}
}
