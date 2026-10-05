package opennox

import (
	"context"
	"log/slog"
	"testing"
)

func TestConsoleLogCommandChangesGUIFilter(t *testing.T) {
	consoleCommandTestFlags(t)
	c, _ := consoleCommandTestConsole(t)
	g := &guiConsole{enabled: true}
	ctx := context.Background()
	if g.Enabled(ctx, slog.LevelInfo) || !g.Enabled(ctx, slog.LevelWarn) {
		t.Fatal("default console warning filter changed")
	}
	consoleCommandTestExec(c, "LOG CONSOLE")
	if !g.Enabled(ctx, slog.LevelDebug) || !g.Enabled(ctx, slog.LevelInfo) {
		t.Fatal("log console flag did not enable lower-level output")
	}
	g.Enable(false)
	if g.Enabled(ctx, slog.LevelError) {
		t.Fatal("disabled GUI console still accepts logs")
	}
	g.Enable(true)
	consoleCommandTestExec(c, "LOG STOP")
	if g.Enabled(ctx, slog.LevelInfo) || !g.Enabled(ctx, slog.LevelWarn) {
		t.Fatal("log stop did not restore the console filter")
	}
}
