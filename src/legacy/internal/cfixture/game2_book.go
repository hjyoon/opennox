package cfixture

/*
#include <stddef.h>
#include <string.h>
#include "game2_book_native.h"
static int book_ref_cursor, book_ref_result, book_ref_len, book_ref_calls;
static unsigned char book_ref_packet[22];
static int game2_book_old_cursor(void) { return book_ref_cursor; }
static int game2_book_old_send(int a,int b,unsigned char* p,int n) {
 if(a!=31||b!=0||n<0||n>22)abort();
 book_ref_len=n;book_ref_calls++;memcpy(book_ref_packet,p,n);return book_ref_result;
}
#define nox_win_unk1 (nox_book_ui_storage()->list_window)
#define nox_book_arrow_images (nox_book_ui_storage()->arrows)
#define nox_book_background_image (nox_book_ui_storage()->background)
#define nox_book_guide_tab_lit_image (nox_book_ui_storage()->guide_tab_lit)
#define nox_book_spell_tab_lit_image (nox_book_ui_storage()->spell_tab_lit)
#define nox_player_netCode_85319C (nox_client_player_net_code())
#define dword_5d4594_1046944 (nox_book_details_storage()->left_arrow)
#define dword_5d4594_1046948 (nox_book_details_storage()->right_arrow)
#define dword_5d4594_1046952 (nox_book_details_storage()->page_icon)
#define dword_5d4594_1046956 (nox_book_details_storage()->reward)
#define dword_5d4594_1046924 (nox_book_details_storage()->forward)
#define dword_5d4594_1046928 (nox_book_details_storage()->backward)
#define dword_5d4594_1046636 (nox_book_state_storage()->reward_x_bits)
#define dword_5d4594_1046640 (nox_book_state_storage()->reward_y_bits)
#define dword_5d4594_1046648 (nox_book_state_storage()->reward_start_frame)
#define dword_5d4594_1046652 (nox_book_state_storage()->reward_ability)
#define dword_5d4594_1046656 (nox_book_state_storage()->line_height)
#define dword_5d4594_1046852 (nox_book_state_storage()->reward_slot)
#define dword_5d4594_1046864 (nox_book_state_storage()->open_requested)
#define dword_5d4594_1046868 (nox_book_state_storage()->page_mode)
#define dword_5d4594_1046872 (nox_book_state_storage()->guide_mode)
#define dword_5d4594_1046932 (nox_book_state_storage()->entry_index)
#define dword_5d4594_1046936 (nox_book_state_storage()->list_page)
#define dword_5d4594_1047512 (nox_book_state_storage()->hidden_entry_count)
#define dword_5d4594_1047516 (nox_book_state_storage()->player_word)
#define dword_5d4594_1047520 (nox_book_state_storage()->reward_active)
#define dword_5d4594_1047524 (nox_book_state_storage()->reward_id)
#define dword_5d4594_1047528 (nox_book_state_storage()->list_drag_id)
#define dword_5d4594_1047532 (nox_book_state_storage()->list_drag_x)
#define dword_5d4594_1047536 (nox_book_state_storage()->list_drag_y)
#define dword_5d4594_1047540 (nox_book_state_storage()->page_drag_id)
#define nox_xxx_aNox_cfg_0_587000_132132 (nox_book_state_storage()->show_detail_page)
#define nox_xxx_aNox_cfg_0_587000_132136 (nox_book_state_storage()->list_drag_index)
#define obj_5d4594_1046620 (nox_book_state_storage()->reward_velocity)
#include "client_native_storage.h"
#define nox_win_width (nox_client_display_storage()->width)
#define nox_win_height (nox_client_display_storage()->height)
#include "game2_book_reference.h"
static void game2_book_test_config(int cursor,int result) {
 book_ref_cursor=cursor;book_ref_result=result;book_ref_len=book_ref_calls=0;memset(book_ref_packet,0,22);
}
static int game2_book_test_send(int old,int ability,void* p,int count,char flag) {
 if(ability)return old?game2_book_old_nox_xxx_clientSendAbil_45DAF0(count):nox_xxx_clientSendAbil_45DAF0(count);
 return old?game2_book_old_nox_xxx_clientSendSpell_45DB20(p,count,flag):nox_xxx_clientSendSpell_45DB20(p,count,flag);
}
static int game2_book_test_capture(void* dst) { memcpy(dst,book_ref_packet,22);return book_ref_len; }
static int game2_book_test_calls(void) { return book_ref_calls; }
static void* game2_book_test_row(int old,void* p,int slot,int kind) {
 if(kind==0)return old?game2_book_old_nox_quickbar_selected_row(p):nox_quickbar_selected_row(p);
 if(kind==1)return old?game2_book_old_nox_quickbar_selected_slot(p,slot):nox_quickbar_selected_slot(p,slot);
 return old?game2_book_old_nox_quickbar_trap_selected_row():nox_quickbar_trap_selected_row();
}
static void* game2_book_test_anim(int old,void* p,void* fn,int op) {
 if(op==0)return old?game2_book_old_nox_book_anim_data(p):nox_book_anim_data(p);
 if(op==1){if(old)game2_book_old_nox_book_anim_set_end(p,fn);else nox_book_anim_set_end(p,fn);}
 else {if(old)game2_book_old_nox_book_anim_start(p);else nox_book_anim_start(p);}
 return NULL;
}
static size_t game2_book_test_layout(int f) {
 switch(f) {
 case 0:return sizeof(nox_things_imageRef_t);case 1:return offsetof(nox_things_imageRef_t,field_24);
 case 2:return sizeof(nox_image_ref_anim_t);case 3:return offsetof(nox_image_ref_anim_t,on_end);
 case 4:return offsetof(nox_image_ref_anim_t,images);case 5:return offsetof(nox_image_ref_anim_t,field_3);
 case 6:return sizeof(nox_drawable);case 7:return offsetof(nox_drawable,flags30);
 default:return (size_t)-1;
 }
}
static void* game2_book_test_callback(int creature) { return creature?(void*)nox_xxx_bookClickCreature_45B200:(void*)nox_xxx_bookClickSpell_45B1F0; }
static int game2_book_test_localflags(int old,uint32_t flags) { return old?game2_book_old_nox_client_localPlayerHasFlags30(flags):nox_client_localPlayerHasFlags30(flags); }
static int game2_book_test_move(int old,void* a,void* b) { return old?game2_book_old_sub_45D7D0(a,b):sub_45D7D0(a,b); }
static int game2_book_test_click(int old,int creature) {
 if(creature)return old?game2_book_old_nox_xxx_bookClickCreature_45B200():nox_xxx_bookClickCreature_45B200();
 return old?game2_book_old_nox_xxx_bookClickSpell_45B1F0():nox_xxx_bookClickSpell_45B1F0();
}
static uint32_t game2_book_test_mode(int field,uint32_t value,int set) {
 uint32_t* p=field==0?&dword_5d4594_1046868:&dword_5d4594_1046872;
 if(set)*p=value;return *p;
}
static int game2_book_test_reset(int old) { return old?game2_book_old_sub_45DB90():sub_45DB90(); }
static int game2_book_test_target(int old,int direct,int id) {
 if(direct)return old?game2_book_old_nox_xxx_guiSpell_45DA10(id):nox_xxx_guiSpell_45DA10(id);
 return old?game2_book_old_nox_xxx_guiSpellTargetClickSet_45D9D0(id):nox_xxx_guiSpellTargetClickSet_45D9D0(id);
}
static uint32_t game2_book_test_busy(uint32_t value,int set) {
 if(set)dword_5d4594_1047520=value;return dword_5d4594_1047520;
}
static int game2_book_test_blocked_callback(int old,int which,void* win,unsigned event,unsigned pos) {
 switch(which) {
 case 0:return old?game2_book_old_nox_xxx_bookWndProc_45B070(0,event):nox_xxx_bookWndProc_45B070(win,event);
 case 1:return old?game2_book_old_nox_xxx_book_45B210(0,event):nox_xxx_book_45B210(win,event);
 case 2:return old?game2_book_old_nox_xxx_bookChildWndProcMB_45B360(win,event):nox_xxx_bookChildWndProcMB_45B360(win,event);
 default:return old?game2_book_old_nox_xxx_bookListWndProc_45B5F0(win,event,pos):nox_xxx_bookListWndProc_45B5F0(win,event,pos);
 }
}
*/
import "C"

import "unsafe"

func Game2BookConfig(cursor, result int32) {
	C.game2_book_test_config(C.int(cursor), C.int(result))
}
func Game2BookSend(old, ability bool, p unsafe.Pointer, count int32, flag byte) int32 {
	return int32(C.game2_book_test_send(C.int(boolInt(old)), C.int(boolInt(ability)), p, C.int(count), C.char(flag)))
}
func boolInt(v bool) int32 {
	if v {
		return 1
	}
	return 0
}
func Game2BookCapture() ([]byte, int) {
	p := make([]byte, 22)
	n := int(C.game2_book_test_capture(unsafe.Pointer(&p[0])))
	return p[:n], int(C.game2_book_test_calls())
}
func Game2BookRow(old bool, p unsafe.Pointer, slot, kind int) unsafe.Pointer {
	return C.game2_book_test_row(C.int(boolInt(old)), p, C.int(slot), C.int(kind))
}
func Game2BookAnim(old bool, p, fn unsafe.Pointer, op int) unsafe.Pointer {
	return C.game2_book_test_anim(C.int(boolInt(old)), p, fn, C.int(op))
}
func Game2BookLayout(f int) uintptr { return uintptr(C.game2_book_test_layout(C.int(f))) }
func Game2BookCallback(creature bool) unsafe.Pointer {
	return C.game2_book_test_callback(C.int(boolInt(creature)))
}
func Game2BookLocalFlags(old bool, flags uint32) int32 {
	return int32(C.game2_book_test_localflags(C.int(boolInt(old)), C.uint32_t(flags)))
}
func Game2BookMove(old bool, a, b unsafe.Pointer) int32 {
	return int32(C.game2_book_test_move(C.int(boolInt(old)), a, b))
}
func Game2BookClick(old, creature bool) int32 {
	return int32(C.game2_book_test_click(C.int(boolInt(old)), C.int(boolInt(creature))))
}
func Game2BookMode(field int, value uint32, set bool) uint32 {
	return uint32(C.game2_book_test_mode(C.int(field), C.uint32_t(value), C.int(boolInt(set))))
}
func Game2BookReset(old bool) int32 {
	return int32(C.game2_book_test_reset(C.int(boolInt(old))))
}
func Game2BookTarget(old, direct bool, id int32) int32 {
	return int32(C.game2_book_test_target(C.int(boolInt(old)), C.int(boolInt(direct)), C.int(id)))
}
func Game2BookBusy(v uint32, set bool) uint32 {
	return uint32(C.game2_book_test_busy(C.uint32_t(v), C.int(boolInt(set))))
}
func Game2BookBlockedCallback(old bool, which int, win unsafe.Pointer, event, pos uint32) int32 {
	return int32(C.game2_book_test_blocked_callback(C.int(boolInt(old)), C.int(which), win, C.uint(event), C.uint(pos)))
}
