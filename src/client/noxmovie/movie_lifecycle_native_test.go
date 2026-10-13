//go:build !server && (darwin || linux)

package noxmovie

import (
	"context"
	"image"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/opennox/libs/client/keybind"
	libseat "github.com/opennox/libs/client/seat"
	"github.com/opennox/opennox/v1/client/seat/headless"
	"github.com/opennox/opennox/v1/legacy/client/audio/ail"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/timshannon/go-openal/openal"
	"golang.org/x/sys/unix"
)

func descriptorsForMovie(t *testing.T, path string) int {
	t.Helper()
	var want unix.Stat_t
	if err := unix.Stat(path, &want); err != nil {
		t.Fatal(err)
	}
	dir, err := os.Open("/dev/fd")
	if err != nil {
		t.Fatal(err)
	}
	entries, err := dir.Readdirnames(-1)
	dir.Close()
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, entry := range entries {
		fd, err := strconv.Atoi(entry)
		if err != nil {
			continue
		}
		var got unix.Stat_t
		if unix.Fstat(fd, &got) == nil && got.Dev == want.Dev && got.Ino == want.Ino {
			count++
		}
	}
	return count
}

// A failed constructor must close the file it opened, independently of GC.
func TestMovieInvalidFileDoesNotLeakDescriptors(t *testing.T) {
	path := filepath.Join(t.TempDir(), "invalid.vqa")
	if err := os.WriteFile(path, []byte("not a VQA movie"), 0600); err != nil {
		t.Fatal(err)
	}
	sc := headless.New(image.Pt(640, 480))
	defer sc.Close()
	before := descriptorsForMovie(t, path)
	for i := 0; i < 8; i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			player, err := NewPlayer(path, sc, 0)
			if err == nil || player != nil {
				t.Fatal("invalid file was admitted")
			}
			if got := descriptorsForMovie(t, path); got != before {
				t.Errorf("failed movie open leaked descriptors: before=%d after=%d", before, got)
			}
		})
	}
}

const movieLifecycleChild = "NOX_TEST_MOVIE_LIFECYCLE_CHILD"

// Real VQA, real OpenAL null output, and real headless input delivery. No
// replacement decoder, shortened movie, playback acceleration or result hook.
func TestMovieNativeInputCancellation(t *testing.T) {
	path := os.Getenv("NOX_TEST_MOVIE_FILE")
	if path == "" {
		t.Skip("set NOX_TEST_MOVIE_FILE to unchanged stock MOVIES/WWLOGO.VQA")
	}
	path, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestMovieNativeInputCancellationChild$", "-test.count=1", "-test.v", "-test.timeout=55s")
	cmd.Dir = t.TempDir()
	for _, kv := range os.Environ() {
		key, _, _ := strings.Cut(kv, "=")
		if strings.HasPrefix(key, "NOX_E2E") || strings.HasPrefix(key, "ALSOFT_") ||
			strings.HasPrefix(key, "NOX_TEST_MOVIE_") || key == "NOX_DATA" || key == "NOX_DEBUG_AUDIO" {
			continue
		}
		cmd.Env = append(cmd.Env, kv)
	}
	cmd.Env = append(cmd.Env, movieLifecycleChild+"=true", "NOX_TEST_MOVIE_FILE="+path,
		"ALSOFT_DRIVERS=null", "ALSOFT_CONF="+filepath.Join(cmd.Dir, "no-personal-openal.ini"))
	out, err := cmd.CombinedOutput()
	t.Logf("%s", out)
	if err != nil {
		t.Fatalf("native movie subprocess: %v", err)
	}
}

type trackedMovieFile struct {
	*os.File
	closes atomic.Int32
}

func (f *trackedMovieFile) Close() error {
	f.closes.Add(1)
	return f.File.Close()
}

func TestMovieNativeInputCancellationChild(t *testing.T) {
	if os.Getenv(movieLifecycleChild) != "true" {
		t.Skip("isolated movie subprocess only")
	}
	if os.Getenv("ALSOFT_DRIVERS") != "null" {
		t.Fatal("movie test must never use speakers")
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	handles.Init()
	defer handles.Release()
	driver := ail.WaveOutOpen()
	if driver == 0 {
		t.Fatal("native OpenAL null device unavailable")
	}
	defer func() {
		if err := driver.Close(); err != nil {
			t.Error(err)
		}
	}()
	events := []struct {
		name  string
		event libseat.InputEvent
	}{
		{"mouse", &libseat.MouseButtonEvent{Button: libseat.MouseButtonLeft, Pressed: true}},
		{"keyboard", &libseat.KeyboardEvent{Key: keybind.KeyEsc, Pressed: true}},
		{"window", libseat.WindowClosed},
	}
	path := os.Getenv("NOX_TEST_MOVIE_FILE")
	for _, tc := range events {
		t.Run(tc.name, func(t *testing.T) {
			file, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			tracked := &trackedMovieFile{File: file}
			sc := headless.New(image.Pt(640, 480))
			defer sc.Close()
			var restored atomic.Int32
			sc.OnInput(func(libseat.InputEvent) { restored.Add(1) })
			player, err := NewPlayerWithHandle(tracked, sc, driver)
			if err != nil {
				tracked.Close()
				t.Fatal(err)
			}
			defer player.Close()
			if player.movie.Header.FramesCount != 130 || player.movie.Header.Fps != 15 || player.movie.Header.SampleRate != 22050 {
				t.Fatal("expected unchanged stock WWLOGO")
			}
			if stats := driver.PlaybackStats(); !stats.Native || stats.SampleSources != 1 {
				t.Fatalf("not a real movie OpenAL source: %+v", stats)
			}
			injected := make(chan bool, 1)
			go func() {
				deadline := time.Now().Add(4 * time.Second)
				for sc.PresentCount() < 5 && time.Now().Before(deadline) {
					time.Sleep(time.Millisecond)
				}
				if sc.PresentCount() < 5 {
					injected <- false
					return
				}
				sc.QueueInput(tc.event)
				injected <- true
			}()
			started := time.Now()
			player.Start()
			player.Play()
			if !<-injected {
				t.Fatal("movie did not present five real frames")
			}
			if elapsed := time.Since(started); elapsed > 4*time.Second {
				t.Fatalf("input did not cancel playback promptly: %s", elapsed)
			}
			if n := sc.PresentCount(); n < 5 || n >= 130 {
				t.Fatalf("invalid cancellation presentation count %d", n)
			}
			if tracked.closes.Load() != 1 {
				t.Errorf("movie file close count=%d, want 1", tracked.closes.Load())
			}
			// Play has joined its audio worker. Look at every real buffer, including
			// one allocated during input delivery; deleting only an earlier slice leaks.
			buffers := append(openal.Buffers(nil), player.audioBuffers...)
			player.Close()
			player.Close()
			if stats := driver.PlaybackStats(); stats.SampleSources != 0 {
				t.Errorf("movie source leaked: %+v", stats)
			}
			if len(buffers) == 0 {
				t.Fatal("no real movie buffers were allocated")
			}
			for _, buf := range buffers {
				_ = buf.GetSize()
				if err := openal.Err(); err == nil {
					t.Errorf("movie buffer %d remains valid after Close", buf)
				}
			}
			if err := openal.Err(); err != nil {
				t.Errorf("unconsumed OpenAL error: %v", err)
			}
			before := restored.Load()
			sc.QueueInput(&libseat.KeyboardEvent{Key: keybind.KeyEsc, Pressed: false})
			sc.InputTick()
			if restored.Load() != before+1 {
				t.Fatal("previous input handlers were not restored")
			}
			t.Logf("real cancellation %s presented=%d buffers=%d closes=%d", tc.name, sc.PresentCount(), len(buffers), tracked.closes.Load())
		})
	}
}
