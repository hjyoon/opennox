package legacy

/*
#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#include "defs.h"
#include "GAME1.h"
#include "GAME1_1.h"
#include "GAME1_2.h"
#include "GAME2.h"
#include "GAME3.h"
#include "common__system__team.h"
#include "operators.h"
extern uint32_t dword_5d4594_1313740,dword_5d4594_1313788;
static unsigned int game3_test_owner(int n){return n?dword_5d4594_1313788:dword_5d4594_1313740;}
static void game3_test_set_owner(int n,unsigned int v){if(n)dword_5d4594_1313788=v;else dword_5d4594_1313740=v;}

static void* game3_test_alloc(size_t n) {
 unsigned char* base=calloc(1,n+64); if(!base)return NULL;
 memset(base,0xA7,32); memset(base+32+n,0xD3,32); return base+32;
}
static int game3_test_guards(void* value,size_t n) {
 unsigned char* p=value; for(int i=0;i<32;i++)if(p[i-32]!=0xA7||p[n+i]!=0xD3)return 0;return 1;
}
static void game3_test_free(void* p){free((unsigned char*)p-32);}
static size_t game3_test_draw_size(void){return sizeof(nox_drawable);}
static size_t game3_test_equipment_size(void){return sizeof(nox_player_equipment_data_t);}
static size_t game3_test_modifier_offset(void){return offsetof(nox_player_equipment_data_t,modifiers);}
static size_t game3_test_draw_union_offset(void){return offsetof(nox_drawable,field_108_1);}
static void* game3_test_item_modifier(nox_drawable* d,int i){return d->item_modifiers[i];}
static void* game3_test_summon_child(nox_drawable* d){void* child;memcpy(&child,&d->field_108_1,sizeof child);return child;}
static uint32_t game3_test_union_word(nox_drawable* d,int i){return d->union_u32[i];}
static uintptr_t game3_test_summon(short id,unsigned short* pos,unsigned short thing,unsigned char direction,short life){return(uintptr_t)nox_xxx_netHandleSummonPacket_4B7C40(id,pos,thing,direction,life);}
static uintptr_t game3_test_shield(unsigned int code,int direction){return(uintptr_t)nox_xxx_fxShield_4B8090(code,direction);}
static uintptr_t game3_test_material(void* entries,uintptr_t bit){return(uintptr_t)sub_4B8E10(entries,(char*)bit);}

static unsigned char game3_ref_bytes[4096];
static size_t game3_ref_len,game3_ref_pos;
static nox_drawable* game3_ref_drawable;
static int game3_ref_nil,game3_ref_flags,game3_ref_teams;
static unsigned int game3_ref_clock;
static uint32_t game3_ref_reads[128][7];
static unsigned int game3_ref_read_count;
static void game3_ref_begin(const void* raw,size_t n,nox_drawable* d,int fail,int flags,unsigned int frame){
 if(n>sizeof game3_ref_bytes)abort();memcpy(game3_ref_bytes,raw,n);game3_ref_len=n;game3_ref_pos=0;
 game3_ref_drawable=d;game3_ref_nil=fail;game3_ref_flags=flags;game3_ref_clock=frame;game3_ref_teams=0;game3_ref_read_count=0;
}
static int game3_ref_rw(void* out,size_t n){
 if(game3_ref_read_count>=128||game3_ref_pos+n>game3_ref_len)abort();
 uint32_t* r=game3_ref_reads[game3_ref_read_count++];r[0]=game3_ref_pos;r[1]=n;
 r[2]=game3_ref_drawable->field_72;r[3]=game3_ref_drawable->flags30;
 memcpy(&r[4],&game3_ref_drawable->shape.box_w,4);r[5]=game3_ref_drawable->field_108_1;r[6]=game3_ref_drawable->field_108_2;
 memcpy(out,game3_ref_bytes+game3_ref_pos,n);game3_ref_pos+=n;return n!=0;
}
static int game3_ref_seek(int n){if(n<0||game3_ref_pos+(size_t)n>game3_ref_len)abort();game3_ref_pos+=n;return 0;}
static nox_drawable* game3_ref_load(int id,int x,int y){if(game3_ref_nil)return NULL;game3_ref_drawable->field_27=id;game3_ref_drawable->pos.x=x;game3_ref_drawable->pos.y=y;return game3_ref_drawable;}
static int game3_ref_gameflags(unsigned int mask){return(game3_ref_flags&mask)!=0;}
static int game3_ref_type(char* name){(void)name;return 999;}
static void game3_ref_team(unsigned char team,nox_object_team_t* data,int active,int code,int flags){(void)team;(void)data;(void)active;(void)code;(void)flags;game3_ref_teams++;}
static void* game3_ref_attach(nox_drawable* d,int x,int y){(void)d;(void)x;(void)y;return NULL;}
static void game3_ref_update(nox_drawable* d,int x,int y){d->pos.x=x;d->pos.y=y;}
static int game3_ref_spark(uint32_t* vp,nox_drawable* d,int a,int b){(void)vp;(void)d;(void)a;(void)b;return 0;}
static void game3_ref_load_error(void){}

#undef nox_xxx_fileReadWrite_426AC0_file3_fread
#define nox_xxx_fileReadWrite_426AC0_file3_fread game3_ref_rw
#define nox_xxx_cryptSeekCur_40E0A0 game3_ref_seek
#define nox_xxx_spriteLoadAdd_45A360_drawable game3_ref_load
#define nox_common_gameFlags_check_40A5C0 game3_ref_gameflags
#define nox_xxx_getTTByNameSpriteMB_44CFC0 game3_ref_type
#define nox_xxx_createAtImpl_4191D0 game3_ref_team
#define sub_410390 game3_ref_attach
#define nox_xxx_updateSpritePosition_49AA90 game3_ref_update
#define sub_4B6880 game3_ref_spark
#define nox_xxx_spriteLoadError_4356E0 game3_ref_load_error
#define gameFrame() game3_ref_clock
#define sub_4ABDA0 game3_ref_sub_4ABDA0
#define nox_xxx_spriteLoadFromMap_4AC020 game3_ref_spriteLoadFromMap_4AC020
#define sub_4B69F0 game3_ref_sub_4B69F0
#define sub_4B6970 game3_ref_sub_4B6970
#define sub_4AC7B0 game3_ref_sub_4AC7B0
#define sub_4ACEF0 game3_ref_sub_4ACEF0
#include "game3_map_reference.h"
#undef sub_4ABDA0
#undef nox_xxx_spriteLoadFromMap_4AC020
#undef sub_4B69F0
#undef sub_4B6970
#undef sub_4AC7B0
#undef sub_4ACEF0
#undef sub_410390
#undef nox_xxx_updateSpritePosition_49AA90
#undef sub_4B6880
#undef nox_xxx_spriteLoadError_4356E0
#undef gameFrame
#undef nox_xxx_createAtImpl_4191D0
#undef nox_xxx_getTTByNameSpriteMB_44CFC0
#undef nox_common_gameFlags_check_40A5C0
#undef nox_xxx_spriteLoadAdd_45A360_drawable
#undef nox_xxx_cryptSeekCur_40E0A0
#undef nox_xxx_fileReadWrite_426AC0_file3_fread
static size_t game3_ref_cursor(void){return game3_ref_pos;}
static short game3_ref_bounce(nox_drawable* d,unsigned int frame){game3_ref_clock=frame;return game3_ref_sub_4B69F0(d);}
static int game3_ref_load_kind(int id,int kind){return kind==0?game3_ref_sub_4AC7B0(id):game3_ref_sub_4ACEF0(id);}
static unsigned int game3_ref_reads_count(void){return game3_ref_read_count;}
static uint32_t game3_ref_read_value(unsigned int i,unsigned int j){if(i>=game3_ref_read_count||j>=7)abort();return game3_ref_reads[i][j];}
static int game3_ref_motion(nox_drawable* d){game3_ref_clock=1;return game3_ref_sub_4B6970(NULL,d,0,0);}
*/
import "C"

import (
	"runtime"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
)

func game3TestAlloc(n uintptr) unsafe.Pointer { return C.game3_test_alloc(C.size_t(n)) }
func game3TestFree(p unsafe.Pointer)          { C.game3_test_free(p) }
func game3TestGuards(p unsafe.Pointer, n uintptr) bool {
	return C.game3_test_guards(p, C.size_t(n)) != 0
}
func game3TestLayouts() [4]uintptr {
	return [4]uintptr{uintptr(C.game3_test_draw_size()), uintptr(C.game3_test_equipment_size()), uintptr(C.game3_test_modifier_offset()), uintptr(C.game3_test_draw_union_offset())}
}
func game3TestReference(raw []byte, d *client.Drawable, id int, version int16, fail bool) (*client.Drawable, uint32, int) {
	var source unsafe.Pointer
	if len(raw) > 0 {
		source = unsafe.Pointer(&raw[0])
	}
	f := C.int(0)
	if fail {
		f = 1
	}
	C.game3_ref_begin(source, C.size_t(len(raw)), (*C.nox_drawable)(d.C()), f, 1, 0)
	var count C.uint32_t
	p := C.game3_ref_spriteLoadFromMap_4AC020(C.int(id), C.short(version), &count)
	runtime.KeepAlive(raw)
	return (*client.Drawable)(unsafe.Pointer(p)), uint32(count), int(C.game3_ref_cursor())
}
func game3TestPublicMap(id int, version int16) (*client.Drawable, uint32) {
	var count C.uint32_t
	p := C.nox_xxx_spriteLoadFromMap_4AC020(C.int(id), C.short(version), &count)
	return (*client.Drawable)(unsafe.Pointer(p)), uint32(count)
}
func game3TestPublicLoad(id, kind int) int32 {
	switch kind {
	case 0:
		return int32(C.sub_4AC7B0(C.int(id)))
	case 1:
		return int32(C.nox_xxx_colorLightClientLoad_4AC980(C.int(id)))
	case 2:
		return int32(C.nox_xxx_cliLoadTeamBase_4ACE00(C.int(id)))
	case 3:
		return int32(C.sub_4ACEF0(C.int(id)))
	default:
		return int32(C.sub_4AD040(C.int(id)))
	}
}
func game3TestReferenceKind(raw []byte, d *client.Drawable, id, kind int) (int32, int, [][7]uint32) {
	var source unsafe.Pointer
	if len(raw) > 0 {
		source = unsafe.Pointer(&raw[0])
	}
	C.game3_ref_begin(source, C.size_t(len(raw)), (*C.nox_drawable)(d.C()), 0, 1, 0)
	result := int32(C.game3_ref_load_kind(C.int(id), C.int(kind)))
	out := make([][7]uint32, int(C.game3_ref_reads_count()))
	for i := range out {
		for j := range out[i] {
			out[i][j] = uint32(C.game3_ref_read_value(C.uint(i), C.uint(j)))
		}
	}
	runtime.KeepAlive(raw)
	return result, int(C.game3_ref_cursor()), out
}
func game3TestReferenceMotion(d *client.Drawable) int32 {
	return int32(C.game3_ref_motion((*C.nox_drawable)(d.C())))
}
func game3TestPublicMotion(d *client.Drawable) int32 {
	return int32(C.sub_4B6970(nil, (*C.nox_drawable)(d.C()), 0, 0))
}
func game3TestReferenceBounce(d *client.Drawable, frame uint32) int16 {
	return int16(C.game3_ref_bounce((*C.nox_drawable)(d.C()), C.uint(frame)))
}
func game3TestPublicBounce(d *client.Drawable) int16 {
	return int16(C.sub_4B69F0((*C.nox_drawable)(d.C())))
}
func game3TestItemModifier(d *client.Drawable, i int) unsafe.Pointer {
	return C.game3_test_item_modifier((*C.nox_drawable)(d.C()), C.int(i))
}
func game3TestSummonChild(d *client.Drawable) unsafe.Pointer {
	return C.game3_test_summon_child((*C.nox_drawable)(d.C()))
}
func game3TestPublicSummon(id int16, pos [2]uint16, thing uint16, dir byte, life int16) uintptr {
	return uintptr(C.game3_test_summon(C.short(id), (*C.ushort)(unsafe.Pointer(&pos[0])), C.ushort(thing), C.uchar(dir), C.short(life)))
}
func game3TestPublicRemoveSummon(id int16) { C.sub_4B7EE0(C.short(id)) }
func game3TestPublicShield(code uint32, dir int32) uintptr {
	return uintptr(C.game3_test_shield(C.uint(code), C.int(dir)))
}
func game3TestPublicMaterial(entries unsafe.Pointer, bit uintptr) uintptr {
	return uintptr(C.game3_test_material(entries, C.uintptr_t(bit)))
}
func game3TestOwner(n int) uint32       { return uint32(C.game3_test_owner(C.int(n))) }
func game3TestSetOwner(n int, v uint32) { C.game3_test_set_owner(C.int(n), C.uint(v)) }
