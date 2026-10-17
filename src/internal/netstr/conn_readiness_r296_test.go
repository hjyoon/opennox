//go:build darwin

package netstr

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"golang.org/x/sys/unix"

	"github.com/opennox/opennox/v1/server/netlib"
)

// Real owned UDP sockets and independent poll(2) arrival checks. No production
// readiness hook, fake PacketConn, global socket, clock, or direct handler call.
func readinessSocketsR296(t *testing.T) (*net.UDPConn, *net.UDPConn) {
	t.Helper()
	open := func() *net.UDPConn {
		t.Helper()
		pc, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
		if err != nil {
			t.Fatal(err)
		}
		// recvRoot currently ignores read errors. Closing this owned socket
		// also bounds a bad readiness/error loop after its read deadline.
		done := make(chan struct{})
		watchdog := time.AfterFunc(4*time.Second, func() {
			defer close(done)
			_ = pc.Close()
		})
		t.Cleanup(func() {
			if !watchdog.Stop() {
				<-done
				t.Error("owned UDP fixture exceeded its four-second watchdog")
			}
			_ = pc.Close()
		})
		if err := pc.SetDeadline(time.Now().Add(3 * time.Second)); err != nil {
			t.Fatal(err)
		}
		return pc
	}
	return open(), open()
}

func readinessPollR296(pc *net.UDPConn, timeoutMS int) (bool, error) {
	raw, err := pc.SyscallConn()
	if err != nil {
		return false, err
	}
	var ready bool
	var pollErr error
	err = raw.Control(func(fd uintptr) {
		if fd > uintptr(^uint32(0)>>1) {
			pollErr = fmt.Errorf("owned UDP descriptor exceeds pollfd range: %d", fd)
			return
		}
		fds := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}
		n, err := unix.Poll(fds, timeoutMS)
		if err != nil {
			pollErr = err
			return
		}
		if fds[0].Revents&(unix.POLLERR|unix.POLLHUP|unix.POLLNVAL) != 0 {
			pollErr = fmt.Errorf("owned UDP poll failure: %#x", fds[0].Revents)
			return
		}
		ready = n > 0 && fds[0].Revents&unix.POLLIN != 0
	})
	if err != nil {
		return false, err
	}
	return ready, pollErr
}

func readinessAwaitR296(t *testing.T, pc *net.UDPConn) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		ready, err := readinessPollR296(pc, 25)
		if errors.Is(err, unix.EINTR) {
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		if ready {
			return
		}
	}
	t.Fatal("independent poll did not observe the owned UDP datagram")
}

func readinessSendR296(t *testing.T, sender, receiver *net.UDPConn, payload []byte) {
	t.Helper()
	n, err := sender.WriteToUDP(payload, receiver.LocalAddr().(*net.UDPAddr))
	if err != nil || n != len(payload) {
		t.Fatalf("actual owned UDP send n=%d/%d err=%v", n, len(payload), err)
	}
}

func readinessReadR296(t *testing.T, receiver, sender *net.UDPConn, want []byte) {
	t.Helper()
	if err := receiver.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
		t.Fatal(err)
	}
	var buf [32]byte
	n, from, err := receiver.ReadFromUDP(buf[:])
	if err != nil || n != len(want) || from == nil || from.AddrPort() != sender.LocalAddr().(*net.UDPAddr).AddrPort() || !bytes.Equal(buf[:n], want) {
		t.Fatalf("owned UDP read n=%d from=%v data=%x err=%v want=%x", n, from, buf[:n], err, want)
	}
}

func readinessEmptyR296(t *testing.T, pc *net.UDPConn) {
	t.Helper()
	ready, err := readinessPollR296(pc, 0)
	if err != nil || ready {
		t.Fatalf("independent poll after consumption ready=%t err=%v", ready, err)
	}
	n, err := canReadConn(false, Log, pc)
	if err != nil || n != 0 {
		t.Fatalf("empty production readiness n=%d err=%v", n, err)
	}
}

func TestDarwinUDPReadinessR296(t *testing.T) {
	t.Run("no-data", func(t *testing.T) {
		receiver, _ := readinessSocketsR296(t)
		readinessEmptyR296(t, receiver)
	})
	for _, tc := range []struct {
		name string
		data []byte
	}{
		{"zero", nil},
		{"positive", []byte{0x39, 0x7a, 0x83}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			receiver, sender := readinessSocketsR296(t)
			readinessSendR296(t, sender, receiver, tc.data)
			readinessAwaitR296(t, receiver)
			n, err := canReadConn(false, Log, receiver)
			if err != nil || n <= 0 {
				t.Errorf("poll-readable datagram has nonpositive production readiness n=%d err=%v", n, err)
			}
			// Read only after measuring production readiness; this control proves
			// the exact real datagram existed, including its zero-length payload.
			readinessReadR296(t, receiver, sender, tc.data)
			readinessEmptyR296(t, receiver)
		})
	}
	t.Run("zero-before-positive-tail", func(t *testing.T) {
		receiver, sender := readinessSocketsR296(t)
		readinessSendR296(t, sender, receiver, nil)
		readinessAwaitR296(t, receiver)
		readinessSendR296(t, sender, receiver, []byte{0x39, 0x7a, 0x83})
		n, err := canReadConn(false, Log, receiver)
		if err != nil || n <= 0 {
			t.Errorf("head zero must be readable even with later positive UDP n=%d err=%v", n, err)
		}
		readinessReadR296(t, receiver, sender, nil)
		readinessAwaitR296(t, receiver)
		readinessReadR296(t, receiver, sender, []byte{0x39, 0x7a, 0x83})
		readinessEmptyR296(t, receiver)
	})
}

type readinessConnR296 struct {
	conn     *Conn
	receiver *net.UDPConn
	sender   *net.UDPConn
	events   [][]byte
}

func readinessNewConnR296(t *testing.T) *readinessConnR296 {
	t.Helper()
	receiver, sender := readinessSocketsR296(t)
	f := &readinessConnR296{receiver: receiver, sender: sender}
	g := NewStreams(func() uint32 { return 1 })
	g.Debug, g.Xor = false, false
	g.GetMaxPlayers = func() int { return 2 }
	client, err := g.NewClient(&Options{Max: 2, BufferSize: 128,
		OnReceive: func(id netlib.StreamID, data []byte) int {
			if id != f.conn {
				t.Errorf("actual Conn dispatched a different stream: %v", id)
			}
			f.events = append(f.events, append([]byte(nil), data...))
			return len(data)
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	f.conn = client.Conn
	f.conn.pc = receiver
	f.conn.setAddr(sender.LocalAddr().(*net.UDPAddr).AddrPort())
	t.Cleanup(func() { _ = client.Close() })
	return f
}

func (f *readinessConnR296) receive(t *testing.T, flags RecvFlags) int {
	t.Helper()
	if err := f.receiver.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
		t.Fatal(err)
	}
	n := f.conn.recvLoop(flags)
	if f.conn.recv.Len() != 0 {
		t.Errorf("actual receive loop retained %d payload bytes", f.conn.recv.Len())
	}
	return n
}

func TestConnZeroDatagramReceiveR296(t *testing.T) {
	t.Run("no-data", func(t *testing.T) {
		f := readinessNewConnR296(t)
		if n := f.receive(t, RecvCanRead); n != -1 || len(f.events) != 0 {
			t.Fatalf("empty Conn receive n=%d callbacks=%x", n, f.events)
		}
		readinessEmptyR296(t, f.receiver)
	})
	t.Run("zero-only", func(t *testing.T) {
		f := readinessNewConnR296(t)
		readinessSendR296(t, f.sender, f.receiver, nil)
		readinessAwaitR296(t, f.receiver)
		if n := f.receive(t, RecvCanRead); n != 0 || len(f.events) != 0 {
			t.Fatalf("actual zero Conn receive n=%d callbacks=%x", n, f.events)
		}
		readinessEmptyR296(t, f.receiver)
	})
	t.Run("zero-before-valid-just-one", func(t *testing.T) {
		f := readinessNewConnR296(t)
		readinessSendR296(t, f.sender, f.receiver, nil)
		readinessAwaitR296(t, f.receiver)
		readinessSendR296(t, f.sender, f.receiver, []byte{0, 0xff, 0xbe}) // registered stream0, current sequence, KEEP_ALIVE
		if n := f.receive(t, RecvCanRead|RecvJustOne); n != 0 || len(f.events) != 0 {
			t.Fatalf("JustOne must consume only head zero n=%d callbacks=%x", n, f.events)
		}
		readinessAwaitR296(t, f.receiver)
		if n := f.receive(t, RecvCanRead|RecvJustOne); n != 3 || len(f.events) != 1 || !bytes.Equal(f.events[0], []byte{0xbe}) {
			t.Fatalf("preserved valid tail n=%d callbacks=%x", n, f.events)
		}
		readinessEmptyR296(t, f.receiver)
	})
	t.Run("valid-before-zero-just-one", func(t *testing.T) {
		f := readinessNewConnR296(t)
		readinessSendR296(t, f.sender, f.receiver, []byte{0, 0xff, 0xbe})
		readinessAwaitR296(t, f.receiver)
		readinessSendR296(t, f.sender, f.receiver, nil)
		if n := f.receive(t, RecvCanRead|RecvJustOne); n != 3 || len(f.events) != 1 || !bytes.Equal(f.events[0], []byte{0xbe}) {
			t.Fatalf("first valid packet n=%d callbacks=%x", n, f.events)
		}
		readinessAwaitR296(t, f.receiver)
		if n := f.receive(t, RecvCanRead|RecvJustOne); n != 0 || len(f.events) != 1 {
			t.Fatalf("preserved zero tail n=%d callbacks=%x", n, f.events)
		}
		readinessEmptyR296(t, f.receiver)
	})
}

func TestConnUDPDrainCallbacksR296(t *testing.T) {
	t.Run("mixed-short-zero-drain", func(t *testing.T) {
		f := readinessNewConnR296(t)
		for _, packet := range [][]byte{nil, {1}, {1, 2}, {0, 0xff, 0xbe}, nil, {0, 0xff, 0xc0}} {
			readinessSendR296(t, f.sender, f.receiver, packet)
		}
		deadline := time.Now().Add(2 * time.Second)
		for {
			readinessAwaitR296(t, f.receiver)
			if err := f.receiver.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
				t.Fatal(err)
			}
			if n := f.conn.RecvLoop(false); n < 0 {
				t.Fatalf("poll-readable queue was not consumed n=%d callbacks=%x", n, f.events)
			}
			if f.conn.recv.Len() != 0 || len(f.events) > 2 || (len(f.events) >= 1 && !bytes.Equal(f.events[0], []byte{0xbe})) || (len(f.events) == 2 && !bytes.Equal(f.events[1], []byte{0xc0})) {
				t.Fatalf("short/zero packets reached callback or valid order changed: %x", f.events)
			}
			ready, err := readinessPollR296(f.receiver, 0)
			if err != nil {
				t.Fatal(err)
			}
			if len(f.events) == 2 && !ready {
				break
			}
			if time.Now().After(deadline) {
				t.Fatalf("actual drain did not finish callbacks=%x ready=%t", f.events, ready)
			}
		}
		readinessEmptyR296(t, f.receiver)
	})
	t.Run("ordinary-valid-public-receive", func(t *testing.T) {
		f := readinessNewConnR296(t)
		readinessSendR296(t, f.sender, f.receiver, []byte{0, 0xff, 0xbe})
		readinessAwaitR296(t, f.receiver)
		if n := f.conn.RecvLoop(false); n != 3 || len(f.events) != 1 || !bytes.Equal(f.events[0], []byte{0xbe}) {
			t.Fatalf("ordinary registered-source datagram n=%d callbacks=%x", n, f.events)
		}
		readinessEmptyR296(t, f.receiver)
	})
	t.Run("ordinary-valid-no-hooks", func(t *testing.T) {
		f := readinessNewConnR296(t)
		readinessSendR296(t, f.sender, f.receiver, []byte{0, 0xff, 0xbe})
		readinessAwaitR296(t, f.receiver)
		if n := f.receive(t, RecvCanRead|RecvNoHooks|RecvJustOne); n != 3 || len(f.events) != 0 {
			t.Fatalf("NoHooks changed consume/suppression n=%d callbacks=%x", n, f.events)
		}
		readinessEmptyR296(t, f.receiver)
	})
}
