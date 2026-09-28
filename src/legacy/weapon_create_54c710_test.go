package legacy

import (
	"testing"

	"github.com/opennox/opennox/v1/server"
)

func TestWeaponCreate54C710RegistrationDispatchesNativeObject(t *testing.T) {
	callback, ok := server.ObjectCreateHandler("WeaponCreate")
	if !ok || callback == nil {
		t.Fatalf("WeaponCreate callback = %p, registered=%t", callback, ok)
	}

	original := weaponCreateCall54C710
	t.Cleanup(func() {
		weaponCreateCall54C710 = original
	})
	want := new(server.Object)
	var got *server.Object
	weaponCreateCall54C710 = func(obj *server.Object) {
		got = obj
	}

	server.CallObjectCreate(callback, want)
	if got != want {
		t.Fatalf("native object = %p, want %p", got, want)
	}
}
