//go:build !server

package cfixture

/*
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "GAME3_1.h"
#include "client__draw__drawrays.h"
#include "client__draw__fx.h"

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

import "unsafe"

func ClientTransientRayPayloadUsesNativeUnion49BDD0() bool {
	return C.nox_client_transient_ray_payload_uses_native_union() != 0
}

func ClientOrbPayloadUsesNativeUnion499490() bool {
	return C.nox_client_orb_payload_uses_native_union() != 0
}

func ClientManaBombPayloadUsesNativeUnion499520() bool {
	return C.nox_client_mana_bomb_payload_uses_native_union() != 0
}

func ClientBallisticFXPayloadUsesNativeUnion499610() bool {
	return C.nox_client_ballistic_fx_payload_uses_native_union() != 0
}

func ClientFallingSparkPayloadUsesNativeUnion499950() bool {
	return C.nox_client_falling_spark_payload_uses_native_union() != 0
}

func ClientWallDrawImageArgumentNative473C10() bool {
	return C.nox_client_wall_draw_image_arg_is_native() != 0
}

func ClientWallImageAddressRoundTrip473C10(addr uintptr) uintptr {
	return uintptr(C.nox_client_wall_image_addr_roundtrip(C.uintptr_t(addr)))
}

func ClientWallScreenPosition473C10(vp, wall unsafe.Pointer) (int, int) {
	var x, y C.int
	C.nox_client_wall_screen_position_473C10(
		(*C.nox_draw_viewport_t)(vp),
		(*C.uchar)(wall),
		&x,
		&y,
	)
	return int(x), int(y)
}
