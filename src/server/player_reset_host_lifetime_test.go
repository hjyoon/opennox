package server

import (
	"reflect"
	"testing"
)

func TestPlayerResetAllInvalidatesHostLifetime(t *testing.T) {
	players := &serverPlayers{list: make([]Player, 3)}
	players.Stats.Base.Health = 27
	players.Mult.Warrior.Health = 2
	hostUnit := &Object{NetCode: 0x1234}
	peerUnit := &Object{NetCode: 0x5678}
	host := &players.list[2]
	*host = Player{Active: 1, PlayerInd: 2, PlayerUnit: hostUnit}
	players.list[0] = Player{Active: 1, PlayerUnit: peerUnit}
	players.SetHost(host, hostUnit)
	if players.Host() != host || players.HostUnit() != hostUnit || players.Count() != 2 {
		t.Fatal("session fixture did not bind host and peer")
	}

	players.ResetAll()
	if players.Host() != nil || players.HostUnit() != nil || players.IsHost(host) {
		t.Errorf("session reset retained host references: player=%p unit=%p isHost=%t", players.Host(), players.HostUnit(), players.IsHost(host))
	}
	if players.First() != nil || players.Count() != 0 || players.HasUnits() {
		t.Error("session reset retained active players or units")
	}
	for i, slot := range players.list {
		if !reflect.DeepEqual(slot, Player{}) {
			t.Errorf("player slot %d retained session state", i)
		}
	}
	if players.Stats.Base.Health != 27 || players.Mult.Warrior.Health != 2 || hostUnit.NetCode != 0x1234 || peerUnit.NetCode != 0x5678 {
		t.Error("reset changed persistent class settings or object storage")
	}

	// A fresh session can reuse a cleared slot without reviving the old host.
	replacement := &Object{NetCode: 0x9abc}
	rebound := &players.list[1]
	*rebound = Player{Active: 1, PlayerInd: 1, PlayerUnit: replacement}
	players.SetHost(rebound, replacement)
	if players.Host() != rebound || players.HostUnit() != replacement || players.IsHost(host) || players.Count() != 1 {
		t.Error("new session did not establish an independent host binding")
	}
	players.ResetAll()
	if players.Host() != nil || players.HostUnit() != nil || players.IsHost(rebound) {
		t.Error("second session reset retained its host binding")
	}

	// Some controlled server callers bind a unit before a player slot exists.
	players.SetHost(nil, hostUnit)
	players.ResetAll()
	if players.Host() != nil || players.HostUnit() != nil {
		t.Error("reset retained a unit-only host binding")
	}
}
