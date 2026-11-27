//go:build !386 && !amd64 && !arm && !arm64

#define NOX_CCALL_FALLBACK
#include "ccall_abi.h"
