package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type shopCancelLegacyServer510DC0 struct {
	Server
	got     *server.TradeSession
	calls   int
	handled bool
}

func (s *shopCancelLegacyServer510DC0) ShopCancelSessionNative510DC0(session *server.TradeSession) bool {
	s.got = session
	s.calls++
	return s.handled
}

func TestShopCancelSessionEntry510DC0KeepsNativePointer(t *testing.T) {
	session, free := alloc.New(server.TradeSession{})
	defer free()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(session)) <= math.MaxUint32 {
		t.Fatalf("session pointer = %p, want address above the PE32 range", session)
	}

	fake := &shopCancelLegacyServer510DC0{handled: true}
	oldGetServer := GetServer
	GetServer = func() Server { return fake }
	t.Cleanup(func() { GetServer = oldGetServer })

	Nox_xxx_shopCancelSession_510DC0(session)
	if fake.calls != 1 || fake.got != session {
		t.Fatalf("native cancellation = calls:%d session:%p, want 1/%p", fake.calls, fake.got, session)
	}
}

func TestShopCancelSessionEntry510DC0RejectsUnknownPointerOn64Bit(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("the PE32 fallback is valid on 32-bit hosts")
	}
	session, free := alloc.New(server.TradeSession{})
	defer free()

	fake := &shopCancelLegacyServer510DC0{}
	oldGetServer := GetServer
	GetServer = func() Server { return fake }
	t.Cleanup(func() { GetServer = oldGetServer })

	// Returning false exercises the C-side 64-bit guard. It must not inspect
	// Field16 at the PE32 offset or enter the original PE32 exit/free chain.
	Nox_xxx_shopCancelSession_510DC0(session)
	if fake.calls != 1 || fake.got != session {
		t.Fatalf("unknown cancellation = calls:%d session:%p, want 1/%p", fake.calls, fake.got, session)
	}
}
