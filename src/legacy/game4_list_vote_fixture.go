package legacy

/*
#include "game4_list_vote_native.h"
extern void* game4_fixture_calloc_go(size_t count, size_t size);
extern nox_alloc_class* game4_fixture_new_class_go(char* name, int size, int count);
// Only the declared allocation services are redirected. The reference
// function definitions and their inputs remain byte-for-byte unchanged.
#define calloc game4_fixture_calloc_go
#define nox_new_alloc_class game4_fixture_new_class_go
#include "game4_list_vote_original_reference.h"
#undef nox_new_alloc_class
#undef calloc
// Keep private static reference state within its C translation unit.
static nox_vote_5066D0* game4_reference_head_get_native(void) {
 return game4_reference_vote_head_5066D0;
}
static void game4_reference_head_set_native(nox_vote_5066D0* head) {
 game4_reference_vote_head_5066D0 = head;
}
*/
import "C"

import (
	"fmt"
	"runtime"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

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

type game4SupportObservation struct {
	Links           []int
	Freed           []int
	Bytes           []byte
	Packets         []game4SupportPacket
	NativeAddresses []uintptr
}
type game4SupportPacket struct {
	Recipient int32
	Bytes     []byte
}

func game4SupportGuarded(size uintptr) (unsafe.Pointer, unsafe.Pointer, func() bool) {
	base := game4SupportAllocate(1, size+64)
	if base == nil {
		panic("game4 support fixture allocation failed")
	}
	data := unsafe.Add(base, 32)
	all := unsafe.Slice((*byte)(base), int(size+64))
	for i := 0; i < 32; i++ {
		all[i] = 0xa5
		all[int(size)+32+i] = 0x5a
	}
	return base, data, func() bool {
		for i := 0; i < 32; i++ {
			if all[i] != 0xa5 || all[int(size)+32+i] != 0x5a {
				return false
			}
		}
		return true
	}
}
func game4SupportRole(p unsafe.Pointer, roles map[unsafe.Pointer]int) int {
	if p == nil {
		return 0
	}
	if role, ok := roles[p]; ok {
		return role
	}
	return -1
}

// Runs the same bounded constructors/lookup/free sequence through the public
// C ABI or the independently renamed complete original C definitions.
func game4SupportWalls(reference, freePayloads bool) (out game4SupportObservation, err error) {
	defer game4SupportHighFixtureMemory()()
	oldHead := game4WallHead()
	game4SetWallHead(nil)
	C.game4_reference_wall_list_set_head_504290(nil)
	defer game4SetWallHead(oldHead)
	var nodes [2]*C.nox_map_wall_list_node_504290
	var data [2]unsafe.Pointer
	roles := make(map[unsafe.Pointer]int)
	for i, xy := range [][2]int8{{-1, 2}, {17, 29}} {
		if reference {
			nodes[i] = C.game4_reference_sub_504290(C.int8_t(xy[0]), C.int8_t(xy[1]))
		} else {
			nodes[i] = C.sub_504290(C.int8_t(xy[0]), C.int8_t(xy[1]))
		}
		if nodes[i] == nil || nodes[i].wall == nil {
			return out, fmt.Errorf("wall allocation")
		}
		data[i] = game4SupportAllocate(32, 1)
		if data[i] == nil {
			return out, fmt.Errorf("wall payload allocation")
		}
		for j := range unsafe.Slice((*byte)(data[i]), 32) {
			unsafe.Slice((*byte)(data[i]), 32)[j] = byte(0x61 + i)
		}
		nodes[i].wall.flags = 4
		nodes[i].wall.data = data[i]
		roles[unsafe.Pointer(nodes[i])] = 10 + i
		roles[unsafe.Pointer(nodes[i].wall)] = 20 + i
		roles[data[i]] = 30 + i
		out.NativeAddresses = append(out.NativeAddresses, uintptr(unsafe.Pointer(nodes[i])), uintptr(unsafe.Pointer(nodes[i].wall)), uintptr(data[i]))
	}
	head := game4WallHead()
	if reference {
		head = C.game4_reference_wall_list_head_504290()
	}
	out.Links = append(out.Links, game4SupportRole(unsafe.Pointer(head), roles), game4SupportRole(unsafe.Pointer(nodes[1].next), roles), game4SupportRole(unsafe.Pointer(nodes[0].previous), roles))
	var hit, miss *C.nox_map_wall_list_node_504290
	if reference {
		hit = C.game4_reference_nox_xxx_cliWallGet_5042F0(255, 2)
		miss = C.game4_reference_nox_xxx_cliWallGet_5042F0(-1, 2)
	} else {
		hit = C.nox_xxx_cliWallGet_5042F0(255, 2)
		miss = C.nox_xxx_cliWallGet_5042F0(-1, 2)
	}
	out.Links = append(out.Links, game4SupportRole(unsafe.Pointer(hit), roles), game4SupportRole(unsafe.Pointer(miss), roles))
	oldObserver := game4SupportFreeObserver
	defer func() { game4SupportFreeObserver = oldObserver }()
	game4SupportFreeObserver = func(p unsafe.Pointer) {
		role := game4SupportRole(p, roles)
		out.Freed = append(out.Freed, role)
		if role >= 30 {
			expected := byte(0x61 + role - 30)
			for _, b := range unsafe.Slice((*byte)(p), 32) {
				if b != expected {
					err = fmt.Errorf("payload canary changed")
				}
			}
		}
	}
	flag := C.int32_t(0)
	if freePayloads {
		flag = 1
	}
	if reference {
		C.game4_reference_nox_mapgen_free_wall_list_503F40(flag)
		if C.game4_reference_wall_list_head_504290() != nil {
			err = fmt.Errorf("reference head not cleared")
		}
	} else {
		C.nox_mapgen_free_wall_list_503F40(flag)
		if game4WallHead() != nil {
			err = fmt.Errorf("native head not cleared")
		}
	}
	if !freePayloads {
		for i, p := range data {
			for _, b := range unsafe.Slice((*byte)(p), 32) {
				if b != byte(0x61+i) {
					err = fmt.Errorf("retained payload changed")
				}
			}
			game4SupportReleaseMemory(p)
		}
	}
	return
}

func game4SupportMapObjects(reference bool) (out game4SupportObservation, err error) {
	defer game4SupportHighFixtureMemory()()
	old := game4MapObjectHead()
	game4SetMapObjectHead(nil)
	C.game4_reference_map_object_list_set_head_5048A0(nil)
	defer game4SetMapObjectHead(old)
	var objects [3]*C.nox_object_t
	var nodes [3]*C.nox_map_object_list_node_5048A0
	var bases [3]unsafe.Pointer
	var guards [3]func() bool
	roles := make(map[unsafe.Pointer]int)
	defer func() {
		for _, p := range bases {
			if p != nil {
				game4SupportReleaseMemory(p)
			}
		}
	}()
	for i := range objects {
		base, data, guard := game4SupportGuarded(C.sizeof_nox_object_t)
		bases[i] = base
		guards[i] = guard
		objects[i] = (*C.nox_object_t)(data)
		if reference {
			nodes[i] = C.game4_reference_nox_xxx_unitAddToList_5048A0(objects[i])
		} else {
			nodes[i] = C.nox_xxx_unitAddToList_5048A0(objects[i])
		}
		if nodes[i] == nil {
			return out, fmt.Errorf("object node allocation")
		}
		roles[unsafe.Pointer(objects[i])] = 1 + i
		roles[unsafe.Pointer(nodes[i])] = 11 + i
		out.NativeAddresses = append(out.NativeAddresses, uintptr(data), uintptr(unsafe.Pointer(nodes[i])))
	}
	head := game4MapObjectHead()
	if reference {
		head = C.game4_reference_map_object_list_head_5048A0()
	}
	for n := head; n != nil; {
		var object *C.nox_object_t
		var next *C.nox_map_object_list_node_5048A0
		if reference {
			object = C.game4_reference_nox_map_object_list_node_object_5048A0(n)
			next = C.game4_reference_nox_map_object_list_node_next_5048A0(n)
		} else {
			object = C.nox_map_object_list_node_object_5048A0(n)
			next = C.nox_map_object_list_node_next_5048A0(n)
		}
		out.Links = append(out.Links, game4SupportRole(unsafe.Pointer(n), roles), game4SupportRole(unsafe.Pointer(n.previous), roles), game4SupportRole(unsafe.Pointer(next), roles), game4SupportRole(unsafe.Pointer(object), roles), game4SupportRole(unsafe.Pointer(object.object_prev), roles), game4SupportRole(unsafe.Pointer(object.object_next), roles))
		n = next
	}
	oldObserver := game4SupportFreeObserver
	defer func() { game4SupportFreeObserver = oldObserver }()
	game4SupportFreeObserver = func(p unsafe.Pointer) { out.Freed = append(out.Freed, game4SupportRole(p, roles)) }
	for n := head; n != nil; {
		next := n.next
		if reference {
			C.game4_reference_nox_map_object_list_node_free_5048A0(n)
		} else {
			C.nox_map_object_list_node_free_5048A0(n)
		}
		n = next
	}
	game4SetMapObjectHead(nil)
	C.game4_reference_map_object_list_set_head_5048A0(nil)
	for _, guard := range guards {
		if !guard() {
			err = fmt.Errorf("object boundary canary changed")
		}
	}
	return
}

func game4SupportPinnedMapObjects(a, b unsafe.Pointer) (bool, func()) {
	releaseMemory := game4SupportHighFixtureMemory()
	old := game4MapObjectHead()
	game4SetMapObjectHead(nil)
	first := C.nox_xxx_unitAddToList_5048A0((*C.nox_object_t)(a))
	second := C.nox_xxx_unitAddToList_5048A0((*C.nox_object_t)(b))
	runtime.GC()
	ok := first != nil && second != nil && C.nox_map_object_list_node_object_5048A0(first) == (*C.nox_object_t)(a) && C.nox_map_object_list_node_object_5048A0(second) == (*C.nox_object_t)(b) && second.next == first && first.previous == second
	return ok, func() {
		if second != nil {
			C.nox_map_object_list_node_free_5048A0(second)
		}
		if first != nil {
			C.nox_map_object_list_node_free_5048A0(first)
		}
		game4SetMapObjectHead(old)
		releaseMemory()
	}
}

func game4SupportVoteAdd(reference bool, voters uint32, count uint8, mask uint32) (ret int8, body []byte, guardOK bool) {
	defer game4SupportHighFixtureMemory()()
	base, data, guard := game4SupportGuarded(C.sizeof_nox_vote_5066D0)
	defer game4SupportReleaseMemory(base)
	vote := (*C.nox_vote_5066D0)(data)
	vote.voters = C.uint32_t(voters)
	vote.count = C.uint8_t(count)
	if reference {
		ret = int8(C.game4_reference_vote_add_voter_5066D0(vote, C.uint32_t(mask)))
	} else {
		ret = int8(game4VoteAddVoter(vote, mask))
	}
	return ret, append([]byte(nil), unsafe.Slice((*byte)(data), C.sizeof_nox_vote_5066D0)...), guard()
}

func game4SupportVoteUnlink(reference bool, removed int) (out game4SupportObservation, err error) {
	defer game4SupportHighFixtureMemory()()
	old := game4VoteHead
	game4VoteHead = nil
	C.game4_reference_head_set_native(nil)
	defer func() { game4VoteHead = old; C.game4_reference_head_set_native(nil) }()
	var nodes [3]*C.nox_vote_5066D0
	roles := make(map[unsafe.Pointer]int)
	for i := range nodes {
		nodes[i] = (*C.nox_vote_5066D0)(game4SupportAllocate(1, C.sizeof_nox_vote_5066D0))
		if nodes[i] == nil {
			return out, fmt.Errorf("vote allocation")
		}
		defer game4SupportReleaseMemory(unsafe.Pointer(nodes[i]))
		roles[unsafe.Pointer(nodes[i])] = i + 1
	}
	for i := range nodes {
		if reference {
			C.game4_reference_nox_xxx_voteAddMB_506AD0(nodes[i])
		} else {
			C.nox_xxx_voteAddMB_506AD0(nodes[i])
		}
		out.NativeAddresses = append(out.NativeAddresses, uintptr(unsafe.Pointer(nodes[i])))
	}
	var result, head *C.nox_vote_5066D0
	if reference {
		result = C.game4_reference_sub_506810(nodes[removed])
		head = C.game4_reference_head_get_native()
	} else {
		result = C.sub_506810(nodes[removed])
		head = game4VoteHead
	}
	out.Links = append(out.Links, game4SupportRole(unsafe.Pointer(result), roles), game4SupportRole(unsafe.Pointer(head), roles))
	for n := head; n != nil; n = n.next {
		out.Links = append(out.Links, game4SupportRole(unsafe.Pointer(n), roles), game4SupportRole(unsafe.Pointer(n.previous), roles), game4SupportRole(unsafe.Pointer(n.next), roles))
	}
	return
}

func game4SupportVoteLifecycle(reference bool) (out game4SupportObservation, err error) {
	defer game4SupportHighFixtureMemory()()
	// Isolate both allocators/heads without resetting any unrelated live owner.
	oldHead, oldAllocator, oldDiagnostic := game4VoteHead, C.nox_alloc_vote_1599652, C.dword_5d4594_1599656
	game4VoteHead = nil
	C.nox_alloc_vote_1599652 = nil
	C.dword_5d4594_1599656 = 0
	defer func() {
		game4VoteHead = oldHead
		C.nox_alloc_vote_1599652 = oldAllocator
		C.dword_5d4594_1599656 = oldDiagnostic
	}()
	oldSend := game4VotePacketObserver
	defer func() { game4VotePacketObserver = oldSend }()
	game4VotePacketObserver = func(recipient int32, b []byte) int32 {
		out.Packets = append(out.Packets, game4SupportPacket{recipient, append([]byte(nil), b...)})
		return 73
	}
	unit := (*C.nox_object_t)(game4SupportAllocate(1, C.sizeof_nox_object_t))
	if unit == nil {
		return out, fmt.Errorf("unit allocation")
	}
	defer game4SupportReleaseMemory(unsafe.Pointer(unit))
	unit.obj_class = 4
	var first, second *C.nox_vote_5066D0
	if reference {
		if C.game4_reference_nox_xxx_allocVoteArray_5066D0() != 1 {
			return out, fmt.Errorf("reference allocator")
		}
		defer C.game4_reference_sub_506720()
		first = C.game4_reference_sub_506A20(2, unit)
		second = C.game4_reference_sub_506A20(3, unit)
	} else {
		if C.nox_xxx_allocVoteArray_5066D0() != 1 {
			return out, fmt.Errorf("native allocator")
		}
		defer C.sub_506720()
		first = C.sub_506A20(2, unit)
		second = C.sub_506A20(3, unit)
	}
	if first == nil || second == nil {
		return out, fmt.Errorf("vote creation")
	}
	out.NativeAddresses = append(out.NativeAddresses, uintptr(unsafe.Pointer(unit)), uintptr(unsafe.Pointer(first)), uintptr(unsafe.Pointer(second)))
	out.Links = append(out.Links, int(first.threshold), int(second.threshold), int(first.frame), int(second.frame))
	first.voters = (1 << 1) | (1 << 31)
	first.count = 2
	if reference {
		C.game4_reference_sub_5067B0(first)
		C.game4_reference_sub_5067B0(second)
		out.Links = append(out.Links, int(C.game4_reference_sub_5071C0()))
	} else {
		C.sub_5067B0(first)
		C.sub_5067B0(second)
		out.Links = append(out.Links, int(C.sub_5071C0()))
	}
	return
}

func game4SupportNameEquality(a, b []uint16) (bool, bool) {
	x, y := (*C.wchar2_t)(unsafe.Pointer(&a[0])), (*C.wchar2_t)(unsafe.Pointer(&b[0]))
	got := game4VoteNamesEqual(x, y)
	want := C.nox_wcscmp(x, y) == 0
	runtime.KeepAlive(a)
	runtime.KeepAlive(b)
	return got, want
}

// Captures the target before the packet transport can reenter and mutate the
// live player's unit. Each backend owns a fresh native allocator, and restores
// any unrelated production head/class without resetting it.
func game4SupportVoteTargetCache(reference bool, unit unsafe.Pointer, name []uint16) (unsafe.Pointer, error) {
	defer game4SupportHighFixtureMemory()()
	oldHead, oldAllocator, oldDiagnostic := game4VoteHead, C.nox_alloc_vote_1599652, C.dword_5d4594_1599656
	oldMinimum := C.nox_server_kickQuestPlayerMinVotes_229992
	game4VoteHead = nil
	C.nox_alloc_vote_1599652 = nil
	C.dword_5d4594_1599656 = 0
	C.nox_server_kickQuestPlayerMinVotes_229992 = 5
	defer func() {
		game4VoteHead = oldHead
		C.nox_alloc_vote_1599652 = oldAllocator
		C.dword_5d4594_1599656 = oldDiagnostic
		C.nox_server_kickQuestPlayerMinVotes_229992 = oldMinimum
	}()
	cname := (*C.wchar2_t)(unsafe.Pointer(&name[0]))
	var vote *C.nox_vote_5066D0
	if reference {
		if C.game4_reference_nox_xxx_allocVoteArray_5066D0() != 1 {
			return nil, fmt.Errorf("reference allocator")
		}
		defer C.game4_reference_sub_506720()
		vote = C.game4_reference_sub_506B80(3, (*C.nox_object_t)(unit), cname)
	} else {
		if C.nox_xxx_allocVoteArray_5066D0() != 1 {
			return nil, fmt.Errorf("native allocator")
		}
		defer C.sub_506720()
		vote = C.sub_506B80(3, (*C.nox_object_t)(unit), cname)
	}
	runtime.KeepAlive(name)
	if vote == nil {
		return nil, fmt.Errorf("kick vote not created")
	}
	return unsafe.Pointer(vote.target), nil
}
