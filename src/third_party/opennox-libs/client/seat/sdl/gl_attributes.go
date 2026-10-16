package sdl

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
)

// configureGLAttributes requests the context attributes before window creation.
// The setter parameter lets the request contract be tested without SDL startup.
func configureGLAttributes(set func(sdl.GLattr, int) error) error {
	if err := set(sdl.GL_CONTEXT_MAJOR_VERSION, 3); err != nil {
		return fmt.Errorf("cannot set OpenGL version: %w", err)
	}
	if err := set(sdl.GL_CONTEXT_MINOR_VERSION, 3); err != nil {
		return fmt.Errorf("cannot set OpenGL version: %w", err)
	}
	if err := set(sdl.GL_CONTEXT_PROFILE_MASK, int(sdl.GL_CONTEXT_PROFILE_CORE)); err != nil {
		return fmt.Errorf("cannot set OpenGL core: %w", err)
	}
	if err := set(sdl.GL_CONTEXT_FLAGS, int(sdl.GL_CONTEXT_FORWARD_COMPATIBLE_FLAG)); err != nil {
		return fmt.Errorf("cannot set OpenGL forward: %w", err)
	}
	if err := set(sdl.GL_DOUBLEBUFFER, 1); err != nil {
		return fmt.Errorf("cannot set OpenGL attribute: %w", err)
	}
	return nil
}
