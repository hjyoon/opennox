package legacy

import (
	"testing"
	"unsafe"
)

func TestItemAmountNativeLayout(t *testing.T) {
	attrsSize, attrsField, drawableModifiers, drawableFieldLow, drawableFieldHigh, callbackSize := itemAmountNativeLayout()
	word := unsafe.Sizeof(uintptr(0))
	if callbackSize != word {
		t.Fatalf("callback slot size = %d, want native pointer size %d", callbackSize, word)
	}
	if word == 4 {
		if attrsSize != 20 || attrsField != 16 || drawableModifiers != 432 || drawableFieldLow != 448 || drawableFieldHigh != 450 {
			t.Fatalf("PE32 layout = attrs %d/%d, drawable %d/%d/%d", attrsSize, attrsField, drawableModifiers, drawableFieldLow, drawableFieldHigh)
		}
		return
	}
	if attrsSize != 40 || attrsField != 32 || drawableModifiers != 560 || drawableFieldLow != 592 || drawableFieldHigh != 594 {
		t.Fatalf("native 64-bit layout = attrs %d/%d, drawable %d/%d/%d", attrsSize, attrsField, drawableModifiers, drawableFieldLow, drawableFieldHigh)
	}
}

func TestItemAmountPointerSlotsPreserveNativeWidth(t *testing.T) {
	want := uintptr(0x12345)
	if unsafe.Sizeof(uintptr(0)) > 4 {
		high := uint64(1)
		high <<= 40
		want += uintptr(high)
	}
	for slot, name := range []string{"dialog", "item", "image", "accept callback", "cancel callback"} {
		if got := itemAmountPointerRoundTrip(slot, want); got != want {
			t.Errorf("%s pointer round trip = %#x, want %#x", name, got, want)
		}
	}
}

func TestItemAmountModifierCopyPreservesNativePointers(t *testing.T) {
	base := uintptr(0x23456)
	if unsafe.Sizeof(uintptr(0)) > 4 {
		high := uint64(1)
		high <<= 40
		base += uintptr(high)
	}
	if !itemAmountAttrsContract(base, 0x89ABCDEF) {
		t.Fatal("four modifier pointers or trailing PE32 field were not copied exactly")
	}
}

func TestItemAmountCallbackPreservesGameEXEArgumentOrder(t *testing.T) {
	if !itemAmountCallbackContract() {
		t.Fatal("callback did not receive position/item ID/thing type/amount/extra in GAME.EXE order")
	}
}
