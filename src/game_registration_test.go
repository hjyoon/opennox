package opennox

import (
	"slices"
	"testing"

	"github.com/opennox/lobby"
	"github.com/opennox/xwis"
)

func TestOnlineGameRegistrationPreservesAccess(t *testing.T) {
	for _, tc := range []struct {
		access lobby.GameAccess
		want   xwis.Access
	}{
		{lobby.AccessOpen, xwis.AccessOpen},
		{lobby.AccessPassword, xwis.AccessPrivate},
		{lobby.AccessClosed, xwis.AccessClosed},
	} {
		info := xwisInfoShort{Access: tc.access}.GameInfo()
		if info.Game.Access != tc.access || info.XWIS.Access != tc.want {
			t.Errorf("access %q registered as lobby %q and XWIS %d", tc.access, info.Game.Access, info.XWIS.Access)
		}
	}
}

func TestOnlineGameRegistrationDetectsRosterChanges(t *testing.T) {
	initial := xwisInfoShort{Players: []lobby.PlayerInfo{
		{Name: "A", Class: "Warrior"},
		{Name: "B", Class: "Wizard"},
	}}
	for _, updatedPlayer := range []lobby.PlayerInfo{
		{Name: "C", Class: "Warrior"},
		{Name: "A", Class: "Conjurer"},
	} {
		updated := initial
		updated.Players = slices.Clone(initial.Players)
		if !initial.Equal(&updated) {
			t.Fatal("identical independent rosters compare differently")
		}
		updated.Players[0] = updatedPlayer
		if initial.Equal(&updated) {
			t.Errorf("same-size roster change to %+v was ignored", updatedPlayer)
		}
	}
}
