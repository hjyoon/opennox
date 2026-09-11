package server

import (
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

// RemotePlayerAudioRuntime501CA0 supplies the services still owned by the root
// game package. All pointer-bearing arguments remain native-width.
type RemotePlayerAudioRuntime501CA0 struct {
	QuestCheckSecretArea func(*Object)
	PolygonAtPoint       func([2]int32, uint32) unsafe.Pointer
	PolygonZone          func(unsafe.Pointer) uint8
	EventZone            func(*types.Pointf, *Object) uint8
	Fade                 func(sound.ID, *types.Pointf, *types.Pointf) int32
	SendDirect           func(*Object, *AudioEvent, int32)
	Flush                func(*Object)
}

// UpdateRemotePlayerAudio501CA0 binds the verified 00501CA0 control flow to
// native Object, PlayerUpdateData, Player, Team, AudioEvent, and Pointf layouts.
func (s *Server) UpdateRemotePlayerAudio501CA0(unit *Object, runtime RemotePlayerAudioRuntime501CA0) {
	remotePlayerAudioUpdate501CA0(unit, remotePlayerAudioHooks501CA0[
		*Object,
		unsafe.Pointer,
		*Player,
		unsafe.Pointer,
		*Team,
		*AudioEvent,
		*types.Pointf,
	]{
		loadUpdate: func(object *Object) unsafe.Pointer {
			return object.UpdateData
		},
		loadPlayer: func(update unsafe.Pointer) *Player {
			return (*PlayerUpdateData)(update).Player
		},
		loadPlayerFlagsLow: func(player *Player) uint8 {
			return uint8(player.Field3680)
		},
		loadCameraTarget: func(player *Player) *Object {
			return player.CameraFollowObj
		},
		loadClassLow: func(object *Object) uint8 {
			return uint8(object.ObjClass)
		},
		loadObjectPositionX: func(object *Object) float32 {
			return object.PosVec.X
		},
		floatToInt: audioEventZoneFloatToInt501C00,
		loadObjectPositionY: func(object *Object) float32 {
			return object.PosVec.Y
		},
		polygonAtPoint:  runtime.PolygonAtPoint,
		loadPolygonZone: runtime.PolygonZone,
		loadCurrentPolygonID: func(player *Player) uint32 {
			return player.field3664
		},
		questCheckSecretArea: runtime.QuestCheckSecretArea,
		loadPlayerAudioZone: func(player *Player) uint8 {
			return uint8(player.field3668)
		},
		resetBitmap: s.Audio.ResetBitmap,
		hasTeam: func(object *Object) bool {
			return object.TeamVal.Has()
		},
		loadTeamID: func(object *Object) uint32 {
			return uint32(object.TeamVal.ID)
		},
		teamByID: func(id uint32) *Team {
			return s.Teams.ByID(TeamID(id))
		},
		firstEvent: func() *AudioEvent {
			return s.Audio.head
		},
		loadEventKind: func(event *AudioEvent) int32 {
			return int32(event.Kind)
		},
		loadEventCode: func(event *AudioEvent) uint32 {
			return event.Code
		},
		loadObjectNetCode: func(object *Object) uint32 {
			return object.NetCode
		},
		loadEventObject: func(event *AudioEvent) *Object {
			return event.Obj
		},
		eventPosition: func(event *AudioEvent) *types.Pointf {
			return &event.Pos
		},
		eventZone: runtime.EventZone,
		loadPhonemeState: func(update unsafe.Pointer) uint8 {
			return (*PlayerUpdateData)(update).Field47_0
		},
		loadEventSound: func(event *AudioEvent) int32 {
			return int32(event.Sound)
		},
		playerPosition: func(player *Player) *types.Pointf {
			return &player.Pos3632Vec
		},
		fade: func(id int32, eventPosition, listenerPosition *types.Pointf) int32 {
			return runtime.Fade(sound.ID(id), eventPosition, listenerPosition)
		},
		loadSoundField20: func(id int32) int32 {
			return int32(s.Audio.Field20(sound.ID(id)))
		},
		addAudio: func(event *AudioEvent, fade int32) {
			s.Audio.AddAudio(event, int(fade))
		},
		sendDirect: runtime.SendDirect,
		loadEventNext: func(event *AudioEvent) *AudioEvent {
			return event.next0
		},
		flush: runtime.Flush,
	})
}
