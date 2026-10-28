package legacy

import (
	"github.com/opennox/opennox/v1/server"
)

func pushUpdateNative53B030(source *server.Object) {
	if source == nil || source.UpdateData == nil {
		return
	}
	server.PushUpdateRaw53B030(source, sparkTrailPushRawRuntime().PushUnits)
}

// Keep this indirection dynamic so tests can prove that object dispatch does
// not cross the C trampoline while retaining the original callback identity.
var pushUpdateCall53B030 = pushUpdateNative53B030
