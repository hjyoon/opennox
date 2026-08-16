package server

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/internal/netlist"
)

const (
	defaultDropCacheBase4ED290          = uintptr(0x5D4594)
	defaultDropLanternCacheOffset4ED290 = uintptr(1568244)
	defaultDropGlyphCacheOffset4ED290   = uintptr(1568252)
	defaultDropTorchCacheOffset4ED290   = uintptr(1568256)
)

// DefaultDropEquipUpdatePrefix4ED290 is the pointer-free three-byte prefix
// read for the replacement Quiver test. Only byte +2 is observed by 004ED290.
type DefaultDropEquipUpdatePrefix4ED290 struct {
	Field0 uint16
	Field2 uint8
}

var _ = [1]struct{}{}[2-unsafe.Offsetof(DefaultDropEquipUpdatePrefix4ED290{}.Field2)]

// DefaultDropRuntime4ED290 contains services not yet owned by server. Object,
// Flag update-data, and Monster update-data access stays native-width; no
// pointer crosses this boundary through an ABI32 integer.
type DefaultDropRuntime4ED290 struct {
	ItemIsDroppable func(*Object) int32
	ItemDropMask    func(*Object, uint32) int32
	DetachInventory func(*Object, *Object)
	CreateAt        func(*Object, *Object, types.Pointf)
	DelayedDelete   func(*Object)
	TeamFlagStatus  func(uint8, uint8, uint8, uint16) int32
	GameFlag        func(uint32) uint32
	SetDecayTime    func(*Object, uint32)
	BuffOff         func(*Object, uint32)
}

type defaultDropNativeDeps4ED290 struct {
	itemIsDroppable   func(*Object) int32
	itemDropMask      func(*Object, uint32) int32
	primaryMessage    func(*Object, string, uint8)
	playAudio         func(uint32, *Object, int32, uint32)
	detachInventory   func(*Object, *Object)
	createAt          func(*Object, *Object, types.Pointf)
	weaponFlags       func(*Object) uint32
	delayedDelete     func(*Object)
	materialIndex     func(*Object) uint32
	informFlagDrop    func(uint8, uint32, uint32)
	markMinimap       func(*Object, uint32)
	loadFrame         func() uint32
	teamFlagStatus    func(uint8, uint8, uint8, uint16) int32
	loadGlyphCache    func() uint32
	storeGlyphCache   func(uint32)
	loadTorchCache    func() uint32
	storeTorchCache   func(uint32)
	loadLanternCache  func() uint32
	storeLanternCache func(uint32)
	lookupType        func(string) uint32
	gameFlag          func(uint32) uint32
	loadGameFPS       func() uint32
	setDecayTime      func(*Object, uint32)
	raise             func(*Object, float32)
	buffOff           func(*Object, uint32)
}

func defaultDropNative4ED290(
	owner, item *Object,
	point *types.Pointf,
	deps defaultDropNativeDeps4ED290,
) int32 {
	return defaultDrop4ED290(defaultDropHooks4ED290[*Object, *types.Pointf, unsafe.Pointer]{
		loadOwnerArg: func() *Object {
			return owner
		},
		loadItemArg: func() *Object {
			return item
		},
		loadInventoryOwner: func(obj *Object) *Object {
			return obj.InvHolder
		},
		loadObjectClassByte: func(obj *Object) uint8 {
			return uint8(obj.ObjClass)
		},
		loadObjectClass: func(obj *Object) uint32 {
			return uint32(obj.ObjClass)
		},
		loadObjectFlags: func(obj *Object) uint32 {
			return uint32(obj.ObjFlags)
		},
		loadObjectNetCode: func(obj *Object) uint32 {
			return obj.NetCode
		},
		loadObjectTeamID: func(obj *Object) uint8 {
			return uint8(obj.TeamVal.ID)
		},
		loadObjectType: func(obj *Object) uint16 {
			return obj.TypeInd
		},
		loadObjectUpdate: func(obj *Object) unsafe.Pointer {
			return obj.UpdateData
		},
		itemIsDroppable: deps.itemIsDroppable,
		itemDropMask:    deps.itemDropMask,
		primaryMessage:  deps.primaryMessage,
		audio:           deps.playAudio,
		detachInventory: deps.detachInventory,
		loadPointArg: func() *types.Pointf {
			return point
		},
		loadPointY: func(point *types.Pointf) float32 {
			return point.Y
		},
		loadPointX: func(point *types.Pointf) float32 {
			return point.X
		},
		createAt: func(item, owner *Object, x, y float32, _ uint32) {
			deps.createAt(item, owner, types.Pointf{X: x, Y: y})
		},
		weaponEquipFlags: deps.weaponFlags,
		loadInventoryHead: func(owner *Object) *Object {
			return owner.InvFirstItem
		},
		loadInventoryNext: func(item *Object) *Object {
			return item.InvNextItem
		},
		loadUpdateByte2: func(update unsafe.Pointer) uint8 {
			return (*DefaultDropEquipUpdatePrefix4ED290)(update).Field2
		},
		delayedDelete:     deps.delayedDelete,
		materialIndex:     deps.materialIndex,
		netInformFlagDrop: deps.informFlagDrop,
		markMinimapForAll: deps.markMinimap,
		loadFrame:         deps.loadFrame,
		storeUpdateFrame: func(update unsafe.Pointer, frame uint32) {
			(*FlagUpdateData4EA490)(update).State = frame
		},
		setTeamFlagStatus: func(teamID, status, material uint8, carrier uint16) {
			deps.teamFlagStatus(teamID, status, material, carrier)
		},
		loadMonsterStatus: func(update unsafe.Pointer) uint32 {
			return uint32((*MonsterUpdateData)(update).StatusFlags)
		},
		storeMonsterAction: func(update unsafe.Pointer, action uint32) {
			(*MonsterUpdateData)(update).AIAction340 = action
		},
		storeMonsterStatus: func(update unsafe.Pointer, status uint32) {
			(*MonsterUpdateData)(update).StatusFlags = object.MonsterStatus(status)
		},
		loadGlyphCache:    deps.loadGlyphCache,
		storeGlyphCache:   deps.storeGlyphCache,
		loadTorchCache:    deps.loadTorchCache,
		storeTorchCache:   deps.storeTorchCache,
		loadLanternCache:  deps.loadLanternCache,
		storeLanternCache: deps.storeLanternCache,
		lookupType:        deps.lookupType,
		gameFlag:          deps.gameFlag,
		loadGameFPS:       deps.loadGameFPS,
		setDecayTime:      deps.setDecayTime,
		raise:             deps.raise,
		buffOff:           deps.buffOff,
	})
}

func defaultDropFlagPacket4ED290(code uint8, netCode, material uint32) [10]byte {
	var packet [10]byte
	packet[0] = byte(netmsg.MSG_INFORM)
	packet[1] = code
	binary.LittleEndian.PutUint32(packet[2:6], netCode)
	binary.LittleEndian.PutUint32(packet[6:10], material)
	return packet
}

func (s *Server) defaultDropInformAll4ED290(code uint8, netCode, material uint32) {
	packet := defaultDropFlagPacket4ED290(code, netCode, material)
	for unit := s.Players.FirstUnit(); unit != nil; unit = s.questNextPlayerUnit4DA7F0(unit) {
		player := (*PlayerUpdateData)(unit.UpdateData).Player
		s.NetList.AddToMsgListCli(ntype.PlayerInd(player.PlayerInd), netlist.Kind1, packet[:])
	}
}

func (s *Server) defaultDropMarkMinimapAll4ED290(obj *Object, flags uint32) {
	for player := s.Players.First(); player != nil; player = s.Players.Next(player) {
		s.Players.Nox_xxx_netMarkMinimapObject_417190(player.PlayerIndex(), obj, flags)
	}
}

func defaultDropLoadCache4ED290(offset uintptr) uint32 {
	return *memmap.PtrUint32(defaultDropCacheBase4ED290, offset)
}

func defaultDropStoreCache4ED290(offset uintptr, value uint32) {
	*memmap.PtrUint32(defaultDropCacheBase4ED290, offset) = value
}

func defaultDropServerDeps4ED290(
	s *Server,
	runtime DefaultDropRuntime4ED290,
) defaultDropNativeDeps4ED290 {
	return defaultDropNativeDeps4ED290{
		itemIsDroppable: runtime.ItemIsDroppable,
		itemDropMask:    runtime.ItemDropMask,
		primaryMessage: func(obj *Object, message string, value uint8) {
			s.NetPriMsgToPlayer(obj, strman.ID(message), value)
		},
		playAudio: func(id uint32, obj *Object, kind int32, code uint32) {
			s.Audio.EventObj(sound.ID(id), obj, int(kind), code)
		},
		detachInventory: runtime.DetachInventory,
		createAt:        runtime.CreateAt,
		weaponFlags:     s.Weapons.Nox_xxx_weaponInventoryEquipFlags_415820,
		delayedDelete:   runtime.DelayedDelete,
		materialIndex:   TeamMaterialObjectIndex4ECBD0,
		informFlagDrop:  s.defaultDropInformAll4ED290,
		markMinimap:     s.defaultDropMarkMinimapAll4ED290,
		loadFrame:       s.Frame,
		teamFlagStatus:  runtime.TeamFlagStatus,
		loadGlyphCache: func() uint32 {
			return defaultDropLoadCache4ED290(defaultDropGlyphCacheOffset4ED290)
		},
		storeGlyphCache: func(value uint32) {
			defaultDropStoreCache4ED290(defaultDropGlyphCacheOffset4ED290, value)
		},
		loadTorchCache: func() uint32 {
			return defaultDropLoadCache4ED290(defaultDropTorchCacheOffset4ED290)
		},
		storeTorchCache: func(value uint32) {
			defaultDropStoreCache4ED290(defaultDropTorchCacheOffset4ED290, value)
		},
		loadLanternCache: func() uint32 {
			return defaultDropLoadCache4ED290(defaultDropLanternCacheOffset4ED290)
		},
		storeLanternCache: func(value uint32) {
			defaultDropStoreCache4ED290(defaultDropLanternCacheOffset4ED290, value)
		},
		lookupType: func(name string) uint32 {
			return uint32(s.Types.IndByID(name))
		},
		gameFlag:     runtime.GameFlag,
		loadGameFPS:  s.TickRate,
		setDecayTime: runtime.SetDecayTime,
		raise: func(obj *Object, z float32) {
			obj.Raise(z)
		},
		buffOff: runtime.BuffOff,
	}
}

// DefaultDrop4ED290 runs the native-width DefaultDrop handler while retaining
// the original callback order and shared GAME.EXE type caches.
func (s *Server) DefaultDrop4ED290(
	owner, item *Object,
	point *types.Pointf,
	runtime DefaultDropRuntime4ED290,
) int32 {
	return defaultDropNative4ED290(owner, item, point, defaultDropServerDeps4ED290(s, runtime))
}
