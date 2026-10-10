//go:build !server

package ail

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/timshannon/go-openal/openal"
)

type mp3FrameReadCloser struct{ *bytes.Reader }

func (mp3FrameReadCloser) Close() error { return nil }

func mp3FrameWAV(t *testing.T, fixture int) []byte {
	t.Helper()
	f := mp3FrameFixtures[fixture]
	data, err := base64.StdEncoding.DecodeString(f.base64)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != f.sha256 {
		t.Fatal("synthetic MP3 fixture hash changed")
	}
	format := make([]byte, 30)
	binary.LittleEndian.PutUint16(format[0:], 0x55)
	binary.LittleEndian.PutUint16(format[2:], uint16(f.channels))
	binary.LittleEndian.PutUint32(format[4:], uint32(f.rate))
	binary.LittleEndian.PutUint32(format[8:], 12000)
	binary.LittleEndian.PutUint16(format[12:], 1)
	binary.LittleEndian.PutUint16(format[16:], 12)
	binary.LittleEndian.PutUint16(format[18:], 1)
	binary.LittleEndian.PutUint16(format[26:], 1)
	var body bytes.Buffer
	body.WriteString("WAVEfmt ")
	_ = binary.Write(&body, binary.LittleEndian, uint32(len(format)))
	body.Write(format)
	body.WriteString("data")
	_ = binary.Write(&body, binary.LittleEndian, uint32(len(data)))
	body.Write(data)
	if len(data)%2 != 0 {
		body.WriteByte(0)
	}
	var wav bytes.Buffer
	wav.WriteString("RIFF")
	_ = binary.Write(&wav, binary.LittleEndian, uint32(body.Len()))
	wav.Write(body.Bytes())
	return wav.Bytes()
}

func openMP3FrameFixture(t *testing.T, data []byte) *wavReader {
	t.Helper()
	r, err := newWavReader(mp3FrameReadCloser{bytes.NewReader(data)}, "mp3-frames-fixture.wav")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(r.Close)
	return r
}

// Use the unchanged minimp3 decoder, a full-sized destination, and its native
// per-channel return count as an independent reference for the stream adapter.
// This does not assert bit-exact agreement with ffmpeg or the PE32 audio engine.
func referenceMP3FramePCM(t *testing.T, data []byte) []int16 {
	t.Helper()
	r := openMP3FrameFixture(t, data)
	var raw *mp3Decoder
	switch d := r.dec.(type) {
	case *mp3Decoder:
		raw = d
	case *mp3FrameDecoder:
		raw = d.mp3Decoder
	default:
		t.Fatalf("expected WAV+MP3 decoder, got %T", r.dec)
	}
	var buffer [1152 * 2]int16
	var result []int16
	for calls := 0; calls < 10000; calls++ {
		n, ok := raw.Decode(buffer[:])
		if n < 0 || n > 1152 || n*r.Channels() > len(buffer) {
			t.Fatalf("invalid native minimp3 frame count %d", n)
		}
		result = append(result, buffer[:n*r.Channels()]...)
		if !ok {
			return result
		}
		if n == 0 {
			t.Fatal("reference made no progress before EOF")
		}
	}
	t.Fatal("reference exceeded finite fixture limit")
	return nil
}

func mp3FramePCMHash(pcm []int16) string {
	var encoded bytes.Buffer
	_ = binary.Write(&encoded, binary.LittleEndian, pcm)
	return fmt.Sprintf("%x", sha256.Sum256(encoded.Bytes()))
}

func readMP3FramePCM(t *testing.T, r audioDecoder) []int16 {
	t.Helper()
	var buffer [1152 * 2]int16
	var result []int16
	initial := r.Position()
	for calls := 0; calls < 10000; calls++ {
		n, ok := r.Decode(buffer[:])
		if n < 0 || n > len(buffer) || n%r.Channels() != 0 {
			t.Fatalf("invalid interleaved sample count: %d", n)
		}
		result = append(result, buffer[:n]...)
		if got, want := r.Position(), initial+len(result)/r.Channels(); got != want {
			t.Fatalf("decoded position=%d, want=%d after %d interleaved samples", got, want, len(result))
		}
		if !ok {
			return result
		}
		if n == 0 {
			t.Fatal("stream made no progress before EOF")
		}
	}
	t.Fatal("stream exceeded finite fixture limit")
	return nil
}

func TestWAVMP3ReferenceFixtures(t *testing.T) {
	for fixture, f := range mp3FrameFixtures {
		t.Run(f.name, func(t *testing.T) {
			pcm := referenceMP3FramePCM(t, mp3FrameWAV(t, fixture))
			nonzero, distinct := false, false
			for i := 0; i < len(pcm); i += f.channels {
				nonzero = nonzero || pcm[i] != 0
				if f.channels == 2 {
					distinct = distinct || pcm[i] != pcm[i+1]
				}
			}
			if len(pcm) == 0 || !nonzero || f.channels == 2 && !distinct {
				t.Fatal("fixture must contain nonzero, distinct channel PCM")
			}
			t.Logf("unchanged minimp3 reference: rate=%d channels=%d frames=%d bytes=%d pcm_sha256=%s", f.rate, f.channels, len(pcm)/f.channels, len(pcm)*2, mp3FramePCMHash(pcm))
		})
	}
}

func TestWAVMP3StreamFramesAndSeek(t *testing.T) {
	for fixture, f := range mp3FrameFixtures {
		t.Run(f.name, func(t *testing.T) {
			data := mp3FrameWAV(t, fixture)
			pcm := referenceMP3FramePCM(t, data)
			frames := len(pcm) / f.channels
			blockFrames := 1152
			if f.rate == 22050 {
				blockFrames = 576
			}
			r := openMP3FrameFixture(t, data)
			t.Run("full-decode", func(t *testing.T) {
				if r.Format() != "WAV+MP3" || r.Channels() != f.channels || r.SampleRate() != f.rate {
					t.Fatal("stream changed the WAV format")
				}
				if got := readMP3FramePCM(t, r); !slices.Equal(got, pcm) {
					t.Fatal("full stream did not preserve all interleaved native PCM")
				}
			})
			// Reuse the same reader after EOF, including a backwards seek.
			positions := []int{0, 1, blockFrames - 1, blockFrames, blockFrames + 1, 2*blockFrames + 3, frames - 1, frames, frames + 23, -7, 1, 0}
			for sequence, offset := range positions {
				t.Run(fmt.Sprintf("seek-%d-at-%d", sequence, offset), func(t *testing.T) {
					want := min(max(offset, 0), frames)
					r.Seek(offset)
					if r.Position() != want {
						t.Fatalf("seek position=%d, want=%d", r.Position(), want)
					}
					if got := readMP3FramePCM(t, r); !slices.Equal(got, pcm[want*f.channels:]) {
						t.Fatalf("seek is not the complete native PCM suffix at frame %d", want)
					}
				})
			}
		})
	}
}

func TestWAVMP3StreamPendingTailAndRewind(t *testing.T) {
	for fixture, f := range mp3FrameFixtures {
		t.Run(f.name, func(t *testing.T) {
			data := mp3FrameWAV(t, fixture)
			pcm := referenceMP3FramePCM(t, data)
			r := openMP3FrameFixture(t, data)
			// Require working Position before testing a small destination. On the
			// unwrapped decoder, C writes a full MP3 frame regardless of its size.
			var bootstrap [1152 * 2]int16
			n, ok := r.Decode(bootstrap[:])
			if !ok || n <= 0 || r.Position() != n/f.channels {
				t.Fatal("stream must report decoded frames before small-buffer checks")
			}
			r.Seek(1)
			if n, ok := r.Decode(nil); n != 0 || !ok || r.Position() != 1 {
				t.Fatal("empty output consumed a frame or reached premature EOF")
			}
			if f.channels == 2 {
				var partial [1]int16
				if n, ok := r.Decode(partial[:]); n != 0 || !ok || r.Position() != 1 {
					t.Fatal("partial stereo frame must not advance position")
				}
			}
			var got []int16
			blockFrames := 1152
			if f.rate == 22050 {
				blockFrames = 576
			}
			for len(got) < (blockFrames-1)*f.channels {
				buffer := make([]int16, 2*f.channels+2)
				buffer[0], buffer[len(buffer)-1] = 777, -777
				n, ok := r.Decode(buffer[1 : len(buffer)-1])
				if !ok || n <= 0 || n > 2*f.channels || n%f.channels != 0 || buffer[0] != 777 || buffer[len(buffer)-1] != -777 {
					t.Fatalf("small output overflow or partial frame: n=%d ok=%t", n, ok)
				}
				got = append(got, buffer[1:1+n]...)
				if r.Position() != 1+len(got)/f.channels {
					t.Fatal("pending PCM changed reported frame position")
				}
			}
			got = append(got, readMP3FramePCM(t, r)...)
			if !slices.Equal(got, pcm[f.channels:]) {
				t.Fatal("pending tail or next MP3 frame was lost or duplicated")
			}
			r.Seek(2)
			r.Seek(1)
			if r.Position() != 1 || !slices.Equal(readMP3FramePCM(t, r), pcm[f.channels:]) {
				t.Fatal("rewind retained a stale pending tail")
			}
			r.Seek(0)
			if r.Position() != 0 || !slices.Equal(readMP3FramePCM(t, r), pcm) {
				t.Fatal("rewind failed to rebuild the MP3 reservoir")
			}
		})
	}
}

func TestWAVMP3StreamStockDialog(t *testing.T) {
	path := os.Getenv("NOX_TEST_MP3_WAV")
	if path == "" {
		t.Skip("set NOX_TEST_MP3_WAV to unchanged stock dialog/c1cap01k.wav")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	before := sha256.Sum256(data)
	r := openMP3FrameFixture(t, data)
	if r.Format() != "WAV+MP3" || r.Channels() != 1 || r.SampleRate() != 22050 {
		t.Fatal("expected stock mono MP3 Korean dialog WAV")
	}
	pcm := referenceMP3FramePCM(t, data)
	if len(pcm) != 449856 || mp3FramePCMHash(pcm) != "defe45e21494c4868ee8852c82da34c9bb9dd86f5e845907f22895572672c9f0" {
		t.Fatal("stock reference changed its established full PCM length/hash")
	}
	for _, offset := range []int{0, 1, 575, 576, 577, 22050, 449855, 449856, 449857} {
		t.Run(fmt.Sprintf("frame-%d", offset), func(t *testing.T) {
			want := min(offset, 449856)
			r.Seek(offset)
			if r.Position() != want || !slices.Equal(readMP3FramePCM(t, r), pcm[want:]) {
				t.Fatalf("stock dialog resume failed at frame %d", want)
			}
		})
	}
	after, err := os.ReadFile(path)
	if err != nil || sha256.Sum256(after) != before {
		t.Fatal("read-only stock dialog test changed its input")
	}
}

func TestOpenALMP3StreamFramesAndSeek(t *testing.T) {
	if os.Getenv("NOX_TEST_OPENAL") != "true" {
		t.Skip("set NOX_TEST_OPENAL=true with ALSOFT_DRIVERS=null for native MP3 streams")
	}
	if os.Getenv("ALSOFT_DRIVERS") != "null" {
		t.Fatal("MP3 stream integration requires isolated OpenAL null output, never speakers")
	}
	for fixture, f := range mp3FrameFixtures {
		t.Run(f.name, func(t *testing.T) {
			h := openPlaybackTestDriver(t)
			data := mp3FrameWAV(t, fixture)
			pcm := referenceMP3FramePCM(t, data)
			frames := len(pcm) / f.channels
			blockFrames := 1152
			if f.rate == 22050 {
				blockFrames = 576
			}
			path := filepath.Join(t.TempDir(), f.name+".wav")
			if err := os.WriteFile(path, data, 0600); err != nil {
				t.Fatal(err)
			}
			stream := h.OpenStream(path)
			if stream == 0 {
				t.Fatal("cannot open native WAV+MP3 stream")
			}
			t.Cleanup(func() { _ = stream.Close() })
			stream.SetVolume(0)
			stream.Start()
			h.get().doWork()
			queued := h.PlaybackStats().StreamBuffersQueued
			if queued == 0 || stream.Position() != frames {
				t.Fatalf("native stream queued=%d position=%d, want=%d frames", queued, stream.Position(), frames)
			}
			// Hardware metadata proves the complete channel count reached the
			// native queue; this is not PCM readback or speaker verification.
			storedBytes := 0
			for _, buffer := range stream.get().hwbuf {
				size := int(buffer.GetSize())
				storedBytes += size
				if size != 0 && (buffer.GetChannels() != uint32(f.channels) || buffer.GetFrequency() != uint32(f.rate) || buffer.GetBits() != 16) {
					t.Fatal("native buffer format does not match decoded PCM")
				}
			}
			if storedBytes != len(pcm)*2 {
				t.Fatalf("native buffers have %d PCM bytes, want=%d", storedBytes, len(pcm)*2)
			}
			for _, offset := range []int{1, blockFrames - 1, blockFrames, blockFrames + 1, frames - 1, frames, 0} {
				t.Run(fmt.Sprintf("frame-%d", offset), func(t *testing.T) {
					stream.SetPosition(offset)
					if stream.Position() != offset || stream.get().eof {
						t.Fatalf("native stream did not restore frame %d and reset EOF", offset)
					}
					if got := readMP3FramePCM(t, stream.get().dec); !slices.Equal(got, pcm[offset*f.channels:]) {
						t.Fatal("native SetPosition did not reach the exact decoder suffix")
					}
					stream.SetPosition(offset)
					stream.Start()
					h.get().doWork()
					if got := openal.Err(); got != nil {
						t.Fatalf("native queue/seek error: %v", got)
					}
				})
			}
			if stats := h.PlaybackStats(); !stats.Native || stats.StreamSources != 1 || stats.StreamBuffersQueued <= queued {
				t.Fatalf("native MP3 seeking did not requeue buffers: %+v", stats)
			}
		})
	}
}
