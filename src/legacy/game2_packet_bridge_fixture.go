package legacy

/*
#include "game2_packet_bridge.h"
int nox_xxx_netOnPacketRecvCli_48EA70_switch(int player, int op, unsigned char* data, int size);
*/
import "C"
import "unsafe"

func game2PacketBridgeFixture(player, op int32, data unsafe.Pointer, size int32, viaSwitch bool) int32 {
	if viaSwitch {
		return int32(C.nox_xxx_netOnPacketRecvCli_48EA70_switch(C.int(player), C.int(op), (*C.uchar)(data), C.int(size)))
	}
	return int32(C.nox_game2_decode_native_object(C.int(player), C.int(op), (*C.uchar)(data), C.int(size)))
}
