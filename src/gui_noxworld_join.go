package opennox

import (
	"net/netip"
	"unsafe"

	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func startNoxWorldJoin() {
	// GAME.EXE 43A02B..43A06C sets 69B558 after opening the player
	// picker. Its retained browser keeps 438770 installed throughout.
	// The native browser releases that owner, so transfer the frame hook
	// after cleanup rather than losing both the gate and the dispatcher.
	legacy.Set_dword_5d4594_815044(1)
	legacy.Sub_43AF90(0)
	noxServer.SetUpdateFunc2(noxWorldPollJoin)
}

func noxWorldPollJoin() bool {
	if legacy.Sub_43AF80() == 0 {
		// Player selection has not sent the request yet. Do not enter the
		// old dispatcher's discovery branch with its unused window globals.
		return true
	}
	// Retain the existing 438770 states, including the 20-second deadline,
	// ping/password handling, delayed retry and connected draw callback.
	return legacy.Sub_438770()
}

func noxWorldJoinDialogOK() {
	switch legacy.Sub_43AF80() {
	case 6:
		text := alloc.GoString16((*uint16)(gui.EventRespPtr(sub449E60(4))))
		buf, free := alloc.Make([]byte{}, 22)
		defer free()
		// Original 43A09D..43A101 copies nine UTF-16 units at +4,
		// then terminates unit eight before sending the 22-byte packet.
		alloc.StrCopy16B(buf[4:22], text)
		*(*uint16)(unsafe.Pointer(&buf[20])) = 0
		addr := netip.AddrPortFrom(legacy.Nox_client_getServerAddr_43B300(), uint16(legacy.ClientGetServerPort()))
		_, _ = sendServerPassword(addr, buf)
		legacy.Sub_43AF90(3)
		legacy.Set_qword_5d4594_814956(platformTicks() + 20000)
		sub449EA0(0)
	case 1:
		noxWorldJoinDialogCancel()
	}
}

func noxWorldJoinDialogCancel() {
	// Original 43A07B resets 43A920 and reopens 4379F0. The native
	// browser is no longer retained, so close its dialog before rebuilding.
	legacy.Set_dword_5d4594_815044(0)
	legacy.Sub_43AF90(0)
	sub_44A400()
	noxGameShowGameSelNative()
}
