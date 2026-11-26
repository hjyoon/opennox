package legacy

import "github.com/opennox/opennox/v1/server"

func playerManaRefreshCall4EECF0(unit *server.Object) uintptr {
	return GetServer().S().PlayerManaRefresh4EECF0(
		unit,
		func(token uint32, maximum int16) uintptr {
			return uintptr(addPlayerProtectedStat(token, uint32(maximum)))
		},
	)
}

// Nox_xxx_playerManaRefresh_4EECF0 gives Go-owned callers the restored native
// path. Its uintptr result is an observable return-register value only and
// must not be converted back into a Go pointer.
func Nox_xxx_playerManaRefresh_4EECF0(unit *server.Object) uintptr {
	return playerManaRefreshCall4EECF0(unit)
}
