package legacy

/*
#include "GAME3_3.h"
#include "GAME4_2.h"
*/
import "C"

import (
	"image"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_netSendVampFx_523270
func nox_xxx_netSendVampFx_523270(op C.char, coordinates *C.short, amount C.short) C.int {
	words := (*[7]C.short)(unsafe.Pointer(coordinates))
	from := image.Pt(int(uint16(words[0])), int(uint16(words[2])))
	to := image.Pt(int(uint16(words[4])), int(uint16(words[6])))
	position := types.Ptf(float32(to.X), float32(to.Y))
	packet := server.VampirismFXPacket523270(netmsg.Op(byte(op)), from, to, uint16(amount))
	return C.int(pointFXRuntimeForCall522FF0().Send(&position, packet[:]))
}

func queueSummonFX(packet []byte) C.int {
	return nox_xxx_netSendPacket0_4E5420(
		255, C.nox_important_payload_t(unsafe.Pointer(unsafe.SliceData(packet))), C.int(len(packet)), nil, 1,
	)
}

//export nox_xxx_sendSummonStartFX_5236F0
func nox_xxx_sendSummonStartFX_5236F0(effectID C.short, point *C.float, direction C.char, typeID, duration C.short) C.int {
	position := *(*types.Pointf)(unsafe.Pointer(point))
	packet := server.SummonStartEffectPacket(uint16(effectID), position, byte(direction), uint16(typeID), uint16(duration))
	return queueSummonFX(packet[:])
}

//export nox_xxx_sendSummonCancelFX_523760
func nox_xxx_sendSummonCancelFX_523760(effectID C.short) C.int {
	packet := server.SummonCancelEffectPacket(uint16(effectID))
	return queueSummonFX(packet[:])
}

//export nox_xxx_sendGeneratorSpawnFX_523830
func nox_xxx_sendGeneratorSpawnFX_523830(coordinates *C.int4, length C.short) {
	values := *(*[4]int32)(unsafe.Pointer(coordinates))
	server.SendGeneratorSpawnFXNative523830(values, uint16(length), pointFXRuntimeForCall522FF0())
}

//export nox_xxx_sendArrowTrapFX_5238A0
func nox_xxx_sendArrowTrapFX_5238A0(position *C.float, variant C.char) {
	server.SendArrowTrapFXNative5238A0((*types.Pointf)(unsafe.Pointer(position)), byte(variant), pointFXRuntimeForCall522FF0())
}
