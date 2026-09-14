package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestLoopAndDamageUpdate53B300NativeFlags(t *testing.T) {
	const (
		loopDisabled = object.Flags(0x1000000)
		loopActive   = object.Flags(0x40)
		keep         = object.Flags(0x8000100)
	)
	for _, tc := range []struct {
		name string
		in   object.Flags
		want object.Flags
	}{
		{"enable", keep, keep | loopActive},
		{"already enabled", keep | loopActive, keep | loopActive},
		{"disable", keep | loopDisabled | loopActive, keep | loopDisabled},
		{"already disabled", keep | loopDisabled, keep | loopDisabled},
	} {
		t.Run(tc.name, func(t *testing.T) {
			obj := &Object{ObjFlags: tc.in}
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(obj)) <= 0xffffffff {
				t.Fatalf("object pointer = %p, want above 4 GiB", obj)
			}
			LoopAndDamageUpdate53B300(obj)
			if obj.ObjFlags != tc.want {
				t.Fatalf("flags = %#x, want %#x", obj.ObjFlags, tc.want)
			}
		})
	}
}
