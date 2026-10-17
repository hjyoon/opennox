package netstr

import (
	"bytes"
	"net"
	"net/netip"
	"runtime"
	"testing"
	"time"
)

// These tests exercise production sendTime, processPong through RecvLoop, and
// ProcessStats with real owned loopback UDP. Only pending-slot setup and the
// existing Now service are controlled; this is not authentication, gameplay,
// a separate-process join, or execution of the original game.
type processStatsWriteR294 struct {
	to   netip.AddrPort
	data []byte
	n    int
	err  error
}

type processStatsSocketR294 struct {
	*net.UDPConn
	writes []processStatsWriteR294
}

func (pc *processStatsSocketR294) WriteTo(data []byte, addr net.Addr) (int, error) {
	n, err := pc.UDPConn.WriteTo(data, addr)
	pc.writes = append(pc.writes, processStatsWriteR294{to: GetAddr(addr), data: append([]byte(nil), data...), n: n, err: err})
	return n, err
}

type processStatsFixtureR294 struct {
	g                             *Streams
	listener                      Server
	pc                            *processStatsSocketR294
	peer, decoy                   *net.UDPConn
	endpoint, peerAddr, decoyAddr netip.AddrPort
	now                           time.Duration
}

func newProcessStatsFixtureR294(t *testing.T) *processStatsFixtureR294 {
	t.Helper()
	if maxStructs != 128 {
		t.Fatal("literal protocol slot matrix requires exactly128 pending records")
	}
	open := func() *net.UDPConn {
		t.Helper()
		pc, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = pc.Close() })
		return pc
	}
	f := &processStatsFixtureR294{now: 5 * time.Second}
	f.pc = &processStatsSocketR294{UDPConn: open()}
	f.peer, f.decoy = open(), open()
	f.endpoint = f.pc.LocalAddr().(*net.UDPAddr).AddrPort()
	f.peerAddr = f.peer.LocalAddr().(*net.UDPAddr).AddrPort()
	f.decoyAddr = f.decoy.LocalAddr().(*net.UDPAddr).AddrPort()
	if f.peerAddr == f.decoyAddr {
		t.Fatal("owned target and canary endpoints must differ")
	}
	f.g = NewStreams(func() uint32 { return 1 })
	f.g.IsHost = func() bool { return true }
	f.g.GetMaxPlayers = func() int { return 4 }
	f.g.Now = func() time.Duration { return f.now }
	f.g.Debug = false // local only: earlier legacy tests leave global DebugSockets enabled
	f.g.Xor = false
	f.g.MaxPacketLoss = 3
	// A real ephemeral socket is attached to the same listener stream shape
	// Listen constructs. No fake transport or protocol callback is installed.
	ns := f.g.newStreamAt(0, -1, &Options{BufferSize: 2048})
	ns.pc = f.pc
	ns.data2hdr()[0] = 0
	f.listener = Server{Conn: ns}
	f.g.lis = f.listener
	return f
}

func (f *processStatsFixtureR294) resetSlots(slot int, selected stream2) {
	f.pc.writes = nil
	f.now = 5 * time.Second
	for i := range f.g.streams2 {
		f.g.streams2[i] = stream2{active: false, lost: uint8(i % 4), cur: uint8(i % 10), addr: f.decoyAddr,
			samples: [10]time.Duration{1, 2, 3, 4, 5, 6, 7, 8, 9, time.Duration(i + 10)}, ticks: time.Duration(i + 1)}
	}
	selected.addr = f.peerAddr
	f.g.streams2[slot] = selected
}

// Verify both the attempted destination and the ACTUAL resulting datagram.
// On the unfixed code a valid decoy endpoint makes wrong-slot delivery directly
// observable without hundreds of expected timeouts or an invalid-address send.
func (f *processStatsFixtureR294) checkDatagram(t *testing.T, index int, wantTo netip.AddrPort, want []byte) {
	t.Helper()
	if index >= len(f.pc.writes) {
		t.Fatalf("missing UDP write %d; have %d", index, len(f.pc.writes))
	}
	w := f.pc.writes[index]
	if w.err != nil || w.n != len(w.data) {
		t.Fatalf("UDP write %d: n=%d payload=%x err=%v", index, w.n, w.data, w.err)
	}
	if w.to != wantTo || !bytes.Equal(w.data, want) {
		t.Errorf("UDP write %d destination=%s payload=%x, want %s/%x", index, w.to, w.data, wantTo, want)
	}
	var receiver *net.UDPConn
	switch w.to {
	case f.peerAddr:
		receiver = f.peer
	case f.decoyAddr:
		receiver = f.decoy
	default:
		t.Fatalf("UDP write escaped the two owned endpoints: %s", w.to)
	}
	if err := receiver.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatal(err)
	}
	var packet [32]byte
	n, from, err := receiver.ReadFromUDP(packet[:])
	if err != nil {
		t.Fatal(err)
	}
	if from.AddrPort() != f.endpoint || !bytes.Equal(packet[:n], w.data) || !bytes.Equal(packet[:n], want) {
		t.Errorf("actual UDP write %d from=%s payload=%x, want source=%s payload=%x", index, from, packet[:n], f.endpoint, want)
	}
}

func checkProcessStatsSlotsR294(t *testing.T, got, want [maxStructs]stream2) {
	t.Helper()
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("pending slot %d changed unexpectedly: got=%+v want=%+v", i, got[i], want[i])
		}
	}
}

func TestProcessStatsAverageRecipientAllSlotsR294(t *testing.T) {
	// Literal existing 8-byte MSG_SERVER_PING packets: 5000..5180ms,
	// 20ms between actual PONGs, producing TEN strictly positive samples.
	// No cnt==0, arithmetic-wrap, sign-policy or packet-loss policy is added.
	pings := [10][8]byte{
		{0, 0, 0x10, 0, 0x88, 0x13, 0, 0}, {0, 0, 0x10, 1, 0x9c, 0x13, 0, 0},
		{0, 0, 0x10, 2, 0xb0, 0x13, 0, 0}, {0, 0, 0x10, 3, 0xc4, 0x13, 0, 0},
		{0, 0, 0x10, 4, 0xd8, 0x13, 0, 0}, {0, 0, 0x10, 5, 0xec, 0x13, 0, 0},
		{0, 0, 0x10, 6, 0, 0x14, 0, 0}, {0, 0, 0x10, 7, 0x14, 0x14, 0, 0},
		{0, 0, 0x10, 8, 0x28, 0x14, 0, 0}, {0, 0, 0x10, 9, 0x3c, 0x14, 0, 0},
	}
	cases := []struct {
		name     string
		min, max time.Duration
		packets  [][]byte
	}{
		{"accept-inclusive-equality", 20 * time.Millisecond, 20 * time.Millisecond, [][]byte{{0, 0, 0x14}}},
		{"low-error-then-OK", 21 * time.Millisecond, -1, [][]byte{{0, 0, 0x13, 0}, {0, 0, 0x14}}},
		{"high-error-then-OK", -1, 19 * time.Millisecond, [][]byte{{0, 0, 0x13, 1}, {0, 0, 0x14}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newProcessStatsFixtureR294(t)
			for slot := 0; slot < 128; slot++ {
				f.resetSlots(slot, stream2{active: true})
				if _, err := f.g.sendTime(slot); err != nil {
					t.Fatal(err)
				}
				for i, ping := range pings {
					if len(f.pc.writes) != i+1 {
						t.Fatalf("slot %d ping %d writes=%d", slot, i, len(f.pc.writes))
					}
					f.checkDatagram(t, i, f.peerAddr, ping[:])
					pong := ping
					pong[2] = 0x12
					f.now += 20 * time.Millisecond
					if n, err := f.peer.WriteToUDP(pong[:], net.UDPAddrFromAddrPort(f.endpoint)); err != nil || n != len(pong) {
						t.Fatalf("slot %d pong %d write=%d/%d err=%v", slot, i, n, len(pong), err)
					}
					deadline := time.Now().Add(2 * time.Second)
					for f.g.streams2[slot].cur == uint8(i) && time.Now().Before(deadline) {
						f.listener.RecvLoop(false)
						runtime.Gosched()
					}
					if f.g.streams2[slot].cur != uint8(i+1) {
						t.Fatalf("slot %d real PONG %d did not reach production decoder", slot, i)
					}
				}
				if len(f.pc.writes) != 10 {
					t.Fatalf("slot %d ten positive PONGs produced %d ping writes", slot, len(f.pc.writes))
				}
				positive := [10]time.Duration{20 * time.Millisecond, 20 * time.Millisecond, 20 * time.Millisecond, 20 * time.Millisecond, 20 * time.Millisecond, 20 * time.Millisecond, 20 * time.Millisecond, 20 * time.Millisecond, 20 * time.Millisecond, 20 * time.Millisecond}
				if f.g.streams2[slot].samples != positive || !f.g.streams2[slot].active {
					t.Fatalf("slot %d did not reach complete active ten-positive-sample state", slot)
				}
				before := f.g.streams2
				want := before
				want[slot].active = false
				f.g.ProcessStats(tc.min, tc.max)
				if len(f.pc.writes) != 10+len(tc.packets) {
					t.Fatalf("slot %d completion writes=%d want=%d", slot, len(f.pc.writes)-10, len(tc.packets))
				}
				for i, packet := range tc.packets {
					f.checkDatagram(t, 10+i, f.peerAddr, packet)
				}
				checkProcessStatsSlotsR294(t, f.g.streams2, want)
			}
			t.Log("all literal pending slots 0..127 completed ten actual UDP PONGs; response destination/order and all 128 slot records checked")
		})
	}
}

func TestProcessStatsSiblingBranchesR294(t *testing.T) {
	cases := []struct {
		name     string
		selected stream2
		packets  [][]byte
		change   func(*stream2)
	}{
		{"inactive-complete", stream2{cur: 10, ticks: 3 * time.Second}, nil, func(*stream2) {}},
		{"pending-exact-two-seconds", stream2{active: true, cur: 3, lost: 1, ticks: 3 * time.Second}, nil, func(*stream2) {}},
		{"pending-before-two-seconds", stream2{active: true, cur: 9, lost: 3, ticks: 3*time.Second + time.Nanosecond}, nil, func(*stream2) {}},
		{"timeout-retry-under-limit", stream2{active: true, cur: 3, lost: 1, ticks: 3*time.Second - time.Nanosecond}, [][]byte{{0, 0, 0x10, 4, 0x88, 0x13, 0, 0}}, func(s *stream2) { s.samples[3] = -1; s.lost = 2; s.cur = 4; s.ticks = 5 * time.Second }},
		{"timeout-retry-equal-limit", stream2{active: true, cur: 4, lost: 2, ticks: 3*time.Second - time.Nanosecond}, [][]byte{{0, 0, 0x10, 5, 0x88, 0x13, 0, 0}}, func(s *stream2) { s.samples[4] = -1; s.lost = 3; s.cur = 5; s.ticks = 5 * time.Second }},
		{"timeout-crosses-loss-limit", stream2{active: true, cur: 4, lost: 3, ticks: 3*time.Second - time.Nanosecond}, [][]byte{{0, 0, 0x13, 1}}, func(s *stream2) { s.samples[4] = -1; s.lost = 4; s.active = false }},
		{"complete-over-loss-limit", stream2{active: true, cur: 10, lost: 4, ticks: 3 * time.Second}, [][]byte{{0, 0, 0x13, 1}, {0, 0, 0x14}}, func(s *stream2) { s.active = false }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newProcessStatsFixtureR294(t)
			for _, slot := range []int{0, 10, 127} {
				f.resetSlots(slot, tc.selected)
				want := f.g.streams2
				tc.change(&want[slot])
				f.g.ProcessStats(20*time.Millisecond, 20*time.Millisecond)
				if len(f.pc.writes) != len(tc.packets) {
					t.Fatalf("slot %d packets=%d want=%d", slot, len(f.pc.writes), len(tc.packets))
				}
				for i, packet := range tc.packets {
					f.checkDatagram(t, i, f.peerAddr, packet)
				}
				checkProcessStatsSlotsR294(t, f.g.streams2, want)
			}
		})
	}
}

func TestProcessStatsIndependentCompletedSlotsR294(t *testing.T) {
	f := newProcessStatsFixtureR294(t)
	positive := [10]time.Duration{20 * time.Millisecond, 20 * time.Millisecond, 20 * time.Millisecond, 20 * time.Millisecond, 20 * time.Millisecond, 20 * time.Millisecond, 20 * time.Millisecond, 20 * time.Millisecond, 20 * time.Millisecond, 20 * time.Millisecond}
	f.resetSlots(0, stream2{active: true, cur: 10, samples: positive})
	// Slot10 is deliberately inactive; overwriting the outer index must not
	// retire it or redirect the two unrelated completed peers to its address.
	f.g.streams2[127] = stream2{active: true, cur: 10, samples: positive, addr: f.decoyAddr, ticks: 1}
	want := f.g.streams2
	want[0].active = false
	want[127].active = false
	f.g.ProcessStats(20*time.Millisecond, 20*time.Millisecond)
	if len(f.pc.writes) != 2 {
		t.Fatalf("independent completed peers produced %d writes, want2", len(f.pc.writes))
	}
	f.checkDatagram(t, 0, f.peerAddr, []byte{0, 0, 0x14})
	f.checkDatagram(t, 1, f.decoyAddr, []byte{0, 0, 0x14})
	checkProcessStatsSlotsR294(t, f.g.streams2, want)
	if f.g.streams2[10].active {
		t.Fatal("inactive slot10 changed")
	}
}
