package server

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/strman"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// ExitCollideData is the pointer-independent 88-byte destination record
// registered by ExitCollide. GAME.EXE stores an 80-byte C-string map name
// followed by the destination X and Y coordinates.
type ExitCollideData struct {
	MapName      [80]byte
	DestinationX float32
	DestinationY float32
}

type exitCollideNativeDeps4E9090 struct {
	warpEnabled          func() int32
	saveBusy             func() int32
	exitAllowed          func(*Object) int32
	setMapLoadRequired   func(int32)
	setSaveFileName      func(string)
	saveCoop             func(int32, *Object, int32)
	questMapFile         func() unsafe.Pointer
	disableAbility       func(*Object, Ability)
	currentQuestStage    func() uint32
	nextStageThreshold   func(uint32) uint32
	sendQuestStage       func(uint8, uint32)
	setPlayerState       func(*Object, PlayerState)
	goObserver           func(*Player, int32, int32)
	storeWarpFrame       func(uint32)
	storeNextMap         func(string)
	setCurrentQuestStage func(uint32)
	setQuestWarping      func(int32)
	mapLoad              func(string)
	countdown            QuestExitCountdownRuntime4E8E60
	delayedDelete        func(*Object)
}

// ExitCollideRuntime4E9090 supplies the remaining legacy-owned save, Quest,
// timer, observer, and map-switch state through native-width callbacks.
type ExitCollideRuntime4E9090 struct {
	WarpEnabled          func() int32
	SaveBusy             func() int32
	ExitAllowed          func(*Object) int32
	SetMapLoadRequired   func(int32)
	SetSaveFileName      func(string)
	SaveCoop             func(int32, *Object, int32)
	QuestMapFile         func() unsafe.Pointer
	DisableAbility       func(*Object, Ability)
	CurrentQuestStage    func() uint32
	NextStageThreshold   func(uint32) uint32
	SendQuestStage       func(uint8, uint32)
	SetPlayerState       func(*Object, PlayerState)
	GoObserver           func(*Player, int32, int32)
	StoreWarpFrame       func(uint32)
	StoreNextMap         func(string)
	SetCurrentQuestStage func(uint32)
	SetQuestWarping      func(int32)
	MapLoad              func(string)
	Countdown            QuestExitCountdownRuntime4E8E60
	DelayedDelete        func(*Object)
}

func exitCollideMapString4E9090(ptr unsafe.Pointer) string {
	_ = *(*byte)(ptr)
	return alloc.GoString((*byte)(ptr))
}

func exitCollideCurTrapsByte4E9090(update *PlayerUpdateData) uint8 {
	return uint8(update.CurTraps)
}

func exitCollideStoreCurTrapsByte4E9090(update *PlayerUpdateData, value uint8) {
	update.CurTraps = update.CurTraps&^uint32(0xff) | uint32(value)
}

func exitCollideUnitPacket4E9090(code uint8, netCode uint32) [6]byte {
	var packet [6]byte
	packet[0] = byte(netmsg.MSG_INFORM)
	packet[1] = code
	binary.LittleEndian.PutUint32(packet[2:], netCode)
	return packet
}

func (s *Server) exitCollideBroadcastUnit4E9090(code uint8, subject *Object) {
	packet := exitCollideUnitPacket4E9090(code, subject.NetCode)
	for unit := s.Players.FirstUnit(); unit != nil; unit = s.questNextPlayerUnit4DA7F0(unit) {
		player := (*PlayerUpdateData)(unit.UpdateData).Player
		s.NetList.AddToMsgListCli(ntype.PlayerInd(player.PlayerInd), netlist.Kind1, packet[:])
	}
}

func (s *Server) exitCollideRecordProgress4D60E0(unit *Object) {
	if uint8(unit.ObjFlags)&0x20 != 0 {
		return
	}
	update := (*PlayerUpdateData)(unit.UpdateData)
	player := update.Player
	if player.Field4792 == 1 {
		player.field4652++
		player = update.Player
		player.field4692 |= 1
	}
}

func (s *Server) exitCollideResetQuestPlayer4D6000(unit *Object, currentQuestStage func() uint32) {
	if unit == nil {
		return
	}
	update := (*PlayerUpdateData)(unit.UpdateData)
	update.Player.field4652 = 0
	update.Player.field4656 = 0
	update.Player.field4660 = 0
	update.Player.field4664 = 0
	update.Player.field4668 = 0
	update.Player.field4672 = 0
	update.Player.field4676 = 0
	update.Player.field4680 = 0
	update.Player.field4684 = 0
	stage := currentQuestStage()
	update.Player.field4688 = stage
	update.Player.field4692 = 63
}

func (s *Server) exitCollideResetQuestPlayers4D60B0(currentQuestStage func() uint32) {
	for unit := s.Players.FirstUnit(); unit != nil; unit = s.questNextPlayerUnit4DA7F0(unit) {
		s.exitCollideResetQuestPlayer4D6000(unit, currentQuestStage)
	}
}

func exitCollideNative4E9090(
	s *Server,
	exit, unit *Object,
	collision unsafe.Pointer,
	deps exitCollideNativeDeps4E9090,
) {
	exitCollide4E9090(exit, unit, collision, exitCollideHooks4E9090[
		*Object, *PlayerUpdateData, *Player, unsafe.Pointer, unsafe.Pointer,
	]{
		glyphType: func() uint32 {
			return uint32(s.Types.GlyphID())
		},
		loadClassByte: func(obj *Object) uint8 {
			return uint8(obj.ObjClass)
		},
		defaultCollide: func(*Object, *Object, unsafe.Pointer) {},
		loadUpdateData: func(obj *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(obj.UpdateData)
		},
		loadMap: func(obj *Object) unsafe.Pointer {
			return obj.CollideData
		},
		gameFlag: func(flag uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(flag)) {
				return 1
			}
			return 0
		},
		loadQuestExit: func(update *PlayerUpdateData) *Object {
			return update.QuestExit
		},
		loadQuestWarpGate: func(update *PlayerUpdateData) *Object {
			return update.QuestWarpGate
		},
		loadSubclassByte: func(obj *Object) uint8 {
			return uint8(obj.ObjSubClass)
		},
		warpEnabled: deps.warpEnabled,
		saveBusy:    deps.saveBusy,
		exitAllowed: deps.exitAllowed,
		paused: func() int32 {
			if noxflags.HasGame(noxflags.GamePause) {
				return 1
			}
			return 0
		},
		mapFirstByte: func(ptr unsafe.Pointer) byte {
			return *(*byte)(ptr)
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadPlayerClass: func(player *Player) uint8 {
			return player.info[66]
		},
		firstOwned: func(obj *Object) *Object {
			return obj.Field129
		},
		nextOwned: func(obj *Object) *Object {
			return obj.Field128
		},
		loadTypeIndex: func(obj *Object) uint16 {
			return obj.TypeInd
		},
		loadInventoryHolder: func(obj *Object) *Object {
			return obj.InvHolder
		},
		delayedDelete: deps.delayedDelete,
		loadCurTrapsByte: func(update *PlayerUpdateData) uint8 {
			return exitCollideCurTrapsByte4E9090(update)
		},
		storeCurTrapsByte: func(update *PlayerUpdateData, value uint8) {
			exitCollideStoreCurTrapsByte4E9090(update, value)
		},
		setMapLoadRequired: deps.setMapLoadRequired,
		setSaveFileName:    deps.setSaveFileName,
		saveCoop:           deps.saveCoop,
		questMapFile:       deps.questMapFile,
		abilityActive: func(obj *Object, ability uint32) int32 {
			if s.Abils.IsActive(obj, Ability(ability)) {
				return 1
			}
			return 0
		},
		disableAbility: func(obj *Object, ability uint32) {
			deps.disableAbility(obj, Ability(ability))
		},
		currentQuestStage: deps.currentQuestStage,
		recordExitProgress: func(obj *Object) {
			s.exitCollideRecordProgress4D60E0(obj)
		},
		loadQuestStage: func(player *Player) uint32 {
			return player.QuestStage
		},
		storeQuestStage: func(player *Player, stage uint32) {
			player.QuestStage = stage
		},
		loadPlayerIndex: func(player *Player) uint8 {
			return player.PlayerInd
		},
		sendQuestStage: deps.sendQuestStage,
		storeQuestExit: func(update *PlayerUpdateData, obj *Object) {
			update.QuestExit = obj
		},
		storeQuestWarpGate: func(update *PlayerUpdateData, obj *Object) {
			update.QuestWarpGate = obj
		},
		setPlayerState: func(obj *Object, state int32) {
			deps.setPlayerState(obj, PlayerState(state))
		},
		goObserver:           deps.goObserver,
		broadcastUnitMessage: s.exitCollideBroadcastUnit4E9090,
		allPlayersExited:     s.QuestAllPlayersExited4E9010,
		frame:                s.Frame,
		storeWarpFrame:       deps.storeWarpFrame,
		firstPlayerUnit:      s.Players.FirstUnit,
		nextPlayerUnit:       s.questNextPlayerUnit4DA7F0,
		sendUnitMessage: func(code uint8, recipient, subject *Object) {
			_ = subject
			player := (*PlayerUpdateData)(recipient.UpdateData).Player
			_ = s.NetInformTextMsg(ntype.PlayerInd(player.PlayerInd), code, 0)
		},
		priorityMessage: func(obj *Object, message string, value uint8) {
			s.NetPriMsgToPlayer(obj, strman.ID(message), value)
		},
		maybeWarp: func() int32 {
			return s.QuestMaybeWarp4E8F60(QuestMaybeWarpRuntime4E8F60{
				CurrentQuestStage:  deps.currentQuestStage,
				NextStageThreshold: deps.nextStageThreshold,
			})
		},
		audio: func(id uint32, obj *Object, a3, a4 int32) {
			s.Audio.EventObj(sound.ID(id), obj, int(a3), uint32(a4))
		},
		exitCountdown: func() int32 {
			return s.QuestExitCountdown4E8E60(deps.countdown)
		},
		copyNextMap: func(ptr unsafe.Pointer) {
			deps.storeNextMap(exitCollideMapString4E9090(ptr))
		},
		nextStageThreshold:   deps.nextStageThreshold,
		setCurrentQuestStage: deps.setCurrentQuestStage,
		setQuestWarping:      deps.setQuestWarping,
		resetQuestPlayers: func() {
			s.exitCollideResetQuestPlayers4D60B0(deps.currentQuestStage)
		},
		mapLoad: func(ptr unsafe.Pointer) {
			deps.mapLoad(exitCollideMapString4E9090(ptr))
		},
	})
}

// ExitCollide4E9090 binds the original callback to native Object,
// PlayerUpdateData, Player, collide-data, and collision pointers.
func (s *Server) ExitCollide4E9090(
	exit, unit *Object,
	collision unsafe.Pointer,
	runtime ExitCollideRuntime4E9090,
) {
	exitCollideNative4E9090(s, exit, unit, collision, exitCollideNativeDeps4E9090{
		warpEnabled:          runtime.WarpEnabled,
		saveBusy:             runtime.SaveBusy,
		exitAllowed:          runtime.ExitAllowed,
		setMapLoadRequired:   runtime.SetMapLoadRequired,
		setSaveFileName:      runtime.SetSaveFileName,
		saveCoop:             runtime.SaveCoop,
		questMapFile:         runtime.QuestMapFile,
		disableAbility:       runtime.DisableAbility,
		currentQuestStage:    runtime.CurrentQuestStage,
		nextStageThreshold:   runtime.NextStageThreshold,
		sendQuestStage:       runtime.SendQuestStage,
		setPlayerState:       runtime.SetPlayerState,
		goObserver:           runtime.GoObserver,
		storeWarpFrame:       runtime.StoreWarpFrame,
		storeNextMap:         runtime.StoreNextMap,
		setCurrentQuestStage: runtime.SetCurrentQuestStage,
		setQuestWarping:      runtime.SetQuestWarping,
		mapLoad:              runtime.MapLoad,
		countdown:            runtime.Countdown,
		delayedDelete:        runtime.DelayedDelete,
	})
}

var (
	_ = [1]struct{}{}[88-unsafe.Sizeof(ExitCollideData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(ExitCollideData{}.MapName)]
	_ = [1]struct{}{}[80-unsafe.Offsetof(ExitCollideData{}.DestinationX)]
	_ = [1]struct{}{}[84-unsafe.Offsetof(ExitCollideData{}.DestinationY)]
)
