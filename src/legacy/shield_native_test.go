package legacy

import (
	"reflect"
	"testing"

	"github.com/opennox/opennox/v1/server"
)

func TestShieldCallbacksUseNativeDurationAndObjectLayouts(t *testing.T) {
	target := &server.Object{ObjClass: 4}
	sp := &server.DurSpell{Level: 3, Target48: target}
	var events []string
	deps := shieldNativeDeps52F5A0{
		balance: func(key string, index int) float64 {
			events = append(events, key)
			if key == "ShieldDuration" {
				return 12.5
			}
			return 41.5
		},
		frame: func() uint32 {
			events = append(events, "frame")
			return 100
		},
		apply: func(got *server.Object, buff int, duration int16, power byte) {
			if got != target || buff != 26 || duration != 12 || power != 3 {
				t.Fatalf("apply = %p/%d/%d/%d", got, buff, duration, power)
			}
			events = append(events, "apply")
		},
		off: func(got *server.Object, buff int) int {
			if got != target || buff != 26 {
				t.Fatalf("off = %p/%d", got, buff)
			}
			events = append(events, "off")
			return 17
		},
	}

	if got := shieldCreateNative52F5A0(sp, deps); got != 0 {
		t.Fatalf("create = %d, want 0", got)
	}
	if sp.Frame68 != 112 || sp.Field72 != 42 {
		t.Fatalf("duration state = frame %d, health %d", sp.Frame68, sp.Field72)
	}
	if want := []string{"ShieldDuration", "apply", "frame", "ShieldHealth"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("create events = %v, want %v", events, want)
	}
	if got := shieldUpdateNative52F650(sp); got != 0 {
		t.Fatalf("update = %d, want 0", got)
	}
	if got := shieldDestroyNative52F670(sp, deps); got != 17 {
		t.Fatalf("destroy = %d, want 17", got)
	}

	target.ObjFlags = 0x20
	if got := shieldUpdateNative52F650(sp); got != 1 {
		t.Fatalf("destroyed target update = %d, want 1", got)
	}
	if got := shieldCreateNative52F5A0(sp, deps); got != 1 {
		t.Fatalf("destroyed target create = %d, want 1", got)
	}
}
