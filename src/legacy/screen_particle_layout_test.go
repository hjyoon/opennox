//go:build !server

package legacy

import (
	"testing"
	"unsafe"
)

func TestScreenParticleNativeLayout(t *testing.T) {
	var p Nox_screenParticle
	word := unsafe.Sizeof(uintptr(0))
	field40 := uintptr(40)
	next := uintptr(44)
	size := uintptr(52)
	if word == 8 {
		field40 = 44
		next = 48
		size = 64
	}
	tests := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{name: "Field_4", got: unsafe.Offsetof(p.Field_4), want: word},
		{name: "Field_40", got: unsafe.Offsetof(p.Field_40), want: field40},
		{name: "Field_44", got: unsafe.Offsetof(p.Field_44), want: next},
		{name: "Field_48", got: unsafe.Offsetof(p.Field_48), want: next + word},
		{name: "sizeof", got: unsafe.Sizeof(p), want: size},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("Nox_screenParticle %s = %d, want %d", tt.name, tt.got, tt.want)
		}
	}
}
