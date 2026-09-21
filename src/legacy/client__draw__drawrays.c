#include "client__draw__drawrays.h"
#include "common__random.h"

#include "GAME2.h"
#include "client__draw__fx.h"
extern uint32_t dword_5d4594_1304328;

// The PE32 client kept these temporary drawable pointers in uint32 memmap
// slots. That truncates every normal pointer on 64-bit hosts. Keep the list in
// native storage; the mirrored count remains for legacy diagnostics only.
static nox_drawable* nox_client_transient_rays[NOX_CLIENT_TRANSIENT_RAY_CAPACITY];
static size_t nox_client_transient_rays_count;

size_t nox_client_transient_ray_count(void) { return nox_client_transient_rays_count; }

nox_drawable* nox_client_transient_ray_at(size_t index) {
	if (index >= nox_client_transient_rays_count) {
		return NULL;
	}
	return nox_client_transient_rays[index];
}

int nox_client_transient_ray_contains(const nox_drawable* dr) {
	for (size_t i = 0; i < nox_client_transient_rays_count; ++i) {
		if (nox_client_transient_rays[i] == dr) {
			return 1;
		}
	}
	return 0;
}

void nox_client_transient_ray_clear(void) {
	memset(nox_client_transient_rays, 0, sizeof(nox_client_transient_rays));
	nox_client_transient_rays_count = 0;
	*getMemU32Ptr(0x5D4594, 1304308) = 0;
}

bool nox_client_transient_ray_add(nox_drawable* dr) {
	if (!dr || nox_client_transient_rays_count >= NOX_CLIENT_TRANSIENT_RAY_CAPACITY) {
		return false;
	}
	nox_client_transient_rays[nox_client_transient_rays_count++] = dr;
	*getMemU32Ptr(0x5D4594, 1304308) = (uint32_t)nox_client_transient_rays_count;
	return true;
}

static uint16_t nox_client_ray_read_u16(const unsigned char* data) {
	uint16_t value;
	memcpy(&value, data, sizeof(value));
	return value;
}

void nox_client_transient_ray_set_payload(nox_drawable* dr, const unsigned char* endpoints) {
	uint8_t* ray = (uint8_t*)&dr->union_u32[0];
	ray[0] = 0;
	memcpy(ray + 5, endpoints, 8);
}

//----- (0049BDD0) --------------------------------------------------------
nox_drawable* nox_xxx_netDrawRays_49BDD0(unsigned char* data) {
	if (!data || nox_client_transient_rays_count >= NOX_CLIENT_TRANSIENT_RAY_CAPACITY) {
		return NULL;
	}
	if (!*getMemU32Ptr(0x5D4594, 1304316)) {
		*getMemU32Ptr(0x5D4594, 1304316) = nox_xxx_getTTByNameSpriteMB_44CFC0("DynamicLightning");
		*getMemU32Ptr(0x5D4594, 1304320) = nox_xxx_getTTByNameSpriteMB_44CFC0("DynamicChainLightning");
		*getMemU32Ptr(0x5D4594, 1304324) = nox_xxx_getTTByNameSpriteMB_44CFC0("DynamicEnergyBolt");
		*getMemU32Ptr(0x5D4594, 1304348) = nox_xxx_getTTByNameSpriteMB_44CFC0("GreenZap");
		dword_5d4594_1304328 = nox_xxx_getTTByNameSpriteMB_44CFC0("OrbRay");
		*getMemU32Ptr(0x5D4594, 1304332) = nox_xxx_getTTByNameSpriteMB_44CFC0("PlasmaRay");
		*getMemU32Ptr(0x5D4594, 1304336) = nox_xxx_getTTByNameSpriteMB_44CFC0("DrainManaOrb");
		*getMemU32Ptr(0x5D4594, 1304340) = nox_xxx_getTTByNameSpriteMB_44CFC0("HealOrb");
		*getMemU32Ptr(0x5D4594, 1304344) = nox_xxx_getTTByNameSpriteMB_44CFC0("CharmOrb");
	}

	uint16_t endpoints[4];
	for (size_t i = 0; i < 4; ++i) {
		endpoints[i] = nox_client_ray_read_u16(data + 1 + 2 * i);
	}
	int x = endpoints[0] + ((int)endpoints[2] - endpoints[0]) / 2;
	int y = endpoints[1] + ((int)endpoints[3] - endpoints[1]) / 2;
	int type;
	switch (data[0]) {
	case 0x7D: // MSG_FX_PLASMA
		type = *getMemU32Ptr(0x5D4594, 1304332);
		break;
	case 0x8C: // MSG_FX_LIGHTNING
		type = *getMemU32Ptr(0x5D4594, 1304316);
		break;
	case 0x8D: // MSG_FX_ENERGY_BOLT
		type = *getMemU32Ptr(0x5D4594, 1304324);
		break;
	case 0x8E: // MSG_FX_CHAIN_LIGHTNING_BOLT
		type = *getMemU32Ptr(0x5D4594, 1304320);
		break;
	case 0x8F: { // MSG_FX_DRAIN_MANA
		type = dword_5d4594_1304328;
		char radius = nox_common_randomIntMinMax_415FF0(6, 12, "C:\\NoxPost\\src\\Client\\Draw\\drawrays.c", 137);
		if (nox_common_randomIntMinMax_415FF0(0, 100, "C:\\NoxPost\\src\\Client\\Draw\\drawrays.c", 140) < 50) {
			int dx = nox_common_randomIntMinMax_415FF0(-20, 20, "C:\\NoxPost\\src\\Client\\Draw\\drawrays.c", 142);
			int dy = nox_common_randomIntMinMax_415FF0(-20, 20, "C:\\NoxPost\\src\\Client\\Draw\\drawrays.c", 143);
			sub_499490(*getMemIntPtr(0x5D4594, 1304336), endpoints, dx, dy, radius, 0);
		}
		break;
	}
	case 0x90: { // MSG_FX_CHARM
		type = dword_5d4594_1304328;
		char radius = nox_common_randomIntMinMax_415FF0(6, 12, "C:\\NoxPost\\src\\Client\\Draw\\drawrays.c", 101);
		if (nox_common_randomIntMinMax_415FF0(0, 100, "C:\\NoxPost\\src\\Client\\Draw\\drawrays.c", 104) < 50) {
			int dx = nox_common_randomIntMinMax_415FF0(-20, 20, "C:\\NoxPost\\src\\Client\\Draw\\drawrays.c", 106);
			int dy = nox_common_randomIntMinMax_415FF0(-20, 20, "C:\\NoxPost\\src\\Client\\Draw\\drawrays.c", 107);
			sub_499490(*getMemIntPtr(0x5D4594, 1304344), endpoints, dx, dy, radius, 0);
		}
		uint16_t reversed[4] = {endpoints[2], endpoints[3], endpoints[0], endpoints[1]};
		if (nox_common_randomIntMinMax_415FF0(0, 100, "C:\\NoxPost\\src\\Client\\Draw\\drawrays.c", 112) < 50) {
			int dx = nox_common_randomIntMinMax_415FF0(-20, 20, "C:\\NoxPost\\src\\Client\\Draw\\drawrays.c", 114);
			int dy = nox_common_randomIntMinMax_415FF0(-20, 20, "C:\\NoxPost\\src\\Client\\Draw\\drawrays.c", 115);
			sub_499490(*getMemIntPtr(0x5D4594, 1304344), reversed, dx, dy, radius, 0);
		}
		break;
	}
	case 0x91: { // MSG_FX_GREATER_HEAL
		type = dword_5d4594_1304328;
		char radius = nox_common_randomIntMinMax_415FF0(6, 12, "C:\\NoxPost\\src\\Client\\Draw\\drawrays.c", 123);
		if (nox_common_randomIntMinMax_415FF0(0, 100, "C:\\NoxPost\\src\\Client\\Draw\\drawrays.c", 126) < 50) {
			int dx = nox_common_randomIntMinMax_415FF0(-20, 20, "C:\\NoxPost\\src\\Client\\Draw\\drawrays.c", 128);
			int dy = nox_common_randomIntMinMax_415FF0(-20, 20, "C:\\NoxPost\\src\\Client\\Draw\\drawrays.c", 129);
			sub_499490(*getMemIntPtr(0x5D4594, 1304340), endpoints, dx, dy, radius, 0);
		}
		break;
	}
	default:
		return NULL;
	}

	nox_drawable* result = nox_xxx_spriteLoadAdd_45A360_drawable(type, x, y);
	if (!result) {
		return NULL;
	}
	nox_client_transient_ray_set_payload(result, data + 1);
	if (!nox_client_transient_ray_add(result)) {
		nox_xxx_spriteDeleteStatic_45A4E0_drawable(result);
		return NULL;
	}
	return result;
}
