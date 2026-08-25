package opennox

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

func (s *Server) netUpdateRemotePlrAudioEventsNative(obj *server.Object, update *server.PlayerUpdateData, zone byte) {
	s.Audio.ResetBitmap()
	var tm *server.Team
	if obj.HasTeam() {
		tm = s.Teams.ByID(obj.TeamVal.ID)
	}
	s.Audio.EachEvent(func(it *server.AudioEvent) {
		if it.Kind == 1 {
			tm2 := s.Teams.ByID(server.TeamID(it.Code))
			if tm == nil || tm2 == nil || tm != tm2 {
				return
			}
		} else if it.Kind == 2 {
			if obj.NetCode != it.Code {
				return
			}
		}
		eventZone := s.audioEventZoneNative501C00(it.Pos, it.Obj)
		if eventZone == zone || eventZone == 0 {
			if update.Field47_0 != 0 || it.Sound < sound.SoundSpellPhonemeUp || it.Sound > sound.SoundSpellPhonemeUpLeft || obj != it.Obj {
				fade := s.ai.soundFadePerc(it.Sound, it.Pos, update.Player.Pos3632()) / 2
				if fade > 0 {
					if s.Audio.Field20(it.Sound) != 0 {
						s.Audio.AddAudio(it, fade)
					} else {
						s.netSendAudioEvent(obj, it, int16(fade))
					}
				}
			}
		}
	})
	s.netSendAudioEvents(obj)
}

// NetUpdateRemotePlrAudioEvents remains as the legacy C callback boundary.
// Native callers use netUpdateRemotePlrAudioEventsNative so PlayerUpdateData
// and Player pointers are never decoded through PE32 byte offsets.
func (s *Server) NetUpdateRemotePlrAudioEvents(obj *server.Object, updatePtr unsafe.Pointer, zone int8) {
	s.netUpdateRemotePlrAudioEventsNative(obj, (*server.PlayerUpdateData)(updatePtr), byte(zone))
}

func (s *Server) netSendAudioEvents(obj *server.Object) {
	s.Audio.EachEventBitmap(func(it *server.AudioEvent) {
		s.netSendAudioEvent(obj, it, int16(it.Perc))
	})
}
