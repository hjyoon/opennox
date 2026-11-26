package legacy

/*
#include "client__gui__window.h"
#include "client__gui__guicon.h"
unsigned int nox_gui_xxx_check_446360();
*/
import "C"
import "github.com/opennox/libs/console"

var (
	GetConsole                     func() *console.Console
	Nox_gui_console_flagXxx_451410 func() int
	Nox_gui_console_Hide_4512B0    func() int
)

//export nox_gui_console_flagXxx_451410
func nox_gui_console_flagXxx_451410() int32 { return int32(Nox_gui_console_flagXxx_451410()) }

//export nox_gui_console_Print_450B90
func nox_gui_console_Print_450B90(cl C.uchar, cstr *wchar2_t) int32 {
	// since legacy code still calls it, we redirect into global printer instead of GUI printer
	GetConsole().Print(console.Color(cl), GoWString(cstr))
	return int32(1)
}

//export nox_gui_console_vprintf_450C00
func nox_gui_console_vprintf_450C00(cl C.uchar, format *C.nox_format_word, args *C.nox_format_args) C.int {
	return consolePrintFormatted(cl, format, formatArguments{native: args})
}

func consolePrintRed(format *wchar2_t, values ...formatValue) {
	consolePrintFormatted(C.NOX_CONSOLE_RED, format, formatArguments{values: values})
}

func consolePrintFormatted(cl C.uchar, format *wchar2_t, args formatArguments) C.int {
	var text [512]wchar2_t
	formatBoundedWide(text[:], format, args)
	return C.int(nox_gui_console_Print_450B90(cl, &text[0]))
}

//export nox_gui_console_PrintOrError_450C30
func nox_gui_console_PrintOrError_450C30(cl C.uchar, cstr *wchar2_t) {
	GetConsole().PrintOrError(console.Color(cl), GoWString(cstr))
}

//export nox_gui_console_Hide_4512B0
func nox_gui_console_Hide_4512B0() int32 { return int32(Nox_gui_console_Hide_4512B0()) }

func Nox_xxx_consoleEsc_49B7A0() {
	nox_xxx_consoleEsc_49B7A0()
}
