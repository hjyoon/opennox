package legacy

/*
#include <stdint.h>
*/
import "C"

//export nox_xxx_gameLoopMemDump_413E30
func nox_xxx_gameLoopMemDump_413E30() { memDumpNative413E30() }

//export nox_xxx_initSinCosTables_414C90
func nox_xxx_initSinCosTables_414C90() C.char { return C.char(trigInitNative414C90()) }
