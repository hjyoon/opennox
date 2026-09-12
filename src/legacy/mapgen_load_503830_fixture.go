package legacy

/*
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "GAME4.h"
#include "common__crypt.h"
#include "memmap.h"

extern void* dword_5d4594_1599588;
extern char* dword_5d4594_1599576;
extern uint32_t dword_5d4594_1599596;
extern uint32_t dword_5d4594_1599644;
extern uint32_t dword_5d4594_1599480;
extern uint32_t dword_5d4594_1599476;
extern uint32_t dword_5d4594_3835396;
extern FILE* nox_file_8;

static uintptr_t nox_test_mapgen_xfer_obj_503830;
static uintptr_t nox_test_mapgen_xfer_bounds_503830;
static unsigned char nox_test_mapgen_xfer_payload_503830[3];

static void nox_test_mapgen_xfer_reset_503830(void) {
	nox_test_mapgen_xfer_obj_503830 = 0;
	nox_test_mapgen_xfer_bounds_503830 = 0;
	memset(nox_test_mapgen_xfer_payload_503830, 0, sizeof(nox_test_mapgen_xfer_payload_503830));
}

static int nox_test_mapgen_xfer_503830(void* obj, void* bounds) {
	nox_test_mapgen_xfer_obj_503830 = (uintptr_t)obj;
	nox_test_mapgen_xfer_bounds_503830 = (uintptr_t)bounds;
	return nox_xxx_fileReadWrite_426AC0_file3_fread(
		nox_test_mapgen_xfer_payload_503830,
		sizeof(nox_test_mapgen_xfer_payload_503830)
	);
}

static void* nox_test_mapgen_xfer_func_503830(void) {
	return (void*)nox_test_mapgen_xfer_503830;
}

static void nox_test_mapgen_xfer_snapshot_503830(uintptr_t* obj, uintptr_t* bounds, unsigned char* payload) {
	*obj = nox_test_mapgen_xfer_obj_503830;
	*bounds = nox_test_mapgen_xfer_bounds_503830;
	memcpy(payload, nox_test_mapgen_xfer_payload_503830, sizeof(nox_test_mapgen_xfer_payload_503830));
}

typedef struct nox_test_mapgen_load_503830_result {
	int indexed;
	int loaded;
	int file_closed;
	uintptr_t records_address;
	uint32_t selected;
	uint32_t loaded_index;
	uint32_t walls[2];
	uint32_t bounds[8];
} nox_test_mapgen_load_503830_result;

static nox_test_mapgen_load_503830_result nox_test_mapgen_load_503830(const char* source) {
	nox_test_mapgen_load_503830_result out = {0};
	char path[0x800] = {0};
	char records[0x26000] = {0};
	if (!source || strlen(source) >= sizeof(path)) {
		out.indexed = -1;
		return out;
	}
	void* old_source = dword_5d4594_1599588;
	char* old_records = dword_5d4594_1599576;
	uint32_t old_count = dword_5d4594_1599596;
	uint32_t old_extra = dword_5d4594_1599644;
	uint32_t old_loaded = dword_5d4594_1599480;
	uint32_t old_dirty = dword_5d4594_1599476;
	uint32_t old_index = dword_5d4594_3835396;
	uint32_t* selected = getMemU32Ptr(0x5D4594, 1599572);
	uint32_t old_selected = *selected;
	uint32_t old_walls[2], old_bounds[8];
	memcpy(old_walls, getMemAt(0x5D4594, 739980), sizeof(old_walls));
	memcpy(old_bounds, getMemAt(0x5D4594, 1599500), sizeof(old_bounds));
	FILE* old_file = nox_file_8;
	strcpy(path, source);
	dword_5d4594_1599588 = path;
	dword_5d4594_1599576 = records;
	dword_5d4594_1599596 = 0;
	nox_file_8 = NULL;
	out.records_address = (uintptr_t)records;
	out.indexed = sub_502B10();
	if (dword_5d4594_1599596 != 0) {
		out.loaded = nox_xxx_mapgenSaveMap_503830(0);
	}
	out.file_closed = nox_file_8 == NULL;
	out.selected = *selected;
	out.loaded_index = dword_5d4594_1599480;
	memcpy(out.walls, getMemAt(0x5D4594, 739980), sizeof(out.walls));
	memcpy(out.bounds, getMemAt(0x5D4594, 1599500), sizeof(out.bounds));
	sub_502DF0();
	dword_5d4594_1599588 = old_source;
	dword_5d4594_1599576 = old_records;
	dword_5d4594_1599596 = old_count;
	dword_5d4594_1599644 = old_extra;
	dword_5d4594_1599480 = old_loaded;
	dword_5d4594_1599476 = old_dirty;
	dword_5d4594_3835396 = old_index;
	*selected = old_selected;
	memcpy(getMemAt(0x5D4594, 739980), old_walls, sizeof(old_walls));
	memcpy(getMemAt(0x5D4594, 1599500), old_bounds, sizeof(old_bounds));
	nox_file_8 = old_file;
	return out;
}
*/
import "C"

import "unsafe"

type mapgenLoadViaCResult503830 struct {
	indexed        int
	loaded         bool
	fileClosed     bool
	recordsAddress uintptr
	selected       uint32
	loadedIndex    uint32
	walls          [2]uint32
	bounds         [8]uint32
}

func mapgenLoadViaC503830(source string) mapgenLoadViaCResult503830 {
	csource := C.CString(source)
	defer C.free(unsafe.Pointer(csource))
	v := C.nox_test_mapgen_load_503830(csource)
	return mapgenLoadViaCResult503830{
		indexed:        int(v.indexed),
		loaded:         v.loaded != 0,
		fileClosed:     v.file_closed != 0,
		recordsAddress: uintptr(v.records_address),
		selected:       uint32(v.selected),
		loadedIndex:    uint32(v.loaded_index),
		walls:          [2]uint32{uint32(v.walls[0]), uint32(v.walls[1])},
		bounds: [8]uint32{
			uint32(v.bounds[0]), uint32(v.bounds[1]), uint32(v.bounds[2]), uint32(v.bounds[3]),
			uint32(v.bounds[4]), uint32(v.bounds[5]), uint32(v.bounds[6]), uint32(v.bounds[7]),
		},
	}
}

func mapgenTestXferFunc503830() unsafe.Pointer { return C.nox_test_mapgen_xfer_func_503830() }

func mapgenTestXferReset503830() { C.nox_test_mapgen_xfer_reset_503830() }

func mapgenTestXferSnapshot503830() (uintptr, uintptr, [3]byte) {
	var obj, bounds C.uintptr_t
	var payload [3]byte
	C.nox_test_mapgen_xfer_snapshot_503830(&obj, &bounds, (*C.uchar)(unsafe.Pointer(&payload[0])))
	return uintptr(obj), uintptr(bounds), payload
}
