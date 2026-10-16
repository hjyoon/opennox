package sdl

import (
	"errors"
	"reflect"
	"testing"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

type glContextRecorderR280 struct {
	t       *testing.T
	token   byte
	events  []string
	failAt  string
	failure error
}

func (r *glContextRecorderR280) handle() sdl.GLContext {
	return sdl.GLContext(unsafe.Pointer(&r.token))
}

func (r *glContextRecorderR280) create() (sdl.GLContext, error) {
	r.events = append(r.events, "create")
	if r.failAt == "create" {
		return nil, r.failure
	}
	return r.handle(), nil
}

func (r *glContextRecorderR280) bind(context sdl.GLContext) error {
	r.t.Helper()
	r.events = append(r.events, "bind")
	if context != r.handle() {
		r.t.Errorf("bound context = %p, want exact created handle %p", context, r.handle())
	}
	if r.failAt == "bind" {
		return r.failure
	}
	return nil
}

func (r *glContextRecorderR280) initializeBackend() error {
	r.events = append(r.events, "backend-init")
	if r.failAt == "backend-init" {
		return r.failure
	}
	return nil
}

func (r *glContextRecorderR280) delete(context sdl.GLContext) {
	r.t.Helper()
	r.events = append(r.events, "delete")
	if context != r.handle() {
		r.t.Errorf("deleted context = %p, want exact created handle %p", context, r.handle())
	}
}

func checkGLContextEventsR280(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("context events = %v, want %v", got, want)
	}
}

// This group independently exposes the old missing retain/delete behavior on
// the reviewed seam-only baseline. No missing-symbol compile error is RED.
// The tokens and four callbacks never call SDL or cross a C boundary.
func TestGLContextOwnerLifecycleR280(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		r := &glContextRecorderR280{t: t}
		var owner glContextOwner
		if err := owner.initialize(r.create, r.bind, r.initializeBackend, r.delete); err != nil {
			t.Fatalf("initialize: %v", err)
		}
		checkGLContextEventsR280(t, r.events, []string{"create", "bind", "backend-init"})
		if owner.context != r.handle() {
			t.Errorf("owned context = %p, want exact created handle %p", owner.context, r.handle())
		}
		owner.release(r.delete)
		checkGLContextEventsR280(t, r.events, []string{"create", "bind", "backend-init", "delete"})
		if owner.context != nil {
			t.Error("released context remains owned")
		}
	})
	for _, tc := range []struct {
		name       string
		wantError  string
		wantEvents []string
	}{
		{"create", "OpenGL creation failed: injected failure", []string{"create"}},
		{"bind", "OpenGL bind failed: injected failure", []string{"create", "bind", "delete"}},
		{"backend-init", "injected failure", []string{"create", "bind", "backend-init", "delete"}},
	} {
		t.Run(tc.name+"-error", func(t *testing.T) {
			r := &glContextRecorderR280{t: t, failAt: tc.name, failure: errors.New("injected failure")}
			var owner glContextOwner
			err := owner.initialize(r.create, r.bind, r.initializeBackend, r.delete)
			if !errors.Is(err, r.failure) {
				t.Errorf("initialize error = %v, want wrapped original sentinel", err)
			}
			if err == nil || err.Error() != tc.wantError {
				t.Errorf("initialize error = %v, want %q", err, tc.wantError)
			}
			if tc.name == "backend-init" && err != r.failure {
				t.Error("backend error identity changed")
			}
			checkGLContextEventsR280(t, r.events, tc.wantEvents)
			if owner.context != nil {
				t.Error("failed initialization retains context")
			}
			owner.release(r.delete)
			checkGLContextEventsR280(t, r.events, tc.wantEvents)
		})
	}
}

// These are additional new owner-contract rules, not independent proof of an
// unmodified original startup failure. Keep this group distinct from Lifecycle.
func TestGLContextOwnerGuardsR280(t *testing.T) {
	t.Run("empty-release", func(t *testing.T) {
		r := &glContextRecorderR280{t: t}
		var owner glContextOwner
		owner.release(r.delete)
		checkGLContextEventsR280(t, r.events, nil)
	})
	t.Run("release-once", func(t *testing.T) {
		r := &glContextRecorderR280{t: t}
		owner := glContextOwner{context: r.handle()}
		owner.release(r.delete)
		owner.release(r.delete)
		checkGLContextEventsR280(t, r.events, []string{"delete"})
		if owner.context != nil {
			t.Error("released context remains owned")
		}
	})
	t.Run("reentrant-release", func(t *testing.T) {
		r := &glContextRecorderR280{t: t}
		owner := glContextOwner{context: r.handle()}
		owner.release(func(context sdl.GLContext) {
			if owner.context != nil {
				t.Error("owner was not cleared before deletion callback")
			}
			r.delete(context)
			owner.release(r.delete)
		})
		checkGLContextEventsR280(t, r.events, []string{"delete"})
	})
	t.Run("duplicate-initialize", func(t *testing.T) {
		r := &glContextRecorderR280{t: t}
		owner := glContextOwner{context: r.handle()}
		err := owner.initialize(r.create, r.bind, r.initializeBackend, r.delete)
		if err == nil || err.Error() != "OpenGL context already initialized" {
			t.Errorf("duplicate initialize error = %v", err)
		}
		checkGLContextEventsR280(t, r.events, nil)
		if owner.context != r.handle() {
			t.Error("duplicate initialization replaced the owned context")
		}
		owner.release(r.delete)
		checkGLContextEventsR280(t, r.events, []string{"delete"})
	})
	t.Run("independent-owners", func(t *testing.T) {
		a, b := &glContextRecorderR280{t: t}, &glContextRecorderR280{t: t}
		if a.handle() == b.handle() {
			t.Fatal("independent test tokens alias")
		}
		first := glContextOwner{context: a.handle()}
		second := glContextOwner{context: b.handle()}
		first.release(a.delete)
		if first.context != nil || second.context != b.handle() {
			t.Error("first release changed the other owner")
		}
		checkGLContextEventsR280(t, a.events, []string{"delete"})
		checkGLContextEventsR280(t, b.events, nil)
		second.release(b.delete)
		checkGLContextEventsR280(t, b.events, []string{"delete"})
	})
	t.Run("retry-after-bind-failure", func(t *testing.T) {
		a, b := &glContextRecorderR280{t: t, failAt: "bind", failure: errors.New("injected failure")}, &glContextRecorderR280{t: t}
		var owner glContextOwner
		if err := owner.initialize(a.create, a.bind, a.initializeBackend, a.delete); !errors.Is(err, a.failure) {
			t.Errorf("first initialize error = %v", err)
		}
		if err := owner.initialize(b.create, b.bind, b.initializeBackend, b.delete); err != nil {
			t.Fatalf("retry: %v", err)
		}
		if owner.context != b.handle() {
			t.Error("retry did not retain the newly created context")
		}
		owner.release(b.delete)
		checkGLContextEventsR280(t, a.events, []string{"create", "bind", "delete"})
		checkGLContextEventsR280(t, b.events, []string{"create", "bind", "backend-init", "delete"})
	})
}
