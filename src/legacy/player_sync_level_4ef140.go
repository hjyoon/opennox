package legacy

import "C"

import "github.com/opennox/opennox/v1/server"

func playerSyncLevelRuntime4EF140() server.PlayerSyncLevelRuntime4EF140 {
	return server.PlayerSyncLevelRuntime4EF140{
		ProtectLevel: func(token uint32, level uint8) {
			sub_56F820(C.int(token), C.uchar(level))
		},
		ReadValues: playerReadValuesCall4EEDC0,
	}
}

func playerSyncLevelCall4EF140(unit *server.Object) int32 {
	return GetServer().S().PlayerSyncLevel4EF140(
		unit,
		playerSyncLevelRuntime4EF140(),
	)
}
