#include "noxstring_format_abi.h"

#include <stdio.h>
#include <string.h>

_Static_assert(sizeof(double) == sizeof(uint64_t), "double bit transport width");

#ifndef NOX_GO_FORMAT_WORDS
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

uint64_t nox_format_arg_double_bits(nox_format_args* args) {
	double value = va_arg(args->list, double);
	uint64_t bits;
	memcpy(&bits, &value, sizeof(bits));
	return bits;
}

void nox_format_host_narrow(nox_format_args* args, char* out, size_t capacity, const char* format) {
	vsnprintf(out, capacity, format, args->list);
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

int nox_vswprintf(wchar2_t* out, const wchar2_t* format, va_list args) {
	return nox_vsnwprintf(out, 0x3fffffff, format, args);
}
#endif

void nox_format_fixed_bits(uint64_t bits, char* out, size_t capacity, int precision) {
	double value;
	memcpy(&value, &bits, sizeof(value));
	snprintf(out, capacity, "%.*f", precision, value);
}

int nox_sprintf(char* out, const char* format, ...) {
	va_list args;
	va_start(args, format);
	int length = nox_vsprintf(out, format, args);
	va_end(args);
	return length;
}

int nox_swprintf(wchar2_t* out, const wchar2_t* format, ...) {
	va_list args;
	va_start(args, format);
	int length = nox_vswprintf(out, format, args);
	va_end(args);
	return length;
}
