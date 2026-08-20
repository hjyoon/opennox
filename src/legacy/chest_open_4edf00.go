package legacy

import "github.com/opennox/opennox/v1/server"

func chestOpenCall4EDF00(chest, unit *server.Object) {
	outer := GetServer()
	outer.S().ChestOpen4EDF00(chest, unit, server.ChestOpenRuntime4EDF00{
		RefreshUnit: Nox_xxx_unit_511810,
		Dispatch:    objectDropDispatchCall4ED790,
	})
}
