#include <assert.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static int allow_open, allow_map, allow_compress;
static unsigned opens, writes, flushes, maps, patches, closes, compressions;
static const char* original;
static char compressed[1024];
static uint32_t suffix = UINT32_C(0x7A786E2E);

static uint32_t* getMemU32Ptr(uintptr_t base, uintptr_t off) {
 assert(base==0x587000 && off==253112);return &suffix;
}
static uint8_t getMemByte(uintptr_t base, uintptr_t off) {
 assert(base==0x587000 && off==253116);return 0;
}
static int nox_xxx_cryptOpen_426910(char* path,int mode,int key) {
 assert(!strcmp(path,original) && mode==0 && key==19);++opens;return allow_open;
}
static void nox_xxx_fileReadWrite_426AC0_file3_fread(void* value,unsigned size) {
 assert(size==4 && *(uint32_t*)value==UINT32_C(0xFADEFACE));++writes;
}
static int nox_xxx_cryptFlush_4268E0(void) {++flushes;return 73;}
static int nox_xxx_map_51E140(void) {++maps;return allow_map;}
static void sub_4268F0(int offset) {assert(offset==73);++patches;}
static void nox_xxx_cryptClose_4269F0(void) {++closes;}
int nox_xxx_nxzCompressFile_57BDD0(char* from,char* to) {
 assert(!strcmp(from,original) && !strcmp(to,compressed));
#if UINTPTR_MAX>UINT32_MAX
 assert((uintptr_t)from>UINT32_MAX && (uintptr_t)to>UINT32_MAX);
#endif
 ++compressions;return allow_compress;
}

// C compatibility save body before the Go port.
int nox_xxx_mapSaveMap_51E010(char* a1, int a2) {
	char* v2;         // edi
	unsigned char v3; // cl
	int result;       // eax
	int v5;           // esi
	int v7;           // [esp+10h] [ebp-804h]
	char v8[1024];    // [esp+14h] [ebp-800h]
	char Mem[1024];   // [esp+414h] [ebp-400h]

	v7 = -86050098;
	strcpy(Mem, a1);
	v8[0] = 0;
	strncat(v8, a1, 1024-1);
	v8[strlen(v8)-4] = 0;
	v2 = &v8[strlen(v8) + 1];
	v3 = getMemByte(0x587000, 253116);
	uint32_t suffix = *getMemU32Ptr(0x587000, 253112);
	memcpy(--v2, &suffix, sizeof(suffix));
	v2[4] = v3;
	result = nox_xxx_cryptOpen_426910(Mem, 0, 19);
	if (result) {
		nox_xxx_fileReadWrite_426AC0_file3_fread(&v7, 4u);
		v5 = nox_xxx_cryptFlush_4268E0();
		if (nox_xxx_map_51E140()) {
			sub_4268F0(v5);
			nox_xxx_cryptClose_4269F0();
			if (!a2 || (result = nox_xxx_nxzCompressFile_57BDD0(Mem, v8)) != 0) {
				result = 1;
			}
		} else {
			nox_xxx_cryptClose_4269F0();
			result = 0;
		}
	}
	return result;
}

int main(void) {
 unsigned cases=0;
 const char* paths[]={"maps/probe.map","owned data/a long theme name.map","maps/$native/$native.map"};
 for(unsigned p=0;p<3;++p) for(int op=0;op<2;++op) for(int mp=0;mp<2;++mp)
 for(int request=0;request<2;++request) for(int cp=0;cp<2;++cp) {
  original=paths[p];strcpy(compressed,original);strcpy(compressed+strlen(compressed)-4,".nxz");
  allow_open=op;allow_map=mp;allow_compress=cp;
  opens=writes=flushes=maps=patches=closes=compressions=0;
  int got=nox_xxx_mapSaveMap_51E010((char*)original,request);
  assert(got==(op && mp && (!request || cp)));
  assert(opens==1 && writes==(unsigned)op && flushes==(unsigned)op && maps==(unsigned)op);
  assert(patches==(unsigned)(op&&mp) && closes==(unsigned)op && compressions==(unsigned)(op&&mp&&request));
  ++cases;
 }
 printf("Mapgen save native R311: %u original-domain cases PASS\n",cases);return 0;
}
