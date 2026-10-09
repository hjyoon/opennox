package opennox

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"

	"github.com/opennox/opennox/v1/client/gui"
	"gopkg.in/yaml.v2"
)

func TestE2ETrapQuickbarVisibilityIncludesAncestors(t *testing.T) {
	g := gui.New(nil)
	root := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 30, 30, nil)
	child := g.NewWindowRaw(root, gui.StatusEnabled, 0, 0, 30, 30, nil)
	leaf := g.NewWindowRaw(child, gui.StatusEnabled, 0, 0, 30, 30, nil)
	t.Cleanup(g.DestroyAll)
	if e2eTrapWindowVisible(nil) || !e2eTrapWindowVisible(leaf) {
		t.Fatal("nil or visible ancestor chain")
	}
	root.Hide()
	if e2eTrapWindowVisible(leaf) || leaf.GetFlags().IsHidden() {
		t.Fatal("parent hiding must not write the child flag")
	}
	root.Show()
	if !e2eTrapWindowVisible(leaf) {
		t.Fatal("showing a previously hidden parent must restore traversal")
	}
	child.Hide()
	root.Hide()
	root.Show()
	if e2eTrapWindowVisible(leaf) || !child.GetFlags().IsHidden() {
		t.Fatal("showing the root must not silently clear an explicitly hidden child")
	}
	child.Show()
	if !e2eTrapWindowVisible(leaf) {
		t.Fatal("the complete ancestor chain should be visible")
	}
}

func TestE2ETrapQuickbarIconsScheduleAndPublicDispatch(t *testing.T) {
	var sc e2eScenario
	sc.CheckTrapQuickbarIcons("trap")
	counts := make(map[string]int)
	for _, step := range sc.steps {
		counts[step.name]++
		if step.ready != nil && step.waitTimeout != 1200 {
			t.Fatalf("unbounded readiness: %s", step.name)
		}
	}
	for name, want := range map[string]int{
		"trap observe starting fixture":   1,
		"trap drag onto actual trap slot": 3,
		"trap real mouse drop":            3,
		"trap next empty row":             2,
		"trap wrapped first row":          1,
		"trap previous third row":         1,
		"trap closed tray":                1,
		"trap reopened tray":              1,
		"trap F11 hidden HUD":             1,
		"trap F11 restored HUD":           1,
		"trap final first row":            1,
	} {
		if counts[name] != want {
			t.Errorf("observation %q: %d, want %d", name, counts[name], want)
		}
	}
	const path = "../scripts/e2e/host-game-trap-quickbar-icons.yaml"
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var file e2eFileYML
	if err := yaml.Unmarshal(raw, &file); err != nil {
		t.Fatal(err)
	}
	fixture, checks := 0, 0
	for _, step := range file.Steps {
		switch step.Action {
		case "set-quickbar-spell":
			fixture++
			if step.Spell != 27 || step.Slot != 0 {
				t.Fatal("unexpected starting quickbar fixture")
			}
		case "check-trap-quickbar-icons":
			checks++
		}
	}
	if fixture != 1 || checks != 1 || file.Steps[len(file.Steps)-1].Action != "quit" {
		t.Fatal("public scenario must exercise the icon check and shut down normally")
	}
	sc = e2eScenario{}
	sc.Load(path)
	if len(sc.steps) < 70 {
		t.Fatal("public dispatcher did not schedule the full input and observation sequence")
	}
}

func TestE2ETrapQuickbarIconsDoesNotSupplyResults(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "e2e_trap_quickbar_icons.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	ast.Inspect(file, func(n ast.Node) bool {
		if assignment, ok := n.(*ast.AssignStmt); ok {
			for _, lhs := range assignment.Lhs {
				ast.Inspect(lhs, func(part ast.Node) bool {
					if field, ok := part.(*ast.SelectorExpr); ok {
						switch field.Sel.Name {
						case "Flags", "Pix", "Cur", "ManaCur", "Buffs", "SpellLvl":
							t.Errorf("fixture writes live outcome %s", field.Sel.Name)
						}
					}
					return true
				})
			}
		}
		if call, ok := n.(*ast.CallExpr); ok {
			if method, ok := call.Fun.(*ast.SelectorExpr); ok {
				switch method.Sel.Name {
				case "Show", "Hide", "SetHidden", "SetDraw", "Draw", "Sub_460EA0", "Nox_xxx_quickBarSetSpell":
					t.Errorf("fixture bypasses actual input/traversal: %s", method.Sel.Name)
				}
			} else if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "copy" {
				t.Error("fixture must not copy expected pixels into the actual frame")
			}
		}
		return true
	})
	source, err := os.ReadFile("legacy/e2e_trap_quickbar_observe.go")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(source), "PtrUint32") || strings.Contains(string(source), "C.int(uintptr") {
		t.Fatal("read-only observation must not truncate native pointers")
	}
}
