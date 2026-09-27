package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/common/memmap"
)

func TestQuitMenuAutoSaveUsesNativePlayerPointer445830(t *testing.T) {
	playerSlot := memmap.PtrPtr(0x852978, 8)
	oldPlayer := *playerSlot
	t.Cleanup(func() { *playerSlot = oldPlayer })

	player := &client.Drawable{}
	var pin runtime.Pinner
	pin.Pin(player)
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(player)) <= math.MaxUint32 {
		t.Fatalf("drawable pointer = %p, want native address above 4 GiB", player)
	}

	*playerSlot = unsafe.Pointer(player)
	if !quitMenuCanAutoSave445830() {
		t.Fatal("living player should be allowed to use AutoSave")
	}

	player.ObjFlags |= object.FlagDead
	if quitMenuCanAutoSave445830() {
		t.Fatal("dead player should not be allowed to use AutoSave")
	}

	*playerSlot = nil
	if quitMenuCanAutoSave445830() {
		t.Fatal("missing player should not be allowed to use AutoSave")
	}
	runtime.KeepAlive(player)
}
