package legacy

/*
#include <stdint.h>

uint32_t* nox_xxx_protectMana_56F9E0(int token, short delta);

static inline uintptr_t nox_playerManaSub_protectMana(int token, int16_t delta) {
	return (uintptr_t)nox_xxx_protectMana_56F9E0(token, (short)delta);
}
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func playerManaSubCall4EEBF0(obj *server.Object, amount int32) uintptr {
	return GetServer().S().PlayerManaSub4EEBF0(
		obj,
		amount,
		func(token uint32, delta int16) uintptr {
			return uintptr(C.nox_playerManaSub_protectMana(
				C.int(int32(token)),
				C.int16_t(delta),
			))
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
