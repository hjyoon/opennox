//go:build !server && (darwin || linux)

package noxmovie

import (
	"context"
	"image"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/opennox/opennox/v1/client/seat/headless"
	"github.com/opennox/opennox/v1/legacy/client/audio/ail"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/timshannon/go-openal/openal"
)

func TestMovieNaturalAndEarlyClose(t *testing.T) {
	path := os.Getenv("NOX_TEST_MOVIE_FILE")
	if path == "" {
		t.Skip("set NOX_TEST_MOVIE_FILE to unchanged stock WWLOGO.VQA")
	}
	path, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestMovieNaturalAndEarlyCloseChild$", "-test.count=1", "-test.v", "-test.timeout=25s")
	cmd.Dir = t.TempDir()
	for _, kv := range os.Environ() {
		key, _, _ := strings.Cut(kv, "=")
		if strings.HasPrefix(key, "NOX_E2E") || strings.HasPrefix(key, "ALSOFT_") ||
			strings.HasPrefix(key, "NOX_TEST_MOVIE_") || key == "NOX_DATA" || key == "NOX_DEBUG_AUDIO" {
			continue
		}
		cmd.Env = append(cmd.Env, kv)
	}
	cmd.Env = append(cmd.Env, "NOX_TEST_MOVIE_NATURAL_CHILD=true", "NOX_TEST_MOVIE_FILE="+path,
		"ALSOFT_DRIVERS=null", "ALSOFT_CONF="+filepath.Join(cmd.Dir, "no-personal-openal.ini"))
	out, err := cmd.CombinedOutput()
	t.Logf("%s", out)
	if err != nil {
		t.Fatalf("natural/early-close subprocess: %v", err)
	}
}

func TestMovieNaturalAndEarlyCloseChild(t *testing.T) {
	if os.Getenv("NOX_TEST_MOVIE_NATURAL_CHILD") != "true" {
		t.Skip("isolated natural movie subprocess only")
	}
	if os.Getenv("ALSOFT_DRIVERS") != "null" {
		t.Fatal("movie playback must never use speakers")
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	handles.Init()
	defer handles.Release()
	driver := ail.WaveOutOpen()
	if driver == 0 {
		t.Fatal("native OpenAL null driver unavailable")
	}
	defer func() {
		if err := driver.Close(); err != nil {
			t.Error(err)
		}
	}()
	for _, mode := range []string{"natural", "close-before-start"} {
		t.Run(mode, func(t *testing.T) {
			file, err := os.Open(os.Getenv("NOX_TEST_MOVIE_FILE"))
			if err != nil {
				t.Fatal(err)
			}
			tracked := &trackedMovieFile{File: file}
			sc := headless.New(image.Pt(640, 480))
			defer sc.Close()
			p, err := NewPlayerWithHandle(tracked, sc, driver)
			if err != nil {
				tracked.Close()
				t.Fatal(err)
			}
			defer p.Close()
			if p.movie.Header.FramesCount != 130 || p.movie.Header.Fps != 15 {
				t.Fatal("expected unchanged complete stock logo")
			}
			started := time.Now()
			if mode == "close-before-start" {
				p.Close()
			}
			p.Start()
			p.Play()
			elapsed := time.Since(started)
			if tracked.closes.Load() != 1 {
				t.Errorf("file close count=%d, want once", tracked.closes.Load())
			}
			if mode == "natural" {
				if sc.PresentCount() != 130 {
					t.Errorf("complete movie presented=%d frames, want 130", sc.PresentCount())
				}
				if elapsed < 8*time.Second || elapsed > 12*time.Second {
					t.Errorf("stock movie natural duration=%s", elapsed)
				}
				if len(p.audioBuffers) == 0 {
					t.Fatal("natural playback allocated no native buffers")
				}
				for _, b := range p.audioBuffers {
					_ = b.GetSize()
					if err := openal.Err(); err == nil {
						t.Errorf("natural EOF retained buffer %d", b)
					}
				}
			} else if sc.PresentCount() != 0 {
				t.Error("closed movie presented frames")
			}
			p.Close()
			p.Close()
			if got := driver.PlaybackStats(); got.SampleSources != 0 {
				t.Errorf("source survived completion: %+v", got)
			}
			if err := openal.Err(); err != nil {
				t.Errorf("OpenAL error after cleanup: %v", err)
			}
			t.Logf("native %s: frames=%d duration=%s closes=%d", mode, sc.PresentCount(), elapsed, tracked.closes.Load())
		})
	}
}
