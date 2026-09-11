package server

import "github.com/opennox/opennox/v1/legacy/common/alloc"

// resetAudioEvents502100 binds GAME.EXE 00502100 to the native allocation
// class and event-list head. The Go-owned collection phase and delayed queues
// remain outside this exact original helper.
func (s *serverAudio) resetAudioEvents502100() {
	resetAudioEvents502100(audioEventResetHooks502100[*alloc.Class]{
		loadClass: func() *alloc.Class {
			return s.alloc.Class
		},
		freeAllObjects: func(class *alloc.Class) {
			class.FreeAllObjects()
		},
		clearHead: func() {
			s.head = nil
		},
	})
}
