package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestPlayerDisconnectClearsOnlyDepartedPeerSlots4DE7C0(t *testing.T) {
	world := objectStatusTestServer(t, 1)
	player := world.Players.NewRaw(21)
	unit, freeUnit := alloc.NewHigh(server.Object{})
	t.Cleanup(freeUnit)
	data, freeData := alloc.NewHigh(server.PlayerUpdateData{})
	t.Cleanup(freeData)
	*unit = server.Object{ObjClass: object.ClassPlayer, NetCode: 21, UpdateData: unsafe.Pointer(data)}
	data.Player, player.PlayerUnit = player, unit

	const departed = 31
	for i := range data.QuestPlayerState {
		data.QuestPlayerState[i] = 0x12345678
		data.RespawnMarkers[i] = 0x33
		data.QuestPlayerFlagsA[i] = 0x44
		data.QuestPlayerFlagsB[i] = 0x55
	}
	clearDisconnectedPeerState4DE7C0(world, departed)
	for i := range data.QuestPlayerState {
		state, marker, a, b := uint32(0x12345678), byte(0x33), byte(0x44), byte(0x55)
		if i == departed {
			state, marker, a, b = 0, 0, 0, 0
		}
		if data.QuestPlayerState[i] != state || data.RespawnMarkers[i] != marker ||
			data.QuestPlayerFlagsA[i] != a || data.QuestPlayerFlagsB[i] != b {
			t.Errorf("disconnected state index %d = %#x/%#x/%#x/%#x", i, data.QuestPlayerState[i],
				data.RespawnMarkers[i], data.QuestPlayerFlagsA[i], data.QuestPlayerFlagsB[i])
		}
	}
}
