package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestAntiSpellProjectileUpdate53BB00RegistrationUsesNativeRecord(t *testing.T) {
	callback, size, ok := server.ObjectUpdateHandler("AntiSpellProjectileUpdate")
	if !ok || callback == nil || size != unsafe.Sizeof(server.MissileUpdateData{}) {
		t.Fatalf("AntiSpellProjectileUpdate registration = %p/%d/%t", callback, size, ok)
	}
	// This must dispatch to Go: the PE32 C body reads the obsolete +748 slot.
	server.CallObjectUpdate(callback, new(server.Object))
}
