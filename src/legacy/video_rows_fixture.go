package legacy

/*
#include <stdint.h>
extern uint8_t** nox_pixbuffer_rows_3798784;
*/
import "C"
import "unsafe"

// Declaration-only access to the real common C row-table global. Server
// builds also allocate/free these rows, so ownership tests need this alias.
var VideoModeRowsC = (***uint16)(unsafe.Pointer(&C.nox_pixbuffer_rows_3798784))
