package server

import (
	"unsafe"

	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
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

func (s *Server) PlayersHaveSilverKey() *Object {
	var found *Object
	for pl := s.Players.First(); pl != nil; pl = s.Players.Next(pl) {
		if !pl.IsActive() {
			continue
		}
		u := pl.PlayerUnit
		if u == nil {
			continue
		}
		cnt := 0
		for it := u.FirstItem(); it != nil; it = it.NextItem() {
			if int(it.TypeInd) == s.Types.SilverKeyID() {
				cnt++
			}
		}
		if cnt > 0 {
			found = u
		}
	}
	if found == nil {
		return nil
	}
	for it := found.FirstItem(); it != nil; it = it.NextItem() {
		if int(it.TypeInd) == s.Types.SilverKeyID() {
			return it
		}
	}
	return nil
}

func (s *Server) DoorCheckKey(u, door *Object) *Object {
	ud2 := door.UpdateDataDoor()
	if ud2.LockCode == 5 {
		return nil
	}
	if door.ObjOwner != nil {
		return nil
	}
	var found *Object
	for it := u.FirstItem(); it != nil; it = it.NextItem() {
		if !it.Class().Has(object.ClassKey) {
			continue
		}
		tname := s.Types.ByInd(int(it.TypeInd)).ID()
		exp := ""
		switch ud2.LockCode {
		case 1:
			exp = "SilverKey"
		case 2:
			exp = "GoldKey"
		case 3:
			exp = "RubyKey"
		case 4:
			exp = "SapphireKey"
		}
		if exp != "" && tname == exp {
			found = it
			break
		}
	}
	if found == nil && u.Class().Has(object.ClassPlayer) && noxflags.HasGame(noxflags.GameModeQuest) &&
		s.Doors.Sub_4D72C0() && ud2.LockCode == 1 {
		return s.PlayersHaveSilverKey()
	}
	return found
}
