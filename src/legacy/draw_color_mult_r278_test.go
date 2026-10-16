//go:build !server

package legacy

import (
	"fmt"
	"testing"
)

type colorMultCaptureRenderR278 struct {
	Render2
	rgb      [3]byte
	calls    int
	returned byte
}

func (r *colorMultCaptureRenderR278) SetColorMultAndIntensityRGB(red, green, blue byte) byte {
	r.calls++
	r.rgb = [3]byte{red, green, blue}
	return r.returned
}

type colorMultCaptureClientR278 struct {
	Client
	render *colorMultCaptureRenderR278
}

func (c *colorMultCaptureClientR278) R2() Render2 { return c.render }

// This directly invokes the Go body of the cgo-exported433E40 function. It
// checks the actual export's argument conversion, one RGB-service invocation,
// and unchanged return propagation; it does not cross a C transition or prove
// the root RGB service's state/pixel behavior (covered by the root fixtures).
func TestPackedColorMultExport433E40AllWordsR278(t *testing.T) {
	render := &colorMultCaptureRenderR278{}
	client := &colorMultCaptureClientR278{render: render}
	oldClient := GetClient
	GetClient = func() Client { return client }
	t.Cleanup(func() { GetClient = oldClient })
	var badRGB, badCalls, badReturn int
	var firstRGB, firstCalls, firstReturn string
	for word := 0; word < 65536; word++ {
		render.calls = 0
		render.rgb = [3]byte{0xa1, 0xb3, 0xc7}
		render.returned = byte(word ^ (word >> 8) ^ 0x5d)
		got := nox_draw_setColorMultAndIntensity_433E40(int32(word))
		// Independent original mode-1 bit split, not the production split helper
		// or ColorNRGBA display expansion. Bit15 does not alter raster RGB.
		want := [3]byte{byte((word & 0x7c00) >> 7), byte((word & 0x03e0) >> 2), byte((word & 0x001f) * 8)}
		if render.rgb != want {
			badRGB++
			if firstRGB == "" {
				firstRGB = fmt.Sprintf("word=%04x got=%v want=%v", word, render.rgb, want)
			}
		}
		if render.calls != 1 {
			badCalls++
			if firstCalls == "" {
				firstCalls = fmt.Sprintf("word=%04x got=%d want=1", word, render.calls)
			}
		}
		if got != int32(render.returned) {
			badReturn++
			if firstReturn == "" {
				firstReturn = fmt.Sprintf("word=%04x got=%d want=%d", word, got, render.returned)
			}
		}
	}
	for _, failure := range []struct {
		kind, first string
		count       int
	}{{"RGB", firstRGB, badRGB}, {"service count", firstCalls, badCalls}, {"return", firstReturn, badReturn}} {
		if failure.count != 0 {
			t.Errorf("all65536 export %s: %d mismatches; first %s", failure.kind, failure.count, failure.first)
		}
	}
}
