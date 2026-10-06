package netstr

import (
	"bytes"
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/opennox/libs/noxnet/netmsg"
)

// Wrap a real UDP socket, keeping it reachable until test cleanup so a Go
// finalizer cannot hide a missing close. The trace also checks send/close order.
type serverCloseTestSocket struct {
	net.PacketConn
	events []string
	closes int
}

func (pc *serverCloseTestSocket) WriteTo(buf []byte, addr net.Addr) (int, error) {
	pc.events = append(pc.events, fmt.Sprintf("send:%x", buf))
	return pc.PacketConn.WriteTo(buf, addr)
}

func (pc *serverCloseTestSocket) Close() error {
	pc.events = append(pc.events, "close")
	pc.closes++
	return pc.PacketConn.Close()
}

func serverCloseTestListen(t *testing.T) (*Streams, Server, *serverCloseTestSocket, int) {
	t.Helper()
	probe, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero})
	if err != nil {
		t.Fatal(err)
	}
	port := probe.LocalAddr().(*net.UDPAddr).Port
	if err := probe.Close(); err != nil {
		t.Fatal(err)
	}
	g := NewStreams(func() uint32 { return 1 })
	g.IsHost = func() bool { return true }
	g.GetMaxPlayers = func() int { return 4 }
	g.Xor = false
	opt := &Options{Port: port, Max: 4, BufferSize: 2048}
	s, err := g.Listen(opt)
	if err != nil {
		t.Fatal(err)
	}
	pc := &serverCloseTestSocket{PacketConn: s.pc}
	s.pc = pc
	t.Cleanup(func() {
		_ = s.Close()
		_ = pc.PacketConn.Close()
	})
	return g, s, pc, opt.Port
}

func serverCloseTestReceiver(t *testing.T) *net.UDPConn {
	t.Helper()
	pc, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pc.Close() })
	return pc
}

func serverCloseTestRead(t *testing.T, pc *net.UDPConn, port int, want []byte) {
	t.Helper()
	if err := pc.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatal(err)
	}
	var buf [2048]byte
	n, from, err := pc.ReadFromUDP(buf[:])
	if err != nil {
		t.Fatal(err)
	}
	if from.Port != port || !bytes.Equal(buf[:n], want) {
		t.Fatalf("UDP from %s payload=%x, want port=%d payload=%x", from, buf[:n], port, want)
	}
}

func TestSendServerCloseListenerFlushesBeforeClose(t *testing.T) {
	g, s, pc, port := serverCloseTestListen(t)
	receiver := serverCloseTestReceiver(t)
	s.setAddr(receiver.LocalAddr().(*net.UDPAddr).AddrPort())
	header := *s.data2hdr()
	pending := []byte{byte(netmsg.MSG_KEEP_ALIVE)}
	if _, err := s.SendUnreliable(pending, false); err != nil {
		t.Fatal(err)
	}
	oldReliable := s.reliable
	s.SendServerClose()
	payload := append(header[:], pending...)
	notice := []byte{header[0], header[1], byte(netmsg.MSG_SERVER_CLOSE)}
	serverCloseTestRead(t, receiver, port, payload)
	serverCloseTestRead(t, receiver, port, notice)
	wantEvents := []string{fmt.Sprintf("send:%x", payload), fmt.Sprintf("send:%x", notice), "close"}
	if !reflect.DeepEqual(pc.events, wantEvents) || pc.closes != 1 {
		t.Fatalf("socket events=%v closes=%d, want %v/1", pc.events, pc.closes, wantEvents)
	}
	if s.pc != nil || g.Host() != nil || s.addr.IsValid() || s.id != 0 || s.ind != 0 || s.sendWrite != 0 || oldReliable.InQueue() != 0 || s.reliable.InQueue() != 0 {
		t.Fatal("listener socket/stream state was not reset")
	}
	// The server's existing teardown calls Close after SendServerClose.
	if err := s.Close(); err != nil || pc.closes != 1 {
		t.Fatalf("second close err=%v closes=%d, want nil/1", err, pc.closes)
	}
}

func TestSendServerCloseListenerPortReuse(t *testing.T) {
	for _, count := range []int{1, 3, 10, 257} {
		for _, closeAgain := range []bool{false, true} {
			t.Run(fmt.Sprintf("retries_%d_caller_close_%t", count, closeAgain), func(t *testing.T) {
				g, s, pc, port := serverCloseTestListen(t)
				receiver := serverCloseTestReceiver(t)
				for i := 0; i < count; i++ {
					s.setAddr(receiver.LocalAddr().(*net.UDPAddr).AddrPort())
					s.SendServerClose()
					if closeAgain {
						if err := s.Close(); err != nil {
							t.Fatal(err)
						}
					}
					// All old sockets remain strongly reachable via cleanup.
					// This bind must succeed immediately, without GC or sleep.
					rebound, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: port})
					if err != nil {
						t.Fatalf("retry %d: server UDP port %d was not released: %v", i+1, port, err)
					}
					if err := rebound.Close(); err != nil {
						t.Fatal(err)
					}
					if pc.closes != 1 || g.Host() != nil {
						t.Fatalf("retry %d: closes=%d host=%v, want 1/nil", i+1, pc.closes, g.Host())
					}
					if i+1 == count {
						break
					}
					opt := &Options{Port: port, Max: 4, BufferSize: 2048}
					s, err = g.Listen(opt)
					if err != nil || opt.Port != port {
						t.Fatalf("retry %d: Listen port=%d err=%v, want %d/nil", i+2, opt.Port, err, port)
					}
					pc = &serverCloseTestSocket{PacketConn: s.pc}
					s.pc = pc
					oldSocket := pc.PacketConn
					t.Cleanup(func() { _ = oldSocket.Close() })
				}
			})
		}
	}
}

func TestSendServerClosePeerKeepsSharedListener(t *testing.T) {
	g, s, pc, port := serverCloseTestListen(t)
	leaving := serverCloseTestReceiver(t)
	remaining := serverCloseTestReceiver(t)
	// Accepted peers share the listener socket, as in processStreamOp0.
	peer := g.newStreamAt(1, 1, &Options{BufferSize: 2048})
	peer.pc = pc
	peer.setAddr(leaving.LocalAddr().(*net.UDPAddr).AddrPort())
	other := g.newStreamAt(2, 2, &Options{BufferSize: 2048})
	other.pc = pc
	other.setAddr(remaining.LocalAddr().(*net.UDPAddr).AddrPort())
	peerHeader := *peer.data2hdr()
	peer.SendServerClose()
	serverCloseTestRead(t, leaving, port, []byte{peerHeader[0], peerHeader[1], byte(netmsg.MSG_SERVER_CLOSE)})
	if pc.closes != 0 || s.pc != pc || g.Host() != s.Conn || g.streams[1] != nil || g.streams[2] != other || peer.pc != nil {
		t.Fatalf("peer close changed listener/other peer: closes=%d", pc.closes)
	}
	otherHeader := *other.data2hdr()
	if _, err := other.SendUnreliable([]byte{byte(netmsg.MSG_KEEP_ALIVE)}, false); err != nil {
		t.Fatal(err)
	}
	if err := other.Flush(); err != nil {
		t.Fatalf("remaining peer cannot send after disconnect: %v", err)
	}
	serverCloseTestRead(t, remaining, port, []byte{otherHeader[0], otherHeader[1], byte(netmsg.MSG_KEEP_ALIVE)})
	other.SendServerClose()
	serverCloseTestRead(t, remaining, port, []byte{otherHeader[0], otherHeader[1], byte(netmsg.MSG_SERVER_CLOSE)})
	if pc.closes != 0 {
		t.Fatal("last accepted peer closed the shared listener")
	}
	s.SendServerClose()
	if pc.closes != 1 {
		t.Fatalf("listener close count=%d, want 1", pc.closes)
	}
}

func TestSendServerCloseNil(t *testing.T) {
	var s *Conn
	s.SendServerClose()
}

func TestSendServerCloseListenReplacesOpenListener(t *testing.T) {
	g, s, pc, port := serverCloseTestListen(t)
	opt := &Options{Port: port, Max: 4, BufferSize: 2048}
	replacement, err := g.Listen(opt)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = replacement.Close() })
	if opt.Port != port || replacement.Conn == s.Conn || g.Host() != replacement.Conn || pc.closes != 1 || !reflect.DeepEqual(pc.events, []string{"close"}) {
		t.Fatalf("Listen reset: port=%d old closes=%d events=%v", opt.Port, pc.closes, pc.events)
	}
	receiver := serverCloseTestReceiver(t)
	replacement.setAddr(receiver.LocalAddr().(*net.UDPAddr).AddrPort())
	originalSocket := replacement.pc
	t.Cleanup(func() { _ = originalSocket.Close() })
	replacement.SendServerClose()
	rebound, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: port})
	if err != nil {
		t.Fatalf("replacement UDP port not released: %v", err)
	}
	_ = rebound.Close()
}

func TestSendServerClosePreservesBusyPortFallback(t *testing.T) {
	_, owner, pc, port := serverCloseTestListen(t)
	other := NewStreams(func() uint32 { return 1 })
	opt := &Options{Port: port, Max: 4, BufferSize: 2048}
	s, err := other.Listen(opt)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	if opt.Port <= port || pc.closes != 0 || owner.pc != pc {
		t.Fatalf("busy port fallback port=%d owner closes=%d, want >%d/0", opt.Port, pc.closes, port)
	}
}
