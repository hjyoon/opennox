package legacy

/*
#include <stdint.h>
int nox_game_showConnection_4AA270_native(void);
int nox_connection_outStart_4AA450_native(void);
int nox_connection_outDone_4AA490_native(void);
*/
import "C"
import "unsafe"

var (
	NoxGameShowConnection4AA270 func() int
	ConnectionOutStart4AA450    func() int
	ConnectionOutDone4AA490     func() int
)

//export nox_game_showConnection_4AA270_native
func nox_game_showConnection_4AA270_native() C.int {
	if NoxGameShowConnection4AA270 == nil {
		return 0
	}
	return C.int(NoxGameShowConnection4AA270())
}

//export nox_connection_outStart_4AA450_native
func nox_connection_outStart_4AA450_native() C.int {
	if ConnectionOutStart4AA450 == nil {
		return 0
	}
	return C.int(ConnectionOutStart4AA450())
}

//export nox_connection_outDone_4AA490_native
func nox_connection_outDone_4AA490_native() C.int {
	if ConnectionOutDone4AA490 == nil {
		return 0
	}
	return C.int(ConnectionOutDone4AA490())
}

func GetNoxGameShowConnection4AA270() unsafe.Pointer {
	return unsafe.Pointer(C.nox_game_showConnection_4AA270_native)
}

func GetConnectionOutStart4AA450() unsafe.Pointer {
	return unsafe.Pointer(C.nox_connection_outStart_4AA450_native)
}

func GetConnectionOutDone4AA490() unsafe.Pointer {
	return unsafe.Pointer(C.nox_connection_outDone_4AA490_native)
}
