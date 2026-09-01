package opennox

import (
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

type playerPhonemeBroadcastNativeDeps4FC960 struct {
	firstUnit       func() *server.Object
	nextUnit        func(*server.Object) *server.Object
	spellGetPhoneme func(uint32, int8) int32
	audioEvent      func(int32, *server.Object, int32, uint32)
}

func playerPhonemeBroadcastNative4FC960(
	source *server.Object,
	phoneme int8,
	deps playerPhonemeBroadcastNativeDeps4FC960,
) int32 {
	return playerPhonemeBroadcast4FC960(
		source,
		phoneme,
		playerPhonemeBroadcastHooks4FC960[*server.Object]{
			firstUnit: deps.firstUnit,
			nextUnit:  deps.nextUnit,
			loadNetCode: func(unit *server.Object) uint32 {
				return unit.NetCode
			},
			spellGetPhoneme: deps.spellGetPhoneme,
			audioEvent:      deps.audioEvent,
		},
	)
}

// PlayerPhonemeBroadcast4FC960 binds GAME.EXE 004FC960 to the native player
// list and native-width Object pointers.
//
//go:noinline
func (s *Server) PlayerPhonemeBroadcast4FC960(source *server.Object, phoneme int8) int32 {
	return playerPhonemeBroadcastNative4FC960(
		source,
		phoneme,
		playerPhonemeBroadcastNativeDeps4FC960{
			firstUnit: s.Players.FirstUnit,
			nextUnit:  s.Players.NextUnit,
			spellGetPhoneme: func(netCode uint32, phoneme int8) int32 {
				return legacy.Nox_xxx_spellGetPhoneme_4FE1C0(netCode, phoneme)
			},
			audioEvent: func(id int32, source *server.Object, kind int32, listener uint32) {
				s.Audio.EventObj(sound.ID(id), source, int(kind), listener)
			},
		},
	)
}
