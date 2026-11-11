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

// Optional allocation-service seams let native-pointer fixtures explicitly
// request foreign high addresses. Nil retains the ordinary product allocator.
var game4SupportAllocationService func(uintptr, uintptr) unsafe.Pointer
var game4SupportReleaseService func(unsafe.Pointer)
var game4SupportClassService func(string, uintptr, int) unsafe.Pointer

func game4SupportAllocate(count, size uintptr) unsafe.Pointer {
	if game4SupportAllocationService != nil {
		return game4SupportAllocationService(count, size)
	}
	return C.calloc(C.size_t(count), C.size_t(size))
}

func game4SupportReleaseMemory(p unsafe.Pointer) {
	if game4SupportReleaseService != nil {
		game4SupportReleaseService(p)
	} else {
		C.free(p)
	}
}

func game4SupportNewClass(name *C.char, size, count C.int) *C.nox_alloc_class {
	if game4SupportClassService != nil {
		return (*C.nox_alloc_class)(game4SupportClassService(C.GoString(name), uintptr(size), int(count)))
	}
	return C.nox_new_alloc_class(name, size, count)
}

func game4SupportFree(p unsafe.Pointer) {
	if game4SupportFreeObserver != nil {
		game4SupportFreeObserver(p)
	}
	game4SupportReleaseMemory(p)
}

func game4VoteSend(recipient int32, packet []byte) int32 {
	if game4VotePacketObserver != nil {
		return game4VotePacketObserver(recipient, packet)
	}
	result := C.nox_xxx_netSendPacket1_4E5390(C.int(recipient), unsafe.Pointer(&packet[0]), C.int(len(packet)), nil, 1)
	runtime.KeepAlive(packet)
	return int32(result)
}
