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

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

type scorePacketPrinter48EA70 struct {
	colors []console.Color
	lines  []string
}

func (p *scorePacketPrinter48EA70) Print(color console.Color, line string) {
	p.colors = append(p.colors, color)
	p.lines = append(p.lines, line)
}

func (p *scorePacketPrinter48EA70) Printf(color console.Color, format string, args ...interface{}) {
	p.Print(color, fmt.Sprintf(format, args...))
}

// The whole real C dispatcher reaches the CGo player registry, localization,
// UTF-16 formatter, centered-message slots/timer and console export. Observe
// only the frontend connection flag, console printer and audio event sink.
// Isolate the original constant blobs, never clearing parent pointer slots.
func TestClientScorePackets48EA70ActualPacketDispatcher(t *testing.T) {
	const child = "OPENNOX_TEST_SCORE_PACKET_48EA70"
	if os.Getenv(child) != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestClientScorePackets48EA70ActualPacketDispatcher$", "-test.count=1", "-test.v")
		cmd.Env = append(os.Environ(), child+"=1")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("isolated score-packet dispatcher: %v\n%s", err, output)
		}
		t.Log(strings.TrimSpace(string(output)))
		return
	}
	InitBlobData()
	type packetCase struct {
		name                        string
		op                          netmsg.Op
		disconnected, missing, host bool
		elimination, debug, updated bool
		announce                    bool
		frame, stamp, cached, total uint32
		first, second               uint32
	}
	lesson := netmsg.MSG_REPORT_LESSON
	treasure := netmsg.MSG_REPORT_TREASURE_COUNT
	cases := []packetCase{
		{name: "lesson raw bits and elimination", op: lesson, elimination: true, updated: true, announce: true, frame: 0xf1234567, first: 0xffffffff, second: 6},
		{name: "lesson below limit", op: lesson, elimination: true, updated: true, frame: 37, first: 0x80000000, second: 5},
		{name: "lesson signed negative deaths", op: lesson, elimination: true, updated: true, frame: 37, first: 1, second: 0x80000000},
		{name: "lesson non elimination", op: lesson, updated: true, frame: 37, first: 5, second: 6},
		{name: "lesson host authoritative but packet announcement", op: lesson, host: true, elimination: true, announce: true, frame: 37, first: 5, second: 6},
		{name: "lesson disconnected", op: lesson, disconnected: true, elimination: true, frame: 37, first: 5, second: 6},
		{name: "lesson missing player", op: lesson, missing: true, elimination: true, frame: 37, first: 5, second: 6},
		{name: "treasure fresh and debug high bit", op: treasure, debug: true, updated: true, announce: true, frame: 11, stamp: 10, cached: 5, total: 9, first: 5, second: 6},
		{name: "treasure equal frame", op: treasure, frame: 10, stamp: 10, cached: 5, total: 9, first: 5, second: 6},
		{name: "treasure old frame cached announcement", op: treasure, announce: true, frame: 9, stamp: 10, cached: 5, total: 6, first: 0, second: 0},
		{name: "treasure host cached DWORD wrap", op: treasure, host: true, announce: true, frame: 11, stamp: 10, cached: 0xffffffff, total: 0, first: 0, second: 0},
		{name: "treasure unsigned frame comparison", op: treasure, updated: true, announce: true, frame: 0x80000000, stamp: 0x7fffffff, first: 0, second: 1},
		{name: "treasure frame zero after wrap", op: treasure, frame: 0, stamp: 0xffffffff, cached: 5, total: 9, first: 5, second: 6},
		{name: "treasure zero total does not narrow wrap", op: treasure, updated: true, frame: 11, stamp: 10, first: 0xffff, second: 0},
		{name: "treasure maximum ushort count", op: treasure, updated: true, announce: true, frame: 11, stamp: 10, first: 0xfffe, second: 0xffff},
		{name: "treasure disconnected", op: treasure, disconnected: true, debug: true, frame: 11, stamp: 10, cached: 5, total: 6, first: 0, second: 1},
		{name: "treasure missing player", op: treasure, missing: true, debug: true, frame: 11, stamp: 10, cached: 5, total: 6, first: 0, second: 1},
	}
	for _, tc := range cases {
		for prior := uint32(0); prior < 3; prior++ {
			t.Run(fmt.Sprintf("%s/center%d", tc.name, prior), func(t *testing.T) {
				s := playerIDNativeFixture417040(t)
				if err := s.Strings().ReadJSON("testdata/client_score_packet_48ea70_strings.json"); err != nil {
					t.Fatal(err)
				}
				s.SetFrame(tc.frame)
				s.SetTickRate(30)
				oldFlags, oldEngine := noxflags.GetGame(), noxflags.GetEngine()
				noxflags.ResetGame()
				noxflags.ResetEngine()
				if tc.host {
					noxflags.SetGame(noxflags.GameHost)
				}
				if tc.elimination {
					noxflags.SetGame(noxflags.GameModeElimination)
				}
				if tc.debug {
					noxflags.SetEngine(noxflags.EngineNetDebug)
				}
				t.Cleanup(func() {
					noxflags.ResetGame()
					noxflags.SetGame(oldFlags)
					noxflags.ResetEngine()
					noxflags.SetEngine(oldEngine)
				})
				oldConnected := Nox_client_isConnected
				Nox_client_isConnected = func() bool { return !tc.disconnected }
				t.Cleanup(func() { Nox_client_isConnected = oldConnected })
				slots := s.Players.ListSlots()
				for _, p := range slots {
					*p = server.Player{}
				}
				const index = 17
				p := slots[index]
				*p = server.Player{Active: 1, NetCodeVal: 0x1234, PlayerInd: 99,
					Lessons: -17, Field2140: 0x77778888, Field2144: 0xaabbccdd,
					Field2152: tc.cached, Field2156: tc.total, GoldVal: 0x5a5a5a5a}
				p.SetName("플레이어🚀")
				setScorePacketStamp48EA70(p, tc.stamp)
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(p)) <= 0xffffffff {
					t.Fatalf("Player not above 4 GiB: %p", p)
				}
				before := make([][]byte, len(slots))
				for i, slot := range slots {
					before[i] = playerIDBytes417040(slot)
				}
				settings := memmap.PtrT[[58]byte](0x5d4594, 371380)
				oldSettings := *settings
				t.Cleanup(func() { *settings = oldSettings })
				for i := range settings {
					settings[i] = 0xa5
				}
				binary.LittleEndian.PutUint16(settings[54:56], 6)
				settingsBefore := *settings
				center := memmap.PtrT[[1932]byte](0x5d4594, 823804)
				oldCenter, oldIndex := *center, scorePacketCenteredIndex48EA70()
				t.Cleanup(func() { *center = oldCenter; setScorePacketCenteredIndex48EA70(oldIndex) })
				for i := range center {
					center[i] = 0x5a
				}
				setScorePacketCenteredIndex48EA70(prior)
				wantCenter, wantIndex := *center, prior
				printer := new(scorePacketPrinter48EA70)
				cn, oldConsole := console.New(printer), GetConsole
				GetConsole = func() *console.Console { return cn }
				t.Cleanup(func() { GetConsole = oldConsole })
				var sounds [][2]int
				oldSound := ClientPlaySoundSpecial
				ClientPlaySoundSpecial = func(id sound.ID, volume int) { sounds = append(sounds, [2]int{int(id), volume}) }
				t.Cleanup(func() { ClientPlaySoundSpecial = oldSound })
				size := 7
				if tc.op == lesson {
					size = 11
				}
				backing := bytes.Repeat([]byte{0x97}, size+4)
				packet := backing[2 : 2+size]
				packet[0] = byte(tc.op)
				code := uint16(0x9234)
				if tc.missing {
					code = 0x9235
				}
				binary.LittleEndian.PutUint16(packet[1:], code)
				if tc.op == lesson {
					binary.LittleEndian.PutUint32(packet[3:], tc.first)
					binary.LittleEndian.PutUint32(packet[7:], tc.second)
				} else {
					binary.LittleEndian.PutUint16(packet[3:], uint16(tc.first))
					binary.LittleEndian.PutUint16(packet[5:], uint16(tc.second))
				}
				packetBefore := bytes.Clone(backing)
				if got := Nox_xxx_netOnPacketRecvCli_48EA70_switch(0, tc.op, packet); got != size {
					t.Fatalf("consumed %d bytes, want %d", got, size)
				}
				wantPlayer := bytes.Clone(before[index])
				put := func(offset uintptr, value uint32) { binary.LittleEndian.PutUint32(wantPlayer[int(offset):], value) }
				if !tc.disconnected && !tc.missing {
					wantPlayer[int(unsafe.Offsetof(p.PlayerInd))] = index
				}
				if tc.updated {
					if tc.op == lesson {
						put(unsafe.Offsetof(p.Lessons), tc.first)
						put(unsafe.Offsetof(p.Field2140), tc.second)
						put(unsafe.Offsetof(p.Field2144), tc.frame)
					} else {
						put(unsafe.Offsetof(p.Field2152), tc.first)
						put(unsafe.Offsetof(p.Field2156), tc.second)
						binary.LittleEndian.PutUint32(wantPlayer[scorePacketStampOffset48EA70():], tc.frame)
					}
				}
				var wantLines []string
				var wantColors []console.Color
				var wantSounds [][2]int
				if tc.announce {
					text := "보물 [플레이어🚀]"
					if tc.op == lesson {
						text = "탈락 [플레이어🚀]"
						wantSounds = [][2]int{{312, 100}}
					}
					wantLines, wantColors = []string{"알림: " + text}, []console.Color{console.ColorRed}
					wantIndex = (prior + 1) % 3
					start := int(wantIndex) * 644
					for i, word := range append(utf16.Encode([]rune(text)), 0) {
						binary.LittleEndian.PutUint16(wantCenter[start+2*i:], word)
					}
					binary.LittleEndian.PutUint32(wantCenter[start+636:], tc.frame+150)
					wantCenter[start+640] = 0
				}
				if !reflect.DeepEqual(printer.lines, wantLines) || !reflect.DeepEqual(printer.colors, wantColors) || !reflect.DeepEqual(sounds, wantSounds) {
					t.Fatalf("output=%q/%v sound=%v, want=%q/%v sound=%v", printer.lines, printer.colors, sounds, wantLines, wantColors, wantSounds)
				}
				if *center != wantCenter || scorePacketCenteredIndex48EA70() != wantIndex {
					t.Fatal("centered text, timer, byte flag or slot wrap differs")
				}
				if *settings != settingsBefore || !bytes.Equal(backing, packetBefore) {
					t.Fatal("settings or packed packet/guards changed")
				}
				for i, slot := range slots {
					want := before[i]
					if i == index {
						want = wantPlayer
					}
					if !bytes.Equal(playerIDBytes417040(slot), want) {
						t.Fatalf("Player %d changed outside the original native stores", i)
					}
				}
				t.Logf("actual op%d Player=%p score=%08x/%08x/%08x treasure=%08x/%08x/%08x centered=%d->%d announce=%t",
					tc.op, p, uint32(p.Lessons), p.Field2140, p.Field2144, p.Field2152, p.Field2156, scorePacketStamp48EA70(p), prior, wantIndex, tc.announce)
			})
		}
	}
}
