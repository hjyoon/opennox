package legacy

/*
#include <stdint.h>

uint32_t* sub_56F8C0(int token, float value);
uint32_t* sub_56F820(int token, unsigned char level);
void* nox_xxx_book_45DBE0(void* kind, int ability, int index);

static inline void nox_playerLevelSet_protectExperience_4EF410(
		uint32_t token, float value) {
	(void)sub_56F8C0((int32_t)token, value);
}

static inline void nox_playerLevelSet_protectLevel_4EF410(
		uint32_t token, uint8_t level) {
	(void)sub_56F820((int32_t)token, (unsigned char)level);
}

static inline void nox_playerLevelSet_book_4EF410(
		int32_t kind, int32_t ability, int32_t index) {
	(void)nox_xxx_book_45DBE0(
		(void*)(uintptr_t)(uint32_t)kind,
		(int)ability,
		(int)index);
}
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func playerLevelSetRuntime4EF410() server.PlayerLevelSetRuntime4EF410 {
	return server.PlayerLevelSetRuntime4EF410{
		ProtectExperience: func(token uint32, value float32) {
			C.nox_playerLevelSet_protectExperience_4EF410(
				C.uint32_t(token),
				C.float(value),
			)
		},
		ProtectLevel: func(token uint32, level uint8) {
			C.nox_playerLevelSet_protectLevel_4EF410(
				C.uint32_t(token),
				C.uint8_t(level),
			)
		},
		ReadValues: playerReadValuesCall4EEDC0,
		BookAbility: func(kind, ability, index int32) {
			C.nox_playerLevelSet_book_4EF410(
				C.int32_t(kind),
				C.int32_t(ability),
				C.int32_t(index),
			)
		},
		// PauseFX 0057AF30 remains the sole pointer-narrowing dependency in
		// this path. Reuse the one isolated adapter until that function is
		// restored rather than introducing another conversion site.
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
