package legacy

import "github.com/opennox/opennox/v1/server"

func playerManaAddCall4EEB80(obj *server.Object, amount int32) uint16 {
	return GetServer().S().PlayerManaAdd4EEB80(
		obj,
		amount,
		func(token uint32, delta int16) {
			addPlayerProtectedStat(token, uint32(delta))
		},
		func(token uint32, maximum uint16) uint16 {
			return uint16(setPlayerProtectedStat(token, uint32(maximum)))
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

// Nox_xxx_playerManaAddFull4EEB80 preserves the original DWORD argument slot
// for ported callers such as Drain Mana. The historical short API above keeps
// its conversion; the native mana core owns word wrapping and protection.
func Nox_xxx_playerManaAddFull4EEB80(obj *server.Object, amount int32) uint16 {
	return playerManaAddCall4EEB80(obj, amount)
}
