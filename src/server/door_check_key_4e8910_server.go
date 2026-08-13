package server

import noxflags "github.com/opennox/opennox/v1/common/flags"

func (s *Server) playersHaveSilverKeyNative4E8A10() *Object {
	return playersHaveSilverKey4E8A10(playersHaveSilverKeyHooks4E8A10[*Object]{
		loadCachedTypeID: func() uint32 {
			return uint32(s.Types.fast.silverKey)
		},
		lookupTypeID: func() uint32 {
			return uint32(uint16(s.Types.IndByID("SilverKey")))
		},
		storeCachedTypeID: func(id uint32) {
			s.Types.fast.silverKey = int(id)
		},
		firstPlayerUnit: s.Players.FirstUnit,
		nextPlayerUnit: func(unit *Object) *Object {
			if unit == nil || uint8(unit.ObjClass)&doorPlayerClassByte4E8910 == 0 {
				return nil
			}
			update := (*PlayerUpdateData)(unit.UpdateData)
			for player := s.Players.Next(update.Player); player != nil; player = s.Players.Next(player) {
				if player.PlayerUnit != nil {
					return player.PlayerUnit
				}
			}
			return nil
		},
		firstItem: func(unit *Object) *Object {
			return unit.InvFirstItem
		},
		nextItem: func(item *Object) *Object {
			return item.InvNextItem
		},
		loadTypeInd: func(item *Object) uint16 {
			return uint16(item.TypeInd)
		},
	})
}

// PlayersHaveSilverKey preserves private GAME.EXE 004E8A10 for existing Go
// callers while keeping object pointers native-width.
func (s *Server) PlayersHaveSilverKey() *Object {
	return s.playersHaveSilverKeyNative4E8A10()
}

func (s *Server) doorCheckKeyNative4E8910(unit, door *Object) *Object {
	return doorCheckKey4E8910(unit, door, doorCheckKeyHooks4E8910[*Object, *DoorUpdateData]{
		loadDoorData: func(obj *Object) *DoorUpdateData {
			return (*DoorUpdateData)(obj.UpdateData)
		},
		loadLockCode: func(data *DoorUpdateData) uint8 {
			return data.LockCode
		},
		loadOwner: func(obj *Object) *Object {
			return obj.ObjOwner
		},
		firstItem: func(obj *Object) *Object {
			return obj.InvFirstItem
		},
		loadClassByte: func(obj *Object) uint8 {
			return uint8(obj.ObjClass)
		},
		loadTypeName: func(obj *Object) string {
			return s.Types.ByInd(int(uint16(obj.TypeInd))).ID()
		},
		nextItem: func(obj *Object) *Object {
			return obj.InvNextItem
		},
		hasQuestGameMode: func() bool {
			return noxflags.HasGame(noxflags.GameModeQuest)
		},
		loadQuestKeyState: func() int32 {
			if s.Doors.Sub_4D72C0() {
				return 1
			}
			return 0
		},
		playersHaveSilverKey: s.playersHaveSilverKeyNative4E8A10,
	})
}

// DoorCheckKey preserves GAME.EXE 004E8910 for the Door collision and pathing
// callers. Object, inventory, owner and update-data links retain native width.
func (s *Server) DoorCheckKey(unit, door *Object) *Object {
	return s.doorCheckKeyNative4E8910(unit, door)
}
