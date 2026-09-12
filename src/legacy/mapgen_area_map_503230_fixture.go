package legacy

/*
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "GAME4.h"
#include "memmap.h"

extern void* dword_5d4594_1599588;
extern void* dword_5d4594_1599592;
extern char* dword_5d4594_1599576;
extern uint32_t dword_5d4594_1599596;
extern FILE* nox_file_8;

typedef struct nox_test_mapgen_rename_503230_result {
	int result;
	uint32_t count;
	int file_closed;
	uintptr_t records_address;
	char first_name[64];
	char second_name[64];
	char third_name[64];
} nox_test_mapgen_rename_503230_result;

static nox_test_mapgen_rename_503230_result nox_test_mapgen_rename_503230(
	const char* source, const char* directory, const char* old_name, const char* new_name) {
	nox_test_mapgen_rename_503230_result out = {0};
	char first[0x800] = {0};
	char second[0x800] = {0};
	char records[0x26000] = {0};
	if (!source || !directory || !old_name || !new_name || strlen(source) >= sizeof(first) ||
		strlen(directory) >= sizeof(first)) {
		out.result = -1;
		return out;
	}
	char* directory_buffer = getMemAt(0x973F18, 42152);
	char old_directory[0x800];
	memcpy(old_directory, directory_buffer, sizeof(old_directory));
	void* old_first = dword_5d4594_1599588;
	void* old_second = dword_5d4594_1599592;
	char* old_records = dword_5d4594_1599576;
	uint32_t old_count = dword_5d4594_1599596;
	FILE* old_file = nox_file_8;
	strcpy(first, source);
	strcpy(directory_buffer, directory);
	dword_5d4594_1599588 = first;
	dword_5d4594_1599592 = second;
	dword_5d4594_1599576 = records;
	dword_5d4594_1599596 = 0;
	nox_file_8 = NULL;

	out.records_address = (uintptr_t)records;
	out.result = sub_503230((char*)old_name, (char*)new_name);
	out.count = dword_5d4594_1599596;
	out.file_closed = nox_file_8 == NULL;
	if (out.count > 0) {
		memcpy(out.first_name, records, sizeof(out.first_name));
	}
	if (out.count > 1) {
		memcpy(out.second_name, records + 76, sizeof(out.second_name));
	}
	if (out.count > 2) {
		memcpy(out.third_name, records + 152, sizeof(out.third_name));
	}

	dword_5d4594_1599588 = old_first;
	dword_5d4594_1599592 = old_second;
	dword_5d4594_1599576 = old_records;
	dword_5d4594_1599596 = old_count;
	nox_file_8 = old_file;
	memcpy(directory_buffer, old_directory, sizeof(old_directory));
	return out;
}
*/
import "C"

import "unsafe"

type mapgenRenameViaCResult503230 struct {
	result         int
	count          uint32
	fileClosed     bool
	recordsAddress uintptr
	firstName      string
	secondName     string
	thirdName      string
}

func mapgenRenameViaC503230(source, directory, oldName, newName string) mapgenRenameViaCResult503230 {
	csource := C.CString(source)
	defer C.free(unsafe.Pointer(csource))
	cdirectory := C.CString(directory)
	defer C.free(unsafe.Pointer(cdirectory))
	coldName := C.CString(oldName)
	defer C.free(unsafe.Pointer(coldName))
	cnewName := C.CString(newName)
	defer C.free(unsafe.Pointer(cnewName))
	v := C.nox_test_mapgen_rename_503230(csource, cdirectory, coldName, cnewName)
	return mapgenRenameViaCResult503230{
		result:         int(v.result),
		count:          uint32(v.count),
		fileClosed:     v.file_closed != 0,
		recordsAddress: uintptr(v.records_address),
		firstName:      C.GoString(&v.first_name[0]),
		secondName:     C.GoString(&v.second_name[0]),
		thirdName:      C.GoString(&v.third_name[0]),
	}
}
