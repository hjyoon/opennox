package legacy

import (
	"testing"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

func TestObjectToggleNative4E7650(t *testing.T) {
	obj := &server.Object{ObjFlags: object.Flags(0xa5010000)}
	result, wasEnabled := objectToggleNative4E7650(
		obj,
		func(got *server.Object) uint32 {
			if got != obj {
				t.Fatalf("setOff object = %p, want %p", got, obj)
			}
			return 0x81234567
		},
		func(*server.Object) byte {
			t.Fatal("setOn called for enabled native object")
			return 0
		},
	)
	if result != 0x67 || !wasEnabled {
		t.Fatalf("result, wasEnabled = %#x, %v; want %#x, true", result, wasEnabled, byte(0x67))
	}
}

func TestObjectToggleNative4E7650Disabled(t *testing.T) {
	obj := &server.Object{ObjFlags: object.Flags(0x80000020)}
	result, wasEnabled := objectToggleNative4E7650(
		obj,
		func(*server.Object) uint32 {
			t.Fatal("setOff called for disabled native object")
			return 0
		},
		func(got *server.Object) byte {
			if got != obj {
				t.Fatalf("setOn object = %p, want %p", got, obj)
			}
			return 0xc3
		},
	)
	if result != 0xc3 || wasEnabled {
		t.Fatalf("result, wasEnabled = %#x, %v; want %#x, false", result, wasEnabled, byte(0xc3))
	}
}

func TestObjectToggleNative4E7650NilFault(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("nil native object did not fault")
		}
	}()
	objectToggleNative4E7650(
		nil,
		func(*server.Object) uint32 { return 0 },
		func(*server.Object) byte { return 0 },
	)
}
