package server

// FlushAudioEvents502060 binds the verified GAME.EXE 00502060 traversal to
// the native Go audio-event records. The send callback receives full native
// pointers; only the original dword scalar fields are narrowed explicitly.
func (s *serverAudio) FlushAudioEvents502060(
	listener *Object,
	send func(*Object, *AudioEvent, int32) bool,
) {
	flushAudioEvents502060(listener, audioEventFlushHooks502060[
		*Object,
		*audioEvent2,
		*AudioEvent,
	]{
		loadBitmapWord: func(word int32) uint32 {
			return s.bitmap[word]
		},
		descriptor: func(soundID int32) *audioEvent2 {
			return &s.bySound[soundID]
		},
		loadHead: func(descriptor *audioEvent2) *AudioEvent {
			return descriptor.Field24
		},
		loadLimit: func(descriptor *audioEvent2) int32 {
			return int32(descriptor.Field20)
		},
		loadPercentage: func(event *AudioEvent) int32 {
			return int32(event.Perc)
		},
		send: send,
		loadNext: func(event *AudioEvent) *AudioEvent {
			return event.list28
		},
	})
}
