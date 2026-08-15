package server

import (
	"encoding/binary"

	"github.com/opennox/libs/noxnet/netmsg"

	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/internal/netlist"
)

// CrownUpdateData is the native-pointer form of CrownUpdate's original
// three-word data record. GAME.EXE 004F3400 caches this record and clears
// PickupTarget last on every Player path.
type CrownUpdateData struct {
	Field0       *Object
	PickupTarget *Object
	Field2       uint32
}

// CrownPickupRuntime4F3400 supplies the enchant application still owned by
// the legacy spell subsystem. All other 004F3400 services have native Server
// bindings.
type CrownPickupRuntime4F3400 struct {
	ApplyEnchant func(obj *Object, enchant EnchantID, duration, power uint32)
}

type crownPickupNativeDeps4F3400 struct {
	defaultPickup func(*Object, *Object, int32, int32) uint32
	loadFrame     func() uint32
	setOwner      func(*Object, *Object)
	applyEnchant  func(*Object, EnchantID, uint32, uint32)
	playAudio     func(sound.ID, *Object, int32, uint32)
	informPickup  func(uint8, uint32, uint32)
	unmarkMinimap func(*Object, uint32)
}

func crownPickupNative4F3400(
	who, crown *Object,
	flag1, flag2 int32,
	deps crownPickupNativeDeps4F3400,
) uint32 {
	return crownPickup4F3400(
		who,
		crown,
		flag1,
		flag2,
		crownPickupHooks4F3400[*Object, *CrownUpdateData, *PlayerUpdateData]{
			loadCrownUpdate: func(obj *Object) *CrownUpdateData {
				return (*CrownUpdateData)(obj.UpdateData)
			},
			loadClassLow: func(obj *Object) uint8 {
				return uint8(obj.ObjClass)
			},
			defaultPickup: deps.defaultPickup,
			loadPlayerUpdate: func(obj *Object) *PlayerUpdateData {
				return (*PlayerUpdateData)(obj.UpdateData)
			},
			loadFrame: deps.loadFrame,
			storePickupFrame: func(update *PlayerUpdateData, frame uint32) {
				update.Field66 = frame
			},
			setOwner: deps.setOwner,
			applyEnchant: func(obj *Object, enchant, duration, power uint32) {
				deps.applyEnchant(obj, EnchantID(enchant), duration, power)
			},
			playAudio: func(id uint32, obj *Object, kind int32, code uint32) {
				deps.playAudio(sound.ID(id), obj, kind, code)
			},
			loadNetCode: func(obj *Object) uint32 {
				return obj.NetCode
			},
			hasTeam: func(obj *Object) bool {
				return obj.TeamVal.Has()
			},
			loadTeamID: func(obj *Object) uint8 {
				return uint8(obj.TeamVal.ID)
			},
			informPickup:  deps.informPickup,
			unmarkMinimap: deps.unmarkMinimap,
			clearPending: func(update *CrownUpdateData) {
				update.PickupTarget = nil
			},
		},
	)
}

func crownPickupInformPacket4F3400(code uint8, netCode, teamID uint32) [10]byte {
	var packet [10]byte
	packet[0] = byte(netmsg.MSG_INFORM)
	packet[1] = code
	binary.LittleEndian.PutUint32(packet[2:6], netCode)
	binary.LittleEndian.PutUint32(packet[6:10], teamID)
	return packet
}

func (s *Server) crownPickupInformAll4F3400(code uint8, netCode, teamID uint32) {
	packet := crownPickupInformPacket4F3400(code, netCode, teamID)
	for unit := s.Players.FirstUnit(); unit != nil; unit = s.questNextPlayerUnit4DA7F0(unit) {
		player := (*PlayerUpdateData)(unit.UpdateData).Player
		s.NetList.AddToMsgListCli(ntype.PlayerInd(player.PlayerInd), netlist.Kind1, packet[:])
	}
}

func crownPickupServerDeps4F3400(
	s *Server,
	runtime CrownPickupRuntime4F3400,
) crownPickupNativeDeps4F3400 {
	return crownPickupNativeDeps4F3400{
		defaultPickup: func(who, crown *Object, flag1, flag2 int32) uint32 {
			if s.Objs.DefaultPickup(who, crown, int(flag1), int(flag2)) {
				return 1
			}
			return 0
		},
		loadFrame: s.Frame,
		setOwner:  s.ObjSetOwner,
		applyEnchant: func(obj *Object, enchant EnchantID, duration, power uint32) {
			runtime.ApplyEnchant(obj, enchant, duration, power)
		},
		playAudio: func(id sound.ID, obj *Object, kind int32, code uint32) {
			s.Audio.EventObj(id, obj, int(kind), code)
		},
		informPickup: s.crownPickupInformAll4F3400,
		unmarkMinimap: func(obj *Object, flags uint32) {
			s.Players.Nox_xxx_netUnmarkMinimapSpec_417470(obj, flags)
		},
	}
}

// CrownPickup4F3400 binds CrownPickup's four-argument IA-32 callback to
// native-width Object, CrownUpdateData, and PlayerUpdateData layouts.
func (s *Server) CrownPickup4F3400(
	who, crown *Object,
	flag1, flag2 int32,
	runtime CrownPickupRuntime4F3400,
) uint32 {
	return crownPickupNative4F3400(
		who,
		crown,
		flag1,
		flag2,
		crownPickupServerDeps4F3400(s, runtime),
	)
}
