package legacy

/*
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "GAME4.h"

extern void* dword_5d4594_1599588;
extern void* dword_5d4594_1599592;
extern char* dword_5d4594_1599576;
extern uint32_t dword_5d4594_1599596;
extern FILE* nox_file_8;

typedef struct nox_test_mapgen_record_502B10 {
	char name[64];
	uint32_t x_bits;
	uint32_t y_bits;
	uint32_t file_offset;
} nox_test_mapgen_record_502B10;

typedef struct nox_test_mapgen_reader_502B10_result {
	int result;
	uint32_t count;
	int file_closed;
	uintptr_t records_address;
	nox_test_mapgen_record_502B10 first;
	nox_test_mapgen_record_502B10 second;
	nox_test_mapgen_record_502B10 last;
} nox_test_mapgen_reader_502B10_result;

_Static_assert(sizeof(nox_test_mapgen_record_502B10) == 76, "fixture record layout must match the oracle");

static nox_test_mapgen_reader_502B10_result nox_test_mapgen_reader_502B10(const char* path) {
	nox_test_mapgen_reader_502B10_result out = {0};
	char first[2048] = {0};
	char second[2048] = {0};
	char records[2048 * 76] = {0};
	if (!path || strlen(path) >= sizeof(first)) {
		return out;
	}
	strcpy(first, path);

	void* old_first = dword_5d4594_1599588;
	void* old_second = dword_5d4594_1599592;
	char* old_records = dword_5d4594_1599576;
	uint32_t old_count = dword_5d4594_1599596;
	FILE* old_file = nox_file_8;
	dword_5d4594_1599588 = first;
	dword_5d4594_1599592 = second;
	dword_5d4594_1599576 = records;
	nox_file_8 = NULL;

	out.records_address = (uintptr_t)records;
	out.result = sub_502B10();
	out.count = dword_5d4594_1599596;
	out.file_closed = nox_file_8 == NULL;
	if (out.count > 0) {
		memcpy(&out.first, records, sizeof(out.first));
		memcpy(&out.last, records + 76 * (out.count - 1), sizeof(out.last));
	}
	if (out.count > 1) {
		memcpy(&out.second, records + 76, sizeof(out.second));
	}

	dword_5d4594_1599588 = old_first;
	dword_5d4594_1599592 = old_second;
	dword_5d4594_1599576 = old_records;
	dword_5d4594_1599596 = old_count;
	nox_file_8 = old_file;
	return out;
}

static int nox_test_mapgen_reader_502B10_lazy_alloc(uintptr_t* first, uintptr_t* second, uintptr_t* records) {
	void* old_first = dword_5d4594_1599588;
	void* old_second = dword_5d4594_1599592;
	char* old_records = dword_5d4594_1599576;
	uint32_t old_count = dword_5d4594_1599596;
	dword_5d4594_1599588 = NULL;
	dword_5d4594_1599592 = NULL;
	dword_5d4594_1599576 = NULL;
	dword_5d4594_1599596 = UINT32_MAX;
	int result = sub_502B10();
	*first = (uintptr_t)dword_5d4594_1599588;
	*second = (uintptr_t)dword_5d4594_1599592;
	*records = (uintptr_t)dword_5d4594_1599576;
	int ok = result == 0 && dword_5d4594_1599596 == 0 &&
		*first != 0 && *second != 0 && *records != 0;
	free(dword_5d4594_1599588);
	free(dword_5d4594_1599592);
	free(dword_5d4594_1599576);
	dword_5d4594_1599588 = old_first;
	dword_5d4594_1599592 = old_second;
	dword_5d4594_1599576 = old_records;
	dword_5d4594_1599596 = old_count;
	return ok;
}
*/
import "C"
import "unsafe"

type mapgenReaderRecord502B10 struct {
	name       string
	xBits      uint32
	yBits      uint32
	fileOffset uint32
}

type mapgenReaderResult502B10 struct {
	result         int
	count          uint32
	fileClosed     bool
	recordsAddress uintptr
	first          mapgenReaderRecord502B10
	second         mapgenReaderRecord502B10
	last           mapgenReaderRecord502B10
}

func mapgenReaderRecordFromC502B10(v C.nox_test_mapgen_record_502B10) mapgenReaderRecord502B10 {
	return mapgenReaderRecord502B10{
		name:       C.GoString(&v.name[0]),
		xBits:      uint32(v.x_bits),
		yBits:      uint32(v.y_bits),
		fileOffset: uint32(v.file_offset),
	}
}

func mapgenReadFile502B10(path string) mapgenReaderResult502B10 {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))
	v := C.nox_test_mapgen_reader_502B10(cpath)
	return mapgenReaderResult502B10{
		result:         int(v.result),
		count:          uint32(v.count),
		fileClosed:     v.file_closed != 0,
		recordsAddress: uintptr(v.records_address),
		first:          mapgenReaderRecordFromC502B10(v.first),
		second:         mapgenReaderRecordFromC502B10(v.second),
		last:           mapgenReaderRecordFromC502B10(v.last),
	}
}

func mapgenReaderLazyAlloc502B10() (uintptr, uintptr, uintptr, bool) {
	var first, second, records C.uintptr_t
	ok := C.nox_test_mapgen_reader_502B10_lazy_alloc(&first, &second, &records) != 0
	return uintptr(first), uintptr(second), uintptr(records), ok
}
