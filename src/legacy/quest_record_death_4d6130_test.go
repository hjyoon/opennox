package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestQuestRecordDeathBridge4D6130NativeCGoRoundTrip(t *testing.T) {
	unit, freeUnit := alloc.New(server.Object{})
	defer freeUnit()
	update, freeUpdate := alloc.New(server.PlayerUpdateData{})
	defer freeUpdate()
	player, freePlayer := alloc.New(server.Player{})
	defer freePlayer()
	unit.UpdateData, update.Player = unsafe.Pointer(update), player
	if got := questRecordDeathBridge4D6130(nil); got != 0 {
		t.Fatalf("nil result = %#x", got)
	}
	if got := questRecordDeathBridge4D6130(unit); got != uintptr(unsafe.Pointer(player)) {
		t.Fatalf("live result = %#x, want Player %p", got, player)
	}
	unit.ObjFlags = object.FlagDestroyed
	if got := questRecordDeathBridge4D6130(unit); got != uintptr(unsafe.Pointer(unit)) {
		t.Fatalf("destroyed result = %#x, want Object %p", got, unit)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]unsafe.Pointer{"unit": unsafe.Pointer(unit), "update": unsafe.Pointer(update), "player": unsafe.Pointer(player)} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("%s native allocation = %p, want above 4 GiB", name, pointer)
			}
		}
	}
}
