package server

import "github.com/opennox/opennox/v1/common/sound"

// ReadAVNT binds the verified GAME.EXE 00502120 parser to the native Go sound
// descriptors. File position changes and descriptor writes retain the
// original ordering; only the caller-owned input abstraction is Go-native.
func (s *serverAudio) ReadAVNT(f File) bool {
	return readAudioEventAVNT502120(audioEventAVNTHooks502120[sound.ID, *audioEvent2]{
		loadInitialized: func() bool {
			return s.inited
		},
		readName: f.ReadString8,
		resolveSound: func(name string) sound.ID {
			return sound.ByName(name)
		},
		zeroSound: sound.SoundNone,
		readU8:    f.ReadU8,
		readI16:   f.ReadI16,
		skip:      f.Skip,
		descriptor: func(id sound.ID) *audioEvent2 {
			return &s.bySound[id]
		},
		storeMaxDistance: func(descriptor *audioEvent2, value int32) {
			descriptor.MaxDist = int(value)
		},
		storeFlags: func(descriptor *audioEvent2, value uint32) {
			descriptor.Flags = value
		},
		storeField8: func(descriptor *audioEvent2, value uint32) {
			descriptor.Field8 = value
		},
		incrementField12: func(descriptor *audioEvent2) {
			descriptor.Field12++
		},
		storeField16: func(descriptor *audioEvent2, value uint32) {
			descriptor.Field16 = value
		},
		storeField20: func(descriptor *audioEvent2, value uint32) {
			descriptor.Field20 = int(value)
		},
	})
}
