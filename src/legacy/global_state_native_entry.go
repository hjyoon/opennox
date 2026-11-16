package legacy

/*
#include "GAME1_2.h"
#include "GAME2_1.h"
#include "GAME2_2.h"
#include "GAME2_3.h"
#include "GAME3_1.h"
#include "GAME4_1.h"
*/
import "C"

//export nox_xxx_clientIsObserver_4372E0
func nox_xxx_clientIsObserver_4372E0() C.int { return C.int(globalStateNativeValue(0)) }

//export sub_467590
func sub_467590() C.int { return C.int(globalStateNativeValue(1)) }

//export sub_4678B0
func sub_4678B0() C.int { return C.int(globalStateNativeValue(2)) }

//export sub_460D40
func sub_460D40() C.int { return C.int(globalStateNativeValue(3)) }

//export sub_4678C0
func sub_4678C0() C.int { return C.int(globalStateNativeValue(4)) }

//export sub_4BFD30
func sub_4BFD30() C.int { return C.int(globalStateNativeValue(5)) }

//export sub_4C12C0
func sub_4C12C0() C.int { return C.int(globalStateNativeValue(6)) }

//export sub_479870
func sub_479870() C.int { return C.int(globalStateNativeValue(7)) }

//export sub_49C810
func sub_49C810() C.int { return C.int(globalStateNativeValue(8)) }

//export sub_49CB40
func sub_49CB40() C.int { return C.int(globalStateNativeValue(9)) }

//export sub_4C5050
func sub_4C5050() { globalStateNativeMutate(10) }

//export sub_481410
func sub_481410() { globalStateNativeMutate(11) }

//export sub_51D0E0
func sub_51D0E0() { globalStateNativeMutate(12) }

//export nox_client_mapZoomIn_4724E0
func nox_client_mapZoomIn_4724E0() { globalStateNativeMutate(13) }

//export nox_client_mapZoomOut_472500
func nox_client_mapZoomOut_472500() { globalStateNativeMutate(14) }

//export sub_470CC0
func sub_470CC0() C.int { return C.int(globalStateNativeValue(15)) }

//export sub_470CD0
func sub_470CD0() C.int { return C.int(globalStateNativeValue(16)) }
