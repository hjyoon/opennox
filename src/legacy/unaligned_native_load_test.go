package legacy

import (
	"encoding/binary"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestUnalignedNativeTileAndSecretWallLoads(t *testing.T) {
	bank, release := alloc.NewHigh([24]byte{})
	defer release()
	target, releaseTarget := alloc.NewHigh([1]byte{})
	defer releaseTarget()
	want := unsafe.Pointer(target)
	if unsafe.Sizeof(want) == 8 && uintptr(want) <= 0xffffffff {
		t.Fatalf("fixture pointer %#x fits in 32 bits", uintptr(want))
	}
	storeWallFixturePointerBits(bank[:], 1, want)
	binary.NativeEndian.PutUint32(bank[11:], 0xf1234567)
	binary.NativeEndian.PutUint16(bank[17:], 0x8765)

	pointerSlot := unsafe.Pointer(&bank[1])
	if got := tileListNativeLoadPointer411350(pointerSlot); got != want {
		t.Fatalf("tile pointer = %p, want %p", got, want)
	}
	if got := secretWallNativeLoadPointer410550(pointerSlot); got != want {
		t.Fatalf("secret wall pointer = %p, want %p", got, want)
	}
	if got := tileListNativeLoadWord411350(unsafe.Pointer(&bank[11])); got != 0xf1234567 {
		t.Fatalf("tile word = %#x, want 0xf1234567", got)
	}
	if got := secretWallNativeLoadWord410550(unsafe.Pointer(&bank[17])); got != 0x8765 {
		t.Fatalf("secret wall word = %#x, want 0x8765", got)
	}
}
