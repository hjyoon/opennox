package legacy

/*
#include "defs.h"
#include "game4_mapgen_state.h"
#include "client__gui__guitrade.h"
#include "client__gui__guirank.h"
#include "client__gui__guisumn.h"
#include "server__script__internal.h"
#include "client_ui_words.h"
#include "vardefs.h"
extern void* nox_xxx_aClosewoodengat_587000_133480;
extern uintptr_t dword_5d4594_1062480;
extern nox_drawable* dword_5d4594_1063116;
extern uint32_t dword_5d4594_1046868;
extern uintptr_t dword_5d4594_529316;
extern uint32_t dword_5d4594_1599656;
extern void* dword_587000_155144;
extern uintptr_t dword_5d4594_1047516;
extern uint32_t dword_5d4594_3798804;
extern uint32_t dword_5d4594_832476;
extern uint32_t nox_xxx_polygonNextIdx_587000_60352;
extern uint32_t dword_5d4594_816348;
extern uint32_t dword_5d4594_1098628;
extern uint32_t dword_5d4594_1046872;
extern uint32_t dword_5d4594_251572;
extern uint32_t dword_5d4594_1047540;
extern uint32_t dword_5d4594_1046936;
extern uint32_t dword_5d4594_1107036;
extern void* dword_5d4594_1301848;
extern uint32_t nox_xxx_aNox_cfg_0_587000_132132;
extern uint32_t dword_5d4594_1047520;
extern uintptr_t dword_5d4594_534808;
extern char* dword_5d4594_1599576;
extern uint32_t dword_5d4594_1599596;
extern void* dword_5d4594_2516344;
extern void* nox_alloc_pixelSpan_1301844;
extern uint32_t dword_5d4594_3798808;
extern uint32_t dword_5d4594_3835396;
extern uint32_t dword_5d4594_3798816;
extern uint32_t dword_5d4594_2487524;
extern void* dword_587000_127004;
extern uint32_t dword_587000_93156;
extern uint32_t dword_587000_93160;
extern uintptr_t dword_5d4594_531648;
extern uint32_t dword_5d4594_815044;
extern void* nox_alloc_tradeItems_2386496;
extern uint32_t dword_5d4594_816340;
extern uint32_t dword_5d4594_1047932;
extern uint32_t dword_5d4594_816368;
extern uint32_t dword_5d4594_1049496;
extern void* dword_587000_93164;
extern uint32_t dword_5d4594_1098624;
extern nox_drawable* dword_5d4594_1320968;
extern uint64_t qword_581450_9552;
extern uint64_t qword_581450_9512;
extern uint32_t dword_5d4594_2386500;
extern uint32_t dword_5d4594_2649712;
extern uint32_t dword_587000_60044;
extern void* dword_5d4594_251560;
extern uint32_t dword_5d4594_3484;
extern void* nox_alloc_screenParticles_806044;
extern void* dword_587000_122852;
extern void* dword_5d4594_1301776;
extern uint32_t dword_5d4594_1550916;
extern void* dword_587000_81128;
extern uint32_t dword_5d4594_3798800;
extern uint32_t dword_5d4594_3798812;
extern uint32_t dword_5d4594_2487248;

extern uint32_t dword_5d4594_1049484;
extern uint32_t dword_5d4594_1046656;
extern uint32_t dword_5d4594_588120;
extern uint32_t dword_5d4594_1047512;
extern void* dword_5d4594_588084;
extern uint32_t dword_5d4594_816372;
extern void* dword_5d4594_1203864;
extern uint32_t dword_5d4594_3835356;
extern uint32_t nox_client_renderBubbles_80844;
extern void* nox_alloc_chat_1197364;
extern uint32_t nox_xxx_aNox_cfg_0_587000_132136;
extern uint32_t dword_5d4594_1046636;
extern nox_gui_trade_slot* dword_5d4594_1320932;
extern uint32_t dword_5d4594_1047528;
extern uint32_t dword_5d4594_608316;
extern uint32_t dword_5d4594_1046932;
extern void* dword_5d4594_588068;
extern uint32_t dword_587000_122848;
extern uint32_t dword_5d4594_831220;
extern uint32_t dword_5d4594_1203840;
extern uint32_t dword_5d4594_741332;
extern uint32_t dword_5d4594_816356;
extern uint32_t dword_5d4594_1203836;
extern void* dword_5d4594_1301780;
extern uint32_t dword_5d4594_1567928;
extern uint32_t dword_5d4594_1307720;
extern uint32_t dword_5d4594_1550912;
extern uint32_t dword_5d4594_1062488;
extern uint32_t dword_5d4594_1047936;
extern void* nox_alloc_tradeSession_2386492;
extern uint32_t dword_5d4594_3835352;
extern uint32_t dword_5d4594_825736;
extern uint32_t dword_5d4594_2487576;
extern void* nox_alloc_friendList_1203860;
extern uint32_t dword_5d4594_3835348;
extern uint32_t dword_5d4594_1046640;
extern nox_gui_trade_slot* dword_5d4594_1320972;
extern uint32_t dword_5d4594_1046648;
extern uint32_t dword_5d4594_10984;
extern nox_drawable* dword_5d4594_1063120;
extern void* nox_alloc_vote_1599652;
extern uint32_t dword_5d4594_2487676;
extern void* dword_5d4594_1049696;
extern uint32_t dword_5d4594_2487672;
extern void* nox_alloc_magicEnt_1569668;
extern uint32_t dword_587000_66116;
extern uint32_t dword_5d4594_2487564;
extern uint32_t nox_server_sanctuaryHelp_54276;
extern uint32_t dword_5d4594_1313740;
extern uintptr_t dword_5d4594_1062492;
extern wchar2_t* dword_5d4594_1098604;
extern uint32_t dword_5d4594_2487584;
extern void* nox_alloc_healthChange_1301772;
extern wchar2_t* dword_5d4594_1098600;
extern uint32_t dword_5d4594_1046864;
extern uint32_t dword_587000_183456;
extern wchar2_t* dword_5d4594_1098596;
extern uint32_t dword_5d4594_1047524;
extern nox_gui_trade_slot* dword_5d4594_1320936;
extern uint32_t dword_5d4594_831076;
extern uint32_t dword_5d4594_1203832;
extern uint32_t dword_5d4594_831244;
extern uint32_t dword_5d4594_1321520;
extern uint32_t dword_5d4594_526276;
extern uint32_t dword_5d4594_1062484;
extern uint32_t dword_5d4594_2487568;
extern uint32_t nox_server_sendMotd_108752;
extern uint32_t dword_5d4594_1123520;
extern uint32_t dword_5d4594_2488720;
extern uint32_t dword_5d4594_1599644;
extern uint32_t nox_server_resetQuestMinVotes_229988;
extern uint32_t dword_5d4594_831260;
extern uint32_t dword_5d4594_1045640;
extern uint32_t nox_server_kickQuestPlayerMinVotes_229992;
extern uint32_t dword_587000_183460;
extern uint32_t dword_5d4594_3835372;
extern uint32_t dword_5d4594_741652;
extern uint32_t dword_5d4594_600116;
extern uint32_t dword_5d4594_815052;
extern uint32_t dword_5d4594_741648;
extern nox_object_t* dword_5d4594_2491592;
extern uint32_t dword_5d4594_2488724;
extern void* dword_5d4594_1548532;
extern uint32_t dword_5d4594_805820;
extern uint32_t dword_5d4594_2489436;
extern char* dword_5d4594_831240;
extern uint32_t dword_5d4594_815748;
extern uint32_t dword_5d4594_1045420;
extern uint32_t dword_5d4594_739392;
extern uint32_t dword_5d4594_1313788;
extern uint32_t dword_5d4594_1047536;
extern void* dword_5d4594_2516352;
extern uint32_t dword_5d4594_1046652;
extern uint32_t dword_5d4594_3835360;
extern uint32_t dword_5d4594_1062552;
extern uint32_t dword_5d4594_1200832;
extern void* nox_alloc_hit_2491548;
extern uint32_t dword_5d4594_3835368;
extern uint32_t nox_xxx_useAudio_587000_80772;
extern void* dword_5d4594_1049692;
extern uint32_t dword_5d4594_2488620;
extern uint32_t dword_5d4594_1047532;
extern uint32_t dword_5d4594_2487656;
extern uint32_t dword_587000_122956;
extern uint32_t dword_5d4594_1046852;



extern void* dword_5d4594_814624;
extern uint32_t dword_5d4594_2649712;
extern void* nox_alloc_chat_1197364;
extern uint32_t dword_5d4594_1599644;
extern void* dword_5d4594_1548532;

extern uint32_t dword_5d4594_10984;
extern unsigned int dword_5d4594_527988;
extern uint32_t dword_5d4594_831220;
extern uint32_t dword_5d4594_831260;
extern nox_screenParticle* dword_5d4594_806052;
extern nox_screenParticle* nox_screenParticles_head;
extern void* nox_alloc_screenParticles_806044;
extern nox_window* dword_5d4594_1082856;
extern void* dword_5d4594_1082864;
extern void* dword_5d4594_1082868;




extern unsigned int dword_5d4594_527988;
extern uint32_t dword_5d4594_815044;

extern unsigned int nox_client_renderBubbles_80844;
extern uint32_t dword_5d4594_3798800;
extern uint32_t dword_5d4594_3798804;
extern uint32_t dword_5d4594_3798808;
extern uint32_t dword_5d4594_3798812;
extern uint32_t dword_5d4594_3798816;
extern uintptr_t dword_8531A0_2576;
extern uint32_t dword_8531A0_2572;
extern unsigned int dword_5d4594_1599628;

extern int nox_win_width;
extern int nox_win_height;

extern uint8_t** nox_pixbuffer_rows_3798784;

*/
import "C"
import (
	"sync/atomic"
	"unsafe"

	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/server"
)

func Sub_460D40() bool {
	return quickbarControls.left != nil
}

func Sub_4D7140(a1 uint32) {
	legacyWords.dword_5d4594_1556144 = C.uint32_t(a1)
}

func Sub_43AF30() uint32 {
	return uint32(C.dword_5d4594_815052)
}

func Get_nox_wol_server_result_cnt_815088() int {
	return int(legacyWords.nox_wol_server_result_cnt_815088)
}
func Get_dword_5d4594_815044() int {
	return int(C.dword_5d4594_815044)
}
func Set_dword_5d4594_815044(v int) {
	C.dword_5d4594_815044 = C.uint(v)
}
func Get_dword_587000_87412() int {
	return int(legacyWords.dword_587000_87412)
}
func Get_dword_5d4594_815000() *gui.Window {
	return asWindow(wolWindows.serverInfo)
}
func Get_dword_587000_87408() int {
	return int(legacyWords.dword_587000_87408)
}
func Get_nox_game_createOrJoin_815048() int {
	return int(C.nox_game_createOrJoin_815048)
}
func Get_dword_5d4594_814996() *gui.Window {
	return asWindow(wolWindows.heading)
}
func Get_dword_5d4594_814984() *gui.Window {
	return asWindow(wolWindows.mapPanel)
}
func Get_dword_5d4594_2488604() int {
	return int(legacyWords.dword_5d4594_2488604)
}
func Get_dword_5d4594_528252() int {
	return int(legacyWords.dword_5d4594_528252)
}
func Get_dword_5d4594_528260() int {
	return int(legacyWords.dword_5d4594_528260)
}
func Get_dword_587000_122852() unsafe.Pointer {
	return C.dword_587000_122852
}
func Get_dword_587000_93164() unsafe.Pointer {
	return C.dword_587000_93164
}
func Get_nox_client_translucentFrontWalls_805844() int {
	return int(legacyAddressableWords.nox_client_translucentFrontWalls_805844)
}
func Get_nox_client_highResFrontWalls_80820() int {
	return int(legacyAddressableWords.nox_client_highResFrontWalls_80820)
}
func Get_nox_client_highResFloors_154952() int {
	return int(legacyAddressableWords.nox_client_highResFloors_154952)
}
func Get_nox_client_fadeObjects_80836() int {
	return int(legacyAddressableWords.nox_client_fadeObjects_80836)
}
func Get_nox_client_renderBubbles_80844() int {
	return int(C.nox_client_renderBubbles_80844)
}
func Get_nox_server_sanctuaryHelp_54276() int {
	return int(C.nox_server_sanctuaryHelp_54276)
}
func Get_nox_server_sendMotd_108752() int {
	return int(C.nox_server_sendMotd_108752)
}
func Get_nox_server_connectionType_3596() int {
	return int(legacyWords.nox_server_connectionType_3596)
}
func Get_nox_server_resetQuestMinVotes_229988() int {
	return int(C.nox_server_resetQuestMinVotes_229988)
}
func Get_nox_server_kickQuestPlayerMinVotes_229992() int {
	return int(C.nox_server_kickQuestPlayerMinVotes_229992)
}
func Get_dword_5d4594_816340() int {
	return int(C.dword_5d4594_816340)
}
func Get_dword_587000_93156() int {
	return int(C.dword_587000_93156)
}
func Get_dword_587000_93160() int {
	return int(C.dword_587000_93160)
}
func Get_dword_5d4594_816348() int {
	return int(C.dword_5d4594_816348)
}

//export sub_435F60
func sub_435F60() C.int {
	value := 1 - legacyWords.dword_5d4594_811904
	legacyWords.dword_5d4594_811904 = value
	return C.int(value)
}

func Get_dword_5d4594_811904() int {
	return int(legacyWords.dword_5d4594_811904)
}
func Get_dword_5d4594_1193156() int {
	return int(legacyWords.dword_5d4594_1193156)
}
func Get_dword_5d4594_1193188() int {
	return int(legacyWords.dword_5d4594_1193188)
}
func Get_dword_5d4594_3798840() int {
	return int(int32(legacyWords.dword_5d4594_3798840))
}
func Get_dword_5d4594_3798824() int {
	return int(int32(legacyWords.dword_5d4594_3798824))
}
func Get_dword_5d4594_3798836() int {
	return int(int32(legacyWords.dword_5d4594_3798836))
}
func Get_dword_5d4594_3798820() int {
	return int(int32(legacyWords.dword_5d4594_3798820))
}
func Get_nox_alloc_chat_1197364() unsafe.Pointer {
	return C.nox_alloc_chat_1197364
}
func Get_dword_5d4594_2660032() int {
	return int(legacyWords.dword_5d4594_2660032)
}
func Get_dword_5d4594_1090048() *gui.Window {
	return asWindow(clientRankRoot)
}
func Get_nox_rank_window(offset int) *gui.Window {
	return asWindow(nox_rank_window_at(C.int(offset)))
}
func Get_dword_5d4594_1090120() int {
	return int(clientRankViewMode)
}
func Get_dword_5d4594_1090100() *gui.Window {
	return asWindow(clientRankWindow(clientRankPositionOffset))
}
func Get_dword_5d4594_814624() unsafe.Pointer {
	return C.dword_5d4594_814624
}
func Set_dword_5d4594_814624(v unsafe.Pointer) {
	C.dword_5d4594_814624 = v
}
func Get_dword_5d4594_251560() unsafe.Pointer {
	return C.dword_5d4594_251560
}
func Get_dword_5d4594_1064868() int {
	return int(*clientChatActive)
}
func Get_dword_5d4594_1316972() *gui.Window {
	return asWindow(serverPerformanceWindow)
}
func Get_dword_5d4594_1599588() unsafe.Pointer {
	return game4Mapgen.archive_path
}
func Get_dword_5d4594_1599592() unsafe.Pointer {
	return game4Mapgen.staging_path
}
func Get_dword_5d4594_1599540() unsafe.Pointer {
	return unsafe.Pointer(game4Mapgen.object_head)
}
func Get_dword_5d4594_1599476() int {
	return int(game4Mapgen.tiles_placed)
}
func Get_dword_5d4594_1599480() uint32 {
	return uint32(game4Mapgen.loaded_index)
}
func Get_dword_5d4594_1599548() unsafe.Pointer {
	return unsafe.Pointer(game4Mapgen.waypoint_head)
}
func Get_dword_5d4594_1599556() unsafe.Pointer {
	return unsafe.Pointer(game4Mapgen.tile_head)
}
func Get_dword_5d4594_1599532() unsafe.Pointer {
	return unsafe.Pointer(game4Mapgen.wall_head)
}
func Get_dword_5d4594_1200804() int {
	return int(legacyWords.dword_5d4594_1200804)
}
func Get_nox_gameDisableMapDraw_5d4594_2650672() int {
	return int(legacyWords.nox_gameDisableMapDraw_5d4594_2650672)
}
func Get_dword_5d4594_528256() int {
	return int(legacyWords.dword_5d4594_528256)
}
func Get_dword_5d4594_527988() int {
	return int(C.dword_5d4594_527988)
}
func Get_nox_alloc_screenParticles_806044() unsafe.Pointer {
	return C.nox_alloc_screenParticles_806044
}
func Get_dword_5d4594_251744() int {
	return int(legacyAddressableWords.dword_5d4594_251744)
}
func Get_dword_5d4594_831260() int {
	return int(C.dword_5d4594_831260)
}
func Get_dword_5d4594_831220() int {
	return int(C.dword_5d4594_831220)
}
func Get_nox_color_black_2650656() uint32 {
	return uint32(renderPalette.black)
}
func Get_nox_color_white_2523948() uint32 {
	return uint32(renderPalette.white)
}
func Get_nox_color_violet_2598268() uint32 {
	return uint32(renderPalette.violet)
}
func Get_nox_color_red_2589776() uint32 {
	return uint32(renderPalette.red)
}
func Get_nox_color_green_2614268() uint32 {
	return uint32(renderPalette.green)
}
func Get_nox_color_cyan_2649820() uint32 {
	return uint32(renderPalette.cyan)
}
func Get_nox_color_blue_2650684() uint32 {
	return uint32(renderPalette.blue)
}
func Get_nox_color_orange_2614256() uint32 {
	return uint32(renderPalette.orange)
}
func Get_nox_color_yellow_2589772() uint32 {
	return uint32(renderPalette.yellow)
}
func Get_nox_color_black_2650656_ptr() *uint32 {
	return (*uint32)(unsafe.Pointer(&renderPalette.black))
}
func Get_nox_color_white_2523948_ptr() *uint32 {
	return (*uint32)(unsafe.Pointer(&renderPalette.white))
}
func Get_nox_color_violet_2598268_ptr() *uint32 {
	return (*uint32)(unsafe.Pointer(&renderPalette.violet))
}
func Get_nox_color_red_2589776_ptr() *uint32 {
	return (*uint32)(unsafe.Pointer(&renderPalette.red))
}
func Get_nox_color_green_2614268_ptr() *uint32 {
	return (*uint32)(unsafe.Pointer(&renderPalette.green))
}
func Get_nox_color_cyan_2649820_ptr() *uint32 {
	return (*uint32)(unsafe.Pointer(&renderPalette.cyan))
}
func Get_nox_color_blue_2650684_ptr() *uint32 {
	return (*uint32)(unsafe.Pointer(&renderPalette.blue))
}
func Get_nox_color_orange_2614256_ptr() *uint32 {
	return (*uint32)(unsafe.Pointer(&renderPalette.orange))
}
func Get_nox_color_yellow_2589772_ptr() *uint32 {
	return (*uint32)(unsafe.Pointer(&renderPalette.yellow))
}
func Get_dword_5d4594_823776() int {
	return int(legacyWords.dword_5d4594_823776)
}
func Get_dword_5d4594_831236() *gui.Window {
	return asWindow(briefingState.outer)
}
func Get_dword_5d4594_1049844() int {
	return int(legacyWords.dword_5d4594_1049844)
}
func Get_dword_5d4594_3799452() int {
	return int(legacyWords.dword_5d4594_3799452)
}
func Get_dword_5d4594_1321236() *gui.Window {
	return asWindow(gameInputBindings.actions)
}
func Get_dword_5d4594_1321240() *gui.Window {
	return asWindow(gameInputBindings.labels)
}
func Get_dword_5d4594_1321244() *gui.Window {
	return asWindow(gameInputBindings.columns[0])
}
func Get_dword_5d4594_1321248() *gui.Window {
	return asWindow(gameInputBindings.columns[1])
}
func Get_dword_5d4594_1522616() *gui.Window {
	return asWindow(shellInputBindings.actions)
}
func Get_dword_5d4594_1522620() *gui.Window {
	return asWindow(shellInputBindings.labels)
}
func Get_dword_5d4594_1522624() *gui.Window {
	return asWindow(shellInputBindings.columns[0])
}
func Get_dword_5d4594_1522628() *gui.Window {
	return asWindow(shellInputBindings.columns[1])
}
func Get_dword_5d4594_1062452() *gui.Window {
	return asWindow(inventoryState.root)
}
func Get_dword_5d4594_2516356() int {
	return int(legacyWords.dword_5d4594_2516356)
}
func Get_dword_5d4594_1548524() int {
	return int(legacyWords.dword_5d4594_1548524)
}
func Get_dword_587000_155144() unsafe.Pointer {
	return C.dword_587000_155144
}
func Get_nox_alloc_magicEnt_1569668() unsafe.Pointer {
	return C.nox_alloc_magicEnt_1569668
}
func Get_nox_client_gui_flag_1556112() int {
	return int(legacyWords.nox_client_gui_flag_1556112)
}

func Set_dword_5d4594_1599628(v int) {
	C.dword_5d4594_1599628 = C.uint(v)
}

func Sub_41E2F0() int {
	return int(C.dword_5d4594_527988)
}

func Get_nox_wol_wnd_world_814980() *gui.Window {
	return asWindow(wolWindows.root)
}

func Get_dword_5d4594_815004() *gui.Window {
	return asWindow(wolWindows.serverDetails)
}

func Nox_xxx_get_57AF20() int {
	if pauseFXStateNative57AF30.active {
		return 1
	}
	return 0
}

func Set_dword_5d4594_1599644(v int) {
	C.dword_5d4594_1599644 = C.uint32_t(v)
}
func Set_dword_5d4594_1599480(v uint32) {
	game4Mapgen.loaded_index = C.uint(v)
}
func Set_dword_5d4594_1599476(v int) {
	game4Mapgen.tiles_placed = C.uint(v)
}
func Set_dword_5d4594_1599540(v unsafe.Pointer) {
	game4Mapgen.object_head = (*C.nox_map_object_list_node_5048A0)(v)
}
func Set_dword_5d4594_1599532(v unsafe.Pointer) {
	game4Mapgen.wall_head = (*C.nox_map_wall_list_node_504290)(v)
}
func Set_dword_5d4594_1599556(v unsafe.Pointer) {
	game4Mapgen.tile_head = (*C.nox_tile_coord_entry_t)(v)
}
func Set_dword_5d4594_1599548(v unsafe.Pointer) {
	game4Mapgen.waypoint_head = (*C.nox_map_waypoint_list_node_5044B0)(v)
}
func Get_dword_5d4594_2386836() int {
	return int(legacyWords.dword_5d4594_2386836)
}
func Set_dword_5d4594_2386836(v int) {
	legacyWords.dword_5d4594_2386836 = C.uint(v)
}
func Set_dword_5d4594_2386848(v int) {
	legacyWords.dword_5d4594_2386848 = C.int(v)
}
func Get_dword_5d4594_2386848() int {
	return int(legacyWords.dword_5d4594_2386848)
}
func Set_dword_5d4594_2386852(v int) {
	legacyWords.dword_5d4594_2386852 = C.uint(v)
}
func Get_dword_5d4594_2386852() uint32 {
	return uint32(legacyWords.dword_5d4594_2386852)
}
func Set_dword_5d4594_815104(v int) {
	legacyWords.dword_5d4594_815104 = C.int(v)
}
func Set_nox_wol_server_result_cnt_815088(v int) {
	legacyWords.nox_wol_server_result_cnt_815088 = C.uint(v)
}
func Set_qword_5d4594_815068(v uint64) {
	wolRefreshDeadline = v
}
func Set_nox_client_connError_814552(v int) {
	wolConnectionError = uint32(v)
}
func Set_dword_5d4594_815056(v int) {
	legacyWords.dword_5d4594_815056 = C.uint(v)
}
func Set_nox_game_createOrJoin_815048(v int) {
	C.nox_game_createOrJoin_815048 = C.uint32_t(v)
}
func Set_dword_587000_87412(v int) {
	legacyWords.dword_587000_87412 = C.int(v)
}
func Set_dword_5d4594_527988(v int) {
	C.dword_5d4594_527988 = C.uint(v)
}
func Set_dword_5d4594_10984(v int) {
	C.dword_5d4594_10984 = C.uint(v)
}
func Set_nox_alloc_screenParticles_806044(v unsafe.Pointer) {
	C.nox_alloc_screenParticles_806044 = v
}
func Set_nox_screenParticles_head(v unsafe.Pointer) {
	C.nox_screenParticles_head = (*C.nox_screenParticle)(v)
}
func Set_dword_5d4594_806052(v unsafe.Pointer) {
	C.dword_5d4594_806052 = (*C.nox_screenParticle)(v)
}
func Set_nox_gameDisableMapDraw_5d4594_2650672(v int) {
	legacyWords.nox_gameDisableMapDraw_5d4594_2650672 = C.uint(v)
}
func Set_dword_5d4594_831224(v int) {
	legacyWords.dword_5d4594_831224 = C.uint(v)
}
func Set_dword_5d4594_1563096(v int) {
	legacyWords.dword_5d4594_1563096 = C.uint(v)
}
func Get_dword_5d4594_1563096() int {
	return int(legacyWords.dword_5d4594_1563096)
}
func Set_nox_win_width(v int) {
	C.nox_win_width = C.int(v)
}
func Set_nox_win_height(v int) {
	C.nox_win_height = C.int(v)
}
func Set_dword_5d4594_823776(v int) {
	legacyWords.dword_5d4594_823776 = C.uint(v)
}
func Set_nox_draw_curDrawData_3799572(v *noxrender.RenderData) {
	currentDrawData = (*nox_render_data_t)(v.C())
}
func Set_nox_color_black_2650656(v uint32) {
	renderPalette.black = C.uint32_t(v)
}
func Set_nox_color_white_2523948(v uint32) {
	renderPalette.white = C.uint32_t(v)
}
func Set_nox_color_violet_2598268(v uint32) {
	renderPalette.violet = C.uint32_t(v)
}
func Set_nox_color_red_2589776(v uint32) {
	renderPalette.red = C.uint32_t(v)
}
func Set_nox_color_green_2614268(v uint32) {
	renderPalette.green = C.uint32_t(v)
}
func Set_nox_color_cyan_2649820(v uint32) {
	renderPalette.cyan = C.uint32_t(v)
}
func Set_nox_color_blue_2650684(v uint32) {
	renderPalette.blue = C.uint32_t(v)
}
func Set_nox_color_orange_2614256(v uint32) {
	renderPalette.orange = C.uint32_t(v)
}
func Set_nox_color_yellow_2589772(v uint32) {
	renderPalette.yellow = C.uint32_t(v)
}
func Set_dword_5d4594_1305748(v int) {
	drawClipSaved = C.uint32_t(v)
}
func Set_nox_pixbuffer_rows_3798784(v unsafe.Pointer) {
	C.nox_pixbuffer_rows_3798784 = (**C.uchar)(v)
}
func Set_dword_5d4594_1193188(v int) {
	legacyWords.dword_5d4594_1193188 = C.uint(v)
}
func Set_dword_5d4594_2649712(v uint32) {
	C.dword_5d4594_2649712 = C.uint(v)
}
func Set_dword_5d4594_1548524(v int) {
	legacyWords.dword_5d4594_1548524 = C.uint(v)
}
func Set_nox_client_gui_flag_1556112(v int) {
	legacyWords.nox_client_gui_flag_1556112 = C.uint(v)
}
func Set_dword_5d4594_3798812(v int) {
	C.dword_5d4594_3798812 = C.uint(v)
}
func Set_dword_5d4594_3798800(v int) {
	C.dword_5d4594_3798800 = C.uint(v)
}
func Set_dword_5d4594_3798816(v int) {
	C.dword_5d4594_3798816 = C.uint(v)
}
func Set_dword_5d4594_3798808(v int) {
	C.dword_5d4594_3798808 = C.uint(v)
}
func Set_dword_5d4594_3798804(v int) {
	C.dword_5d4594_3798804 = C.uint(v)
}
func Set_dword_5d4594_3798820(v int) {
	legacyWords.dword_5d4594_3798820 = C.uint(v)
}
func Set_dword_5d4594_3798824(v int) {
	legacyWords.dword_5d4594_3798824 = C.uint(v)
}
func Set_dword_5d4594_3798828(v int) {
	legacyWords.dword_5d4594_3798828 = C.uint(v)
}
func Set_dword_5d4594_3798832(v int) {
	legacyWords.dword_5d4594_3798832 = C.uint(v)
}
func Set_dword_5d4594_3798836(v int) {
	legacyWords.dword_5d4594_3798836 = C.uint(v)
}
func Set_dword_5d4594_3798840(v int) {
	legacyWords.dword_5d4594_3798840 = C.uint(v)
}
func Set_dword_5d4594_3799452(v int) {
	legacyWords.dword_5d4594_3799452 = C.uint(v)
}
func Set_nox_alloc_chat_1197364(v unsafe.Pointer) {
	C.nox_alloc_chat_1197364 = v
}
func Set_dword_587000_145664(v int) {
	clientRankDirty = uint32(v)
}
func Set_dword_587000_145668(v int) {
	clientRankPendingMode = uint32(v)
}
func Set_dword_5d4594_1090120(v int) {
	clientRankViewMode = uint32(v)
}
func Set_dword_5d4594_251744(v int) {
	legacyAddressableWords.dword_5d4594_251744 = C.uint32_t(v)
}
func Set_dword_5d4594_816340(v int) {
	C.dword_5d4594_816340 = C.uint(v)
}
func Set_dword_5d4594_816348(v int) {
	C.dword_5d4594_816348 = C.uint(v)
}
func Set_dword_5d4594_1548532(v unsafe.Pointer) {
	C.dword_5d4594_1548532 = v
}
func Set_dword_5d4594_831260(v int) {
	C.dword_5d4594_831260 = C.uint(v)
}
func Set_nox_client_translucentFrontWalls_805844(v int) {
	legacyAddressableWords.nox_client_translucentFrontWalls_805844 = C.uint(v)
}
func Set_nox_client_highResFrontWalls_80820(v int) {
	legacyAddressableWords.nox_client_highResFrontWalls_80820 = C.uint(v)
}
func Set_nox_client_highResFloors_154952(v int) {
	legacyAddressableWords.nox_client_highResFloors_154952 = C.uint(v)
}
func Set_nox_client_fadeObjects_80836(v int) {
	legacyAddressableWords.nox_client_fadeObjects_80836 = C.uint(v)
}
func Set_nox_client_renderBubbles_80844(v int) {
	C.nox_client_renderBubbles_80844 = C.uint(v)
}
func Set_nox_server_sanctuaryHelp_54276(v int) {
	C.nox_server_sanctuaryHelp_54276 = C.uint(v)
}
func Set_nox_server_sendMotd_108752(v int) {
	C.nox_server_sendMotd_108752 = C.uint(v)
}
func Set_nox_server_connectionType_3596(v int) {
	legacyWords.nox_server_connectionType_3596 = C.uint(v)
}
func Set_nox_server_resetQuestMinVotes_229988(v int) {
	C.nox_server_resetQuestMinVotes_229988 = C.uint(v)
}
func Set_nox_server_kickQuestPlayerMinVotes_229992(v int) {
	C.nox_server_kickQuestPlayerMinVotes_229992 = C.uint(v)
}
func Set_dword_5d4594_1308156(v unsafe.Pointer) {
	streetPantsModifier = (*server.Modifier)(v)
}
func Set_dword_5d4594_1308160(v unsafe.Pointer) {
	streetShirtModifier = (*server.Modifier)(v)
}
func Set_dword_5d4594_1308164(v unsafe.Pointer) {
	streetSneakersModifier = (*server.Modifier)(v)
}
func Set_dword_5d4594_2650676(v int) {
	legacyWords.dword_5d4594_2650676 = C.uint(v)
}
func Set_dword_5d4594_2650680(v int) {
	legacyWords.dword_5d4594_2650680 = C.uint(v)
}
func Set_qword_5d4594_814956(v uint64) {
	wolConnectDeadline = v
}
func Set_dword_5d4594_1200804(v uint32) {
	legacyWords.dword_5d4594_1200804 = C.uint(v)
}
func Set_dword_5d4594_1200832(v int) {
	C.dword_5d4594_1200832 = C.uint(v)
}
func Set_nox_alloc_magicEnt_1569668(v unsafe.Pointer) {
	C.nox_alloc_magicEnt_1569668 = v
}
func Set_dword_5d4594_1569672(v int) {
	legacyWords.dword_5d4594_1569672 = C.uint(v)
}
func Get_nox_xxx_waypointCounterMB_587000_154948() uint32 {
	return uint32(legacyAddressableWords.nox_xxx_waypointCounterMB_587000_154948)
}
func Inc_nox_xxx_waypointCounterMB_587000_154948() {
	legacyAddressableWords.nox_xxx_waypointCounterMB_587000_154948++
}
func Inc_dword_5d4594_2516356() {
	legacyWords.dword_5d4594_2516356++
}
func Inc_dword_5d4594_2386848() {
	legacyWords.dword_5d4594_2386848++
}
func Inc_nox_wol_server_result_cnt_815088() {
	legacyWords.nox_wol_server_result_cnt_815088++
}
func Set_dword_5d4594_1599588(v unsafe.Pointer) {
	game4Mapgen.archive_path = v
}
func Set_dword_5d4594_1599592(v unsafe.Pointer) {
	game4Mapgen.staging_path = v
}
func Get_dword_5d4594_1599592_ptr() *uint32 {
	return (*uint32)(unsafe.Pointer(&C.dword_5d4594_2649712))
}
func Get_dword_5d4594_1123524() *gui.Window {
	return AsWindowP(unsafe.Pointer(npcDialogWindow))
}
func Set_dword_5d4594_1123520(v int) {
	C.dword_5d4594_1123520 = C.uint(v)
}
func Get_dword_5d4594_826036() *byte {
	return (*byte)(motdBuffer)
}

// Keep a typed reference alongside the uintptr mirror consumed by legacy C.
var currentPlayer8531A02576 atomic.Pointer[server.Player]

func Get_dword_8531A0_2576() *server.Player {
	return currentPlayer8531A02576.Load()
}
func Set_dword_8531A0_2576(p *server.Player) {
	currentPlayer8531A02576.Store(p)
	C.dword_8531A0_2576 = C.uintptr_t(uintptr(p.C()))
}
func Get_dword_8531A0_2572() uint32 {
	return uint32(C.dword_8531A0_2572)
}
func Get_dword_8531A0_2572_ptr() *uint32 {
	return (*uint32)(unsafe.Pointer(&C.dword_8531A0_2572))
}
func Set_dword_8531A0_2572(v uint32) {
	C.dword_8531A0_2572 = C.uint(v)
}
func Get_dword_5d4594_2488620() uint32 {
	return uint32(C.dword_5d4594_2488620)
}

// Glyph casts share this original C-owned DWORD, not its retired blob slot.
func Get_dword_5d4594_2487712_ptr() *uint32 {
	return (*uint32)(unsafe.Pointer(&legacyAddressableWords.dword_5d4594_2487712))
}

// Meteor casts share this original C-owned DWORD, not its retired blob slot.
func Get_dword_5d4594_2487804_ptr() *uint32 {
	return (*uint32)(unsafe.Pointer(&legacyAddressableWords.dword_5d4594_2487804))
}

func Set_dword_5d4594_2488620(v uint32) {
	C.dword_5d4594_2488620 = C.uint32_t(v)
}
func Get_dword_5d4594_608316() uint32 {
	return uint32(C.dword_5d4594_608316)
}
func Set_dword_5d4594_608316(v uint32) {
	C.dword_5d4594_608316 = C.uint32_t(v)
}
func Nox_xxx_playerGet_470A90() int {
	return int(meterState.barsVisible)
}
