//go:build !server

package opennox

import (
	"fmt"
	"image"
	"os"
	"testing"

	"github.com/opennox/opennox/v1/client/seat/headless"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

// API fixture with an actual renderer/headless seat and C-to-Go entry.
// The actual in-game constructor/Close path has its own integration fixture.
func TestOptionsVideoApply(t *testing.T) {
	for _, phase := range []string{"video-apply", "video-apply-e2e"} {
		t.Run(phase, func(t *testing.T) {
			runOptionsConfigChild(t, t.TempDir(), phase, 0)
		})
	}
}

func testOptionsVideoApply(t *testing.T, c *Client, sc *headless.Seat, phase string) {
	t.Helper()
	const sentinel = "private options apply sentinel\n"
	for _, path := range []string{"nox.cfg", "opennox.yml"} {
		if err := os.WriteFile(path, []byte(sentinel), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	initial := image.Pt(640, 480)
	for index, res := range getResolutionOptions() {
		if res == (image.Point{}) {
			continue
		}
		for _, mode := range []int{-3, -1, -2} {
			c.videoSetGameMode(initial)
			c.UpdateFullScreen(mode)
			g_fullscreen_cfg = 99 // A deliberately different save checkpoint.
			guiOptionsRes = res
			ccall.CallVoidVoid(legacy.OptionsApplyVideoModeCEntry)
			wantSize, wantCheckpoint := res, mode
			if phase == "video-apply-e2e" {
				wantSize, wantCheckpoint = initial, 99
			}
			if sc.ScreenSize() != wantSize || c.videoGetGameMode() != wantSize || g_fullscreen_cfg != wantCheckpoint || c.GetWindowMode() != mode {
				t.Fatalf("%s resolution %d mode %d: seat=%v game=%v checkpoint=%d live=%d; want %v/%d", phase, index, mode, sc.ScreenSize(), c.videoGetGameMode(), g_fullscreen_cfg, c.GetWindowMode(), wantSize, wantCheckpoint)
			}
		}
	}
	for _, path := range []string{"nox.cfg", "opennox.yml"} {
		if data, err := os.ReadFile(path); err != nil || string(data) != sentinel {
			t.Fatal(fmt.Sprintf("apply-only operation changed %s: %v", path, err))
		}
	}
}
