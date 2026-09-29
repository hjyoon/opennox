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
static int nox_client_transient_ray_add_addr(uintptr_t addr) {
	return nox_client_transient_ray_add((nox_drawable*)addr);
}
static uintptr_t nox_client_transient_ray_at_addr(size_t index) {
	return (uintptr_t)nox_client_transient_ray_at(index);
}
static int nox_client_transient_ray_contains_addr(uintptr_t addr) {
	return nox_client_transient_ray_contains((nox_drawable*)addr);
}
static void nox_client_ray_lightning_particles(int from_x, int from_y, int to_x, int to_y) {
	int2 from = {from_x, from_y};
	int2 to = {to_x, to_y};
	nox_xxx_makeLightningParticles_4999D0((int)dword_5d4594_1200776, &from, &to);
}
static void nox_client_ray_plasma_particles(int x, int y) {
	nox_xxx_drawEnergyBolt_499710(x, y, 10, (int)dword_5d4594_1200776);
}
static int nox_client_transient_ray_payload_uses_native_union(void) {
	nox_drawable dr;
	const unsigned char endpoints[8] = {0x34, 0x12, 0x78, 0x56, 0xbc, 0x9a, 0xf0, 0xde};
	memset(&dr, 0xa5, sizeof(dr));
	nox_client_transient_ray_set_payload(&dr, endpoints);
	const unsigned char* ray = (const unsigned char*)&dr.union_u32[0];
	if (ray[0] != 0 || memcmp(ray + 5, endpoints, sizeof(endpoints)) != 0) {
		return 0;
	}
#if UINTPTR_MAX > UINT32_MAX
	if (((const unsigned char*)&dr)[432] != 0xa5) {
		return 0;
	}
#endif
	return 1;
}
static int nox_client_orb_payload_uses_native_union(void) {
	nox_drawable dr;
	const uint16_t destination[2] = {0x1234, 0x5678};
	memset(&dr, 0xa5, sizeof(dr));
	dr.field_92 = &dr;
#if UINTPTR_MAX > UINT32_MAX
	unsigned char pe32_payload_region[15];
	memcpy(pe32_payload_region, ((const unsigned char*)&dr) + 432, sizeof(pe32_payload_region));
#endif
	nox_client_orb_set_payload_499490(&dr, destination, 0x9a, 0xbc, 0xde);
	const unsigned char* payload = (const unsigned char*)&dr.union_u32[0];
	const unsigned char expected[4] = {0x34, 0x12, 0x78, 0x56};
	if (memcmp(payload, expected, sizeof(expected)) != 0 || payload[11] != 0x9a || payload[12] != 0xbc ||
		payload[13] != 0xde || payload[14] != 0xde || dr.field_92 != &dr) {
		return 0;
	}
#if UINTPTR_MAX > UINT32_MAX
	if (memcmp(((const unsigned char*)&dr) + 432, pe32_payload_region, sizeof(pe32_payload_region)) != 0) {
		return 0;
	}
#endif
	return 1;
}
static int nox_client_mana_bomb_payload_uses_native_union(void) {
	nox_drawable dr;
	const int16_t path[4] = {100, -200, 130, -160};
	memset(&dr, 0xa5, sizeof(dr));
	dr.field_92 = &dr;
#if UINTPTR_MAX > UINT32_MAX
	unsigned char pe32_payload_region[15];
	memcpy(pe32_payload_region, ((const unsigned char*)&dr) + 432, sizeof(pe32_payload_region));
#endif
	nox_client_mana_bomb_orb_set_payload_499520(&dr, path, 0x7a, 1, 9, 4);
	const unsigned char* payload = (const unsigned char*)&dr.union_u32[0];
	uint16_t distance = 0;
	memcpy(&distance, payload + 8, sizeof(distance));
	if (memcmp(payload, path, sizeof(path)) != 0 || distance != 50 || payload[10] != 0x7a ||
		payload[11] != 1 || payload[12] != 9 || payload[13] != 4 || payload[14] != 4 ||
		dr.field_116 != (void*)sub_4CA720 || (dr.field_127 & 0xffffu) != 0x7a ||
		(dr.field_127 & 0xffff0000u) != 0xa5a50000u || dr.field_92 != &dr) {
		return 0;
	}
#if UINTPTR_MAX > UINT32_MAX
	if (memcmp(((const unsigned char*)&dr) + 432, pe32_payload_region, sizeof(pe32_payload_region)) != 0) {
		return 0;
	}
#endif
	return 1;
}
static int nox_client_ballistic_fx_payload_uses_native_union(void) {
	nox_drawable dr;
	memset(&dr, 0xa5, sizeof(dr));
	dr.pos.x = 0x1234;
	dr.pos.y = 0x2345;
	dr.field_92 = &dr;
#if UINTPTR_MAX > UINT32_MAX
	unsigned char pe32_payload_region[20];
	memcpy(pe32_payload_region, ((const unsigned char*)&dr) + 432, sizeof(pe32_payload_region));
#endif
	nox_client_ballistic_fx_set_state_499610(&dr, 0x345678, 123, 456, 0x9a);
	if (dr.union_u32[0] != 0x1234000 || dr.union_u32[1] != 0x2345000 ||
		dr.union_u32[2] != 0x345678 || dr.union_u32[3] != 123 || dr.union_u32[4] != 456 ||
		dr.field_74_4 != 0x9a || dr.field_92 != &dr) {
		return 0;
	}
#if UINTPTR_MAX > UINT32_MAX
	if (memcmp(((const unsigned char*)&dr) + 432, pe32_payload_region, sizeof(pe32_payload_region)) != 0) {
		return 0;
	}
#endif
	return 1;
}
static int nox_client_falling_spark_payload_uses_native_union(void) {
	nox_drawable dr;
	const int2 origin = {0x12345678, -0x1234567};
	memset(&dr, 0xa5, sizeof(dr));
	dr.field_92 = &dr;
#if UINTPTR_MAX > UINT32_MAX
	unsigned char pe32_payload_region[11];
	memcpy(pe32_payload_region, ((const unsigned char*)&dr) + 432, sizeof(pe32_payload_region));
#endif
	nox_client_falling_spark_set_payload_499950(&dr, &origin, 0x4567, -9, 8);
	const unsigned char* payload = (const unsigned char*)&dr.union_u32[0];
	uint16_t payload_z = 0;
	memcpy(&payload_z, payload + 8, sizeof(payload_z));
	if (memcmp(payload, &origin, sizeof(origin)) != 0 || payload_z != 0x4567 || payload[10] != 8 ||
		dr.z != 0x4567 || dr.field_26_1 != 0 || dr.vel_z != -9 || dr.field_92 != &dr) {
		return 0;
	}
#if UINTPTR_MAX > UINT32_MAX
	if (memcmp(((const unsigned char*)&dr) + 432, pe32_payload_region, sizeof(pe32_payload_region)) != 0) {
		return 0;
	}
#endif
	return 1;
}
static int nox_client_wall_draw_image_arg_is_native(void) {
	typedef int* (*wall_edge_draw_func)(nox_video_bag_image_t*, int, int, int*, int*, int, int, int, int, int);
	return _Generic(&nox_xxx_edgeDraw_480EF0, wall_edge_draw_func: 1, default: 0);
}
static uintptr_t nox_client_wall_image_addr_roundtrip(uintptr_t addr) {
	nox_video_bag_image_t* image = (nox_video_bag_image_t*)addr;
	return (uintptr_t)image;
}
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
	C.nox_xxx_clientDrawAll_436100_draw_A()
}

func Nox_xxx_drawMinimapAndLines_4738E0() {
	C.nox_xxx_drawMinimapAndLines_4738E0()
}

func Nox_xxx_clientDrawAll_436100_draw_B() {
	C.nox_xxx_clientDrawAll_436100_draw_B()
}

func Sub_436F50() {
	C.sub_436F50()
}

func Sub_437100() {
	C.sub_437100()
}

func Sub_470DE0() {
	C.sub_470DE0()
}

func Nox_xxx_clientEnumHover_476FA0() {
	C.nox_xxx_clientEnumHover_476FA0()
}

func Nox_xxx_spriteDeleteSomeList_49C4B0() {
	C.nox_xxx_spriteDeleteSomeList_49C4B0()
}

func clientTransientRayAddAddress49BDD0(addr uintptr) bool {
	return C.nox_client_transient_ray_add_addr(C.uintptr_t(addr)) != 0
}

func clientTransientRayCount49BDD0() int {
	return int(C.nox_client_transient_ray_count())
}

func clientTransientRayAddress49BDD0(index int) uintptr {
	return uintptr(C.nox_client_transient_ray_at_addr(C.size_t(index)))
}

func clientTransientRayContainsAddress49BDD0(addr uintptr) bool {
	return C.nox_client_transient_ray_contains_addr(C.uintptr_t(addr)) != 0
}

func clientTransientRayClear49BDD0() {
	C.nox_client_transient_ray_clear()
}

func clientTransientRayPayloadUsesNativeUnion49BDD0() bool {
	return C.nox_client_transient_ray_payload_uses_native_union() != 0
}

// Nox_xxx_netDrawRays_49BDD0 keeps the legacy drawable construction behind a
// fixed-size packet boundary. The C routine never retains the packet bytes;
// the created drawable itself is tracked in native-width storage.
func Nox_xxx_netDrawRays_49BDD0(data [9]byte) {
	C.nox_xxx_netDrawRays_49BDD0((*C.uchar)(unsafe.Pointer(&data[0])))
}

func Nox_xxx_makeRayLightningParticles_49BDD0(from, to image.Point) {
	C.nox_client_ray_lightning_particles(C.int(from.X), C.int(from.Y), C.int(to.X), C.int(to.Y))
}

func Nox_xxx_makeRayPlasmaParticles_49BDD0(to image.Point) {
	C.nox_client_ray_plasma_particles(C.int(to.X), C.int(to.Y))
}

func clientOrbPayloadUsesNativeUnion499490() bool {
	return C.nox_client_orb_payload_uses_native_union() != 0
}

func clientManaBombPayloadUsesNativeUnion499520() bool {
	return C.nox_client_mana_bomb_payload_uses_native_union() != 0
}

func clientBallisticFXPayloadUsesNativeUnion499610() bool {
	return C.nox_client_ballistic_fx_payload_uses_native_union() != 0
}

func clientFallingSparkPayloadUsesNativeUnion499950() bool {
	return C.nox_client_falling_spark_payload_uses_native_union() != 0
}

func clientWallDrawImageArgumentNative473C10() bool {
	return C.nox_client_wall_draw_image_arg_is_native() != 0
}

func clientWallImageAddressRoundTrip473C10(addr uintptr) uintptr {
	return uintptr(C.nox_client_wall_image_addr_roundtrip(C.uintptr_t(addr)))
}

func clientWallScreenPosition473C10(vp *noxrender.Viewport, wall *server.Wall) (int, int) {
	var x, y C.int
	C.nox_client_wall_screen_position_473C10(
		(*C.nox_draw_viewport_t)(vp.C()),
		(*C.uchar)(wall.C()),
		&x,
		&y,
	)
	return int(x), int(y)
}

func Sub_49BBC0() {
	C.sub_49BBC0()
}

func Nox_xxx_polygonDrawColor_421B80() {
	C.nox_xxx_polygonDrawColor_421B80()
}

func Nox_xxx_cliToggleObsWindow_4357A0() {
	C.nox_xxx_cliToggleObsWindow_4357A0()
}

func Nox_xxx_motd_4467F0() {
	C.nox_xxx_motd_4467F0()
}

func Sub_42EBA0() int {
	return int(C.sub_42EBA0())
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
	return int(C.nox_xxx_drawAllMB_475810_draw_B((*nox_draw_viewport_t)(vp.C())))
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
		(*C.nox_draw_viewport_t)(vp.C()),
		(*C.uint16_t)(unsafe.Pointer(&ray[0])),
		&cfrom, &cto,
	)
	return image.Pt(int(cfrom.field_0), int(cfrom.field_4)), image.Pt(int(cto.field_0), int(cto.field_4))
}
func Nox_xxx_drawWalls_473C10(vp *noxrender.Viewport, a2 *server.Wall) {
	C.nox_xxx_drawWalls_473C10((*nox_draw_viewport_t)(vp.C()), a2.C())
}
func Sub_476080(a1 unsafe.Pointer) int {
	return int(C.sub_476080((*C.uchar)(a1)))
}
func Sub_459DB0(dr *client.Drawable) int {
	return int(C.sub_459DB0((*nox_drawable)(dr.C())))
}
func Sub_472540(dr *client.Drawable) int {
	return int(C.sub_472540((*nox_drawable)(dr.C())))
}
func Sub_49A6A0(vp *noxrender.Viewport, dr *client.Drawable) {
	C.sub_49A6A0((*nox_draw_viewport_t)(vp.C()), (*nox_drawable)(dr.C()))
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
	C.nox_xxx_tileDrawImpl_4826A0((*nox_draw_viewport_t)(vp.C()))
}
func Nox_xxx_tileDrawMB_481C20_A(vp *noxrender.Viewport, a2 int) {
	C.nox_xxx_tileDrawMB_481C20_A((*nox_draw_viewport_t)(vp.C()), C.int(a2))
}
func Nox_xxx_tileDrawMB_481C20_B(vp *noxrender.Viewport, a2 int) {
	C.nox_xxx_tileDrawMB_481C20_B((*nox_draw_viewport_t)(vp.C()), C.int(a2))
}
func Nox_xxx_tileCheckRedrawMB_482570(vp *noxrender.Viewport) int {
	return int(C.nox_xxx_tileCheckRedrawMB_482570((*nox_draw_viewport_t)(vp.C())))
}
