//go:build !server

package ail

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/opennox/libs/env"
	"github.com/timshannon/go-openal/openal"

	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
)

// These tests use the actual default playback device, not OpenAL's null driver.
// The generated PCM is silent so desktop verification does not produce a tone.
func openPlaybackTestDriver(t *testing.T) Driver {
	t.Helper()
	if os.Getenv("NOX_TEST_OPENAL") != "true" {
		t.Skip("set NOX_TEST_OPENAL=true to test the default OpenAL playback device")
	}
	if env.IsE2E() {
		t.Fatal("unset NOX_E2E and NOX_E2E_RECORD for real OpenAL tests")
	}
	runtime.LockOSThread()
	t.Cleanup(runtime.UnlockOSThread)
	handles.Init()
	t.Cleanup(handles.Release)
	h := WaveOutOpen()
	if h == 0 {
		t.Fatal("cannot open the default OpenAL playback device")
	}
	t.Logf("native playback driver=%#x renderer=%q", h, openal.GetString(0xB003 /* AL_RENDERER */))
	t.Cleanup(func() {
		if err := h.Close(); err != nil {
			t.Errorf("close playback driver: %v", err)
		}
	})
	return h
}

func silentPlaybackWAV(t *testing.T, duration time.Duration) string {
	t.Helper()
	const rate = 22050
	pcmSize := int(duration*rate/time.Second) * 2
	b := make([]byte, 44+pcmSize)
	copy(b, "RIFF")
	binary.LittleEndian.PutUint32(b[4:8], uint32(len(b)-8))
	copy(b[8:16], "WAVEfmt ")
	binary.LittleEndian.PutUint32(b[16:20], 16)
	binary.LittleEndian.PutUint16(b[20:22], 1)
	binary.LittleEndian.PutUint16(b[22:24], 1)
	binary.LittleEndian.PutUint32(b[24:28], rate)
	binary.LittleEndian.PutUint32(b[28:32], rate*2)
	binary.LittleEndian.PutUint16(b[32:34], 2)
	binary.LittleEndian.PutUint16(b[34:36], 16)
	copy(b[36:40], "data")
	binary.LittleEndian.PutUint32(b[40:44], uint32(pcmSize))
	path := filepath.Join(t.TempDir(), "silence.wav")
	if err := os.WriteFile(path, b, 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestOpenALStreamPause(t *testing.T) {
	h := openPlaybackTestDriver(t)
	s := h.OpenStream(silentPlaybackWAV(t, time.Second))
	if s == 0 {
		t.Fatal("cannot open PCM stream")
	}
	t.Cleanup(func() { _ = s.Close() })
	s.SetVolume(0)
	s.Start()
	h.get().doWork()
	if got := s.get().source.State(); got != openal.Playing {
		t.Fatalf("started source = %v, want Playing", got)
	}
	s.Pause(true)
	if got := s.get().source.State(); got != openal.Paused {
		t.Fatalf("paused source = %v, want Paused", got)
	}
	pos := s.get().source.Geti(openal.AlSampleOffset)
	time.Sleep(40 * time.Millisecond)
	if got := s.get().source.Geti(openal.AlSampleOffset); got != pos {
		t.Fatalf("paused hardware position = %d, want unchanged %d", got, pos)
	}
	s.Pause(false)
	if got := s.get().source.State(); got != openal.Playing {
		t.Fatalf("resumed source = %v, want Playing", got)
	}
}

func TestOpenALShortStreamDrain(t *testing.T) {
	h := openPlaybackTestDriver(t)
	s := h.OpenStream(silentPlaybackWAV(t, 40*time.Millisecond))
	if s == 0 {
		t.Fatal("cannot open short PCM stream")
	}
	t.Cleanup(func() { _ = s.Close() })
	s.SetVolume(0)
	s.Start()
	h.get().doWork()
	if got := s.get().source.State(); got != openal.Playing {
		t.Fatalf("short source = %v, want Playing even after decoder EOF", got)
	}
	if got := s.Status(); got != 4 {
		t.Fatalf("draining stream status = %d, want 4", got)
	}
	deadline := time.Now().Add(2 * time.Second)
	for s.Status() == 4 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
		h.get().doWork()
	}
	if got := s.Status(); got != 2 {
		t.Fatalf("finished stream status = %d, want 2", got)
	}
	if got := openal.Err(); got != nil {
		t.Fatalf("playback error: %v", got)
	}
}

func TestOpenALStreamCloseReleasesSource(t *testing.T) {
	h := openPlaybackTestDriver(t)
	path := silentPlaybackWAV(t, time.Second)
	for i := 0; i < 32; i++ {
		s := h.OpenStream(path)
		if s == 0 {
			t.Fatalf("cannot open stream %d", i)
		}
		if err := s.Close(); err != nil {
			t.Fatal(err)
		}
		if s.get() != nil || h.get().streamHead != nil {
			t.Fatalf("closed stream %d retains its decoder/source or driver link", i)
		}
		if err := s.Close(); err != nil {
			t.Fatalf("repeated close: %v", err)
		}
	}
	if got := openal.Err(); got != nil {
		t.Fatalf("close error: %v", got)
	}
}

func TestOpenALDriverCloseReleasesHandles(t *testing.T) {
	h := openPlaybackTestDriver(t)
	s := h.OpenStream(silentPlaybackWAV(t, time.Second))
	sample := h.AllocateSample()
	if s == 0 || sample == 0 {
		t.Fatal("cannot allocate stream and sample")
	}
	if err := h.Close(); err != nil {
		t.Fatal(err)
	}
	if h.get() != nil || s.get() != nil || sample.get() != nil {
		t.Fatal("closed driver retains native playback handles")
	}
	if err := h.Close(); err != nil {
		t.Fatalf("repeated driver close: %v", err)
	}
}

func TestOpenALShutdownReleasesHandles(t *testing.T) {
	h := openPlaybackTestDriver(t)
	s := h.OpenStream(silentPlaybackWAV(t, time.Second))
	sample := h.AllocateSample()
	timer := RegisterTimer(func(uint32) {})
	timer.SetFrequency(30)
	timer.Start()
	Shutdown()
	if h.get() != nil || s.get() != nil || sample.get() != nil || timer.get() != nil {
		t.Fatal("shutdown retains native playback or timer handles")
	}
	Shutdown()
}
