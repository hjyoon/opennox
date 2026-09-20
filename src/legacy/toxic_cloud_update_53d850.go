package legacy

import "github.com/opennox/opennox/v1/server"

func toxicCloudUpdateRuntime53D850() server.ToxicCloudUpdateRuntime53D850 {
	outer := GetServer()
	return server.ToxicCloudUpdateRuntime53D850{
		ActivatePoison: Nox_xxx_activatePoison_4EE7E0,
		DelayedDelete:  outer.DelayedDelete,
	}
}

func toxicCloudUpdateNative53D850(cloud *server.Object) {
	GetServer().S().ToxicCloudUpdate53D850(cloud, toxicCloudUpdateRuntime53D850())
}

func smallToxicCloudUpdateNative53D960(cloud *server.Object) {
	GetServer().S().SmallToxicCloudUpdate53D960(cloud, toxicCloudUpdateRuntime53D850())
}

// Keep these indirections dynamic so tests can prove direct Go dispatch while
// preserving the legacy callback addresses stored in thing.bin objects.
var (
	toxicCloudUpdateCall53D850      = toxicCloudUpdateNative53D850
	smallToxicCloudUpdateCall53D960 = smallToxicCloudUpdateNative53D960
)
