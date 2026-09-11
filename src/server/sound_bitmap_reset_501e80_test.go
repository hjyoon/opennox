package server

import (
	"testing"
	"unsafe"
)

func TestResetSoundBitmap501E80(t *testing.T) {
	if got := unsafe.Sizeof(soundBitmap501E80{}); got != 0x80 {
		t.Fatalf("bitmap size: got %#x, want 0x80", got)
	}

	type guardedBitmap struct {
		before uint64
		bitmap soundBitmap501E80
		after  uint64
	}
	const (
		before = uint64(0x0123456789ABCDEF)
		after  = uint64(0xFEDCBA9876543210)
	)
	g := guardedBitmap{before: before, after: after}
	for i := range g.bitmap {
		g.bitmap[i] = uint32(i+1) * 0x01010101
	}

	if got := resetSoundBitmap501E80(&g.bitmap); got != 0 {
		t.Fatalf("return value: got %d, want 0", got)
	}
	for i, word := range g.bitmap {
		if word != 0 {
			t.Fatalf("bitmap[%d]: got %#08x, want 0", i, word)
		}
	}
	if g.before != before || g.after != after {
		t.Fatalf("neighboring state changed: before=%#x after=%#x", g.before, g.after)
	}
}

func TestResetSoundBitmap501E80DoesNotCloseAudioCollection(t *testing.T) {
	audio := serverAudio{inAudio: true}
	for i := range audio.bitmap {
		audio.bitmap[i] = ^uint32(0)
	}

	if got := resetSoundBitmap501E80(&audio.bitmap); got != 0 {
		t.Fatalf("return value: got %d, want 0", got)
	}
	if !audio.inAudio {
		t.Fatal("00501E80 changed the Go-owned audio collection lifetime")
	}

	audio.resetBitmapForRemotePlayer501CA0()
	if audio.inAudio {
		t.Fatal("remote-player wrapper did not close the audio collection")
	}
}
