#include <assert.h>
#include <stdarg.h>
#include <stdio.h>
#include "defs.h"
#include "GAME3_2.h"
#include "GAME4_2.h"
#include "common/fs/nox_fs.h"

static uint8_t info[0x5B8];
static uint32_t packed_pointer_slot = 0xA1B2C3D4;
static void* side_pointer_slot;
static nox_object_t *generated[2], *preserved, *restored;
static unsigned char before[2][sizeof(nox_object_t)];
static unsigned switches, saves, flag_sets, flag_clears;
static int empty_generated;

void* mem_getPtr(uintptr_t base, uintptr_t off) {
	assert(base == 0x973F18 && off == 2408);
	return info;
}
uint32_t* mem_getU32Ptr(uintptr_t base, uintptr_t off) {
	assert(base == 0x5D4594 && off == 1550924);
	return &packed_pointer_slot;
}
int32_t* mem_getI32Ptr(uintptr_t base, uintptr_t off) {
	return (int32_t*)mem_getU32Ptr(base, off);
}
void* mem_getPtrValue(uintptr_t base, uintptr_t off) {
	assert(base == 0x5D4594 && off == 1550924);
	return side_pointer_slot;
}
void mem_setPtrValue(uintptr_t base, uintptr_t off, void* value) {
	assert(base == 0x5D4594 && off == 1550924);
	side_pointer_slot = value;
}
void nox_xxx_mapSwitchLevel_4D12E0(int save_players) {
	assert(save_players == 1);
	++switches;
}
void nox_xxx_setGameFlags_40A4D0(unsigned int flags) { assert(flags == 0x400000); ++flag_sets; }
void nox_common_gameFlags_unset_40A540(unsigned int flags) { assert(flags == 0x400000); ++flag_clears; }
nox_object_t* nox_get_and_zero_server_objects_4DA3C0(void) { return preserved; }
void nox_set_server_objects_4DA3E0(nox_object_t* objects) { restored = objects; }
nox_object_t* nox_server_getFirstObject_4DA790(void) { return empty_generated ? NULL : generated[0]; }
nox_object_t* nox_server_getNextObject_4DA7A0(nox_object_t* object) {
	assert(object == generated[0] || object == generated[1]);
	return object == generated[0] ? generated[1] : NULL;
}
int nox_xxx_mapGenStep_4D44E0(void) { return 1; }
char* nox_fs_root(void) { return "/owned"; }
int nox_fs_access(char* path, int mode) { assert(path && mode == 0); return -1; }
bool nox_fs_remove(const char* path) { assert(path); return true; }
bool nox_fs_move(char* source, char* destination) { assert(source && destination); return true; }
bool nox_fs_mkdir(char* path) { assert(!strcmp(path, "/owned\\Maps\\$probeclassic")); return true; }
char* nox_xxx_getRandMapName_4D4310(void) { return "probeclassic"; }
void nox_xxx_mapGenMakeInfo_4D5DB0(void* pointer) { assert(pointer == info); }
int nox_sprintf(char* target, const char* format, ...) {
	va_list args;
	va_start(args, format);
	// The extracted entry uses three 1024-byte path arrays.
	int n = vsnprintf(target, 1024, format, args);
	assert(n >= 0 && n < 1024);
	va_end(args);
	return n;
}
int nox_xxx_mapSaveMap_51E010(char* path, int compress) {
	assert(compress == 1 && !strcmp(path, "/owned\\Maps\\$probeclassic\\$probeclassic.map"));
	++saves;
	for (int i = 0; !empty_generated && i < 2; ++i) {
		unsigned char expected[sizeof(nox_object_t)];
		memcpy(expected, before[i], sizeof(expected));
		memset(expected + offsetof(nox_object_t, script_id), 0, sizeof(generated[i]->script_id));
		assert(!memcmp(expected, generated[i], sizeof(expected)));
	}
	return 1;
}

// C compatibility body before the Go entry port.
int nox_xxx_mapGenStart_4D4320() {
	int v0;                      // esi
	int v1;                      // eax
	char* v2;                    // eax
	char* v3;                    // eax
	int result;                  // eax
	nox_object_t* i;                       // eax
	char* v6;                    // eax
	char* v7;                    // eax
	char* v8;                    // [esp-8h] [ebp-C10h]
	char* v9;                    // [esp-4h] [ebp-C0Ch]
	char* v10;                   // [esp-4h] [ebp-C0Ch]
	char PathName[1024];         // [esp+8h] [ebp-C00h]
	char FileName[1024];         // [esp+408h] [ebp-800h]
	char ExistingFileName[1024]; // [esp+808h] [ebp-400h]

	v0 = 100;
	nox_xxx_mapSwitchLevel_4D12E0(1);
	nox_xxx_setGameFlags_40A4D0(0x400000);
	setMemPtr(0x5D4594, 1550924, nox_get_and_zero_server_objects_4DA3C0());
	memset(getMemAt(0x973F18, 2408), 0, 0x5B8u);
	while (1) {
		v1 = nox_xxx_mapGenStep_4D44E0();
		if (v1 == 1) {
			break;
		}
		if (v1) {
			if (--v0) {
				continue;
			}
		}
		return 0;
	}
	v2 = nox_fs_root();
	nox_sprintf(FileName, "%s\\nc.obj", v2);
	v3 = nox_fs_root();
	nox_sprintf(ExistingFileName, "%s\\blend.obj", v3);
	if (nox_fs_access(FileName, 0) == -1 || (result = nox_fs_remove(FileName))) {
		if (nox_fs_access(ExistingFileName, 0) == -1 || (result = nox_fs_move(ExistingFileName, FileName))) {
			for (i = nox_server_getFirstObject_4DA790(); i; i = nox_server_getNextObject_4DA7A0(i)) {
				i->script_id = 0;
			}
			nox_xxx_mapGenMakeInfo_4D5DB0(getMemAt(0x973F18, 2408));
			v9 = nox_xxx_getRandMapName_4D4310();
			v6 = nox_fs_root();
			nox_sprintf(PathName, "%s\\Maps\\$%s", v6, v9);
			nox_fs_mkdir(PathName);
			v10 = nox_xxx_getRandMapName_4D4310();
			v8 = nox_xxx_getRandMapName_4D4310();
			v7 = nox_fs_root();
			nox_sprintf(PathName, "%s\\Maps\\$%s\\$%s.map", v7, v8, v10);
			if (!nox_xxx_mapSaveMap_51E010(PathName, 1)) {
				return 0;
			}
			nox_xxx_mapSwitchLevel_4D12E0(1);
			nox_set_server_objects_4DA3E0((nox_object_t*)getMemPtr(0x5D4594, 1550924));
			setMemPtr(0x5D4594, 1550924, NULL);
			nox_common_gameFlags_unset_40A540(0x400000);
			result = 1;
		}
	}
	return result;
}

int main(int argc, char** argv) {
	empty_generated = argc > 1 && !strcmp(argv[1], "empty");
	for (int i = 0; i < 2; ++i) {
		generated[i] = malloc(sizeof(*generated[i]));
		assert(generated[i] && (uintptr_t)generated[i] > UINT32_MAX);
		memset(generated[i], 0x5A + i, sizeof(*generated[i]));
		generated[i]->script_id = i ? -123456 : 987654;
		generated[i]->extent = 0xDEADC0DEu + i;
		memcpy(before[i], generated[i], sizeof(*generated[i]));
	}
	preserved = malloc(sizeof(*preserved));
	assert(preserved && (uintptr_t)preserved > UINT32_MAX);
	memset(preserved, 0xC7, sizeof(*preserved));
	assert(nox_xxx_mapGenStart_4D4320() == 1);
	assert(restored == preserved);
	assert(side_pointer_slot == NULL && packed_pointer_slot == 0xA1B2C3D4);
	assert(switches == 2 && saves == 1 && flag_sets == 1 && flag_clears == 1);
	for (int i = 0; i < 2; ++i) free(generated[i]);
	free(preserved);
	printf("Classic entry native pointers and script_id: %s PASS\n", empty_generated ? "empty" : "generated");
	return 0;
}
