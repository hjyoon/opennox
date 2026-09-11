package server

func (s *serverAudio) descriptor501EA0(soundID int32) *audioEvent2 {
	return &s.bySound[uint32(soundID)]
}

// AddAudio binds GAME.EXE 00501EA0, 00501EF0, 00501F10, and 00501F30 to
// native Go pointers. PE32 scalar fields are narrowed explicitly to dwords so
// amd64 and arm64 builds retain the original signed overflow behavior.
func (s *serverAudio) AddAudio(event *AudioEvent, percentage int) {
	audioEventInsert501EA0(event, int32(percentage), audioEventInsertHooks501EA0[
		*audioEvent2,
		*AudioEvent,
	]{
		loadEventSound: func(event *AudioEvent) int32 {
			return int32(event.Sound)
		},
		descriptor: s.descriptor501EA0,
		testBitmap: func(soundID int32) uint32 {
			return testSoundBitmap501EF0(&s.bitmap, soundID)
		},
		setBitmap: func(soundID int32) {
			setSoundBitmap501F10(&s.bitmap, soundID)
		},
		storeListHead: func(descriptor *audioEvent2, head *AudioEvent) {
			descriptor.Field24 = head
		},
		storeEventPercent: func(event *AudioEvent, percentage int32) {
			event.Perc = int(percentage)
		},
		insert: s.insertAudioEventList501F30,
	})
}

func (s *serverAudio) insertAudioEventList501F30(descriptor *audioEvent2, event *AudioEvent) {
	insertAudioEventList501F30(descriptor, event, audioEventListHooks501F30[
		*audioEvent2,
		*AudioEvent,
	]{
		loadHead: func(descriptor *audioEvent2) *AudioEvent {
			return descriptor.Field24
		},
		storeHead: func(descriptor *audioEvent2, head *AudioEvent) {
			descriptor.Field24 = head
		},
		loadPercent: func(event *AudioEvent) int32 {
			return int32(event.Perc)
		},
		loadSound: func(event *AudioEvent) int32 {
			return int32(event.Sound)
		},
		descriptor: s.descriptor501EA0,
		loadFlagsLow: func(descriptor *audioEvent2) uint8 {
			return uint8(descriptor.Field16)
		},
		loadLimit: func(descriptor *audioEvent2) int32 {
			return int32(descriptor.Field20)
		},
		loadNext: func(event *AudioEvent) *AudioEvent {
			return event.list28
		},
		storeNext: func(event, next *AudioEvent) {
			event.list28 = next
		},
	})
}
