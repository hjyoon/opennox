package legacy

/*
#include <math.h>
*/
import "C"

// Retain the same host libm services as the former C entry. Control flow,
// conversion, table ownership and directed-rounding correction live in Go.
func trigSin414C90(v float64) float64  { return float64(C.sin(C.double(v))) }
func trigAcos414C90(v float64) float64 { return float64(C.acos(C.double(v))) }
