package legacy

/*
#include <stdlib.h>
#include <string.h>

#include "GAME4.h"
#include "memmap.h"

extern void* dword_5d4594_1599588;

static int nox_test_mapgen_prepare_area_map_503140(const char* source, const char* directory) {
	char* directory_buffer = getMemAt(0x973F18, 42152);
	char old_directory[0x800];
	char* source_buffer = calloc(1, 0x800);
	if (!source_buffer) {
		return -1;
	}
	memcpy(old_directory, directory_buffer, sizeof(old_directory));
	void* old_source = dword_5d4594_1599588;
	strncpy(source_buffer, source, 0x7ff);
	memset(directory_buffer, 0, 0x800);
	strncpy(directory_buffer, directory, 0x7ff);
	dword_5d4594_1599588 = source_buffer;
	int result = sub_503140();
	dword_5d4594_1599588 = old_source;
	memcpy(directory_buffer, old_directory, sizeof(old_directory));
	free(source_buffer);
	return result;
}
*/
import "C"

import "unsafe"

func mapgenPrepareAreaMapViaC503140(source, directory string) int {
	csource := C.CString(source)
	defer C.free(unsafe.Pointer(csource))
	cdirectory := C.CString(directory)
	defer C.free(unsafe.Pointer(cdirectory))
	return int(C.nox_test_mapgen_prepare_area_map_503140(csource, cdirectory))
}
