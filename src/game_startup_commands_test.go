package opennox

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/opennox/libs/console"
	noxflags "github.com/opennox/opennox/v1/common/flags"
)

func TestGameStartupCommandsPreserveRuleOrderAndPermissions(t *testing.T) {
	for _, tc := range []struct {
		name              string
		server, dedicated bool
	}{
		{name: "graphical-host", server: true},
		{name: "dedicated-host", server: true, dedicated: true},
		{name: "client"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			consoleCommandTestFlags(t)
			oldConsole, oldServer, oldExec, oldWait := noxConsole, isServer, serverExec, consoleWaitSysOpPass
			t.Cleanup(func() {
				noxConsole, isServer, serverExec, consoleWaitSysOpPass = oldConsole, oldServer, oldExec, oldWait
			})
			c, _ := consoleCommandTestConsole(t)
			c.SetCheats(false)
			noxConsole, isServer, consoleWaitSysOpPass = c, tc.server, false
			if tc.server {
				noxflags.SetGame(noxflags.GameHost)
			}
			if tc.dedicated {
				noxflags.SetEngine(noxflags.EngineNoRendering)
			}
			var events []string
			c.Register(&console.Command{Token: "startuprule", Flags: console.ClientServer, Func: func(ctx context.Context, _ *console.Console, args []string) bool {
				if console.IsServer(ctx) != tc.server || console.IsCheats(ctx) != tc.dedicated {
					t.Fatal("startup rule context changed")
				}
				events = append(events, "rule:"+args[0])
				return true
			}})
			c.Register(&console.Command{Token: "startupprobe", Flags: console.Server | console.Cheat, Func: func(ctx context.Context, _ *console.Console, args []string) bool {
				if !console.IsServer(ctx) || !console.IsCheats(ctx) {
					t.Fatal("startup CLI command context changed")
				}
				events = append(events, "cli:"+args[0])
				return true
			}})
			cycle := consoleCommandAt(c, "set cycle")
			cycle.LegacyFunc = nil
			cycle.Func = func(ctx context.Context, _ *console.Console, args []string) bool {
				if !console.IsServer(ctx) || !console.IsCheats(ctx) || len(args) != 1 || args[0] != "on" {
					t.Fatal("default cycle context or argument changed")
				}
				events = append(events, "cycle:on")
				return true
			}
			c.Localize(c.Strings(), "on", "off")
			dir := t.TempDir()
			t.Chdir(dir)
			if err := os.WriteFile(filepath.Join(dir, "autoexec.rul"), []byte("STARTUPRULE \"Rule 한글\"\n"), 0600); err != nil {
				t.Fatal(err)
			}
			serverExec = []string{"", " \t ", "STARTUPPROBE \"Keep This 한글\"", "unknown-startup-command", "STARTUPPROBE Tail"}
			beforeExec := append([]string(nil), serverExec...)
			beforeGame, beforeEngine := noxflags.GetGame(), noxflags.GetEngine()
			execGameStartupCommands(context.Background())
			want := []string{"rule:Rule 한글"}
			if tc.server {
				want = append(want, "cycle:on", "cli:Keep This 한글", "cli:Tail")
			}
			if !reflect.DeepEqual(events, want) {
				t.Fatalf("startup events=%q, want %q", events, want)
			}
			if noxflags.GetGame() != beforeGame || noxflags.GetEngine() != beforeEngine || !reflect.DeepEqual(serverExec, beforeExec) || c.Cheats() {
				t.Fatal("startup dispatcher changed flags, CLI input, or persistent cheat permission")
			}
		})
	}
}

func TestGameStartupCommandsMissingRuleKeepsCLI(t *testing.T) {
	consoleCommandTestFlags(t)
	oldConsole, oldServer, oldExec, oldWait := noxConsole, isServer, serverExec, consoleWaitSysOpPass
	t.Cleanup(func() {
		noxConsole, isServer, serverExec, consoleWaitSysOpPass = oldConsole, oldServer, oldExec, oldWait
	})
	c, _ := consoleCommandTestConsole(t)
	c.SetCheats(false)
	noxConsole, isServer, serverExec, consoleWaitSysOpPass = c, true, []string{"startupprobe Tail"}, false
	noxflags.SetGame(noxflags.GameHost)
	noxflags.SetEngine(noxflags.EngineNoRendering)
	var events []string
	c.Register(&console.Command{Token: "startupprobe", Flags: console.Server | console.Cheat, Func: func(_ context.Context, _ *console.Console, args []string) bool {
		events = append(events, args[0])
		return true
	}})
	cycle := consoleCommandAt(c, "set cycle")
	cycle.LegacyFunc = nil
	cycle.Func = func(context.Context, *console.Console, []string) bool { events = append(events, "cycle"); return true }
	c.Localize(c.Strings(), "on", "off")
	t.Chdir(t.TempDir())
	execGameStartupCommands(context.Background())
	if !reflect.DeepEqual(events, []string{"cycle", "Tail"}) {
		t.Fatalf("missing optional rule changed startup: %q", events)
	}
}

// Runtime checks separately exercise actual map load, rule execution, native
// console commands and authenticated shutdown. This guards their placement
// outside the tick loop and after the successful initial-map error boundary.
func TestDedicatedServerRunsStartupCommandsAfterInitialMap(t *testing.T) {
	f, err := parser.ParseFile(token.NewFileSet(), "gameloop.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var body *ast.BlockStmt
	for _, decl := range f.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "MainLoop" {
			body = fn.Body
		}
	}
	if body == nil {
		t.Fatal("dedicated server loop missing")
	}
	var mapGuard, startup, tick token.Pos
	for _, stmt := range body.List {
		switch stmt := stmt.(type) {
		case *ast.IfStmt:
			if stmt.Init == nil {
				continue
			}
			ast.Inspect(stmt.Init, func(node ast.Node) bool {
				if call, ok := node.(*ast.CallExpr); ok {
					if selector, ok := call.Fun.(*ast.SelectorExpr); ok && selector.Sel.Name == "InitialMapLoad" {
						mapGuard = stmt.End()
					}
				}
				return true
			})
		case *ast.ExprStmt:
			if call, ok := stmt.X.(*ast.CallExpr); ok {
				if name, ok := call.Fun.(*ast.Ident); ok && name.Name == "execGameStartupCommands" {
					startup = stmt.Pos()
				}
			}
		case *ast.ForStmt:
			tick = stmt.Pos()
		}
	}
	if mapGuard == token.NoPos || startup == token.NoPos || tick == token.NoPos || startup <= mapGuard || startup >= tick {
		t.Fatal("dedicated startup must run once after initial map success and before the tick loop")
	}
}
