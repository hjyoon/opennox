package opennox

import (
	"image"
	"net/netip"
	"testing"
	"time"
	"unsafe"

	"github.com/opennox/lobby"

	"github.com/opennox/opennox/v1/common/discover"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
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

func TestNoxWorldServerEndpointRequiresLegacyIPv4AndValidPort(t *testing.T) {
	tests := []struct {
		name    string
		server  discover.Server
		want    string
		wantErr bool
	}{
		{
			name: "discovered IP wins",
			server: discover.Server{
				Game: lobby.Game{Address: "198.51.100.1", Port: 18590},
				IP:   netip.MustParseAddr("192.0.2.10"),
			},
			want: "192.0.2.10:18590",
		},
		{
			name:   "address and embedded port",
			server: discover.Server{Game: lobby.Game{Address: "192.0.2.20:19000"}},
			want:   "192.0.2.20:19000",
		},
		{
			name:    "IPv6 unsupported by GAME.EXE record",
			server:  discover.Server{Game: lobby.Game{Address: "2001:db8::1", Port: 18590}},
			wantErr: true,
		},
		{
			name:    "invalid port",
			server:  discover.Server{Game: lobby.Game{Address: "192.0.2.30", Port: 65536}},
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := noxWorldServerEndpoint(tc.server)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("endpoint = %s, want error", got)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got.String() != tc.want {
				t.Fatalf("endpoint = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestNoxWorldSelectedServerPreservesLegacyRecordOn64Bit(t *testing.T) {
	oldHost := clientGetServerHost()
	defer func() {
		noxWorldClearSelectedServer()
		clientSetServerHost(oldHost)
	}()
	srv := discover.Server{
		Game: lobby.Game{
			Name:    "0123456789abcdef",
			Address: "192.0.2.44",
			Port:    19044,
			Map:     "estate",
			Mode:    lobby.ModeQuest,
			Access:  lobby.AccessPassword,
			Players: lobby.PlayersInfo{Cur: 3, Max: 6},
			Quest:   &lobby.QuestInfo{Stage: 12},
		},
		Ping: 37 * time.Millisecond,
	}
	if err := noxWorldSetSelectedServer(srv); err != nil {
		t.Fatal(err)
	}
	if noxWorldSelected == nil {
		t.Fatal("selected record is nil")
	}
	if got, want := legacy.Get_dword_5d4594_814624(), unsafe.Pointer(noxWorldSelected); got != want {
		t.Fatalf("legacy record pointer = %p, want %p", got, want)
	}
	if got, want := legacy.Nox_client_getServerAddr_43B300(), netip.MustParseAddr("192.0.2.44"); got != want {
		t.Fatalf("legacy address = %s, want %s", got, want)
	}
	if got, want := legacy.ClientGetServerPort(), 19044; got != want {
		t.Fatalf("legacy port = %d, want %d", got, want)
	}
	if got, want := noxWorldSelected.ServerName(), "0123456789abcde"; got != want {
		t.Fatalf("server name = %q, want %q", got, want)
	}
	if got, want := noxWorldSelected.Flags(), noxflags.GameModeQuest; got != want {
		t.Fatalf("flags = %v, want %v", got, want)
	}
	if got, want := noxWorldSelected.QuestLevel(), 12; got != want {
		t.Fatalf("quest level = %d, want %d", got, want)
	}
	if got, want := noxWorldSelected.Ping(), 37; got != want {
		t.Fatalf("ping = %d, want %d", got, want)
	}
	if got, want := clientGetServerHost(), "192.0.2.44"; got != want {
		t.Fatalf("client host = %q, want %q", got, want)
	}
}

func TestNoxWorldJoinSelectedRequiresSelectionAndUsesExactServer(t *testing.T) {
	srv := discover.Server{Game: lobby.Game{Name: "join-me", Address: "192.0.2.55", Port: 18590}}
	st := &noxWorldNativeState{
		rows:     []noxWorldServerRow{{server: srv}},
		selected: -1,
	}
	if err := st.joinSelected(); err == nil {
		t.Fatal("join without a selection succeeded")
	}
	var got discover.Server
	st.selected = 0
	st.onJoin = func(s discover.Server) error {
		got = s
		return nil
	}
	if err := st.joinSelected(); err != nil {
		t.Fatal(err)
	}
	if got.Name != srv.Name || got.Address != srv.Address || got.Port != srv.Port {
		t.Fatalf("join callback server = %+v, want %+v", got, srv)
	}
}

func TestNoxWorldServerInfoPopupMatchesLegacyPlacementAndCoreFields(t *testing.T) {
	for _, tc := range []struct {
		in, want image.Point
	}{
		{image.Pt(0, 0), image.Pt(216, 27)},
		{image.Pt(400, 200), image.Pt(335, 180)},
		{image.Pt(640, 480), image.Pt(470, 331)},
	} {
		if got := noxWorldInfoPosition(tc.in); got != tc.want {
			t.Errorf("popup position for %v = %v, want %v", tc.in, got, tc.want)
		}
	}
	row := noxWorldServerRow{
		name:    "quest-server",
		players: "2/4",
		mode:    "[Quest] 9",
		ping:    "31",
		server: discover.Server{Game: lobby.Game{
			Map:   "war01a",
			Mode:  lobby.ModeQuest,
			Quest: &lobby.QuestInfo{Stage: 9},
		}},
	}
	got := noxWorldServerInfoLines(row, noxWorldTestLocalize)
	want := []string{
		"[Name]", "quest-server", "",
		"[Ping]", "31", "",
		"[GameType]", "[Quest] 9", "[Stage]", "9",
		"", "[Map]", "war01a", "", "[Occupancy]", "2/4",
	}
	if len(got) != len(want) {
		t.Fatalf("info line count = %d, want %d: %q", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("info line %d = %q, want %q", i, got[i], want[i])
		}
	}
}
