package opennox

import (
	"context"
	"reflect"
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

func TestConsoleServerCommandNormalizesInput(t *testing.T) {
	consoleCommandTestFlags(t)
	old := noxConsole
	t.Cleanup(func() { noxConsole = old })
	noxConsole = console.New(&consoleCapture{})
	noxConsole.Localize(strman.New())
	var got []string
	noxConsole.Register(&console.Command{Token: "probe", Flags: console.Server | console.Cheat, Func: func(ctx context.Context, _ *console.Console, args []string) bool {
		if console.IsClient(ctx) || !console.IsCheats(ctx) {
			t.Fatal("server execution lost its original context")
		}
		got = append([]string(nil), args...)
		return true
	}})
	execServerCmd("  PrObE\t\"한글 Red\" OFF  ")
	if !reflect.DeepEqual(got, []string{"한글 Red", "OFF"}) {
		t.Fatalf("server command arguments changed: %#v", got)
	}
}
