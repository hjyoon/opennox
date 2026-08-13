package legacy

import "github.com/opennox/opennox/v1/server"

func objectToggleNative4E7650(
	obj *server.Object,
	setOff func(*server.Object) uint32,
	setOn func(*server.Object) byte,
) (result byte, wasEnabled bool) {
	return objectToggle4E7650(obj, objectToggleHooks4E7650[*server.Object]{
		flags: func(obj *server.Object) uint32 {
			return uint32(obj.ObjFlags)
		},
		setOff: setOff,
		setOn:  setOn,
	})
}
