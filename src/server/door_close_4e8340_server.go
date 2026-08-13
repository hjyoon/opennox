package server

import (
	noxflags "github.com/opennox/opennox/v1/common/flags"
)

func doorCloseNative4E8340(
	door *Object,
	target *DoorTilePoint,
	quest func() int32,
	questSync func(*Object) int32,
) {
	doorClose4E8340(door, target, doorCloseHooks4E8340[*Object, *DoorUpdateData, *DoorTilePoint]{
		class: func(obj *Object) uint32 {
			return uint32(obj.ObjClass)
		},
		updateData: func(obj *Object) *DoorUpdateData {
			return obj.UpdateDataDoor()
		},
		tileX: func(update *DoorUpdateData) int32 {
			return update.TileX
		},
		targetX: func(target *DoorTilePoint) int32 {
			return target.X
		},
		tileY: func(update *DoorUpdateData) int32 {
			return update.TileY
		},
		targetY: func(target *DoorTilePoint) int32 {
			return target.Y
		},
		storeLockCode: func(update *DoorUpdateData, value uint8) {
			update.LockCode = value
		},
		quest:     quest,
		questSync: questSync,
	})
}

func doorQuestSyncNative4E8390(
	door *Object,
	sendExtent func(int32, *Object) int32,
) int32 {
	return doorQuestSync4E8390(door, doorQuestSyncHooks4E8390[*Object, *DoorUpdateData]{
		updateData: func(obj *Object) *DoorUpdateData {
			return obj.UpdateDataDoor()
		},
		storeSyncByte: func(update *DoorUpdateData, value uint8) {
			update.QuestSync = value
		},
		sendExtent: sendExtent,
	})
}

func (s *Server) DoorExtentPacket4D6A20(recipient int32, obj *Object) int32 {
	return doorExtentPacket4D6A20(recipient, obj, doorExtentPacketHooks4D6A20[*Object]{
		extent: func(obj *Object) uint16 {
			return uint16(obj.Extent)
		},
		send: func(recipient int32, packet [4]byte, relatedObject uintptr, removeIfDisconnected int32) int32 {
			if relatedObject != 0 {
				panic("004D6A20 related object must be nil")
			}
			return int32(s.NetSendPacketXxx0(int(recipient), packet[:], nil, int(removeIfDisconnected)))
		},
	})
}

func (s *Server) DoorQuestSync4E8390(door *Object) int32 {
	return doorQuestSyncNative4E8390(door, s.DoorExtentPacket4D6A20)
}

func (s *Server) DoorCloseAtTile4E8340(door *Object, target *DoorTilePoint) {
	doorCloseNative4E8340(door, target,
		func() int32 {
			if noxflags.HasGame(noxflags.GameModeQuest) {
				return 1
			}
			return 0
		},
		s.DoorQuestSync4E8390,
	)
}
