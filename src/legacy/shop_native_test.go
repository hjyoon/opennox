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

func TestShopViewportUsesNativeWidth(t *testing.T) {
	if !shopViewportNativeContract() {
		t.Fatal("shop viewport fields do not preserve the GAME.EXE 1024x768 contract")
	}
}

func TestInventoryCapacityPreservesNativeDrawablePointer(t *testing.T) {
	available, highPointer := inventoryCapacityContract(1234, 1234, 0, 1, 1)
	if unsafe.Sizeof(uintptr(0)) > 4 && !highPointer {
		t.Fatal("test drawable did not exercise a pointer above the PE32 address range")
	}
	if available != 80 {
		t.Fatalf("available slots with a matching stack = %d, want 80", available)
	}
}

func TestInventoryCapacityPreservesGameEXEStackLimits(t *testing.T) {
	tests := []struct {
		name          string
		requestedType uint32
		drawableType  uint32
		flags         uint32
		count         uint8
		amount        int
		want          int
	}{
		{name: "nonmatching occupied cell", requestedType: 1234, drawableType: 1235, count: 1, amount: 1, want: 79},
		{name: "regular stack at capacity", requestedType: 1234, drawableType: 1234, count: 30, amount: 1, want: 80},
		{name: "regular stack overflow", requestedType: 1234, drawableType: 1234, count: 31, amount: 1, want: 79},
		{name: "food stack at capacity", requestedType: 1234, drawableType: 1234, flags: 0x10, count: 2, amount: 1, want: 80},
		{name: "food stack overflow", requestedType: 1234, drawableType: 1234, flags: 0x10, count: 3, amount: 1, want: 79},
		{name: "not stackable", requestedType: 1234, drawableType: 1234, flags: 0x4000000, count: 1, amount: 1, want: 79},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			available, _ := inventoryCapacityContract(
				tc.requestedType, tc.drawableType, tc.flags, tc.count, tc.amount,
			)
			if available != tc.want {
				t.Fatalf("available slots = %d, want %d", available, tc.want)
			}
		})
	}
}

func TestInventoryItemStateUsesNativeGrid(t *testing.T) {
	found, count, currentHealth, maximumHealth := inventoryItemStateContract()
	if !found || count != 3 || currentHealth != 0x2345 || maximumHealth != 0x6789 {
		t.Fatalf("inventory state = found:%t count:%d health:%#x/%#x", found, count, currentHealth, maximumHealth)
	}
}

func TestInventoryItemLocationUsesNativeGrid(t *testing.T) {
	found, column, row, netCode := inventoryItemLocationContract()
	if !found || column != 2 || row != 19 || netCode != 0x89abcdef {
		t.Fatalf("inventory location = found:%t column:%d row:%d netcode:%#x", found, column, row, netCode)
	}
}

func TestInventoryTooltipPreservesNativeDrawablePointer(t *testing.T) {
	if !inventoryTooltipPointerContract() {
		t.Fatal("inventory tooltip drawable did not survive a native-width round trip")
	}
}

func TestInventoryTradeSelectsLastNativeNetCode(t *testing.T) {
	first, second := inventoryTradeNetCodeContract()
	if first != 0x11112222 || second != 0x33334444 {
		t.Fatalf("selected net codes = %#x/%#x, want %#x/%#x", first, second, 0x11112222, 0x33334444)
	}
}
