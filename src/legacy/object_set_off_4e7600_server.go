package legacy

import (
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

func objectSetOffNative4E7600(obj *server.Object, audio func(*server.Object)) uint32 {
	return objectSetOff4E7600(obj, objectSetOffHooks4E7600[*server.Object]{
		flags: func(obj *server.Object) uint32 {
			return uint32(obj.ObjFlags)
		},
		class: func(obj *server.Object) uint32 {
			return uint32(obj.ObjClass)
		},
		audio: audio,
		setOnOff: func(obj *server.Object, enabled bool) {
			obj.SetOnOff(enabled)
		},
		setFlags: func(obj *server.Object, flags uint32) {
			obj.ObjFlags = object.Flags(flags)
		},
	})
}
