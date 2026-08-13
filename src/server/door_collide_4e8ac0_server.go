package server

import (
	"unsafe"

	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
)

type doorCollideNativeDeps4E8AC0 struct {
	frame              func() uint32
	ticks              func() uint64
	loadFeedbackTicks  func() uint64
	storeFeedbackTicks func(uint64)
	audio              func(uint32, *Object)
	priorityMessage    func(*Object, string)
	keyMessage         func(*Object, string, uint8)
	findKey            func(*Object, *Object) *Object
	questMode          func() bool
	questSync          func(*Object) int32
	storeQuestFrame    func(uint32)
	eachObjectInRect   func(doorCollideRect4E8AC0, DoorTilePoint)
	questKeyState      func() int32
	delayedDelete      func(*Object)
}

// DoorCollideRuntime4E8AC0 supplies state still shared with legacy collision
// handlers. In particular, FeedbackTicks is the same global used by Chest and
// other original handlers rather than a Door-only replacement.
type DoorCollideRuntime4E8AC0 struct {
	Ticks              func() uint64
	LoadFeedbackTicks  func() uint64
	StoreFeedbackTicks func(uint64)
	StoreQuestFrame    func(uint32)
	DelayedDelete      func(*Object)
}

func doorCollideNative4E8AC0(
	door, unit *Object,
	collision unsafe.Pointer,
	deps doorCollideNativeDeps4E8AC0,
) {
	doorCollide4E8AC0(door, unit, collision, doorCollideHooks4E8AC0[*Object, *DoorUpdateData]{
		loadUpdateData: func(obj *Object) *DoorUpdateData {
			return obj.UpdateDataDoor()
		},
		loadCurrentDirection: func(update *DoorUpdateData) int32 {
			return update.CurrentDirection
		},
		loadTargetDirection: func(update *DoorUpdateData) int32 {
			return update.TargetDirection
		},
		loadOwner: func(obj *Object) *Object {
			return obj.ObjOwner
		},
		loadOwnerExpiryFrame: func(obj *Object) uint32 {
			return obj.Field34
		},
		frame: deps.frame,
		storeOwner: func(obj, owner *Object) {
			obj.ObjOwner = owner
		},
		ticks:              deps.ticks,
		loadFeedbackTicks:  deps.loadFeedbackTicks,
		storeFeedbackTicks: deps.storeFeedbackTicks,
		loadSubclassByte: func(obj *Object) uint8 {
			return uint8(obj.ObjSubClass)
		},
		audio:           deps.audio,
		priorityMessage: deps.priorityMessage,
		loadLockCode: func(update *DoorUpdateData) uint8 {
			return update.LockCode
		},
		findKey:    deps.findKey,
		keyMessage: deps.keyMessage,
		loadTileX: func(update *DoorUpdateData) int32 {
			return update.TileX
		},
		loadTileY: func(update *DoorUpdateData) int32 {
			return update.TileY
		},
		storeLockCode: func(update *DoorUpdateData, value uint8) {
			update.LockCode = value
		},
		questMode:        deps.questMode,
		questSync:        deps.questSync,
		storeQuestFrame:  deps.storeQuestFrame,
		eachObjectInRect: deps.eachObjectInRect,
		loadInventoryHolder: func(obj *Object) *Object {
			return obj.InvHolder
		},
		loadClassByte: func(obj *Object) uint8 {
			return uint8(obj.ObjClass)
		},
		questKeyState: deps.questKeyState,
		delayedDelete: deps.delayedDelete,
	})
}

// DoorCollide4E8AC0 binds the original Door collision to native-width Object
// pointers while retaining the fixed-width DoorUpdate record.
func (s *Server) DoorCollide4E8AC0(
	door, unit *Object,
	collision unsafe.Pointer,
	runtime DoorCollideRuntime4E8AC0,
) {
	doorCollideNative4E8AC0(door, unit, collision, doorCollideNativeDeps4E8AC0{
		frame:              s.Frame,
		ticks:              runtime.Ticks,
		loadFeedbackTicks:  runtime.LoadFeedbackTicks,
		storeFeedbackTicks: runtime.StoreFeedbackTicks,
		audio: func(id uint32, obj *Object) {
			s.Audio.EventObj(sound.ID(id), obj, 0, 0)
		},
		priorityMessage: func(unit *Object, message string) {
			s.NetPriMsgToPlayer(unit, strman.ID(message), 0)
		},
		keyMessage: s.doorCollideKeyMessage4E8AC0,
		findKey:    s.DoorCheckKey,
		questMode: func() bool {
			return noxflags.HasGame(noxflags.GameModeQuest)
		},
		questSync:       s.DoorQuestSync4E8390,
		storeQuestFrame: runtime.StoreQuestFrame,
		eachObjectInRect: func(rect doorCollideRect4E8AC0, target DoorTilePoint) {
			s.Map.EachObjInRect(types.Rectf{
				Min: types.Ptf(rect.MinX, rect.MinY),
				Max: types.Ptf(rect.MaxX, rect.MaxY),
			}, func(obj *Object) bool {
				s.DoorCloseAtTile4E8340(obj, &target)
				return true
			})
		},
		questKeyState: func() int32 {
			if s.Doors.Sub_4D72C0() {
				return 1
			}
			return 0
		},
		delayedDelete: runtime.DelayedDelete,
	})
}

func doorCollideKeyPacket4E8AC0(message string, lockCode uint8) ([52]byte, bool) {
	var packet [52]byte
	if message == "" || len(message) > 48 {
		return packet, false
	}
	packet[0] = 0xf0
	packet[1] = 33
	copy(packet[2:51], message)
	packet[51] = lockCode
	return packet, true
}

func (s *Server) doorCollideKeyMessage4E8AC0(unit *Object, message string, lockCode uint8) {
	if unit == nil || uint8(unit.ObjClass)&doorCollidePlayerClassByte4E8AC0 == 0 {
		return
	}
	packet, ok := doorCollideKeyPacket4E8AC0(message, lockCode)
	if !ok {
		return
	}
	player := unit.UpdateDataPlayer().Player
	s.NetSendPacketXxx0(int(player.PlayerInd), packet[:], nil, 1)
}
