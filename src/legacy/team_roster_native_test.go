package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestTeamRosterNativeReportsDescriptorsAndMembers4DF2E0(t *testing.T) {
	world := objectStatusTestServer(t, 3)
	setImportantRecipientMaskC(1 << 6)
	flags := noxflags.GetGame()
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameHost | noxflags.GameModeCoopTeam)
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(flags)
	})
	first, second := world.Teams.Create(2), world.Teams.Create(7)
	team := (*nox_team_t)(first.C())
	sub_418800((*wchar2_t)(first.C()), internWStr("Téam"), 0x9c)
	sub_418830(team, 0x78563412)
	first.Lessons = -1

	const code = 0x1234abcd
	player := world.Players.NewRaw(code)
	unit, freeUnit := alloc.NewHigh(server.Object{})
	t.Cleanup(freeUnit)
	update, freeUpdate := alloc.NewHigh(server.PlayerUpdateData{})
	t.Cleanup(freeUpdate)
	*unit = server.Object{ObjClass: object.ClassPlayer, NetCode: code, TypeInd: 0xfedc, UpdateData: unsafe.Pointer(update)}
	unit.TeamVal.ID = first.ID()
	player.PlayerUnit, update.Player = unit, player
	if got := nox_xxx_objGetTeamByNetCode_418C80(code); unsafe.Pointer(got) != unsafe.Pointer(&unit.TeamVal) {
		t.Fatalf("team record = %p, want %p", got, &unit.TeamVal)
	}
	if got := sub_418BC0(team); got != 1 {
		t.Fatalf("team member count = %d", got)
	}
	if got := sub_4189D0(); asTeam(got) != second {
		t.Fatalf("least occupied team = %p, want %p", got, second)
	}
	if got := sub_418A10(); asTeam(got) != second {
		t.Fatalf("team without ID word = %p, want %p", got, second)
	}
	if got := sub_418A80(1); asTeam(got) != first {
		t.Fatalf("first team color = %p, want %p", got, first)
	}
	if got := sub_418A40(internWStr("téAM")); got != team || sub_4190F0(internWStr("unknown")) != 0 {
		t.Fatalf("case-insensitive team lookup = %p, want %p", got, team)
	}

	sub_4DF2E0(server.HostPlayerIndex)
	if packet, _ := importantNativeListC(); packet != nil {
		t.Fatal("team roster enqueued for the host")
	}
	sub_4DF2E0(6)
	expected := [][]byte{
		{0xc4, 0, 7, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 2, 0},
		{0xc4, 1, 2, 0, 0, 0, 0xcd, 0xab, 0xdc, 0xfe},
		{0xc4, 0, 2, 0, 0, 0, 0x12, 0x34, 0x56, 0x78, 0xff, 0xff, 0xff, 0xff,
			1, 4, 1, 0x9c, 'T', 0, 0xe9, 0, 'a', 0, 'm', 0},
	}
	packet, _ := importantNativeListC()
	for _, want := range expected {
		checkObjectStatusPacket(t, packet, 6, want, 1, 1)
		packet = importantPacketNextC(packet)
	}
	if packet != nil {
		t.Fatal("extra team roster packets")
	}
}
