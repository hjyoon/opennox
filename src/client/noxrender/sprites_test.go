package noxrender

import (
	"testing"
	"unsafe"
)

func TestRenderSpritesPackedImageKey(t *testing.T) {
	const key uint32 = 0x80000020
	img := &Image{}
	b := &RenderSprites{byHandle: map[uintptr]*Image{uintptr(key): img}}
	if got := b.ImageByLegacyKey(key); got != img {
		t.Fatalf("packed key %#x resolved to %p, want %p", key, got, img)
	}
	if got := b.ImageByLegacyKey(0); got != nil {
		t.Fatalf("zero key resolved to %p", got)
	}
}

func TestRenderSpritesLegacyHandleRecovery(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) <= 4 {
		t.Skip("legacy aliases are native handles on 32-bit targets")
	}

	var token byte
	h := ImageHandle(unsafe.Pointer(&token))
	img := &Image{}
	b := &RenderSprites{
		byHandle:       map[uintptr]*Image{uintptr(h): img},
		byLegacyHandle: make(map[uint32]ImageHandle),
	}
	b.registerLegacyHandle(h)

	low := uint32(uintptr(h))
	zeroExtended := uintptr(low)
	signExtended := uintptr(uint64(int64(int32(low))))
	if got := b.ImageByLegacyKey(low); got != img {
		t.Fatalf("packed alias %#x resolved to %p, want %p", low, got, img)
	}
	if got := b.imageByLegacyHandle(zeroExtended); got != img {
		t.Fatalf("zero-extended alias %#x resolved to %p, want %p", zeroExtended, got, img)
	}
	if got := b.imageByLegacyHandle(signExtended); got != img {
		t.Fatalf("sign-extended alias %#x resolved to %p, want %p", signExtended, got, img)
	}

	wrongHigh := uintptr(uint64(0x12345678)<<32 | uint64(low))
	if got := b.imageByLegacyHandle(wrongHigh); got != nil {
		t.Fatalf("non-legacy address %#x resolved to %p", wrongHigh, got)
	}
	if got := b.imageByLegacyHandle(uintptr(low + 1)); got != nil {
		t.Fatalf("unknown alias %#x resolved to %p", low+1, got)
	}

	// A collision cannot occur in the current 16 MiB handle arena, but fail
	// closed if a future allocator ever produces one.
	b.byLegacyHandle[low] = nil
	if got := b.imageByLegacyHandle(zeroExtended); got != nil {
		t.Fatalf("ambiguous alias %#x resolved to %p", zeroExtended, got)
	}
}
