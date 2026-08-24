package legacy

import "C"
import "github.com/opennox/opennox/v1/client/gui"

//export nox_client_setCursorType_477610
func nox_client_setCursorType_477610(v_cgo int32) int32 {
	v := int(v_cgo)
	GetClient().Nox_client_setCursorType(gui.Cursor(v))
	return int32(v)
}

//export nox_xxx_cursorGetTypePrev_477630
func nox_xxx_cursorGetTypePrev_477630() int32 {
	return int32(int(GetClient().Cli().CursorPrev))
}
