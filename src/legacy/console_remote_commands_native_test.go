package legacy

import (
	"context"
	"testing"

	"github.com/opennox/libs/console"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestRemoteConsoleDispatchKeepsAuthorizationAndModeGates(t *testing.T) {
	world := objectStatusTestServer(t, 1)
	player := world.Players.NewRaw(15)
	if player == nil || player.Index() != 0 {
		t.Fatal("remote console test needs player slot zero")
	}
	unit, freeUnit := alloc.NewHigh(server.Object{})
	player.PlayerUnit = unit
	defer func() {
		player.PlayerUnit = nil
		freeUnit()
	}()
	pl := (*nox_playerInfo)(player.C())
	const command = "show mem"
	text, freeText := CWString(command)
	defer freeText()

	previousFlags, previousCode := noxflags.GetGame(), ClientPlayerNetCode()
	previousExec, previousConsole := ExecConsoleCmd, GetConsole
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameModeArena)
	ClientSetPlayerNetCode(int(server.HostPlayerIndex))
	sub_4D11D0()
	t.Cleanup(func() {
		sub_4D11D0()
		noxflags.ResetGame()
		noxflags.SetGame(previousFlags)
		ClientSetPlayerNetCode(previousCode)
		ExecConsoleCmd, GetConsole = previousExec, previousConsole
	})
	con := console.New(console.NewMultiPrinter())
	GetConsole = func() *console.Console { return con }
	calls := 0
	ExecConsoleCmd = func(_ context.Context, got string) bool {
		calls++
		if got != command {
			t.Errorf("remote command = %q, want %q", got, command)
		}
		return true
	}
	dispatch := func(want int) {
		t.Helper()
		if got := nox_xxx_serverHandleClientConsole_443E90(pl, 2, text); got != 1 || calls != want {
			t.Fatalf("remote console result = %d, command calls = %d, want 1 and %d", got, calls, want)
		}
	}
	dispatch(0)
	sub_4D1210(0)
	dispatch(1)
	noxflags.SetGame(noxflags.GameFlag15)
	dispatch(1)
	noxflags.UnsetGame(noxflags.GameFlag15)
	sub_4D1250(0)
	dispatch(1)
	noxflags.SetGame(noxflags.GameModeCoop)
	dispatch(2)
}
