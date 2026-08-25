#include <stdio.h>

#include "client__draw__debugdraw.h"
#include "client__draw__parse__parse.h"
#include "client__drawable__drawable.h"
#include "client__drawable__update__dball.h"
#include "client__drawable__update__drainup.h"
#include "client__drawable__update__healup.h"
#include "client__drawable__update__manabomb.h"
#include "client__drawable__update__mmislup.h"
#include "client__drawable__update__mtailup.h"
#include "client__drawable__update__sparklup.h"
#include "client__drawable__update__telwake.h"
#include "client__drawable__update__vortexup.h"

#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_2.h"
#include "GAME2.h"
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "GAME3_1.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4.h"
#include "GAME4_3.h"
#include "client__gui__window.h"
#include "client__video__draw_common.h"
#include "common__strman.h"
#include "operators.h"

extern uint64_t qword_581450_9544;
extern uint64_t qword_581450_9552;
extern uint32_t dword_5d4594_251572;

extern int nox_parse_thing_draw_funcs_cnt;

//----- (0044B2D0) --------------------------------------------------------
bool nox_parse_thing_light_dir(nox_thing* obj, nox_memfile* f, char* attr_value) {
	int deg = 0;
	if (sscanf(attr_value, "%d", &deg) != 1) {
		return 0;
	}
	if (deg < 0 || deg >= 360) {
		return 0;
	}
	obj->light_dir = (long long)((double)deg * *getMemDoublePtr(0x581450, 9560) * *(double*)&qword_581450_9552 +
								 *(double*)&qword_581450_9544);
	obj->field_10 = 0;
	return 1;
}

//----- (0044B330) --------------------------------------------------------
bool nox_parse_thing_light_penumbra(nox_thing* obj, nox_memfile* f, char* attr_value) {
	int deg = 0;
	if (sscanf(attr_value, "%d", &deg) != 1) {
		return 0;
	}
	if (deg < 0 || deg >= 180) {
		return 0;
	}
	obj->light_penumbra = (long long)((double)deg * *getMemDoublePtr(0x581450, 9560) * *(double*)&qword_581450_9552 +
									  *(double*)&qword_581450_9544);
	return 1;
}

//----- (004B5C40) --------------------------------------------------------
bool nox_parse_thing_client_update(nox_thing* obj, nox_memfile* f, char* attr_value) {
	static const struct {
		const char* name;
		void* function;
	} updates[] = {
		{"ColorLightUpdateDraw", (void*)nox_xxx_updDrawColorlight_4CE390},
		{"CloudUpdateDraw", (void*)nox_xxx_updDrawCloud_4CE1D0},
		{"SmallCloudUpdateDraw", (void*)sub_4CE360},
		{"DeathBallUpdateDraw", (void*)nox_xxx_updDrawDBall_4CDF80},
		{"DeathBallFragmentUpdateDraw", (void*)sub_4CE0A0},
		{"DeathBallChargeUpdateDraw", (void*)nox_xxx_updDrawDBallCharge_4CE0C0},
		{"MagicUpdateDraw", (void*)nox_xxx_updDrawMagic_4CDD80},
		{"SparkleTrailUpdateDraw", (void*)nox_xxx_updDrawSparkleTrail_4CDBF0},
		{"MagicMissileUpdateDraw", (void*)nox_xxx_updDrawMagicMissile_4CD9E0},
		{"TeleportWakeUpdateDraw", (void*)nox_xxx_updDrawTeleportWake_4CD8D0},
		{"DrainManaUpdateDraw", (void*)sub_4CD690},
		{"HealUpdateDraw", (void*)sub_4CD450},
		{"CharmUpdateDraw", (void*)sub_4CD400},
		{"TitanFireballUpdateDraw", (void*)sub_4CCE70},
		{"StrongFireballUpdateDraw", (void*)sub_4CD090},
		{"FireballUpdateDraw", (void*)sub_4CD0C0},
		{"WeakFireballUpdateDraw", (void*)sub_4CD0F0},
		{"PitifulFireballUpdateDraw", (void*)sub_4CD120},
		{"FistUpdateDraw", (void*)nox_xxx_updDrawFist_4CCDB0},
		{"MeteorUpdateDraw", (void*)sub_4CCD00},
		{"UndeadKillerClientUpdate", (void*)nox_xxx_updDrawUndeadKiller_4CCCF0},
		{"ManaBombChargeClientUpdate", (void*)nox_xxx_updDrawManabombCharge_4CCAC0},
		{"VortexSourceClientUpdate", (void*)nox_xxx_updDrawVortexSource_4CC950},
		{"LinearOrbUpdateDraw", (void*)sub_4CA650},
		{"MonsterGeneratorUpdateDraw", (void*)nox_xxx_updDrawMonsterGen_4BC920},
	};
	char* name = strtok(attr_value, " \t\n\r");
	if (!name) {
		return false;
	}
	for (int i = 0; i < sizeof(updates) / sizeof(updates[0]); i++) {
		if (strcmp(updates[i].name, name) == 0) {
			obj->client_update = updates[i].function;
			return true;
		}
	}
	return false;
}

//----- (0044C500) --------------------------------------------------------
bool nox_parse_thing_pretty_image(nox_thing* obj, nox_memfile* f, char* attr_value) {
	char v10[128];

	const uint32_t known_idx = nox_memfile_read_u32(f);
	if (known_idx != -1) {
		obj->pretty_image = nox_xxx_readImgMB_42FAA0(known_idx, 0, v10);
		return 1;
	}

	// TODO: After cleanup: This branch appears to never be taken. Figure out what these values are.
	const int v8 = nox_memfile_read_u8(f);
	const int n = nox_memfile_read_u8(f);
	nox_memfile_read(v10, 1u, n, f);
	obj->pretty_image = nox_xxx_readImgMB_42FAA0(known_idx, v8, v10);
	return 1;
}

//----- (00485CF0) --------------------------------------------------------
extern uint32_t nox_tile_def_cnt;
extern nox_tileDef_t nox_tile_defs_arr[176];
int nox_free_tile_defs() {
	for (int i = 0; i < nox_tile_def_cnt; i++) {
		nox_tileDef_t* it = &nox_tile_defs_arr[i];
		if (it->data_32) {
			free(it->data_32);
			it->data_32 = 0;
		}
	}
	return 1;
}

//----- (00485F30) --------------------------------------------------------
int sub_485F30() {
	if (*(int*)&dword_5d4594_251572 <= 0) {
		return 1;
	}
	for (int i = 0; i < *(int*)&dword_5d4594_251572; i++) {
		if (nox_edge_images_native[i]) {
			free(nox_edge_images_native[i]);
			nox_edge_images_native[i] = NULL;
		}
		*getMemU32Ptr(0x85B3FC, 28676 + 60 * i) = 0;
	}
	return 1;
}

// GAME.EXE 004F0640 reward definition initialization is restored by
// server.RewardDefinitionsInit4F0640. Its original PE32 pointer tables cannot
// represent native ModifierEff pointers on 64-bit targets, so the sole active
// caller now uses native Go tables with exact uint32 game fields.

// Go owns the native layouts for draw kinds 5-8. Mirror only the fields that
// the destructor traverses; all image arrays and nested records live on the C
// heap. The original routine used PE32 byte strides and truncated every heap
// address through int, which is invalid on 64-bit hosts.
typedef struct nox_native_animation_vector_t {
	uint32_t size;
	void* frames[9];
	uint16_t frame_count;
	uint16_t frame_delay;
	uint32_t animation_kind;
} nox_native_animation_vector_t;

typedef struct nox_native_animation_state_draw_data_t {
	uint32_t size;
	nox_native_animation_vector_t anim[3];
} nox_native_animation_state_draw_data_t;

typedef struct nox_native_monster_draw_data_t {
	uint32_t size;
	nox_native_animation_vector_t anim[16];
} nox_native_monster_draw_data_t;

typedef struct nox_native_player_equip_animation_t {
	uint32_t size;
	void* frames[9];
} nox_native_player_equip_animation_t;

typedef struct nox_native_player_animation_t {
	nox_native_animation_vector_t base;
	nox_native_player_equip_animation_t* naked;
	nox_native_player_equip_animation_t* armor[26];
	nox_native_player_equip_animation_t* weapon[27];
} nox_native_player_animation_t;

typedef struct nox_native_player_draw_data_t {
	uint32_t size;
	nox_native_player_animation_t anim[55];
} nox_native_player_draw_data_t;

_Static_assert(sizeof(nox_native_animation_vector_t) == (sizeof(void*) == 4 ? 48 : 88),
	"wrong native animation-vector size");
_Static_assert(offsetof(nox_native_animation_vector_t, frames) == (sizeof(void*) == 4 ? 4 : 8),
	"wrong native animation-vector frame offset");
_Static_assert(sizeof(nox_native_animation_state_draw_data_t) == (sizeof(void*) == 4 ? 148 : 272),
	"wrong native animation-state data size");
_Static_assert(sizeof(nox_native_monster_draw_data_t) == (sizeof(void*) == 4 ? 772 : 1416),
	"wrong native monster data size");
_Static_assert(sizeof(nox_native_player_equip_animation_t) == (sizeof(void*) == 4 ? 40 : 80),
	"wrong native player-equipment animation size");
_Static_assert(sizeof(nox_native_player_animation_t) == (sizeof(void*) == 4 ? 264 : 520),
	"wrong native player-animation size");
_Static_assert(sizeof(nox_native_player_draw_data_t) == (sizeof(void*) == 4 ? 14524 : 28608),
	"wrong native player data size");

//----- (0044C780) --------------------------------------------------------
static void nox_xxx_draw_44C780(void* frames[9]) {
	for (int i = 0; i < 9; i++) {
		// Direction 4 is unused by the original eight-direction loader.
		if (i != 4 && frames[i]) {
			free(frames[i]);
			frames[i] = NULL;
		}
	}
}

static void nox_free_player_equip_animation(nox_native_player_equip_animation_t* anim) {
	if (!anim) {
		return;
	}
	nox_xxx_draw_44C780(anim->frames);
	free(anim);
}

//----- (0044C7B0) --------------------------------------------------------
static void sub_44C7B0(nox_native_player_draw_data_t* data) {
	for (int i = 0; i < 55; i++) {
		nox_native_player_animation_t* anim = &data->anim[i];
		nox_xxx_draw_44C780(anim->base.frames);
		nox_free_player_equip_animation(anim->naked);
		for (int j = 0; j < 26; j++) {
			nox_free_player_equip_animation(anim->armor[j]);
		}
		for (int j = 0; j < 27; j++) {
			nox_free_player_equip_animation(anim->weapon[j]);
		}
	}
}

//----- (0044C650) --------------------------------------------------------
void nox_xxx_draw_44C650_free_kind(void* lpMem, int kind) {
	void** v7 = 0;
	int v8 = 0;

	switch (kind) {
	case 2:
	case 3:
		if (((nox_static_random_draw_data_t*)lpMem)->images) {
			free(((nox_static_random_draw_data_t*)lpMem)->images);
		}
		free(lpMem);
		break;
	case 4:
		v7 = (void**)((nox_cond_animate_draw_data_t*)lpMem)->images;
		v8 = 5;
		do {
			if (*v7) {
				free(*v7);
			}
			++v7;
			--v8;
		} while (v8);
		free(lpMem);
		break;
	case 5:
		nox_xxx_draw_44C780(((nox_native_animation_vector_t*)lpMem)->frames);
		free(lpMem);
		break;
	case 6:
		sub_44C7B0((nox_native_player_draw_data_t*)lpMem);
		free(lpMem);
		break;
	case 7:
		for (int i = 0; i < 16; i++) {
			nox_xxx_draw_44C780(((nox_native_monster_draw_data_t*)lpMem)->anim[i].frames);
		}
		free(lpMem);
		break;
	case 8:
		for (int i = 0; i < 3; i++) {
			nox_xxx_draw_44C780(((nox_native_animation_state_draw_data_t*)lpMem)->anim[i].frames);
		}
		free(lpMem);
		break;
	default:
		free(lpMem);
		break;
	}
}
