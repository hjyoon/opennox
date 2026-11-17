package legacy

import (
	"bytes"
	_ "embed"
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/team_membership_queries_original_v1.json.gz
var teamMembershipOriginalQueries []byte

type teamMembershipTestServer struct {
	Server
	native *server.Server
}

func (s *teamMembershipTestServer) S() *server.Server { return s.native }

type teamMembershipQueryCase struct {
	Input struct {
		Name         string
		Entry        string
		ArgumentRole int `json:"argument_role"`
		Lists        [2][]int
		Seed         byte
		IDs          [4]byte
	}
	ReturnRole int `json:"return_node_role"`
}

func TestTeamMembershipNativeLayout(t *testing.T) {
	var team server.Team
	var node server.ObjectTeam
	word := unsafe.Sizeof(uintptr(0))
	want := [13]uintptr{word, unsafe.Sizeof(team), unsafe.Alignof(team), 44, 4, unsafe.Sizeof(node), unsafe.Alignof(node), unsafe.Offsetof(node.Field0), unsafe.Sizeof(node.Field0), unsafe.Offsetof(node.ID), unsafe.Sizeof(node.ID), word, word}
	if got := teamMembershipNativeLayoutC(); got != want || got[5] != 8 || got[9] != 4 {
		t.Fatal("native public function types and fixed record", got, want)
	}
}

func TestTeamMembershipNativeOriginalQueries(t *testing.T) {
	var cases []teamMembershipQueryCase
	npcVoiceOriginalDecode(t, teamMembershipOriginalQueries, "ca0158f2c2b279e17312902fd55c008be1e4178c667bcd169099b420352c801d", "8c835149bc19cda73729450624a4f99d503f43bac107e5e4a2d29dc211ba6e57", 2088, &cases)
	const ownerBytes, guard = 65536, 32
	for route := 0; route < 3; route++ {
		for _, place := range []int{0, 512, 4096} {
			t.Run(fmt.Sprintf("route%d/place%d", route, place), func(t *testing.T) {
				teams, freeTeams := alloc.MakeHigh([]byte{}, ownerBytes+2*guard)
				defer freeTeams()
				buffer, freeNodes := alloc.MakeHigh([]byte{}, ownerBytes+2*guard)
				defer freeNodes()
				srv := server.NewHigh(nil, nil, strman.New())
				defer srv.Close()
				oldServer := GetServer
				GetServer = func() Server { return &teamMembershipTestServer{native: srv} }
				defer func() { GetServer = oldServer }()
				srv.Teams.Arr = unsafe.Slice((*server.Team)(unsafe.Pointer(&teams[guard+0x1000+place])), 17)
				var nodes [4]*server.ObjectTeam
				for i := range nodes {
					nodes[i] = (*server.ObjectTeam)(unsafe.Pointer(&buffer[guard+0x2000+place+0x100*i]))
				}
				if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(srv.Teams.Arr[0].C()) <= math.MaxUint32 || uintptr(unsafe.Pointer(nodes[0])) <= math.MaxUint32) {
					t.Fatal("both native owners must exceed4GiB")
				}
				// Reset sees valid null pointer fields from the first iteration too.
				teamMembershipNativeSeedC(srv.Teams.Arr[0].C())
				for _, c := range cases {
					srv.Teams.Reset()
					for i := range teams {
						teams[i] = 0xd7
						buffer[i] = 0xe9
					}
					for i := 0; i < ownerBytes; i++ {
						teams[guard+i] = byte(int(c.Input.Seed) + 17*i + 3*(i>>8))
						buffer[guard+i] = byte(int(c.Input.Seed) + 29*i + 5*(i>>8))
					}
					teamMembershipNativeSeedC(srv.Teams.Arr[0].C())
					srv.Teams.ActiveCnt = 2
					for _, n := range nodes {
						n.Field0 = 0
						n.ID = 0
					}
					for i, list := range c.Input.Lists {
						for j := len(list) - 1; j >= 0; j-- {
							srv.Teams.AttachObject(nodes[list[j]], server.TeamID(i+1))
						}
					}
					for i, n := range nodes {
						n.ID = server.TeamID(c.Input.IDs[i])
					}
					beforeTeams, beforeNodes := bytes.Clone(teams), bytes.Clone(buffer)
					var argument unsafe.Pointer
					kind := 0
					if c.Input.Entry == "head" {
						if c.Input.ArgumentRole != 0 {
							argument = srv.Teams.Arr[c.Input.ArgumentRole].C()
						}
					} else if c.Input.Entry == "next" {
						kind = 1
						if c.Input.ArgumentRole != 0 {
							argument = unsafe.Pointer(nodes[c.Input.ArgumentRole-1])
						}
					} else {
						t.Fatal("undeclared entry", c.Input.Entry)
					}
					var wanted uintptr
					if c.ReturnRole != 0 {
						wanted = uintptr(unsafe.Pointer(nodes[c.ReturnRole-1]))
					}
					if got := teamMembershipNativeQuery(route, kind, argument); got != wanted {
						t.Fatal("original full pointer role differs", c.Input.Name, got, wanted)
					}
					if !bytes.Equal(teams, beforeTeams) || !bytes.Equal(buffer, beforeNodes) {
						t.Fatal("query changed complete foreign owner or guards", c.Input.Name)
					}
				}
			})
		}
	}
}
