package opennox

import (
	"testing"

	"github.com/opennox/opennox/v1/server"
)

func TestObjectDeleteFinish4E5E80OrderAndArgument(t *testing.T) {
	obj := &server.Object{}
	want := []string{
		"transfer slaves",
		"clear owner",
		"clear activators",
		"decay",
		"drop all items",
		"finalize",
		"free",
	}
	var got []string
	step := func(name string) func(*server.Object) {
		return func(arg *server.Object) {
			if arg != obj {
				t.Fatalf("%s received %p, want %p", name, arg, obj)
			}
			if len(got) >= len(want) {
				t.Fatalf("unexpected callback %d = %q", len(got), name)
			}
			if name != want[len(got)] {
				t.Fatalf("callback %d = %q, want %q", len(got), name, want[len(got)])
			}
			got = append(got, name)
		}
	}

	objectDeleteFinish_4E5E80(obj, objectDeleteFinish4E5E80Hooks{
		transferSlaves:  step("transfer slaves"),
		clearOwner:      step("clear owner"),
		clearActivators: step("clear activators"),
		decay:           step("decay"),
		dropAllItems:    step("drop all items"),
		finalize:        step("finalize"),
		free:            step("free"),
	})
	if len(got) != len(want) {
		t.Fatalf("callback count = %d, want %d", len(got), len(want))
	}
}

func TestObjectDeleteFinish4E5E80DoesNotPrecheckObject(t *testing.T) {
	calls := 0
	acceptNil := func(obj *server.Object) {
		if obj != nil {
			t.Fatalf("callback received %p, want nil", obj)
		}
		calls++
	}
	objectDeleteFinish_4E5E80(nil, objectDeleteFinish4E5E80Hooks{
		transferSlaves:  acceptNil,
		clearOwner:      acceptNil,
		clearActivators: acceptNil,
		decay:           acceptNil,
		dropAllItems:    acceptNil,
		finalize:        acceptNil,
		free:            acceptNil,
	})
	if calls != 7 {
		t.Fatalf("callback count = %d, want 7", calls)
	}
}
