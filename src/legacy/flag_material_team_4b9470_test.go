//go:build !server

package legacy

import (
	"math"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestFlagMaterialTeam4B9470ReadsNativePointerSlotsAndPackedIDs(t *testing.T) {
	InitBlobData()
	for row := uintptr(0); row < 9; row++ {
		namePtr := *memmap.PtrPtr(0x587000, 177488+8*row)
		if namePtr == nil {
			t.Fatalf("missing shipped flag material name at row %d", row)
		}
		name := GoStringP(namePtr)
		input, free := alloc.CString(name)
		if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(namePtr) <= math.MaxUint32 || uintptr(unsafe.Pointer(input)) <= math.MaxUint32) {
			free()
			t.Fatalf("flag material %q pointers must be above 4 GiB: table=%p input=%p", name, namePtr, input)
		}
		want := int(memmap.Int32(0x587000, 177492+8*row))
		if got := flagMaterialTeamCall4B9470(input); got != want {
			free()
			t.Fatalf("flag material %q team = %d, want %d", name, got, want)
		}
		if got := flagMaterialDrawableTeamCall4B94E0(input, 0x10000000); got != want {
			free()
			t.Fatalf("drawable material %q team = %d, want %d", name, got, want)
		}
		if got := flagMaterialDrawableTeamCall4B94E0(input, 0); got != 0 {
			free()
			t.Fatalf("non-flag drawable material %q team = %d, want 0", name, got)
		}
		free()
	}
	if got := flagMaterialTeamNilCall4B9470(); got != 0 {
		t.Fatalf("nil material result = %d, want 0", got)
	}
	for _, name := range []string{"not-a-flag-material", strings.ToLower(GoStringP(*memmap.PtrPtr(0x587000, 177488)))} {
		input, free := alloc.CString(name)
		got := flagMaterialTeamCall4B9470(input)
		free()
		if got != 0 {
			t.Fatalf("unknown/case-mismatched material %q team = %d, want 0", name, got)
		}
	}
}

func TestFlagMaterialTeam4B9470LiveSlotSentinelAndFullWidthID(t *testing.T) {
	InitBlobData()
	first := memmap.PtrPtr(0x587000, 177488)
	oldName := *first
	id := memmap.PtrInt32(0x587000, 177492)
	oldID := *id
	input, free := alloc.CString("native-flag-material")
	t.Cleanup(free)
	t.Cleanup(func() { *first, *id = oldName, oldID })
	*first, *id = unsafe.Pointer(input), 0x12345678
	if got := flagMaterialTeamCall4B9470(input); got != int(*id) {
		t.Fatalf("live native material team = %#x, want %#x", got, *id)
	}
	*id = -7
	if got := flagMaterialTeamCall4B9470(input); got != -7 {
		t.Fatalf("signed team result = %d, want -7", got)
	}
	*first = nil
	if got := flagMaterialTeamCall4B9470(input); got != 0 {
		t.Fatalf("empty table result = %d, want 0", got)
	}
}
