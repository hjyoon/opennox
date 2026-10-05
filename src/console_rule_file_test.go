package opennox

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/opennox/libs/console"
	noxflags "github.com/opennox/opennox/v1/common/flags"
)

func TestConsoleRuleFilePreservesCaseAndQuotedPath(t *testing.T) {
	consoleCommandTestFlags(t)
	noxflags.SetGame(noxflags.GameHost)
	c, output := consoleCommandTestConsole(t)
	oldConsole, oldWait := noxConsole, consoleWaitSysOpPass
	noxConsole, consoleWaitSysOpPass = c, false
	t.Cleanup(func() { noxConsole, consoleWaitSysOpPass = oldConsole, oldWait })
	fixture := filepath.Join(t.TempDir(), "strings.json")
	if err := os.WriteFile(fixture, []byte(`{"entries":[{"id":"ExecutingRul","vals":[{"str":"Executing %S"}]}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := c.Strings().ReadJSON(fixture); err != nil {
		t.Fatal(err)
	}
	var payload string
	c.Register(&console.Command{Token: "ruleprobe", Flags: console.Server, Func: func(_ context.Context, _ *console.Console, args []string) bool {
		payload = strings.Join(args, " ")
		return true
	}})
	path := filepath.Join(t.TempDir(), "Mixed Case 한글.RUL")
	if err := os.WriteFile(path, []byte("RULEPROBE \"Keep This 한글\"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if !consoleCommandTestExec(c, "EXECRUL \""+path+"\"") || payload != "Keep This 한글" {
		t.Fatal("rule file did not execute through the native console parser")
	}
	if len(output.lines) == 0 || !strings.Contains(output.lines[0], path) {
		t.Fatal("rule path was case folded or given a duplicate extension")
	}
}
