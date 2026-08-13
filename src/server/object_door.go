package server

import (
	"unsafe"
)

// DoorUpdateData is the 52-byte DoorUpdate record used by GAME.EXE. Only the
// fields whose meaning is established are named; the remaining bytes stay
// explicit so every supported pointer width has the original wire layout.
type DoorUpdateData struct {
	Field0    uint8    // 0, 0
	LockCode  uint8    // 0, 1
	_         [14]byte // 0, 2
	TileX     int32    // 4, 16
	TileY     int32    // 5, 20
	_         [24]byte // 6, 24
	QuestSync uint8    // 12, 48
	_         [3]byte  // 12, 49
}

var _ = [1]struct{}{}[52-unsafe.Sizeof(DoorUpdateData{})]
var _ = [1]struct{}{}[1-unsafe.Offsetof(DoorUpdateData{}.LockCode)]
var _ = [1]struct{}{}[16-unsafe.Offsetof(DoorUpdateData{}.TileX)]
var _ = [1]struct{}{}[20-unsafe.Offsetof(DoorUpdateData{}.TileY)]
var _ = [1]struct{}{}[48-unsafe.Offsetof(DoorUpdateData{}.QuestSync)]

func (obj *Object) UpdateDataDoor() *DoorUpdateData {
	return (*DoorUpdateData)(obj.UpdateData)
}

type serverDoors struct {
	flagXxx bool

	keyHolder    *Object
	keyHolderSet bool
}

func (s *serverDoors) Sub_4D72C0() bool {
	return s.flagXxx
}

func (s *serverDoors) Sub_4D72B0(a1 bool) {
	s.flagXxx = a1
}
