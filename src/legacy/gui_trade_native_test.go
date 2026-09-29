package legacy

import (
	"math"
	"testing"
	"unsafe"
)

func TestGUITradeSlotsPreserveNativePointers(t *testing.T) {
	value := uintptr(0x76543210)
	if unsafe.Sizeof(value) == 8 {
		value = uintptr(0x1234567887654321)
		if value <= math.MaxUint32 {
			t.Fatalf("test pointer = %#x, want value above 4 GiB", value)
		}
	}
	if !guiTradePointerRoundTrip(value) {
		t.Fatalf("trade GUI truncated native pointer %#x", value)
	}
}

func TestGUITradeSlotScalarContract(t *testing.T) {
	if !guiTradeSlotContract() {
		t.Fatal("trade GUI slot item lookup, removal, or compatibility contract failed")
	}
	wantSize := uintptr(140)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantSize = 144
	}
	size, totalOffset := guiTradeSlotNativeLayout()
	if size != wantSize {
		t.Fatalf("trade slot size = %d, want %d", size, wantSize)
	}
	wantTotalOffset := unsafe.Sizeof(uintptr(0)) + 33*4
	if totalOffset != wantTotalOffset {
		t.Fatalf("trade total offset = %d, want %d", totalOffset, wantTotalOffset)
	}
}

func TestGUITradeSlotSelectionPreservesVisualOrder(t *testing.T) {
	if !guiTradeSlotSelectionOrder() {
		t.Fatal("trade GUI slot selection did not preserve the original 0, 2, 1, 3 visual order")
	}
}
