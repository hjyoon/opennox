package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestQuestPlayerCount4E3CE0CEntryNativePointersAndExactState(t *testing.T) {
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	oldServer := GetServer
	GetServer = func() Server { return &questCanJoinLegacyServer4E4100{srv: srv} }
	t.Cleanup(func() { GetServer = oldServer })
	oldGame, oldEngine := noxflags.GetGame(), noxflags.GetEngine()
	noxflags.ResetGame()
	noxflags.ResetEngine()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldGame)
		noxflags.ResetEngine()
		noxflags.SetEngine(oldEngine)
	})
	if got := Nox_xxx_player_4E3CE0(); got != 0 {
		t.Fatalf("empty C entry = %d, want 0", got)
	}
	states := []uint32{0, 1, 2, math.MaxUint32, 0x80000000, 1, 1}
	indices := []int{0, 1, 2, 3, 4, 5, server.HostPlayerIndex}
	for i, state := range states {
		player := srv.Players.NewRaw(100 + i)
		if player == nil {
			t.Fatal("player fixture allocation failed")
		}
		update, freeUpdate := alloc.New(server.PlayerUpdateData{})
		t.Cleanup(freeUpdate)
		unit, freeUnit := alloc.New(server.Object{})
		t.Cleanup(freeUnit)
		update.Player = player
		*unit = server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
		player.PlayerUnit, player.PlayerInd, player.Field4792 = unit, uint8(indices[i]), state
		if unsafe.Sizeof(uintptr(0)) > 4 {
			for _, ptr := range []unsafe.Pointer{unsafe.Pointer(player), unsafe.Pointer(update), unsafe.Pointer(unit)} {
				if uintptr(ptr) <= math.MaxUint32 {
					t.Fatalf("native pointer = %p, want above 4 GiB", ptr)
				}
			}
		}
	}
	for _, tc := range []struct {
		host, noRendering bool
		want              int
	}{
		{false, false, 3}, {false, true, 3}, {true, false, 3}, {true, true, 2},
	} {
		noxflags.ResetGame()
		noxflags.ResetEngine()
		if tc.host {
			noxflags.SetGame(noxflags.GameHost)
		}
		if tc.noRendering {
			noxflags.SetEngine(noxflags.EngineNoRendering)
		}
		if got := Nox_xxx_player_4E3CE0(); got != tc.want {
			t.Fatalf("host=%t noRendering=%t C entry = %d, want %d", tc.host, tc.noRendering, got, tc.want)
		}
		for i, player := range srv.Players.List() {
			if player.Field4792 != states[i] || player.PlayerInd != uint8(indices[i]) ||
				player.PlayerUnit == nil || player.PlayerUnit.UpdateData == nil ||
				(*server.PlayerUpdateData)(player.PlayerUnit.UpdateData).Player != player {
				t.Fatal("read-only C count changed native player fields")
			}
		}
	}
}
