package legacy

/*
#include "client__gui__window.h"
*/
import "C"

import "github.com/opennox/opennox/v1/client/gui"

var inputCfgRestoreFocusCEntry = C.nox_gui_input_cfg_restore_focus

// The shell InputCfg controls are NOFOCUS. Its key-capture modal must return
// keyboard input to the same background used by the shell animation, not to
// a binding column or the InputCfg root. The in-game modal does not use this.
//
//export nox_gui_input_cfg_restore_focus
func nox_gui_input_cfg_restore_focus() { gui.FocusMainBg() }
