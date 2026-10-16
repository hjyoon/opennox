package opennox

import (
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/opennox/libs/noxnet"
	"github.com/opennox/libs/noxnet/netmsg"

	"github.com/opennox/opennox/v1/internal/netstr"
)

// Uses actual UDP, the existing Options.CheckPass boundary, the real message
// decoder, and the production authentication helper. This is not a fake auth
// service or full hosting/WAN test. No external server or new hook is used.
func TestNetworkPasswordUDPTransportR283(t *testing.T) {
	cases := []struct {
		name, configured, supplied string
		minimum, maximum           int16
		wantReply                  []byte
	}{
		{"real-not-decoy", "Secret", "Secret", -1, -1, []byte{0, 0x36, 20}},
		{"reject-decoy", "Secret", "decoy", -1, -1, []byte{0, 0x36, 19, 6}},
		{"case-sensitive", "Secret", "secret", -1, -1, []byte{0, 0x36, 19, 6}},
		{"empty", "", "", -1, -1, []byte{0, 0x36, 20}},
		{"non-ASCII", "비밀鍵", "비밀鍵", -1, -1, []byte{0, 0x36, 20}},
		{"eight-UTF16-words-with-pair", "Ab😀1234", "Ab😀1234", -1, -1, []byte{0, 0x36, 20}},
		{"matching-minimum-ping", "Secret", "Secret", 1, -1, []byte{0, 0, 16, 0, 0xd2, 4, 0, 0}},
		{"matching-maximum-ping", "Secret", "Secret", -1, 1, []byte{0, 0, 16, 0, 0xd2, 4, 0, 0}},
		{"wrong-before-ping", "Secret", "other", 1, 1, []byte{0, 0x36, 19, 6}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			f := newNetworkPasswordFixtureR283(t, test.configured, test.minimum, test.maximum)
			// Listen does not accept port0. Probe only an OS-assigned ephemeral
			// port, close the probe, then use the listener's actual returned port.
			probe, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero})
			if err != nil {
				t.Fatal(err)
			}
			port := probe.LocalAddr().(*net.UDPAddr).Port
			if err := probe.Close(); err != nil {
				t.Fatal(err)
			}
			if port < 1024 || port > 65535 {
				t.Fatalf("OS ephemeral probe port %d is outside the existing listener contract", port)
			}
			streams := netstr.NewStreams(func() uint32 { return 1 })
			streams.IsHost = func() bool { return true }
			streams.GetMaxPlayers = func() int { return 4 }
			streams.Xor = false
			// Deterministic existing clock service for literal ping bytes only.
			streams.Now = func() time.Duration { return 1234 * time.Millisecond }
			var decoded []noxnet.MsgServerPass
			options := &netstr.Options{Port: port, Max: 4, BufferSize: 2048,
				CheckPass: func(packet *noxnet.MsgServerPass) netmsg.Message {
					decoded = append(decoded, *packet)
					return noxnetCheckPassword(packet)
				},
			}
			listener, err := streams.Listen(options)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := listener.Close(); err != nil {
					t.Error(err)
				}
			})
			if options.Port < 1024 || options.Port > 65535 {
				t.Fatalf("listener returned invalid probe-derived port %d", options.Port)
			}
			peer, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = peer.Close() })
			message := &noxnet.MsgServerPass{Unk0: 0x72, Pass: test.supplied}
			packet, err := netmsg.Append([]byte{0xff, 0x36}, message)
			if err != nil {
				t.Fatal(err)
			}
			if len(packet) != 22 || packet[2] != byte(netmsg.MSG_SERVER_PASSWORD) {
				t.Fatalf("password encoder produced an unexpected request %x", packet)
			}
			endpoint := &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: options.Port}
			if n, err := peer.WriteToUDP(packet, endpoint); err != nil || n != len(packet) {
				t.Fatalf("password UDP write = %d/%d, %v", n, len(packet), err)
			}
			// The listener's public RecvLoop is nonblocking. Poll only this
			// owned socket on the test thread; no goroutine accesses Settings.
			deadline := time.Now().Add(3 * time.Second)
			for len(decoded) == 0 && time.Now().Before(deadline) {
				listener.RecvLoop(false)
				if len(decoded) == 0 {
					time.Sleep(time.Millisecond)
				}
			}
			if len(decoded) != 1 || decoded[0] != *message {
				t.Fatalf("real UDP password decoder/callback observations = %+v, want one %+v", decoded, *message)
			}
			if err := peer.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
				t.Fatal(err)
			}
			var reply [256]byte
			n, from, err := peer.ReadFromUDP(reply[:])
			if err != nil {
				t.Fatal(err)
			}
			if !from.IP.IsLoopback() || from.Port != options.Port || !bytes.Equal(reply[:n], test.wantReply) {
				t.Errorf("actual UDP reply from %v = %x, want probe port %d bytes %x", from, reply[:n], options.Port, test.wantReply)
			}
			f.checkUnchanged(t)
		})
	}
}
