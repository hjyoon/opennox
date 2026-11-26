#ifndef NOX_STRING_FORMAT_ABI_H
#define NOX_STRING_FORMAT_ABI_H

#include "noxstring.h"

typedef struct nox_format_args {
	va_list list;
} nox_format_args;

#if defined(__i386__) || defined(__x86_64__) || defined(__arm__) || defined(__aarch64__)
#define NOX_GO_FORMAT_WORDS
#if defined(__x86_64__) && !defined(_WIN32)
_Static_assert(sizeof(va_list) == 24 && _Alignof(va_list) == 8, "SysV AMD64 va_list layout");
#elif defined(__aarch64__) && !defined(__APPLE__) && !defined(_WIN32)
_Static_assert(sizeof(va_list) == 32 && _Alignof(va_list) == 8, "AAPCS64 va_list layout");
#else
_Static_assert(sizeof(va_list) == sizeof(void*) &&
	_Alignof(va_list) == _Alignof(void*), "variadic pointer cursor layout");
#endif
#endif

typedef const char nox_format_byte;
typedef const wchar2_t nox_format_word;

int nox_format_narrow_go(char* out, size_t capacity, nox_format_byte* format, nox_format_args* args);
int nox_format_wide_go(wchar2_t* out, size_t capacity, nox_format_word* format, nox_format_args* args);
void nox_format_bounded_wide_go(wchar2_t* out, size_t capacity, nox_format_word* format, nox_format_args* args);

void nox_format_wide_number_abi(wchar2_t* out, size_t capacity, const wchar2_t* format, int value);
void nox_format_wide_strings_abi(wchar2_t* out, size_t capacity, const wchar2_t* format,
                               const wchar2_t* first, const wchar2_t* second);

#ifndef NOX_GO_FORMAT_WORDS
int nox_format_arg_int(nox_format_args* args);
unsigned int nox_format_arg_uint(nox_format_args* args);
const char* nox_format_arg_narrow(nox_format_args* args);
const wchar2_t* nox_format_arg_wide(nox_format_args* args);
#endif
void nox_format_arg_fixed(nox_format_args* args, char* out, size_t capacity, int precision);
void nox_format_host_narrow(nox_format_args* args, char* out, size_t capacity, const char* format);
void nox_format_abort(int code);

#endif
