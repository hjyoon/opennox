#include <stdio.h>
#include <limits.h>
#ifdef _WIN32
#include <windows.h>
#else
#include <sys/mman.h>
#endif
#include "../client__gui__guigen.h"
#include "../common__strman.h"
#include "../GAME1_2.h"
#include "../GAME1_3.h"
#include "../GAME2.h"
#include "../GAME2_2.h"
#include "../client__video__draw_common.h"

int nox_win_width, nox_win_height;
uint32_t nox_color_white_2523948=0x7fff;
static uint8_t* table;
static wchar2_t* texts[5];
static void* cached_text;
static int64_t input[10], expected[12][5];
static uint32_t expected_events, after[3], event_index, cases;
static int32_t green=0x03e0;
static void fail(const char* why) {
	fprintf(stderr,"case=%u event=%u cached=%p: %s\n",cases,event_index,cached_text,why);
	exit(1);
}
static int kind(void* text) {
	for (int i=0;i<5;i++) if (texts[i]==text) return i;
	fail("native cached UTF-16 pointer differs from original semantic text");
	return -1;
}
static void emit(int64_t op,int64_t a,int64_t b,int64_t c,int64_t d) {
	const int64_t row[5]={op,a,b,c,d};
	if (event_index>=expected_events || memcmp(row,expected[event_index],sizeof(row)))
		fail("original branch/argument/helper order differs");
	event_index++;
}
static void* mapgen_mem(uintptr_t blob, uintptr_t off) {
	if (blob==0x852978 && off==4) return &green;
	if (blob!=0x5d4594 || off<1309516 || off>=1309680) fail("unexpected memmap access");
	return table+off-1309516;
}
void* mem_getPtrValue(uintptr_t blob,uintptr_t off) {
	if (blob!=0x5d4594 || off!=1309660) fail("unexpected native pointer slot");
	return cached_text;
}
void mem_setPtrValue(uintptr_t blob,uintptr_t off,void* value) {
	if (blob!=0x5d4594 || off!=1309660) fail("unexpected native pointer slot store");
	cached_text=value;
}
void nox_xxx_clientPlaySoundSpecial_452D80(int a,int b) { emit(1,a,b,0,0); }
int sub_430B50(int a,int b,int c,int d) { emit(2,a,b,c,d);return -1; }
void nox_client_clearScreen_440900(void) { emit(3,0,0,0,0); }
void nox_client_drawImageAt_47D2C0(nox_video_bag_image_t* image,int x,int y) { emit(4,(uint32_t)(uintptr_t)image,x,y,0); }
wchar2_t* nox_strman_loadString_40F1D0(char* name,char** extra,char* source,int line) {
	int k=!strcmp(name,"Generating") ? 1 : !strcmp(name,"Assembling") ? 2 : !strcmp(name,"Populating") ? 3 : 0;
	if (!k || extra || strcmp(source,"C:\\NoxPost\\src\\client\\Gui\\guigen.c")) fail("original localization key/source");
	emit(5,k,line,0,0);return texts[k];
}
int nox_xxx_drawGetStringSize_43F840(void* font,wchar2_t* text,int* width,int* height,int max_width) {
	if (font || !width || height || max_width) fail("original measure contract");
	emit(6,kind(text),0,0,0);
	*width=(int)input[6];
	if (input[9]) cached_text=texts[4];
	return 1;
}
int nox_xxx_guiFontHeightMB_43F320(void* font) { if (font) fail("original default font");emit(7,0,0,0,0);return (int)input[7]; }
void nox_xxx_drawSetTextColor_434390(int color) { emit(8,(uint32_t)color,0,0,0); }
void nox_xxx_drawSetColor_4343E0(int color) { emit(9,(uint32_t)color,0,0,0); }
int nox_draw_drawStringHL_43F730(void* font,wchar2_t* text,int x,int y) {
	if (font) fail("original text font");
	emit(10,0,kind(text),x,y);return 777;
}
void nox_video_callCopyBackBuffer_4AD170(void) { emit(11,0,0,0,0); }
#undef getMemAt
#undef getMemU32Ptr
#undef getMemIntPtr
#define getMemAt mapgen_mem
#define getMemU32Ptr(B,O) ((uint32_t*)mapgen_mem(B,O))
#define getMemIntPtr(B,O) ((int32_t*)mapgen_mem(B,O))
#pragma GCC diagnostic push
#pragma GCC diagnostic ignored "-Wpointer-to-int-cast"
#pragma GCC diagnostic ignored "-Wint-conversion"
#pragma GCC diagnostic ignored "-Wpointer-sign"
// MAPGEN_BODY
#pragma GCC diagnostic pop

static void* native_map(void) {
#ifdef _WIN32
	return VirtualAlloc(NULL,4096,MEM_RESERVE|MEM_COMMIT,PAGE_READWRITE);
#else
	void* p=mmap(NULL,4096,PROT_READ|PROT_WRITE,MAP_PRIVATE|MAP_ANON,-1,0);
	return p==MAP_FAILED ? NULL : p;
#endif
}
int main(void) {
	uint32_t count;
	if (fread(&count,4,1,stdin)!=1) return 1;
	uint8_t* region=native_map();
	if (!region || (sizeof(void*)>4 && (uintptr_t)region<=UINT32_MAX)) fail("high-address native map required");
	table=region+64;
	for (unsigned i=0;i<5;i++) { texts[i]=(wchar2_t*)(region+512+i*64);texts[i][0]=(wchar2_t)('A'+i);texts[i][1]=0; }
	uint8_t* packet=region+1025;
	for (unsigned row=0;row<count;row++) {
		if (fread(input,sizeof(input),1,stdin)!=1 || fread(&expected_events,4,1,stdin)!=1 ||
		    fread(expected,sizeof(expected),1,stdin)!=1 || fread(after,sizeof(after),1,stdin)!=1) return 1;
		for (unsigned pattern=0;pattern<3;pattern++) {
			const uint8_t fill=pattern==0 ? 0 : pattern==1 ? 0x5a : 0xff;
			memset(table-32,fill,160+64);
			for (unsigned i=0;i<32;i++) ((uint32_t*)table)[i]=0x11000000+i;
			for (unsigned i=0;i<4;i++) ((uint32_t*)table)[32+i]=0x22000000+i;
			((uint32_t*)table)[38]=(uint32_t)input[2];((uint32_t*)table)[39]=(uint32_t)input[3];
			cached_text=texts[input[8]];event_index=0;cases++;
			nox_win_width=(int)input[4];nox_win_height=(int)input[5];
			memset(packet-16,fill,35);
			packet[0]=(uint8_t)input[0];packet[1]=(uint8_t)input[1];packet[2]=(uint8_t)(input[1]>>8);
			uint8_t before[160];memcpy(before,table,sizeof(before));
			short returned=nox_xxx_mapGenClientText_4A9D00(packet);
			if (returned!=(short)input[1] || event_index!=expected_events) fail("decoded short API/complete semantic events");
			if (((uint32_t*)table)[38]!=after[0] || ((uint32_t*)table)[39]!=after[1] || kind(cached_text)!=(int)after[2]) fail("original final sequence/frame/live cached text");
			if (memcmp(table,before,152)) fail("native pointer changed packed image/text/neighbor DWORD bytes");
			for (unsigned i=0;i<32;i++) if (table[-(int)i-1]!=fill || table[160+i]!=fill) fail("table guards");
			for (unsigned i=0;i<16;i++) if (packet[-(int)i-1]!=fill || packet[3+i]!=fill) fail("packet guards");
			if (packet[0]!=(uint8_t)input[0] || packet[1]!=(uint8_t)input[1] || packet[2]!=(uint8_t)(input[1]>>8)) fail("packet bytes changed");
		}
	}
	printf("%u complete high-address map-generation cases; original semantic branch/order/coordinates/live text/frames, decoded API, packed neighbors and unaligned packet guards verified\n",cases);
	return 0;
}
