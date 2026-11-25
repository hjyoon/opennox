package legacy

import (
	"bytes"
	"testing"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestClientRankNativeOrderAndScoreRestoration(t *testing.T) {
	world := objectStatusTestServer(t, 1)
	flags, engine := noxflags.GetGame(), noxflags.GetEngine()
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameHost | noxflags.GameModeCoopTeam)
	noxflags.ResetEngine()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(flags)
		noxflags.ResetEngine()
		noxflags.SetEngine(engine)
	})
	width := memmap.PtrInt32(0x5D4594, 1084036)
	oldWidth := *width
	*width = 0
	t.Cleanup(func() { *width = oldWidth })
	for _, region := range [][]byte{
		memmap.Slice(0x5D4594, 1084132)[:4*80],
		memmap.Slice(0x5D4594, 1087204)[:2*56],
		memmap.Slice(0x5D4594, 1090116)[:3],
	} {
		old := bytes.Clone(region)
		t.Cleanup(func() { copy(region, old) })
	}
	if unsafe.Sizeof(clientRankPlayerRow{}) != 80 || unsafe.Offsetof(clientRankPlayerRow{}.Status) != 76 ||
		unsafe.Sizeof(clientRankTeamRow{}) != 56 || unsafe.Offsetof(clientRankTeamRow{}.ID) != 44 {
		t.Fatal("rank rows differ from the stock widget layout")
	}
	input := []struct {
		lessons int32
		deaths  uint32
		ladder  uint32
		status  uint32
	}{
		{5, 3, 0, 0},
		{5, 2, 7, 0},
		{30, 0, 1, 0x01},
		{-2, 1, 5, 0x21},
	}
	var players []*server.Player
	for i, entry := range input {
		player := world.Players.NewRaw(101 + i)
		player.Lessons, player.Field2140, player.Field2108, player.Field3680 = entry.lessons, entry.deaths, entry.ladder, entry.status
		player.SetName("Håna")
		players = append(players, player)
	}
	alpha, beta := world.Teams.Create(5), world.Teams.Create(7)
	alpha.Lessons, beta.Lessons = 10, -2
	alpha.SetNameAnd68("Alpha", 0)
	beta.SetNameAnd68("Beta", 0)

	cases := []struct {
		name    string
		mode    clientRankMode
		order   [4]byte
		display [4]int32
		playing byte
	}{
		{"lessons", clientRankLessons, [4]byte{0, 1, 3, 2}, [4]int32{5, 5, -2, 30}, 3},
		{"elimination", clientRankElimination, [4]byte{3, 1, 0, 2}, [4]int32{1, 2, 3, 0}, 3},
		{"ladder", clientRankLadder, [4]byte{3, 1, 0, 2}, [4]int32{5, 7, 0, 1}, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			*memmap.PtrUint8(0x5D4594, 1090117) = 0
			*memmap.PtrUint8(0x5D4594, 1090118) = 0
			clientRankStorePlayers(tc.mode, clientRankPlayerCount())
			if rows, playing := memmap.Uint8(0x5D4594, 1090117), memmap.Uint8(0x5D4594, 1090118); rows != 4 || playing != tc.playing {
				t.Fatalf("row count %d, playing %d, want 4 and %d", rows, playing, tc.playing)
			}
			for i, who := range tc.order {
				row := clientRankPlayerAt(byte(i))
				if row.NetCode != players[who].NetCodeVal || int32(row.Score) != tc.display[i] || row.TeamID != -1 ||
					row.Status != input[who].status || alloc.GoString16S(row.Name[:]) != "Håna" {
					t.Fatalf("row %d = %+v, want player %d and score %d", i, *row, players[who].NetCodeVal, tc.display[i])
				}
			}
			for i, player := range players {
				entry := input[i]
				if player.Lessons != entry.lessons || player.Field2140 != entry.deaths || player.Field2108 != entry.ladder {
					t.Fatalf("player %d scores changed to %d/%d/%d", i, player.Lessons, player.Field2140, player.Field2108)
				}
			}
		})
	}
	for _, ascending := range []bool{false, true} {
		clientRankStoreTeams(ascending)
		first, second := alpha, beta
		if ascending {
			first, second = beta, alpha
		}
		if count := memmap.Uint8(0x5D4594, 1090116); count != 2 {
			t.Fatalf("team rows = %d", count)
		}
		for i, team := range []*server.Team{first, second} {
			row := clientRankTeamAt(byte(i))
			if row.ID != uint32(team.ID()) || row.Score != team.Lessons ||
				row.Color != byte(team.ColorInd) || alloc.GoString16S(row.Name[:]) != team.Name() {
				t.Fatalf("team row %d = %+v, want %s and score %d", i, *row, team.Name(), team.Lessons)
			}
		}
	}
}
