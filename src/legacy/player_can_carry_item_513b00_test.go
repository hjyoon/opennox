package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

type carryLegacyServer513B00 struct {
	Server
	srv *server.Server
}

func (s *carryLegacyServer513B00) S() *server.Server { return s.srv }

func TestPlayerCanCarryItem513B00PreservesNativeObject(t *testing.T) {
	srv := new(server.Server)
	oldGetServer := GetServer
	GetServer = func() Server { return &carryLegacyServer513B00{srv: srv} }
	t.Cleanup(func() { GetServer = oldGetServer })
	glyph := memmap.PtrUint32(0x5D4594, 2386856)
	oldGlyph := *glyph
	oldPickupCount := Get_dword_5d4594_2386848()
	oldWarned := Get_dword_5d4594_2386852()
	// The C globals are restored through their existing setters after this
	// test. We intentionally avoid another pointer-bearing C helper here.
	Set_dword_5d4594_2386848(80)
	Set_dword_5d4594_2386852(0)
	*glyph = 1
	t.Cleanup(func() {
		*glyph = oldGlyph
		Set_dword_5d4594_2386848(oldPickupCount)
		Set_dword_5d4594_2386852(int(oldWarned))
	})

	owner := &server.Object{InvFirstItem: &server.Object{TypeInd: 7, Worth: 999999}}
	incoming := &server.Object{TypeInd: 8}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(owner)) <= 0xffffffff || uintptr(unsafe.Pointer(incoming)) <= 0xffffffff) {
		t.Skip("allocator did not provide high-address objects")
	}
	// The original C path crashed at incoming+4 before it could test capacity.
	// At a saturated count, the inventory walker and cost path also run; the
	// threshold cost leaves no item to drop and needs no map or player fixture.
	Nox_xxx_playerCanCarryItem_513B00(owner, incoming)
	if owner.InvFirstItem == nil || owner.InvFirstItem.Worth != 999999 {
		t.Fatal("native inventory link was lost")
	}
}
