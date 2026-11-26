package legacy

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
			setPlayerProtectedStat(token, value)
		},
		ProtectUint16: func(token uint32, value uint16) {
			setPlayerProtectedStat(token, uint32(value))
		},
		WideLen: func(info *server.PlayerInfo) uint32 {
			return uint32(nox_wcslen((*wchar2_t)(info.C())))
		},
		ProtectName: func(info *server.PlayerInfo, size, token uint32) int32 {
			return int32(sub_56FB00((*C.int)(info.C()), C.uint(size), C.int(token)))
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
