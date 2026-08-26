package server

import (
	"unsafe"
)

// DoorUpdateData is the 52-byte DoorUpdate record used by GAME.EXE. Only the
// fields whose meaning is established are named; the remaining bytes stay
// explicit so every supported pointer width has the original wire layout.
type DoorUpdateData struct {
	Field0           uint8   // 0, 0
	LockCode         uint8   // 0, 1
	_                [2]byte // 0, 2
	TargetDirection  int32   // 1, 4
	SyncedDirection  int32   // 2, 8
	CurrentDirection int32   // 3, 12
	TileX            int32   // 4, 16
	TileY            int32   // 5, 20
	_                [4]byte // 6, 24
	Queued           uint32  // 7, 28
	AngularVelocity  float32 // 8, 32
	_                [4]byte // 9, 36
	FractionalDir    int16   // 10, 40
	_                [2]byte // 10, 42
	LastMoveFrame    uint32  // 11, 44
	QuestSync        uint8   // 12, 48
	_                [3]byte // 12, 49
}

var _ = [1]struct{}{}[52-unsafe.Sizeof(DoorUpdateData{})]
var _ = [1]struct{}{}[1-unsafe.Offsetof(DoorUpdateData{}.LockCode)]
var _ = [1]struct{}{}[4-unsafe.Offsetof(DoorUpdateData{}.TargetDirection)]
var _ = [1]struct{}{}[8-unsafe.Offsetof(DoorUpdateData{}.SyncedDirection)]
var _ = [1]struct{}{}[12-unsafe.Offsetof(DoorUpdateData{}.CurrentDirection)]
var _ = [1]struct{}{}[16-unsafe.Offsetof(DoorUpdateData{}.TileX)]
var _ = [1]struct{}{}[20-unsafe.Offsetof(DoorUpdateData{}.TileY)]
var _ = [1]struct{}{}[28-unsafe.Offsetof(DoorUpdateData{}.Queued)]
var _ = [1]struct{}{}[32-unsafe.Offsetof(DoorUpdateData{}.AngularVelocity)]
var _ = [1]struct{}{}[40-unsafe.Offsetof(DoorUpdateData{}.FractionalDir)]
var _ = [1]struct{}{}[44-unsafe.Offsetof(DoorUpdateData{}.LastMoveFrame)]
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
