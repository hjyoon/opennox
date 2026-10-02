package opennox

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/timer"
)

const optionsFXObserverChild = "NOX_TEST_OPTIONS_FX_OBSERVER_CHILD"

// Observer-only fixture. It runs the read-only before/after steps, never the
// scheduled mouse input or an option handler. Actual queued GUI input and
// native OpenAL source gain have separate headless/integration regressions.
func TestOptionsAuditFXObserver(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestOptionsAuditFXObserverChild$", "-test.count=1", "-test.v")
	cmd.Dir = t.TempDir()
	for _, kv := range os.Environ() {
		key, _, _ := strings.Cut(kv, "=")
		if key == optionsFXObserverChild || key == "NOX_DATA" || strings.HasPrefix(key, "NOX_E2E") {
			continue
		}
		cmd.Env = append(cmd.Env, kv)
	}
	cmd.Env = append(cmd.Env, optionsFXObserverChild+"=true")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("FX observer subprocess failed: %v\n%s", err, out)
	}
}

func TestOptionsAuditFXObserverChild(t *testing.T) {
	if os.Getenv(optionsFXObserverChild) != "true" {
		t.Skip("only run in the isolated FX observer subprocess")
	}
	legacy.InitBlobData()
	if legacy.Sub_453070() != 1 {
		t.Fatal("fresh FX observer fixture is not enabled")
	}
	g := gui.New(nil)
	t.Cleanup(g.DestroyAll)
	root := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 640, 480, nil)
	root.SetID(300)
	slider := g.NewWindowRaw(root, gui.StatusEnabled, 0, 0, 70, 12, nil)
	slider.SetID(351)
	data, free := alloc.New(gui.SliderData{})
	slider.WidgetData = unsafe.Pointer(data)
	t.Cleanup(func() { slider.WidgetData = nil; free() })
	noxClient = &Client{Client: &client.Client{GUI: g}}
	fx := (*timer.TimerGroup)(legacy.Get_dword_587000_127004())
	for _, mode := range []int{0, 1} {
		for _, startup := range []int{0, 1, 8192, VolumeMax} {
			for _, live := range []int{1, 4850, 8126, VolumeMax} {
				for _, fault := range []string{"", "target", "startup"} {
					t.Run(fmt.Sprintf("mode%d/startup%d/live%d/fault%s", mode, startup, live, fault), func(t *testing.T) {
						configSetVolume(startup, VolumeFX)
						*fx = timer.TimerGroup{}
						fx.Timers[0].Current, fx.Timers[0].Target = VolumeMax<<16, uint32(live)<<16
						*data = gui.SliderData{Min: 0, Max: VolumeMax, Field3: uint32(live)}
						a, sc := &optionsAudit{mode: mode}, new(e2eScenario)
						a.slider(sc, 351, float64(live)/VolumeMax)
						for _, step := range sc.steps {
							if step.name == "audit FX startup value before slider" {
								step.fnc()
							}
						}
						switch fault {
						case "target":
							fx.Timers[0].Target ^= 256 << 16
						case "startup":
							configSetVolume(startup^1, VolumeFX)
						}
						beforeTimer, beforeData, beforeStartup, beforeInput := *fx, *data, configGetVolume(VolumeFX), len(e2e.input)
						sc.steps[len(sc.steps)-1].fnc()
						wantFailed := 0
						if fault != "" {
							wantFailed = 1
						}
						if a.failed != wantFailed || a.passed != 4-wantFailed {
							t.Fatalf("observer passed=%d failed=%d; want %d/%d without treating startup as live gain", a.passed, a.failed, 4-wantFailed, wantFailed)
						}
						if *fx != beforeTimer || *data != beforeData || configGetVolume(VolumeFX) != beforeStartup || len(e2e.input) != beforeInput || legacy.Sub_453070() != 1 || g.Focused() != nil {
							t.Fatal("read-only observer changed settings, timer, widget, queued input, enabled flag or focus")
						}
						if current, _, _ := optionsAuditVolume(VolumeFX); current != VolumeMax {
							t.Fatal("observer advanced the inactive timer")
						}
					})
				}
			}
		}
	}
}
