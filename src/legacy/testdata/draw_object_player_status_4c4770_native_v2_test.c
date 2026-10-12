// The Go test inserts the UNMODIFIED production 4C4770 body between this
// fixture preamble and its main. Only external world/render services are
// bounded stubs; player records use the real native defs.h layout.
#define _Static_assert(...)
#include "defs.h"
#undef _Static_assert
#include "common__system__team.h"
#include <math.h>
#include <stdio.h>

_Static_assert(offsetof(nox_playerInfo, field_3680) ==
    (sizeof(void*) == 4 ? 3680 : 4976), "native player status offset");

static nox_playerInfo* probe_info;
static nox_playerInfo* probe_viewer;
static nox_drawable* probe_subject;
static nox_drawable* probe_local;
static int probe_invisible;
static int probe_draws, probe_use_alpha, probe_alpha, probe_draw_alpha;
static uint32_t probe_zero32;
static uint16_t probe_zero16;
static int32_t probe_zero_int;
static uint8_t probe_light[12] = {255,0,0,0,255,0,0,0,255,0,0,0};
static uint32_t dword_5d4594_1321520 = 9999;
static uint32_t nox_player_netCode_85319C = 10;
static uintptr_t dword_8531A0_2576;
static uint32_t dword_8531A0_2572;
static uint32_t nox_color_blue_2650684, nox_color_white_2523948;
static int nox_win_width = 640;

static int probe_buff(nox_drawable* dr, int buff) {
    return dr == probe_subject && buff == 0 && probe_invisible;
}
static void probe_image(void* image, int x, int y) {
    (void)image; (void)x; (void)y;
    probe_draws++;
    probe_draw_alpha = probe_use_alpha ? probe_alpha : 255;
}

#undef getMemU32Ptr
#undef getMemU16Ptr
#undef getMemIntPtr
#undef getMemAt
#undef getMemPtr
#define getMemU32Ptr(...) (&probe_zero32)
#define getMemU16Ptr(...) (&probe_zero16)
#define getMemIntPtr(...) (&probe_zero_int)
#define getMemAt(...) ((void*)probe_light)
#define getMemPtr(...) ((void*)probe_local)
#define nox_xxx_getTTByNameSpriteMB_44CFC0(...) (9999)
#define nox_common_playerInfoGetByID_417040(...) ((void*)probe_info)
#define nox_xxx_objGetTeamByNetCode_418C80(...) ((nox_object_team_t*)0)
#define nox_xxx_servCompareTeams_419150(...) (0)
#define sub_4C5630(...) (0)
#define sub_4C42A0(...) (1)
#define sub_469920(...) ((void*)probe_light)
#define nox_client_drawable_testBuff_4356C0(dr, buff) probe_buff(dr, buff)
#define nox_xxx_draw_434600(...) ((void)0)
#define nox_draw_setColorMultAndIntensity_433E40(...) ((void)0)
#define sub_4345F0(...) ((void)0)
#define nox_draw_setColorMultAndIntensityRGB_433CD0(...) ((void)0)
#define gameFrame() (100)
#define gameFPS() (30)
#define sub_4C4EC0(...) (128)
#define nox_client_drawEnableAlpha_434560(value) (probe_use_alpha = (value))
#define nox_client_drawSetAlpha_434580(value) (probe_alpha = (uint8_t)(value))
#define nox_common_gameFlags_check_40A5C0(...) (0)
#define sub_4C3B70(...) ((void)0)
#define nox_xxx_wndDraw_49F7F0(...) ((void)0)
#define sub_49F7C0_def(...) ((void)0)
#define sub_47D370(...) ((void)0)
#define nox_client_drawImageAt_47D2C0(image, x, y) probe_image(image, x, y)
#define sub_49F860(...) ((void)0)
#define nox_double2int(value) ((int)(value))

// PRODUCTION_BODY_INSERTION_POINT

int main(int argc, char** argv) {
    if (argc != 6) return 90;
    uint32_t status = (uint32_t)strtoul(argv[1], NULL, 0);
    uint8_t legacy_byte = (uint8_t)strtoul(argv[2], NULL, 0);
    int local = atoi(argv[3]);
    int moving = atoi(argv[4]);
    probe_invisible = atoi(argv[5]);
    probe_info = calloc(1, sizeof(*probe_info));
    probe_viewer = calloc(1, sizeof(*probe_viewer));
    probe_subject = calloc(1, sizeof(*probe_subject));
    probe_local = calloc(1, sizeof(*probe_local));
    if (!probe_info || !probe_viewer || !probe_subject || !probe_local) return 91;
    // This is legitimate armor memory on 64-bit. It must never select the
    // observer branch. On 32-bit the status store is the authoritative slot.
    if (sizeof(void*) != 4) ((uint8_t*)probe_info)[3680] = legacy_byte;
    probe_info->field_3680 = status;
    nox_playerInfo before = *probe_info;
    probe_subject->flags28 = 4;
    probe_subject->field_32 = local ? 10 : 20;
    probe_subject->field_27 = 713;
    probe_subject->pos.x = moving ? 8 : 0;
    probe_subject->field_5 = 1;
    probe_subject->field_10 = 0;
    probe_local->flags28 = 4;
    probe_local->field_32 = 10;
    probe_viewer->field_3680 = local ? status : 0;
    dword_8531A0_2576 = (uintptr_t)probe_viewer;
    probe_alpha = 255;
    probe_draw_alpha = -1;
    nox_draw_viewport_t viewport = {0};
    uint8_t image = 1;
    nox_xxx_drawObject_4C4770_draw(&viewport, probe_subject, &image);
    if (memcmp(&before, probe_info, sizeof(before)) != 0) return 92;
    printf("draws=%d alpha=%d status_offset=%zu pointer_bytes=%zu status=%#x legacy_byte=%#x\n",
        probe_draws, probe_draw_alpha, offsetof(nox_playerInfo, field_3680), sizeof(void*),
        probe_info->field_3680, ((uint8_t*)probe_info)[3680]);
    free(probe_info); free(probe_viewer); free(probe_subject); free(probe_local);
    return 0;
}
