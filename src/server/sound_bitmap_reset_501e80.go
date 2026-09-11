package server

const soundBitmapWordCount501E80 = 0x20

// soundBitmap501E80 is the exact 0x80-byte bitmap cleared by GAME.EXE
// 00501E80. Each bit selects one of the 1024 sound definitions.
type soundBitmap501E80 [soundBitmapWordCount501E80]uint32

// resetSoundBitmap501E80 preserves GAME.EXE 00501E80: clear exactly 32
// dwords and leave a zero return value in EAX. The original function does not
// change the lifetime of the surrounding audio-event collection.
func resetSoundBitmap501E80(bitmap *soundBitmap501E80) int32 {
	*bitmap = soundBitmap501E80{}
	return 0
}

// resetBitmapForRemotePlayer501CA0 keeps the Go-owned event queue transition
// separate from the original 00501E80 helper. The transition was introduced
// when events produced outside the audio phase began to be delayed until the
// next frame; it is not an additional write performed by GAME.EXE 00501E80.
func (s *serverAudio) resetBitmapForRemotePlayer501CA0() {
	_ = resetSoundBitmap501E80(&s.bitmap)
	s.inAudio = false
}
