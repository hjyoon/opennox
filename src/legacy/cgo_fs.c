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
