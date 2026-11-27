//go:build !386 && !amd64 && !arm && !arm64

#define NOX_NATIVE_CALL_FALLBACK
#include "native_callback_calls.h"
#include "native_death_calls.h"
