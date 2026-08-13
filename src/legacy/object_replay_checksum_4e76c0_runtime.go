package legacy

import "github.com/opennox/opennox/v1/server"

// Sub_4E76C0 runs the replay checksum pass over the live object list.
func Sub_4E76C0() {
	objectReplayChecksumPassNative4E76C0(func() *server.Object {
		return GetServer().S().Objs.First()
	})
}
