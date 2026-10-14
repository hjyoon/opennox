package legacy

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"unicode/utf16"
	"unsafe"

	"github.com/opennox/libs/console"
	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/strman"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

// Real C dispatcher, native Player registry, string manager, guide lookup,
// UTF-16 formatting, centered-message storage and console export. Spell and
// ability title services use controlled fixture data; no lookup/format/output
// result is substituted. The original blobs are isolated from parent tests.
func TestClientQuestNotices48EA70ActualPacketDispatcher(t *testing.T) {
	const child = "OPENNOX_TEST_QUEST_NOTICES_48EA70"
	if os.Getenv(child) != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestClientQuestNotices48EA70ActualPacketDispatcher$", "-test.count=1", "-test.v")
		cmd.Env = append(os.Environ(), child+"=1")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("isolated Quest/name dispatcher: %v\n%s", err, out)
		}
		t.Log(strings.TrimSpace(string(out)))
		return
	}
	InitBlobData()
	type noticeCase struct {
		name                              string
		sub                               byte
		disconnected, missing             bool
		identity                          int
		settings, acquired, chat, pending bool
		index                             byte
	}
	var cases []noticeCase
	for sub := byte(30); sub <= 32; sub++ {
		for connected := 0; connected < 2; connected++ {
			for found := 0; found < 2; found++ {
				for identity := 0; identity < 3; identity++ {
					cases = append(cases, noticeCase{name: fmt.Sprintf("reward%d/connected%d/found%d/identity%d", sub, connected, found, identity),
						sub: sub, disconnected: connected == 0, missing: found == 0, identity: identity})
				}
			}
		}
	}
	for acquired := 0; acquired < 2; acquired++ {
		for index := 0; index < 2; index++ {
			for chat := 0; chat < 2; chat++ {
				for pending := 0; pending < 2; pending++ {
					cases = append(cases, noticeCase{name: fmt.Sprintf("name/acquired%d/index%d/chat%d/pending%d", acquired, index, chat, pending),
						settings: true, acquired: acquired != 0, index: byte(index), chat: chat != 0, pending: pending != 0})
				}
			}
		}
	}
	for _, tc := range cases {
		for prior := uint32(0); prior < 3; prior++ {
			t.Run(fmt.Sprintf("%s/center%d", tc.name, prior), func(t *testing.T) {
				s := playerIDNativeFixture417040(t)
				if err := s.Strings().ReadJSON("testdata/client_quest_notices_48ea70_strings.json"); err != nil {
					t.Fatal(err)
				}
				s.SetFrame(37)
				s.SetTickRate(30)
				oldFlags := noxflags.GetGame()
				noxflags.ResetGame()
				if tc.chat {
					noxflags.SetGame(noxflags.GameModeChat)
				}
				t.Cleanup(func() { noxflags.ResetGame(); noxflags.SetGame(oldFlags) })
				oldConnected := Nox_client_isConnected
				Nox_client_isConnected = func() bool { return !tc.disconnected }
				t.Cleanup(func() { Nox_client_isConnected = oldConnected })
				slots := s.Players.ListSlots()
				for _, p := range slots {
					*p = server.Player{}
				}
				p, other := slots[17], slots[18]
				*p = server.Player{Active: 1, NetCodeVal: 0xffff, PlayerInd: 99, GoldVal: 0x5a5a5a5a}
				*other = server.Player{Active: 1, NetCodeVal: 0x7ffe, PlayerInd: 18}
				p.SetName("수혜자🚀")
				other.SetName("다른🚀")
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(p)) <= 0xffffffff {
					t.Fatalf("Player not above 4 GiB: %p", p)
				}
				before := make([][]byte, len(slots))
				for i, p := range slots {
					before[i] = playerIDBytes417040(p)
				}
				oldPlayer := Get_dword_8531A0_2576()
				t.Cleanup(func() { Set_dword_8531A0_2576(oldPlayer) })
				switch tc.identity {
				case 0:
					Set_dword_8531A0_2576(p)
				case 1:
					Set_dword_8531A0_2576(other)
				default:
					Set_dword_8531A0_2576(nil)
				}
				var titleCalls []int
				oldSpell, oldAbility := Nox_xxx_spellTitle_424930, Nox_xxx_abilityGetName_0_425260
				Nox_xxx_spellTitle_424930 = func(id int) (string, bool) {
					titleCalls = append(titleCalls, id)
					return s.Strings().GetStringInFile(strman.ID("FixtureSpell"), "cdecode.c"), true
				}
				Nox_xxx_abilityGetName_0_425260 = func(id int) string {
					titleCalls = append(titleCalls, id)
					return s.Strings().GetStringInFile(strman.ID("FixtureAbility"), "cdecode.c")
				}
				t.Cleanup(func() { Nox_xxx_spellTitle_424930 = oldSpell; Nox_xxx_abilityGetName_0_425260 = oldAbility })
				t.Cleanup(questNoticeGuideFixture48EA70("生物🚀"))
				printer := new(scorePacketPrinter48EA70)
				cn, oldConsole := console.New(printer), GetConsole
				GetConsole = func() *console.Console { return cn }
				t.Cleanup(func() { GetConsole = oldConsole })
				center := memmap.PtrT[[1932]byte](0x5d4594, 823804)
				oldCenter, oldIndex := *center, scorePacketCenteredIndex48EA70()
				t.Cleanup(func() { *center = oldCenter; setScorePacketCenteredIndex48EA70(oldIndex) })
				for i := range center {
					center[i] = 0x5a
				}
				setScorePacketCenteredIndex48EA70(prior)
				wantCenter, wantIndex := *center, prior
				settings := memmap.PtrT[[116]byte](0x5d4594, 371380)
				oldSettings := *settings
				t.Cleanup(func() { *settings = oldSettings })
				for i := range settings {
					settings[i] = 0xa5
				}
				settings[52], settings[58+52] = 0x1f, 0x1f
				wantSettings := *settings
				oldAcquired := *memmap.PtrUint32(0x5d4594, 371704)
				t.Cleanup(func() { *memmap.PtrUint32(0x5d4594, 371704) = oldAcquired })
				*memmap.PtrUint32(0x5d4594, 371704) = 0
				if tc.acquired {
					*memmap.PtrUint32(0x5d4594, 371704) = 1
				}
				oldPending, oldHelp := questNoticeNamePending48EA70(), Get_nox_server_sanctuaryHelp_54276()
				Set_dword_5d4594_1200832(0)
				if tc.pending {
					Set_dword_5d4594_1200832(1)
				}
				Set_nox_server_sanctuaryHelp_54276(0)
				t.Cleanup(func() { Set_dword_5d4594_1200832(oldPending); Set_nox_server_sanctuaryHelp_54276(oldHelp) })
				op, size := netmsg.MSG_GAUNTLET, 5
				if tc.settings {
					op, size = netmsg.MSG_UPDATE_GUI_GAME_SETTINGS, 60
				}
				backing := bytes.Repeat([]byte{0x97}, size+4)
				packet := backing[2 : 2+size]
				packet[0] = byte(op)
				var wantText string
				var wantTitles []int
				if tc.settings {
					packet[1] = tc.index
					for i := 0; i < 58; i++ {
						packet[i+2] = byte(0x81 + i)
					}
					packet[54] = 0x1f // Do not request the separate GUI mode refresh.
					copy(wantSettings[int(tc.index)*58:], packet[2:])
					if !tc.acquired && tc.index == 1 {
						copy(wantSettings[:58], wantSettings[58:])
					}
					if !tc.acquired && tc.chat && tc.pending {
						wantText = "이름 [수혜자🚀]"
					}
				} else {
					packet[1], packet[2] = tc.sub, 1
					code := uint16(0xffff)
					if tc.missing {
						code = 0xfffe
					}
					binary.LittleEndian.PutUint16(packet[3:], code)
					if !tc.disconnected && !tc.missing {
						prefix, title := "스펠", "주문🚀"
						if tc.sub == 31 {
							prefix, title = "도감", "生物🚀"
						} else {
							if tc.sub == 32 {
								prefix, title = "스킬", "기술🚀"
							}
							wantTitles = []int{1}
						}
						if tc.identity == 0 {
							wantText = prefix + " [" + title + "]"
						} else {
							wantText = prefix + " [수혜자🚀] [" + title + "]"
						}
					}
				}
				packetBefore := bytes.Clone(backing)
				if got := Nox_xxx_netOnPacketRecvCli_48EA70_switch(0, op, packet); got != size {
					t.Fatalf("consumed %d bytes, want %d", got, size)
				}
				wantPlayer := bytes.Clone(before[17])
				if !tc.settings && !tc.missing {
					wantPlayer[int(unsafe.Offsetof(p.PlayerInd))] = 17 // Lookup precedes connection gate.
				}
				var wantLines []string
				var wantColors []console.Color
				if wantText != "" {
					wantLines, wantColors = []string{"알림: " + wantText}, []console.Color{console.ColorRed}
					wantIndex = (prior + 1) % 3
					start := int(wantIndex) * 644
					for i, word := range append(utf16.Encode([]rune(wantText)), 0) {
						binary.LittleEndian.PutUint16(wantCenter[start+2*i:], word)
					}
					binary.LittleEndian.PutUint32(wantCenter[start+636:], 187)
					wantCenter[start+640] = 0
				}
				if !reflect.DeepEqual(printer.lines, wantLines) || !reflect.DeepEqual(printer.colors, wantColors) || !reflect.DeepEqual(titleCalls, wantTitles) {
					t.Fatalf("output=%q/%v titles=%v, want=%q/%v titles=%v", printer.lines, printer.colors, titleCalls, wantLines, wantColors, wantTitles)
				}
				if *center != wantCenter || scorePacketCenteredIndex48EA70() != wantIndex || *settings != wantSettings || !bytes.Equal(backing, packetBefore) {
					t.Fatal("centered slots/timer, settings or packed packet/guards differ")
				}
				if tc.settings {
					wantPending := 0
					if tc.pending && wantText == "" {
						wantPending = 1
					}
					if *memmap.PtrUint32(0x5d4594, 371704) != 1 || questNoticeNamePending48EA70() != wantPending {
						t.Fatal("settings acquisition/name flag differs")
					}
				}
				for i, slot := range slots {
					want := before[i]
					if i == 17 {
						want = wantPlayer
					}
					if !bytes.Equal(playerIDBytes417040(slot), want) {
						t.Fatalf("Player %d changed outside original index store", i)
					}
				}
				t.Logf("actual Quest/name op%d sub%d Player=%p name=%q center%d->%d", op, packet[1], p, wantText, prior, wantIndex)
			})
		}
	}
}
