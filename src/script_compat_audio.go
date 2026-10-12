package opennox

import (
	"github.com/opennox/libs/script"
	"github.com/opennox/noxscript/ns/v4/audio"
)

// The legacy compatibility API uses the active game, as do its Global,
// HostPlayer, and CinemaPlayers entry points. Keep its sound-name and position
// semantics in the existing native namespace, including next-frame queuing.
func scriptCompatAudioEffect(name string, pos script.Positioner) {
	if pos == nil {
		return
	}
	noxServer.NoxScript().AudioEvent(audio.Name(name), pos)
}
