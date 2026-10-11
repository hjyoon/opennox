package opennox

import (
	"bytes"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

// Compile the actual full function body, with bounded service receipts. The
// headless campaign scenario separately exercises real F1/Enter, map loading,
// real summon/charm completion, ownership, HP and native client drawables.
func TestMapLoadCompanionRepositionContract(t *testing.T) {
	source, err := os.ReadFile("parsecmd_maps.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "parsecmd_maps.go", source, 0)
	if err != nil {
		t.Fatal(err)
	}
	var body []byte
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "noxLoadMap" {
			if body != nil || fn.Recv != nil || fn.Body == nil {
				t.Fatal("ambiguous loader body")
			}
			body = source[fset.Position(fn.Pos()).Offset:fset.Position(fn.End()).Offset]
		}
	}
	fixture, err := os.ReadFile("testdata/map_load_companion_fixture.go.txt")
	if err != nil {
		t.Fatal(err)
	}
	const marker = "// PRODUCTION_NOX_LOAD_MAP"
	if body == nil || bytes.Count(fixture, []byte(marker)) != 1 {
		t.Fatal("missing full production body or marker")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "main.go"), bytes.Replace(fixture, []byte(marker), body, 1), 0600); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(dir, "map-load-companions")
	cmd := exec.Command(filepath.Join(runtime.GOROOT(), "bin", "go"), "build", "-p=2", "-o", binary, "main.go")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GO111MODULE=off")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile unchanged loader body: %v\n%s", err, output)
	}
	type input struct {
		Name                                  string
		Game, Mode                            uint32
		Force, CheckError, StatError, Pending bool
		Teams                                 int
	}
	type result struct {
		Error   string
		Pending bool
		Busy    int
		Map     string
		Calls   []string
	}
	tests := []struct {
		name             string
		in               input
		queued, relocate bool
		errorPart        string
		teams            bool
	}{
		{"solo-fresh", input{Name: "con02a", Game: 0x800, Mode: 0x200}, true, true, "", false},
		{"solo-extension", input{Name: "con02a.map", Game: 0x800, Mode: 0x200}, true, true, "", false},
		{"solo-existing-pending", input{Name: "con02a", Game: 0x800, Mode: 0x200, Pending: true}, true, true, "", false},
		{"solo-forced-non-solo", input{Name: "estate", Game: 0x800, Mode: 0x100, Force: true}, true, true, "", false},
		{"solo-generated", input{Name: "#generated", Game: 0x800, Mode: 0x200}, true, true, "", false},
		{"solo-unknown-game-flags", input{Name: "con02a", Game: 0x1800803, Mode: 0x200}, true, true, "", false},
		{"quest-matching", input{Name: "quest", Game: 0x3000, Mode: 0x1000}, true, false, "", false},
		{"quest-forced-solo", input{Name: "con02a", Game: 0x3000, Mode: 0x200, Force: true}, true, false, "", false},
		{"quest-existing-pending", input{Name: "quest", Game: 0x3000, Mode: 0x1000, Pending: true}, true, false, "", false},
		{"online-coop-bit", input{Name: "quest", Game: 0x3800, Mode: 0x1000}, true, false, "", false},
		{"ctf-matching", input{Name: "estate", Game: 0x2020, Mode: 0x20}, true, false, "", false},
		{"online-forced-empty-mode", input{Name: "estate", Game: 0x2020, Mode: 0, Force: true}, true, false, "", false},
		{"chat-two-teams", input{Name: "estate", Game: 0x2080, Mode: 0x20, Teams: 2}, true, false, "", false},
		{"chat-create-teams", input{Name: "estate", Game: 0x2080, Mode: 0x40, Teams: 1}, true, false, "", true},
		{"chat-no-team-mode", input{Name: "estate", Game: 0x2080, Mode: 0x100}, true, false, "", false},
		{"solo-reject-non-solo", input{Name: "estate", Game: 0x800, Mode: 0x100}, false, false, "non-Solo", false},
		{"solo-reject-pending-preserved", input{Name: "estate", Game: 0x800, Mode: 0x100, Pending: true}, false, false, "non-Solo", false},
		{"quest-reject-solo", input{Name: "con02a", Game: 0x3000, Mode: 0x200}, false, false, "Solo", false},
		{"online-reject-zero", input{Name: "estate", Game: 0x2020, Mode: 0}, false, false, "Solo", false},
		{"online-reject-mismatch", input{Name: "estate", Game: 0x2020, Mode: 0x40}, false, false, "different mode", false},
		{"invalid-file", input{Name: "missing", Game: 0x800, Mode: 0x200, CheckError: true}, false, false, "checking map", false},
		{"invalid-file-pending-preserved", input{Name: "missing", Game: 0x800, Mode: 0x200, CheckError: true, Pending: true}, false, false, "checking map", false},
		{"inaccessible-file", input{Name: "missing", Game: 0x800, Mode: 0x200, StatError: true}, false, false, "Cannot access", false},
		{"inaccessible-forced-file", input{Name: "missing", Game: 0x800, Mode: 0x100, Force: true, StatError: true, Pending: true}, false, false, "Cannot access", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.in)
			if err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(binary)
			cmd.Stdin = bytes.NewReader(data)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("native loader fixture: %v\n%s", err, output)
			}
			var got result
			if err := json.Unmarshal(output, &got); err != nil {
				t.Fatalf("receipt: %v %s", err, output)
			}
			if (got.Error == "") != tc.queued || tc.errorPart != "" && !strings.Contains(got.Error, tc.errorPart) {
				t.Errorf("error=%q want=%q queued=%t", got.Error, tc.errorPart, tc.queued)
			}
			wantPending := tc.in.Pending || tc.relocate
			if got.Pending != wantPending {
				t.Errorf("companion relocation pending=%t want=%t at accepted cheat-load", got.Pending, wantPending)
			}
			busy := 0
			wantMap := ""
			if tc.queued {
				busy = 1
				wantMap = tc.in.Name
				if !strings.HasSuffix(wantMap, ".map") {
					wantMap += ".map"
				}
			}
			if got.Busy != busy || got.Map != wantMap {
				t.Errorf("busy/map=%d/%q want=%d/%q", got.Busy, got.Map, busy, wantMap)
			}
			var effects []string
			for _, v := range got.Calls {
				if v == "busy" || v == "relocate" || v == "queue" || v == "pause" || v == "create-teams" {
					effects = append(effects, v)
				}
			}
			var want []string
			if tc.teams {
				want = append(want, "create-teams")
			}
			if tc.queued {
				want = append(want, "busy")
				if tc.relocate {
					want = append(want, "relocate")
				}
				want = append(want, "queue", "pause")
			}
			if !reflect.DeepEqual(effects, want) {
				t.Errorf("transition service order=%v want=%v", effects, want)
			}
			if strings.HasPrefix(tc.in.Name, "#") {
				for _, v := range got.Calls {
					if strings.HasPrefix(v, "check:") || strings.HasPrefix(v, "stat:") {
						t.Errorf("generated map touched disk: %s", v)
					}
				}
			}
		})
	}
}
