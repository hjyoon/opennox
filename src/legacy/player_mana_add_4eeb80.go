package legacy

/*
#include <stdint.h>

uint32_t* nox_xxx_protectPlayerHPMana_56F870(int token, unsigned short value);
uint32_t* nox_xxx_protectMana_56F9E0(int token, short delta);

static inline void nox_playerManaAdd_protectMana(int token, short delta) {
	(void)nox_xxx_protectMana_56F9E0(token, delta);
}

static inline uint16_t nox_playerManaAdd_protectMaximum(int token, unsigned short value) {
	return (uint16_t)(uintptr_t)nox_xxx_protectPlayerHPMana_56F870(token, value);
}
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func playerManaAddCall4EEB80(obj *server.Object, amount int32) uint16 {
	return GetServer().S().PlayerManaAdd4EEB80(
		obj,
		amount,
		func(token uint32, delta int16) {
			C.nox_playerManaAdd_protectMana(C.int(int32(token)), C.short(delta))
		},
		func(token uint32, maximum uint16) uint16 {
			return uint16(C.nox_playerManaAdd_protectMaximum(
				C.int(int32(token)),
				C.ushort(maximum),
			))
		},
	)
}

// Nox_xxx_playerManaAdd_4EEB80 gives Go-owned callers the restored native
// path without a CGo round trip through the historical fixed-offset body.
// The old short parameter conversion sign-extends into the exact 32-bit
// amount slot modeled by the server implementation.
func Nox_xxx_playerManaAdd_4EEB80(obj *server.Object, amount int) uint16 {
	return playerManaAddCall4EEB80(obj, int32(int16(amount)))
}
