#include <stdint.h>
#include "compat.h"
#include "common/fs/nox_fs.h"
#include "noxstring_format_abi.h"

int nox_fs_vfprintf_go(FILE* file, nox_format_byte* format, nox_format_args* args);

int nox_fs_fprintf(FILE* file, const char* format, ...) {
	nox_format_args args;
	va_start(args.list, format);
	int result = nox_fs_vfprintf_go(file, format, &args);
	va_end(args.list);
	return result;
}

#ifdef _WIN32
uintptr_t nox_win_find_first_file_word(const char* path, WIN32_FIND_DATAA* entry) {
	return (uintptr_t)FindFirstFileA(path, entry);
}

int nox_win_find_next_file_word(uintptr_t handle, WIN32_FIND_DATAA* entry) {
	return FindNextFileA((HANDLE)handle, entry);
}

int nox_win_find_close_word(uintptr_t handle) {
	return FindClose((HANDLE)handle);
}
#else
extern uintptr_t nox_compat_find_first_file_word(const char*, WIN32_FIND_DATAA*);
extern int nox_compat_find_next_file_word(uintptr_t, WIN32_FIND_DATAA*);
extern int nox_compat_find_close_word(uintptr_t);

HANDLE compatFindFirstFileA(const char* path, WIN32_FIND_DATAA* entry) {
	return (HANDLE)nox_compat_find_first_file_word(path, entry);
}

int compatFindNextFileA(HANDLE handle, WIN32_FIND_DATAA* entry) {
	return nox_compat_find_next_file_word((uintptr_t)handle, entry);
}

int compatFindClose(HANDLE handle) {
	return nox_compat_find_close_word((uintptr_t)handle);
}
#endif
