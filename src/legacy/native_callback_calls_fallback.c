//go:build !386 && !amd64 && !arm && !arm64

#define NOX_NATIVE_CALLBACK_FALLBACK
#include "native_callback_calls.h"
