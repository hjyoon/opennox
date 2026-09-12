package legacy

import (
	"math"
	"testing"
	"unsafe"
)

func TestSliderEventCallbacksReceiveNativeWindowPointer(t *testing.T) {
	for _, tc := range []struct {
		name     string
		vertical bool
	}{
		{name: "horizontal"},
		{name: "vertical", vertical: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			address, ok := sliderEventCallbackContract(tc.vertical)
			if !ok {
				t.Fatal("callback did not preserve its window and widget-data pointers")
			}
			if unsafe.Sizeof(uintptr(0)) > 4 && address <= math.MaxUint32 {
				t.Fatalf("test window address = %#x, want a pointer above the PE32 range", address)
			}
		})
	}
}

func TestSliderWindowPointerSlotsPreserveNativeWidth(t *testing.T) {
	want := uintptr(0x45678)
	if unsafe.Sizeof(uintptr(0)) > 4 {
		highBit := uint64(1) << 40
		want |= uintptr(highBit)
	}
	if !sliderPointerSlotsContract(want) {
		t.Fatalf("slider pointer slots truncated %#x", want)
	}
}

func TestSliderNativeLayout(t *testing.T) {
	windowSize, widgetData, style, owner, enabledColor, highlightColor, thumb, bgImage, disabledImage := sliderNativeLayout()
	if unsafe.Sizeof(uintptr(0)) == 4 {
		if windowSize != 404 || widgetData != 32 || style != 44 || owner != 52 || enabledColor != 64 ||
			highlightColor != 72 || thumb != 400 || bgImage != 24 || disabledImage != 48 {
			t.Fatalf("PE32 slider layout = %d/%d/%d/%d/%d/%d/%d/%d/%d",
				windowSize, widgetData, style, owner, enabledColor, highlightColor, thumb, bgImage, disabledImage)
		}
		return
	}
	if windowSize != 528 || widgetData != 56 || style != 72 || owner != 80 || enabledColor != 104 ||
		highlightColor != 120 || thumb != 520 || bgImage != 32 || disabledImage != 80 {
		t.Fatalf("native 64-bit slider layout = %d/%d/%d/%d/%d/%d/%d/%d/%d",
			windowSize, widgetData, style, owner, enabledColor, highlightColor, thumb, bgImage, disabledImage)
	}
}
