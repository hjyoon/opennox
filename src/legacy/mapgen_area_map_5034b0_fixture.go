package legacy

/*
#include "world_native_storage.h"
#define dword_5d4594_1599576 (nox_archive_mapgen_storage()->archive_records)
#define dword_5d4594_1599596 (nox_archive_mapgen_storage()->archive_record_count)
#include "native_fixture_alloc.h"
#include "internal/cfixture/game4_mapgen_storage_fixture.h"
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "GAME4.h"
#include "memmap.h"


typedef struct nox_test_mapgen_extract_5034B0_result {
	int indexed;
	int payload;
	int attachment;
	int file_closed;
	uintptr_t records_address;
} nox_test_mapgen_extract_5034B0_result;

static nox_test_mapgen_extract_5034B0_result nox_test_mapgen_extract_5034B0(
	const char* source, const char* directory, const char* name, const char* attachment_path) {
	nox_test_mapgen_extract_5034B0_result out = {0};
	char* first = nox_native_fixture_calloc_high(1, 0x800);
	char* second = nox_native_fixture_calloc_high(1, 0x800);
	char* records = nox_native_fixture_calloc_high(1, 0x26000);
	if (!first || !second || !records) abort();
	if (!source || !directory || !name || !attachment_path ||
		strlen(source) >= (0x800) || strlen(directory) >= (0x800)) {
		out.indexed = -1;
		nox_native_fixture_free_high(records);
		nox_native_fixture_free_high(second);
		nox_native_fixture_free_high(first);
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
	out.indexed = sub_502B10();
	out.payload = sub_5034B0((char*)name);
	out.attachment = sub_5036D0((char*)name, (char*)attachment_path);
	out.file_closed = nox_file_8 == NULL;
	sub_502DF0();
	dword_5d4594_1599588 = old_first;
	dword_5d4594_1599592 = old_second;
	dword_5d4594_1599576 = old_records;
	dword_5d4594_1599596 = old_count;
	nox_file_8 = old_file;
	memcpy(directory_buffer, old_directory, sizeof(old_directory));
	nox_native_fixture_free_high(records);
	nox_native_fixture_free_high(second);
	nox_native_fixture_free_high(first);
	return out;
}
*/
import "C"

import "unsafe"

type mapgenExtractViaCResult5034B0 struct {
	indexed        int
	payload        bool
	attachment     bool
	fileClosed     bool
	recordsAddress uintptr
}

func mapgenExtractViaC5034B0(source, directory, name, attachmentPath string) mapgenExtractViaCResult5034B0 {
	csource := CString(source)
	defer C.free(unsafe.Pointer(csource))
	cdirectory := CString(directory)
	defer C.free(unsafe.Pointer(cdirectory))
	cname := CString(name)
	defer C.free(unsafe.Pointer(cname))
	cattachment := CString(attachmentPath)
	defer C.free(unsafe.Pointer(cattachment))
	v := C.nox_test_mapgen_extract_5034B0(csource, cdirectory, cname, cattachment)
	return mapgenExtractViaCResult5034B0{
		indexed:        int(v.indexed),
		payload:        v.payload != 0,
		attachment:     v.attachment != 0,
		fileClosed:     v.file_closed != 0,
		recordsAddress: uintptr(v.records_address),
	}
}
