package legacy

// #include <stdint.h>
import "C"

import "github.com/opennox/opennox/v1/server"

func playerLevelSetRuntime4EF410() server.PlayerLevelSetRuntime4EF410 {
	return server.PlayerLevelSetRuntime4EF410{
		ProtectExperience: func(token uint32, value float32) {
			sub_56F8C0(C.int(token), C.float(value))
		},
		ProtectLevel: func(token uint32, level uint8) {
			sub_56F820(C.int(token), C.uchar(level))
		},
		ReadValues: playerReadValuesCall4EEDC0,
		BookAbility: func(kind, ability, index int32) {
			nox_xxx_book_45DBE0(C.uintptr_t(uint32(kind)), C.int(ability), C.int(index))
		},
		PauseFX: experienceLevelUpdateRuntime4EF2E0().PauseFX,
	}
}

func playerLevelSetCall4EF410(unit *server.Object, level uint8) {
	GetServer().S().PlayerLevelSet4EF410(
		unit,
		level,
		playerLevelSetRuntime4EF410(),
	)
}
