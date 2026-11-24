package legacy

import "C"

var (
	Sub_57C490_2 func(key string)
)

//export sub_57C490_2
func sub_57C490_2(a1 *C.char) {
	Sub_57C490_2(GoString(a1))
}
