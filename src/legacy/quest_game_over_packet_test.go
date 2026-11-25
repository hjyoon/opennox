package legacy

import (
	"bytes"
	"encoding/binary"
	"image"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/internal/cfixture"
	"github.com/opennox/opennox/v1/server"
)

func TestQuestGameOverWindowPacket(t *testing.T) {
	const child = "OPENNOX_TEST_QUEST_GAME_OVER_PACKET"
	if os.Getenv(child) != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestQuestGameOverWindowPacket$", "-test.count=1")
		cmd.Env = append(os.Environ(), child+"=1")
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("isolated game-over packet: %v\n%s", err, output)
		}
		return
	}
	strings := strman.New()
	path := filepath.Join(t.TempDir(), "strings.json")
	const localized = `{"lang": 8, "entries": [
		{"id":"GGOver.wnd:GeneratorsDestroyed","vals":[{"str":"발생기 %u"}]},
		{"id":"GGOver.wnd:NumSecretsFound","vals":[{"str":"비밀 %d"}]},
		{"id":"GGOver.wnd:Kills","vals":[{"str":"처치 %u"}]}
	]}`
	if err := os.WriteFile(path, []byte(localized), 0600); err != nil {
		t.Fatal(err)
	}
	if err := strings.ReadJSON(path); err != nil {
		t.Fatal(err)
	}
	srv := server.NewHigh(nil, nil, strings)
	t.Cleanup(srv.Close)
	cli := client.NewClientHigh(nil, nil, srv)
	t.Cleanup(cli.Close)
	t.Cleanup(cli.GUI.DestroyAll)
	GetClient = func() Client { return &questGameOverTimerClient49B6E0{cli: cli} }
	GetServer = func() Server { return &questGameOverTimerServer49B6E0{srv: srv} }
	Nox_client_isConnected = func() bool { return true }
	Set_nox_win_width(1024)
	Set_nox_win_height(769)

	const op = netmsg.MSG_GAUNTLET
	packet := [14]byte{byte(op), 2}
	binary.LittleEndian.PutUint16(packet[2:4], 65535)
	binary.LittleEndian.PutUint16(packet[4:6], 32768)
	binary.LittleEndian.PutUint16(packet[6:8], 1)
	binary.LittleEndian.PutUint16(packet[8:10], 999)
	wantPacket := packet
	for _, route := range []string{"legacy ABI", "Go client"} {
		t.Run(route, func(t *testing.T) {
			root := cli.GUI.NewWindowRaw(nil, gui.StatusHidden, 19, 29, 421, 251, nil)
			*questGameOverTimerRootSlot49B6E0() = root
			start := g2ServerWord(1303456)
			*start = 7
			frame := uint32(43210)
			gameFrameHook = func() uint32 {
				frame++
				return frame
			}
			type row struct {
				id   uint
				text string
			}
			var got []row
			want := []row{{10710, "발생기 65535"}, {10705, "현재"}, {10706, "기록"}, {10707, "비밀 32768"}, {10708, "처치 1"}, {10711, "계속 %"}}
			for _, count := range want {
				window := cli.GUI.NewWindowRaw(root, 0, 0, 0, 10, 10, func(_ *gui.Window, event gui.WindowEvent) gui.WindowEventResp {
					if text, ok := event.(*gui.StaticTextSetText); ok {
						got = append(got, row{count.id, text.Str})
					}
					return gui.RawEventResp(0)
				})
				window.SetID(count.id)
			}
			for offset, text := range map[uintptr]string{1302940: "현재", 1302684: "기록", 1303460: "계속 %%"} {
				alloc.StrCopy16(unsafe.Slice((*uint16)(g2ServerMem(offset)), 8), text)
			}
			var audio [][2]int
			ClientPlaySoundSpecial = func(id sound.ID, volume int) {
				audio = append(audio, [2]int{int(id), volume})
				if root.GetFlags().IsHidden() || !root.GetFlags().IsEnabled() || root.GlobalPos() != image.Pt(19, 29) {
					t.Fatal("window must be shown and enabled before the sound, with positioning still pending")
				}
			}
			var used int
			if route == "legacy ABI" {
				used = int(cfixture.Game2PacketBridgeFixture(0, int32(op), unsafe.Pointer(&packet[0]), int32(len(packet)), true))
			} else {
				used = Nox_xxx_netOnPacketRecvCli_48EA70_switch(0, op, packet[:])
			}
			if used != len(packet) || !bytes.Equal(packet[:], wantPacket[:]) {
				t.Fatalf("consumed %d bytes with payload %x", used, packet)
			}
			if pos := root.GlobalPos(); pos != image.Pt(302, 259) {
				t.Fatalf("centered position %v", pos)
			}
			if !reflect.DeepEqual(got, want) || !reflect.DeepEqual(audio, [][2]int{{1007, 100}}) || *start != 43212 || frame != 43212 {
				t.Fatalf("rows=%v, audio=%v, timer start=%d, frame=%d", got, audio, *start, frame)
			}
		})
	}
}
