package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestMagicMissileUpdate53BDA0RegistrationUsesNativeRecord(t *testing.T) {
	callback, size, ok := server.ObjectUpdateHandler("MagicMissileUpdate")
	if !ok || callback == nil || size != unsafe.Sizeof(server.MissileUpdateData{}) {
		t.Fatalf("MagicMissileUpdate registration = %p/%d/%t", callback, size, ok)
	}
	// The Go callback handles absent update data; the PE32 C implementation
	// would dereference the native object at its obsolete +748 offset.
	server.CallObjectUpdate(callback, new(server.Object))
}
