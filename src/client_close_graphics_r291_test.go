//go:build !server

package opennox

import (
	"errors"
	"image"
	"reflect"
	"testing"

	"github.com/opennox/libs/client/seat"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/render"
)

type r291CloseSeat struct {
	seat.Seat
	t        *testing.T
	events   []string
	resize   func(image.Point)
	surface  *r291CloseSurface
	closeErr error
	closed   bool
}

func (s *r291CloseSeat) ScreenSize() image.Point { return image.Pt(64, 32) }

func (s *r291CloseSeat) NewSurface(image.Point, bool) seat.Surface {
	s.surface = &r291CloseSurface{owner: s}
	return s.surface
}

func (s *r291CloseSeat) OnScreenResize(f func(image.Point)) { s.resize = f }

func (s *r291CloseSeat) Close() error {
	s.events = append(s.events, "seat")
	if s.closeErr != nil {
		return s.closeErr
	}
	s.closed = true
	return nil
}

type r291CloseSurface struct {
	seat.Surface
	owner    *r291CloseSeat
	destroys int
}

func (s *r291CloseSurface) Destroy() {
	s.destroys++
	if s.owner.closed {
		s.owner.t.Error("surface destroyed after its seat/context was closed")
	}
	s.owner.events = append(s.owner.events, "surface")
}

func r291PreserveCloseCallbacks(t *testing.T) {
	t.Helper()
	oldInput, oldScenario := e2e.inputToWindow, e2e.scenarioToWindow
	t.Cleanup(func() { e2e.inputToWindow, e2e.scenarioToWindow = oldInput, oldScenario })
}

func r291CloseFixture(t *testing.T) (*Client, *r291CloseSeat) {
	t.Helper()
	r291PreserveCloseCallbacks(t)
	s := &r291CloseSeat{t: t}
	r, err := render.New(s)
	if err != nil {
		t.Fatal(err)
	}
	return &Client{Client: &client.Client{Seat: s, Win: r}}, s
}

func TestClientCloseGraphicsOrdersRetirement(t *testing.T) {
	c, s := r291CloseFixture(t)
	e2e.inputToWindow = func(p image.Point) image.Point { return p }
	e2e.scenarioToWindow = func(p image.Point) image.Point { return p }

	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	if want := []string{"surface", "seat"}; !reflect.DeepEqual(s.events, want) {
		t.Fatalf("retirement order = %v, want %v", s.events, want)
	}
	if c.Win != nil || c.Seat != nil || !s.closed || s.surface.destroys != 1 {
		t.Fatal("graphics owners were not retired exactly once")
	}
	if e2e.inputToWindow != nil || e2e.scenarioToWindow != nil {
		t.Fatal("coordinate callbacks retain the retired seat")
	}
	if s.resize == nil {
		t.Fatal("fixture did not retain the seat resize callback")
	}
	s.resize(image.Pt(128, 64))
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(s.events, []string{"surface", "seat"}) || s.surface.destroys != 1 {
		t.Fatal("late resize/repeated close used the retired graphics owner")
	}
}

func TestClientCloseGraphicsRetainsFailedSeat(t *testing.T) {
	c, s := r291CloseFixture(t)
	closeErr := errors.New("seat close failed")
	s.closeErr = closeErr
	if err := c.Close(); !errors.Is(err, closeErr) {
		t.Fatalf("close error = %v, want %v", err, closeErr)
	}
	if c.Win != nil || c.Seat != s || s.closed || s.surface.destroys != 1 {
		t.Fatal("failed seat owner lost, or surface not retired")
	}
	if !reflect.DeepEqual(s.events, []string{"surface", "seat"}) {
		t.Fatalf("failed close order = %v", s.events)
	}
	s.closeErr = nil
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	if c.Seat != nil || !s.closed || s.surface.destroys != 1 ||
		!reflect.DeepEqual(s.events, []string{"surface", "seat", "seat"}) {
		t.Fatal("retry did not finish only the outstanding seat cleanup")
	}
}

func TestClientCloseGraphicsWithoutRenderer(t *testing.T) {
	r291PreserveCloseCallbacks(t)
	s := &r291CloseSeat{t: t}
	c := &Client{Client: &client.Client{Seat: s}}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	if c.Seat != nil || !s.closed || !reflect.DeepEqual(s.events, []string{"seat"}) {
		t.Fatal("seat-only owner was not closed")
	}
}

func TestClientCloseGraphicsEmpty(t *testing.T) {
	r291PreserveCloseCallbacks(t)
	for _, c := range []*Client{nil, {}, {Client: &client.Client{}}} {
		if err := c.Close(); err != nil {
			t.Fatal(err)
		}
	}
}
