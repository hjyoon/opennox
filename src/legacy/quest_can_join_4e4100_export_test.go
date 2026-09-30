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

type questCanJoinLegacyServer4E4100 struct {
	Server
	srv *server.Server
}

func (s *questCanJoinLegacyServer4E4100) S() *server.Server { return s.srv }

func TestQuestCanJoinCEntry4E4100PreservesHighPlayerPointers(t *testing.T) {
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	oldGetServer := GetServer
	GetServer = func() Server { return &questCanJoinLegacyServer4E4100{srv: srv} }
	t.Cleanup(func() { GetServer = oldGetServer })
	oldGame, oldEngine := noxflags.GetGame(), noxflags.GetEngine()
	noxflags.ResetGame()
	noxflags.ResetEngine()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldGame)
		noxflags.ResetEngine()
		noxflags.SetEngine(oldEngine)
	})
	if got := Sub_4E4100(); got != 1 {
		t.Fatalf("empty C entry = %d, want 1", got)
	}
	for i, state := range []uint32{1, 2, math.MaxUint32, 3, 0x80000000, 7, 1} {
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
		player.PlayerUnit = unit
		player.Field4792 = state
		if unsafe.Sizeof(uintptr(0)) > 4 {
			for _, pointer := range []unsafe.Pointer{unsafe.Pointer(player), unsafe.Pointer(update), unsafe.Pointer(unit)} {
				if uintptr(pointer) <= math.MaxUint32 {
					t.Fatalf("native pointer = %p, want above 4 GiB", pointer)
				}
			}
		}
		want := uint32(1)
		if i >= 5 {
			want = 0
		}
		if got := Sub_4E4100(); got != want {
			t.Fatalf("C entry with %d players = %d, want %d", i+1, got, want)
		}
		if player.Field4792 != state || player.PlayerUnit != unit || update.Player != player ||
			unit.UpdateData != unsafe.Pointer(update) {
			t.Fatal("C entry mutated native player fields")
		}
	}
}
