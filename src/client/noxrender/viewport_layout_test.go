package noxrender

import (
	"testing"
	"unsafe"
)

func TestViewportNativeLayout(t *testing.T) {
	var vp Viewport
	word := unsafe.Sizeof(uintptr(0))
	tests := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{name: "Size", got: unsafe.Offsetof(vp.Size), want: 8 * word},
		{name: "Field10", got: unsafe.Offsetof(vp.Field10), want: 10 * word},
		{name: "Field11", got: unsafe.Offsetof(vp.Field11), want: 11 * word},
		{name: "Jiggle12", got: unsafe.Offsetof(vp.Jiggle12), want: 12 * word},
		{name: "sizeof", got: unsafe.Sizeof(vp), want: 13 * word},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("Viewport %s = %d, want %d", tt.name, tt.got, tt.want)
		}
	}
}
