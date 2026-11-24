package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestQuestObserverRecoveryChecksLiveUpdateAndPlayerState(t *testing.T) {
	player := server.Player{Field3680: 0x400011, Field4792: 1}
	update := server.PlayerUpdateData{Player: &player}
	unit := server.Object{UpdateData: unsafe.Pointer(&update)}
	if got := questObserverReadyPlayer4D7150(&unit); got != &player {
		t.Fatalf("eligible native player returned %p, want %p", got, &player)
	}
	for _, row := range []struct {
		name         string
		flags, state uint32
		exit, warp   *server.Object
	}{
		{"inactive", 0x10, 1, nil, nil},
		{"observing flag cleared", 1, 1, nil, nil},
		{"wrong stage", 0x11, 2, nil, nil},
		{"occupied exit", 0x11, 1, &server.Object{}, nil},
		{"occupied warp gate", 0x11, 1, nil, &server.Object{}},
	} {
		t.Run(row.name, func(t *testing.T) {
			player.Field3680, player.Field4792 = row.flags, row.state
			update.QuestExit, update.QuestWarpGate = row.exit, row.warp
			if got := questObserverReadyPlayer4D7150(&unit); got != nil {
				t.Fatalf("ineligible native player %p", got)
			}
		})
	}
	update.Player = nil
	if got := questObserverReadyPlayer4D7150(&unit); got != nil {
		t.Fatalf("missing native player returned %p", got)
	}
	unit.UpdateData = nil
	if got := questObserverReadyPlayer4D7150(&unit); got != nil {
		t.Fatalf("missing native update record returned %p", got)
	}
}
