package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestSpellProjectileUpdate53B940RegistrationUsesNativeRecord(t *testing.T) {
	callback, size, ok := server.ObjectUpdateHandler("SpellProjectileUpdate")
	if !ok || callback == nil || size != unsafe.Sizeof(server.SpellProjectileUpdateData{}) {
		t.Fatalf("SpellProjectileUpdate registration = %p/%d/%t", callback, size, ok)
	}
	// Dispatch must enter Go. The legacy C body reads the obsolete PE32 +748
	// update-data slot and would fault even for this empty object.
	server.CallObjectUpdate(callback, new(server.Object))
}
