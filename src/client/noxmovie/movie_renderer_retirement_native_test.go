//go:build !server && (darwin || linux)

package noxmovie

import (
	"context"
	"image"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	libseat "github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/noximage"
	"github.com/opennox/opennox/v1/client/render"
	"github.com/opennox/opennox/v1/client/seat/headless"
	"github.com/opennox/opennox/v1/legacy/client/audio/ail"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
)

// Record ownership operations while every rendering/input operation still
// delegates to the production headless backend, not a replacement renderer.
type observedMovieSurface struct {
	libseat.Surface
	destroyed atomic.Int32
}

func (s *observedMovieSurface) Destroy() { s.destroyed.Add(1); s.Surface.Destroy() }

type observedMovieSeat struct {
	*headless.Seat
	surfaces []*observedMovieSurface
}

func (s *observedMovieSeat) NewSurface(sz image.Point, filter bool) libseat.Surface {
	surface := &observedMovieSurface{Surface: s.Seat.NewSurface(sz, filter)}
	s.surfaces = append(s.surfaces, surface)
	return surface
}

func TestMovieRendererRetirement(t *testing.T) {
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
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestMovieRendererRetirementChild$", "-test.count=1", "-test.v", "-test.timeout=25s")
	cmd.Dir = t.TempDir()
	for _, kv := range os.Environ() {
		key, _, _ := strings.Cut(kv, "=")
		if strings.HasPrefix(key, "NOX_E2E") || strings.HasPrefix(key, "ALSOFT_") ||
			strings.HasPrefix(key, "NOX_TEST_MOVIE_") || key == "NOX_DATA" || key == "NOX_DEBUG_AUDIO" {
			continue
		}
		cmd.Env = append(cmd.Env, kv)
	}
	cmd.Env = append(cmd.Env, "NOX_TEST_MOVIE_RENDERER_CHILD=true", "NOX_TEST_MOVIE_FILE="+path,
		"ALSOFT_DRIVERS=null", "ALSOFT_CONF="+filepath.Join(cmd.Dir, "no-personal-openal.ini"))
	out, err := cmd.CombinedOutput()
	t.Logf("%s", out)
	if err != nil {
		t.Fatalf("movie renderer subprocess: %v", err)
	}
}

func TestMovieRendererRetirementChild(t *testing.T) {
	if os.Getenv("NOX_TEST_MOVIE_RENDERER_CHILD") != "true" {
		t.Skip("isolated movie renderer subprocess only")
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
		t.Fatal("real OpenAL null driver unavailable")
	}
	defer func() {
		if err := driver.Close(); err != nil {
			t.Error(err)
		}
	}()
	for _, mode := range []string{"cancel", "close-before-start", "constructor-no-audio-source"} {
		t.Run(mode, func(t *testing.T) {
			sc := &observedMovieSeat{Seat: headless.New(image.Pt(640, 480))}
			defer sc.Close()
			active, err := render.New(sc)
			if err != nil {
				t.Fatal(err)
			}
			frame := noximage.NewImage16(image.Rect(0, 0, 640, 480))
			for i := range frame.Pix {
				frame.Pix[i] = uint16((i*31 + 0x1234) & 0x7fff)
			}
			active.CopyBuffer(frame)
			beforeSurfaces := len(sc.surfaces)
			f, err := os.Open(os.Getenv("NOX_TEST_MOVIE_FILE"))
			if err != nil {
				t.Fatal(err)
			}
			file := &trackedMovieFile{File: f}
			selected := driver
			if mode == "constructor-no-audio-source" {
				selected = 0
			}
			p, err := NewPlayerWithHandle(file, sc, selected)
			var retired *render.Renderer
			if mode == "constructor-no-audio-source" {
				if err == nil || p != nil {
					t.Fatal("missing audio source should reject the constructor")
				}
				if file.closes.Load() != 0 {
					t.Error("borrowed file ownership changed on constructor error")
				}
				file.Close()
			} else {
				if err != nil {
					file.Close()
					t.Fatal(err)
				}
				retired = p.renderer
				if mode == "cancel" {
					ready := make(chan bool, 1)
					go func() {
						deadline := time.Now().Add(4 * time.Second)
						for sc.PresentCount() < 6 && time.Now().Before(deadline) {
							time.Sleep(time.Millisecond)
						}
						if sc.PresentCount() < 6 {
							ready <- false
							return
						}
						sc.QueueInput(&libseat.MouseButtonEvent{Button: libseat.MouseButtonLeft, Pressed: true})
						ready <- true
					}()
					p.Start()
					p.Play()
					if !<-ready {
						t.Fatal("original movie failed to render five frames")
					}
				}
				p.Close()
				p.Close()
				if file.closes.Load() != 1 {
					t.Error("movie file was not closed exactly once")
				}
			}
			movieSurfaces := sc.surfaces[beforeSurfaces:]
			if len(movieSurfaces) == 0 {
				t.Fatal("no real movie surface was created")
			}
			for i, surface := range movieSurfaces {
				if n := surface.destroyed.Load(); n != 1 {
					t.Errorf("movie surface %d destroy count=%d, want 1", i, n)
				}
			}
			active.CopyBuffer(frame)
			var retiredTicks uint
			if retired != nil {
				retiredTicks = retired.Ticks()
			}
			for _, size := range []image.Point{{800, 600}, {1920, 1080}, {640, 480}} {
				beforePresent := sc.PresentCount()
				sc.ResizeScreen(size)
				actual, view := sc.Snapshot()
				control := headless.New(size)
				controlRenderer, err := render.New(control)
				if err != nil {
					t.Fatal(err)
				}
				controlRenderer.CopyBuffer(frame)
				want, wantView := control.Snapshot()
				if actual == nil || actual.Rect != want.Rect || !slices.Equal(actual.Pix, want.Pix) || view != wantView {
					t.Errorf("retired movie overwrote active GUI after resize to %v: actual-view=%v want=%v", size, view, wantView)
				}
				control.Close()
				if n := sc.PresentCount() - beforePresent; n != 1 {
					t.Errorf("resize presented %d renderers, want only active renderer", n)
				}
				if retired != nil && retired.Ticks() != retiredTicks {
					t.Error("retired movie renderer received a resize redraw")
				}
			}
			if n := sc.surfaces[0].destroyed.Load(); n != 0 {
				t.Error("movie retirement destroyed active GUI surface")
			}
			if got := driver.PlaybackStats(); got.SampleSources != 0 {
				t.Errorf("movie source survived retirement: %+v", got)
			}
			t.Logf("retirement mode=%s movie-surfaces=%d current-GUI-preserved=true", mode, len(movieSurfaces))
		})
	}
}
