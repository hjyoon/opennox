package legacy

import "github.com/opennox/opennox/v1/server"

func unitAdjustHPRuntime4EE460() server.UnitAdjustHPRuntime4EE460 {
	return server.UnitAdjustHPRuntime4EE460{
		SetHP: Nox_xxx_unitSetHP_4E4560,
	}
}

func unitAdjustHPCall4EE460(unit *server.Object, delta int32) {
	GetServer().S().UnitAdjustHP4EE460(unit, delta, unitAdjustHPRuntime4EE460())
}

func mobInformOwnerHPCall4EE4C0(obj *server.Object) {
	GetServer().S().MobInformOwnerHP4EE4C0(obj)
}

func currentHPReportCall4D8620(recipient int32, obj *server.Object) int32 {
	return GetServer().S().CurrentHPReport4D8620(recipient, obj)
}
