#ifndef NOX_PORT_GAME2_3
#define NOX_PORT_GAME2_3

#include "defs.h"

void sub_48C580(pixel8888* a1, int num);
unsigned int sub_48C690(int a1, int a2, int a3, int a4);
unsigned int sub_48C6B0(int a1, int a2);
int nox_xxx_showObserverWindow_48CA70(int a1);
int sub_48CAD0();
int sub_48D000_initGuiKick();
int nox_xxx_guiKick_48D0A0(int a1, int a2, int* a3, int a4);
int sub_48D120();
char* nox_xxx_voteSend_48D260(wchar2_t* a1);
char* nox_xxx_netSendRenameMb_48D2D0(wchar2_t* a1);
int sub_48D340();
int sub_48D3B0();
int nox_xxx_clientVote_48D3E0();
uint32_t* sub_48D410();
int sub_48D450();
int sub_48D4A0();
int sub_48D4B0(int a1);
int sub_48D4F0(unsigned short a1, unsigned short a2);
int sub_48D560(unsigned short a1);
int sub_48D5A0(const uint8_t* data);
int sub_48D660();
int sub_48D740();
void sub_48D760();
int* sub_48D7B0();
int sub_48D800();

// The first 692 bytes preserve the original PE32 chat-bubble data layout.
// Runtime-only pointers live after that prefix so 64-bit clients do not lose
// their upper address bits while keeping the offsets used by the layout code.
typedef struct nox_chat_bubble {
	wchar2_t text[318];       // 0
	uint32_t duration_hint;   // 636
	uint32_t expire_frame;    // 640
	uint16_t drawable_x;      // 644
	uint16_t drawable_y;      // 646
	int32_t x;                // 648
	int32_t y;                // 652
	uint32_t net_code;        // 656
	uint32_t visible;         // 660
	uint32_t draw_tail;       // 664
	uint32_t legacy_drawable; // 668; reserved PE32 slot
	int32_t width;            // 672
	int32_t height;           // 676
	int32_t order;            // 680
	uint32_t legacy_next;     // 684; reserved PE32 slot
	uint32_t legacy_prev;     // 688; reserved PE32 slot
	nox_drawable* drawable;
	struct nox_chat_bubble* next;
	struct nox_chat_bubble* prev;
} nox_chat_bubble;
_Static_assert(offsetof(nox_chat_bubble, duration_hint) == 636,
	"wrong offset of nox_chat_bubble.duration_hint!");
_Static_assert(offsetof(nox_chat_bubble, expire_frame) == 640,
	"wrong offset of nox_chat_bubble.expire_frame!");
_Static_assert(offsetof(nox_chat_bubble, net_code) == 656,
	"wrong offset of nox_chat_bubble.net_code!");
_Static_assert(offsetof(nox_chat_bubble, width) == 672,
	"wrong offset of nox_chat_bubble.width!");
_Static_assert(offsetof(nox_chat_bubble, order) == 680,
	"wrong offset of nox_chat_bubble.order!");
_Static_assert(offsetof(nox_chat_bubble, drawable) >= 692,
	"native chat-bubble fields overlap the PE32 prefix!");

nox_chat_bubble* nox_xxx_netCode2ChatBubble_48D850(int a1);
void sub_48D990(nox_draw_viewport_t* a1);
void sub_48DCF0(nox_draw_viewport_t* a1);
bool sub_48E000(int4* a1, uint32_t* a2);
char sub_48E240(int a1, uint32_t* a2);
int sub_48E480(uint32_t* a1, uint32_t* a2);
int sub_48E530(int a1, int a2);
int sub_48E5C0(uint32_t* a1, int a2, int a3);
int* sub_48E6A0(char a1, uint32_t* a2, uint32_t* a3, int* a4, int* a5);
void sub_48E8E0(int a1);
void sub_48E940();
nox_drawable* nox_xxx_spriteCreate_48E970(int a1, unsigned short a2, int a3, int a4);
char* sub_4947E0(int a1);
int sub_4948B0(int a1);
int nox_xxx_netCliProcUpdateStream_494A60(unsigned char* a1, int a2, uint32_t* a3);
unsigned char* nox_xxx_netCliUpdateStream2_494C30(unsigned char* a1, int a2, int* a3);
int sub_494F00();
char* sub_494FF0();
char* sub_495020(int a1);
int sub_495060(int a1, short a2, short a3);
int sub_4950C0(int a1);
int sub_4950F0(int a1, char a2);
int sub_495120(int a1, short a2, short a3);
int sub_495150(int a1, short a2);
int sub_495180(int a1, uint16_t* a2, uint16_t* a3, uint8_t* a4);
char* sub_4951C0();
int nox_xxx_unitSpriteCheckAlly_4951F0(int a1);
int sub_495210(uint8_t* data);
int* sub_495500(int* a1);
int sub_4958F0();
int nox_xxx_allocClassListFriends_495980();
void sub_4959B0();
int sub_4959D0();
uint32_t* nox_xxx_cliAddObjFriend_4959F0(int a1);
void sub_495A20(int a1);
int sub_495A80(int a1);
void sub_495BB0(nox_drawable* dr, nox_draw_viewport_t* vp);
void sub_495B50(nox_drawable_fx* fx);
int sub_495BF0(nox_drawable* dr, nox_drawable_fx* fx, nox_draw_viewport_t* vp);
int sub_495D00(nox_drawable* dr, nox_drawable_fx* fx, nox_draw_viewport_t* vp);
void sub_495FC0(nox_drawable_fx* fx, nox_drawable* dr);
nox_point sub_499290(int a1);
int sub_4992B0(int a1, int a2);
int nox_xxx_loadReflSheild_499360();
int sub_499450();
int nox_xxx_drawShield_499810(nox_draw_viewport_t* vp, nox_drawable* dr);
uint32_t* nox_xxx_fxDrawTurnUndead_499880(short* a1);
void nox_xxx_drawPointMB_499B70(int xLeft, int yTop, int a3);
void nox_xxx_bookRewardCli_499CF0(int* a1, int a2, int a3);
void sub_499F60(int a1, int a2, int a3, short a4, char a5, char a6, char a7, char a8, char a9, int a10);
void* nox_npc_by_id(int id);
char* nox_xxx_clientEquip_49A3D0(uint8_t opcode, int npc_id, uint32_t item_type, const uint8_t modifiers[4]);

typedef struct nox_health_change {
	uint32_t drawable_id;
	int16_t delta;
	uint16_t reserved_6;
	uint32_t frame;
	struct nox_health_change* next;
	struct nox_health_change* prev;
} nox_health_change;
_Static_assert(offsetof(nox_health_change, drawable_id) == 0,
	"wrong offset of nox_health_change.drawable_id!");
_Static_assert(offsetof(nox_health_change, delta) == 4,
	"wrong offset of nox_health_change.delta!");
_Static_assert(offsetof(nox_health_change, frame) == 8,
	"wrong offset of nox_health_change.frame!");
_Static_assert(offsetof(nox_health_change, next) == (sizeof(void*) == 4 ? 12 : 16),
	"wrong native offset of nox_health_change.next!");
_Static_assert(offsetof(nox_health_change, prev) == (sizeof(void*) == 4 ? 16 : 24),
	"wrong native offset of nox_health_change.prev!");
_Static_assert(sizeof(nox_health_change) == (sizeof(void*) == 4 ? 20 : 32),
	"wrong native size of nox_health_change structure!");

int nox_xxx_allocArrayHealthChanges_49A5F0();
void sub_49A630();
nox_health_change* nox_xxx_cliAddHealthChange_49A650(int a1, short a2);
void sub_49A6A0(nox_draw_viewport_t* vp, nox_drawable* dr);
void sub_49A880(nox_health_change* change);
int sub_49A8C0();
void nox_xxx_sprite_49AA00_drawable(nox_drawable* dr);
void nox_xxx_updateSpritePosition_49AA90(nox_drawable* dr, int a2, int a3);
void nox_xxx_forEachSprite_49AB00(int4*, void*, void*);
nox_drawable* nox_drawable_find_49ABF0(nox_point* pt, int r);
int sub_49AEA0();
int sub_49B3E0();
int sub_49B420(int a1, int a2, int* a3, int a4);
int sub_49B490();
int sub_49B6B0();
void nox_xxx_consoleEsc_49B7A0();
void nox_xxx_spriteTransparentDecay_49B950(nox_drawable* a1, int a2);
void nox_xxx_spriteToSightDestroyList_49BAB0_drawable(nox_drawable* a1);
void sub_49BBB0();
void sub_49BBC0();
void nox_xxx_spriteToList_49BC80_drawable(nox_drawable* a1);
void sub_49BCD0(nox_drawable* dr);
uint32_t* nox_xxx_clientAddRayEffect_49C160(int a1);
void nox_xxx_clientRemoveRayEffect_49C450(int a1);
void nox_xxx_spriteDeleteSomeList_49C4B0();
void nox_xxx_sprite_49C4F0();
int nox_xxx_wnd_49C760(int a1, int a2, int* a3, int a4);
int sub_49C7A0();
int sub_49C810();
int sub_49CA60(int a1, int a2, int* a3, int a4);
int sub_49CB40();
void nox_client_drawBorderLines_49CC70(int xLeft, int yTop, int a3, int a4);
void sub_49CD30(int xLeft, int yTop, int a3, int a4, int a5, int a6);
void nox_client_drawRectFilledOpaque_49CE30(int xLeft, int yTop, int a3, int a4);
void nox_client_drawRectFilledAlpha_49CF10(int xLeft, int yTop, int a3, int a4);
int sub_49D1C0(void* a1, int a2, int a3);
void sub_49E3C0(uint32_t* a1, int a2, unsigned int a3);
int nox_client_drawLineFromPoints_49E4B0();
int sub_49E4F0(int a1);
void nox_client_drawPixel_49EFA0(int a1, int a2);
void nox_client_drawAddPoint_49F500(int a1, int a2);
void nox_xxx_rasterPointRel_49F570(int a1, int a2);
int sub_49F6D0(int a1);
int4* nox_client_copyRect_49F6F0(int xLeft, int yTop, int a3, int a4);
int4* sub_49F780(int xLeft, int a2);
void sub_49F7C0_def();
void nox_xxx_wndDraw_49F7F0();
int sub_49F860();
int4* nox_xxx_utilRect_49F930(int4* a1, int4* a2, int4* a3);
void sub_49FDB0(int a1);
uint32_t* sub_49FF20();
// The server payload is an exact 169-byte PE32/wire record. Keep native list
// links and the dynamically-created GUI row outside that payload so neither
// pointer is narrowed into its legacy 32-bit fields.
typedef struct nox_gui_server_node_t {
	nox_list_item_t list;
	nox_gui_server_ent_t server;
	nox_window* row;
} nox_gui_server_node_t;
_Static_assert(sizeof(nox_gui_server_ent_t) == 169,
	"WOL server payload must retain its packed wire size");
_Static_assert(offsetof(nox_gui_server_node_t, server) >= sizeof(nox_list_item_t),
	"native WOL list links overlap the wire payload");
_Static_assert(sizeof(((nox_gui_server_node_t*)0)->row) == sizeof(void*),
	"WOL row window must remain native-width");

static inline nox_gui_server_node_t* nox_wol_server_node_from_list(nox_list_item_t* item) {
	return item ? (nox_gui_server_node_t*)((uint8_t*)item - offsetof(nox_gui_server_node_t, list)) : NULL;
}

static inline const nox_gui_server_node_t* nox_wol_server_node_from_list_const(const nox_list_item_t* item) {
	return item ? (const nox_gui_server_node_t*)((const uint8_t*)item - offsetof(nox_gui_server_node_t, list)) : NULL;
}

static inline nox_gui_server_node_t* nox_wol_server_node_from_record(nox_gui_server_ent_t* server) {
	return server ? (nox_gui_server_node_t*)((uint8_t*)server - offsetof(nox_gui_server_node_t, server)) : NULL;
}

static inline const nox_gui_server_node_t* nox_wol_server_node_from_record_const(
	const nox_gui_server_ent_t* server) {
	return server ? (const nox_gui_server_node_t*)((const uint8_t*)server - offsetof(nox_gui_server_node_t, server)) : NULL;
}

nox_window* nox_wol_server_row_get(const nox_gui_server_ent_t* server);
void nox_wol_server_row_set(nox_gui_server_ent_t* server, nox_window* row);
nox_list_item_t* sub_49FFA0(int a1);
nox_list_item_t* sub_4A0020(void);
int nox_wol_servers_addResult_4A0030(nox_gui_server_ent_t* srv);
void nox_wol_servers_sortBtnHandler_4A0290(int id);
nox_list_item_t* sub_4A0360(void);
nox_list_item_t* sub_4A0390(void);
int sub_4A0410(const char* a1, short a2);
nox_gui_server_ent_t* sub_4A0490(int a1);
nox_gui_server_ent_t* sub_4A04C0(int a1);
nox_window* nox_new_window_from_file(char* cname, void* fnc);

unsigned sub_48C730(unsigned int a1);

#endif // NOX_PORT_GAME2_3
