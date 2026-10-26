#include <stdio.h>
#include <limits.h>
#ifdef _WIN32
#include <windows.h>
#else
#include <sys/mman.h>
#endif
#include "../defs.h"
#include "../GAME2_1.h"
#include "../GAME3_1.h"
#include "../GAME5_2.h"
#include "../operators.h"

static nox_inventory_cell_t* grid;
static nox_drawable* drawables;
_Alignas(nox_inventory_cell_t) static uint8_t expected_grid[NOX_INVENTORY_CELLS_MAX*sizeof(nox_inventory_cell_t)];
static nox_drawable expected_drawables[NOX_INVENTORY_CELLS_MAX];
uintptr_t dword_5d4594_1062480;
uint32_t dword_5d4594_1062484;
static uintptr_t selected;
static uint32_t previous_code;
static uint16_t wire_code;
static int return_value, sends;
static unsigned cases;
int nox_xxx_netClientSend2_4E53C0(int recipient, const void* payload, int size, nox_object_t* related, int disconnect);
#define nox_client_inventory_grid_1050020 grid
#pragma GCC diagnostic push
#pragma GCC diagnostic ignored "-Wpointer-to-int-cast"
#pragma GCC diagnostic ignored "-Wsign-compare"
// SELECT_BODY
// CLEAR_BODY
// REPORT_BODY
// CODE_BODY
#pragma GCC diagnostic pop

static int grid_intact(void) {
	if (memcmp(grid,expected_grid,sizeof(expected_grid))) return 0;
	uint8_t* bytes=(uint8_t*)grid;
	for (int i=1;i<=32;i++) if (bytes[-i]!=0xa5 || bytes[sizeof(expected_grid)-1+i]!=0xa5) return 0;
	return 1;
}
static void die(const char* what) {
	fprintf(stderr,"case %u: %s grid=%p selected=%p\n",cases,what,(void*)grid,(void*)selected);
	exit(1);
}
int nox_xxx_netClientSend2_4E53C0(int recipient, const void* payload, int size, nox_object_t* related, int disconnect) {
	const uint8_t* bytes=payload;
	if (recipient!=31 || size!=3 || related!=NULL || disconnect!=1 ||
		bytes[0]!=0xe0 || ((uint16_t)bytes[1]|(uint16_t)bytes[2]<<8)!=wire_code) die("original three-byte secondary packet/routing");
	if (dword_5d4594_1062480!=selected || dword_5d4594_1062484!=previous_code) die("native selected pointer/full previous code before send");
	if (!grid_intact()) die("all occupied markers must be reset before send");
	if (memcmp(drawables,expected_drawables,sizeof(expected_drawables))) die("drawable changes exceed selected code before send");
	sends++;
	return return_value;
}
static const uint32_t codes[]={0,0x7fffu,0x8000u,UINT32_MAX};
static void seed(unsigned pattern, unsigned hole) {
	memset((uint8_t*)grid-32,0xa5,sizeof(expected_grid)+64);
	memset(grid,0x5a,sizeof(expected_grid));
	memset(drawables,0x69,sizeof(expected_drawables));
	for (unsigned i=0;i<NOX_INVENTORY_CELLS_MAX;i++) {
		grid[i].field_0=&drawables[i];
		grid[i].field_4=codes[(i+pattern)%4];
		grid[i].field_132=0xdead0000u+i;
		grid[i].field_136=0x80000000u+i;
		const uint8_t amounts[]={0,1,32,255};
		grid[i].field_140=i==hole ? 0 : amounts[(i+pattern)%4];
		drawables[i].field_32=0xfedc0000u+i;
		const uint32_t flags[]={0,0x00400000u,0x20000000u};
		drawables[i].flags28=flags[(i+pattern)%3];
	}
	memcpy(expected_grid,grid,sizeof(expected_grid));
	memcpy(expected_drawables,drawables,sizeof(expected_drawables));
	nox_inventory_cell_t* rows=(nox_inventory_cell_t*)expected_grid;
	for (unsigned i=0;i<NOX_INVENTORY_CELLS_MAX;i++) if (rows[i].field_140) rows[i].field_136=0;
	dword_5d4594_1062480=0;
	dword_5d4594_1062484=0x12345678u;
	sends=0;
}
static int run_case(const char* mode, unsigned pattern, unsigned first, unsigned second) {
	cases++;
	seed(pattern,first);
	if (strcmp(mode,"clear")==0) {
		char* expected=(char*)grid+offsetof(nox_inventory_cell_t,field_140)+
			(5*NOX_INVENTORY_ROW_COUNT-1)*sizeof(nox_inventory_cell_t);
		if (sub_4619F0()!=expected) die("original full native reset return");
		if (!grid_intact()) die("all 84 cells including four equipment cells; empty flags/padding retained");
		if (sends || dword_5d4594_1062480 || dword_5d4594_1062484!=0x12345678u ||
			memcmp(drawables,expected_drawables,sizeof(expected_drawables))) die("reset modified unrelated globals/drawables");
	} else if (strcmp(mode,"select")==0) {
		// Every old/new cell, including same-cell and nil transitions. The
		// selected record may be empty; original selector has no count gate.
		dword_5d4594_1062480=first<NOX_INVENTORY_CELLS_MAX ? (uintptr_t)&grid[first] : 0;
		previous_code=first<NOX_INVENTORY_CELLS_MAX ? grid[first].field_4 : 0;
		selected=second<NOX_INVENTORY_CELLS_MAX ? (uintptr_t)&grid[second] : 0;
		wire_code=0;
		if (selected) {
			uint32_t code=grid[second].field_4;
			((nox_inventory_cell_t*)expected_grid)[second].field_136=1;
			expected_drawables[second].field_32=code;
			if (code<0x8000) wire_code=(uint16_t)code | (drawables[second].flags28&0x20400000u ? 0x8000u : 0);
		}
		const int values[]={0,1,-17,INT_MAX}; return_value=values[pattern];
		int got=nox_xxx_clientSetAltWeapon_461550((nox_inventory_cell_t*)selected);
		if (got!=return_value || sends!=1 || !grid_intact() ||
			memcmp(drawables,expected_drawables,sizeof(expected_drawables))) die("selector return/one-send/whole-state contract");
	} else abort();
	return 0;
}
static void* map_near(uintptr_t address) {
#if UINTPTR_MAX <= UINT32_MAX
	/* Keep every data/guard case on native 32-bit allocations. Exact high
	 * address and 4-GiB boundary placements remain mandatory on 64-bit. */
	address=0;
#endif
#ifdef _WIN32
	return VirtualAlloc((void*)address,0x80000,MEM_RESERVE|MEM_COMMIT,PAGE_READWRITE);
#else
	void* ptr=mmap((void*)address,0x80000,PROT_READ|PROT_WRITE,MAP_PRIVATE|MAP_ANON,-1,0);
	return ptr==MAP_FAILED ? NULL : ptr;
#endif
}
static void release_map(void* ptr) {
#ifdef _WIN32
	VirtualFree(ptr,0,MEM_RELEASE);
#else
	munmap(ptr,0x80000);
#endif
}
int main(int argc, char** argv) {
	if (argc!=2) return 1;
	drawables=malloc(sizeof(expected_drawables)); if (!drawables) return 1;
	if (sizeof(void*)>4 && (uintptr_t)drawables<=UINT32_MAX) die("native drawables required");
	const uintptr_t bases[]={(uintptr_t)0x37ffc0000ull,(uintptr_t)0x3fffc0000ull,(uintptr_t)0x440000000ull};
	for (unsigned map=0;map<3;map++) {
		// Hints only; never overwrite an existing mapping with MAP_FIXED.
		void* ptr=map_near(bases[map]); if (!ptr) die("map native grid");
		grid=(nox_inventory_cell_t*)((uint8_t*)ptr+(map<2 ? 0x40000-1600 : 64));
		if (sizeof(void*)>4 && ((uintptr_t)ptr!=bases[map] || (uintptr_t)grid<=UINT32_MAX)) die("exact high-address test placement");
		unsigned seconds=strcmp(argv[1],"clear")==0 ? 1 : NOX_INVENTORY_CELLS_MAX+1;
		for (unsigned pattern=0;pattern<4;pattern++) for (unsigned first=0;first<=NOX_INVENTORY_CELLS_MAX;first++)
			for (unsigned second=0;second<seconds;second++) run_case(argv[1],pattern,first,second);
		release_map(ptr);
	}
	free(drawables);
	printf("%s: %u complete native secondary-grid cases; 84 cells, whole grid/drawables, native pointers, previous ID, packet/routing/returns and guards verified\n",argv[1],cases);
	return 0;
}
