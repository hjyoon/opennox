package opennox

import (
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/player"
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/server"
)

func TestPlayerAbilityRuntimeTickNative4FBEE0BindsFixedMatrixAndGlobalList(t *testing.T) {
	base := server.New(nil, nil, strman.New())
	t.Cleanup(base.Close)
	s := &Server{Server: base}
	s.abilities.s = s
	s.Abils.Init4FB990()

	warriorUnit := new(server.Object)
	warrior := s.Players.ResetInd(3)
	warrior.PlayerUnit = warriorUnit
	warrior.Info().SetPlayerClass(player.Warrior)
	s.Abils.SetPlayerAbilityCooldownAt(warrior.PlayerInd, server.AbilityHarpoon, 2)

	wizardUnit := new(server.Object)
	wizard := s.Players.ResetInd(4)
	wizard.PlayerUnit = wizardUnit
	wizard.Info().SetPlayerClass(player.Wizard)
	s.Abils.SetPlayerAbilityCooldownAt(wizard.PlayerInd, server.AbilityHarpoon, 2)

	dead := &server.ExecAbilityClass{
		Unit: &server.Object{ObjFlags: object.FlagDead},
		Abil: server.AbilityWarcry,
	}
	s.Abils.SetExecHead(dead)

	s.abilities.playerAbilityRuntimeTick4FBEE0()

	if got := s.Abils.PlayerAbilityCooldownAt(warrior.PlayerInd, server.AbilityHarpoon); got != 1 {
		t.Fatalf("Warrior cooldown = %d, want 1", got)
	}
	if got := s.Abils.PlayerAbilityCooldownAt(wizard.PlayerInd, server.AbilityHarpoon); got != 2 {
		t.Fatalf("non-Warrior cooldown = %d, want 2", got)
	}
	if got := s.Abils.ExecHead(); got != nil {
		t.Fatalf("execution-list head = %p, want nil", got)
	}
	if *dead != (server.ExecAbilityClass{}) {
		t.Fatalf("freed execution record = %+v, want zero", *dead)
	}
}
