package legacy

/*
#include "client__network__inform.h"
#include "server__magic__plyrspel.h"
*/
import "C"

func spellResultPacketInformCall4C9BF0(status uint32) int {
	packet := [6]C.uint8_t{
		0xA9,
		0,
		C.uint8_t(status),
		C.uint8_t(status >> 8),
		C.uint8_t(status >> 16),
		C.uint8_t(status >> 24),
	}
	return int(C.nox_client_handlePacketInform_4C9BF0(&packet[0]))
}

func spellResultExportCall4FB0B0(status uint32) {
	C.nox_xxx_abilGetError_4FB0B0_magic_plyrspel(C.uint32_t(status))
}

//export nox_xxx_abilGetError_4FB0B0_magic_plyrspel
func nox_xxx_abilGetError_4FB0B0_magic_plyrspel(status C.uint32_t) {
	GetClient().SpellResult4FB0B0(uint32(status))
}
