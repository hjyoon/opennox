package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

func TestRigidCollisionNativeCoefficientOwnership(t *testing.T) {
	push, spring := collisionPushScalePtr550F80(), collisionSpringPtr5488B0()
	if unsafe.Sizeof(uintptr(0)) > 4 && (uintptr(unsafe.Pointer(push)) <= math.MaxUint32 || uintptr(unsafe.Pointer(spring)) <= math.MaxUint32) {
		t.Fatal("C-owned coefficients below 4 GiB")
	}
	oldPush, oldSpring := *push, *spring
	t.Cleanup(func() { *push = oldPush; *spring = oldSpring })
	if math.Float32bits(oldPush) != 0x42200000 || math.Float32bits(oldSpring) != 0x42c80000 {
		t.Fatal("stock collision coefficients do not match original PE data")
	}
	// Inspect the retired bytes directly while the safe build guards live access.
	retired := memmap.BlobByAddr(0x587000).Data[292488:292496]
	before := [8]byte(retired)
	t.Logf("C-owned push=%08x spring=%08x retired slots=%x", math.Float32bits(oldPush), math.Float32bits(oldSpring), before)
	*push = 37.5
	*spring = 77.25
	if collisionPushScaleNative550F80() != 37.5 || collisionSpringNative5488B0() != 77.25 {
		t.Fatal("coefficient providers did not reload live C-owned values")
	}
	if got := [8]byte(retired); got != before {
		t.Fatalf("retired slots changed from %x to %x", before, got)
	}
}
