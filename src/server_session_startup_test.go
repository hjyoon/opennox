package opennox

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestCompleteServerSessionStartupOrder(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(fmt.Sprintf("failure=%v", fail), func(t *testing.T) {
			failure := errors.New("HTTP listen failed")
			var calls []string
			got := completeServerSessionStartup(func() error {
				calls = append(calls, "services")
				if fail {
					return failure
				}
				return nil
			}, func() {
				calls = append(calls, "session cleanup")
			}, func() {
				calls = append(calls, "restore startup")
			})
			wantCalls := []string{"services"}
			var wantErr error
			if fail {
				wantCalls = append(wantCalls, "session cleanup", "restore startup")
				wantErr = failure
			}
			if got != wantErr || !reflect.DeepEqual(calls, wantCalls) {
				t.Fatalf("result=%v calls=%v, want original error=%v calls=%v", got, calls, wantErr, wantCalls)
			}
		})
	}
}

// Use real native allocation classes here, not a relaxed free/list stub. The
// headless port-collision scenarios separately exercise the complete engine
// session teardown; this checks the helper's release-before-retry ordering.
func TestCompleteServerSessionStartupRetries(t *testing.T) {
	for _, failures := range []int{1, 3, 10, 257} {
		t.Run(fmt.Sprintf("failures=%d", failures), func(t *testing.T) {
			failure := errors.New("occupied HTTP port")
			var live sessionStartupTestClass
			t.Cleanup(func() {
				if live.active {
					live.class.Free()
				}
			})
			cleanupCalls := 0
			restoreCalls := 0
			for attempt := 0; attempt <= failures; attempt++ {
				if live.active {
					t.Fatal("previous session is still live at retry")
				}
				live.class = alloc.NewClassT("sessionStartupTest", uintptr(0), 1)
				live.object = live.class.NewObject()
				live.active = true
				*live.object = uintptr(attempt + 1)
				got := completeServerSessionStartup(func() error {
					if !live.active || *live.object != uintptr(attempt+1) {
						t.Fatal("services started without the current session")
					}
					if attempt < failures {
						return failure
					}
					return nil
				}, func() {
					live.class.FreeObjectFirst(live.object)
					live.class.Free()
					live.active = false
					cleanupCalls++
				}, func() {
					if live.active || cleanupCalls != attempt+1 {
						t.Fatal("startup intent restored before session cleanup")
					}
					restoreCalls++
				})
				if attempt < failures {
					if got != failure || live.active {
						t.Fatalf("failed session result=%v active=%v", got, live.active)
					}
				} else {
					if got != nil || !live.active || *live.object != uintptr(attempt+1) {
						t.Fatalf("successful session result=%v active=%v", got, live.active)
					}
					live.class.FreeObjectFirst(live.object)
					live.class.Free()
					live.active = false
				}
			}
			if cleanupCalls != failures || restoreCalls != failures {
				t.Fatalf("cleanup count=%d restore count=%d, want %d of each", cleanupCalls, restoreCalls, failures)
			}
		})
	}
}

type sessionStartupTestClass struct {
	class  alloc.ClassT[uintptr]
	object *uintptr
	active bool
}
