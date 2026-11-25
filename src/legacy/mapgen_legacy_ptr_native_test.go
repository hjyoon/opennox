package legacy

import (
	"testing"
	"unsafe"
)

func TestMapgenPointerRegistryTokens(t *testing.T) {
	r := mapgenPointerRegistry{nextToken: 0xf0000000}
	const address = uintptr(0x1234)
	if token := r.register(address); token != uint32(address) || r.register(address) != token {
		t.Fatalf("original token %#x", token)
	}
	if r.register(0) != 0 || r.resolve(0) != 0 {
		t.Fatal("null address changed")
	}
	if unsafe.Sizeof(uintptr(0)) > 4 {
		upper := uint64(1) << 32
		other := uintptr(upper + uint64(address))
		zeroLow := uintptr(upper)
		if token := r.register(other); token != 0xf0000000 || r.resolve(uintptr(token)) != other || r.resolve(other) != other {
			t.Fatalf("collision token %#x", token)
		}
		if token := r.register(zeroLow); token != 0xf0000001 || r.resolve(uintptr(token)) != zeroLow {
			t.Fatalf("zero-word token %#x", token)
		}
		r.forget(other)
		if r.resolve(0xf0000000) != 0xf0000000 || r.resolve(address) != address {
			t.Fatal("forget did not release only the collided token")
		}
		if token := r.register(other); token != 0xf0000002 {
			t.Fatalf("reused a stale token %#x", token)
		}
	}
	r.forget(address)
	if r.register(address) != uint32(address) {
		t.Fatal("released low token was not available")
	}
}
