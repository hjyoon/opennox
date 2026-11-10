/* Exact current-C definitions, renamed only by preprocessor tokens.
 * Not an original-CPU oracle; valid-input regression reference only. */
#ifndef OPENNOX_GAME4_MAP_ARCHIVE_REFERENCE_H
#define OPENNOX_GAME4_MAP_ARCHIVE_REFERENCE_H
#include <limits.h>
#include <math.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "GAME4.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME1_2.h"
#include "common/fs/nox_fs.h"
#include "common__crypt.h"
#include "noxstring.h"
extern char* dword_5d4594_1599576;
extern void* dword_5d4594_1599588;
extern void* dword_5d4594_1599592;
extern uint32_t dword_5d4594_1599596;
extern uint32_t dword_5d4594_1599480;
extern uint32_t dword_5d4594_1599476;
extern uint32_t dword_5d4594_1599644;
extern uint32_t dword_5d4594_3835396;
extern FILE* nox_file_8;
extern int nox_mapgenRewriteAreaMap_502ED0(char*, char*, char*);
extern int nox_mapgenPrepareAreaMap_503140(char*, char*);
extern int nox_mapgenRenameAreaMap_503230(char*, char*, char*, char*);
#define sub_5029A0 game4_ref_sub_5029A0
#define sub_5029F0 game4_ref_sub_5029F0
#define sub_502A20 game4_ref_sub_502A20
#define sub_502A30 game4_ref_sub_502A30
#define sub_502A50 game4_ref_sub_502A50
#define sub_502A90 game4_ref_sub_502A90
#define sub_502AB0 game4_ref_sub_502AB0
#define sub_502AF0 game4_ref_sub_502AF0
#define nox_mapgen_read_exact_502B10 game4_ref_nox_mapgen_read_exact_502B10
#define sub_502B10 game4_ref_sub_502B10
#define sub_502D70 game4_ref_sub_502D70
#define sub_502DA0 game4_ref_sub_502DA0
#define sub_502DF0 game4_ref_sub_502DF0
#define sub_502E10 game4_ref_sub_502E10
#define sub_502E50 game4_ref_sub_502E50
#define sub_502E70 game4_ref_sub_502E70
#define sub_502EA0 game4_ref_sub_502EA0
#define sub_502ED0 game4_ref_sub_502ED0
#define sub_503140 game4_ref_sub_503140
#define sub_503230 game4_ref_sub_503230
#define nox_mapgen_read_extract_prefix_5034B0 game4_ref_nox_mapgen_read_extract_prefix_5034B0
#define nox_mapgen_copy_exact_5034B0 game4_ref_nox_mapgen_copy_exact_5034B0
#define sub_5034B0 game4_ref_sub_5034B0
#define sub_5036D0 game4_ref_sub_5036D0
#define nox_xxx_mapgenSaveMap_503830 game4_ref_nox_xxx_mapgenSaveMap_503830
#define nox_mapgenClearPendingScriptIDs_503B30 game4_ref_nox_mapgenClearPendingScriptIDs_503B30
#define nox_mapgenTruncQwordLow_503B30 game4_ref_nox_mapgenTruncQwordLow_503B30
#define nox_mapgenBuildPlaceBounds_503B30 game4_ref_nox_mapgenBuildPlaceBounds_503B30
#define nox_mapgenPlaceOffset_503B30 game4_ref_nox_mapgenPlaceOffset_503B30
#define nox_mapgenWallSpan_503B30 game4_ref_nox_mapgenWallSpan_503B30
#define nox_mapgenFixCoordsNative_503B30 game4_ref_nox_mapgenFixCoordsNative_503B30
#define nox_mapgenPlaceGeometryNative_503B30 game4_ref_nox_mapgenPlaceGeometryNative_503B30
#define sub_503B30 game4_ref_sub_503B30
#define sub_503EC0 game4_ref_sub_503EC0
static int sub_5029A0(char* a1);
static char* sub_5029F0(int a1);
static int sub_502A20();
static int sub_502A30(char* name);
static int sub_502A50(char* a1);
static char* sub_502A90();
static int sub_502AB0(char* a1);
static char* sub_502AF0();
static int nox_mapgen_read_exact_502B10(FILE* file, void* dst, int size);
static int sub_502B10();
static int sub_502D70(int a1);
static FILE* sub_502DA0(char* a1);
static FILE* sub_502DF0();
static FILE* sub_502E10(int a1);
static FILE* sub_502E50(char* name);
static double sub_502E70(int a1);
static double sub_502EA0(int a1);
static int sub_502ED0(char* name);
static int sub_503140(void);
static int sub_503230(char* old_name, char* new_name);
static int nox_mapgen_read_extract_prefix_5034B0(FILE* input, int32_t* remaining, uint8_t* version);
static int nox_mapgen_copy_exact_5034B0(FILE* input, FILE* output, int32_t count);
static int sub_5034B0(char* name);
static int sub_5036D0(char* name, char* output_path);
static int nox_xxx_mapgenSaveMap_503830(int a1);
static void nox_mapgenClearPendingScriptIDs_503B30(nox_object_t* first);
static int32_t nox_mapgenTruncQwordLow_503B30(double value);
static void nox_mapgenBuildPlaceBounds_503B30(
	const float2* at, const float2* fixed_at, const void* record, int32_t corners[8], int4* bounds);
static int32_t nox_mapgenPlaceOffset_503B30(float fixed_coordinate, int32_t map_origin);
static float nox_mapgenWallSpan_503B30(uint32_t wall_count);
static int nox_mapgenFixCoordsNative_503B30(float x, float y, float* fixed_x, float* fixed_y);
static int nox_mapgenPlaceGeometryNative_503B30(float x, float y, float fixed_x, float fixed_y,
	const void* record, int32_t map_origin_x, int32_t map_origin_y, int32_t bounds[4], int32_t delta[2]);
static int sub_503B30(float2* a1);
static int sub_503EC0(nox_object_t* object, float2* out);
typedef struct nox_mapgen_name_record_502B10 { char name[64]; uint32_t x_bits; uint32_t y_bits; uint32_t file_offset; } nox_mapgen_name_record_502B10;
_Static_assert(sizeof(nox_mapgen_name_record_502B10) == 76, "wire record layout");

static int sub_5029A0(char* a1) {
	int v1; // edi
	int i;  // esi

	v1 = 0;
	if (*(int*)&dword_5d4594_1599596 <= 0) {
		return -1;
	}
	for (i = 0; nox_strcmpi(a1, (const char*)(i + dword_5d4594_1599576)); i += 76) {
		if (++v1 >= *(int*)&dword_5d4594_1599596) {
			return -1;
		}
	}
	return v1;
}

static char* sub_5029F0(int a1) {
	char* result; // eax

	if (a1 < 0 || a1 > *(int*)&dword_5d4594_1599596) {
		result = 0;
	} else {
		result = dword_5d4594_1599576 + 76 * a1;
	}
	return result;
}

static int sub_502A20() { return dword_5d4594_1599596; }

static int sub_502A30(char* name) { return sub_502D70(sub_5029A0(name)); }

static int sub_502A50(char* a1) {
	int result; // eax

	sub_502DF0();
	if (a1) {
		strncpy(dword_5d4594_1599588, a1, 0x7FFu);
		result = 1;
	} else {
		**(uint8_t**)&dword_5d4594_1599588 = getMemByte(0x5D4594, 1599608);
		result = 0;
	}
	return result;
}

static char* sub_502A90() {
	char* name = dword_5d4594_1599588;
	return strlen(name) ? name : NULL;
}

static int sub_502AB0(char* a1) {
	int result; // eax

	if (a1) {
		strncpy(dword_5d4594_1599592, a1, 0x7FFu);
		result = 1;
	} else {
		**(uint8_t**)&dword_5d4594_1599592 = getMemByte(0x5D4594, 1599612);
		result = 0;
	}
	return result;
}

static char* sub_502AF0() {
	char* name = dword_5d4594_1599592;
	return strlen(name) ? name : NULL;
}

static int nox_mapgen_read_exact_502B10(FILE* file, void* dst, int size) {
	uint8_t* next = dst;
	while (size > 0) {
		int n = nox_fs_fread(file, next, size);
		if (n <= 0 || n > size) {
			return 0;
		}
		next += n;
		size -= n;
	}
	return 1;
}

static int sub_502B10() {
	dword_5d4594_1599596 = 0;
	if (!dword_5d4594_1599588) {
		dword_5d4594_1599588 = calloc(1, 0x800u);
	}
	if (!dword_5d4594_1599592) {
		dword_5d4594_1599592 = calloc(1, 0x800u);
	}
	if (!dword_5d4594_1599576) {
		dword_5d4594_1599576 = calloc(1, 0x26000u);
	}
	if (!dword_5d4594_1599588 || !dword_5d4594_1599592 || !dword_5d4594_1599576) {
		return 0;
	}

	char* first_name = dword_5d4594_1599588;
	// The original calls strlen here. Its 2048-byte path buffer can be left
	// unterminated by sub_502A50, so do not let this scan escape that buffer.
	if (!first_name[0] || !memchr(first_name, 0, 0x800u)) {
		return 0;
	}
	if (!sub_502DA0(first_name)) {
		return 0;
	}

	int result = 0;
	uint32_t magic;
	if (!nox_mapgen_read_exact_502B10(nox_file_8, &magic, sizeof(magic)) || magic != 0xCAFEDEADu) {
		goto done;
	}
	for (;;) {
		int32_t record_length;
		if (!nox_mapgen_read_exact_502B10(nox_file_8, &record_length, sizeof(record_length))) {
			goto done;
		}
		if (record_length == 0) {
			result = 1;
			break;
		}
		if (dword_5d4594_1599596 >= 2048 || record_length < 11) {
			goto done;
		}

		long after_length = nox_fs_ftell(nox_file_8);
		if (after_length < 4 || (uint64_t)(after_length - 4) > UINT32_MAX) {
			goto done;
		}
		uint8_t name_length;
		if (!nox_mapgen_read_exact_502B10(nox_file_8, &name_length, 1) ||
			name_length >= 64 || record_length < 11 + name_length) {
			goto done;
		}
		char name[64];
		if (!nox_mapgen_read_exact_502B10(nox_file_8, name, name_length)) {
			goto done;
		}
		name[name_length] = 0;
		uint8_t flags[2];
		uint32_t x_bits, y_bits;
		if (!nox_mapgen_read_exact_502B10(nox_file_8, flags, sizeof(flags)) ||
			!nox_mapgen_read_exact_502B10(nox_file_8, &x_bits, sizeof(x_bits)) ||
			!nox_mapgen_read_exact_502B10(nox_file_8, &y_bits, sizeof(y_bits))) {
			goto done;
		}

		nox_mapgen_name_record_502B10* record =
			(nox_mapgen_name_record_502B10*)(dword_5d4594_1599576 + 76 * dword_5d4594_1599596);
		// strcpy in the original stops at an embedded NUL, even though the
		// full declared name length is consumed from the file.
		size_t copy_length = strlen(name) + 1;
		memcpy(record->name, name, copy_length);
		record->x_bits = x_bits;
		record->y_bits = y_bits;
		record->file_offset = (uint32_t)(after_length - 4);
		++dword_5d4594_1599596;

		if (nox_fs_fseek_cur(nox_file_8, record_length - 11 - name_length) != 0) {
			goto done;
		}
	}
done:
	sub_502DF0();
	return result;
}

static int sub_502D70(int a1) {
	if (a1 < 0 || a1 >= *(int*)&dword_5d4594_1599596) {
		return 0;
	}
	dword_5d4594_3835396 = a1;
	return nox_xxx_mapgenSaveMap_503830(a1) == 0;
}

static FILE* sub_502DA0(char* a1) {
	FILE* result; // eax

	result = nox_file_8;
	if (nox_file_8 || (result = (FILE*)nox_xxx_cryptOpen_426910(a1, 1, -1)) != 0 &&
						  (result = nox_xxx_mapgenGetSomeFile_426A60(), (nox_file_8 = result) != 0)) {
		nox_fs_fseek_start(result, 0);
		result = (FILE*)1;
	}
	return result;
}

static FILE* sub_502DF0() {
	FILE* result; // eax

	result = nox_file_8;
	if (nox_file_8) {
		nox_xxx_cryptClose_4269F0();
		nox_file_8 = 0;
	}
	return result;
}

static FILE* sub_502E10(int a1) {
	if (!nox_file_8 || a1 < 0 || a1 >= *(int*)&dword_5d4594_1599596) {
		return 0;
	}
	nox_fs_fseek_start(nox_file_8, *(uint32_t*)(dword_5d4594_1599576 + 76 * a1 + 72));
	return nox_file_8;
}

static FILE* sub_502E50(char* name) {
	if (!name) {
		return NULL;
	}
	return sub_502E10(sub_5029A0(name));
}

static double sub_502E70(int a1) {
	double result; // st7

	if (a1 < 0 || a1 >= *(int*)&dword_5d4594_1599596) {
		result = -1.0;
	} else {
		result = *(float*)(dword_5d4594_1599576 + 76 * a1 + 64);
	}
	return result;
}

static double sub_502EA0(int a1) {
	double result; // st7

	if (a1 < 0 || a1 >= *(int*)&dword_5d4594_1599596) {
		result = -1.0;
	} else {
		result = *(float*)(dword_5d4594_1599576 + 76 * a1 + 68);
	}
	return result;
}

static int sub_502ED0(char* name) {
	if (!name || !sub_503140()) {
		return 0;
	}
	int copied_any = nox_mapgenRewriteAreaMap_502ED0(name, dword_5d4594_1599588, getMemAt(0x973F18, 42152));
	if (copied_any < 0) {
		return 0;
	}
	sub_502B10();
	return copied_any;
}

static int sub_503140(void) {
	sub_502DF0();
	return nox_mapgenPrepareAreaMap_503140(dword_5d4594_1599588, getMemAt(0x973F18, 42152));
}

static int sub_503230(char* old_name, char* new_name) {
	if (!old_name || !new_name || !sub_503140()) {
		return 0;
	}
	if (!nox_mapgenRenameAreaMap_503230(old_name, new_name, dword_5d4594_1599588, getMemAt(0x973F18, 42152))) {
		return 0;
	}
	sub_502B10();
	return 1;
}

static int nox_mapgen_read_extract_prefix_5034B0(FILE* input, int32_t* remaining, uint8_t* version) {
	int32_t length;
	uint8_t name_length;
	char name[64];
	uint8_t flags[2];
	uint8_t coordinates[8];
	if (!nox_mapgen_read_exact_502B10(input, &length, 4) ||
		!nox_mapgen_read_exact_502B10(input, &name_length, 1) ||
		name_length >= sizeof(name) || length < 11 + name_length ||
		!nox_mapgen_read_exact_502B10(input, name, name_length) ||
		!nox_mapgen_read_exact_502B10(input, flags, sizeof(flags)) ||
		!nox_mapgen_read_exact_502B10(input, coordinates, sizeof(coordinates))) {
		return 0;
	}
	*remaining = length - 11 - name_length;
	*version = flags[1];
	return 1;
}

static int nox_mapgen_copy_exact_5034B0(FILE* input, FILE* output, int32_t count) {
	uint8_t buffer[4096];
	while (count > 0) {
		int chunk = count < (int)sizeof(buffer) ? count : (int)sizeof(buffer);
		if (!nox_mapgen_read_exact_502B10(input, buffer, chunk) || nox_fs_fwrite(output, buffer, chunk) != chunk) {
			return 0;
		}
		count -= chunk;
	}
	return 1;
}

static int sub_5034B0(char* name) {
	char* source = sub_502A90();
	if (!name || !source) {
		return 0;
	}
	char* directory = getMemAt(0x973F18, 42152);
	size_t dir_length = 0;
	while (dir_length < 0x800 && directory[dir_length]) {
		++dir_length;
	}
	static const char suffix[] = "\\copy.tmp";
	if (!dir_length || dir_length + sizeof(suffix) > 0x800) {
		return 0;
	}
	char output_path[0x800];
	memcpy(output_path, directory, dir_length);
	memcpy(output_path + dir_length, suffix, sizeof(suffix));

	sub_502DA0(source);
	FILE* input = sub_502E50(name);
	if (!input) {
		sub_502DF0();
		return 0;
	}
	int32_t remaining;
	uint8_t version;
	if (!nox_mapgen_read_extract_prefix_5034B0(input, &remaining, &version)) {
		sub_502DF0();
		return 0;
	}
	if (version > 1) {
		int32_t extra_length;
		if (remaining < 4 || !nox_mapgen_read_exact_502B10(input, &extra_length, 4)) {
			sub_502DF0();
			return 0;
		}
		remaining -= 4;
		if (extra_length < 0 || extra_length > remaining || nox_fs_fseek_cur(input, extra_length) != 0) {
			sub_502DF0();
			return 0;
		}
		remaining -= extra_length;
	}
	uint32_t magic;
	if (remaining < 4 || !nox_mapgen_read_exact_502B10(input, &magic, 4) || magic != 0xCAFEDEAD) {
		sub_502DF0();
		return 0;
	}
	remaining -= 4;
	FILE* output = nox_fs_create(output_path);
	if (!output) {
		sub_502DF0();
		return 0;
	}
	uint32_t output_magic = 0xABEDFACE;
	int ok = nox_fs_fwrite(output, &output_magic, 4) == 4 &&
		nox_mapgen_copy_exact_5034B0(input, output, remaining);
	sub_502DF0();
	nox_fs_close(output);
	if (!ok) {
		nox_fs_remove(output_path);
	}
	return ok;
}

static int sub_5036D0(char* name, char* output_path) {
	if (!name || !output_path) {
		return 0;
	}
	// DeleteFileA is the original first operation, including failure paths.
	nox_fs_remove(output_path);
	char* source = sub_502A90();
	if (!source) {
		return 0;
	}
	sub_502DA0(source);
	FILE* input = sub_502E50(name);
	if (!input) {
		sub_502DF0();
		return 0;
	}
	int32_t remaining;
	uint8_t version;
	if (!nox_mapgen_read_extract_prefix_5034B0(input, &remaining, &version) || version <= 1 || remaining < 4) {
		sub_502DF0();
		return 0;
	}
	int32_t extra_length;
	if (!nox_mapgen_read_exact_502B10(input, &extra_length, 4) || extra_length <= 0 || extra_length > remaining - 4) {
		sub_502DF0();
		return 0;
	}
	FILE* output = nox_fs_create(output_path);
	if (!output) {
		sub_502DF0();
		return 0;
	}
	int ok = nox_mapgen_copy_exact_5034B0(input, output, extra_length);
	sub_502DF0();
	nox_fs_close(output);
	if (!ok) {
		nox_fs_remove(output_path);
	}
	return ok;
}

static int nox_xxx_mapgenSaveMap_503830(int a1) {
	if (a1 < 0 || a1 >= *(int*)&dword_5d4594_1599596) {
		return 0;
	}
	nox_xxx_free_503F40();
	*getMemU32Ptr(0x5D4594, 1599572) = -1;
	dword_5d4594_1599644 = 0;
	if (!sub_502DA0(dword_5d4594_1599588) || !sub_502E10(a1)) {
		goto fail;
	}
	if (!nox_mapgenLoadNative_503830(a1)) {
		goto fail;
	}
	dword_5d4594_1599480 = a1;
	dword_5d4594_1599476 = 0;
	dword_5d4594_3835396 = a1;
	sub_502DF0();
	return 1;
fail:
	sub_502DF0();
	return 0;
}

static void nox_mapgenClearPendingScriptIDs_503B30(nox_object_t* first) {
	for (nox_object_t* obj = first; obj; obj = nox_server_getNextObjectUninited_4DA880(obj)) {
		obj->script_id = 0;
	}
}

static int32_t nox_mapgenTruncQwordLow_503B30(double value) {
	if (!isfinite(value) || value < -0x1p63 || value >= 0x1p63) {
		return 0;
	}
	uint32_t low = (uint32_t)(uint64_t)(int64_t)value;
	int32_t result;
	memcpy(&result, &low, sizeof(result));
	return result;
}

static void nox_mapgenBuildPlaceBounds_503B30(
	const float2* at, const float2* fixed_at, const void* record, int32_t corners[8], int4* bounds) {
	const uint8_t* bytes = record;
	float width, height;
	memcpy(&width, bytes + 64, sizeof(width));
	memcpy(&height, bytes + 68, sizeof(height));
	float2 corner;
	float2 fixed;

	corners[2] = nox_mapgenTruncQwordLow_503B30(fixed_at->field_0);
	corners[3] = nox_mapgenTruncQwordLow_503B30(fixed_at->field_4);

	corner.field_0 = width + at->field_0;
	corner.field_4 = height + at->field_4;
	nox_xxx_mapGenFixCoords_4D3D90(&corner, &fixed);
	corners[4] = nox_mapgenTruncQwordLow_503B30(fixed.field_0);
	corners[5] = nox_mapgenTruncQwordLow_503B30(fixed.field_4);

	corner.field_4 = at->field_4;
	nox_xxx_mapGenFixCoords_4D3D90(&corner, &fixed);
	corners[0] = nox_mapgenTruncQwordLow_503B30(fixed.field_0);
	corners[1] = nox_mapgenTruncQwordLow_503B30(fixed.field_4);

	corner.field_0 = at->field_0;
	corner.field_4 = height + at->field_4;
	nox_xxx_mapGenFixCoords_4D3D90(&corner, &fixed);
	corners[6] = nox_mapgenTruncQwordLow_503B30(fixed.field_0);
	corners[7] = nox_mapgenTruncQwordLow_503B30(fixed.field_4);

	sub_4D3C80((uint32_t*)corners);
	sub_428170(corners, bounds);
}

static int32_t nox_mapgenPlaceOffset_503B30(float fixed_coordinate, int32_t map_origin) {
	return nox_mapgenTruncQwordLow_503B30((double)fixed_coordinate - (double)map_origin);
}

static float nox_mapgenWallSpan_503B30(uint32_t wall_count) {
	uint32_t low = wall_count * 23u;
	int32_t wrapped;
	memcpy(&wrapped, &low, sizeof(wrapped));
	return (float)wrapped;
}

static int nox_mapgenFixCoordsNative_503B30(float x, float y, float* fixed_x, float* fixed_y) {
	float2 at = {x, y};
	float2 fixed;
	if (!nox_xxx_mapGenFixCoords_4D3D90(&at, &fixed)) {
		return 0;
	}
	*fixed_x = fixed.field_0;
	*fixed_y = fixed.field_4;
	return 1;
}

static int nox_mapgenPlaceGeometryNative_503B30(float x, float y, float fixed_x, float fixed_y,
	const void* record, int32_t map_origin_x, int32_t map_origin_y, int32_t bounds[4], int32_t delta[2]) {
	if (!record || !bounds || !delta) {
		return 0;
	}
	float2 at = {x, y};
	float2 fixed = {fixed_x, fixed_y};
	int32_t corners[8];
	int4 native_bounds;
	nox_mapgenBuildPlaceBounds_503B30(&at, &fixed, record, corners, &native_bounds);
	memcpy(bounds, &native_bounds, sizeof(native_bounds));
	delta[0] = nox_mapgenPlaceOffset_503B30(fixed_x, map_origin_x);
	delta[1] = nox_mapgenPlaceOffset_503B30(fixed_y, map_origin_y);
	return 1;
}

static int sub_503B30(float2* a1) {
	if (!a1) {
		return 0;
	}
	return nox_mapgenPlaceNative_503B30(a1->field_0, a1->field_4);
}

static int sub_503EC0(nox_object_t* object, float2* out) {
	float2 a1a; // [esp+0h] [ebp-18h]
	float2 v4;  // [esp+8h] [ebp-10h]
	float2 a2a; // [esp+10h] [ebp-8h]

	if (dword_5d4594_1599480 != dword_5d4594_3835396 || *(int*)&dword_5d4594_1599480 == -1 ||
		dword_5d4594_1599476 == 1) {
		return 0;
	}
	a1a.field_0 = (double)*getMemIntPtr(0x5D4594, 1599508);
	a1a.field_4 = (double)*getMemIntPtr(0x5D4594, 1599512);
	sub_4D3E30(&a1a, &a2a);
	// GAME.EXE passes the object's PE32 +56 position pair directly to
	// 004D3E30, which may clamp both live coordinates in place. Address the
	// same pair through the native object layout instead of truncating the
	// object pointer or retaining the fixed PE32 offset on 64-bit hosts.
	sub_4D3E30((float2*)&object->x, &v4);
	out->field_0 = v4.field_0 - a2a.field_0;
	out->field_4 = v4.field_4 - a2a.field_4;
	return 1;
}

#undef sub_5029A0
#undef sub_5029F0
#undef sub_502A20
#undef sub_502A30
#undef sub_502A50
#undef sub_502A90
#undef sub_502AB0
#undef sub_502AF0
#undef nox_mapgen_read_exact_502B10
#undef sub_502B10
#undef sub_502D70
#undef sub_502DA0
#undef sub_502DF0
#undef sub_502E10
#undef sub_502E50
#undef sub_502E70
#undef sub_502EA0
#undef sub_502ED0
#undef sub_503140
#undef sub_503230
#undef nox_mapgen_read_extract_prefix_5034B0
#undef nox_mapgen_copy_exact_5034B0
#undef sub_5034B0
#undef sub_5036D0
#undef nox_xxx_mapgenSaveMap_503830
#undef nox_mapgenClearPendingScriptIDs_503B30
#undef nox_mapgenTruncQwordLow_503B30
#undef nox_mapgenBuildPlaceBounds_503B30
#undef nox_mapgenPlaceOffset_503B30
#undef nox_mapgenWallSpan_503B30
#undef nox_mapgenFixCoordsNative_503B30
#undef nox_mapgenPlaceGeometryNative_503B30
#undef sub_503B30
#undef sub_503EC0
#endif
