package legacy

import (
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/internal/cfixture"
)

func TestItemAmountNativeLayout(t *testing.T) {
	attrsSize, attrsField, drawableModifiers, drawableFieldLow, drawableFieldHigh, callbackSize := cfixture.ItemAmountNativeLayout()
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
	root, freeRoot := alloc.NewHigh(nox_window{})
	t.Cleanup(freeRoot)
	item, freeItem := alloc.NewHigh(nox_drawable{})
	t.Cleanup(freeItem)
	image, freeImage := alloc.NewHigh(byte(0))
	t.Cleanup(freeImage)
	accept, cancel := cfixture.ItemAmountCallbacks()
	saved := itemAmountDialog
	t.Cleanup(func() { itemAmountDialog = saved })

	itemAmountDialog.root, itemAmountDialog.drawable = root, item
	itemAmountDialog.images[7] = (*nox_video_bag_image_t)(unsafe.Pointer(image))
	itemAmountDialog.accept, itemAmountDialog.cancel = itemAmountCallback(accept), itemAmountCallback(cancel)
	runtime.GC()

	slots := []struct {
		name      string
		got, want unsafe.Pointer
	}{
		{"dialog", Get_nox_gui_itemAmount_dialog_1319228().C(), unsafe.Pointer(root)},
		{"item", ClientItemAmountDrawable().C(), unsafe.Pointer(item)},
		{"image", unsafe.Pointer(itemAmountDialog.images[7]), unsafe.Pointer(image)},
		{"accept callback", unsafe.Pointer(itemAmountDialog.accept), accept},
		{"cancel callback", unsafe.Pointer(itemAmountDialog.cancel), cancel},
	}
	for i, slot := range slots {
		if slot.got != slot.want {
			t.Errorf("%s pointer = %p, want %p", slot.name, slot.got, slot.want)
		}
		if i < 3 && unsafe.Sizeof(uintptr(0)) == 8 && uint64(uintptr(slot.got)) <= 0xffffffff {
			t.Errorf("%s pointer %p does not exercise the upper address bits", slot.name, slot.got)
		}
	}
}

func TestItemAmountModifierCopyPreservesNativePointers(t *testing.T) {
	modifiers, free := alloc.NewHigh([4]byte{})
	defer free()
	if !cfixture.ItemAmountAttrsContract(modifiers, 0x89ABCDEF) {
		t.Fatal("four modifier pointers or trailing PE32 field were not copied exactly")
	}
}

func TestItemAmountCallbackPreservesGameEXEArgumentOrder(t *testing.T) {
	accept, cancel := cfixture.ItemAmountCallbacks()
	if !cfixture.ItemAmountCallbackContract(accept, 1) || !cfixture.ItemAmountCallbackContract(cancel, 2) {
		t.Fatal("accept or cancel did not receive position/item ID/thing type/amount/extra in GAME.EXE order")
	}
}
