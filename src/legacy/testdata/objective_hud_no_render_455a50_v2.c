#include <assert.h>
#include <stdio.h>
#include "GAME1.h"
#include "GAME1_2.h"
#include "GAME1_3.h"
#include "GAME2.h"
#include "client__gui__gui_ctf.h"
#include "operators.h"

static _Alignas(8) unsigned char mapped[128];
static nox_window ctf, fb, child;
static unsigned engine_flags;
static int lookups, gui_calls, creates;
nox_window* dword_5d4594_1045604;
nox_window* dword_5d4594_1045636;
uint32_t nox_player_netCode_85319C = 0x1234;
uint32_t dword_5d4594_1045640;

static void gui_call(void) {
    if (engine_flags & NOX_ENGINE_FLAG_DISABLE_GRAPHICS_RENDERING) {
        fprintf(stderr, "objective HUD reached GUI while NoRendering\n");
        abort();
    }
    gui_calls++;
}
uint8_t* mem_getU8Ptr(uintptr_t base, uintptr_t offset) {
    assert(base==0x5d4594 && offset>=1045600 && offset<1045728);
    return &mapped[offset-1045600];
}
uint32_t* mem_getU32Ptr(uintptr_t base, uintptr_t offset) {
    assert(offset+4<=1045728);
    return (uint32_t*)mem_getU8Ptr(base,offset);
}
bool nox_common_getEngineFlag(const nox_engine_flag flag) {
    return (engine_flags & (unsigned)flag)!=0;
}
nox_object_team_t* nox_xxx_objGetTeamByNetCode_418C80(int code) {
    assert(code==0x1234);lookups++;return NULL;
}
int sub_455C30(void) { gui_call();creates++;dword_5d4594_1045604=&ctf;return 1; }
int sub_456070(void) { gui_call();creates++;dword_5d4594_1045636=&fb;return 1; }
void nox_xxx_gameGetScreenBoundaries_43BEB0_get_video_mode(int* w,int* h,int* d) { gui_call();*w=1024;*h=768;*d=16; }
void nox_xxx_screenGetSize_430C50_get_video_max(int* w,int* h) { gui_call();*w=640;*h=480; }
void sub_46ACE0(nox_window* window,int first,int last,int hidden) { gui_call();assert(window==&ctf && first==8811 && last==8826 && hidden==1); }
nox_window* nox_xxx_wndGetChildByID_46B0C0(nox_window* window,int id) { gui_call();assert(window==&ctf && id>=8811 && id<=8826);return &child; }
int nox_window_set_hidden(nox_window* window,int hidden) { gui_call();assert(window==&child && hidden==0);return 1; }
int nox_xxx_wndShowModalMB_46A8C0(nox_window* window) { gui_call();assert(window==&ctf || window==&fb);return 55; }

/* Complete active bodies, inserted without any rewriting by the Go test. */
/* CTF_BODY */
/* FLAGBALL_BODY */

int main(int argc,char** argv) {
    assert(argc==2);
    int flagball=strcmp(argv[1],"flagball")==0;
    assert(flagball || strcmp(argv[1],"ctf")==0);
    unsigned modes[4]={0,NOX_ENGINE_FLAG_DISABLE_TEXT_RENDERING|NOX_ENGINE_FLAG_DISABLE_FLOOR_RENDERING,
        NOX_ENGINE_FLAG_DISABLE_GRAPHICS_RENDERING,NOX_ENGINE_FLAG_DISABLE_GRAPHICS_RENDERING|NOX_ENGINE_FLAG_ADMIN};
    unsigned patterns[3]={0,0x5a,0xff};
    for (int m=0;m<4;m++) for (int existing=0;existing<2;existing++) for (int p=0;p<3;p++) {
        engine_flags=modes[m];lookups=gui_calls=creates=0;
        dword_5d4594_1045640=0xa55aa55a;
        memset(mapped,patterns[p],sizeof(mapped));
        memset(&ctf,0,sizeof(ctf));memset(&fb,0,sizeof(fb));memset(&child,0,sizeof(child));
        ctf.width=fb.width=300;ctf.height=fb.height=200;
        ctf.off_x=fb.off_x=71;ctf.off_y=fb.off_y=73;
        dword_5d4594_1045604=existing ? &ctf : NULL;
        dword_5d4594_1045636=existing ? &fb : NULL;
        unsigned char before[sizeof(mapped)];memcpy(before,mapped,sizeof(mapped));
        nox_window ctf_before=ctf,fb_before=fb;
        int result=flagball ? sub_455F60() : (unsigned char)sub_455A50(2);
        assert(lookups==1);
        if (engine_flags & NOX_ENGINE_FLAG_DISABLE_GRAPHICS_RENDERING) {
            assert(result==0 && gui_calls==0 && creates==0);
            assert(dword_5d4594_1045640==0xa55aa55a);
            assert(memcmp(before,mapped,sizeof(mapped))==0);
            assert(memcmp(&ctf_before,&ctf,sizeof(ctf))==0 && memcmp(&fb_before,&fb,sizeof(fb))==0);
            assert(dword_5d4594_1045604==(existing ? &ctf : NULL));
            assert(dword_5d4594_1045636==(existing ? &fb : NULL));
        } else {
            assert(result==55 && gui_calls>0 && creates==!existing);
            if (flagball) {
                assert(fb.off_x==449 && fb.end_x==749 && fb.off_y==360 && fb.end_y==560);
                assert(dword_5d4594_1045640==1 && *mem_getU8Ptr(0x5d4594,1045644)==0);
            } else {
                assert(ctf.off_x==399 && ctf.end_x==699 && ctf.off_y==400 && ctf.end_y==600);
                assert(*mem_getU32Ptr(0x5d4594,1045608)==1 && *mem_getU8Ptr(0x5d4594,1045628)==2);
            }
        }
    }
    puts("complete native HUD body: 24 render/no-render, preexisting-window and guard-pattern cases passed");
    return 0;
}
