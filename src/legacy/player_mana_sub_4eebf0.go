package legacy

import "github.com/opennox/opennox/v1/server"

func playerManaSubCall4EEBF0(obj *server.Object, amount int32) uintptr {
	return GetServer().S().PlayerManaSub4EEBF0(
		obj,
		amount,
		func(token uint32, delta int16) uintptr {
			return uintptr(addPlayerProtectedStat(token, uint32(delta)))
		},
	)
}

// Nox_xxx_playerManaSub_4EEBF0 gives Go-owned callers the restored native
// path. Narrowing amount to int32 preserves the original whole C int slot.
// The uintptr result is an observable register value only and must not be
// converted back into a pointer.
func Nox_xxx_playerManaSub_4EEBF0(obj *server.Object, amount int) uintptr {
	return playerManaSubCall4EEBF0(obj, int32(amount))
}
