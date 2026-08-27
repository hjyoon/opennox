package opennox

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

func TestPlayerIncrementElimDeath4D8D40UsesNativePlayerChainAndWraps(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})

	player := &server.Player{Field2140: math.MaxUint32}
	update := &server.PlayerUpdateData{Player: player}
	obj := &server.Object{
		ObjClass:   object.ClassPlayer | object.Class(0x80000000),
		UpdateData: unsafe.Pointer(update),
	}
	s := &Server{}
	s.playerIncrementElimDeath4D8D40(obj)
	if player.Field2140 != 0 {
		t.Fatalf("wrapped deaths = %08x, want 00000000", player.Field2140)
	}

	nonPlayer := &server.Object{ObjClass: object.Class(0x80000000)}
	s.playerIncrementElimDeath4D8D40(nonPlayer)
}

func TestPlayerIncrementElimDeath4D8D40PlayerNilUpdateFaults(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})

	defer func() {
		if recover() == nil {
			t.Fatal("Player nil UpdateData did not fault")
		}
	}()
	(&Server{}).playerIncrementElimDeath4D8D40(&server.Object{ObjClass: object.ClassPlayer})
}
