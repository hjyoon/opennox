package opennox

import (
	"encoding/json"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/player"
	"github.com/opennox/opennox/v1/server"
)

func TestPlayerTargetStatInputsAreCopied(t *testing.T) {
	pl := &server.Player{Level: 3}
	pl.Info().SetPlayerClass(player.Wizard)
	ud := &server.PlayerUpdateData{Player: pl}
	unit := &server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(ud)}
	world := new(server.Server)
	world.Players.Stats.Base = server.ClassStats{Health: 20, Mana: 78, Speed: 17250, Strength: 20}
	world.Players.Stats.Wizard = server.ClassStats{Health: 75, Mana: 170, Speed: 18500, Strength: 35}
	world.Players.Stats.Warrior = server.ClassStats{Health: 150, Mana: 0, Speed: 21000, Strength: 125}
	world.Players.Mult.Wizard = server.ClassStats{Health: 1, Mana: 1, Speed: 1, Strength: 1}
	got, err := e2eReadPlayerTarget(1, unit, nil, world)
	if err != nil {
		t.Fatal(err)
	}
	if !got.StatInputs.Present || got.StatInputs.Level != 3 || got.StatInputs.Class != uint8(player.Wizard) ||
		got.StatInputs.Base[1] != math.Float32bits(78) || got.StatInputs.Maximum[1] != math.Float32bits(170) || got.StatInputs.Warrior[0] != math.Float32bits(150) || got.StatInputs.Multipliers[0] != math.Float32bits(1) {
		t.Fatalf("wrong native stat inputs %+v", got.StatInputs)
	}
	pl.Level = 9
	world.Players.Stats.Wizard.Mana = 0
	world.Players.Stats.Base.Mana = 0
	if got.StatInputs.Level != 3 || got.StatInputs.Base[1] != math.Float32bits(78) || got.StatInputs.Maximum[1] != math.Float32bits(170) {
		t.Fatal("later native updates changed the copied inputs")
	}
	if _, err := json.Marshal(got); err != nil {
		t.Fatal(err)
	}
	absent, err := e2eReadPlayerTarget(2, unit, nil)
	if err != nil || absent.StatInputs.Present {
		t.Fatal("no world must not fabricate stat inputs")
	}
}
