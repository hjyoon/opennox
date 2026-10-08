//go:build !server

package ail

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/opennox/libs/env"
	"github.com/timshannon/go-openal/openal"

	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
)

const stockAudioBankChild = "NOX_TEST_STOCK_FX_CHILD"

// This opt-in test reads the bank without modifying assets or personal settings.
// It plays every complete sample at its stock rate through the real OpenAL null
// driver, not mock audio. This does not verify speakers or gameplay event gates.
func TestNativeStockAudioBankPlayback(t *testing.T) {
	base := os.Getenv("NOX_TEST_STOCK_FX_DIR")
	if base == "" {
		t.Skip("set NOX_TEST_STOCK_FX_DIR to a directory containing Audio.idx and Audio.bag")
	}
	base, err := filepath.Abs(base)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestNativeStockAudioBankPlaybackChild$", "-test.count=1", "-test.v", "-test.timeout=10m")
	cmd.Dir = t.TempDir()
	for _, kv := range os.Environ() {
		key, _, _ := strings.Cut(kv, "=")
		if strings.HasPrefix(key, "NOX_TEST_STOCK_FX_") || strings.HasPrefix(key, "NOX_E2E") || strings.HasPrefix(key, "ALSOFT_") || key == "NOX_DATA" || key == "NOX_DEBUG_AUDIO" {
			continue
		}
		cmd.Env = append(cmd.Env, kv)
	}
	cmd.Env = append(cmd.Env, stockAudioBankChild+"=true", "NOX_TEST_STOCK_FX_DIR="+base, "ALSOFT_DRIVERS=null", "ALSOFT_CONF="+filepath.Join(cmd.Dir, "no-personal-openal.ini"))
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("stock audio bank subprocess failed: %v\n%s", err, out)
	}
	t.Logf("%s", out)
}

type stockPlaybackEntry struct {
	name     string
	rate     uint32
	block    uint32
	channels int
	data     []byte
	frames   uint64
}

func readStockPlaybackEntries(t *testing.T, idx, bag []byte) []stockPlaybackEntry {
	t.Helper()
	// Parse the raw index independently of the root package's bank loader, so a
	// lost or overwritten entry in that loader cannot reduce playback coverage.
	if len(idx) < 12 || string(idx[:4]) != "GABA" {
		t.Fatal("invalid GABA bank header")
	}
	version := binary.LittleEndian.Uint32(idx[4:8])
	count := uint64(binary.LittleEndian.Uint32(idx[8:12]))
	if version != 2 || count == 0 || uint64(len(idx)) != 12+36*count {
		t.Fatalf("unsupported stock bank: version=%d count=%d indexBytes=%d", version, count, len(idx))
	}
	entries := make([]stockPlaybackEntry, 0, int(count))
	names := make(map[string]bool)
	for i := uint64(0); i < count; i++ {
		rec := idx[12+36*i : 12+36*(i+1)]
		name := string(bytes.TrimRight(rec[:16], "\x00"))
		key := strings.ToLower(name)
		if name == "" || strings.ContainsRune(name, 0) || names[key] {
			t.Fatalf("empty, embedded-NUL or duplicate bank entry %d: %q", i, name)
		}
		names[key] = true
		off := uint64(binary.LittleEndian.Uint32(rec[16:20]))
		size := uint64(binary.LittleEndian.Uint32(rec[20:24]))
		rate := binary.LittleEndian.Uint32(rec[24:28])
		flags := binary.LittleEndian.Uint32(rec[28:32])
		block := binary.LittleEndian.Uint32(rec[32:36])
		channels := 1 + int(flags&1)
		if off+size > uint64(len(bag)) || size == 0 || rate == 0 || flags&^uint32(13) != 0 || flags&8 == 0 || block < uint32(4*channels) || size%uint64(block) != 0 || (channels == 2 && (block-8)%8 != 0) {
			t.Fatalf("invalid stock ADPCM entry %q: offset=%d size=%d rate=%d flags=%#x block=%d", name, off, size, rate, flags, block)
		}
		framesPerBlock := uint64(1 + (int(block)-4*channels)*2/channels)
		entries = append(entries, stockPlaybackEntry{
			name: name, rate: rate, block: block, channels: channels,
			data: bag[off : off+size], frames: size / uint64(block) * framesPerBlock,
		})
	}
	return entries
}

func decodeStockPlaybackEntries(t *testing.T, entries []stockPlaybackEntry) uint64 {
	t.Helper()
	pcmHash := sha256.New()
	var blocks, values, nonzeroValues uint64
	var decoded []int16
	var pcmBytes []byte
	for _, entry := range entries {
		_, _ = pcmHash.Write([]byte(entry.name + "\x00"))
		for off := 0; off < len(entry.data); off += int(entry.block) {
			data := entry.data[off : off+int(entry.block)]
			for ch := 0; ch < entry.channels; ch++ {
				if data[4*ch+2] > 88 {
					t.Fatalf("%s block %d channel %d: ADPCM index=%d", entry.name, off/int(entry.block), ch, data[4*ch+2])
				}
			}
			if entry.channels == 1 {
				decoded = decodeADPCMMono(decoded[:0], data)
			} else {
				decoded = decodeADPCMStereo(decoded[:0], data)
			}
			want := entry.channels + 2*(int(entry.block)-4*entry.channels)
			if len(decoded) != want {
				t.Fatalf("%s block %d: decoded %d PCM values, want %d", entry.name, off/int(entry.block), len(decoded), want)
			}
			for ch := 0; ch < entry.channels; ch++ {
				if decoded[ch] != int16(binary.LittleEndian.Uint16(data[4*ch:])) {
					t.Fatalf("%s block %d channel %d: initial predictor lost", entry.name, off/int(entry.block), ch)
				}
			}
			if cap(pcmBytes) < 2*want {
				pcmBytes = make([]byte, 2*want)
			}
			pcmBytes = pcmBytes[:2*want]
			for i, value := range decoded {
				binary.LittleEndian.PutUint16(pcmBytes[2*i:], uint16(value))
				if value != 0 {
					nonzeroValues++
				}
			}
			_, _ = pcmHash.Write(pcmBytes)
			values += uint64(want)
			blocks++
		}
	}
	if nonzeroValues == 0 {
		t.Fatal("the entire stock bank decoded to silence")
	}
	// This fingerprint records what was decoded; it is not an independent
	// Windows/Miles PCM oracle or a replacement for existing audio goldens.
	t.Logf("decoded entries=%d blocks=%d PCM-values=%d nonzero-values=%d PCM-SHA256=%x", len(entries), blocks, values, nonzeroValues, pcmHash.Sum(nil))
	return blocks
}

func playStockPlaybackBatch(t *testing.T, driver Driver, voices []Sample, entries []stockPlaybackEntry) {
	t.Helper()
	before := driver.PlaybackStats()
	var blocks uint64
	var longest time.Duration
	for i, entry := range entries {
		voice := voices[i]
		voice.Init()
		format := int32(5)
		if entry.channels == 2 {
			format = 7
		}
		voice.SetType(format, 0)
		voice.SetADPCMBlockSize(entry.block)
		voice.SetPlaybackRate(int(entry.rate))
		voice.SetVolume(127)
		if ready := voice.BufferReady(); ready != 0 {
			t.Fatalf("%s: initial ready buffer=%d, want 0", entry.name, ready)
		}
		voice.LoadBuffer(0, entry.data)
		blocks += uint64(len(entry.data)) / uint64(entry.block)
		duration := time.Duration((entry.frames*uint64(time.Second) + uint64(entry.rate) - 1) / uint64(entry.rate))
		if duration > longest {
			longest = duration
		}
	}
	deadline := time.Now().Add(longest + 3*time.Second)
	for {
		Serve()
		finished := true
		for i := range entries {
			if status := voices[i].Status(); status != 2 {
				if status != 4 {
					t.Fatalf("%s: unexpected playback status=%d", entries[i].name, status)
				}
				finished = false
			}
		}
		stats := driver.PlaybackStats()
		queued := stats.SampleBuffersQueued - before.SampleBuffersQueued
		processed := stats.SampleBuffersProcessed - before.SampleBuffersProcessed
		if queued > blocks || processed > queued {
			t.Fatalf("duplicated stock buffers: expected=%d queued=%d processed=%d", blocks, queued, processed)
		}
		if finished && queued == blocks && processed == blocks {
			if err := openal.Err(); err != nil {
				t.Fatalf("stock OpenAL playback error: %v", err)
			}
			return
		}
		if time.Now().After(deadline) {
			for i, entry := range entries {
				t.Logf("undrained batch sample=%s blocks=%d rate=%d channels=%d status=%d", entry.name, len(entry.data)/int(entry.block), entry.rate, entry.channels, voices[i].Status())
			}
			t.Fatalf("stock buffers did not drain: expected=%d queued=%d processed=%d finished=%t stats=%+v", blocks, queued, processed, finished, stats)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestNativeStockAudioBankPlaybackChild(t *testing.T) {
	if os.Getenv(stockAudioBankChild) != "true" {
		t.Skip("only run in the isolated stock audio bank subprocess")
	}
	if env.IsE2E() || os.Getenv("ALSOFT_DRIVERS") != "null" {
		t.Fatal("stock test requires real OpenAL null output, not E2E mock audio or speakers")
	}
	base := os.Getenv("NOX_TEST_STOCK_FX_DIR")
	if base == "" || !filepath.IsAbs(base) {
		t.Fatal("stock audio bank subprocess requires an absolute asset directory")
	}
	indexPath, bagPath := filepath.Join(base, "Audio.idx"), filepath.Join(base, "Audio.bag")
	idx, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatal(err)
	}
	bag, err := os.ReadFile(bagPath)
	if err != nil {
		t.Fatal(err)
	}
	indexHash, bagHash := sha256.Sum256(idx), sha256.Sum256(bag)
	t.Logf("Audio.idx bytes=%d SHA256=%x; Audio.bag bytes=%d SHA256=%x", len(idx), indexHash, len(bag), bagHash)
	entries := readStockPlaybackEntries(t, idx, bag)
	blocks := decodeStockPlaybackEntries(t, entries)

	runtime.LockOSThread()
	t.Cleanup(runtime.UnlockOSThread)
	handles.Init()
	t.Cleanup(handles.Release)
	driver := WaveOutOpen()
	if driver == 0 {
		t.Fatal("cannot open the OpenAL null device")
	}
	t.Cleanup(func() {
		if err := driver.Close(); err != nil {
			t.Errorf("close stock playback driver: %v", err)
		}
	})
	voices := make([]Sample, 16)
	for i := range voices {
		voices[i] = driver.AllocateSample()
		if voices[i] == 0 || voices[i].GetSource() == nil {
			t.Fatalf("stock voice %d has no native OpenAL source", i)
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(driver) <= uintptr(^uint32(0)) || uintptr(voices[i]) <= uintptr(^uint32(0))) {
			t.Fatalf("stock driver/voice %d handles did not exceed 4 GiB", i)
		}
	}
	t.Logf("native driver=%#x voice=%#x renderer=%q", driver, voices[0], openal.GetString(0xB003 /* AL_RENDERER */))
	// Sorting only schedules this isolated test. No sample is shortened, no
	// playback rate is accelerated, and no gameplay RNG or event is modified.
	sort.Slice(entries, func(i, j int) bool {
		a, b := entries[i], entries[j]
		left, right := a.frames*uint64(b.rate), b.frames*uint64(a.rate)
		if left != right {
			return left < right
		}
		return a.name < b.name
	})
	started := time.Now()
	for first := 0; first < len(entries); first += len(voices) {
		end := first + len(voices)
		if end > len(entries) {
			end = len(entries)
		}
		playStockPlaybackBatch(t, driver, voices, entries[first:end])
		if end%256 == 0 || end == len(entries) {
			t.Logf("played %d/%d complete stock samples in %s: %+v", end, len(entries), time.Since(started).Round(time.Millisecond), driver.PlaybackStats())
		}
	}
	stats := driver.PlaybackStats()
	if !stats.Native || stats.SampleSources != len(voices) || stats.PlayingSources != 0 || stats.StreamSources != 0 || stats.SampleBuffersQueued != blocks || stats.SampleBuffersProcessed != blocks {
		t.Fatalf("stock bank native playback incomplete: want %d blocks, got %+v", blocks, stats)
	}
	for _, voice := range voices {
		voice.Release()
	}
	if got := driver.PlaybackStats(); got.SampleSources != 0 || got.SampleBuffersQueued != blocks || got.SampleBuffersProcessed != blocks {
		t.Fatalf("voice release lost or duplicated stock playback history: %+v", got)
	}
	if err := openal.Err(); err != nil {
		t.Fatalf("stock voice release error: %v", err)
	}
	if sha256.Sum256(idx) != indexHash || sha256.Sum256(bag) != bagHash {
		t.Fatal("decoding/playback modified stock bank input bytes")
	}
	for _, file := range []struct {
		path string
		hash [sha256.Size]byte
	}{{indexPath, indexHash}, {bagPath, bagHash}} {
		data, err := os.ReadFile(file.path)
		if err != nil {
			t.Fatal(err)
		}
		if sha256.Sum256(data) != file.hash {
			t.Fatalf("stock bank file changed: %s", file.path)
		}
	}
}
