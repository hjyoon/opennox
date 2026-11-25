package legacy

/*
#include "game4_list_vote_native.h"
*/
import "C"

import (
	"fmt"
	"runtime"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
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
	result := C.nox_xxx_netSendPacket1_4E5390(C.int(recipient), C.nox_important_payload_t(unsafe.Pointer(&packet[0])), C.int(len(packet)), nil, 1)
	runtime.KeepAlive(packet)
	return int32(result)
}

//export game4_fixture_calloc_go
func game4_fixture_calloc_go(count, size C.size_t) unsafe.Pointer {
	return game4SupportAllocate(uintptr(count), uintptr(size))
}

//export game4_fixture_release_go
func game4_fixture_release_go(p unsafe.Pointer) { game4SupportReleaseMemory(p) }

//export game4_fixture_new_class_go
func game4_fixture_new_class_go(name *C.char, size, count C.int) *C.nox_alloc_class {
	return game4SupportNewClass(name, size, count)
}

func game4SupportHighFixtureMemory() func() {
	oldAllocate, oldRelease, oldClass := game4SupportAllocationService, game4SupportReleaseService, game4SupportClassService
	owners := make(map[unsafe.Pointer]func())
	game4SupportAllocationService = func(count, size uintptr) unsafe.Pointer {
		p, release := alloc.CallocHigh(int(count), size)
		owners[p] = release
		return p
	}
	game4SupportReleaseService = func(p unsafe.Pointer) {
		if release, ok := owners[p]; ok {
			delete(owners, p)
			release()
		} else if oldRelease != nil {
			oldRelease(p)
		} else {
			C.free(p)
		}
	}
	game4SupportClassService = func(name string, size uintptr, count int) unsafe.Pointer {
		return alloc.NewClassHigh(name, size, count).UPtr()
	}
	return func() {
		game4SupportAllocationService, game4SupportReleaseService, game4SupportClassService = oldAllocate, oldRelease, oldClass
		if len(owners) != 0 {
			panic(fmt.Sprintf("GAME4 high fixture retained %d foreign allocations", len(owners)))
		}
	}
}

// These diagnostic callbacks preserve the real free operation and copy the
// synchronous reference packet before its C stack storage expires.
//
//export game4_reference_free_go
func game4_reference_free_go(p unsafe.Pointer) { game4SupportFree(p) }

//export game4_reference_send_packet_go
func game4_reference_send_packet_go(recipient C.int, packet unsafe.Pointer, size C.int, related unsafe.Pointer, reliable C.int) C.int {
	return C.int(game4VoteSend(int32(recipient), C.GoBytes(packet, size)))
}
