package legacy

import (
	"testing"
	"unsafe"
)

func TestShopCellNativeLayout(t *testing.T) {
	size, countOffset, codesOffset, priceOffset := shopCellNativeLayout()
	if unsafe.Sizeof(uintptr(0)) == 4 {
		if size != 140 || countOffset != 4 || codesOffset != 8 || priceOffset != 136 {
			t.Fatalf("PE32 layout = size %d, offsets %d/%d/%d", size, countOffset, codesOffset, priceOffset)
		}
		return
	}
	if size != 144 || countOffset != 8 || codesOffset != 12 || priceOffset != 140 {
		t.Fatalf("native 64-bit layout = size %d, offsets %d/%d/%d", size, countOffset, codesOffset, priceOffset)
	}
}

func TestShopLookupPreservesNativePointer(t *testing.T) {
	want := uintptr(0x12345)
	if unsafe.Sizeof(uintptr(0)) > 4 {
		want += uintptr(1) << 32
	}
	if got := shopLookupRoundTrip(want); got != want {
		t.Fatalf("drawable pointer round trip = %#x, want %#x", got, want)
	}
}

func TestShopLookupPreservesGameEXETraversalOrder(t *testing.T) {
	if got := shopLookupOrder(); got != 0x1111 {
		t.Fatalf("first matching drawable = %#x, want row 0/column 1 drawable %#x", got, uintptr(0x1111))
	}
}

func TestShopNetCodeShiftPreservesGameEXEContract(t *testing.T) {
	got := shopShiftContract()
	index := byte(got)
	code0 := byte(got >> 8)
	code1 := byte(got >> 16)
	code2 := byte(got >> 24)
	code30 := byte(got >> 32)
	code31 := byte(got >> 40)
	if index != 1 || code0 != 10 || code1 != 30 || code2 != 0 || code30 != 99 || code31 != 0 {
		t.Fatalf("shift = index %d, codes %d/%d/%d/.../%d/%d", index, code0, code1, code2, code30, code31)
	}
}
