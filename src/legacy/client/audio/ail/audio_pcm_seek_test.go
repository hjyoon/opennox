//go:build !server

package ail

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/timshannon/go-openal/openal"
)

type pcmSeekReadCloser struct{ *bytes.Reader }

func (pcmSeekReadCloser) Close() error { return nil }

// Generate signed 16-bit samples independently of the decoder. Neither the
// expected suffix nor the seek position is derived from production output.
func pcmSeekWAV(t *testing.T, channels, rate int, split bool) ([]byte, []int16) {
	t.Helper()
	const frames = 8197
	pcm := make([]int16, frames*channels)
	for frame := 0; frame < frames; frame++ {
		for channel := 0; channel < channels; channel++ {
			pcm[frame*channels+channel] = int16((frame*311+channel*197+17)%65536 - 32768)
		}
	}
	var encoded bytes.Buffer
	if err := binary.Write(&encoded, binary.LittleEndian, pcm); err != nil {
		t.Fatal(err)
	}
	var body bytes.Buffer
	body.WriteString("WAVE")
	chunk := func(name string, payload []byte) {
		body.WriteString(name)
		_ = binary.Write(&body, binary.LittleEndian, uint32(len(payload)))
		body.Write(payload)
	}
	format := make([]byte, 16)
	binary.LittleEndian.PutUint16(format[0:], 1)
	binary.LittleEndian.PutUint16(format[2:], uint16(channels))
	binary.LittleEndian.PutUint32(format[4:], uint32(rate))
	binary.LittleEndian.PutUint32(format[8:], uint32(rate*channels*2))
	binary.LittleEndian.PutUint16(format[12:], uint16(channels*2))
	binary.LittleEndian.PutUint16(format[14:], 16)
	chunk("fmt ", format)
	if split {
		// Existing WAV decoding supports frame-aligned data chunks. Include
		// metadata between them without extending that admission contract.
		at := 2053 * channels * 2
		chunk("data", encoded.Bytes()[:at])
		chunk("JUNK", []byte{1, 2, 3, 4})
		chunk("data", encoded.Bytes()[at:])
	} else {
		chunk("data", encoded.Bytes())
	}
	var wav bytes.Buffer
	wav.WriteString("RIFF")
	_ = binary.Write(&wav, binary.LittleEndian, uint32(body.Len()))
	wav.Write(body.Bytes())
	return wav.Bytes(), pcm
}

func openPCMSeekFixture(t *testing.T, data []byte) *wavReader {
	t.Helper()
	r, err := newWavReader(pcmSeekReadCloser{bytes.NewReader(data)}, "pcm-seek-fixture.wav")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(r.Close)
	return r
}

func readPCMSeekSuffix(t *testing.T, r audioDecoder) []int16 {
	t.Helper()
	initial := r.Position()
	var buffer [2048]int16
	var result []int16
	for calls := 0; calls < 10000; calls++ {
		n, ok := r.Decode(buffer[:])
		if n < 0 || n > len(buffer) || n%r.Channels() != 0 {
			t.Fatalf("invalid complete-frame PCM count %d", n)
		}
		result = append(result, buffer[:n]...)
		if r.Position() != initial+len(result)/r.Channels() {
			t.Fatal("delivered PCM frame position diverged from samples")
		}
		if !ok {
			return result
		}
		if n == 0 {
			t.Fatal("PCM made no progress before EOF")
		}
	}
	t.Fatal("PCM fixture exceeded finite decode limit")
	return nil
}

func TestWAVPCMSeekExactFrames(t *testing.T) {
	for _, channels := range []int{1, 2} {
		for _, rate := range []int{22050, 44100} {
			for _, split := range []bool{false, true} {
				t.Run(fmt.Sprintf("channels-%d-rate-%d-split-%t", channels, rate, split), func(t *testing.T) {
					data, pcm := pcmSeekWAV(t, channels, rate, split)
					frames := len(pcm) / channels
					r := openPCMSeekFixture(t, data)
					t.Run("full-decode", func(t *testing.T) {
						if r.Format() != "WAV+PCM" || r.Channels() != channels || r.SampleRate() != rate || !slices.Equal(readPCMSeekSuffix(t, r), pcm) {
							t.Fatal("unchanged PCM decoding must preserve the generated samples and format")
						}
					})
					// Reuse the same reader after EOF, across the data boundary,
					// and for huge offsets without multiplying MaxInt by channels.
					positions := []int{0, 1, 2052, 2053, 2054, 4097, frames - 1, frames, frames + 31, int(^uint(0) >> 1), -9, 1, 0}
					for sequence, offset := range positions {
						t.Run(fmt.Sprintf("seek-%d-at-%d", sequence, offset), func(t *testing.T) {
							want := min(max(offset, 0), frames)
							r.Seek(offset)
							if got := r.Position(); got != want {
								t.Fatalf("PCM seek position=%d, want=%d", got, want)
							}
							if got := readPCMSeekSuffix(t, r); !slices.Equal(got, pcm[want*channels:]) {
								t.Fatalf("PCM seek must return the entire independent suffix at frame %d", want)
							}
						})
					}
				})
			}
		}
	}
}

func TestWAVPCMSeekAfterBufferedRead(t *testing.T) {
	for _, channels := range []int{1, 2} {
		for _, split := range []bool{false, true} {
			t.Run(fmt.Sprintf("channels-%d-split-%t", channels, split), func(t *testing.T) {
				data, pcm := pcmSeekWAV(t, channels, 22050, split)
				r := openPCMSeekFixture(t, data)
				out := make([]int16, 7*channels)
				if n, ok := r.Decode(out); !ok || n != len(out) || r.Position() != 7 || !slices.Equal(out, pcm[:len(out)]) {
					t.Fatal("initial ordinary read must fill the WAV read-ahead buffer")
				}
				for _, offset := range []int{1, 2054, 2, 0} {
					r.Seek(offset)
					if r.Position() != offset {
						t.Fatalf("buffered seek position=%d, want=%d", r.Position(), offset)
					}
					buffer := make([]int16, 2*channels+2)
					buffer[0], buffer[len(buffer)-1] = 777, -777
					n, ok := r.Decode(buffer[1 : len(buffer)-1])
					if !ok || n != 2*channels || r.Position() != offset+2 || buffer[0] != 777 || buffer[len(buffer)-1] != -777 || !slices.Equal(buffer[1:1+n], pcm[offset*channels:(offset+2)*channels]) {
						t.Fatal("seek retained stale read-ahead PCM or overflowed the destination")
					}
				}
				if n, ok := r.Decode(nil); n != 0 || !ok {
					t.Fatal("empty destination consumed PCM or reached premature EOF")
				}
			})
		}
	}
}

func TestWAVPCMSeekClampsTruncatedData(t *testing.T) {
	for _, channels := range []int{1, 2} {
		t.Run(fmt.Sprintf("channels-%d", channels), func(t *testing.T) {
			data, pcm := pcmSeekWAV(t, channels, 22050, false)
			const presentFrames = 19
			data = data[:44+presentFrames*channels*2]
			r := openPCMSeekFixture(t, data)
			r.Seek(100)
			if r.Position() != presentFrames {
				t.Fatalf("truncated data position=%d, want=%d actually available frames", r.Position(), presentFrames)
			}
			if n, ok := r.Decode(make([]int16, 2*channels)); n != 0 || ok {
				t.Fatal("truncated PCM must stop at the actual data end")
			}
			r.Seek(1)
			if r.Position() != 1 || !slices.Equal(readPCMSeekSuffix(t, r), pcm[channels:presentFrames*channels]) {
				t.Fatal("truncated reader could not resume after EOF")
			}
		})
	}
}

func TestOpenALPCMStreamFramesAndSeek(t *testing.T) {
	if os.Getenv("NOX_TEST_OPENAL") != "true" {
		t.Skip("set NOX_TEST_OPENAL=true with ALSOFT_DRIVERS=null for native PCM streams")
	}
	if os.Getenv("ALSOFT_DRIVERS") != "null" {
		t.Fatal("PCM stream integration requires isolated OpenAL null output, never speakers")
	}
	for _, channels := range []int{1, 2} {
		t.Run(fmt.Sprintf("channels-%d", channels), func(t *testing.T) {
			h := openPlaybackTestDriver(t)
			data, pcm := pcmSeekWAV(t, channels, 22050, true)
			frames := len(pcm) / channels
			path := filepath.Join(t.TempDir(), "pcm-seek.wav")
			if err := os.WriteFile(path, data, 0600); err != nil {
				t.Fatal(err)
			}
			stream := h.OpenStream(path)
			if stream == 0 {
				t.Fatal("cannot open native PCM WAV stream")
			}
			t.Cleanup(func() { _ = stream.Close() })
			stream.SetVolume(0)
			stream.Start()
			h.get().doWork()
			queued := h.PlaybackStats().StreamBuffersQueued
			if queued == 0 || stream.Position() != frames {
				t.Fatal("native PCM stream did not queue all frames")
			}
			storedBytes := 0
			for _, buffer := range stream.get().hwbuf {
				size := int(buffer.GetSize())
				storedBytes += size
				if size != 0 && (buffer.GetChannels() != uint32(channels) || buffer.GetFrequency() != 22050 || buffer.GetBits() != 16) {
					t.Fatal("native PCM buffer format changed")
				}
			}
			if storedBytes != len(pcm)*2 {
				t.Fatal("native PCM queue did not preserve the full interleaved sample byte count")
			}
			for _, offset := range []int{1, 2052, 2053, 2054, frames - 1, frames, 0} {
				t.Run(fmt.Sprintf("frame-%d", offset), func(t *testing.T) {
					stream.SetPosition(offset)
					if stream.Position() != offset || stream.get().eof {
						t.Fatal("native PCM SetPosition failed to restore the frame and reset EOF")
					}
					if got := readPCMSeekSuffix(t, stream.get().dec); !slices.Equal(got, pcm[offset*channels:]) {
						t.Fatal("native PCM SetPosition did not reach the independent decoder suffix")
					}
					stream.SetPosition(offset)
					stream.Start()
					h.get().doWork()
					if err := openal.Err(); err != nil {
						t.Fatalf("native PCM queue/seek error: %v", err)
					}
				})
			}
			if stats := h.PlaybackStats(); !stats.Native || stats.StreamSources != 1 || stats.StreamBuffersQueued <= queued {
				t.Fatal("native PCM seeking did not requeue buffers")
			}
		})
	}
}
