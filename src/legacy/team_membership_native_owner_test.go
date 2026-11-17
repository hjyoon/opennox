package legacy

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/team_membership_owner_native_v1.json.gz
var teamMembershipNativeOwnerOriginal []byte

type teamMembershipOwnerInput struct {
	Name    string
	Graph   int
	Lists   [2][]int
	Program string
	Seed    byte
}

type teamMembershipOwnerEvent struct {
	Operation       string
	ReturnedTeam    int    `json:"returned_team"`
	ActiveTeams     int    `json:"active_teams"`
	TeamArrayLength int    `json:"team_array_length"`
	TeamSHA32       string `json:"team_sha256_32"`
	TeamSHA64       string `json:"team_sha256_64"`
	NodeSHA         string `json:"node_owner_sha256"`
	IDs             [4]byte
	RawCounts       []uint32 `json:"raw_counts"`
	MemberCounts    []int    `json:"member_counts"`
	OrderedLists    [2][]int `json:"ordered_lists"`
}

type teamMembershipOwnerCase struct {
	Input  teamMembershipOwnerInput
	Events []teamMembershipOwnerEvent
}

func teamMembershipOwnerHash(data []byte) string {
	s := sha256.Sum256(data)
	return hex.EncodeToString(s[:])
}

func teamMembershipRunOwnerCase(t *testing.T, c teamMembershipOwnerCase, parentRoute int) {
	t.Helper()
	buffer, free := alloc.MakeHigh([]byte{}, 65536)
	defer free()
	srv := server.NewHigh(nil, nil, strman.New())
	closed := false
	defer func() {
		if !closed {
			srv.Close()
		}
	}()
	oldServer := GetServer
	GetServer = func() Server { return &teamMembershipTestServer{native: srv} }
	defer func() { GetServer = oldServer }()
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(srv.Teams.Arr[0].C()) <= math.MaxUint32 || uintptr(unsafe.Pointer(&buffer[0])) <= math.MaxUint32) {
		t.Fatal("owned allocations must exceed4GiB", c.Input.Name)
	}
	for i := range buffer {
		buffer[i] = byte(int(c.Input.Seed) + 17*i + 3*(i>>8))
	}
	var nodes [4]*server.ObjectTeam
	for i := range nodes {
		nodes[i] = (*server.ObjectTeam)(unsafe.Pointer(&buffer[0x2000+0x100*i]))
		nodes[i].Field0 = 0
		nodes[i].ID = 0
	}
	team := func(role int) *server.Team {
		if role == 0 {
			return nil
		}
		return &srv.Teams.Arr[role]
	}
	teamMembershipNativeSeedC(srv.Teams.Arr[0].C())
	srv.Teams.ActiveCnt = 2
	teamPointers := [2]unsafe.Pointer{team(1).C(), team(2).C()}
	for i, list := range c.Input.Lists {
		for j := len(list) - 1; j >= 0; j-- {
			srv.Teams.AttachObject(nodes[list[j]], server.TeamID(i+1))
		}
	}
	index := 0
	emit := func(op string, returned *server.Team) {
		t.Helper()
		if index >= len(c.Events) {
			t.Fatal("extra owner event", c.Input.Name, op)
		}
		e := c.Events[index]
		index++
		returnedID := 0
		if returned != nil {
			returnedID = int(returned.ID())
		}
		if e.Operation != op || e.ReturnedTeam != returnedID || e.ActiveTeams != srv.Teams.ActiveCnt || e.TeamArrayLength != len(srv.Teams.Arr) {
			t.Fatal("existing owner control/return differs", c.Input.Name, op, e, returnedID)
		}
		var IDs [4]byte
		for i, n := range nodes {
			IDs[i] = byte(n.ID)
			if n.Field0 != 0 {
				t.Fatal("native pointer entered fixed PE32 slot", c.Input.Name, op, i)
			}
		}
		if IDs != e.IDs || teamMembershipOwnerHash(buffer) != e.NodeSHA {
			t.Fatal("complete incoming node owner differs", c.Input.Name, op, IDs, e.IDs)
		}
		wantTeamHash := e.TeamSHA64
		if unsafe.Sizeof(uintptr(0)) == 4 {
			wantTeamHash = e.TeamSHA32
		}
		checkTeams := func() {
			if len(srv.Teams.Arr) == 0 {
				return
			}
			data := unsafe.Slice((*byte)(srv.Teams.Arr[0].C()), len(srv.Teams.Arr)*int(unsafe.Sizeof(server.Team{})))
			if teamMembershipOwnerHash(data) != wantTeamHash {
				t.Fatal("complete incoming Team array differs", c.Input.Name, op)
			}
		}
		checkTeams()
		if !closed {
			raw := []uint32{teamMembershipNativeCountC(team(1).C()), teamMembershipNativeCountC(team(2).C())}
			counts := []int{srv.Teams.MemberCount(team(1)), srv.Teams.MemberCount(team(2))}
			if !reflect.DeepEqual(raw, e.RawCounts) || !reflect.DeepEqual(counts, e.MemberCounts) {
				t.Fatal("raw/deferred/legacy count boundary differs", c.Input.Name, op, raw, counts, e.RawCounts, e.MemberCounts)
			}
		}
		var next [4]int
		for _, list := range e.OrderedLists {
			for i := 0; i+1 < len(list); i++ {
				next[list[i]] = list[i+1] + 1
			}
		}
		for route := 0; route < 3; route++ {
			for i, list := range e.OrderedLists {
				var wanted uintptr
				if len(list) > 0 {
					wanted = uintptr(unsafe.Pointer(nodes[list[0]]))
				}
				if got := teamMembershipNativeQuery(route, 0, teamPointers[i]); got != wanted {
					t.Fatal("original head order differs", c.Input.Name, op, route, i, got, wanted)
				}
			}
			for i, n := range nodes {
				var wanted uintptr
				if next[i] != 0 {
					wanted = uintptr(unsafe.Pointer(nodes[next[i]-1]))
				}
				if got := teamMembershipNativeQuery(route, 1, unsafe.Pointer(n)); got != wanted {
					t.Fatal("original next order differs", c.Input.Name, op, route, i, got, wanted)
				}
			}
		}
		// After Close, old Team addresses are opaque keys only. Never read freed
		// Team bytes; cleared native ownership must return no links.
		checkTeams()
		if teamMembershipOwnerHash(buffer) != e.NodeSHA {
			t.Fatal("query wrote the owner", c.Input.Name, op)
		}
	}
	link := func(n, role int) {
		var v *server.ObjectTeam
		if n >= 0 {
			v = nodes[n]
		}
		var returned *server.Team
		if parentRoute == 0 {
			returned = srv.Teams.LinkObject(v, server.TeamID(role))
		} else if teamMembershipNativeParentLinkC(unsafe.Pointer(v), byte(role)) {
			returned = team(role)
		}
		emit(fmt.Sprintf("link-%d-%d", n, role), returned)
	}
	commit := func(role int) { srv.Teams.CommitObjectAttach(team(role)); emit(fmt.Sprintf("commit-%d", role), nil) }
	detach := func(n int) {
		var v *server.ObjectTeam
		if n >= 0 {
			v = nodes[n]
		}
		var returned *server.Team
		if parentRoute == 0 {
			returned = srv.Teams.DetachObject(v)
		} else if address := teamMembershipNativeParentDetachC(unsafe.Pointer(v)); address != 0 {
			for role := 1; role <= 2; role++ {
				if address == uintptr(team(role).C()) {
					returned = team(role)
					break
				}
			}
			if returned == nil {
				t.Fatal("C detach return lost native owner address", c.Input.Name, address)
			}
		}
		emit(fmt.Sprintf("detach-%d", n), returned)
	}
	all := func(role int) { srv.Teams.DetachAllObjects(team(role)); emit(fmt.Sprintf("detach-all-%d", role), nil) }
	reset := func() { srv.Teams.Reset(); emit("reset", nil) }
	emit("initial", nil)
	switch c.Input.Program {
	case "link-commit-detach":
		for n := 0; n < 4; n++ {
			link(n, 1)
			commit(1)
		}
		for n := 0; n < 4; n++ {
			detach(n)
		}
		all(1)
		all(2)
		reset()
	case "move-deferred-commit":
		link(0, 1)
		link(0, 1)
		commit(1)
		link(0, 2)
		commit(2)
		link(-1, 1)
		commit(0)
		detach(0)
		detach(-1)
		all(1)
		all(2)
		reset()
	case "detach-ID-fallback":
		detach(0)
		nodes[0].ID = 1
		teamMembershipNativeSetCountC(team(1).C(), 7)
		emit("seed-legacy-ID-with-set", nil)
		detach(0)
		for _, n := range []int{2, 0, 3, 1} {
			detach(n)
		}
		all(1)
		nodes[0].ID = 1
		teamMembershipNativeSetCountC(team(1).C(), 3)
		emit("seed-legacy-ID-without-set", nil)
		detach(0)
		all(1)
		all(2)
		reset()
	case "reset-reuse":
		reset()
		teamMembershipNativeSeedC(srv.Teams.Arr[0].C())
		srv.Teams.ActiveCnt = 2
		emit("reuse-team-slots", nil)
		link(3, 2)
		commit(2)
		link(0, 1)
		commit(1)
		link(1, 2)
		commit(2)
		all(2)
	default:
		t.Fatal("undeclared owner program", c.Input.Program)
	}
	srv.Close()
	closed = true
	emit("closed-no-Team-dereference", nil)
	if index != len(c.Events) {
		t.Fatal("missing owner events", c.Input.Name, index, len(c.Events))
	}
}

func TestTeamMembershipNativeOwnerCompatibility(t *testing.T) {
	var cases []teamMembershipOwnerCase
	npcVoiceOriginalDecode(t, teamMembershipNativeOwnerOriginal, "d62bd0b52667a1cbc0280608eaa0b2c0ea0cffb8d9c1d4041d77f61b73f38a71", "cb456c9fbcebfbf81dc38be28c9c2398b84902e2186a157356a6413a534a37e1", 1044, &cases)
	seen, events := 0, 0
	for parentRoute := 0; parentRoute < 2; parentRoute++ {
		t.Run(fmt.Sprintf("parents%d", parentRoute), func(t *testing.T) {
			for _, program := range []string{"link-commit-detach", "move-deferred-commit", "detach-ID-fallback", "reset-reuse"} {
				t.Run(program, func(t *testing.T) {
					count := 0
					for _, c := range cases {
						if c.Input.Program != program {
							continue
						}
						teamMembershipRunOwnerCase(t, c, parentRoute)
						count++
						seen++
						events += len(c.Events)
					}
					if count != 261 {
						t.Fatal("owner graph coverage", parentRoute, program, count)
					}
				})
			}
		})
	}
	if seen != 2088 || events != 29754 {
		t.Fatal("complete incoming owner baseline", seen, events)
	}
}
