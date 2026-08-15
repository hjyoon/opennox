package server

import (
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
)

// SoulGateCollideRuntime4EBE40 supplies the three operations still owned by
// the legacy/root runtime. Native Object, PlayerUpdateData, Player and
// SoulGateCollideData pointers never cross through a fixed-width surrogate.
type SoulGateCollideRuntime4EBE40 struct {
	SetQuestMode  func(int32)
	SetQuestTimer func(uint32)
	PointFX       func(uint32, *types.Pointf) uint32
}

type soulGateCollideNativeDeps4EBE40 struct {
	gameFlagsCheck  func(uint32) uint32
	setQuestMode    func(int32)
	firstPlayerUnit func() *Object
	nextPlayerUnit  func(*Object) *Object
	loadFrame       func() uint32
	setQuestTimer   func(uint32)
	loadFPS         func() uint32
	audio           func(uint32, *Object, int32, int32)
	pointFX         func(uint32, *types.Pointf) uint32
	priorityMessage func(*Object, string, int32)
}

func soulGateCollideNative4EBE40(
	source, target *Object,
	collision *types.Pointf,
	deps soulGateCollideNativeDeps4EBE40,
) {
	soulGateCollide4EBE40(
		source,
		target,
		collision,
		soulGateCollideHooks4EBE40[
			*Object,
			*SoulGateCollideData,
			*PlayerUpdateData,
			*Player,
		]{
			loadSourceCollideData: func(obj *Object) *SoulGateCollideData {
				return (*SoulGateCollideData)(obj.CollideData)
			},
			gameFlagsCheck: deps.gameFlagsCheck,
			loadTargetClassLow: func(obj *Object) uint8 {
				return uint8(obj.ObjClass)
			},
			setQuestMode:    deps.setQuestMode,
			firstPlayerUnit: deps.firstPlayerUnit,
			nextPlayerUnit:  deps.nextPlayerUnit,
			loadPlayerUpdate: func(obj *Object) *PlayerUpdateData {
				return (*PlayerUpdateData)(obj.UpdateData)
			},
			loadPlayer: func(update *PlayerUpdateData) *Player {
				return update.Player
			},
			loadQuestState: func(player *Player) uint32 {
				return player.Field4792
			},
			loadSoulGate: func(update *PlayerUpdateData) *Object {
				return update.SoulGate
			},
			loadFrame:     deps.loadFrame,
			setQuestTimer: deps.setQuestTimer,
			loadLastUsedFrame: func(data *SoulGateCollideData) uint32 {
				return data.LastUsedFrame
			},
			loadFPS: deps.loadFPS,
			audio:   deps.audio,
			pointFX: func(id uint32, obj *Object) uint32 {
				return deps.pointFX(id, &obj.PosVec)
			},
			priorityMessage: deps.priorityMessage,
			storeSoulGate: func(update *PlayerUpdateData, gate *Object) {
				update.SoulGate = gate
			},
			storeLastUsedFrame: func(data *SoulGateCollideData, frame uint32) {
				data.LastUsedFrame = frame
			},
		},
	)
}

func soulGateCollideServerDeps4EBE40(
	s *Server,
	runtime SoulGateCollideRuntime4EBE40,
) soulGateCollideNativeDeps4EBE40 {
	return soulGateCollideNativeDeps4EBE40{
		gameFlagsCheck: func(flag uint32) uint32 {
			if noxflags.HasGame(noxflags.GameFlag(flag)) {
				return 1
			}
			return 0
		},
		setQuestMode:    runtime.SetQuestMode,
		firstPlayerUnit: s.Players.FirstUnit,
		nextPlayerUnit:  s.questNextPlayerUnit4DA7F0,
		loadFrame:       s.Frame,
		setQuestTimer:   runtime.SetQuestTimer,
		loadFPS:         s.TickRate,
		audio: func(id uint32, obj *Object, first, second int32) {
			s.Audio.EventObj(sound.ID(id), obj, int(first), uint32(second))
		},
		pointFX: runtime.PointFX,
		priorityMessage: func(obj *Object, message string, value int32) {
			s.NetPriMsgToPlayer(obj, strman.ID(message), byte(value))
		},
	}
}

// SoulGateCollide4EBE40 binds GAME.EXE 004EBE40 to native-width server
// layouts. The collision point remains in the registered callback signature
// but is deliberately unread, matching the original function.
func (s *Server) SoulGateCollide4EBE40(
	source, target *Object,
	collision *types.Pointf,
	runtime SoulGateCollideRuntime4EBE40,
) {
	soulGateCollideNative4EBE40(
		source,
		target,
		collision,
		soulGateCollideServerDeps4EBE40(s, runtime),
	)
}
