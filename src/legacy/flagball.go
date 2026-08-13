package legacy

import "C"

var (
	Sub_4E8290 func(state uint8, netCode uint16) int32
)

//export sub_4E8290
func sub_4E8290(state uint8, netCode uint16) int32 {
	return Sub_4E8290(state, netCode)
}
