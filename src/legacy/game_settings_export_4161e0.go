package legacy

/*
#include <stdint.h>
*/
import "C"
import "unsafe"

//export sub_4161E0
func sub_4161E0() C.int { return C.int(sub4161E0Native()) }

//export sub_4164F0
func sub_4164F0() { sub4164F0Native() }

//export sub_416580
func sub_416580() C.int { return C.int(sub416580Native()) }

//export sub_4165B0
func sub_4165B0() *C.char { return (*C.char)(sub4165B0Native()) }

//export sub_4165D0
func sub_4165D0(index C.int) *C.char { return (*C.char)(sub4165D0Native(int32(index))) }

//export sub_4165F0
func sub_4165F0(from, to C.int) C.int { return C.int(sub4165F0Native(int32(from), int32(to))) }

//export sub_416630
func sub_416630() *C.char { return (*C.char)(sub416630Native()) }

//export sub_416640
func sub_416640() unsafe.Pointer { return sub416640Native() }

//export sub_416650
func sub_416650() C.int { return C.int(sub416650Native()) }

//export sub_416690
func sub_416690() { sub416690Native() }

//export sub_4169C0
func sub_4169C0() C.int { return C.int(sub4169C0Native()) }

//export sub_4169F0
func sub_4169F0() *C.char { return (*C.char)(sub4169F0Native()) }
