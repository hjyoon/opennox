//go:build !server

package legacy

/*
#include "defs.h"
#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "GAME2_3.h"
#include "GAME3_1.h"
#include "client__draw__drawrays.h"
#include "client__draw__fx.h"
#include "client__draw__glowdraw.h"
#include "client__gui__guiggovr.h"
extern uint32_t dword_5d4594_1200776;
void nox_xxx_tileDrawMB_481C20_A(nox_draw_viewport_t* vp, int v3);
void nox_xxx_tileDrawMB_481C20_B(nox_draw_viewport_t* vp, int v78);
void nox_xxx_tileDrawMB_481C20_C_textured(nox_draw_viewport_t* vp, int v72, int v78);
void  nox_xxx_cliLight16_469140(nox_drawable* dr, nox_draw_viewport_t* vp);
void nox_xxx_clientDrawAll_436100_draw_A();
void nox_xxx_clientDrawAll_436100_draw_B();
void nox_xxx_drawAllMB_475810_draw_A(nox_draw_viewport_t* vp);
int nox_xxx_drawAllMB_475810_draw_B(nox_draw_viewport_t* vp);
void nox_xxx_drawAllMB_475810_draw_C(nox_draw_viewport_t* vp, int v36, int v7);
int sub_436F50();
*/
import "C"
import (
	"image"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/server"
)

func Nox_xxx_clientDrawAll_436100_draw_A() {
	nox_xxx_clientDrawAll_436100_draw_A()
}

func Nox_xxx_drawMinimapAndLines_4738E0() {
	nox_xxx_drawMinimapAndLines_4738E0()
}

func Nox_xxx_clientDrawAll_436100_draw_B() {
	nox_xxx_clientDrawAll_436100_draw_B()
}

func Sub_436F50() {
	sub_436F50()
}

func Sub_437100() {
	sub_437100()
}

func Sub_470DE0() {
	C.sub_470DE0()
}

func Nox_xxx_clientEnumHover_476FA0() {
	nox_xxx_clientEnumHover_476FA0()
}

func Nox_xxx_netDrawRays_49BDD0(data [9]byte) {
	drawTransientRay49BDD0(&data)
}

func Nox_xxx_makeRayLightningParticles_49BDD0(from, to image.Point) {
	first := C.int2{field_0: C.int(from.X), field_4: C.int(from.Y)}
	last := C.int2{field_0: C.int(to.X), field_4: C.int(to.Y)}
	nox_xxx_makeLightningParticles_4999D0(C.int(C.dword_5d4594_1200776), &first, &last)
}

func Nox_xxx_makeRayPlasmaParticles_49BDD0(to image.Point) {
	nox_xxx_drawEnergyBolt_499710(C.int(to.X), C.int(to.Y), 10, C.int(C.dword_5d4594_1200776))
}

func Sub_49BBC0() {
	sub_49BBC0()
}

func Nox_xxx_polygonDrawColor_421B80() {
	nox_xxx_polygonDrawColor_421B80()
}

func Nox_xxx_cliToggleObsWindow_4357A0() {
	C.nox_xxx_cliToggleObsWindow_4357A0()
}

func Nox_xxx_motd_4467F0() {
	C.nox_xxx_motd_4467F0()
}

func Sub_42EBA0() int {
	return int(sub_42EBA0())
}

func Sub_49B6E0() {
	C.sub_49B6E0()
}

func Get_nox_thing_glow_orb_draw() unsafe.Pointer {
	return C.nox_thing_glow_orb_draw
}

func Get_nox_thing_glow_orb_move_draw() unsafe.Pointer {
	return C.nox_thing_glow_orb_move_draw
}

func Nox_xxx_drawAllMB_475810_draw_B(vp *noxrender.Viewport) int {
	return int(nox_xxx_drawAllMB_475810_draw_B((*nox_draw_viewport_t)(vp.C())))
}
func Sub_4C5060(vp *noxrender.Viewport) {
	C.sub_4C5060((*nox_draw_viewport_t)(vp.C()))
}
func AddSentryRay4C5020(from, to image.Point) {
	C.nox_client_addSentryRay_4C5020(
		C.uint16_t(from.X), C.uint16_t(from.Y),
		C.uint16_t(to.X), C.uint16_t(to.Y),
	)
}
func SentryRayCount4C5020() int {
	return int(C.nox_client_sentryRayCount_4C5020())
}
func SentryRayAt4C5020(index int) (image.Point, image.Point, bool) {
	var ray [4]C.uint16_t
	if C.nox_client_sentryRayAt_4C5020(C.int(index), &ray[0]) == 0 {
		return image.Point{}, image.Point{}, false
	}
	return image.Pt(int(ray[0]), int(ray[1])), image.Pt(int(ray[2]), int(ray[3])), true
}
func ClearSentryRays4C5050() {
	C.sub_4C5050()
}
func clientSentryRayScreenPosition4C5060(vp *noxrender.Viewport, from, to image.Point) (image.Point, image.Point) {
	ray := [4]C.uint16_t{
		C.uint16_t(from.X), C.uint16_t(from.Y),
		C.uint16_t(to.X), C.uint16_t(to.Y),
	}
	var cfrom, cto C.int2
	C.nox_client_sentryRayScreenPosition_4C5060(
		(*C.nox_sentry_const_viewport_t)(vp.C()),
		(*C.nox_sentry_const_u16_t)(unsafe.Pointer(&ray[0])),
		&cfrom, &cto,
	)
	return image.Pt(int(cfrom.field_0), int(cfrom.field_4)), image.Pt(int(cto.field_0), int(cto.field_4))
}
func Nox_xxx_drawWalls_473C10(vp *noxrender.Viewport, a2 *server.Wall) {
	nox_xxx_drawWalls_473C10((*nox_draw_viewport_t)(vp.C()), a2.C())
}
func Sub_476080(a1 unsafe.Pointer) int {
	return int(sub_476080((*C.uchar)(a1)))
}
func Sub_459DB0(dr *client.Drawable) int {
	return int(C.sub_459DB0((*nox_drawable)(dr.C())))
}
func Sub_472540(dr *client.Drawable) int {
	return int(sub_472540((*nox_drawable)(dr.C())))
}
func Sub_49A6A0(vp *noxrender.Viewport, dr *client.Drawable) {
	sub_49A6A0((*nox_draw_viewport_t)(vp.C()), (*nox_drawable)(dr.C()))
}
func Nox_xxx_sprite_4756E0_drawable(dr *client.Drawable) int {
	return int(C.nox_xxx_sprite_4756E0_drawable((*nox_drawable)(dr.C())))
}
func Nox_xxx_sprite_475740_drawable(dr *client.Drawable) int {
	return int(C.nox_xxx_sprite_475740_drawable((*nox_drawable)(dr.C())))
}
func Nox_xxx_sprite_4757A0_drawable(dr *client.Drawable) int {
	return int(C.nox_xxx_sprite_4757A0_drawable((*nox_drawable)(dr.C())))
}
func Sub_4757D0_drawable(dr *client.Drawable) int {
	return int(C.sub_4757D0_drawable((*nox_drawable)(dr.C())))
}
func Nox_xxx_tileDrawImpl_4826A0(vp *noxrender.Viewport) {
	nox_xxx_tileDrawImpl_4826A0((*nox_draw_viewport_t)(vp.C()))
}
func Nox_xxx_tileDrawMB_481C20_A(vp *noxrender.Viewport, a2 int) {
	nox_xxx_tileDrawMB_481C20_A((*nox_draw_viewport_t)(vp.C()), C.int(a2))
}
func Nox_xxx_tileDrawMB_481C20_B(vp *noxrender.Viewport, a2 int) {
	nox_xxx_tileDrawMB_481C20_B((*nox_draw_viewport_t)(vp.C()), C.int(a2))
}
func Nox_xxx_tileCheckRedrawMB_482570(vp *noxrender.Viewport) int {
	return int(nox_xxx_tileCheckRedrawMB_482570((*nox_draw_viewport_t)(vp.C())))
}
