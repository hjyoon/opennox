package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/client"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type matchWinnerTestClient struct {
	Client
	core client.Client
}

func (c *matchWinnerTestClient) Cli() *client.Client { return &c.core }

func TestMatchWinnerScoresPreserveHostAndClientMembership(t *testing.T) {
	oldFlags := noxflags.GetGame()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})
	for _, tc := range []struct {
		name  string
		flags noxflags.GameFlag
	}{
		{"host arena", noxflags.GameHost | noxflags.GameModeArena},
		{"client arena", noxflags.GameModeArena},
		{"host elimination", noxflags.GameHost | noxflags.GameModeElimination},
		{"client elimination", noxflags.GameModeElimination},
	} {
		t.Run(tc.name, func(t *testing.T) {
			const limit = 12
			noxflags.ResetGame()
			noxflags.SetGame(tc.flags)
			settings := (*server.Settings2)(cliGamedataNative416590(0))
			modeIndex := uintptr(sub409A70Native(int16(tc.flags)))
			hostLimit := memmap.PtrUint16(0x5D4594, 3488+2*modeIndex)
			oldHost, oldClient := *hostLimit, settings.Field54
			t.Cleanup(func() { *hostLimit, settings.Field54 = oldHost, oldClient })
			if tc.flags.Has(noxflags.GameHost) {
				*hostLimit, settings.Field54 = limit, 99
			} else {
				*hostLimit, settings.Field54 = 99, limit
			}

			world := playerIDNativeFixture417040(t)
			slots := world.Players.ListSlots()
			winTeam, loseTeam := world.Teams.Create(1), world.Teams.Create(2)
			front := &matchWinnerTestClient{}
			oldClientGetter := GetClient
			GetClient = func() Client { return front }
			t.Cleanup(func() { GetClient = oldClientGetter })
			winTeam.Lessons, loseTeam.Lessons = 4, 22

			for i, player := range slots[:3] {
				*player = server.Player{PlayerInd: byte(i), Active: 1, NetCodeVal: 0x80000000 + uint32(i)}
				id := loseTeam.ID()
				if i == 0 {
					id = winTeam.ID()
				}
				if tc.flags.Has(noxflags.GameHost) {
					unit, freeUnit := alloc.NewHigh(server.Object{})
					update, freeUpdate := alloc.NewHigh(server.PlayerUpdateData{})
					t.Cleanup(freeUnit)
					t.Cleanup(freeUpdate)
					*unit = server.Object{ObjClass: object.ClassPlayer, NetCode: player.NetCodeVal, UpdateData: unsafe.Pointer(update)}
					unit.TeamVal.ID, player.PlayerUnit, update.Player = id, unit, player
				} else {
					drawable := &client.Drawable{NetCode32: player.NetCodeVal}
					drawable.TeamVal.ID = id
					front.Cli().Objs.PlayerListAdd(drawable)
				}
			}
			slots[2].Field3680 = 1

			setScores := func() {
				for _, player := range slots[:3] {
					player.Lessons, player.Field2140 = 22, 2
				}
				slots[0].Field2140 = 22
			}
			check := func(lessons [3]int32, deaths [3]uint32) {
				t.Helper()
				for i, player := range slots[:3] {
					if player.Lessons != lessons[i] || player.Field2140 != deaths[i] {
						t.Fatalf("player %d score/deaths = %d/%d, want %d/%d", i, player.Lessons, player.Field2140, lessons[i], deaths[i])
					}
				}
			}
			setScores()
			sub_4947E0((*nox_playerInfo)(slots[0].C()))
			elimination := tc.flags.Has(noxflags.GameModeElimination)
			if elimination {
				check([3]int32{22, 22, 22}, [3]uint32{11, 12, 2})
			} else {
				check([3]int32{12, 11, 22}, [3]uint32{22, 2, 2})
			}

			setScores()
			sub_4948B0((*nox_team_t)(winTeam.C()))
			if elimination {
				check([3]int32{22, 22, 22}, [3]uint32{22, 12, 2})
				if winTeam.Lessons != 4 || loseTeam.Lessons != 22 {
					t.Fatal("elimination winner changed team points")
				}
			} else {
				check([3]int32{22, 11, 22}, [3]uint32{22, 2, 2})
				if winTeam.Lessons != 12 || loseTeam.Lessons != 11 {
					t.Fatalf("team scores = %d/%d, want 12/11", winTeam.Lessons, loseTeam.Lessons)
				}
			}
		})
	}
}
