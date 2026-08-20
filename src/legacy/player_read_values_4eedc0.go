package legacy

/*
#include <stddef.h>
#include <stdint.h>

uint32_t* sub_56F780(int token, int value);
uint32_t* nox_xxx_protectPlayerHPMana_56F870(int token, unsigned short value);
size_t nox_wcslen(const uint16_t* value);
int sub_56FB00(int* value, unsigned int size, int token);

static inline void nox_playerReadValues_protectInt_4EEDC0(
		uint32_t token, uint32_t value) {
	(void)sub_56F780((int32_t)token, (int32_t)value);
}

static inline void nox_playerReadValues_protectUint16_4EEDC0(
		uint32_t token, uint16_t value) {
	(void)nox_xxx_protectPlayerHPMana_56F870((int32_t)token, value);
}

static inline uint32_t nox_playerReadValues_wideLen_4EEDC0(void* value) {
	return (uint32_t)nox_wcslen((const uint16_t*)value);
}

static inline int32_t nox_playerReadValues_protectName_4EEDC0(
		void* value, uint32_t size, uint32_t token) {
	return (int32_t)sub_56FB00((int*)value, size, (int32_t)token);
}
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func playerReadValuesRuntime4EEDC0() server.PlayerReadValuesRuntime4EEDC0 {
	return server.PlayerReadValuesRuntime4EEDC0{
		SetHP: Nox_xxx_unitSetHP_4E4560,
		SoloMode: func() int32 {
			return int32(Sub_4D6F30())
		},
		Ability: abilityGivePlayerAllRuntime4EED40(),
		ProtectInt: func(token, value uint32) {
			C.nox_playerReadValues_protectInt_4EEDC0(
				C.uint32_t(token),
				C.uint32_t(value),
			)
		},
		ProtectUint16: func(token uint32, value uint16) {
			C.nox_playerReadValues_protectUint16_4EEDC0(
				C.uint32_t(token),
				C.uint16_t(value),
			)
		},
		WideLen: func(info *server.PlayerInfo) uint32 {
			return uint32(C.nox_playerReadValues_wideLen_4EEDC0(info.C()))
		},
		ProtectName: func(info *server.PlayerInfo, size, token uint32) int32 {
			return int32(C.nox_playerReadValues_protectName_4EEDC0(
				info.C(),
				C.uint32_t(size),
				C.uint32_t(token),
			))
		},
	}
}

func playerReadValuesCall4EEDC0(unit *server.Object, rewardArg int32) int32 {
	return GetServer().S().PlayerReadValues4EEDC0(
		unit,
		rewardArg,
		playerReadValuesRuntime4EEDC0(),
	)
}
