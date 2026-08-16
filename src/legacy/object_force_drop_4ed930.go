package legacy

import "github.com/opennox/opennox/v1/server"

func objectForceDropRuntime4ED930() server.ObjectForceDropRuntime4ED930 {
	return server.ObjectForceDropRuntime4ED930{
		Dispatch: objectDropDispatchCall4ED790,
	}
}

func objectForceDropCall4ED930(owner, item *server.Object) int32 {
	return GetServer().S().ObjectForceDrop4ED930(
		owner,
		item,
		objectForceDropRuntime4ED930(),
	)
}

func Nox_xxx_invForceDropItem_4ED930(owner, item *server.Object) int {
	return int(objectForceDropCall4ED930(owner, item))
}
