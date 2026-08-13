package legacy

import (
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

func objectSetOnNative4E75B0(
	obj *server.Object,
	audio func(*server.Object),
	hasCollideOrUpdate func(*server.Object) byte,
) byte {
	return objectSetOn4E75B0(obj, objectSetOnHooks4E75B0[*server.Object]{
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
		clearFlags: func(obj *server.Object, flags uint32) {
			obj.ObjFlags &^= object.Flags(flags)
		},
		hasCollideOrUpdate: hasCollideOrUpdate,
	})
}
