package server

import "github.com/opennox/opennox/v1/common/sound"

// ReadAUD binds the verified GAME.EXE 00502320 list reader to the native Go
// file abstraction and audio descriptor table.
func (s *serverAudio) ReadAUD(f File) bool {
	return readAudioEventAUDList502320(audioEventAUDListHooks502320{
		readCount: f.ReadI32,
		readRecord: func() bool {
			return s.readAUDRec(f)
		},
	})
}

// readAUDRec binds the verified GAME.EXE 00502370 record reader. File
// position changes and descriptor writes retain the original ordering; only
// the caller-owned input abstraction is Go-native.
func (s *serverAudio) readAUDRec(f File) bool {
	return readAudioEventAUDRecord502370(audioEventAUDRecordHooks502370[sound.ID, *audioEvent2]{
		readName: f.ReadString8,
		resolveSound: func(name string) sound.ID {
			return sound.ByName(name)
		},
		zeroSound: sound.SoundNone,
		loadInitialized: func() bool {
			return s.inited
		},
		readI16: f.ReadI16,
		readU8:  f.ReadU8,
		readI8:  f.ReadI8,
		skip:    f.Skip,
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
