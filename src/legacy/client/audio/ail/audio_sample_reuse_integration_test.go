//go:build !server

package ail

import (
	"context"
	"encoding/binary"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/timshannon/go-openal/openal"
)

// Every input block is valid in its declared IMA ADPCM channel layout. In a
// mono block, byte6 is payload0x93, not a right-channel header index147.
func sampleReuseADPCM(t *testing.T, stereo bool, blocks int) []byte {
	t.Helper()
	data := make([]byte, 1024*blocks)
	for off := 0; off < len(data); off += 1024 {
		binary.LittleEndian.PutUint16(data[off:], 1000)
		data[off+2] = 7
		if stereo {
			binary.LittleEndian.PutUint16(data[off+4:], 2000)
			data[off+6] = 5
		} else {
			data[off+6] = 147
		}
		var pcm []int16
		if stereo {
			pcm = decodeADPCMStereo(nil, data[off:off+1024])
		} else {
			pcm = decodeADPCMMono(nil, data[off:off+1024])
		}
		want := 2041
		if stereo {
			want = 2034
		}
		if len(pcm) != want {
			t.Fatalf("valid source block values=%d want=%d", len(pcm), want)
		}
	}
	return data
}

func TestOpenALSampleReuseDoesNotDecodePreviousFormat(t *testing.T) {
	const child = "NOX_TEST_ADPCM_SAMPLE_REUSE_CHILD"
	mode := os.Getenv(child)
	if mode == "" {
		if os.Getenv("NOX_TEST_OPENAL") != "true" {
			t.Skip("set NOX_TEST_OPENAL=true for native sample reuse")
		}
		for _, mode := range []string{"mono-stereo", "stereo-mono", "mono-mono", "stereo-stereo"} {
			t.Run(mode, func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
				defer cancel()
				cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestOpenALSampleReuseDoesNotDecodePreviousFormat$", "-test.v")
				cmd.Dir = t.TempDir()
				for _, kv := range os.Environ() {
					k, _, _ := strings.Cut(kv, "=")
					if k == child || strings.HasPrefix(k, "NOX_E2E") {
						continue
					}
					cmd.Env = append(cmd.Env, kv)
				}
				cmd.Env = append(cmd.Env, child+"="+mode)
				out, err := cmd.CombinedOutput()
				t.Logf("%s", out)
				if err != nil {
					t.Fatalf("actual sample reuse %s: %v", mode, err)
				}
			})
		}
		return
	}
	parts := strings.Split(mode, "-")
	if len(parts) != 2 {
		t.Fatal("mode")
	}
	oldStereo, newStereo := parts[0] == "stereo", parts[1] == "stereo"
	driver := openPlaybackTestDriver(t)
	voice := driver.AllocateSample()
	if voice == 0 {
		t.Fatal("native sample")
	}
	defer voice.Release()
	configure := func(stereo bool) {
		voice.Init()
		format := int32(5)
		if stereo {
			format = 7
		}
		voice.SetType(format, 0)
		voice.SetADPCMBlockSize(1024)
		voice.SetPlaybackRate(22050)
		voice.SetVolume(0)
	}
	configure(oldStereo)
	first, second := sampleReuseADPCM(t, oldStereo, 8), sampleReuseADPCM(t, oldStereo, 1)
	if ready := voice.BufferReady(); ready != 0 {
		t.Fatal("first old buffer")
	}
	voice.LoadBuffer(0, first)
	if ready := voice.BufferReady(); ready != 1 {
		t.Fatal("second old buffer")
	}
	voice.LoadBuffer(1, second)
	driver.get().doWork()
	old := voice.get()
	if old.buffers[0].pos == 0 || len(old.buffers[1].buf) == 0 || old.buffers[1].pos != 0 || old.source.BuffersQueued() != 4 {
		t.Fatal("old playback must retain an unconsumed second input and four real queued buffers")
	}
	queued := driver.PlaybackStats().SampleBuffersQueued
	configure(newStereo)
	if old.source.BuffersQueued() != 0 {
		t.Fatal("reinitialization retained old hardware buffers")
	}
	if ready := voice.BufferReady(); ready != 0 {
		t.Fatal("first new buffer")
	}
	voice.LoadBuffer(0, sampleReuseADPCM(t, newStereo, 1))
	// Before the fix this reaches decodeNibble index147 for mono->stereo,
	// exactly matching the reported crash. No recover/clamping hides it.
	driver.get().doWork()
	if delta := driver.PlaybackStats().SampleBuffersQueued - queued; delta != 1 {
		t.Fatalf("new generation submitted%d blocks, want only its one new block", delta)
	}
	if old.buffers[0].buf != nil || old.buffers[1].buf != nil {
		t.Fatal("completed/newly-reset inputs retain old bytes")
	}
	if err := openal.Err(); err != nil {
		t.Fatal(err)
	}
	t.Logf("reuse=%s native=%+v; old queue discarded, new block decoded in its declared format", mode, driver.PlaybackStats())
}

func TestOpenALSampleCompletesOnlySubmittedBuffers(t *testing.T) {
	driver := openPlaybackTestDriver(t)
	voice := driver.AllocateSample()
	if voice == 0 {
		t.Fatal("native sample")
	}
	defer voice.Release()
	voice.Init()
	voice.SetType(5, 0)
	voice.SetADPCMBlockSize(1024)
	voice.SetPlaybackRate(22050)
	voice.SetVolume(0)
	for i := 0; i < 2; i++ {
		if got := voice.BufferReady(); got != i {
			t.Fatal("input sequence")
		}
		voice.LoadBuffer(uint32(i), sampleReuseADPCM(t, false, 1))
	}
	driver.get().doWork()
	if stats := driver.PlaybackStats(); stats.SampleBuffersQueued != 2 {
		t.Fatalf("submitted=%d want2", stats.SampleBuffersQueued)
	}
	if s := voice.get(); s.buffers[0].buf != nil || s.buffers[1].buf != nil {
		t.Fatal("completed source slots must be empty")
	}
}

func TestOpenALSampleReadySlotAfterDrain(t *testing.T) {
	driver := openPlaybackTestDriver(t)
	voice := driver.AllocateSample()
	if voice == 0 {
		t.Fatal("native sample")
	}
	defer voice.Release()
	voice.Init()
	voice.SetType(5, 0)
	voice.SetADPCMBlockSize(1024)
	voice.SetPlaybackRate(22050)
	voice.SetVolume(0)
	if got := voice.BufferReady(); got != 0 {
		t.Fatal("initial ready slot")
	}
	voice.LoadBuffer(0, sampleReuseADPCM(t, false, 1))
	driver.get().doWork()
	if got := voice.BufferReady(); got != 1 {
		t.Fatalf("after drain current ready slot=%d want1", got)
	}
	voice.LoadBuffer(1, sampleReuseADPCM(t, false, 1))
	driver.get().doWork()
	if got := driver.PlaybackStats().SampleBuffersQueued; got != 2 {
		t.Fatalf("refill submitted=%d want2", got)
	}
}
