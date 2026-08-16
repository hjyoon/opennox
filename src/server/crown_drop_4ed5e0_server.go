package server

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/netlist"
)

// CrownDropRuntime4ED5E0 supplies the three services still owned by legacy C.
// Object, CrownUpdateData, PlayerUpdateData, player iteration, owner removal,
// network delivery, and minimap tracking remain native-width in server.
type CrownDropRuntime4ED5E0 struct {
	DefaultDrop  func(*Object, *Object, *types.Pointf) int32
	TeamContains func(*ObjectTeam, TeamID) int32
	BuffOff      func(*Object, EnchantID)
}

type crownDropNativeDeps4ED5E0 struct {
	gameFlag       func(uint32) int32
	gameplayFlag   func(uint32) int32
	loadFrame      func() uint32
	firstPlayer    func() *Object
	nextPlayer     func(*Object) *Object
	defaultDrop    func(*Object, *Object, *types.Pointf) int32
	teamContains   func(*ObjectTeam, TeamID) int32
	clearOwner     func(*Object)
	buffOff        func(*Object, EnchantID)
	informDrop     func(uint8, uint32, uint32)
	markMinimapAll func(*Object, uint32)
}

func crownDropNative4ED5E0(
	owner, crown *Object,
	point *types.Pointf,
	deps crownDropNativeDeps4ED5E0,
) int32 {
	return crownDrop4ED5E0(crownDropHooks4ED5E0[
		*Object,
		*CrownUpdateData,
		*PlayerUpdateData,
		*types.Pointf,
	]{
		gameFlag:     deps.gameFlag,
		gameplayFlag: deps.gameplayFlag,
		loadCrownArg: func() *Object {
			return crown
		},
		loadFrame: deps.loadFrame,
		loadTeamID: func(obj *Object) uint8 {
			return uint8(obj.TeamVal.ID)
		},
		loadCrownUpdate: func(obj *Object) *CrownUpdateData {
			return (*CrownUpdateData)(obj.UpdateData)
		},
		firstPlayer: deps.firstPlayer,
		loadPlayerData: func(obj *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(obj.UpdateData)
		},
		teamContains: func(obj *Object, teamID uint8) int32 {
			return deps.teamContains(&obj.TeamVal, TeamID(teamID))
		},
		loadPickupFrame: func(data *PlayerUpdateData) uint32 {
			return data.Field66
		},
		nextPlayer: deps.nextPlayer,
		loadOwnerArg: func() *Object {
			return owner
		},
		storePickupTarget: func(data *CrownUpdateData, target *Object) {
			data.PickupTarget = target
		},
		loadPointArg: func() *types.Pointf {
			return point
		},
		defaultDrop: deps.defaultDrop,
		clearOwner:  deps.clearOwner,
		buffOff: func(obj *Object, enchant uint32) {
			deps.buffOff(obj, EnchantID(enchant))
		},
		loadNetCode: func(obj *Object) uint32 {
			return obj.NetCode
		},
		hasTeam: func(obj *Object) int32 {
			if obj.TeamVal.Has() {
				return 1
			}
			return 0
		},
		informDrop:  deps.informDrop,
		markMinimap: deps.markMinimapAll,
	})
}

func crownDropInformPacket4ED5E0(code uint8, netCode, teamID uint32) [10]byte {
	var packet [10]byte
	packet[0] = byte(netmsg.MSG_INFORM)
	packet[1] = code
	binary.LittleEndian.PutUint32(packet[2:6], netCode)
	binary.LittleEndian.PutUint32(packet[6:10], teamID)
	return packet
}

func (s *Server) crownDropInformAll4ED5E0(code uint8, netCode, teamID uint32) {
	packet := crownDropInformPacket4ED5E0(code, netCode, teamID)
	for unit := s.Players.FirstUnit(); unit != nil; unit = s.questNextPlayerUnit4DA7F0(unit) {
		player := (*PlayerUpdateData)(unit.UpdateData).Player
		s.NetList.AddToMsgListCli(ntype.PlayerInd(player.PlayerInd), netlist.Kind1, packet[:])
	}
}

func crownDropServerDeps4ED5E0(
	s *Server,
	runtime CrownDropRuntime4ED5E0,
) crownDropNativeDeps4ED5E0 {
	return crownDropNativeDeps4ED5E0{
		gameFlag: func(flag uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(flag)) {
				return 1
			}
			return 0
		},
		gameplayFlag: func(flag uint32) int32 {
			if noxflags.HasGamePlay(noxflags.GameplayFlag(flag)) {
				return 1
			}
			return 0
		},
		loadFrame:    s.Frame,
		firstPlayer:  s.Players.FirstUnit,
		nextPlayer:   s.questNextPlayerUnit4DA7F0,
		defaultDrop:  runtime.DefaultDrop,
		teamContains: runtime.TeamContains,
		clearOwner:   s.ObjClearOwner,
		buffOff:      runtime.BuffOff,
		informDrop:   s.crownDropInformAll4ED5E0,
		markMinimapAll: func(obj *Object, flags uint32) {
			s.defaultDropMarkMinimapAll4ED290(obj, flags)
		},
	}
}

// CrownDrop4ED5E0 binds GAME.EXE 004ED5E0 to native-width Object, Crown
// update, Player update, and Pointf pointers while retaining the remaining
// legacy services as typed dependencies.
func (s *Server) CrownDrop4ED5E0(
	owner, crown *Object,
	point *types.Pointf,
	runtime CrownDropRuntime4ED5E0,
) int32 {
	return crownDropNative4ED5E0(owner, crown, point, crownDropServerDeps4ED5E0(s, runtime))
}

var (
	_ = [1]struct{}{}[4-unsafe.Offsetof(ObjectTeam{}.ID)]
)
