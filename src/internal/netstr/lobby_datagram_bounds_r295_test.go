package netstr

import (
	"bytes"
	"fmt"
	"net"
	"net/netip"
	"runtime"
	"testing"
	"time"

	"github.com/opennox/libs/noxnet"
)

// Literal UDP fixtures exercise the actual lobby receive boundary. They do not
// emulate a host, change a global socket, authenticate, or execute GAME.EXE.
type lobbyEventR295 struct {
	kind string
	from netip.AddrPort
	data []byte
	code noxnet.ConnectError
}

func lobbySocketsR295(t *testing.T) (*net.UDPConn, *net.UDPConn) {
	t.Helper()
	open := func() *net.UDPConn {
		t.Helper()
		pc, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = pc.Close() })
		if err := pc.SetDeadline(time.Now().Add(3 * time.Second)); err != nil {
			t.Fatal(err)
		}
		return pc
	}
	return open(), open()
}

func lobbySendR295(t *testing.T, sender, receiver *net.UDPConn, packet []byte) {
	t.Helper()
	n, err := sender.WriteToUDP(packet, receiver.LocalAddr().(*net.UDPAddr))
	if err != nil || n != len(packet) {
		t.Fatalf("actual UDP send n=%d/%d err=%v", n, len(packet), err)
	}
}

// Wait only for the fixture's UDP delivery before exercising a nonblocking
// poll. A successful local WriteTo does not guarantee immediate readability.
func lobbyAwaitReadableR295(t *testing.T, receiver *net.UDPConn) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		n, err := canReadConn(false, Log, receiver)
		if err != nil {
			t.Fatal(err)
		}
		if n > 0 {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("owned UDP fixture did not become readable")
		}
		runtime.Gosched()
	}
}

func lobbyReceiveR295(t *testing.T, pc net.PacketConn, selected netip.Addr, flags RecvFlags, opts LobbyWaitOptions) (n int, err error) {
	t.Helper()
	defer func() {
		if fault := recover(); fault != nil {
			t.Errorf("actual lobby UDP receive panicked: %v", fault)
			err = fmt.Errorf("lobby receive panic: %v", fault)
		}
	}()
	return WaitForLobbyResults(pc, selected, flags, opts)
}

func lobbyCallbacksR295(events *[]lobbyEventR295, pc *net.UDPConn, continueError bool) LobbyWaitOptions {
	return LobbyWaitOptions{
		OnResult: func(from netip.AddrPort, data []byte) {
			// The production GUI consumes the name beginning at byte72. An
			// empty name is safe and remains admitted; no new NUL policy.
			_ = data[72:]
			*events = append(*events, lobbyEventR295{kind: "info", from: from, data: append([]byte(nil), data...)})
		},
		OnPassRequired: func() { *events = append(*events, lobbyEventR295{kind: "password"}) },
		OnPing: func(from netip.AddrPort, data []byte) {
			*events = append(*events, lobbyEventR295{kind: "ping", from: from, data: append([]byte(nil), data...)})
			// Same eight-byte echo as the actual GUI callback. A short slice
			// must not expose bytes remaining in the shared receive buffer.
			data[2] = 18
			if _, err := pc.WriteToUDP(data[:8], net.UDPAddrFromAddrPort(from)); err != nil {
				panic(err)
			}
		},
		OnConnectErr: func(code noxnet.ConnectError) bool {
			*events = append(*events, lobbyEventR295{kind: "error", code: code})
			return continueError
		},
		OnJoinOK:   func() { *events = append(*events, lobbyEventR295{kind: "join"}) },
		OnJoinFail: func() { *events = append(*events, lobbyEventR295{kind: "fail"}) },
	}
}

func TestLobbyShortDatagramsR295(t *testing.T) {
	for _, selected := range []bool{true, false} {
		name := "selected-IP"
		if !selected {
			name = "other-IP"
		}
		t.Run(name, func(t *testing.T) {
			cases := []struct {
				name     string
				full     []byte
				min, max int
			}{
				{"empty", nil, 0, 0},
				{"header-one", []byte{0}, 1, 1},
				{"header-two", []byte{0, 1}, 2, 2},
				{"error-prefix", []byte{0, 1, 19}, 3, 3},
				{"ping-prefixes", []byte{0, 1, 16, 9, 0x78, 0x56, 0x34, 0x12}, 3, 7},
				{"info-prefixes", append([]byte{0, 1, 13}, make([]byte, 69)...), 3, 71},
			}
			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					receiver, sender := lobbySocketsR295(t)
					address := sender.LocalAddr().(*net.UDPAddr).AddrPort().Addr()
					if !selected {
						address = netip.MustParseAddr("192.0.2.1") // comparison only; never contacted
					}
					for size := tc.min; size <= tc.max; size++ {
						var events []lobbyEventR295
						lobbySendR295(t, sender, receiver, tc.full[:size])
						n, err := lobbyReceiveR295(t, receiver, address, 0, lobbyCallbacksR295(&events, receiver, false))
						if err != nil || n != size || len(events) != 0 {
							t.Errorf("received short UDP n=%d: returned=%d err=%v callbacks=%+v", size, n, err, events)
						}
					}
				})
			}
		})
	}
}

func TestLobbyValidDatagramsR295(t *testing.T) {
	info := append([]byte{0xfe, 7, 13}, make([]byte, 69)...)
	cases := []struct {
		name   string
		packet []byte
		kind   string
	}{
		{"info-empty-name", info, "info"},
		{"info-name-and-extra", append(append([]byte(nil), info...), 'N', 'o', 'x', 0, 0x91), "info"},
		{"password", []byte{0xfe, 7, 15}, "password"},
		{"ping", []byte{0xfe, 7, 16, 9, 0x78, 0x56, 0x34, 0x12}, "ping"},
		{"error", []byte{0xfe, 7, 19, 7}, "error"},
		{"join", []byte{0xfe, 7, 20}, "join"},
		{"fail", []byte{0xfe, 7, 21}, "fail"},
		{"unknown-low", []byte{0xfe, 7, 22}, ""},
		{"accepted-stream-not-lobby", []byte{0xfe, 7, 32}, ""},
	}
	for _, tc := range cases {
		for _, selected := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/selected-%t", tc.name, selected), func(t *testing.T) {
				receiver, sender := lobbySocketsR295(t)
				peer := sender.LocalAddr().(*net.UDPAddr).AddrPort()
				address := peer.Addr()
				if !selected {
					address = netip.MustParseAddr("192.0.2.1")
				}
				var events []lobbyEventR295
				lobbySendR295(t, sender, receiver, tc.packet)
				n, err := lobbyReceiveR295(t, receiver, address, 0, lobbyCallbacksR295(&events, receiver, true))
				if err != nil || n != len(tc.packet) {
					t.Fatalf("actual valid receive n=%d err=%v", n, err)
				}
				want := tc.kind
				if !selected && want != "info" {
					want = ""
				}
				if want == "" {
					if len(events) != 0 {
						t.Fatalf("unexpected callbacks: %+v", events)
					}
					return
				}
				if len(events) != 1 || events[0].kind != want {
					t.Fatalf("callbacks=%+v want exactly %s", events, want)
				}
				got := events[0]
				if want == "info" || want == "ping" {
					if got.from != peer || !bytes.Equal(got.data, tc.packet) {
						t.Fatalf("source/payload changed: %+v want=%s/%x", got, peer, tc.packet)
					}
				}
				if want == "error" && got.code != 7 {
					t.Fatalf("connect error changed: %d", got.code)
				}
				if want == "ping" {
					var reply [32]byte
					n, from, err := sender.ReadFromUDP(reply[:])
					expected := append([]byte(nil), tc.packet...)
					expected[2] = 18
					if err != nil || from.AddrPort() != receiver.LocalAddr().(*net.UDPAddr).AddrPort() || !bytes.Equal(reply[:n], expected) {
						t.Fatalf("actual PONG from=%v n=%d data=%x err=%v want=%x", from, n, reply[:n], err, expected)
					}
				}
			})
		}
	}
}

func TestLobbyDatagramDrainFlagsR295(t *testing.T) {
	t.Run("drain-short-and-valid", func(t *testing.T) {
		receiver, sender := lobbySocketsR295(t)
		info := append([]byte{0, 7, 13}, make([]byte, 69)...)
		packets := [][]byte{{0, 7}, {0, 7, 16}, {0, 7, 16, 9, 0x78, 0x56, 0x34, 0x12}, info[:71], info, {0, 7, 19}, {0, 7, 19, 3}, {0, 7, 20}}
		for _, packet := range packets {
			lobbySendR295(t, sender, receiver, packet)
		}
		lobbyAwaitReadableR295(t, receiver)
		var events []lobbyEventR295
		n, err := lobbyReceiveR295(t, receiver, sender.LocalAddr().(*net.UDPAddr).AddrPort().Addr(), RecvCanRead, lobbyCallbacksR295(&events, receiver, false))
		if err != nil || n != 3 {
			t.Fatalf("drain result n=%d err=%v", n, err)
		}
		var kinds []string
		for _, event := range events {
			kinds = append(kinds, event.kind)
		}
		if fmt.Sprint(kinds) != "[ping info error join]" {
			t.Fatalf("callback order=%v", kinds)
		}
		if len(events[0].data) != 8 || len(events[1].data) != 72 || events[2].code != 3 {
			t.Fatalf("callback content=%+v", events)
		}
	})
	t.Run("just-one-keeps-next-datagram", func(t *testing.T) {
		receiver, sender := lobbySocketsR295(t)
		lobbySendR295(t, sender, receiver, []byte{0})
		lobbySendR295(t, sender, receiver, []byte{0, 7, 20})
		lobbyAwaitReadableR295(t, receiver)
		var events []lobbyEventR295
		opts := lobbyCallbacksR295(&events, receiver, false)
		addr := sender.LocalAddr().(*net.UDPAddr).AddrPort().Addr()
		n, err := lobbyReceiveR295(t, receiver, addr, RecvCanRead|RecvJustOne, opts)
		if err != nil || n != 1 || len(events) != 0 {
			t.Fatalf("single malformed result n=%d err=%v events=%+v", n, err, events)
		}
		n, err = lobbyReceiveR295(t, receiver, addr, 0, opts)
		if err != nil || n != 3 || len(events) != 1 || events[0].kind != "join" {
			t.Fatalf("next real datagram n=%d err=%v events=%+v", n, err, events)
		}
	})
}
