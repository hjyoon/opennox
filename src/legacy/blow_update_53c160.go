package legacy

import "github.com/opennox/opennox/v1/server"

func blowUpdateNative53C160(source *server.Object) {
	GetServer().S().BlowUpdate53C160(source)
}

// Keep this indirection dynamic so tests can prove direct native-width Go
// dispatch while retaining the legacy callback identity stored in thing.bin.
var blowUpdateCall53C160 = blowUpdateNative53C160
