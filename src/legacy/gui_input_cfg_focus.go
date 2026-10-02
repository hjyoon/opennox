package legacy

/*
#include "client__gui__window.h"
#include "GAME3_1.h"
extern nox_window* dword_5d4594_1522612;
extern nox_window* dword_5d4594_1522632;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/client/gui"
)

var (
	inputCfgRestoreFocusCEntry         = C.nox_gui_input_cfg_restore_focus
	inputCfgCapturePromptCEntry4CC170  = C.sub_4CC170
	inputCfgCapturePromptSlot4CC170    = (**gui.Window)(unsafe.Pointer(&C.dword_5d4594_1522612))
	inputCfgCaptureSelectionSlot4CC170 = (**gui.Window)(unsafe.Pointer(&C.dword_5d4594_1522632))
)

// The shell InputCfg controls are NOFOCUS. Its key-capture modal must return
// keyboard input to the same background used by the shell animation, not to
// a binding column or the InputCfg root. The in-game modal does not use this.
//
//export nox_gui_input_cfg_restore_focus
func nox_gui_input_cfg_restore_focus() { gui.FocusMainBg() }
