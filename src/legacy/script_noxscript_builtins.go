package legacy

/*
#include "defs.h"
int nox_script_RetreatLevel_515DF0();
int nox_script_RetreatLevelGroup_515E50();
int nox_script_SetResumeLevel_515E80();
int nox_script_SetResumeLevelGroup_515EE0();
void nox_script_StartupScreen_516600_A();
int sub_512E80(wchar2_t* a1);
*/
import "C"
import (
	"github.com/opennox/noxscript/ns/asm"

	"github.com/opennox/opennox/v1/server/noxscript"
)

var (
	Nox_script_shouldReadMoreXxx     func(fi asm.Builtin) bool
	Nox_script_shouldReadEvenMoreXxx func(fi asm.Builtin) bool
)

//export nox_script_shouldReadMoreXxx
func nox_script_shouldReadMoreXxx(fi_cgo int32) C.bool {
	fi := int(fi_cgo)
	return C.bool(Nox_script_shouldReadMoreXxx(asm.Builtin(fi)))
}

//export nox_script_shouldReadEvenMoreXxx
func nox_script_shouldReadEvenMoreXxx(fi_cgo int32) C.bool {
	fi := int(fi_cgo)
	return C.bool(Nox_script_shouldReadEvenMoreXxx(asm.Builtin(fi)))
}

func CallScriptBuiltin(fi asm.Builtin) (int, bool) {
	if fi < 0 || int(fi) >= len(noxScriptBuiltins) {
		return 0, false
	}
	fnc := noxScriptBuiltins[fi]
	if fnc == nil {
		return 0, false
	}
	res := fnc(GetServer().NoxScriptC())
	return res, true
}

func Nox_script_StartupScreen_516600_A() {
	C.nox_script_StartupScreen_516600_A()
}

func Sub_512E80(str string) int {
	cstr, _ := CWString(str)
	return int(C.sub_512E80(cstr))
}

var noxScriptBuiltins = [asm.BuiltinGetScore + 1]noxscript.Builtin{
	asm.BuiltinIsTalking: noxScriptIsTalkingBuiltin5166A0,
	asm.BuiltinIsTrading: noxScriptPlayerIsTradingBuiltin5166E0,
}
