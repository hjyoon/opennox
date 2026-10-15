package legacy

/*
#include "defs.h"
#include "GAME3_1.h"
#include "GAME3_2.h"
#include "client__drawable__update__mmislup.h"
#include "client__drawable__update__telwake.h"
#include "client__drawable__update__sparklup.h"
#include "client__drawable__update__vortexup.h"
#include "client__drawable__update__manabomb.h"
#include "client__drawable__update__mtailup.h"
*/
import "C"

// These entries keep the parser's C symbol identities while delegating native
// pointers to the existing renderer. A dedicated server has no renderer hook.

//export nox_xxx_updDrawMagicMissile_4CD9E0
func nox_xxx_updDrawMagicMissile_4CD9E0(vp *nox_draw_viewport_t, dr *nox_drawable) int32 {
	if ClientUpdateMagicMissile4CD9E0 == nil {
		return 1
	}
	return int32(ClientUpdateMagicMissile4CD9E0(asViewport(vp), asDrawable(dr)))
}

//export nox_xxx_updDrawTeleportWake_4CD8D0
func nox_xxx_updDrawTeleportWake_4CD8D0(vp *nox_draw_viewport_t, dr *nox_drawable) int32 {
	if ClientUpdateTeleportWake4CD8D0 == nil {
		return 1
	}
	return int32(ClientUpdateTeleportWake4CD8D0(asViewport(vp), asDrawable(dr)))
}

//export nox_xxx_updDrawSparkleTrail_4CDBF0
func nox_xxx_updDrawSparkleTrail_4CDBF0(vp *nox_draw_viewport_t, dr *nox_drawable) int32 {
	if ClientUpdateSparkleTrail4CDBF0 == nil {
		return 1
	}
	return int32(ClientUpdateSparkleTrail4CDBF0(asViewport(vp), asDrawable(dr)))
}

//export nox_xxx_updDrawVortexSource_4CC950
func nox_xxx_updDrawVortexSource_4CC950(vp *nox_draw_viewport_t, dr *nox_drawable) int32 {
	if ClientUpdateVortexSource4CC950 == nil {
		return 1
	}
	return int32(ClientUpdateVortexSource4CC950(asViewport(vp), asDrawable(dr)))
}

//export nox_xxx_updDrawManabombCharge_4CCAC0
func nox_xxx_updDrawManabombCharge_4CCAC0(vp *nox_draw_viewport_t, dr *nox_drawable) int32 {
	if ClientUpdateManaBombCharge4CCAC0 == nil {
		return 1
	}
	return int32(ClientUpdateManaBombCharge4CCAC0(asViewport(vp), asDrawable(dr)))
}

//export sub_4CA720
func sub_4CA720(vp *nox_draw_viewport_t, dr *nox_drawable) int32 {
	if ClientUpdateManaBombOrb4CA720 == nil {
		return 1
	}
	return int32(ClientUpdateManaBombOrb4CA720(asViewport(vp), asDrawable(dr)))
}

//export nox_xxx_updDrawCloud_4CE1D0
func nox_xxx_updDrawCloud_4CE1D0(vp *nox_draw_viewport_t, dr *nox_drawable) int32 {
	if ClientUpdateCloud4CE1D0 == nil {
		return 1
	}
	return int32(ClientUpdateCloud4CE1D0(asViewport(vp), asDrawable(dr), 75))
}

//export sub_4CE360
func sub_4CE360(vp *nox_draw_viewport_t, dr *nox_drawable) int32 {
	if ClientUpdateCloud4CE1D0 == nil {
		return 1
	}
	return int32(ClientUpdateCloud4CE1D0(asViewport(vp), asDrawable(dr), 35))
}

//export sub_4CE340
func sub_4CE340(vp *nox_draw_viewport_t, dr *nox_drawable) int32 {
	if ClientUpdateCloudRise4CE340 == nil {
		return 1
	}
	return int32(ClientUpdateCloudRise4CE340(asViewport(vp), asDrawable(dr)))
}

//export sub_4CCD00
func sub_4CCD00(vp *nox_draw_viewport_t, dr *nox_drawable) int32 {
	if ClientUpdateMeteor4CCD00 == nil {
		return 1
	}
	return int32(ClientUpdateMeteor4CCD00(asViewport(vp), asDrawable(dr)))
}

//export nox_xxx_updDrawFist_4CCDB0
func nox_xxx_updDrawFist_4CCDB0(vp *nox_draw_viewport_t, dr *nox_drawable) int32 {
	if ClientUpdateFist4CCDB0 == nil {
		return 1
	}
	return int32(ClientUpdateFist4CCDB0(asViewport(vp), asDrawable(dr)))
}

//export nox_xxx_sprite_4CA540
func nox_xxx_sprite_4CA540(vp *nox_draw_viewport_t, dr *nox_drawable) int32 {
	if ClientUpdatePredictLinear4CA540 == nil {
		return 1
	}
	return int32(ClientUpdatePredictLinear4CA540(asViewport(vp), asDrawable(dr)))
}

//export nox_xxx_updDrawMagic_4CDD80
func nox_xxx_updDrawMagic_4CDD80(vp *nox_draw_viewport_t, dr *nox_drawable) {
	if ClientUpdateMagic4CDD80 != nil {
		ClientUpdateMagic4CDD80(asViewport(vp), asDrawable(dr))
	}
}
