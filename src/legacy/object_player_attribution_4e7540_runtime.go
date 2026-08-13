package legacy

import "github.com/opennox/opennox/v1/server"

func recordPlayerAttributionRuntime4E7540(source, target *server.Object) {
	recordPlayerAttributionNative4E7540(source, target, gameFrameHook)
}
