package opennox

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/opennox/libs/client/keybind"
	"github.com/opennox/libs/console"
	"github.com/opennox/libs/strman"
)

type consoleCapture struct{ lines []string }

func (p *consoleCapture) Print(_ console.Color, s string)                     { p.lines = append(p.lines, s) }
func (p *consoleCapture) Printf(cl console.Color, s string, _ ...interface{}) { p.Print(cl, s) }

func TestConsoleInputCommandPathAndArguments(t *testing.T) {
	c := console.New(&consoleCapture{})
	c.Register(&console.Command{Token: "set", Flags: console.ClientServer, Sub: []*console.Command{
		{Token: "name", Flags: console.ClientServer}, {Token: "spell", Flags: console.ClientServer},
		{Token: "quality", Flags: console.ClientServer, Sub: []*console.Command{{Token: "LAN", Flags: console.ClientServer}}},
		{Token: "monsters", Flags: console.ClientServer},
	}})
	for _, token := range []string{"say", "exec", "sysop", "watch"} {
		c.Register(&console.Command{Token: token, Flags: console.ClientServer})
	}
	c.Register(&console.Command{Token: "script", Flags: console.ClientServer, Raw: true})
	c.Localize(strman.New(), "on", "off", "respawn")
	c.Commands()[len(c.Commands())-1].Alias = "스크립트"
	cases := []struct{ input, want string }{
		{"  SET\tNaMe Red Dragon  ", `set name "Red" "Dragon"`},
		{`set name "off"`, `set name "off"`},
		{`SET SPELL "마법 화살" OFF`, `set spell "마법 화살" "off"`},
		{"sEt QuAliTy lan", "set quality LAN"},
		{"HELP SET QUALITY lan", "help set quality LAN"},
		{"QUES\tSET\tNAME", "ques set name"},
		{"set monsters ReSpAwN OFF", `set monsters "respawn" "off"`},
		{"SAY\tAbC  off 한글", "say AbC  off 한글"},
		{"EXEC MyFunction", "exec MyFunction"},
		{"SYSOP SecretCase", "sysop SecretCase"},
		{"SCRIPT print(\"SET OFF\")  +  X", "script print(\"SET OFF\")  +  X"},
		{"스크립트 print(\"한글\")", "script print(\"한글\")"},
		{"RaCoIaWs", "racoiaws"},
		{"watch SET", `watch "SET"`},
		{"set nonexistent NAME", `set "nonexistent" "NAME"`},
		{" ", ""}, {"bogus AbC", "bogus AbC"},
	}
	for _, tt := range cases {
		t.Run(tt.input, func(t *testing.T) {
			if got := normalizeConsoleInput(c, tt.input); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestConsoleInputPreservesAliasNamedArguments(t *testing.T) {
	c := console.New(&consoleCapture{})
	var got []string
	c.Register(&console.Command{Token: "watch", Flags: console.ClientServer, Func: func(_ context.Context, _ *console.Console, args []string) bool {
		got = append([]string(nil), args...)
		return true
	}})
	c.Register(&console.Command{Token: "cmd", Flags: console.ClientServer})
	sm := strman.New()
	fixture := filepath.Join(t.TempDir(), "strings.json")
	if err := os.WriteFile(fixture, []byte(`{"entries":[{"id":"cmd_token:cmd","vals":[{"str":"별칭"}]}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := sm.ReadJSON(fixture); err != nil {
		t.Fatal(err)
	}
	c.Localize(sm)
	// A genuine command alias must not be applied to a player's name.
	c.Exec(console.AsClient(context.Background()), `watch 별칭`)
	if !reflect.DeepEqual(got, []string{"cmd"}) {
		t.Fatalf("fixture did not activate shared parser alias conversion: %#v", got)
	}
	input := `WATCH "Red Dragon" 별칭 한글`
	if !c.Exec(console.AsClient(context.Background()), normalizeConsoleInput(c, input)) {
		t.Fatal("not executed")
	}
	if want := []string{"Red Dragon", "별칭", "한글"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
}

func TestConsoleInputMacroKeysAndLocalizedSwitches(t *testing.T) {
	c := console.New(&consoleCapture{})
	var got []string
	c.Register(&console.Command{Token: "set", Flags: console.ClientServer, Sub: []*console.Command{
		{Token: "cycle", Flags: console.ClientServer, Func: func(_ context.Context, _ *console.Console, args []string) bool {
			got = append([]string(nil), args...)
			return true
		}},
	}})
	sm := strman.New()
	fixture := filepath.Join(t.TempDir(), "strings.json")
	if err := os.WriteFile(fixture, []byte(`{"entries":[{"id":"cmd_token:set","vals":[{"str":"설정"}]},{"id":"cmd_token:cycle","vals":[{"str":"순환"}]},{"id":"cmd_token:on","vals":[{"str":"켜기"}]},{"id":"cmd_token:off","vals":[{"str":"끄기"}]}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := sm.ReadJSON(fixture); err != nil {
		t.Fatal(err)
	}
	c.Localize(sm, "on", "off")
	ctx := console.AsServer(context.Background())
	for _, input := range []string{"설정 순환 켜기", "SET CYCLE ON"} {
		if !c.Exec(ctx, normalizeConsoleInput(c, input)) || !reflect.DeepEqual(got, []string{"on"}) {
			t.Fatalf("input %q: %#v", input, got)
		}
	}
	c.SetExec(func(_ context.Context, command string) bool {
		return c.Exec(ctx, normalizeConsoleInput(c, command))
	})
	if !c.Exec(ctx, normalizeConsoleInput(c, "BIND f2 SET CYCLE OFF")) {
		t.Fatal("bind lowercase key")
	}
	if !c.ExecMacros(ctx, keybind.KeyByName("F2")) || !reflect.DeepEqual(got, []string{"off"}) {
		t.Fatalf("macro execution: %#v", got)
	}
	if !c.Exec(ctx, normalizeConsoleInput(c, "UNBIND f2")) || c.ExecMacros(ctx, keybind.KeyByName("F2")) {
		t.Fatal("unbind lowercase key")
	}
}

func TestConsoleInputKeepsCheatAndSideGates(t *testing.T) {
	c := console.New(&consoleCapture{})
	calls := 0
	c.Register(&console.Command{Token: "privileged", Flags: console.Server | console.Cheat, Func: func(context.Context, *console.Console, []string) bool { calls++; return true }})
	c.Localize(strman.New())
	input := normalizeConsoleInput(c, "PRIVILEGED")
	c.Exec(console.AsClient(context.Background()), input)
	c.Exec(console.AsServer(context.Background()), input)
	if calls != 0 {
		t.Fatal("permission bypass")
	}
	c.SetCheats(true)
	c.Exec(console.AsClient(context.Background()), input)
	if calls != 0 {
		t.Fatal("server-only bypass")
	}
	c.Exec(console.AsServer(context.Background()), input)
	if calls != 1 {
		t.Fatalf("server calls = %d", calls)
	}
}
