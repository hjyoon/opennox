package opennox

import (
	"math"
	"testing"

	"github.com/opennox/libs/types"
)

func TestE2EAIFearRouteReadOnly(t *testing.T) {
	for _, mode := range []string{"clear", "detour"} {
		t.Run(mode, func(t *testing.T) {
			from, to := types.Ptf(0, 0), types.Ptf(460, 0)
			clear := func(a, b types.Pointf) bool {
				return mode == "clear" || e2eWarriorLaneMissesCircle(a, b, types.Ptf(230, 0), 92)
			}
			path, ok := e2eAIFearRoute(from, to, clear, clear)
			if !ok || len(path) == 0 {
				t.Fatal("clear/detour route missing")
			}
			previous := from
			detoured := false
			for _, p := range path {
				if !clear(previous, p) || previous.Sub(p).Len() > 33 {
					t.Fatal("route crossed blocked geometry or skipped a grid edge")
				}
				if p.Y != 0 {
					detoured = true
				}
				previous = p
			}
			if previous.Sub(to).Len() > 48 || !clear(previous, to) || mode == "detour" && !detoured {
				t.Fatal("route did not reach independent visible proximity")
			}
		})
	}
}

func TestE2EAIFearRouteBounds(t *testing.T) {
	for _, mode := range []string{"nil clear", "nil visible", "nearby", "nearby blocked", "disconnected", "NaN", "infinity", "outside bounded grid"} {
		t.Run(mode, func(t *testing.T) {
			from, to := types.Pointf{}, types.Ptf(460, 0)
			calls := 0
			clear := func(_, _ types.Pointf) bool { calls++; return mode != "disconnected" }
			visible := func(_, _ types.Pointf) bool { return mode != "nearby blocked" }
			switch mode {
			case "nil clear":
				clear = nil
			case "nil visible":
				visible = nil
			case "nearby", "nearby blocked":
				to = types.Ptf(48, 0)
			case "NaN":
				to.X = float32(math.NaN())
			case "infinity":
				from.X = float32(math.Inf(1))
			case "outside bounded grid":
				to = types.Ptf(5000, 0)
			}
			path, ok := e2eAIFearRoute(from, to, clear, visible)
			if ok != (mode == "nearby") || len(path) != 0 || calls > 4096*8 {
				t.Fatalf("bounded route accepted=%t path=%d calls=%d", ok, len(path), calls)
			}
		})
	}
}

func TestE2EAIFearApproachInactive(t *testing.T) {
	for _, mode := range []string{"before expiry", "already hit"} {
		t.Run(mode, func(t *testing.T) {
			f := &e2eAIFearFixture{expired: mode == "already hit", hit: mode == "already hit"}
			if aim, walk := f.approach(); walk || aim != (types.Pointf{}) {
				t.Fatal("inactive movement observer accessed live game state")
			}
		})
	}
}
