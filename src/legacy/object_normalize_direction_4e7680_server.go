package legacy

import "github.com/opennox/opennox/v1/server"

func objectNormalizeDirectionNative4E7680(obj *server.Object) *server.Object {
	return objectNormalizeDirection4E7680(obj, objectNormalizeDirectionHooks4E7680[*server.Object]{
		direction: func(obj *server.Object) int16 {
			return int16(obj.Direction1)
		},
		addDirection: func(obj *server.Object, delta int16) {
			obj.Direction1 = server.Dir16(uint16(obj.Direction1) + uint16(delta))
		},
	})
}
