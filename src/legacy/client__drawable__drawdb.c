#include "draw_allocation.h"
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
#include "floor_tile_storage.h"
#include "operators.h"

extern uint64_t qword_581450_9544;
extern uint64_t qword_581450_9552;
extern uint32_t dword_5d4594_251572;

extern int nox_parse_thing_draw_funcs_cnt;

// Preserve the address return for call-scoped Go drawables.
void nox_drawable_set_light_color_go(nox_drawable* dr, int r, int g, int b);

nox_drawable* nox_drawable_change_light_color_native(nox_drawable* dr, int r, int g, int b) {
	nox_drawable_set_light_color_go(dr, r, g, b);
	return dr;
}

void nox_drawable_set_packed_light_color_go(uint32_t* light, int r, int g, int b);

uint32_t* nox_xxx_spriteChangeLightColor_484BE0(uint32_t* light, int r, int g, int b) {
	nox_drawable_set_packed_light_color_go(light, r, g, b);
	return light;
}

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

// GAME.EXE 004F0640 reward definition initialization is restored by
// server.RewardDefinitionsInit4F0640. Its original PE32 pointer tables cannot
// represent native ModifierEff pointers on 64-bit targets, so the sole active
// caller now uses native Go tables with exact uint32 game fields.
