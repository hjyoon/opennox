package legacy

import "C"

var (
	Sub_4706C0 func(a1 int)
)

//export sub_4706C0
func sub_4706C0(a1_cgo int32) { a1 := int(a1_cgo); Sub_4706C0(a1) }
