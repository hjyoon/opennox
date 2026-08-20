package legacy

import "github.com/opennox/opennox/v1/server"

func unitHPSetOnMaxRuntime4EE6F0() server.UnitHPSetOnMaxRuntime4EE6F0 {
	return server.UnitHPSetOnMaxRuntime4EE6F0{
		SetHP: Nox_xxx_unitSetHP_4E4560,
	}
}

func unitHPSetOnMaxCall4EE6F0(unit *server.Object) {
	GetServer().S().UnitHPSetOnMax4EE6F0(unit, unitHPSetOnMaxRuntime4EE6F0())
}
