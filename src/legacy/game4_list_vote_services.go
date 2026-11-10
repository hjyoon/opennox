package legacy

/*
#include "game4_list_vote_native.h"
*/
import "C"

import (
	"runtime"
	"unsafe"
)

// Optional synchronous diagnostic boundaries. Production always uses the real
// allocator/network service when no observer is installed. An observer must
// not retain the packet argument or mutate objects being released.
var game4SupportFreeObserver func(unsafe.Pointer)
var game4VotePacketObserver func(int32, []byte) int32

func game4SupportFree(p unsafe.Pointer) {
	if game4SupportFreeObserver != nil {
		game4SupportFreeObserver(p)
	}
	C.free(p)
}

func game4VoteSend(recipient int32, packet []byte) int32 {
	if game4VotePacketObserver != nil {
		return game4VotePacketObserver(recipient, packet)
	}
	result := C.nox_xxx_netSendPacket1_4E5390(C.int(recipient), unsafe.Pointer(&packet[0]), C.int(len(packet)), nil, 1)
	runtime.KeepAlive(packet)
	return int32(result)
}
