package legacy

import "github.com/opennox/opennox/v1/server"

func sentryGlobeUpdateNative510E60(source *server.Object) {
	GetServer().S().SentryGlobeUpdate510E60(source)
}

// Keep this indirection dynamic so tests can prove native-width Go dispatch
// while retaining the callback identity stored in thing.bin.
var sentryGlobeUpdateCall510E60 = sentryGlobeUpdateNative510E60
