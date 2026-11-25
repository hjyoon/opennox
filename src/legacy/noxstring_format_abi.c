#include "noxstring_format_abi.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int nox_format_arg_int(nox_format_args* args) {
	return va_arg(args->list, int);
}

unsigned int nox_format_arg_uint(nox_format_args* args) {
	return va_arg(args->list, unsigned int);
}

const char* nox_format_arg_narrow(nox_format_args* args) {
	return va_arg(args->list, const char*);
}

const wchar2_t* nox_format_arg_wide(nox_format_args* args) {
	return va_arg(args->list, const wchar2_t*);
}

size_t nox_format_arg_float(nox_format_args* args, char* out, int precision) {
	char format[32];
	snprintf(format, sizeof(format), "%%.%df", precision);
	snprintf(out, 32, format, va_arg(args->list, double));
	return strlen(out);
}

void nox_format_abort(int code) {
	printf("Unhandled format character: '%c'", code);
	abort();
}

int nox_vsnprintf(char* out, size_t capacity, const char* format, va_list args) {
	nox_format_args state;
	va_copy(state.list, args);
	int length = nox_format_narrow_go(out, capacity, format, &state);
	va_end(state.list);
	return length;
}

int nox_vsnwprintf(wchar2_t* out, size_t capacity, const wchar2_t* format, va_list args) {
	nox_format_args state;
	va_copy(state.list, args);
	int length = nox_format_wide_go(out, capacity, format, &state);
	va_end(state.list);
	return length;
}

int nox_vsprintf(char* out, const char* format, va_list args) {
	return nox_vsnprintf(out, 0x7fffffff, format, args);
}

int nox_sprintf(char* out, const char* format, ...) {
	va_list args;
	va_start(args, format);
	int length = nox_vsprintf(out, format, args);
	va_end(args);
	return length;
}

int nox_vswprintf(wchar2_t* out, const wchar2_t* format, va_list args) {
	return nox_vsnwprintf(out, 0x3fffffff, format, args);
}

int nox_swprintf(wchar2_t* out, const wchar2_t* format, ...) {
	va_list args;
	va_start(args, format);
	int length = nox_vswprintf(out, format, args);
	va_end(args);
	return length;
}
