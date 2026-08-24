//go:build !server

package opennox

import (
	"image"
	"unicode/utf16"

	"github.com/spf13/viper"
	"github.com/tawesoft/golib/v2/dialog"

	"github.com/opennox/libs/client/seat"
	seatsdl "github.com/opennox/libs/client/seat/sdl"
	"github.com/opennox/libs/env"

	"github.com/opennox/opennox/v1/client/input"
	"github.com/opennox/opennox/v1/client/render"
	"github.com/opennox/opennox/v1/internal/version"
	"github.com/opennox/opennox/v1/legacy"
)

func init() {
	viper.SetDefault(configVideoFiltering, true)
	viper.SetDefault(configVideoStretch, false)
}

func (c *Client) initSeat(sz image.Point) error {
	if err := prepareSeatOpenGL(); err != nil {
		return err
	}
	sst, err := seatsdl.New(c.Log, "OpenNox "+version.ClientVersion(), sz)
	if err != nil {
		return err
	}
	c.Seat = sst
	if env.IsE2E() {
		c.Seat = e2eWrapSeat(c.Seat)
	}
	c.Win, err = render.New(c.Seat)
	if err != nil {
		_ = c.Seat.Close()
		return err
	}

	var (
		inp            *input.Handler
		renderViewport image.Rectangle
		lastDrawable   image.Point
		lastLogical    image.Point
	)
	syncInputViewport := func() {
		if inp == nil || renderViewport.Empty() {
			return
		}
		drawable := c.Seat.ScreenSize()
		logical := sdlLogicalWindowSize(drawable)
		if drawable == lastDrawable && logical == lastLogical {
			return
		}
		lastDrawable, lastLogical = drawable, logical
		view := input.ScaleViewport(renderViewport, drawable, logical)
		inp.SetWinSize(view)
		c.Log.Info("input viewport", "drawable", drawable, "logical", logical, "view", view)
	}
	// Refresh before the input handler consumes the first event after focus or
	// a display/DPI change. SDL reports mouse positions in logical window units
	// while OpenGL renders in drawable pixels.
	c.Seat.OnInput(func(ev seat.InputEvent) {
		switch ev.(type) {
		case *seat.MouseMoveEvent, *seat.MouseButtonEvent, seat.WindowEvent:
			syncInputViewport()
		}
	})
	inp = input.New(c.Log, c.Seat, false, c.Strings().Lang())
	c.Inp = inp
	c.GUI.SetInput(inp)

	inp.OnQuit(mainloopStop)
	inp.OnToggleFullScreen(c.Win.ToggleWindowMode)
	inp.OnKeyPress(gameexOnKeyboardPress)
	inp.OnMouseWheel(func(dv int) {
		// mix event handler is triggered only for wheel events
		call_OnLibraryNotice_265(dv)
	})
	inp.OnInputString(func(str string) {
		for _, c := range utf16.Encode([]rune(str)) {
			legacy.NoxInputOnChar(c)
		}
	})

	c.Win.OnViewResize(func(view image.Rectangle) {
		renderViewport = view
		lastDrawable = image.Point{}
		lastLogical = image.Point{}
		syncInputViewport()
	})
	OnPixBufferResize(inp.SetDrawWinSize)

	c.Win.SetFiltering(viper.GetBool(configVideoFiltering))
	c.Win.SetStretched(viper.GetBool(configVideoStretch))
	if err != nil {
		return err
	}
	sst.SetGamma(getGamma())
	return nil
}

func (c *Client) freeSeat() {
	if c.Seat != nil {
		c.Seat.Close()
		c.Seat = nil
	}
}

func errorMessage(format string, args ...any) {
	dialog.Error(format, args...)
}
