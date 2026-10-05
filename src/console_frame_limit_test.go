package opennox

import (
	"context"
	"testing"

	"github.com/opennox/libs/console"
)

func TestConsoleFrameRateLimiterCommandsAndPermission(t *testing.T) {
	consoleCommandTestFlags(t)
	c, output := consoleCommandTestConsole(t)
	old := useFrameLimit
	t.Cleanup(func() { useFrameLimit = old })
	useFrameLimit = false
	c.SetCheats(false)
	consoleCommandTestExec(c, "SET FRAMERATELIMITER")
	if useFrameLimit {
		t.Fatal("cheat-gated frame limiter ran without authorization")
	}
	c.SetCheats(true)
	c.Exec(console.AsClient(context.Background()), normalizeConsoleInput(c, "SET FRAMERATELIMITER"))
	if useFrameLimit {
		t.Fatal("server-only frame limiter ran in client context")
	}
	if !consoleCommandTestExec(c, "SET FRAMERATELIMITER") || !useFrameLimit {
		t.Fatal("authorized original set command did not enable the limiter")
	}
	before := len(output.lines)
	consoleCommandTestExec(c, "UNSET FRAMERATELIMITER extra")
	// ParseToken reports handled commands as true even when their handler
	// rejects arguments and prints help. Inspect the real state and output.
	if !useFrameLimit || len(output.lines) != before+1 || output.lines[before] != c.HelpString(consoleCommandAt(c, "unset frameratelimiter")) {
		t.Fatal("invalid unset arguments changed state or failed to print help")
	}
	c.SetCheats(false)
	if !consoleCommandTestExec(c, "UNSET FRAMERATELIMITER") || useFrameLimit {
		t.Fatal("original unset command did not restore the disabled limiter")
	}
	c.SetCheats(true)
	before = len(output.lines)
	consoleCommandTestExec(c, "SET FRAMERATELIMITER extra")
	if useFrameLimit || len(output.lines) != before+1 || output.lines[before] != c.HelpString(consoleCommandAt(c, "set frameratelimiter")) {
		t.Fatal("invalid set arguments changed state or failed to print help")
	}
}
