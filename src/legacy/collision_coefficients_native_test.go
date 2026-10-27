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
	t.Logf("C-owned push=%08x spring=%08x retired slots push=%08x spring=%08x", math.Float32bits(oldPush), math.Float32bits(oldSpring), math.Float32bits(memmap.Float32(0x587000, 292488)), math.Float32bits(memmap.Float32(0x587000, 292492)))
	*push = 37.5
	*spring = 77.25
	if collisionPushScaleNative550F80() != 37.5 || collisionSpringNative5488B0() != 77.25 {
		t.Fatal("coefficient providers did not reload live C-owned values")
	}
	if memmap.Float32(0x587000, 292488) == 37.5 || memmap.Float32(0x587000, 292492) == 77.25 {
		t.Fatal("retired slots unexpectedly alias C-owned coefficients")
	}
}
