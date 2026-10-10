//go:build !server

package music

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/opennox/libs/env"
	"github.com/opennox/opennox/v1/legacy/client/audio/ail"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/legacy/timer"
)

const musicLoopChild = "NOX_TEST_MUSIC_LOOP_CHILD"

// Isolate the cached E2E backend and timer globals from other package tests.
func runMusicLoopChild(t *testing.T, mode string, extra ...string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestMusicLoopChild$", "-test.count=1", "-test.v", "-test.timeout=85s")
	cmd.Dir = t.TempDir()
	for _, kv := range os.Environ() {
		key, _, _ := strings.Cut(kv, "=")
		if strings.HasPrefix(key, "NOX_E2E") || strings.HasPrefix(key, "ALSOFT_") || strings.HasPrefix(key, "NOX_TEST_MUSIC_") || key == "NOX_DATA" {
			continue
		}
		cmd.Env = append(cmd.Env, kv)
	}
	cmd.Env = append(cmd.Env, musicLoopChild+"="+mode)
	if mode == "mock" {
		cmd.Env = append(cmd.Env, "NOX_E2E=music-loop-state", "NOX_E2E_AUDIO=mock")
	} else {
		cmd.Env = append(cmd.Env, "ALSOFT_DRIVERS=null", "ALSOFT_CONF="+filepath.Join(cmd.Dir, "isolated-openal.ini"))
	}
	cmd.Env = append(cmd.Env, extra...)
	out, err := cmd.CombinedOutput()
	t.Logf("%s", out)
	if err != nil {
		t.Fatalf("music loop child %s: %v", mode, err)
	}
}

func TestMusicUpdatePersistsResumeAndLoopState(t *testing.T) {
	runMusicLoopChild(t, "mock")
}

// Opt-in actual decoding/queueing on OpenAL's null device, never speakers or
// a GUI. Read-only stock input and an isolated child directory/configuration.
func TestMusicChap1NativeResumeThenFullLoop(t *testing.T) {
	path := os.Getenv("NOX_TEST_MUSIC_WAV")
	if path == "" {
		t.Skip("set NOX_TEST_MUSIC_WAV to the unchanged stock music/chap1.wav")
	}
	path, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	runMusicLoopChild(t, "native", "NOX_TEST_MUSIC_WAV="+path)
}

func newMusicLoopModule(t *testing.T, dir string, driver ail.Driver, tick *uint64) *Module {
	t.Helper()
	oldClock := timer.PlatformTicks
	timer.PlatformTicks = func() uint64 { return *tick }
	t.Cleanup(func() { timer.PlatformTicks = oldClock })
	var enabled, ready, fallback, initialized, pending, unused1, unused2 uint32
	var actualDriver ail.Driver
	var volume timer.TimerGroup
	volume.Init()
	volumePtr := &volume
	m := NewModule(dir, &unused1, &unused2, &actualDriver, &enabled, &fallback, &initialized,
		&volume, &volumePtr, &pending, func() ail.Driver { return driver }, func() bool { return false },
		func() string { return "" }, func() uint64 { return *tick })
	m.Init()
	ready = *m.dword_587000_93156
	if ready != 1 || *m.dword_5d4594_816376 != driver {
		t.Fatal("fixture must use the actual initialization path")
	}
	t.Cleanup(m.Sub_43D650)
	return m
}

func TestMusicLoopChild(t *testing.T) {
	mode := os.Getenv(musicLoopChild)
	if mode == "" {
		t.Skip("isolated music-loop subprocess only")
	}
	handles.Init()
	t.Cleanup(handles.Release)
	if mode == "mock" {
		if !env.IsE2E() {
			t.Fatal("mock state test requires isolated E2E audio")
		}
		for _, slot := range []uint32{0, 1} {
			for _, position := range []uint32{0, 1, 22050, 725469} {
				t.Run(fmt.Sprintf("slot-%d/resume-%d", slot, position), func(t *testing.T) {
					var tick uint64
					m := newMusicLoopModule(t, "music", ail.Driver(handles.New()), &tick)
					m.block_idx = 1 - slot
					untouched := MusicBlock{Inner: MusicState{MusicIdx: 29, Volume: 17, Position: 123, D: 71}, Loaded: 9}
					m.blocks[m.block_idx] = untouched
					m.SetNextMusic(MusicState{MusicIdx: 1, Volume: 80, Position: position, D: 99})
					m.Update()
					if m.playingStream == 0 || m.currentPlaying.Position != position {
						t.Fatal("first playback must resume at the saved position")
					}
					if got := m.blocks[slot]; got.Inner.Position != 0 || got.Loaded != 1 || got.Inner.MusicIdx != 1 || got.Inner.Volume != 80 || got.Inner.D != 99 {
						t.Fatalf("resume position/Loaded must persist in the selected block, got %+v", got)
					}
					for loop := 0; loop < 4; loop++ {
						// Mock streams report EOF; exercise the actual state1->2->0
						// transitions and startPlay rather than resetting test state.
						m.Update()
						m.Update()
						m.Update()
						if m.playingStream == 0 || m.currentPlaying.Position != 0 || m.blocks[slot].Inner.Position != 0 || *m.dword_5d4594_816348 != 1 {
							t.Fatalf("loop %d reused a stale resume position: playing=%+v block=%+v", loop, m.currentPlaying, m.blocks[slot])
						}
					}
					if m.blocks[1-slot] != untouched {
						t.Fatal("inactive block changed")
					}
				})
			}
		}
		for _, index := range []uint32{0, uint32(len(audioTable)), ^uint32(0)} {
			t.Run(fmt.Sprintf("invalid-%d", index), func(t *testing.T) {
				var tick uint64
				m := newMusicLoopModule(t, "music", ail.Driver(handles.New()), &tick)
				m.SetNextMusic(MusicState{MusicIdx: index, Volume: 75, Position: 123})
				m.Update()
				if m.playingStream != 0 || m.blocks[m.block_idx].Inner.MusicIdx != 0 || *m.dword_5d4594_816348 != 0 {
					t.Fatal("invalid music selection must be cleared in the persistent block")
				}
			})
		}
		return
	}
	if mode != "native" || env.IsE2E() || os.Getenv("ALSOFT_DRIVERS") != "null" {
		t.Fatal("native test requires OpenAL null output, not mock audio or speakers")
	}
	path := os.Getenv("NOX_TEST_MUSIC_WAV")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != "chap1.wav" || len(data) != 753804 || string(data[8:16]) != "WAVEfmt " || binary.LittleEndian.Uint16(data[20:22]) != 0x11 || binary.LittleEndian.Uint16(data[22:24]) != 2 {
		t.Fatal("expected unchanged stock stereo IMA ADPCM chap1.wav")
	}
	const frames = 747519
	const rate = 22050
	if binary.LittleEndian.Uint32(data[24:28]) != rate || string(data[40:44]) != "fact" || binary.LittleEndian.Uint32(data[48:52]) != frames {
		t.Fatal("stock sample rate/fact duration mismatch")
	}
	before := sha256.Sum256(data)
	runtime.LockOSThread()
	t.Cleanup(runtime.UnlockOSThread)
	driver := ail.WaveOutOpen()
	if driver == 0 {
		t.Fatal("cannot open actual OpenAL null device")
	}
	t.Cleanup(func() {
		if err := driver.Close(); err != nil {
			t.Error(err)
		}
	})
	var tick uint64
	m := newMusicLoopModule(t, filepath.Dir(path), driver, &tick)
	m.SetNextMusic(MusicState{MusicIdx: 1, Volume: 50, Position: frames - rate})
	m.Update()
	if m.currentPlaying.Position != frames-rate || m.blocks[m.block_idx].Inner.Position != 0 || m.blocks[m.block_idx].Loaded != 1 {
		t.Fatalf("native first resume did not consume its position: playing=%+v block=%+v", m.currentPlaying, m.blocks[m.block_idx])
	}
	firstStream := m.playingStream
	var fullStream ail.Stream
	var fullStarted time.Time
	maxPosition := 0
	deadline := time.Now().Add(70 * time.Second)
	for time.Now().Before(deadline) {
		ail.Serve()
		tick += 10
		m.Sub_43D2D0()
		m.Update()
		if m.playingStream != 0 && m.playingStream != firstStream {
			if fullStream == 0 {
				fullStream, fullStarted = m.playingStream, time.Now()
				if m.currentPlaying.Position != 0 {
					t.Fatal("after resumed tail, next loop must start at zero")
				}
			} else if m.playingStream != fullStream {
				elapsed := time.Since(fullStarted)
				if m.currentPlaying.Position != 0 || maxPosition != frames || elapsed < 33*time.Second || driver.PlaybackStats().StreamBuffersProcessed == 0 {
					t.Fatalf("shortened native loop: next=%+v frames=%d elapsed=%s stats=%+v", m.currentPlaying, maxPosition, elapsed, driver.PlaybackStats())
				}
				after, err := os.ReadFile(path)
				if err != nil || sha256.Sum256(after) != before {
					t.Fatal("stock audio changed during playback")
				}
				t.Logf("actual OpenAL null ADPCM: resumed last second -> full %d frames at %d Hz (%.6f seconds, observed %s) -> next loop at zero; stats=%+v", frames, rate, float64(frames)/rate, elapsed, driver.PlaybackStats())
				return
			}
		}
		if m.playingStream == fullStream && fullStream != 0 {
			maxPosition = max(maxPosition, m.playingStream.Position())
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("native resume/full-loop timeout: stats=%+v max position=%d", driver.PlaybackStats(), maxPosition)
}
