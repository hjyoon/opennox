#ifndef NOX_STRING_FORMAT_ABI_H
#define NOX_STRING_FORMAT_ABI_H

#include "noxstring.h"

typedef struct nox_format_args nox_format_args;
typedef const char nox_format_byte;
typedef const wchar2_t nox_format_word;

int nox_format_narrow_go(char* out, size_t capacity, nox_format_byte* format, nox_format_args* args);
int nox_format_wide_go(wchar2_t* out, size_t capacity, nox_format_word* format, nox_format_args* args);

int nox_format_arg_int(nox_format_args* args);
unsigned int nox_format_arg_uint(nox_format_args* args);
const char* nox_format_arg_narrow(nox_format_args* args);
const wchar2_t* nox_format_arg_wide(nox_format_args* args);
size_t nox_format_arg_float(nox_format_args* args, char* out, int precision);
void nox_format_abort(int code);

#endif
