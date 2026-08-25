package opennox

import (
	"net/netip"
	"testing"
	"time"

	"github.com/opennox/lobby"

	"github.com/opennox/opennox/v1/common/discover"
)

func noxWorldTestLocalize(id string) string { return "[" + id + "]" }

func TestNoxWorldServerRowsMatchLegacyColumns(t *testing.T) {
	servers := []discover.Server{
		{
			Game: lobby.Game{
				Name:    "abcdefghijklmnop",
				Address: "192.0.2.1",
				Port:    18590,
				Mode:    lobby.ModeArena,
				Access:  lobby.AccessOpen,
				Players: lobby.PlayersInfo{Cur: 1, Max: 4},
			},
			Ping: 42*time.Millisecond + 900*time.Microsecond,
		},
		{
			Game: lobby.Game{
				Port:    18590,
				Mode:    lobby.ModeQuest,
				Access:  lobby.AccessPassword,
				Players: lobby.PlayersInfo{Cur: 4, Max: 4},
				Quest:   &lobby.QuestInfo{Stage: 7},
			},
			IP: netip.MustParseAddr("192.0.2.2"),
		},
		{
			// GAME.EXE rejects duplicate address/port pairs before insertion.
			Game: lobby.Game{
				Name:    "duplicate",
				Address: "192.0.2.1",
				Port:    18590,
			},
		},
		{
			// GAME.EXE also rejects entries that have no usable address.
			Game: lobby.Game{Name: "missing-address", Port: 18590},
		},
		{
			Game: lobby.Game{
				Name:    "closed",
				Address: "192.0.2.3",
				Port:    18590,
				Mode:    lobby.ModeElimination,
				Access:  lobby.AccessClosed,
				Players: lobby.PlayersInfo{Cur: 0, Max: 16},
			},
			NoPing: true,
		},
	}

	rows := noxWorldServerRows(servers, noxWorldTestLocalize)
	if len(rows) != 3 {
		t.Fatalf("row count = %d, want 3", len(rows))
	}
	if got, want := rows[0].name, "abcdefghijklmno"; got != want {
		t.Fatalf("name = %q, want %q", got, want)
	}
	if got, want := rows[0].players, "1/4"; got != want {
		t.Fatalf("players = %q, want %q", got, want)
	}
	if got, want := rows[0].mode, "[Arena]"; got != want {
		t.Fatalf("mode = %q, want %q", got, want)
	}
	if got, want := rows[0].ping, "42"; got != want {
		t.Fatalf("ping = %q, want %q", got, want)
	}
	if got, want := rows[0].status, "[Open]"; got != want {
		t.Fatalf("status = %q, want %q", got, want)
	}
	if got, want := rows[1].name, "192.0.2.2:18590"; got != want {
		t.Fatalf("fallback name = %q, want %q", got, want)
	}
	if got, want := rows[1].mode, "[Quest] 7"; got != want {
		t.Fatalf("quest mode = %q, want %q", got, want)
	}
	if got, want := rows[1].ping, "--"; got != want {
		t.Fatalf("unknown ping = %q, want %q", got, want)
	}
	if got, want := rows[1].status, "[Noxworld.wnd:private]"; got != want {
		t.Fatalf("private status = %q, want %q", got, want)
	}
	if got, want := rows[2].mode, "[Highlander]"; got != want {
		t.Fatalf("elimination mode = %q, want %q", got, want)
	}
	if got, want := rows[2].status, "[Noxworld.wnd:closed]"; got != want {
		t.Fatalf("closed status = %q, want %q", got, want)
	}
}

func TestNoxWorldServerRowsPreserveUTF8AndCapResults(t *testing.T) {
	if got, want := noxWorldTruncateServerName("12345678901234가"), "12345678901234"; got != want {
		t.Fatalf("UTF-8 truncation = %q, want %q", got, want)
	}

	servers := make([]discover.Server, noxWorldMaxServers+2)
	for i := range servers {
		servers[i].Name = "server"
		servers[i].Address = netip.AddrFrom4([4]byte{10, byte(i >> 16), byte(i >> 8), byte(i)}).String()
		servers[i].Port = 18590
	}
	if got := len(noxWorldServerRows(servers, noxWorldTestLocalize)); got != noxWorldMaxServers {
		t.Fatalf("row cap = %d, want %d", got, noxWorldMaxServers)
	}
}

func TestNoxWorldSortButtonsMatchGAMEEXEToggles(t *testing.T) {
	for _, tc := range []struct {
		button int
		base   int
	}{
		{10047, 0},
		{10048, 2},
		{10049, 4},
		{10050, 6},
		{10051, 8},
	} {
		if got := noxWorldNextSorting(-1, tc.button); got != tc.base {
			t.Errorf("button %d first sort = %d, want %d", tc.button, got, tc.base)
		}
		if got := noxWorldNextSorting(tc.base, tc.button); got != tc.base+1 {
			t.Errorf("button %d toggle sort = %d, want %d", tc.button, got, tc.base+1)
		}
		if got := noxWorldNextSorting(tc.base+1, tc.button); got != tc.base {
			t.Errorf("button %d second toggle = %d, want %d", tc.button, got, tc.base)
		}
	}
}

func TestNoxWorldSortRowsMatchLegacyKeys(t *testing.T) {
	base := []noxWorldServerRow{
		{name: "charlie", mode: "Arena", pingValue: 99, statusValue: 0x20, server: discover.Server{Game: lobby.Game{Players: lobby.PlayersInfo{Cur: 1}}}},
		{name: "Alpha", mode: "Quest", pingValue: 9999, statusValue: 0, server: discover.Server{Game: lobby.Game{Players: lobby.PlayersInfo{Cur: 3}}}},
		{name: "bravo", mode: "CTF", pingValue: 10, statusValue: 0x10, server: discover.Server{Game: lobby.Game{Players: lobby.PlayersInfo{Cur: 2}}}},
	}
	tests := []struct {
		sorting int
		want    []string
	}{
		{0, []string{"Alpha", "bravo", "charlie"}},
		{1, []string{"charlie", "bravo", "Alpha"}},
		{2, []string{"charlie", "bravo", "Alpha"}},
		{3, []string{"Alpha", "bravo", "charlie"}},
		{4, []string{"charlie", "bravo", "Alpha"}},
		{5, []string{"Alpha", "bravo", "charlie"}},
		{6, []string{"bravo", "charlie", "Alpha"}},
		{7, []string{"Alpha", "charlie", "bravo"}},
		{8, []string{"Alpha", "bravo", "charlie"}},
		{9, []string{"charlie", "bravo", "Alpha"}},
	}
	for _, tc := range tests {
		rows := append([]noxWorldServerRow(nil), base...)
		noxWorldSortRows(rows, tc.sorting)
		for i, want := range tc.want {
			if got := rows[i].name; got != want {
				t.Errorf("sorting %d row %d = %q, want %q", tc.sorting, i, got, want)
			}
		}
	}
}

func TestNoxWorldSortRowsReverseEqualKeysLikeLegacyInsertion(t *testing.T) {
	rows := []noxWorldServerRow{
		{name: "same", server: discover.Server{Game: lobby.Game{Address: "first"}}},
		{name: "SAME", server: discover.Server{Game: lobby.Game{Address: "second"}}},
		{name: "Same", server: discover.Server{Game: lobby.Game{Address: "third"}}},
	}
	noxWorldSortRows(rows, 0)
	for i, want := range []string{"third", "second", "first"} {
		if got := rows[i].server.Address; got != want {
			t.Fatalf("first sort row %d = %q, want %q", i, got, want)
		}
	}
	noxWorldSortRows(rows, 1)
	for i, want := range []string{"first", "second", "third"} {
		if got := rows[i].server.Address; got != want {
			t.Fatalf("second sort row %d = %q, want %q", i, got, want)
		}
	}
}

func TestNoxWorldGameModeStringsMatchLegacyPriority(t *testing.T) {
	for mode, want := range map[lobby.GameMode]string{
		lobby.ModeQuest:       "[Quest]",
		lobby.ModeCTF:         "[CTF]",
		lobby.ModeElimination: "[Highlander]",
		lobby.ModeKOTR:        "[KotR]",
		lobby.ModeFlagBall:    "[Flagball]",
		lobby.ModeChat:        "[Chat]",
		lobby.ModeArena:       "[Arena]",
		lobby.ModeCoop:        "[Arena]",
		lobby.ModeCustom:      "[Arena]",
	} {
		if got := noxWorldGameModeString(mode, noxWorldTestLocalize); got != want {
			t.Errorf("mode %q = %q, want %q", mode, got, want)
		}
	}
}
