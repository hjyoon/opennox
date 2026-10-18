//go:build !server

package ail

import (
	"testing"
	"time"

	"github.com/timshannon/go-openal/openal"
)

// Recycle exactly one real queued buffer, leaving the remaining original
// buffers in the OpenAL source. The free-list slice now has duplicate IDs;
// closing must still delete the immutable set of originally allocated IDs.
func recycleOneNativeCloseBuffer(t *testing.T, source openal.Source, buffers openal.Buffers, ready *int, unqueue func(bool)) []openal.Buffer {
	t.Helper()
	original := append([]openal.Buffer(nil), buffers...)
	for i := len(buffers) - 1; i >= 0; i-- {
		buffers[i].SetDataInt16(openal.FormatMono16, make([]int16, 22050), 22050)
		source.QueueBuffer(buffers[i])
		*ready--
	}
	source.SetGain(0)
	source.Play()
	source.Pause()
	source.Seti(openal.AlSampleOffset, 22051)
	if err := openal.Err(); err != nil {
		t.Fatalf("prepare real partial queue: %v", err)
	}
	if got := source.BuffersProcessed(); got != 1 {
		t.Fatalf("partially consumed buffers = %d, want exactly 1", got)
	}
	unqueue(true)
	if *ready != 1 {
		t.Fatalf("free pool = %d, want 1", *ready)
	}
	*ready--
	source.QueueBuffer(buffers[*ready])
	if err := openal.Err(); err != nil {
		t.Fatalf("requeue native free buffer: %v", err)
	}
	seen := make(map[openal.Buffer]bool)
	duplicate := false
	for _, b := range buffers {
		duplicate = duplicate || seen[b]
		seen[b] = true
	}
	if !duplicate {
		t.Fatal("fixture did not retain the production free-list alias state")
	}
	return original
}

func assertNativeCloseBuffersDeleted(t *testing.T, buffers []openal.Buffer) {
	t.Helper()
	for _, b := range buffers {
		_ = openal.Err()
		size := b.GetSize()
		if err := openal.Err(); err == nil {
			t.Errorf("owned OpenAL buffer %d survived close with size %d", b, size)
		}
	}
}

func TestOpenALRecycledSampleReleaseDeletesEveryBuffer(t *testing.T) {
	driver := openPlaybackTestDriver(t)
	handle := driver.AllocateSample()
	sample := handle.get()
	if sample == nil {
		t.Fatal("cannot allocate native sample")
	}
	buffers := recycleOneNativeCloseBuffer(t, sample.source, sample.hwbuf, &sample.hwready, sample.unqueueBuffers)
	handle.Release()
	assertNativeCloseBuffersDeleted(t, buffers)
}

func TestOpenALRecycledStreamCloseDeletesEveryBuffer(t *testing.T) {
	driver := openPlaybackTestDriver(t)
	handle := driver.OpenStream(silentPlaybackWAV(t, time.Second))
	stream := handle.get()
	if stream == nil {
		t.Fatal("cannot allocate native stream")
	}
	buffers := recycleOneNativeCloseBuffer(t, stream.source, stream.hwbuf, &stream.hwready, stream.unqueueBuffers)
	if err := handle.Close(); err != nil {
		t.Fatal(err)
	}
	assertNativeCloseBuffersDeleted(t, buffers)
}
