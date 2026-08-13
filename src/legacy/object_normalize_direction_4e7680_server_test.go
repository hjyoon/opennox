package legacy

import (
	"testing"

	"github.com/opennox/opennox/v1/server"
)

func TestObjectNormalizeDirectionNative4E7680(t *testing.T) {
	obj := &server.Object{
		Mass:       17.5,
		Direction1: server.Dir16(0xfeff),
		Direction2: server.Dir16(0xa55a),
		Field32:    0x89abcdef,
	}
	got := objectNormalizeDirectionNative4E7680(obj)
	if got != obj {
		t.Fatalf("return object = %p, want %p", got, obj)
	}
	if obj.Direction1 != 255 {
		t.Fatalf("Direction1 = %#04x, want %#04x", obj.Direction1, server.Dir16(255))
	}
	if obj.Mass != 17.5 || obj.Direction2 != 0xa55a || obj.Field32 != 0x89abcdef {
		t.Fatalf("adjacent fields changed: Mass=%v Direction2=%#04x Field32=%#08x", obj.Mass, obj.Direction2, obj.Field32)
	}
}

func TestObjectNormalizeDirectionNative4E7680AllSignedExtremes(t *testing.T) {
	for _, initial := range []server.Dir16{0, 255, 256, 0x7fff, 0x8000, 0xffff} {
		obj := &server.Object{Direction1: initial}
		objectNormalizeDirectionNative4E7680(obj)
		if want := server.Dir16(uint16(initial) & 0xff); obj.Direction1 != want {
			t.Fatalf("initial %#04x: Direction1 = %#04x, want %#04x", initial, obj.Direction1, want)
		}
	}
}

func TestObjectNormalizeDirectionNative4E7680NilFault(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("nil native object did not fault")
		}
	}()
	objectNormalizeDirectionNative4E7680(nil)
}
