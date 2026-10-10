//go:build !server

package ail

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/timshannon/go-openal/openal"
)

type adpcmSeekReadCloser struct{ *bytes.Reader }

func (adpcmSeekReadCloser) Close() error { return nil }

// Distinct predictors and nibbles make restarting at a block boundary visibly
// different from the requested PCM frame. The last block is fact-trimmed.
func adpcmSeekWAV(t *testing.T, channels, blockSize int, split bool) ([]byte, int, int) {
	t.Helper()
	framesPerBlock := (blockSize/channels-4)*2 + 1
	frames := 3*framesPerBlock - 5
	data := make([]byte, 3*blockSize)
	for block := 0; block < 3; block++ {
		part := data[block*blockSize : (block+1)*blockSize]
		for i := range part {
			part[i] = byte(17*block + 13*i + 1)
		}
		for channel := 0; channel < channels; channel++ {
			binary.LittleEndian.PutUint16(part[4*channel:], uint16(1000*block+200*channel+73))
			part[4*channel+2] = byte(2 + block)
			part[4*channel+3] = 0
		}
	}
	var body bytes.Buffer
	body.WriteString("WAVE")
	chunk := func(name string, data []byte) {
		body.WriteString(name)
		_ = binary.Write(&body, binary.LittleEndian, uint32(len(data)))
		body.Write(data)
	}
	format := make([]byte, 20)
	binary.LittleEndian.PutUint16(format[0:], 0x11)
	binary.LittleEndian.PutUint16(format[2:], uint16(channels))
	binary.LittleEndian.PutUint32(format[4:], 22050)
	binary.LittleEndian.PutUint32(format[8:], uint32(22050*blockSize/framesPerBlock))
	binary.LittleEndian.PutUint16(format[12:], uint16(blockSize))
	binary.LittleEndian.PutUint16(format[14:], 4)
	binary.LittleEndian.PutUint16(format[16:], 2)
	binary.LittleEndian.PutUint16(format[18:], uint16(framesPerBlock))
	chunk("fmt ", format)
	fact := make([]byte, 4)
	binary.LittleEndian.PutUint32(fact, uint32(frames))
	chunk("fact", fact)
	if split {
		// Two block-aligned data chunks, as required by this decoder.
		// Metadata is skipped by the ordinary WAV reader, not the fixture.
		chunk("data", data[:blockSize])
		chunk("JUNK", []byte{1, 2, 3, 4})
		chunk("data", data[blockSize:])
	} else {
		chunk("data", data)
	}
	var wav bytes.Buffer
	wav.WriteString("RIFF")
	_ = binary.Write(&wav, binary.LittleEndian, uint32(body.Len()))
	wav.Write(body.Bytes())
	return wav.Bytes(), frames, framesPerBlock
}

func openADPCMSeekFixture(t *testing.T, data []byte) *wavReader {
	t.Helper()
	r, err := newWavReader(adpcmSeekReadCloser{bytes.NewReader(data)}, "adpcm-seek-fixture.wav")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(r.Close)
	return r
}

func readADPCMSeekPCM(t *testing.T, r audioDecoder) []int16 {
	t.Helper()
	buffer := make([]int16, 16*1024*3)
	var result []int16
	initial := r.Position()
	for calls := 0; ; calls++ {
		n, ok := r.Decode(buffer)
		if n < 0 || n > len(buffer) || n%r.Channels() != 0 || calls > 10000 {
			t.Fatalf("invalid decoder progress: count=%d channels=%d calls=%d", n, r.Channels(), calls)
		}
		result = append(result, buffer[:n]...)
		if got, want := r.Position(), initial+len(result)/r.Channels(); got != want {
			t.Fatalf("position=%d, want=%d after %d interleaved samples", got, want, len(result))
		}
		if !ok {
			return result
		}
		if n == 0 {
			t.Fatal("decoder made no progress before EOF")
		}
	}
}

func TestWAVADPCMSeekExactFrames(t *testing.T) {
	for _, channels := range []int{1, 2} {
		t.Run(fmt.Sprintf("channels-%d", channels), func(t *testing.T) {
			for _, split := range []bool{false, true} {
				t.Run(fmt.Sprintf("split-%t", split), func(t *testing.T) {
					data, frames, blockFrames := adpcmSeekWAV(t, channels, 32, split)
					pcm := readADPCMSeekPCM(t, openADPCMSeekFixture(t, data))
					if len(pcm) != frames*channels {
						t.Fatal("fixture must decode its fact length before any seek")
					}
					r := openADPCMSeekFixture(t, data)
					// Reuse one reader after EOF, including a pending-tail rewind.
					positions := []int{0, 1, blockFrames - 1, blockFrames, blockFrames + 1, 2*blockFrames + 3, frames - 1, frames, frames + 17, -7, 1, 0}
					for sequence, offset := range positions {
						t.Run(fmt.Sprintf("seek-%d-at-%d", sequence, offset), func(t *testing.T) {
							want := min(max(offset, 0), frames)
							r.Seek(offset)
							if got := r.Position(); got != want {
								t.Fatalf("seek position=%d, want=%d", got, want)
							}
							if got := readADPCMSeekPCM(t, r); !slices.Equal(got, pcm[want*channels:]) {
								t.Fatalf("seek PCM is not the unchanged full decode suffix at frame %d", want)
							}
						})
					}
				})
			}
		})
	}
}

func TestWAVADPCMSeekPreservesBlockTail(t *testing.T) {
	for _, channels := range []int{1, 2} {
		t.Run(fmt.Sprintf("channels-%d", channels), func(t *testing.T) {
			data, _, blockFrames := adpcmSeekWAV(t, channels, 32, false)
			pcm := readADPCMSeekPCM(t, openADPCMSeekFixture(t, data))
			r := openADPCMSeekFixture(t, data)
			r.Seek(1)
			var got []int16
			for len(got) < (blockFrames-1)*channels {
				buffer := make([]int16, 2*channels)
				n, ok := r.Decode(buffer)
				if !ok || n <= 0 || n > len(buffer) || n%channels != 0 {
					t.Fatalf("pending tail did not respect output capacity: n=%d ok=%t", n, ok)
				}
				got = append(got, buffer[:n]...)
				if r.Position() != 1+len(got)/channels {
					t.Fatal("pending PCM changed the reported frame position")
				}
			}
			got = append(got, readADPCMSeekPCM(t, r)...)
			if !slices.Equal(got, pcm[channels:]) {
				t.Fatal("pending tail or its following ADPCM block was lost or duplicated")
			}
			// Rewind before consuming the next pending block, not only at EOF.
			r.Seek(2)
			r.Seek(1)
			if r.Position() != 1 || !slices.Equal(readADPCMSeekPCM(t, r), pcm[channels:]) {
				t.Fatal("rewind retained a stale pending block")
			}
		})
	}
}

func TestWAVADPCMSeekStockChap1(t *testing.T) {
	path := os.Getenv("NOX_TEST_ADPCM_WAV")
	if path == "" {
		t.Skip("set NOX_TEST_ADPCM_WAV to unchanged stock music/chap1.wav")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	before := sha256.Sum256(data)
	r := openADPCMSeekFixture(t, data)
	if r.Format() != "WAV+ADPCM" || r.Channels() != 2 || r.SampleRate() != 22050 {
		t.Fatal("expected stock stereo ADPCM chap1.wav")
	}
	pcm := readADPCMSeekPCM(t, r)
	var encoded bytes.Buffer
	_ = binary.Write(&encoded, binary.LittleEndian, pcm)
	if len(pcm) != 747519*2 || fmt.Sprintf("%x", sha256.Sum256(encoded.Bytes())) != "bf1193102407cfb80f4fa2c7814594a9e56518ddcbf1f2841d7f31fed143ac10" {
		t.Fatal("stock full decode must retain the established PCM length/hash")
	}
	for _, offset := range []int{0, 1, 2040, 2041, 2042, 22050, 725469, 747518, 747519, 747520} {
		t.Run(fmt.Sprintf("frame-%d", offset), func(t *testing.T) {
			want := min(offset, 747519)
			r.Seek(offset)
			if r.Position() != want || !slices.Equal(readADPCMSeekPCM(t, r), pcm[want*2:]) {
				t.Fatalf("stock resume did not preserve the full-decode suffix at frame %d", want)
			}
		})
	}
	after, err := os.ReadFile(path)
	if err != nil || sha256.Sum256(after) != before {
		t.Fatal("read-only stock seek test changed its input")
	}
}

func TestOpenALADPCMStreamSetPosition(t *testing.T) {
	if os.Getenv("NOX_TEST_OPENAL") != "true" {
		t.Skip("set NOX_TEST_OPENAL=true with ALSOFT_DRIVERS=null for actual stream seeking")
	}
	if os.Getenv("ALSOFT_DRIVERS") != "null" {
		t.Fatal("ADPCM seek integration requires isolated OpenAL null output, never speakers")
	}
	h := openPlaybackTestDriver(t)
	data, frames, blockFrames := adpcmSeekWAV(t, 2, 256, false)
	pcm := readADPCMSeekPCM(t, openADPCMSeekFixture(t, data))
	path := filepath.Join(t.TempDir(), "seek-stereo.wav")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	stream := h.OpenStream(path)
	if stream == 0 {
		t.Fatal("cannot open actual stereo ADPCM stream")
	}
	t.Cleanup(func() { _ = stream.Close() })
	stream.SetVolume(0)
	stream.Start()
	h.get().doWork()
	queued := h.PlaybackStats().StreamBuffersQueued
	if queued == 0 {
		t.Fatal("fixture did not queue actual native stream buffers")
	}
	for _, offset := range []int{1, blockFrames - 1, blockFrames, blockFrames + 1, frames - 1, frames, 0} {
		t.Run(fmt.Sprintf("frame-%d", offset), func(t *testing.T) {
			stream.SetPosition(offset)
			if stream.Position() != offset || stream.get().eof {
				t.Fatalf("actual stream did not restore frame %d and reset EOF", offset)
			}
			if got := readADPCMSeekPCM(t, stream.get().dec); !slices.Equal(got, pcm[offset*2:]) {
				t.Fatal("Stream.SetPosition did not reach the exact decoder suffix")
			}
			stream.SetPosition(offset)
			stream.Start()
			h.get().doWork()
			if got := openal.Err(); got != nil {
				t.Fatalf("actual native queue/seek error: %v", got)
			}
		})
	}
	if stats := h.PlaybackStats(); !stats.Native || stats.StreamBuffersQueued <= queued || stats.StreamSources != 1 {
		t.Fatalf("actual stream seeking did not requeue native buffers: %+v", stats)
	}
}
