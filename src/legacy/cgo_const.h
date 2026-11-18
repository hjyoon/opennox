#ifndef NOX_CGO_CONST_H
#define NOX_CGO_CONST_H

#include <stdint.h>

// Pointee aliases let cgo exports retain the public C const qualifiers.
typedef const char nox_cgo_const_char;
typedef const void nox_cgo_const_void;
typedef const uint32_t nox_cgo_const_uint32;

#endif
