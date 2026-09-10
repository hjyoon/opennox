package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestSummonUnitExport5016C0PreservesExactWidthsAndNativePointers(t *testing.T) {
	old := Nox_xxx_unitDoSummonAt_5016C0
	t.Cleanup(func() {
		Nox_xxx_unitDoSummonAt_5016C0 = old
	})

	position, freePosition := alloc.New(types.Pointf{})
	defer freePosition()
	*position = types.Pointf{X: 12.5, Y: -7.25}
	owner, freeOwner := alloc.New(server.Object{})
	defer freeOwner()
	created, freeCreated := alloc.New(server.Object{})
	defer freeCreated()

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for index, pointer := range []unsafe.Pointer{
			unsafe.Pointer(position),
			unsafe.Pointer(owner),
			unsafe.Pointer(created),
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("pointer %d = %p, want native address above 4 GiB", index, pointer)
			}
		}
	}

	var observedTypeID int32
	var observedPosition *types.Pointf
	var observedOwner *server.Object
	var observedDirection uint8
	Nox_xxx_unitDoSummonAt_5016C0 = func(
		typeID int32,
		position *types.Pointf,
		owner *server.Object,
		direction uint8,
	) *server.Object {
		observedTypeID = typeID
		observedPosition = position
		observedOwner = owner
		observedDirection = direction
		return created
	}

	got := summonUnitExportCall5016C0(math.MinInt32, position, owner, math.MaxUint8)
	if got != created {
		t.Fatalf("created = %p, want %p", got, created)
	}
	if observedTypeID != math.MinInt32 || observedPosition != position || observedOwner != owner || observedDirection != math.MaxUint8 {
		t.Fatalf(
			"arguments = %d/%p/%p/%#x, want %d/%p/%p/%#x",
			observedTypeID,
			observedPosition,
			observedOwner,
			observedDirection,
			int32(math.MinInt32),
			position,
			owner,
			uint8(math.MaxUint8),
		)
	}

	Nox_xxx_unitDoSummonAt_5016C0 = func(
		typeID int32,
		position *types.Pointf,
		owner *server.Object,
		direction uint8,
	) *server.Object {
		observedTypeID = typeID
		observedPosition = position
		observedOwner = owner
		observedDirection = direction
		return nil
	}
	if got := summonUnitExportCall5016C0(math.MaxInt32, nil, nil, 0); got != nil {
		t.Fatalf("nil result = %p", got)
	}
	if observedTypeID != math.MaxInt32 || observedPosition != nil || observedOwner != nil || observedDirection != 0 {
		t.Fatalf("nil arguments = %d/%p/%p/%#x", observedTypeID, observedPosition, observedOwner, observedDirection)
	}

	runtime.KeepAlive(position)
	runtime.KeepAlive(owner)
	runtime.KeepAlive(created)
}
