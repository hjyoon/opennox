package legacy

import (
	"bytes"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestPlayerIncomingNative4DDF60KeepsPreInitPosition(t *testing.T) {
	old := noxflags.GetGame()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(old)
	})
	for _, coop := range []bool{false, true} {
		noxflags.ResetGame()
		if coop {
			noxflags.SetGame(noxflags.GameModeCoopTeam)
		}
		position := types.Pointf{X: 123.5, Y: -45.25}
		oldPosition := types.Pointf{X: 1, Y: 2}
		unit := &server.Object{PosVec: position}
		player := &server.Player{PlayerUnit: unit, Field3676: 2, Field4700: 19, Pos3632Vec: oldPosition}
		initializeIncomingPlayer4DDF60(player, position, func() {
			if player.Field4700 != 0 || player.Field3676 != 2 {
				t.Fatalf("fields at initialization = %d/%d", player.Field4700, player.Field3676)
			}
			unit.PosVec = types.Pointf{X: 300, Y: 400}
		})
		want := position
		if coop {
			want = oldPosition
		}
		if player.Field4700 != 0 || player.Field3676 != 3 || player.Pos3632Vec != want {
			t.Fatalf("incoming state with coop=%v = %d/%d/%v; want position %v", coop, player.Field4700, player.Field3676, player.Pos3632Vec, want)
		}
	}
}

func TestPlayerIncomingCampingMarkers4DE270(t *testing.T) {
	world := objectStatusTestServer(t, 1)
	var joining *server.Player
	for i, camping := range []uint32{1, 0, 1} {
		code := 0x400 + i
		player := world.Players.NewRaw(code)
		unit, freeUnit := alloc.NewHigh(server.Object{})
		t.Cleanup(freeUnit)
		data, freeData := alloc.NewHigh(server.PlayerUpdateData{})
		t.Cleanup(freeData)
		*unit = server.Object{ObjClass: object.ClassPlayer, NetCode: uint32(code), TypeInd: 0x2345, UpdateData: unsafe.Pointer(data)}
		data.Player, data.IsCamping = player, camping
		player.PlayerUnit = unit
		if i == 0 {
			joining = player
		}
	}
	calls := 0
	world.NetSendPacketXxx = func(recipient int, buf []byte, related *server.Object, remove, sequence int) int {
		calls++
		want := []byte{0xd2, 0x02, 0x04, 0x45, 0x23, 1, 2}
		if !bytes.Equal(buf, want) || recipient != joining.Index() || related != nil || remove != 1 || sequence != 0 {
			t.Fatalf("incoming marker = %x to %d, related=%p, remove=%d, sequence=%d", buf, recipient, related, remove, sequence)
		}
		return 17
	}
	if result := world.SendIncomingCampingMarkers4DE270(joining); result != 17 || calls != 1 {
		t.Fatalf("incoming marker result = %d, calls = %d", result, calls)
	}
}
