package legacy

/*
#include <stddef.h>
#include "game2_server_fixture.h"
static uintptr_t game2_server_public_call(uint32_t entry,void* p,int a,int b) {
 switch(entry){
 case 0x4535E0: return (uintptr_t)(intptr_t)sub_4535E0((int*)p);
 case 0x4535F0: return (uintptr_t)(intptr_t)sub_4535F0(a);
 case 0x453600: return (uintptr_t)(intptr_t)sub_453600();
 case 0x453610: return (uintptr_t)(intptr_t)sub_453610();
 case 0x453620: return (uintptr_t)(intptr_t)sub_453620((uint8_t*)p,(char)a,b);
 case 0x453640: return (uintptr_t)(intptr_t)sub_453640((uint32_t*)p,a,b);
 case 0x453660: return (uintptr_t)(intptr_t)sub_453660(a);
 case 0x453690: return (uintptr_t)(intptr_t)sub_453690(a);
 case 0x453F90: return (uintptr_t)(intptr_t)sub_453F90();
 case 0x453FA0: return (uintptr_t)(intptr_t)sub_453FA0(p,a,b);
 case 0x454000: return (uintptr_t)(intptr_t)sub_454000(p,a);
 case 0x4559B0: return (uintptr_t)(intptr_t)sub_4559B0((wchar2_t*)p);
 case 0x456E40: return (uintptr_t)(intptr_t)sub_456E40(a,b);
 case 0x456F10: return (uintptr_t)(intptr_t)sub_456F10((wchar2_t*)p,a);
 case 0x457140: return (uintptr_t)(intptr_t)sub_457140(a,(wchar2_t*)p);
 case 0x4571A0: return (uintptr_t)(intptr_t)sub_4571A0(a,b);
 case 0x459560: return (uintptr_t)(intptr_t)sub_459560(a);
 case 0x459870: return (uintptr_t)(intptr_t)sub_459870();
 case 0x459D50: return (uintptr_t)(intptr_t)sub_459D50(a);
 case 0x459D60: return (uintptr_t)(intptr_t)sub_459D60();
 case 0x459D80: return (uintptr_t)(intptr_t)sub_459D80(a);
 case 0x459DA0: return (uintptr_t)(intptr_t)sub_459DA0();
 case 0x459DB0: return (uintptr_t)(intptr_t)sub_459DB0((nox_drawable*)p);
 case 0x459EC0: return (uintptr_t)(intptr_t)nox_xxx_cliNextMinimapObj_459EC0((nox_drawable*)p);
 case 0x45A010: return (uintptr_t)(intptr_t)sub_45A010((nox_drawable*)p);
 case 0x45A070: return (uintptr_t)(intptr_t)nox_drawable_next_45A070((nox_drawable*)p);
 case 0x45A990: return (uintptr_t)(intptr_t)nox_xxx_spriteSetActiveMB_45A990_drawable((nox_drawable*)p);
 case 0x45AB80: return (uintptr_t)(intptr_t)nox_xxx_spriteSetFrameMB_45AB80_drawable((nox_drawable*)p,a);
 case 0x453F70: sub_453F70(p);return 0;
 case 0x457350: sub_457350(a,b);return 0;
 default:abort();
 }
}
static void** game2_server_fixture_owner(int field){switch(field){
case 0:return (void**)&dword_5d4594_1045468;
case 1:return (void**)&dword_5d4594_1045464;
case 2:return (void**)&dword_5d4594_1045484;
case 3:return (void**)&dword_5d4594_1045480;
case 4:return (void**)&dword_5d4594_1045508;
case 5:return (void**)&dword_5d4594_1045684;
case 6:return (void**)&dword_5d4594_1045688;
case 7:return (void**)&dword_5d4594_1045692;
case 8:return (void**)&dword_5d4594_1046492;
case 9:return (void**)&dword_5d4594_1045604;
case 10:return (void**)&dword_5d4594_1045636;
case 100:return (void**)&nox_gui_server_access.root;
case 101:return (void**)&nox_gui_server_access.active_list;
case 10102:return (void**)&nox_gui_server_access.wnd_10102;
case 10103:return (void**)&nox_gui_server_access.wnd_10103;
case 10104:return (void**)&nox_gui_server_access.wnd_10104;
case 10105:return (void**)&nox_gui_server_access.wnd_10105;
case 10109:return (void**)&nox_gui_server_access.wnd_10109;
case 10111:return (void**)&nox_gui_server_access.wnd_10111;
case 10112:return (void**)&nox_gui_server_access.wnd_10112;
case 10113:return (void**)&nox_gui_server_access.wnd_10113;
case 10123:return (void**)&nox_gui_server_access.wnd_10123;
case 10125:return (void**)&nox_gui_server_access.wnd_10125;
case 10126:return (void**)&nox_gui_server_access.wnd_10126;
case 10127:return (void**)&nox_gui_server_access.wnd_10127;
case 10128:return (void**)&nox_gui_server_access.wnd_10128;
case 10129:return (void**)&nox_gui_server_access.wnd_10129;
case 10130:return (void**)&nox_gui_server_access.wnd_10130;
case 10131:return (void**)&nox_gui_server_access.wnd_10131;
case 10132:return (void**)&nox_gui_server_access.wnd_10132;
case 10133:return (void**)&nox_gui_server_access.wnd_10133;
case 10200:return (void**)&nox_gui_server_access.wnd_10200;
default:abort();}}

static void game2_server_fixture_drawable(void* raw,uint32_t cl,uint32_t sub,uint32_t flags,uint32_t anim,uint32_t frame,uint32_t prev,void* link){nox_drawable* p=raw;p->flags28=cl;p->flags29=sub;p->flags30=flags;p->field_69=anim;p->field_77=frame;p->field_78=prev;p->field_92=link;p->field_102=link;p->field_104=link;}
static void* game2_server_fixture_head(int team){return team?(void*)&nox_gui_server_players.teams:(void*)&nox_gui_server_players.players;}
static size_t game2_server_fixture_layout(int field){switch(field){
case 0:return sizeof(nox_window);case 1:return sizeof(nox_window_data);case 2:return sizeof(nox_scrollListBox_data);case 3:return sizeof(nox_scrollListBox_item);
case 4:return sizeof(nox_list_item_t);case 5:return sizeof(nox_drawable);case 6:return offsetof(nox_gui_server_player_entry,name);case 7:return offsetof(nox_gui_server_player_entry,net_code);case 8:return offsetof(nox_gui_server_team_entry,team_id);case 9:return offsetof(nox_gui_server_team_entry,color);case 10:return sizeof(nox_gui_server_player_entry);case 11:return sizeof(nox_gui_server_team_entry);default:abort();}}
static void* game2_server_fixture_team(uint32_t id,uint8_t color,const wchar2_t* name){nox_gui_server_team_entry* p=calloc(1,sizeof(*p));if(!p)return NULL;p->list.field_0=&p->list;p->list.field_1=&p->list;p->team_id=id;p->color=color;nox_wcscpy(p->name,name);return p;}
static int game2_server_fixture_access_event(void* w,int event,uintptr_t a,uintptr_t b){return nox_xxx_windowAccessProc_454BA0(w,event,a,b);}
*/
import "C"
import (
	"runtime"
	"unsafe"
)

func game2ServerFixtureCall(old bool, entry uint32, p unsafe.Pointer, a, b int32) uintptr {
	var v C.uintptr_t
	if old {
		v = C.game2_server_reference_call(C.uint32_t(entry), p, C.int(a), C.int(b))
	} else {
		v = C.game2_server_public_call(C.uint32_t(entry), p, C.int(a), C.int(b))
	}
	runtime.KeepAlive(p)
	return uintptr(v)
}
func game2ServerFixtureOwner(field int) unsafe.Pointer {
	return *C.game2_server_fixture_owner(C.int(field))
}
func game2ServerFixtureSetOwner(field int, p unsafe.Pointer) {
	*C.game2_server_fixture_owner(C.int(field)) = p
	runtime.KeepAlive(p)
}
func game2ServerFixtureHead(team bool) *accessListNode416720 {
	v := C.int(0)
	if team {
		v = 1
	}
	return (*accessListNode416720)(C.game2_server_fixture_head(v))
}
func game2ServerFixtureLayout(field int) uintptr {
	return uintptr(C.game2_server_fixture_layout(C.int(field)))
}
func game2ServerFixtureAlloc(n uintptr) unsafe.Pointer { return C.calloc(1, C.size_t(n)) }
func game2ServerFixtureFree(p unsafe.Pointer)          { C.free(p) }
func game2ServerFixtureFirstVisible(p unsafe.Pointer) int {
	v := int(C.game2_server_reference_first_visible(p))
	runtime.KeepAlive(p)
	return v
}
func game2ServerFixtureTeam(id uint32, color byte, name unsafe.Pointer) unsafe.Pointer {
	p := C.game2_server_fixture_team(C.uint32_t(id), C.uint8_t(color), (*C.wchar2_t)(name))
	runtime.KeepAlive(name)
	return p
}
func game2ServerFixtureAccessEvent(w unsafe.Pointer, event int, a, b uintptr) int {
	v := int(C.game2_server_fixture_access_event(w, C.int(event), C.uintptr_t(a), C.uintptr_t(b)))
	runtime.KeepAlive(w)
	return v
}

func game2ServerFixtureClearHead(team bool) {
	h := game2ServerFixtureHead(team)
	for p := accessListNativeNext416720(h); p != nil; {
		next := accessListNativeNext416720(p)
		accessListNativeRemove416720(p)
		game2ServerFixtureFree(unsafe.Pointer(p))
		p = next
	}
	accessListNativeClear416720(h)
}
func game2ServerFixtureDrawable(p unsafe.Pointer, cl, sub, flags, anim, frame, prev uint32, link unsafe.Pointer) {
	C.game2_server_fixture_drawable(p, C.uint32_t(cl), C.uint32_t(sub), C.uint32_t(flags), C.uint32_t(anim), C.uint32_t(frame), C.uint32_t(prev), link)
	runtime.KeepAlive(link)
	runtime.KeepAlive(p)
}
func game2ServerFixtureNoRender(set bool) bool {
	old := bool(C.nox_common_getEngineFlag(C.NOX_ENGINE_FLAG_DISABLE_GRAPHICS_RENDERING))
	if set {
		C.nox_common_setEngineFlag(C.NOX_ENGINE_FLAG_DISABLE_GRAPHICS_RENDERING)
	} else {
		C.nox_common_resetEngineFlag(C.NOX_ENGINE_FLAG_DISABLE_GRAPHICS_RENDERING)
	}
	return old
}
func game2ServerFixtureHUD(ctf bool, count byte) int {
	if ctf {
		return int(C.sub_455A50(C.char(count)))
	}
	return int(C.sub_455F60())
}

func game2ServerFixtureControls() unsafe.Pointer { return unsafe.Pointer(C.sub_416630()) }
