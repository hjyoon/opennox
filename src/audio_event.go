package opennox

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

func (s *Server) remotePlayerAudioRuntimeNative501CA0() server.RemotePlayerAudioRuntime501CA0 {
	polygonRuntime := audioEventZoneRuntimeNative501C00()
	return server.RemotePlayerAudioRuntime501CA0{
		QuestCheckSecretArea: s.questCheckSecretAreaNative421C70,
		PolygonAtPoint:       polygonRuntime.PolygonAtPoint,
		PolygonZone:          polygonRuntime.PolygonZone,
		EventZone:            s.audioEventZonePtrNative501C00,
		Fade: func(id sound.ID, eventPosition, listenerPosition *types.Pointf) int32 {
			return int32(s.ai.soundFadePerc(id, *eventPosition, *listenerPosition))
		},
		SendDirect: func(listener *server.Object, event *server.AudioEvent, fade int32) {
			s.netSendAudioEvent(listener, event, fade)
		},
		Flush: s.netSendAudioEvents,
	}
}

func (s *Server) netUpdateRemotePlayerNative501CA0(unit *server.Object) {
	s.Server.UpdateRemotePlayerAudio501CA0(unit, s.remotePlayerAudioRuntimeNative501CA0())
}

func (s *Server) netSendAudioEvents(obj *server.Object) {
	s.Audio.EachEventBitmap(func(it *server.AudioEvent) {
		s.netSendAudioEvent(obj, it, int32(it.Perc))
	})
}
