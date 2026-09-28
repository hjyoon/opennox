//go:build !server

package legacy

import (
	"math"
	"testing"
	"unsafe"
)

func TestTransientRayRegistry49BDD0PreservesNativeAddresses(t *testing.T) {
	clientTransientRayClear49BDD0()
	defer clientTransientRayClear49BDD0()

	base := uintptr(0x10000)
	if unsafe.Sizeof(uintptr(0)) > 4 {
		base = uintptr(math.MaxUint32) + 0x10001
	}
	for i := 0; i < 96; i++ {
		addr := base + uintptr(i)*0x1000
		if !clientTransientRayAddAddress49BDD0(addr) {
			t.Fatalf("add ray %d at %#x failed", i, addr)
		}
	}
	if got := clientTransientRayCount49BDD0(); got != 96 {
		t.Fatalf("ray count = %d, want 96", got)
	}
	for _, index := range []int{0, 47, 95} {
		want := base + uintptr(index)*0x1000
		if got := clientTransientRayAddress49BDD0(index); got != want {
			t.Fatalf("ray %d address = %#x, want %#x", index, got, want)
		}
		if !clientTransientRayContainsAddress49BDD0(want) {
			t.Fatalf("registry does not contain ray %d at %#x", index, want)
		}
	}
	if clientTransientRayAddAddress49BDD0(base + 96*0x1000) {
		t.Fatal("registry accepted a 97th ray")
	}
	if clientTransientRayAddress49BDD0(96) != 0 {
		t.Fatal("out-of-range registry lookup returned a pointer")
	}

	clientTransientRayClear49BDD0()
	if got := clientTransientRayCount49BDD0(); got != 0 {
		t.Fatalf("ray count after clear = %d, want 0", got)
	}
	if clientTransientRayContainsAddress49BDD0(base) {
		t.Fatal("cleared registry retained the first ray")
	}
}

func TestTransientRayPayload49BDD0UsesNativeDrawableUnion(t *testing.T) {
	if !clientTransientRayPayloadUsesNativeUnion49BDD0() {
		t.Fatal("ray payload used the PE32 byte offset instead of the native drawable union")
	}
}

func TestCharmOrbPayload499490UsesNativeDrawableUnion(t *testing.T) {
	if !clientOrbPayloadUsesNativeUnion499490() {
		t.Fatal("charm orb payload used the PE32 byte offset or corrupted the drawable links")
	}
}

func TestManaBombOrbPayload499520UsesNativeDrawableUnion(t *testing.T) {
	if !clientManaBombPayloadUsesNativeUnion499520() {
		t.Fatal("mana bomb orb payload used the PE32 byte offset or corrupted the drawable links")
	}
}

func TestBallisticFXPayload499610UsesNativeDrawableUnion(t *testing.T) {
	if !clientBallisticFXPayloadUsesNativeUnion499610() {
		t.Fatal("ballistic FX payload used the PE32 byte offset or corrupted the drawable links")
	}
}

func TestFallingSparkPayload499950UsesNativeDrawableUnion(t *testing.T) {
	if !clientFallingSparkPayloadUsesNativeUnion499950() {
		t.Fatal("falling spark payload used the PE32 byte offset or corrupted the drawable links")
	}
}
