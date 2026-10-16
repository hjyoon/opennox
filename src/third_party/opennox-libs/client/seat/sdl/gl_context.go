package sdl

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
)

// glContextOwner owns the SDL context separately from its window and GL objects.
type glContextOwner struct {
	context sdl.GLContext
}

func (owner *glContextOwner) initialize(create func() (sdl.GLContext, error), bind func(sdl.GLContext) error, initializeBackend func() error, deleteContext func(sdl.GLContext)) error {
	if owner.context != nil {
		return fmt.Errorf("OpenGL context already initialized")
	}
	context, err := create()
	if err != nil {
		return fmt.Errorf("OpenGL creation failed: %w", err)
	}
	owner.context = context
	if err := bind(context); err != nil {
		owner.release(deleteContext)
		return fmt.Errorf("OpenGL bind failed: %w", err)
	}
	if err := initializeBackend(); err != nil {
		owner.release(deleteContext)
		return err
	}
	return nil
}

func (owner *glContextOwner) release(deleteContext func(sdl.GLContext)) {
	if owner.context == nil {
		return
	}
	context := owner.context
	owner.context = nil
	deleteContext(context)
}
