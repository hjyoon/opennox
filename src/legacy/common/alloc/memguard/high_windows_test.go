package memguard

import (
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

func highFixtureMemoryInfo(address uintptr) (windows.MemoryBasicInformation, error) {
	var info windows.MemoryBasicInformation
	if unsafe.Sizeof(uintptr(0)) == 8 {
		err := windows.VirtualQuery(address, &info, unsafe.Sizeof(info))
		return info, err
	}
	// winnt.h includes PartitionId only under _WIN64. The x/sys transport
	// structure includes it unconditionally, so decode the actual 28-byte
	// Windows/386 result from its larger, suitably aligned backing storage.
	type native32 struct {
		BaseAddress, AllocationBase, AllocationProtect uint32
		RegionSize, State, Protect, Type               uint32
	}
	var shape native32
	err := windows.VirtualQuery(address, &info, unsafe.Sizeof(shape))
	value := *(*native32)(unsafe.Pointer(&info))
	return windows.MemoryBasicInformation{
		BaseAddress: uintptr(value.BaseAddress), AllocationBase: uintptr(value.AllocationBase),
		AllocationProtect: value.AllocationProtect, RegionSize: uintptr(value.RegionSize),
		State: value.State, Protect: value.Protect, Type: value.Type,
	}, err
}

func TestNewHighPageProtection(t *testing.T) {
	size := PageSize() + 1
	b, release := NewHigh(size)
	released := false
	defer func() {
		if !released {
			release()
		}
	}()
	address := uintptr(unsafe.Pointer(unsafe.SliceData(b)))
	if len(b) != size || address%uintptr(PageSize()) != 0 {
		t.Fatal("high opaque reservation extent or alignment changed")
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && address <= uintptr(^uint32(0)) {
		t.Fatal("high opaque reservation does not exercise native upper bits")
	}
	info, err := highFixtureMemoryInfo(address)
	if err != nil {
		t.Fatal(err)
	}
	if info.AllocationBase != address || info.RegionSize < uintptr(size) || info.State != windows.MEM_RESERVE || info.AllocationProtect != windows.PAGE_NOACCESS {
		t.Fatalf("opaque fixture is not a reserved inaccessible region: %+v", info)
	}
	release()
	released = true
	info, err = highFixtureMemoryInfo(address)
	if err != nil {
		t.Fatal(err)
	}
	const memFree = 0x10000 // MEMORY_BASIC_INFORMATION.State: winnt.h MEM_FREE.
	if info.State != memFree {
		t.Fatalf("released opaque reservation remains allocated: %+v", info)
	}
}
