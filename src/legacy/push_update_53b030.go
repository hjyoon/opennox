package legacy

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

func pushUpdateNative53B030(source *server.Object) {
	srv := GetServer()
	server.PushUpdate53B030(source, server.PushUpdateRuntime53B030{
		PushUnits: func(position types.Pointf, outer, inner, force float32) {
			srv.S().MapPushUnitsAround52E040(position, outer, inner, force, server.MapPushUnitsAroundRuntime52E040{
				ApplyForce: srv.ApplyForce,
			})
		},
	})
}

// Keep this indirection dynamic so tests can prove that object dispatch does
// not cross the C trampoline while retaining the original callback identity.
var pushUpdateCall53B030 = pushUpdateNative53B030
