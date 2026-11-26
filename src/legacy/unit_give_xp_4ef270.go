package legacy

import "github.com/opennox/opennox/v1/server"

func unitGiveXPRuntime4EF270() server.UnitGiveXPRuntime4EF270 {
	return server.UnitGiveXPRuntime4EF270{
		ProtectExperience: func(token uint32, award float32) {
			addPlayerProtectedExperience(token, award)
		},
		SyncLevel: func(unit *server.Object) {
			experienceLevelUpdateCall4EF2E0(unit)
		},
	}
}

func unitGiveXPCall4EF270(unit *server.Object, target float32) float64 {
	return GetServer().S().UnitGiveXP4EF270(unit, target, unitGiveXPRuntime4EF270())
}
