package legacy

/*
#include <stdint.h>

*/
import "C"
import (
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/client/audio/ail"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/dialog"
	"github.com/opennox/opennox/v1/legacy/timer"
)

func Set_dword_5d4594_831080(v uint32) {
	*memmap.PtrUint32(0x5D4594, 831080) = v
}

var (
	Dialogs            *dialog.Dialog
	dialogFallbackMode uint32 = 1
)

func initDialog() {
	Dialogs = dialog.NewDialog(
		"dialog",
		(*uint32)(&briefingState.dialog_enabled),
		(*uint32)(&clientAudioStorage.dialog_state), (*uint32)(&clientAudioStorage.dialog_flag), &dialogFallbackMode, (*uint32)(&clientAudioStorage.dialog_initialized), &dialogAudioDriver,
		memmap.PtrUint32(0x5D4594, 830860),
		func() *strman.StringManager {
			return GetServer().S().Strings()
		},
		memmap.PtrT[timer.TimerGroup](0x5D4594, 830876),
		memmap.PtrT[timer.TimerGroup](0x5D4594, 830980),
		(*timer.TimerGroup)(legacyPointers.ordinaryMusicTimer),
		(*timer.TimerGroup)(legacyPointers.dialogTimer),
		Sub_413890,

		func() ail.Driver {
			drv := Sub_43F130()
			setDialogAudioDriver(drv)
			return drv
		},
		func() { MusicModule.Sub_43DBD0() },
		func() { MusicModule.Sub_43DBE0() },
		func() bool { return MusicModule.Sub_43DC40() },
		Set_dword_5d4594_831080,
		func() uint32 { return uint32(Get_dword_587000_93160()) },
	)
}

//export sub_44D930
func sub_44D930() int32 {
	return int32(bool2int(Dialogs.Sub_44D930()))
}

//export nox_xxx_playDialogFile_44D900
func nox_xxx_playDialogFile_44D900(a1p *byte, a2_cgo int32) int32 {
	a2 := int(a2_cgo)
	return int32(Dialogs.PlayFile(alloc.GoString(a1p), a2))
}
