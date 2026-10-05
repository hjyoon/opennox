package opennox

import (
	"context"
	"testing"

	"github.com/opennox/opennox/v1/common/memmap"
)

func TestConsoleTelnetGameExitClosesListener(t *testing.T) {
	oldContinue, oldMenu := mainloopContinue, continueMenuOrHost
	oldConnected := memmap.Uint32(0x5D4594, 815764)
	*memmap.PtrUint32(0x5D4594, 815764) = 0
	t.Cleanup(func() {
		consoleTelnetTestStop(t, &noxTelnet)
		mainloopContinue, continueMenuOrHost = oldContinue, oldMenu
		*memmap.PtrUint32(0x5D4594, 815764) = oldConnected
	})
	if err := noxTelnet.start(0, func(context.Context, bool, string) (bool, []string) { return true, nil }, "Password", "Bad password"); err != nil {
		t.Fatal(err)
	}
	noxTelnet.mu.Lock()
	run := noxTelnet.active
	noxTelnet.mu.Unlock()
	nox_game_exit_xxx2()
	if noxTelnet.address() != "" || mainloopContinue || continueMenuOrHost {
		t.Fatal("game exit left its console listener/game loop active")
	}
	var stopped consoleTelnetService
	stopped.active = run
	consoleTelnetTestStop(t, &stopped)
}
