#include <stdio.h>
#include <limits.h>
#ifdef _WIN32
#include <windows.h>
#else
#include <sys/mman.h>
#endif
#include "../defs.h"
#include "../GAME1_2.h"
#include "../GAME1_3.h"
#pragma GCC diagnostic push
#pragma GCC diagnostic ignored "-Wsign-compare"
#ifdef __clang__
#pragma GCC diagnostic ignored "-Wincompatible-pointer-types-discards-qualifiers"
#else
#pragma GCC diagnostic ignored "-Wdiscarded-qualifiers"
#endif
#include "../noxstring.c"
#pragma GCC diagnostic pop

char* nox_itoa(int value,char* dst,int radix) {
	if (radix==10) sprintf(dst,"%d",value);
	else if (radix==8) sprintf(dst,"%o",(unsigned)value);
	else sprintf(dst,"%x",(unsigned)value);
	return dst;
}
int nox_win_width;
uint32_t dword_5d4594_825736;
uint32_t nox_color_black_2650656=0,nox_color_white_2523948=0x7fff;
static int32_t older=0x03e0;
static uint8_t *table,*shadows;
static nox_draw_viewport_t viewport;
static int64_t input[10],expected[32][5];
static uint32_t event_count,event_index,returned,cases,clear_cases,draw_cases;
static const wchar2_t sources[3][5]={{0},{0x41,0},{0xac00,0xd83d,0xde80,0x42,0}};
static void fail(const char* why) {
	fprintf(stderr,"case=%u mode=%lld events=%u/%u table=%p shadows=%p: %s\n",cases,(long long)input[0],event_index,event_count,(void*)table,(void*)shadows,why);
	exit(1);
}
static int kind(const void* text) {
	for (int i=0;i<3;i++) if (text==table+644*i) return i;
	fail("wrong native text pointer");return -1;
}
static void emit(int64_t op,int64_t a,int64_t b,int64_t c) {
	int64_t row[5]={op,a,b,c,0};
	if (event_index>=event_count || memcmp(row,expected[event_index],sizeof(row))) fail("original complete trace/coordinates/order differs");
	event_index++;
}
static void* center_mem(uintptr_t base,uintptr_t off) {
	if (base==0x5d4594 && off==2597996) return &older;
	if (base==0x5d4594 && ((off>=823804 && off<=825772) || off==826372)) return table+off-823804;
	if (base==0x587000 && off>=107848 && off<=107880) return shadows+off-107848;
	fail("unexpected memmap access");return NULL;
}
uint32_t gameFrame(void) { return (uint32_t)input[3]; }
nox_draw_viewport_t* nox_draw_getViewport_437250(void) { emit(1,0,0,0);return &viewport; }
void nox_xxx_drawSetTextColor_434390(int color) { emit(2,(uint32_t)color,0,0); }
int nox_xxx_drawGetStringSize_43F840(void* font,wchar2_t* text,int* width,int* height,int wrap) {
	if (font || !width || height || wrap) fail("original measure arguments");
	int row=kind(text);emit(3,row,0,0);*width=(int)input[7]+8*row;return 1;
}
int nox_xxx_drawString_43F6E0(void* font,wchar2_t* text,int x,int y) {
	if (font) fail("original default font");emit(4,kind(text),x,y);return (int)0xdeadbeefu;
}
int nox_xxx_guiFontHeightMB_43F320(void* font) { if (font) fail("original height font");emit(5,0,0,0);return (int)input[8]; }
static wchar2_t* center_copy(wchar2_t* dst,const wchar2_t* src) {
	if (src!=(wchar2_t*)(table+1936)) fail("original empty/source pointer");
	emit(6,kind(dst),0,0);return nox_wcscpy(dst,src);
}
#undef getMemAt
#undef getMemU32Ptr
#undef getMemU16Ptr
#undef getMemIntPtr
#define getMemAt center_mem
#define getMemU32Ptr(B,O) ((uint32_t*)center_mem(B,O))
#define getMemU16Ptr(B,O) ((uint16_t*)center_mem(B,O))
#define getMemIntPtr(B,O) ((int32_t*)center_mem(B,O))
#define nox_wcscpy center_copy
#pragma GCC diagnostic push
#pragma GCC diagnostic ignored "-Wpointer-to-int-cast"
#pragma GCC diagnostic ignored "-Wpointer-sign"
// CENTER_BODIES
#pragma GCC diagnostic pop
#undef nox_wcscpy

static void* map_near(uintptr_t address) {
#ifdef _WIN32
	return VirtualAlloc((void*)address,0x80000,MEM_RESERVE|MEM_COMMIT,PAGE_READWRITE);
#else
	void* p=mmap((void*)address,0x80000,PROT_READ|PROT_WRITE,MAP_PRIVATE|MAP_ANON,-1,0);
	return p==MAP_FAILED ? NULL : p;
#endif
}
static void release_map(void* p) {
#ifdef _WIN32
	VirtualFree(p,0,MEM_RELEASE);
#else
	munmap(p,0x80000);
#endif
}
int main(int argc,char** argv) {
	if (argc!=2) return 1;
	int mode=strcmp(argv[1],"clear")==0 ? 0 : strcmp(argv[1],"draw")==0 ? 1 : -1;
	if (mode<0) return 1;
	uint32_t count;int32_t offsets[8];
	if (fread(offsets,sizeof(offsets),1,stdin)!=1 || fread(&count,4,1,stdin)!=1) return 1;
	const uintptr_t bases[]={(uintptr_t)0x37ffc0000ull,(uintptr_t)0x3fffc0000ull,(uintptr_t)0x440000000ull};
	const uintptr_t shadow_bases[]={(uintptr_t)0x47ffc0000ull,(uintptr_t)0x4fffc0000ull,(uintptr_t)0x540000000ull};
	const unsigned clear_offsets[]={16,1280,1920,2048,4096},shadow_offsets[]={8,16,24,32,128};
	for (unsigned row=0;row<count;row++) {
		if (fread(input,sizeof(input),1,stdin)!=1 || fread(&event_count,4,1,stdin)!=1 || fread(expected,sizeof(expected),1,stdin)!=1 || fread(&returned,4,1,stdin)!=1) return 1;
		if ((int)input[0]!=mode) continue;
		for (unsigned mapping=0;mapping<3;mapping++) {
			void* main_map=map_near(bases[mapping]);void* shadow_map=map_near(shadow_bases[mapping]);
			if (!main_map || !shadow_map || (sizeof(void*)>4 && ((uintptr_t)main_map!=bases[mapping] || (uintptr_t)shadow_map!=shadow_bases[mapping]))) fail("exact high-address placements unavailable");
			for (unsigned placement=0;placement<5;placement++) for (unsigned pattern=0;pattern<3;pattern++) {
				uint8_t fill=pattern==0 ? 0 : pattern==1 ? 0x5a : 0xff;
				table=(uint8_t*)main_map+(mapping<2 ? 0x40000-clear_offsets[placement]-636 : 64+clear_offsets[placement]);
				shadows=(uint8_t*)shadow_map+(mapping<2 ? 0x40000-shadow_offsets[placement] : 64+shadow_offsets[placement]);
				memset(table-32,fill,1932+64);memset(shadows-32,fill,32+64);memcpy(shadows,offsets,32);
				dword_5d4594_825736=(uint32_t)input[1];event_index=0;cases++;
				nox_win_width=(int)input[4];viewport.height=(intptr_t)input[5];viewport.y1=(intptr_t)input[6];
				if (mode==0) memcpy(table+1936,sources[input[9]],sizeof(sources[0]));
				else for (unsigned i=0;i<3;i++) {
					((uint16_t*)(table+644*i))[0]=(uint16_t)(0x41+i);((uint16_t*)(table+644*i))[1]=0;
					uint32_t expires=input[2]&(1u<<i) ? (uint32_t)input[3] : (uint32_t)input[3]-1;
					memcpy(table+644*i+636,&expires,4);
				}
				uint8_t before[1932+64],want[1932+64],shadow_before[96];
				memcpy(before,table-32,sizeof(before));memcpy(want,before,sizeof(want));memcpy(shadow_before,shadows-32,sizeof(shadow_before));
				if (mode==0) {
					for (unsigned i=0;i<3;i++) { unsigned at=32+644*i;memcpy(want+at,sources[input[9]],2*(nox_wcslen(sources[input[9]])+1));memset(want+at+636,0,5); }
					wchar2_t* result=sub_445450();clear_cases++;
					if (result!=(wchar2_t*)(table+644*returned) || dword_5d4594_825736!=0) fail("original reset return/index");
				} else { draw_cases++;if ((uint32_t)nox_xxx_drawMessageLines_445530()!=returned || dword_5d4594_825736!=(uint32_t)input[1]) fail("original draw return/index"); }
				if (event_index!=event_count || memcmp(table-32,want,sizeof(want)) || memcmp(shadows-32,shadow_before,sizeof(shadow_before))) fail("complete events/message/shadow bytes or guards");
			}
			 release_map(main_map);release_map(shadow_map);
		}
	}
	printf("center %s: %u native cases; clear=%u draw=%u; complete original traces, UTF-16 copy, return/index, full packed bytes and native boundary guards verified\n",argv[1],cases,clear_cases,draw_cases);
	return 0;
}
