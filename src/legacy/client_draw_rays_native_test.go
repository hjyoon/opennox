//go:build !server

package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/internal/cfixture"
)

func TestTransientRayRegistry49BDD0PreservesNativeAddresses(t *testing.T) {
	pool, free := alloc.NewHigh([97]nox_drawable{})
	defer free()
	defer clientTransientRayClear49BDD0()

	if len(transientRays.drawables) != len(pool)-1 {
		t.Fatalf("ray capacity = %d, want 96", len(transientRays.drawables))
	}
	if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(pool)) <= math.MaxUint32 {
		t.Fatal("drawable pool must be above 4 GiB")
	}
	if !cfixture.ClientTransientRayRegistry49BDD0(unsafe.Pointer(pool)) {
		t.Fatal("C registry did not round-trip and clear the native drawable pointers")
	}
}

func TestTransientRayPayload49BDD0UsesNativeDrawableUnion(t *testing.T) {
	if !cfixture.ClientTransientRayPayloadUsesNativeUnion49BDD0() {
		t.Fatal("ray payload used the PE32 byte offset instead of the native drawable union")
	}
}

func TestCharmOrbPayload499490UsesNativeDrawableUnion(t *testing.T) {
	if !cfixture.ClientOrbPayloadUsesNativeUnion499490() {
		t.Fatal("charm orb payload used the PE32 byte offset or corrupted the drawable links")
	}
}

func TestManaBombOrbPayload499520UsesNativeDrawableUnion(t *testing.T) {
	if !cfixture.ClientManaBombPayloadUsesNativeUnion499520() {
		t.Fatal("mana bomb orb payload used the PE32 byte offset or corrupted the drawable links")
	}
}

func TestBallisticFXPayload499610UsesNativeDrawableUnion(t *testing.T) {
	if !cfixture.ClientBallisticFXPayloadUsesNativeUnion499610() {
		t.Fatal("ballistic FX payload used the PE32 byte offset or corrupted the drawable links")
	}
}

func TestFallingSparkPayload499950UsesNativeDrawableUnion(t *testing.T) {
	if !cfixture.ClientFallingSparkPayloadUsesNativeUnion499950() {
		t.Fatal("falling spark payload used the PE32 byte offset or corrupted the drawable links")
	}
}
