package legacy

import (
	"testing"
)

type netLineMessagePlayer4D9EB0 struct {
	index byte
}

type netLineMessageUpdateData4D9EB0 struct {
	player *netLineMessagePlayer4D9EB0
}

func TestNetLineMessagePlayerIndex4D9EB0(t *testing.T) {
	player := &netLineMessagePlayer4D9EB0{index: 0xfe}
	updateData := &netLineMessageUpdateData4D9EB0{player: player}
	var events []string
	got := netLineMessagePlayerIndex4D9EB0(
		updateData,
		func(updateData *netLineMessageUpdateData4D9EB0) *netLineMessagePlayer4D9EB0 {
			events = append(events, "player")
			return updateData.player
		},
		func(player *netLineMessagePlayer4D9EB0) byte {
			events = append(events, "index")
			return player.index
		},
	)
	if want := byte(0xfe); got != want {
		t.Fatalf("player index = %#x, want %#x", got, want)
	}
	if got, want := len(events), 2; got != want || events[0] != "player" || events[1] != "index" {
		t.Fatalf("events = %v, want [player index]", events)
	}

	player.index = 0x7b
	if got, want := netLineMessagePlayerIndex4D9EB0(
		updateData,
		func(updateData *netLineMessageUpdateData4D9EB0) *netLineMessagePlayer4D9EB0 {
			return updateData.player
		},
		func(player *netLineMessagePlayer4D9EB0) byte { return player.index },
	), byte(0x7b); got != want {
		t.Fatalf("reloaded player index = %#x, want %#x", got, want)
	}
}

func TestNetLineMessagePlayerIndex4D9EB0FaultsOnMissingPointers(t *testing.T) {
	tests := []struct {
		name       string
		updateData *netLineMessageUpdateData4D9EB0
	}{
		{name: "update data", updateData: nil},
		{name: "player", updateData: &netLineMessageUpdateData4D9EB0{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("missing legacy pointer did not fault")
				}
			}()
			_ = netLineMessagePlayerIndex4D9EB0(
				test.updateData,
				func(updateData *netLineMessageUpdateData4D9EB0) *netLineMessagePlayer4D9EB0 {
					return updateData.player
				},
				func(player *netLineMessagePlayer4D9EB0) byte { return player.index },
			)
		})
	}
}
