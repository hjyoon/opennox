package render

import (
	libseat "github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/noximage"
	"github.com/opennox/opennox/v1/client/seat/headless"
	"image"
	"testing"
)

type rendererRetirementSurface struct {
	libseat.Surface
	destroyed int
}

func (s *rendererRetirementSurface) Destroy() { s.destroyed++; s.Surface.Destroy() }

type rendererRetirementSeat struct {
	*headless.Seat
	surfaces []*rendererRetirementSurface
}

func (s *rendererRetirementSeat) NewSurface(sz image.Point, filter bool) libseat.Surface {
	p := &rendererRetirementSurface{Surface: s.Seat.NewSurface(sz, filter)}
	s.surfaces = append(s.surfaces, p)
	return p
}

func TestRendererRetirementReleasesStorageAndIgnoresResize(t *testing.T) {
	for _, mode := range []string{"before-frame", "after-frame", "after-recreate"} {
		t.Run(mode, func(t *testing.T) {
			sc := &rendererRetirementSeat{Seat: headless.New(image.Pt(640, 480))}
			defer sc.Close()
			r, err := New(sc)
			if err != nil {
				t.Fatal(err)
			}
			callbacks := 0
			r.OnViewResize(func(image.Rectangle) { callbacks++ })
			frame := noximage.NewImage16(image.Rect(0, 0, 640, 480))
			if mode != "before-frame" {
				r.CopyBuffer(frame)
				r.CopyBuffer(frame)
			}
			if mode == "after-recreate" {
				r.CopyBuffer(noximage.NewImage16(image.Rect(0, 0, 640, 400)))
			}
			closer, ok := any(r).(interface{ Close() })
			if !ok {
				t.Fatal("renderer has no surface retirement operation")
			}
			closer.Close()
			closer.Close()
			if r.backbuf != nil || r.buf != nil || r.prev != nil || r.sc != nil || len(r.onResize) != 0 {
				t.Error("retired renderer retained surface/frame/seat/view callbacks")
			}
			for i, p := range sc.surfaces {
				if p.destroyed != 1 {
					t.Errorf("surface %d destroyed %d times, want once", i, p.destroyed)
				}
			}
			surfaceCount := len(sc.surfaces)
			ticks, presents, beforeCallbacks := r.Ticks(), sc.PresentCount(), callbacks
			r.CopyBuffer(frame)
			r.OnViewResize(func(image.Rectangle) { t.Error("closed renderer registered another view callback") })
			sc.ResizeScreen(image.Pt(1920, 1080))
			if r.Ticks() != ticks || sc.PresentCount() != presents || callbacks != beforeCallbacks {
				t.Error("retired renderer rendered or notified during resize/copy")
			}
			if err := r.Init(image.Pt(640, 480)); err == nil {
				t.Error("closed renderer reinitialized without an error")
			}
			if len(sc.surfaces) != surfaceCount {
				t.Error("retired renderer allocated another surface")
			}
		})
	}
}
