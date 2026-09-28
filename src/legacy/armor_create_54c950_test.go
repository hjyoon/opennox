package legacy

import (
	"testing"

	"github.com/opennox/opennox/v1/server"
)

func TestArmorCreate54C950RegistrationDispatchesNativeObject(t *testing.T) {
	callback, ok := server.ObjectCreateHandler("ArmorCreate")
	if !ok || callback == nil {
		t.Fatalf("ArmorCreate callback = %p, registered=%t", callback, ok)
	}

	original := armorCreateCall54C950
	t.Cleanup(func() {
		armorCreateCall54C950 = original
	})
	want := new(server.Object)
	var got *server.Object
	armorCreateCall54C950 = func(obj *server.Object) {
		got = obj
	}

	server.CallObjectCreate(callback, want)
	if got != want {
		t.Fatalf("native object = %p, want %p", got, want)
	}
}
