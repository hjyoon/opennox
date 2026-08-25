// Package headless provides an in-memory seat for deterministic GUI tests.
// It implements the same screen and input contracts as the SDL seat without
// creating an operating-system window or an OpenGL context.
package headless

import (
	"image"
	"sync"

	libseat "github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/noximage"
)

var _ libseat.Seat = (*Seat)(nil)

// New creates an in-memory seat at the requested drawable size.
func New(sz image.Point) *Seat {
	return &Seat{
		size: sz,
		max:  sz,
		mode: libseat.Windowed,
	}
}

// Seat is an in-memory implementation of seat.Seat. Input events are only
// delivered when QueueInput is called, and rendered frames remain in memory.
type Seat struct {
	mu sync.RWMutex

	size       image.Point
	max        image.Point
	mode       libseat.ScreenMode
	gamma      float32
	textInput  bool
	closed     bool
	onResize   []func(image.Point)
	onInput    libseat.InputConfig
	inputQueue []libseat.InputEvent

	lastSurface  *Surface
	lastView     image.Rectangle
	clearCount   uint64
	presentCount uint64
}

func (s *Seat) ScreenSize() image.Point {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.size
}

func (s *Seat) ScreenMaxSize() image.Point {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.max
}

func (s *Seat) ResizeScreen(sz image.Point) {
	s.mu.Lock()
	if s.size == sz {
		s.mu.Unlock()
		return
	}
	s.size = sz
	if sz.X > s.max.X {
		s.max.X = sz.X
	}
	if sz.Y > s.max.Y {
		s.max.Y = sz.Y
	}
	callbacks := append([]func(image.Point){}, s.onResize...)
	s.mu.Unlock()
	for _, fnc := range callbacks {
		fnc(sz)
	}
}

func (s *Seat) SetScreenMode(mode libseat.ScreenMode) {
	s.mu.Lock()
	s.mode = mode
	s.mu.Unlock()
}

func (s *Seat) SetGamma(v float32) {
	s.mu.Lock()
	s.gamma = v
	s.mu.Unlock()
}

func (s *Seat) OnScreenResize(fnc func(image.Point)) {
	s.mu.Lock()
	s.onResize = append(s.onResize, fnc)
	s.mu.Unlock()
}

func (s *Seat) NewSurface(sz image.Point, filter bool) libseat.Surface {
	return &Surface{
		seat:   s,
		size:   sz,
		filter: filter,
		frame:  noximage.NewImage16(image.Rectangle{Max: sz}),
	}
}

func (s *Seat) Clear() {
	s.mu.Lock()
	s.clearCount++
	s.mu.Unlock()
}

func (s *Seat) Present() {
	s.mu.Lock()
	s.presentCount++
	s.mu.Unlock()
}

func (s *Seat) ReplaceInputs(cfg libseat.InputConfig) libseat.InputConfig {
	s.mu.Lock()
	old := s.onInput
	s.onInput = cfg
	s.mu.Unlock()
	return old
}

func (s *Seat) OnInput(fnc func(libseat.InputEvent)) {
	s.mu.Lock()
	s.onInput = append(s.onInput, fnc)
	s.mu.Unlock()
}

// QueueInput schedules synthetic input for delivery on the next InputTick.
func (s *Seat) QueueInput(evs ...libseat.InputEvent) {
	s.mu.Lock()
	s.inputQueue = append(s.inputQueue, evs...)
	s.mu.Unlock()
}

func (s *Seat) InputTick() {
	s.mu.Lock()
	evs := append([]libseat.InputEvent(nil), s.inputQueue...)
	s.inputQueue = s.inputQueue[:0]
	callbacks := append(libseat.InputConfig(nil), s.onInput...)
	s.mu.Unlock()
	for _, ev := range evs {
		for _, fnc := range callbacks {
			fnc(ev)
		}
	}
}

func (s *Seat) SetTextInput(enable bool) {
	s.mu.Lock()
	s.textInput = enable
	s.mu.Unlock()
}

func (s *Seat) Close() error {
	s.mu.Lock()
	s.closed = true
	s.inputQueue = nil
	s.mu.Unlock()
	return nil
}

// Snapshot returns a copy of the last surface drawn by the renderer and its
// destination viewport. A nil image means that no surface has been drawn yet.
func (s *Seat) Snapshot() (*noximage.Image16, image.Rectangle) {
	s.mu.RLock()
	surf, view := s.lastSurface, s.lastView
	s.mu.RUnlock()
	if surf == nil {
		return nil, view
	}
	return surf.snapshot(), view
}

// PresentCount returns the number of frames presented by the renderer.
func (s *Seat) PresentCount() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.presentCount
}

var _ libseat.Surface = (*Surface)(nil)

// Surface stores the most recently uploaded software-rendered frame.
type Surface struct {
	mu sync.RWMutex

	seat      *Seat
	size      image.Point
	filter    bool
	frame     *noximage.Image16
	destroyed bool
}

func (s *Surface) Size() image.Point {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.size
}

func (s *Surface) Update(data *noximage.Image16) {
	if data == nil {
		return
	}
	s.mu.Lock()
	if s.destroyed {
		s.mu.Unlock()
		return
	}
	frame := noximage.NewImage16(data.Rect)
	for y := data.Rect.Min.Y; y < data.Rect.Max.Y; y++ {
		copy(frame.Row(y), data.Row(y))
	}
	s.frame = frame
	s.size = data.Size()
	s.mu.Unlock()
}

func (s *Surface) Draw(view image.Rectangle) {
	s.mu.RLock()
	destroyed := s.destroyed
	s.mu.RUnlock()
	if destroyed {
		return
	}
	s.seat.mu.Lock()
	s.seat.lastSurface = s
	s.seat.lastView = view
	s.seat.mu.Unlock()
}

func (s *Surface) Destroy() {
	s.mu.Lock()
	s.destroyed = true
	s.frame = nil
	s.mu.Unlock()
}

func (s *Surface) snapshot() *noximage.Image16 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.destroyed || s.frame == nil {
		return nil
	}
	out := noximage.NewImage16(s.frame.Rect)
	for y := s.frame.Rect.Min.Y; y < s.frame.Rect.Max.Y; y++ {
		copy(out.Row(y), s.frame.Row(y))
	}
	return out
}
