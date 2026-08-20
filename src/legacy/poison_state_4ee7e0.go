package legacy

import "github.com/opennox/opennox/v1/server"

func activatePoisonRuntime4EE7E0() server.ActivatePoisonRuntime4EE7E0 {
	return server.ActivatePoisonRuntime4EE7E0{
		PoisonProtectEngage: PoisonProtectEffectPointer4DFDE0(),
	}
}

// Nox_xxx_activatePoison_4EE7E0 exposes the native Go path to Go-owned
// callers without a CGo round trip.
func Nox_xxx_activatePoison_4EE7E0(unit *server.Object, increment, maximum int32) int32 {
	return GetServer().S().ActivatePoison4EE7E0(unit, increment, maximum, activatePoisonRuntime4EE7E0())
}

func setSomePoisonDataCall4EEA90(unit *server.Object, value int32) {
	GetServer().S().SetPoison4EEA90(unit, value)
}
