package legacy

import (
	"encoding/binary"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/opennox/libs/noxnet/netmsg"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestClientScorePacketsMatchWinnerReports(t *testing.T) {
	const child = "OPENNOX_TEST_CLIENT_MATCH_WINNER"
	if os.Getenv(child) != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestClientScorePacketsMatchWinnerReports$", "-test.count=1")
		cmd.Env = append(os.Environ(), child+"=1")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("isolated client result reports: %v\n%s", err, out)
		}
		return
	}
	world := playerIDNativeFixture417040(t)
	GetClient = func() Client { return &matchWinnerTestClient{} }
	Nox_client_isConnected = func() bool { return true }
	const now = 1901
	PlatformTicks = func() uint64 { return now }
	noxflags.SetEngine(noxflags.EngineNoRendering)
	const local = 0x1234
	ClientSetPlayerNetCode(local)
	player := world.Players.ListSlots()[3]
	*player = server.Player{Active: 1, NetCodeVal: local}
	player.SetName("전사🚀")
	team := world.Teams.Create(1)
	team.SetNameAnd68("청군🚀", 0)

	path := filepath.Join(t.TempDir(), "strings.json")
	const localized = `{"lang":8,"entries":[
		{"id":"cdecode.c:TimeLimitReached","vals":[{"str":"시간 | "}]},
		{"id":"cdecode.c:DM_MaleVictory","vals":[{"str":"승리"}]},
		{"id":"cdecode.c:DM_FemaleVictory","vals":[{"str":"여전사 승리"}]},
		{"id":"cdecode.c:DM_Loss","vals":[{"str":"패배 [%s]"}]},
		{"id":"cdecode.c:CTF_Tie","vals":[{"str":"무승부"}]},
		{"id":"cdecode.c:CTF_Victory","vals":[{"str":"깃발 [%s]"}]},
		{"id":"cdecode.c:Team","vals":[{"str":"팀 %s"}]},
		{"id":"cdecode.c:teamformat","vals":[{"str":"구단 %s"}]},
		{"id":"cdecode.c:FB_Victory","vals":[{"str":"%s 승리"}]},
		{"id":"cdecode.c:HL_Header","vals":[{"str":"생존 | "}]},
		{"id":"cdecode.c:HL_Victory","vals":[{"str":"승자 %s"}]}
	]}`
	if err := os.WriteFile(path, []byte(localized), 0600); err != nil {
		t.Fatal(err)
	}
	if err := world.Strings().ReadJSON(path); err != nil {
		t.Fatal(err)
	}

	const oldText = "기존 메시지"
	const oldStyle = 99
	textBuffer := memmap.PtrT[[264]uint16](0x5d4594, 811376)
	styleSlot := memmap.PtrInt32(0x5d4594, 811060)
	timeSlot := memmap.PtrUint64(0x5d4594, 811908)
	for _, tc := range []struct {
		name            string
		op              netmsg.Op
		code            uint16
		flags           noxflags.GameFlag
		frame, previous uint32
		female          byte
		handled         bool
		text            string
		style           int32
	}{
		{"local winner", netmsg.MSG_REPORT_DEATHMATCH_WINNER, local | 0x8000, noxflags.GameModeArena, 11, 10, 0, true, "시간 | 승리", 0},
		{"local female winner", netmsg.MSG_REPORT_DEATHMATCH_WINNER, local, noxflags.GameModeArena, 11, 10, 1, true, "시간 | 여전사 승리", 0},
		{"signed solo frame gate", netmsg.MSG_REPORT_DEATHMATCH_WINNER, local, noxflags.GameModeArena, 0x80000000, 0x7fffffff, 0, false, oldText, oldStyle},
		{"unsigned flag frame", netmsg.MSG_REPORT_FLAG_WINNER, 0, noxflags.GameModeCTF, 0x80000000, 0x7fffffff, 0, true, "시간 | 무승부", 0},
		{"flag team winner", netmsg.MSG_REPORT_FLAG_WINNER, 1, noxflags.GameModeCTF, 11, 10, 0, true, "시간 | 깃발 [청군🚀]", 1},
		{"ball winner signed frame", netmsg.MSG_REPORT_FLAG_BALL_WINNER, 1, noxflags.GameModeFlagBall, 0x7fffffff, 0x80000000, 0, true, "구단 청군🚀 승리", 1},
		{"team deathmatch", netmsg.MSG_REPORT_DEATHMATCH_TEAM_WINNER, 1, noxflags.GameModeArena, 0x80000000, 0x7fffffff, 0, true, "시간 | 패배 [팀 청군🚀]", 1},
		{"team elimination", netmsg.MSG_REPORT_DEATHMATCH_TEAM_WINNER, 1, noxflags.GameModeElimination, 11, 10, 0, true, "시간 | 생존 | 승자 팀 청군🚀", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			noxflags.ResetGame()
			noxflags.SetGame(tc.flags)
			player.Info().SetIsFemale(tc.female)
			Set_dword_5d4594_1200804(tc.previous)
			alloc.StrCopy16(textBuffer[:], oldText)
			*styleSlot, *timeSlot = oldStyle, 12
			calls := 0
			ClientPlaySoundSpecial = func(id sound.ID, volume int) {
				calls++
				if id != 309 || volume != 100 || noxflags.HasGame(noxflags.GameFlag4) {
					t.Fatalf("winner audio: id=%d volume=%d flag=%t", id, volume, noxflags.HasGame(noxflags.GameFlag4))
				}
			}
			packet := [8]byte{byte(tc.op), 0, 0, 1}
			binary.LittleEndian.PutUint16(packet[1:3], tc.code)
			binary.LittleEndian.PutUint32(packet[4:8], tc.frame)
			if used := Nox_xxx_netOnPacketRecvCli_48EA70_switch(0, tc.op, packet[:]); used != len(packet) {
				t.Fatalf("winner consumed %d bytes", used)
			}
			wantTime, wantCalls := uint64(12), 0
			if tc.handled {
				wantTime, wantCalls = now, 1
			}
			if text := alloc.GoString16S(textBuffer[:]); text != tc.text || *styleSlot != tc.style || *timeSlot != wantTime ||
				calls != wantCalls || noxflags.HasGame(noxflags.GameFlag4) != tc.handled {
				t.Fatalf("winner text=%q style=%d clock=%d sounds=%d flags=%v", text, *styleSlot, *timeSlot, calls, noxflags.GetGame())
			}
		})
	}
}
