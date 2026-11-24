package legacy

import (
	"testing"

	"github.com/opennox/opennox/v1/common/memmap"
)

func TestRemoteConsoleGrantsNativeLifetime4D11A0(t *testing.T) {
	world := objectStatusTestServer(t, 3)
	sub_4D11D0()
	flag := memmap.PtrUint32(0x5D4594, 1548504)
	old := *flag
	*flag = 0
	t.Cleanup(func() {
		sub_4D11D0()
		*flag = old
	})
	sub_4D1210(-1)
	sub_4D1210(32)
	sub_4D1210(0)
	if len(remoteConsoleGrants) != 0 {
		t.Fatal("inactive player received a remote console grant")
	}

	first := world.Players.NewRaw(15)
	second := world.Players.NewRaw(16)
	if first.Index() != 0 || second.Index() != 1 {
		t.Fatalf("test players occupy slots %d/%d", first.Index(), second.Index())
	}
	sub_4D11A0()
	sub_4D1210(0)
	sub_4D1210(1)
	sub_4D1210(0)
	sub_4D11A0()
	if *flag != 1 || len(remoteConsoleGrants) != 2 || sub_4D12A0(0) != 1 || sub_4D12A0(1) != 1 {
		t.Fatal("remote console initialization or duplicate detection lost grants")
	}
	marker := remoteConsoleGrants[0].token
	if sub_4D1250(0) != marker || sub_4D1250(0) != nil || sub_4D12A0(0) != 0 || sub_4D12A0(1) != 1 {
		t.Fatal("remote console revocation did not remove exactly one grant")
	}
	first.Active = 0
	replacement := world.Players.NewRaw(17)
	if replacement.Index() != 0 || sub_4D12A0(0) != 0 {
		t.Fatal("reused player slot inherited a remote console grant")
	}

	sub_4D11D0()
	if *flag != 1 || len(remoteConsoleGrants) != 0 || sub_4D12A0(1) != 0 {
		t.Fatal("remote console shutdown failed to clear grants")
	}
}
