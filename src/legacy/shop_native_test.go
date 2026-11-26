package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/internal/cfixture"
)

func TestShopCellNativeLayout(t *testing.T) {
	size, countOffset, codesOffset, priceOffset := cfixture.ShopCellNativeLayout()
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
		high := uint64(1)
		high <<= 32
		want += uintptr(high)
	}
	if got := cfixture.ShopLookupRoundTrip(want); got != want {
		t.Fatalf("drawable pointer round trip = %#x, want %#x", got, want)
	}
}

func TestShopLookupPreservesGameEXETraversalOrder(t *testing.T) {
	if got := cfixture.ShopLookupOrder(); got != 0x1111 {
		t.Fatalf("first matching drawable = %#x, want row 0/column 1 drawable %#x", got, uintptr(0x1111))
	}
}

func TestShopLookupUsesCountAndStableStorage(t *testing.T) {
	if !cfixture.ShopStorageContract() {
		t.Fatal("shop cells must keep their column-major addresses and use count to admit lookup")
	}
}

func TestShopNetCodeShiftPreservesGameEXEContract(t *testing.T) {
	got := cfixture.ShopShiftContract()
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

func TestShopStockFindsSpaceAfterFullStack(t *testing.T) {
	drawable, free := alloc.New(client.Drawable{})
	defer free()
	const typeID = 81
	dr := (*nox_drawable)(drawable.C())
	dr.field_27 = typeID
	var saved [shopRows][shopCols]shopInventoryCell
	for row := range shopRows {
		for col := range shopCols {
			cell := shopCell(row, col)
			saved[row][col], *cell = *cell, shopInventoryCell{}
		}
	}
	defer func() {
		for row := range shopRows {
			for col := range shopCols {
				*shopCell(row, col) = saved[row][col]
			}
		}
	}()

	full, empty, partial := shopCell(0, 0), shopCell(0, 1), shopCell(0, 2)
	full.drawable, full.count, full.price = dr, shopStackMax, 11
	partial.drawable, partial.count, partial.net_codes[0] = dr, 1, 32
	full.net_codes[shopStackMax-1] = 31
	if got := sub_479300(typeID, 33, 17, 0, nil); got != 2 {
		t.Fatalf("stack size after insertion = %d, want 2", got)
	}
	if full.count != shopStackMax || full.price != 11 || sub_4780A0(31) != full {
		t.Fatal("insertion changed the full stack")
	}
	if sub_4780A0(33) != partial || partial.price != 17 || empty.count != 0 {
		t.Fatal("insertion did not reuse the matching stack with room")
	}
	partial.count = shopStackMax
	if got := sub_4793C0(typeID); got != empty {
		t.Fatalf("cell after both stacks fill = %p, want %p", got, empty)
	}
}

func TestShopViewportUsesNativeWidth(t *testing.T) {
	if !cfixture.ShopViewportNativeContract() {
		t.Fatal("shop viewport fields do not preserve the GAME.EXE 1024x768 contract")
	}
}

func TestInventoryCapacityPreservesNativeDrawablePointer(t *testing.T) {
	available, highPointer := cfixture.InventoryCapacityContract(1234, 1234, 0, 1, 1)
	if unsafe.Sizeof(uintptr(0)) > 4 && !highPointer {
		t.Fatal("test drawable did not exercise a pointer above the PE32 address range")
	}
	if available != 80 {
		t.Fatalf("available slots with a matching stack = %d, want 80", available)
	}
}

func TestInventoryTotalWeightPreservesNativeDrawablePointers(t *testing.T) {
	weight, highPointers := cfixture.InventoryTotalWeightContract()
	if unsafe.Sizeof(uintptr(0)) > 4 && !highPointers {
		t.Fatal("test drawables did not exercise pointers above the PE32 address range")
	}
	if weight != 29 {
		t.Fatalf("inventory weight = %d, want 2*7 + 3*5 = 29", weight)
	}
}

func TestInventoryAlternateWeaponSyncPreservesNativeCellPointer(t *testing.T) {
	const want = uint64(1<<10) - 1
	if got := cfixture.InventoryAltWeaponSyncContract(); got != want {
		t.Fatalf("alternate weapon sync contract bits = %#x, want %#x", got, want)
	}
}

func TestInventoryCellNativeLayout(t *testing.T) {
	size, altOffset, countOffset := cfixture.InventoryCellNativeLayout()
	if unsafe.Sizeof(uintptr(0)) == 4 {
		if size != 148 || altOffset != 136 || countOffset != 140 {
			t.Fatalf("PE32 inventory cell = size %d, offsets %d/%d", size, altOffset, countOffset)
		}
		return
	}
	if size != 152 || altOffset != 140 || countOffset != 144 {
		t.Fatalf("native 64-bit inventory cell = size %d, offsets %d/%d", size, altOffset, countOffset)
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
			available, _ := cfixture.InventoryCapacityContract(
				tc.requestedType, tc.drawableType, tc.flags, tc.count, tc.amount,
			)
			if available != tc.want {
				t.Fatalf("available slots = %d, want %d", available, tc.want)
			}
		})
	}
}

func TestInventoryTooltipPreservesNativeDrawablePointer(t *testing.T) {
	if !cfixture.InventoryTooltipPointerContract() {
		t.Fatal("inventory tooltip drawable did not survive a native-width round trip")
	}
}

func TestInventoryTradeSelectsLastNativeNetCode(t *testing.T) {
	first, second := cfixture.InventoryTradeNetCodeContract()
	if first != 0x11112222 || second != 0x33334444 {
		t.Fatalf("selected net codes = %#x/%#x, want %#x/%#x", first, second, 0x11112222, 0x33334444)
	}
}

func newShopTestDrawable(t *testing.T) *nox_drawable {
	t.Helper()
	p, free := alloc.CallocHigh(1, unsafe.Sizeof(nox_drawable{}))
	t.Cleanup(free)
	if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(p) <= 0xffffffff {
		t.Fatal("test drawable did not exercise a high native address")
	}
	return (*nox_drawable)(p)
}

func TestShopItemLocationUsesNativeCellsAndScrollOffset(t *testing.T) {
	drawable := newShopTestDrawable(t)
	t.Cleanup(cfixture.ClearShopCells())
	active, scroll := cfixture.ShopViewWords()
	left := memmap.PtrInt32(0x5D4594, 1098380)
	top := memmap.PtrInt32(0x5D4594, 1098384)
	oldActive, oldScroll, oldLeft, oldTop := *active, *scroll, *left, *top
	t.Cleanup(func() {
		*active, *scroll, *left, *top = oldActive, oldScroll, oldLeft, oldTop
	})
	*active, *scroll, *left, *top = 1, 50, 10, 20

	drawable.field_27 = 0x12345678
	cell := shopCell(7, 5)
	cell.drawable, cell.count, cell.price = drawable, 2, 0xFEDCBA98
	cell.net_codes[0], cell.net_codes[1], cell.net_codes[2] = 123, 0x89ABCDEF, 456
	found, x, y, thingType, count, price := Nox_client_shopItemLocation(0x89ABCDEF)
	if !found || x != 285 || y != 345 || thingType != 0x12345678 || count != 2 || price != 0xFEDCBA98 {
		t.Fatalf("shop location: found %t, position (%d,%d), type %#x, count %d, price %#x", found, x, y, thingType, count, price)
	}
	if drawable.field_32 != 0 || cell.count != 2 || cell.net_codes[1] != 0x89ABCDEF {
		t.Fatal("shop lookup modified the cell or drawable")
	}
	missing := func(netCode uint32) {
		t.Helper()
		if found, _, _, _, _, _ := Nox_client_shopItemLocation(netCode); found {
			t.Fatalf("unavailable shop item %#x was returned", netCode)
		}
	}
	missing(456)
	missing(0)
	*active = 0
	missing(123)
	*active = 1
	cell.count = shopStackMax + 1
	missing(123)
	cell.count = 2
	cell.drawable = nil
	missing(123)
}

func TestInventoryItemStateUsesNativeGrid(t *testing.T) {
	first, second := newShopTestDrawable(t), newShopTestDrawable(t)
	t.Cleanup(cfixture.ClearInventoryCells())
	first.field_27, first.field_73_1, first.field_73_2 = 0x12345678, 0x2345, 0x6789
	second.field_27, second.field_73_1, second.field_73_2 = 0x12345678, 0x1111, 0x2222
	// The row-first lookup sees first even though second precedes it physically.
	firstCell, secondCell := inventoryGridCell(0, 1), inventoryGridCell(1, 0)
	firstCell.field_0, firstCell.field_140 = first, 3
	secondCell.field_0, secondCell.field_140 = second, 2

	found, count, currentHealth, maximumHealth := Nox_client_inventoryItemState(0x12345678)
	if !found || count != 5 || currentHealth != 0x2345 || maximumHealth != 0x6789 {
		t.Fatalf("inventory state: found %t, count %d, current %#x, maximum %#x", found, count, currentHealth, maximumHealth)
	}
	if found, count, currentHealth, maximumHealth := Nox_client_inventoryItemState(0); found || count != 0 || currentHealth != 0 || maximumHealth != 0 {
		t.Fatalf("missing inventory state: found %t, count %d, current %#x, maximum %#x", found, count, currentHealth, maximumHealth)
	}
}

func TestInventoryItemLocationUsesNativeGrid(t *testing.T) {
	drawable := newShopTestDrawable(t)
	t.Cleanup(cfixture.ClearInventoryCells())
	drawable.field_27 = 0x12345678
	cell := inventoryGridCell(19, 2)
	cell.field_0, cell.field_4, cell.field_140 = drawable, 0x89ABCDEF, 1
	found, column, row, netCode := Nox_client_inventoryItemLocation(0x12345678)
	if !found || column != 2 || row != 19 || netCode != 0x89ABCDEF {
		t.Fatalf("inventory location: found %t, column %d, row %d, netcode %#x", found, column, row, netCode)
	}
	if found, column, row, netCode := Nox_client_inventoryItemLocation(0); found || column != -1 || row != -1 || netCode != 0 {
		t.Fatalf("missing item: found %t, column %d, row %d, netcode %#x", found, column, row, netCode)
	}
}
