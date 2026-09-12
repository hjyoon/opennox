package legacy

/*
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "GAME4.h"
#include "common/fs/nox_fs.h"

extern char* dword_5d4594_1599576;
extern uint32_t dword_5d4594_1599596;
extern uint32_t dword_5d4594_3835396;
extern FILE* nox_file_8;

typedef struct nox_test_mapgen_stream_access_502D70_result {
	uintptr_t records_address;
	uintptr_t file_address;
	int invalid_index;
	int invalid_coordinate;
	int closed_seek;
	int opened;
	int seek_by_index;
	int seek_by_name;
	int absent_name;
	int nil_name;
	int reopen_rewinds;
	int closed;
	int missing_file;
	int idempotent_close;
	double x;
	double y;
} nox_test_mapgen_stream_access_502D70_result;

static nox_test_mapgen_stream_access_502D70_result nox_test_mapgen_stream_access_502D70(const char* path, const char* missing) {
	nox_test_mapgen_stream_access_502D70_result out = {0};
	char records[2 * 76] = {0};
	uint32_t offset0 = 7, offset1 = 11;
	float x = 1.5f, y = -2.5f;
	memcpy(records, "Alpha", 6);
	memcpy(records + 76, "Beta", 5);
	memcpy(records + 64, &x, sizeof(x));
	memcpy(records + 76 + 68, &y, sizeof(y));
	memcpy(records + 72, &offset0, sizeof(offset0));
	memcpy(records + 76 + 72, &offset1, sizeof(offset1));

	char* old_records = dword_5d4594_1599576;
	uint32_t old_count = dword_5d4594_1599596;
	uint32_t old_selected = dword_5d4594_3835396;
	FILE* old_file = nox_file_8;
	dword_5d4594_1599576 = records;
	dword_5d4594_1599596 = 2;
	dword_5d4594_3835396 = 733;
	nox_file_8 = NULL;
	out.records_address = (uintptr_t)records;

	out.invalid_index = sub_502D70(-1) == 0 && sub_502D70(2) == 0 && dword_5d4594_3835396 == 733;
	out.x = sub_502E70(0);
	out.y = sub_502EA0(1);
	out.invalid_coordinate = sub_502E70(-1) == -1.0 && sub_502E70(2) == -1.0 &&
		sub_502EA0(-1) == -1.0 && sub_502EA0(2) == -1.0;
	out.closed_seek = sub_502E10(0) == NULL && sub_502E50((char*)"Alpha") == NULL;

	out.opened = sub_502DA0((char*)path) != NULL && nox_file_8 != NULL;
	if (out.opened) {
		FILE* opened_file = nox_file_8;
		out.file_address = (uintptr_t)opened_file;
		out.seek_by_index = nox_fs_ftell(opened_file) == 0 &&
			sub_502E10(1) == opened_file && nox_fs_ftell(opened_file) == 11;
		out.seek_by_name = sub_502E50((char*)"aLpHa") == opened_file && nox_fs_ftell(opened_file) == 7;
		out.absent_name = sub_502E50((char*)"Absent") == NULL && nox_fs_ftell(opened_file) == 7;
		out.nil_name = sub_502E50(NULL) == NULL && nox_fs_ftell(opened_file) == 7;
		out.reopen_rewinds = sub_502DA0((char*)missing) != NULL &&
			nox_file_8 == opened_file && nox_fs_ftell(opened_file) == 0;
		sub_502DF0();
		out.closed = nox_file_8 == NULL && sub_502E10(0) == NULL;
	}
	sub_502DF0();
	out.idempotent_close = nox_file_8 == NULL;
	out.missing_file = sub_502DA0((char*)missing) == NULL && nox_file_8 == NULL;

	dword_5d4594_1599576 = old_records;
	dword_5d4594_1599596 = old_count;
	dword_5d4594_3835396 = old_selected;
	nox_file_8 = old_file;
	return out;
}
*/
import "C"
import "unsafe"

type mapgenStreamAccessResult502D70 struct {
	recordsAddress    uintptr
	fileAddress       uintptr
	invalidIndex      bool
	invalidCoordinate bool
	closedSeek        bool
	opened            bool
	seekByIndex       bool
	seekByName        bool
	absentName        bool
	nilName           bool
	reopenRewinds     bool
	closed            bool
	missingFile       bool
	idempotentClose   bool
	x                 float64
	y                 float64
}

func mapgenStreamAccess502D70(path, missing string) mapgenStreamAccessResult502D70 {
	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))
	cmissing := C.CString(missing)
	defer C.free(unsafe.Pointer(cmissing))
	v := C.nox_test_mapgen_stream_access_502D70(cpath, cmissing)
	return mapgenStreamAccessResult502D70{
		recordsAddress:    uintptr(v.records_address),
		fileAddress:       uintptr(v.file_address),
		invalidIndex:      v.invalid_index != 0,
		invalidCoordinate: v.invalid_coordinate != 0,
		closedSeek:        v.closed_seek != 0,
		opened:            v.opened != 0,
		seekByIndex:       v.seek_by_index != 0,
		seekByName:        v.seek_by_name != 0,
		absentName:        v.absent_name != 0,
		nilName:           v.nil_name != 0,
		reopenRewinds:     v.reopen_rewinds != 0,
		closed:            v.closed != 0,
		missingFile:       v.missing_file != 0,
		idempotentClose:   v.idempotent_close != 0,
		x:                 float64(v.x),
		y:                 float64(v.y),
	}
}
