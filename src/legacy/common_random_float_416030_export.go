package legacy

/*
#include "common__random.h"
*/
import "C"

// Exercise the original C float/float/double ABI, including the cgo callback,
// rather than a Go-only copy of the exported function.
func commonRandomFloatCEntry416030(min, max float32) float64 {
	return float64(C.nox_common_randomFloat_416030(C.float(min), C.float(max)))
}
