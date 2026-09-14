package legacy

import "github.com/opennox/opennox/v1/server"

var undeadKillerUpdateCall53E190 = func(obj *server.Object) {
	outer := GetServer()
	server.UndeadKillerUpdate53E190(obj, outer.S().Frame(), outer.DelayedDelete)
}
