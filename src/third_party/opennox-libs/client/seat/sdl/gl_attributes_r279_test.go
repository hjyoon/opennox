package sdl

import (
	"errors"
	"reflect"
	"testing"

	"github.com/veandco/go-sdl2/sdl"
)

type glAttributeCallR279 struct {
	attr  sdl.GLattr
	value int
}

func expectedGLAttributeCallsR279() []glAttributeCallR279 {
	return []glAttributeCallR279{
		{sdl.GL_CONTEXT_MAJOR_VERSION, 3},
		{sdl.GL_CONTEXT_MINOR_VERSION, 3},
		{sdl.GL_CONTEXT_PROFILE_MASK, int(sdl.GL_CONTEXT_PROFILE_CORE)},
		{sdl.GL_CONTEXT_FLAGS, int(sdl.GL_CONTEXT_FORWARD_COMPATIBLE_FLAG)},
		{sdl.GL_DOUBLEBUFFER, 1},
	}
}

// SDL_GLattr keys and the profile/flag values are different enum domains.
// Numeric values here are independently read from the installed SDL2 header,
// not computed from the implementation or from a fake GL context.
func checkGLAttributeEnumDomainsR279(t *testing.T) {
	t.Helper()
	for _, tc := range []struct {
		name string
		got  int
		want int
	}{
		{"major key", int(sdl.GL_CONTEXT_MAJOR_VERSION), 17},
		{"minor key", int(sdl.GL_CONTEXT_MINOR_VERSION), 18},
		{"profile mask key", int(sdl.GL_CONTEXT_PROFILE_MASK), 21},
		{"context flags key", int(sdl.GL_CONTEXT_FLAGS), 20},
		{"double buffer key", int(sdl.GL_DOUBLEBUFFER), 5},
		{"core profile value", int(sdl.GL_CONTEXT_PROFILE_CORE), 1},
		{"forward compatible flag value", int(sdl.GL_CONTEXT_FORWARD_COMPATIBLE_FLAG), 2},
		{"green channel key", int(sdl.GL_GREEN_SIZE), 1},
		{"blue channel key", int(sdl.GL_BLUE_SIZE), 2},
	} {
		if tc.got != tc.want {
			t.Errorf("%s = %d, want %d", tc.name, tc.got, tc.want)
		}
	}
}

func TestConfigureGLAttributesR279(t *testing.T) {
	checkGLAttributeEnumDomainsR279(t)
	var calls []glAttributeCallR279
	err := configureGLAttributes(func(attr sdl.GLattr, value int) error {
		calls = append(calls, glAttributeCallR279{attr, value})
		return nil
	})
	if err != nil {
		t.Fatalf("configureGLAttributes: %v", err)
	}
	if want := expectedGLAttributeCallsR279(); !reflect.DeepEqual(calls, want) {
		t.Fatalf("ordered attribute calls = %v, want %v", calls, want)
	}
}

func TestConfigureGLAttributesFailureR279(t *testing.T) {
	wantCalls := expectedGLAttributeCallsR279()
	for failAt, tc := range []struct {
		name    string
		message string
	}{
		{"major", "cannot set OpenGL version"},
		{"minor", "cannot set OpenGL version"},
		{"profile", "cannot set OpenGL core"},
		{"flags", "cannot set OpenGL forward"},
		{"doublebuffer", "cannot set OpenGL attribute"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sentinel := errors.New("injected setting failure")
			var calls []glAttributeCallR279
			err := configureGLAttributes(func(attr sdl.GLattr, value int) error {
				calls = append(calls, glAttributeCallR279{attr, value})
				if len(calls) == failAt+1 {
					return sentinel
				}
				return nil
			})
			if !errors.Is(err, sentinel) {
				t.Errorf("error = %v, want wrapped sentinel", err)
			}
			if want := tc.message + ": " + sentinel.Error(); err == nil || err.Error() != want {
				t.Errorf("error = %v, want %q", err, want)
			}
			if want := wantCalls[:failAt+1]; !reflect.DeepEqual(calls, want) {
				t.Errorf("fail-fast attribute calls = %v, want %v", calls, want)
			}
		})
	}
}
