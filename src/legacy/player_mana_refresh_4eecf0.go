package legacy

/*
#include <stdint.h>

uint32_t* nox_xxx_protectMana_56F9E0(int token, short delta);

static inline uintptr_t nox_playerManaRefresh_protectMana(int token, int16_t delta) {
	return (uintptr_t)nox_xxx_protectMana_56F9E0(token, (short)delta);
}
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func playerManaRefreshCall4EECF0(unit *server.Object) uintptr {
	return GetServer().S().PlayerManaRefresh4EECF0(
		unit,
		func(token uint32, maximum int16) uintptr {
			return uintptr(C.nox_playerManaRefresh_protectMana(
				C.int(int32(token)),
				C.int16_t(maximum),
			))
		},
	)
}

// Nox_xxx_playerManaRefresh_4EECF0 gives Go-owned callers the restored native
// path. Its uintptr result is an observable return-register value only and
// must not be converted back into a Go pointer.
func Nox_xxx_playerManaRefresh_4EECF0(unit *server.Object) uintptr {
	return playerManaRefreshCall4EECF0(unit)
}
