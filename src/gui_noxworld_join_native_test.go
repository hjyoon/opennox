package opennox

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"net/netip"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"testing"
	"time"
	"unsafe"

	"github.com/opennox/libs/noxnet"
	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/strman"
	"github.com/opennox/lobby"
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/common/discover"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/netstr"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

const noxWorldJoinChild = "OPENNOX_TEST_JOIN_OWNER_R181"

// Real loopback datagrams enter the production lobby decoder and real C state
// dispatcher through the installed frame hook. Only translation/font/dialog
// presentation and the retry sender are observed external services. Each child
// has its own legacy globals; no live game's sockets or callbacks are replaced.
func TestNativeNoxWorldJoin(t *testing.T) {
	if mode := os.Getenv(noxWorldJoinChild); mode != "" {
		testNativeNoxWorldJoinChild(t, mode)
		return
	}
	for _, mode := range []string{"pending", "accepted", "password", "error", "timeout", "retry", "inactive", "foreign", "nonblocking"} {
		t.Run(mode, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestNativeNoxWorldJoin$", "-test.v")
			cmd.Env = append(os.Environ(), noxWorldJoinChild+"="+mode)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("native join %s failed: %v\n%s", mode, err, out)
			}
			if !bytes.Contains(out, []byte("NATIVE NOXWORLD JOIN PASS: "+mode)) {
				t.Fatalf("missing completed join receipt: %s", out)
			}
		})
	}
}

func testNativeNoxWorldJoinChild(t *testing.T, mode string) {
	s := server.New(nil, nil, strman.New())
	t.Cleanup(s.Close)
	noxServer = &Server{Server: s}
	g := gui.New(nil)
	t.Cleanup(g.DestroyAll)
	noxClient = &Client{Client: &client.Client{Server: s, GUI: g}}
	gui.MainBg = g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 640, 480, nil)
	g.NewWindowRaw(gui.MainBg, gui.StatusEnabled, 0, 0, 10, 10, nil).SetID(99)
	var events []string
	legacy.Sub_44A4A0 = func() int { return 1 }
	legacy.Sub_449E00 = func(string) int { events = append(events, "title"); return 1 }
	legacy.Sub_449E30 = func(string) int { events = append(events, "text"); return 1 }
	legacy.Sub_449EA0 = func(flags gui.DialogFlags) { events = append(events, fmt.Sprintf("buttons:%d", flags)) }
	legacy.Sub_44A360 = func(size int) { events = append(events, fmt.Sprintf("font:%d", size)) }
	var now uint64 = 5000
	legacy.PlatformTicks = func() uint64 { return now }
	startNoxWorldJoin()
	if legacy.Sub_43B6D0() != 1 || legacy.Sub_43AF80() != 0 {
		t.Fatal("native join did not arm the original packet gate/state")
	}
	if !s.Update2() || len(events) != 0 || noxClient.DrawFunc != nil {
		t.Fatal("player selection entered the unused C discovery window branch")
	}
	peer, err := net.ListenUDP("udp4", net.UDPAddrFromAddrPort(netip.MustParseAddrPort("127.0.0.1:0")))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = peer.Close() })
	conn, err := net.ListenUDP("udp4", net.UDPAddrFromAddrPort(netip.MustParseAddrPort("127.0.0.1:0")))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	lobbyBroadcast = conn
	endpoint := peer.LocalAddr().(*net.UDPAddr).AddrPort()
	if err := noxWorldSetSelectedServer(discover.Server{Game: lobby.Game{Address: endpoint.Addr().String(), Port: int(endpoint.Port())}}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(noxWorldClearSelectedServer)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(noxWorldSelected)) <= 0xffffffff {
		t.Fatal("selected server fixture is not above 4GiB")
	}
	receive := func(op netmsg.Op, extra ...byte) {
		t.Helper()
		packet := append([]byte{0, 0, byte(op)}, extra...)
		if _, err := peer.WriteToUDP(packet, conn.LocalAddr().(*net.UDPAddr)); err != nil {
			t.Fatal(err)
		}
		if err := conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
			t.Fatal(err)
		}
		if n, err := waitForLobbyResults(conn, netstr.RecvJustOne); err != nil || n != len(packet) {
			t.Fatalf("actual lobby datagram: n=%d err=%v", n, err)
		}
	}
	legacy.Sub_43AF90(3)
	legacy.Set_qword_5d4594_814956(25000)
	switch mode {
	case "pending":
		if !s.Update2() || legacy.Sub_43AF80() != 3 || len(events) != 0 {
			t.Fatal("pending request advanced")
		}
	case "accepted":
		receive(netmsg.MSG_SERVER_JOIN_OK)
		if legacy.Sub_43AF80() != 7 {
			t.Fatal("join approval was ignored")
		}
		if !s.Update2() || legacy.Sub_43AF80() != 1 || noxClient.DrawFunc == nil {
			t.Fatal("accepted request did not install connected draw")
		}
		if want := []string{"font:1", "text", "buttons:0"}; !reflect.DeepEqual(events, want) {
			t.Fatalf("accepted service order=%v want=%v", events, want)
		}
		receive(netmsg.MSG_SERVER_JOIN_OK)
		if !s.Update2() || legacy.Sub_43AF80() != 1 || len(events) != 3 {
			t.Fatal("duplicate approval repeated connection")
		}
	case "password":
		receive(netmsg.MSG_PASSWORD_REQUIRED)
		if !s.Update2() || legacy.Sub_43AF80() != 6 {
			t.Fatal("password request did not enter state 6")
		}
		if want := []string{"title", "text", "buttons:7", "font:0"}; !reflect.DeepEqual(events, want) {
			t.Fatalf("password service order=%v want=%v", events, want)
		}
		root := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 100, 60, nil)
		for _, id := range []uint{4001, 4002, 4009, 4010, 4006, 4007, 4008} {
			g.NewWindowRaw(root, gui.StatusEnabled|gui.StatusHidden, 0, 0, 10, 10, nil).SetID(id)
		}
		entry := gui.NewEntryFieldRaw(g, root, gui.StatusEnabled, 0, 0, 40, 20, &gui.WindowData{Style: gui.StyleEntryField}, &gui.EntryFieldData{Field_1040: 64})
		entry.SetID(4003)
		text, free := alloc.CString16("abcdefghXYZ")
		defer free()
		entry.Func94(gui.AsWindowEvent(0x401E, uintptr(unsafe.Pointer(text)), 0))
		nox_gui_curDialog_830224 = root
		*memmap.PtrUint32(0x5D4594, 830240) = 7
		noxWorldJoinDialogOK()
		if err := peer.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
			t.Fatal(err)
		}
		var wire [64]byte
		n, from, err := peer.ReadFromUDP(wire[:])
		if err != nil || n != 22 || from.Port != conn.LocalAddr().(*net.UDPAddr).Port || wire[2] != byte(netmsg.MSG_SERVER_PASSWORD) {
			t.Fatalf("password datagram n=%d from=%v err=%v", n, from, err)
		}
		if got := alloc.GoString16B(wire[4:22]); got != "abcdefgh" || binary.LittleEndian.Uint16(wire[20:22]) != 0 || legacy.Sub_43AF80() != 3 {
			t.Fatalf("original eight-unit password/state lost: %q state=%d", got, legacy.Sub_43AF80())
		}
	case "error":
		receive(netmsg.MSG_SERVER_ERROR, byte(noxnet.ErrWrongVer))
		if legacy.Sub_43AF80() != 2 {
			t.Fatal("server error ignored")
		}
		if !s.Update2() || legacy.Sub_43AF80() != 1 || legacy.Sub_43B6D0() != 0 {
			t.Fatal("error did not clear join gate")
		}
	case "timeout":
		now = 24999
		if !s.Update2() || legacy.Sub_43AF80() != 3 {
			t.Fatal("deadline expired early")
		}
		now = 25000
		if !s.Update2() || legacy.Sub_43AF80() != 2 {
			t.Fatal("original inclusive timeout boundary lost")
		}
	case "retry":
		receive(netmsg.MSG_SERVER_JOIN_FAIL)
		if !s.Update2() || legacy.Sub_43AF80() != 9 {
			t.Fatal("join failure did not schedule retry")
		}
		attempts := 0
		legacy.Nox_client_joinGame_438A90 = func() int { attempts++; return 1 }
		now = 6000
		if !s.Update2() || attempts != 0 {
			t.Fatal("retry ran at inclusive boundary")
		}
		now = 6001
		if !s.Update2() || attempts != 1 {
			t.Fatal("retry did not run after original one-second delay")
		}
	case "inactive":
		legacy.Set_dword_5d4594_815044(0)
		receive(netmsg.MSG_SERVER_JOIN_OK)
		if legacy.Sub_43AF80() != 3 || noxClient.DrawFunc != nil {
			t.Fatal("inactive join accepted late approval")
		}
	case "foreign":
		noxWorldSelected.SetAddr("192.0.2.44")
		receive(netmsg.MSG_SERVER_JOIN_OK)
		if legacy.Sub_43AF80() != 3 {
			t.Fatal("foreign address changed join state")
		}
	case "nonblocking":
		if _, err := peer.WriteToUDP([]byte{0, 0, byte(netmsg.MSG_SERVER_JOIN_OK)}, conn.LocalAddr().(*net.UDPAddr)); err != nil {
			t.Fatal(err)
		}
		deadline := time.Now().Add(time.Second)
		for legacy.Sub_43AF80() == 3 && time.Now().Before(deadline) {
			clientWaitForLobbyResults()
			runtime.Gosched()
		}
		if legacy.Sub_43AF80() != 7 {
			t.Fatal("production nonblocking frame did not consume actual UDP approval")
		}
	default:
		t.Fatal("unknown child mode")
	}
	t.Log("NATIVE NOXWORLD JOIN PASS: " + mode)
}
