package netstr

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"net/netip"
	"testing"
	"time"
)

// Exercise the real socket reader, not a substituted readFrom result. Bytes
// after the datagram are deliberately nonzero and must never reach a handler.
func TestExtensionDatagramBoundsR270(t *testing.T) {
	receiver, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	defer receiver.Close()
	sender, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	defer sender.Close()
	// Darwin's default UDP send buffer is smaller than the receive buffer
	// exercised below. Change only this test-owned socket, not the OS limit.
	if err := sender.SetWriteBuffer(64 << 10); err != nil {
		t.Fatal(err)
	}
	streams := NewStreams(nil)
	wantFrom := sender.LocalAddr().(*net.UDPAddr).AddrPort()
	exchange := func(t *testing.T, packet, buffer []byte) (n int, from netip.AddrPort, err error) {
		t.Helper()
		if err := receiver.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
			t.Fatal(err)
		}
		if n, err := sender.WriteToUDP(packet, receiver.LocalAddr().(*net.UDPAddr)); err != nil || n != len(packet) {
			t.Fatalf("actual loopback send = %d, %v", n, err)
		}
		return streams.recvRaw(receiver, buffer)
	}
	for _, guard := range []byte{0, 0x5a, 0xff} {
		t.Run(fmt.Sprintf("guard_%02x", guard), func(t *testing.T) {
			for _, size := range []int{2, 3, 4, 5, 8, 9, 22, 64, 4096, 16384} {
				packet := bytes.Repeat([]byte{0xa5}, size)
				binary.LittleEndian.PutUint16(packet, 0xf13a)
				buffer := bytes.Repeat([]byte{guard}, 16384)
				calls := 0
				streams.OnExtPacket = func(pc net.PacketConn, got []byte, from netip.AddrPort) {
					calls++
					if pc != receiver || from != wantFrom || !bytes.Equal(got, packet) {
						t.Errorf("datagram=%d: callback length=%d, socket/from/payload must be the actual UDP datagram", size, len(got))
					}
				}
				n, from, err := exchange(t, packet, buffer)
				if err != nil || n != 0 || from != wantFrom || calls != 1 {
					t.Fatalf("extension dispatch = (%d, %v, %v), calls=%d", n, from, err, calls)
				}
				if !bytes.Equal(buffer[:size], packet) || !bytes.Equal(buffer[size:], bytes.Repeat([]byte{guard}, len(buffer)-size)) {
					t.Fatal("socket reader changed bytes outside the datagram")
				}
			}
		})
	}
	t.Run("non_extension", func(t *testing.T) {
		streams.OnExtPacket = func(net.PacketConn, []byte, netip.AddrPort) { t.Error("ordinary datagram reached extension handler") }
		for _, packet := range [][]byte{nil, {0x3a}, {0x3a, 0xf0}, {1, 2, 3, 4, 5}} {
			buffer := bytes.Repeat([]byte{0x5a}, 64)
			n, from, err := exchange(t, packet, buffer)
			if err != nil || n != len(packet) || from != wantFrom || !bytes.Equal(buffer[:n], packet) || !bytes.Equal(buffer[n:], bytes.Repeat([]byte{0x5a}, len(buffer)-n)) {
				t.Fatalf("ordinary datagram = (%d, %v, %v), want % x", n, from, err, packet)
			}
		}
	})
	t.Run("no_extension_handler", func(t *testing.T) {
		streams.OnExtPacket = nil
		for _, packet := range [][]byte{{0x3a, 0xf1}, {0x3a, 0xf1, 0, 0}} {
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("unsupported extension must not panic without a handler: %v", r)
					}
				}()
				n, from, err := exchange(t, packet, make([]byte, 64))
				if err != nil || n != 0 || from != wantFrom {
					t.Errorf("unsupported extension = (%d, %v, %v)", n, from, err)
				}
			}()
		}
	})
}
