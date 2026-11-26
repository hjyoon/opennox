#include "GAME2_1.h"
#include "GAME2_2.h"
#include "client__system__parsecmd.h"
#include "client__gui__guitrade.h"
#include "client__gui__guisumn.h"
#include "client__gui__guispell.h"
#include "stats_start_native.h"
#include "client__gui__servopts__playrlst.h"
#include "client__gui__servopts__access.h"
#include "client_ui_words.h"
#include "vardefs.h"

void nullsub_68();

obj_5D4594_2650668_t** ptr_5D4594_2650668 = 0;
const int ptr_5D4594_2650668_cap = 128;



void* nox_xxx_aClosewoodengat_587000_133480 = 0;
nox_list_item_t nox_gui_wol_servers_list = {0};
uintptr_t dword_5d4594_1062480 = 0;
nox_drawable* dword_5d4594_1063116 = 0;
nox_video_bag_image_t* nox_ctf_flag_team_border;
uintptr_t dword_5d4594_529316 = 0;
uint32_t dword_5d4594_1599656 = 0;
void* dword_587000_155144 = 0;
uint32_t dword_5d4594_3798804 = 0;
uint32_t dword_5d4594_832476 = 0;
uint32_t nox_game_createOrJoin_815048 = 0;
uint32_t nox_xxx_polygonNextIdx_587000_60352 = 0x1;
uint32_t dword_5d4594_816348 = 0;
uint32_t dword_5d4594_1098628 = 0;
uint32_t dword_5d4594_251572 = 0;
void* nox_edge_images_native[64] = {0};
uint32_t dword_5d4594_1107036 = 0;
void* dword_5d4594_1301848 = 0;
void* dword_5d4594_1193360 = 0;
void* dword_5d4594_1193348 = 0;
uintptr_t dword_5d4594_534808 = 0;
char* dword_5d4594_1599576 = 0;
uint32_t dword_5d4594_1599596 = 0;
void* dword_5d4594_2516344 = 0;
void* nox_alloc_pixelSpan_1301844 = 0;
uint32_t dword_5d4594_3798808 = 0;
uint32_t dword_5d4594_3835396 = 0;
uint32_t dword_5d4594_3798816 = 0;
nox_respawn_record_t* dword_5d4594_1568024 = NULL;
uint32_t dword_5d4594_2487524 = 0;
void* dword_587000_127004 = 0;
uint32_t dword_587000_93156 = 0x1;
uint32_t dword_587000_93160 = 0x1;
uintptr_t dword_5d4594_531648 = 0;
uint32_t dword_5d4594_815044 = 0;
void* nox_alloc_tradeItems_2386496 = 0;
uint32_t dword_5d4594_816340 = 0;
uint32_t dword_5d4594_816368 = 0;
void* dword_587000_93164 = 0;
uint32_t dword_5d4594_1098624 = 0;
nox_drawable* dword_5d4594_1320968 = 0;
uint64_t qword_581450_9552 = 0x40f0000000000000;
uint64_t qword_581450_9512 = 0x3ff0000000000000; // 1.0d
uint32_t dword_5d4594_2386500 = 0;
uint32_t dword_5d4594_2649712 = 0;
uint32_t dword_587000_60044 = 0xffffffff;
void* dword_5d4594_251560 = 0;
uint32_t dword_5d4594_3484 = 0;
void* nox_alloc_screenParticles_806044 = 0;
void* dword_587000_122852 = 0;
void* dword_5d4594_1301776 = 0;
uint32_t dword_5d4594_1550916 = 0;
void* dword_587000_81128 = 0;
uint32_t dword_5d4594_3798800 = 0;
uint32_t dword_5d4594_3798812 = 0;
const char nox_startup_pause_names[2][6] = {"Pause", "Pause"};
uint32_t dword_5d4594_2487248 = 0;
uint32_t dword_5d4594_588120 = 0;
void* dword_5d4594_588084 = 0;
uint32_t dword_5d4594_816372 = 0;
void* dword_5d4594_1203864 = 0;
uint32_t dword_5d4594_3835356 = 0;
uint32_t nox_client_renderBubbles_80844 = 0x1;
void* nox_alloc_chat_1197364 = 0;
nox_gui_trade_slot* dword_5d4594_1320932 = 0;
uint32_t dword_5d4594_830864 = 0;
uint32_t dword_5d4594_608316 = 0;
void* dword_5d4594_588068 = 0;
uint32_t dword_5d4594_1203840 = 0;
uint32_t dword_5d4594_741332 = 0;
unsigned int dword_5d4594_1599628 = 0;
unsigned int dword_5d4594_2489420 = 0;
nox_stats_quest_columns_native nox_stats_quest_columns = {0};
uint32_t dword_5d4594_816356 = 0;
uint32_t dword_5d4594_1203836 = 0;
void* dword_5d4594_1301780 = 0;
void* nox_alloc_respawn_1568020 = 0;
uint32_t dword_5d4594_1567928 = 0;
uint32_t dword_5d4594_1307720 = 0;
uint32_t dword_5d4594_1550912 = 0;
uint32_t dword_5d4594_1062488 = 0;
void* nox_alloc_tradeSession_2386492 = 0;
uint32_t dword_5d4594_3835352 = 0;
uint32_t dword_5d4594_825736 = 0;
uint32_t dword_5d4594_2487576 = 0;
uint32_t dword_5d4594_831084 = 0;
void* nox_alloc_friendList_1203860 = 0;
uint32_t dword_5d4594_3835348 = 0;
nox_gui_trade_slot* dword_5d4594_1320972 = 0;
uint32_t dword_5d4594_10984 = 0;
nox_drawable* dword_5d4594_1063120 = 0;
void* nox_alloc_vote_1599652 = 0;
uint32_t dword_5d4594_2487676 = 0;
uint32_t dword_5d4594_2487672 = 0;
void* nox_alloc_magicEnt_1569668 = 0;
uint32_t dword_587000_66116 = 0x1d;
uint32_t dword_5d4594_2487564 = 0;
uint32_t dword_5d4594_814548 = 0;
uint32_t nox_server_sanctuaryHelp_54276 = 0x1;
uint32_t dword_5d4594_1313740 = 0;
uintptr_t dword_5d4594_1062492 = 0;
wchar2_t* dword_5d4594_1098604 = 0;
char* nox_shop_dialog_file_1098608 = 0;
uint32_t dword_5d4594_2487584 = 0;
void* nox_alloc_healthChange_1301772 = 0;
wchar2_t* dword_5d4594_1098600 = 0;
uint32_t dword_587000_183456 = 0x67;
nox_waypoint_t* dword_5d4594_2487244 = 0;
wchar2_t* dword_5d4594_1098596 = 0;
nox_gui_trade_slot* dword_5d4594_1320936 = 0;
uint32_t dword_5d4594_831076 = 0;
uint32_t dword_5d4594_1203832 = 0;
uint32_t dword_5d4594_1321520 = 0;
uint32_t dword_5d4594_526276 = 0;
uint32_t dword_5d4594_1062484 = 0;
uint32_t dword_5d4594_2487568 = 0;
uint32_t nox_server_sendMotd_108752 = 0x1;
uint32_t dword_5d4594_1123520 = 0;
uint32_t dword_5d4594_2488720 = 0;
uint32_t dword_5d4594_1599644 = 0;
uint32_t nox_server_resetQuestMinVotes_229988 = 0x6;
uint32_t dword_5d4594_1045640 = 0;
uint32_t nox_server_kickQuestPlayerMinVotes_229992 = 0x5;
uint32_t dword_587000_183460 = 0x26;
uint32_t dword_5d4594_3835372 = 0;
uint32_t dword_5d4594_741652 = 0;
uint32_t dword_5d4594_600116 = 0;
uint32_t dword_5d4594_815052 = 0;
uint32_t dword_5d4594_741648 = 0;
nox_object_t* dword_5d4594_2491592 = 0;
uint32_t dword_5d4594_2488724 = 0;
void* dword_5d4594_1548532 = 0;
uint32_t dword_5d4594_805820 = 0;
uint32_t dword_5d4594_2489436 = 0;
uint32_t dword_5d4594_527988 = 0;
uint32_t dword_5d4594_815748 = 0;
uint32_t dword_5d4594_1045420 = 0;
uint32_t dword_5d4594_739392 = 0;
uint32_t dword_5d4594_1313788 = 0;
void* dword_5d4594_2516352 = 0;
uint32_t dword_5d4594_3835360 = 0;
uint32_t dword_5d4594_1062552 = 0;
uint32_t dword_5d4594_1200832 = 0;
void* nox_alloc_hit_2491548 = 0;
uint32_t dword_5d4594_3835368 = 0;
uint32_t nox_xxx_useAudio_587000_80772 = 0x5;
uint32_t dword_5d4594_2488620 = 0;
uint32_t dword_5d4594_2487656 = 0;
uint32_t dword_587000_122956 = 0x1;
uintptr_t dword_8531A0_2576 = 0;
uint32_t dword_8531A0_2572 = 0;

uint8_t** nox_pixbuffer_rows_3798784 = 0;

void* dword_5d4594_1096640 = 0;

void* nox_client_spriteUnderCursorXxx_1096644 = 0;

int nox_win_width = 0;
int nox_win_height = 0;
obj_5D4594_754088_t* ptr_5D4594_754088 = 0;
int ptr_5D4594_754088_cnt = 0;
obj_5D4594_754088_t* ptr_5D4594_754092 = 0;
int ptr_5D4594_754092_cnt = 0;
nox_screenParticle* nox_screenParticles_head = 0;
nox_screenParticle* dword_5d4594_806052 = 0;
nox_gui_server_ent_t* dword_5d4594_814624 = 0;
nox_gui_animation* nox_wnd_xxx_815040 = 0;
void* dword_5d4594_805984 = 0;
