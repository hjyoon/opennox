package opennox

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/opennox/libs/env"
	"github.com/opennox/libs/platform"
)

const e2eOrdinaryTimeChild = "NOX_TEST_E2E_CLOCK_CHILD"

func TestE2EOrdinaryTimeContract(t *testing.T) {
	if mode := os.Getenv(e2eOrdinaryTimeChild); mode != "" {
		virtual := mode == "virtual" || mode == "default" || mode == "recording"
		if got := env.IsE2EVirtualTime(); got != virtual {
			t.Fatalf("virtual clock = %t, want %t", got, virtual)
		}
		previous := platform.Get()
		t.Cleanup(func() { platform.Set(previous) })
		frozen := &frameLimitTestClock{Platform: previous, now: 9123 * time.Millisecond}
		platform.Set(frozen)
		want := uint64(9123)
		if virtual {
			want = uint64(9123 * time.Millisecond)
		}
		if got := platformTicks(); got != want {
			t.Fatalf("game timestamp = %d, want %d", got, want)
		}
		for _, requested := range []bool{false, true} {
			if got, want := e2eFrameLimitDisabled(requested), requested || virtual; got != want {
				t.Fatalf("frame-limit disable flag(%t) = %t, want %t", requested, got, want)
			}
		}
		intro, brief, fade := 25, 15000, 25
		if virtual {
			intro, brief, fade = 10, 10, 10
		}
		if got := nox_client_getIntroScreenDuration_44E3B0(); got != intro {
			t.Fatalf("intro duration = %d, want %d", got, intro)
		}
		if got := nox_client_getBriefDuration(); got != brief {
			t.Fatalf("brief duration = %d, want %d", got, brief)
		}
		if got := getFadeDuration(); got != fade {
			t.Fatalf("fade duration = %d, want %d", got, fade)
		}
		fmt.Println("E2E_ORDINARY_TIME_CONTRACT_OK:" + mode)
		return
	}
	for _, mode := range []string{"production", "default", "virtual", "ordinary", "recording", "ordinary-recording"} {
		t.Run(mode, func(t *testing.T) {
			command := exec.Command(os.Args[0], "-test.run=^TestE2EOrdinaryTimeContract$", "-test.v")
			for _, value := range os.Environ() {
				key, _, _ := strings.Cut(value, "=")
				if !strings.HasPrefix(key, "NOX_E2E") && key != e2eOrdinaryTimeChild && key != "NOX_DEV" {
					command.Env = append(command.Env, value)
				}
			}
			command.Env = append(command.Env, e2eOrdinaryTimeChild+"="+mode)
			switch mode {
			case "production":
			case "recording", "ordinary-recording":
				command.Env = append(command.Env, "NOX_E2E_RECORD=isolated-clock-contract")
			default:
				command.Env = append(command.Env, "NOX_E2E=isolated-clock-contract")
			}
			switch mode {
			case "ordinary", "ordinary-recording":
				command.Env = append(command.Env, "NOX_E2E_CLOCK=ordinary")
			case "virtual":
				command.Env = append(command.Env, "NOX_E2E_CLOCK=virtual")
			}
			output, err := command.CombinedOutput()
			if err != nil || !strings.Contains(string(output), "E2E_ORDINARY_TIME_CONTRACT_OK:"+mode) {
				t.Fatalf("clock contract %s: %v\n%s", mode, err, output)
			}
		})
	}
}
