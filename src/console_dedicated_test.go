package opennox

import (
	"context"
	"testing"

	"github.com/opennox/libs/console"
	"github.com/opennox/libs/strman"
	noxflags "github.com/opennox/opennox/v1/common/flags"
)

func TestConsoleDedicatedOperatorCheats(t *testing.T) {
	consoleCommandTestFlags(t)
	old := noxConsole
	t.Cleanup(func() { noxConsole = old })
	noxConsole = console.New(&consoleCapture{})
	noxConsole.Localize(strman.New())
	calls := 0
	noxConsole.Register(&console.Command{Token: "probe", Flags: console.Server | console.Cheat, Func: func(context.Context, *console.Console, []string) bool { calls++; return true }})
	noxflags.SetGame(noxflags.GameHost)
	execConsoleCmdAuthed(context.Background(), "PROBE")
	if calls != 0 {
		t.Fatal("graphical host acquired cheat permission")
	}
	noxflags.SetEngine(noxflags.EngineNoRendering)
	execConsoleCmdAuthed(context.Background(), "PROBE")
	if calls != 1 {
		t.Fatal("dedicated operator command was rejected")
	}
	noxflags.UnsetGame(noxflags.GameHost)
	execConsoleCmdAuthed(context.Background(), "PROBE")
	if calls != 1 {
		t.Fatal("headless client acquired server/cheat permission")
	}
}
