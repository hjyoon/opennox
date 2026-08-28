package legacy

/*
#include "xfer_inventory_4f3e30.h"
*/
import "C"

import (
	"fmt"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

const inventoryXferTOCVersion4F3E30 = uint16(60)

type inventoryXferDeps4F3E30[T comparable] struct {
	readName      func() (string, error)
	readTOC       func() (uint16, error)
	lookupName    func(string) int32
	lookupTOC     func(uint16) int32
	readCRC       func() error
	newObject     func(uint16) T
	callXfer      func(T) error
	storePrevious func(T, T)
	loadFirst     func(T) T
	storeNext     func(T, T)
	storeFirst    func(T, T)
	storeHolder   func(T, T)
}

// xferInventory4F3E30 preserves GAME.EXE's signed count gate, uint16 version
// split, low-word type lookup, recursive transfer, and exact prepend order.
// The owner is deliberately not inspected before the first item's transfer.
func xferInventory4F3E30[T comparable](
	version uint16,
	owner T,
	count int32,
	deps inventoryXferDeps4F3E30[T],
) error {
	if count <= 0 {
		return nil
	}
	var zero T
	for i := int32(0); i < count; i++ {
		var typeInd uint16
		if version < inventoryXferTOCVersion4F3E30 {
			name, err := deps.readName()
			if err != nil {
				return fmt.Errorf("inventory[%d] type name: %w", i, err)
			}
			typeInd = uint16(deps.lookupName(name))
		} else {
			objectTOC, err := deps.readTOC()
			if err != nil {
				return fmt.Errorf("inventory[%d] TOC: %w", i, err)
			}
			typeInd = uint16(deps.lookupTOC(objectTOC))
		}
		if typeInd == 0 {
			return fmt.Errorf("inventory[%d] has unknown type", i)
		}
		if err := deps.readCRC(); err != nil {
			return fmt.Errorf("inventory[%d] CRC: %w", i, err)
		}
		item := deps.newObject(typeInd)
		if item == zero {
			return fmt.Errorf("inventory[%d] cannot allocate type %d", i, typeInd)
		}
		if err := deps.callXfer(item); err != nil {
			return fmt.Errorf("inventory[%d]: %w", i, err)
		}

		deps.storePrevious(item, zero)
		deps.storeNext(item, deps.loadFirst(owner))
		if first := deps.loadFirst(owner); first != zero {
			deps.storePrevious(first, item)
		}
		deps.storeFirst(owner, item)
		deps.storeHolder(item, owner)
	}
	return nil
}

func readInventoryTypeName4F3E30(cf *cryptfile.CryptFile) (string, error) {
	size, err := monsterRWU8(cf, 0)
	if err != nil {
		return "", err
	}
	buf := make([]byte, int(size))
	if err := monsterRWBytes528DB0(cf, buf); err != nil {
		return "", err
	}
	return cStringBytes528DB0(buf), nil
}

func xferInventoryNative4F3E30(
	cf *cryptfile.CryptFile,
	s *server.Server,
	owner *server.Object,
	version uint16,
	count int32,
) error {
	return xferInventory4F3E30(version, owner, count, inventoryXferDeps4F3E30[*server.Object]{
		readName: func() (string, error) {
			return readInventoryTypeName4F3E30(cf)
		},
		readTOC: func() (uint16, error) {
			return monsterRWU16(cf, 0)
		},
		lookupName: func(name string) int32 {
			return int32(s.Types.IndByID(name))
		},
		lookupTOC: func(objectTOC uint16) int32 {
			return int32(Nox_xxx_objectTOCgetTT(objectTOC))
		},
		readCRC: func() error {
			var crc [4]byte
			return cf.ReadMaybeAlign(crc[:])
		},
		newObject: func(typeInd uint16) *server.Object {
			return s.NewObjectByTypeInd(int(typeInd))
		},
		callXfer: func(item *server.Object) error {
			return item.CallXfer(nil)
		},
		storePrevious: func(item, previous *server.Object) {
			item.Field125 = previous
		},
		loadFirst: func(owner *server.Object) *server.Object {
			return owner.InvFirstItem
		},
		storeNext: func(item, next *server.Object) {
			item.InvNextItem = next
		},
		storeFirst: func(owner, first *server.Object) {
			owner.InvFirstItem = first
		},
		storeHolder: func(item, holder *server.Object) {
			item.InvHolder = holder
		},
	})
}

func xferInventoryCall4F3E30(owner *server.Object, version uint16, count int32) int32 {
	// GAME.EXE returns before touching the owner, stream, or server state.
	// Keep this gate outside the argument evaluation below as well.
	if count <= 0 {
		return 1
	}
	if err := xferInventoryNative4F3E30(cryptfile.Global(), GetServer().S(), owner, version, count); err != nil {
		mapLog.Printf("nox_xxx_xfer_4F3E30: %v", err)
		return 0
	}
	return 1
}

//export nox_xxx_xfer_4F3E30
func nox_xxx_xfer_4F3E30(version C.uint16_t, owner *C.nox_object_t, count C.int32_t) C.int32_t {
	return C.int32_t(xferInventoryCall4F3E30(
		asObjectS((*nox_object_t)(owner)),
		uint16(version),
		int32(count),
	))
}

func Nox_xxx_xfer_4F3E30(version int, owner *server.Object, count uint32) int {
	return int(xferInventoryCall4F3E30(owner, uint16(version), int32(count)))
}

func monsterXferInventory4F3E30(
	cf *cryptfile.CryptFile,
	owner *server.Object,
	version int,
	count uint32,
) error {
	signedCount := int32(count)
	if signedCount <= 0 {
		return nil
	}
	return xferInventoryNative4F3E30(cf, GetServer().S(), owner, uint16(version), signedCount)
}
