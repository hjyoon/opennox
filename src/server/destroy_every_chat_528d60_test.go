package server

import (
	"bytes"
	"testing"

	"github.com/opennox/libs/common"

	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/netlist"
)

func TestDestroyEveryChat528D60SendsExactSentinelToPlayerUnits(t *testing.T) {
	s := new(Server)
	s.Players.list = make([]Player, common.MaxPlayers)
	s.NetList = netlist.New()
	s.NetList.Init()
	t.Cleanup(s.NetList.Free)

	s.Players.list[0].Active = 1
	s.Players.list[0].PlayerInd = 0
	s.Players.list[0].PlayerUnit = &Object{}
	s.Players.list[3].Active = 1
	s.Players.list[3].PlayerInd = 3
	s.Players.list[3].PlayerUnit = &Object{}
	s.Players.list[5].Active = 1
	s.Players.list[5].PlayerInd = 5

	s.DestroyEveryChat528D60()

	want := []byte{0xCA, 0xAD, 0xDE}
	for _, ind := range []ntype.PlayerInd{0, 3} {
		if got := s.NetList.CopyPacketsA(ind, netlist.Kind1); !bytes.Equal(got, want) {
			t.Fatalf("player %d packet = % x, want % x", ind, got, want)
		}
	}
	if got := s.NetList.CopyPacketsA(5, netlist.Kind1); len(got) != 0 {
		t.Fatalf("active player without a unit received % x", got)
	}
}
