package legacy

import (
	"testing"

	"github.com/opennox/opennox/v1/server"
)

func TestProjectileTrailUpdate53AEC0RegistrationDispatchesToGo(t *testing.T) {
	callback, size, ok := server.ObjectUpdateHandler("ProjectileTrailUpdate")
	if !ok || callback == nil || size != 0 {
		t.Fatalf("ProjectileTrailUpdate registration = %p/%d/%t", callback, size, ok)
	}
	// Native callback accepts nil; the old C updater dereferenced a PE32 object.
	server.CallObjectUpdate(callback, nil)
}
