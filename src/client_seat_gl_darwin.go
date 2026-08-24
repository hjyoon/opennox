//go:build darwin && !server

package opennox

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
)

// prepareSeatOpenGL supplies the profile attributes that the current
// opennox/libs SDL seat accidentally passes as attribute identifiers. macOS
// requires a core, forward-compatible profile for OpenGL 3.3 contexts.
func prepareSeatOpenGL() error {
	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_TIMER); err != nil {
		return fmt.Errorf("SDL initialization failed: %w", err)
	}
	attrs := []struct {
		attr  sdl.GLattr
		value int
	}{
		{sdl.GL_CONTEXT_PROFILE_MASK, sdl.GL_CONTEXT_PROFILE_CORE},
		{sdl.GL_CONTEXT_FLAGS, sdl.GL_CONTEXT_FORWARD_COMPATIBLE_FLAG},
	}
	for _, it := range attrs {
		if err := sdl.GLSetAttribute(it.attr, it.value); err != nil {
			sdl.Quit()
			return fmt.Errorf("cannot configure macOS OpenGL context: %w", err)
		}
	}
	return nil
}
