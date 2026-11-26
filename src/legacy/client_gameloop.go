//go:build !server

package legacy

/*
#include "GAME1.h"
#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "GAME2_1.h"
#include "GAME2_3.h"
#include "GAME3.h"
#include "GAME3_1.h"
#include "GAME3_3.h"
#include "GAME5.h"
#include "GAME5_2.h"

*/
import "C"
import "github.com/opennox/opennox/v1/client/noxrender"

func Sub_43CEB0() {
	C.sub_43CEB0()
}
func Sub_40A710(a1 int) uint32 {
	return uint32(C.sub_40A710(C.int(a1)))
}
func Nox_client_screenParticlesDraw_431720(vp *noxrender.Viewport) {
	nox_client_screenParticlesDraw_431720((*nox_draw_viewport_t)(vp.C()))
}
func Nox_client_newScreenParticle_431540(a1, a2, a3, a4, a5, a6, a7, a8, a9, a10 int) {
	nox_client_newScreenParticle_431540(C.int(a1), C.int(a2), C.int(a3), C.int(a4), C.int(a5), C.int(a6), C.char(a7), C.char(a8), C.char(a9), C.char(a10))
}
