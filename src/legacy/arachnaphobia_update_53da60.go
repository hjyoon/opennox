package legacy

import "github.com/opennox/opennox/v1/server"

func arachnaphobiaUpdateNative53DA60(source *server.Object) {
	if source == nil {
		return
	}
	server.ArachnaphobiaUpdateRaw53DA60(source, arachnaphobiaRawRuntime53DA60())
}

// Keep this indirection dynamic so tests can prove direct Go dispatch while
// preserving the legacy callback address stored in thing.bin objects.
var arachnaphobiaUpdateCall53DA60 = arachnaphobiaUpdateNative53DA60
