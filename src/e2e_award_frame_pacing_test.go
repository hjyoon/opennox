package opennox

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/opennox/libs/types"
)

func TestE2EAwardFramePacingRequiresRealWaitAndFrozenGameplay(t *testing.T) {
	pos := types.Pointf{X: 12, Y: 34}
	for _, c := range []struct {
		name    string
		elapsed time.Duration
		loops   int
		draws   int
		frame   uint32
		pos     types.Pointf
		wantErr bool
	}{
		{"real_30_fps", 500 * time.Millisecond, 15, 15, 123, pos, false},
		{"slower_renderer", time.Second, 15, 15, 123, pos, false},
		{"missing_limiter", 30 * time.Millisecond, 15, 15, 123, pos, true},
		{"zero_time", 0, 15, 15, 123, pos, true},
		{"no_rendering", 500 * time.Millisecond, 15, 0, 123, pos, true},
		{"extra_rendering", 500 * time.Millisecond, 15, 30, 123, pos, true},
		{"clock_was_changed", 500 * time.Millisecond, 30, 15, 123, pos, true},
		{"unpaused_simulation", 500 * time.Millisecond, 15, 15, 124, pos, true},
		{"player_moved", 500 * time.Millisecond, 15, 15, 123, types.Pointf{}, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			err := e2eAwardFramePacingError(c.elapsed, c.loops, c.draws, 123, c.frame, pos, c.pos)
			if (err != nil) != c.wantErr {
				t.Fatalf("pacing validation: err=%v wantErr=%t", err, c.wantErr)
			}
		})
	}
}

func TestE2EAwardFramePacingNeverInjectsPauseClockOrAwardResults(t *testing.T) {
	f, err := parser.ParseFile(token.NewFileSet(), "e2e_award_frame_pacing.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	calls := make(map[string]int)
	ast.Inspect(f, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			var name string
			switch fun := call.Fun.(type) {
			case *ast.Ident:
				name = fun.Name
			case *ast.SelectorExpr:
				name = fun.Sel.Name
			}
			calls[name]++
			switch name {
			case "Sleep", "LoopSleep", "RateWait", "SetRateLimit", "SetFrame", "IncFrame", "SetGame", "UnsetGame", "SetEngine", "UnsetEngine", "PauseFXStart57AF30", "Sub_57B0A0", "Sub_413A00", "SetPlayerState", "tick", "nox_ticks_reset_416D40":
				t.Errorf("observer injects pause/pacing outcome through %s", name)
			}
		}
		if assign, ok := n.(*ast.AssignStmt); ok {
			for _, lhs := range assign.Lhs {
				if sel, ok := lhs.(*ast.SelectorExpr); ok {
					switch sel.Sel.Name {
					case "ticks", "DrawCnt", "Experience", "Level", "EquippedWeapon", "State", "PosVec":
						t.Errorf("observer overwrites production result %s", sel.Sel.Name)
					}
				}
				if index, ok := lhs.(*ast.IndexExpr); ok {
					if sel, ok := index.X.(*ast.SelectorExpr); ok && sel.Sel.Name == "SpellLvl" {
						t.Error("observer overwrites spell knowledge")
					}
				}
			}
		}
		return true
	})
	for _, name := range []string{"Nox_xxx_plyrGiveExp_4EF3A0_exp_level", "Nox_xxx_spellGrantToPlayer_4FB550", "SetHalberd", "Now", "Since", "Ticks", "ClientQuickbarSnapshot", "Nox_xxx_get_57AF20", "HasGame", "Frame", "setEnableFrameLimit", "Screen", "Key"} {
		if calls[name] == 0 {
			t.Errorf("missing real award/pacing observer: %s", name)
		}
	}
}

func TestE2EAwardFramePacingCoversActualAwardFamiliesAndNaturalResume(t *testing.T) {
	data, err := os.ReadFile("e2e_award_frame_pacing.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{"level-up", "new-spell", "OblivionHalberd", "OblivionHeart", "OblivionWierdling", "OblivionOrb", "e2e.slow = 0", "measure live render pacing", "natural award completion", "real gameplay resume and award outcome", "close book through B input", "oldLimit", "oldSlow"} {
		if !strings.Contains(string(data), required) {
			t.Errorf("award coverage missing %s", required)
		}
	}
	data, err = os.ReadFile("../scripts/e2e/solo-wizard-award-frame-pacing.yaml")
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{"open solo campaign", "select wizard", "action: check-award-frame-pacing", "action: quit"} {
		if !strings.Contains(string(data), required) {
			t.Errorf("actual Solo scenario missing %s", required)
		}
	}
}
