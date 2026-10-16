package server

import (
	"bytes"
	"encoding/binary"
	"io"
	"slices"
	"testing"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/player"

	"github.com/opennox/opennox/v1/common/ntype"
)

// The wire table is stated independently of CtrlCode.DataSize. Unlisted low
// bytes retain the existing zero-payload policy; this is not a code whitelist.
func playerInputWireSizeR272(code byte) int {
	switch code {
	case 1, 2, 3, 4, 5, 28:
		return 1
	case 29, 30:
		return 4
	default:
		return 0
	}
}

func playerInputRecordR272(code byte, high [3]byte, value uint32) ([]byte, PlayerCtrl) {
	n := playerInputWireSizeR272(code)
	record := make([]byte, 4+n)
	copy(record, []byte{code, high[0], high[1], high[2]})
	var payload [4]byte
	binary.LittleEndian.PutUint32(payload[:], value)
	copy(record[4:], payload[:n])
	if n == 0 {
		value = 0
	} else if n == 1 {
		value &= 0xff
	}
	return record, PlayerCtrl{Code: player.CtrlCode(code), Data: value, Active: true}
}

// Give the actual-length slice a readable tail. A decoder must respect len,
// not silently consume bytes merely because they are inside the allocation.
func guardedPlayerInputR272(data []byte) (packet, allocation []byte) {
	allocation = bytes.Repeat([]byte{0xa5}, 16+len(data)+512)
	copy(allocation[16:], data)
	packet = allocation[16 : 16+len(data)]
	return packet, allocation
}

func TestNetworkPlayerInputDecoderWidthsR272(t *testing.T) {
	seed := []PlayerCtrl{{Code: player.CCJump, Data: 0x89abcdef, Active: false}}
	for code := 0; code < 256; code++ {
		if got, want := player.CtrlCode(code).DataSize(), playerInputWireSizeR272(byte(code)); got != want {
			t.Fatalf("wire width for code %d = %d, want %d", code, got, want)
		}
		for _, high := range [][3]byte{{}, {0x5a, 0xa5, 0xff}, {0xff, 0xff, 0xff}} {
			for _, value := range []uint32{0, 0x89abcdef, 0xffffffff} {
				record, event := playerInputRecordR272(byte(code), high, value)
				packet, allocation := guardedPlayerInputR272(record)
				before := bytes.Clone(allocation)
				out := slices.Clone(seed)
				got, err := netDecodePlayerInput(packet, out)
				want := append(slices.Clone(seed), event)
				if err != nil || !slices.Equal(got, want) || !bytes.Equal(allocation, before) {
					t.Fatalf("code=%d high=%x value=%08x: got=%v err=%v want=%v input_changed=%t", code, high, value, got, err, want, !bytes.Equal(allocation, before))
				}
			}
		}
	}
	for _, out := range [][]PlayerCtrl{nil, {}, seed} {
		got, err := netDecodePlayerInput(nil, out)
		if err != nil || len(got) != len(out) || cap(got) != cap(out) || (got == nil) != (out == nil) || !slices.Equal(got, out) {
			t.Fatalf("empty input: got=%v cap=%d err=%v, initial=%v cap=%d", got, cap(got), err, out, cap(out))
		}
	}
}

func TestNetworkPlayerInputDecoderTruncationAtomicR272(t *testing.T) {
	first, _ := playerInputRecordR272(6, [3]byte{1, 2, 3}, 0)
	second, _ := playerInputRecordR272(29, [3]byte{4, 5, 6}, 0x76543210)
	prefix := append(first, second...)
	for code := 0; code < 256; code++ {
		record, _ := playerInputRecordR272(byte(code), [3]byte{0xff, 0xa5, 0x5a}, 0x12345678)
		for cut := 1; cut < len(record); cut++ {
			for _, validPrefix := range [][]byte{nil, prefix} {
				wire := append(bytes.Clone(validPrefix), record[:cut]...)
				packet, allocation := guardedPlayerInputR272(wire)
				inputBefore := bytes.Clone(allocation)
				for _, withOut := range []bool{false, true} {
					var backing [12]PlayerCtrl
					for i := range backing {
						backing[i] = PlayerCtrl{Code: player.CtrlCode(90 + i), Data: 0x87650000 + uint32(i), Active: i%2 == 0}
					}
					backingBefore := backing
					var out []PlayerCtrl
					if withOut {
						out = backing[:2]
					}
					got, err := netDecodePlayerInput(packet, out)
					if err != io.ErrUnexpectedEOF || len(got) != len(out) || cap(got) != cap(out) || (got == nil) != (out == nil) || !slices.Equal(got, out) || backing != backingBefore || !bytes.Equal(allocation, inputBefore) {
						t.Fatalf("code=%d cut=%d/%d valid_prefix=%d with_out=%t: got=%v cap=%d err=%v original_cap=%d backing_changed=%t input_changed=%t", code, cut, len(record), len(validPrefix), withOut, got, cap(got), err, cap(out), backing != backingBefore, !bytes.Equal(allocation, inputBefore))
					}
					if withOut && &got[0] != &out[0] {
						t.Fatalf("code=%d cut=%d replaced original output on rejection", code, cut)
					}
				}
			}
		}
	}
}

func TestNetworkPlayerInputEnvelopeBoundsR272(t *testing.T) {
	for _, flags := range []uint32{0, 0x10} {
		s := new(Server)
		pl := &Player{PlayerInd: 3, Field3680: flags}
		queue := s.Players.Control.Player(pl.Index())
		queue.Append([]PlayerCtrl{{Code: player.CCJump, Active: true}})
		beforeQueue := *queue
		if n, err := s.netOnPlayerInput(pl, nil); n != 0 || err != io.ErrUnexpectedEOF || *queue != beforeQueue {
			t.Fatalf("flags=%x empty envelope: n=%d err=%v queue_changed=%t", flags, n, err, *queue != beforeQueue)
		}
		for declared := 1; declared < 256; declared++ {
			for present := 0; present < declared; present++ {
				wire := make([]byte, 1+present)
				wire[0] = byte(declared)
				packet, allocation := guardedPlayerInputR272(wire)
				beforeInput := bytes.Clone(allocation)
				n, err := s.netOnPlayerInput(pl, packet)
				if n != 0 || err != io.ErrUnexpectedEOF || *queue != beforeQueue || !bytes.Equal(allocation, beforeInput) {
					t.Fatalf("flags=%x declared=%d present=%d: n=%d err=%v queue_changed=%t input_changed=%t", flags, declared, present, n, err, *queue != beforeQueue, !bytes.Equal(allocation, beforeInput))
				}
			}
		}
	}
}

func TestNetworkPlayerInputInactiveGateR272(t *testing.T) {
	for _, flags := range []uint32{0, 1, 8, 0xffffffef} {
		s := new(Server)
		pl := &Player{PlayerInd: 3, Field3680: flags}
		queue := s.Players.Control.Player(pl.Index())
		queue.Append([]PlayerCtrl{{Code: player.CCJump, Active: true}})
		beforeQueue := *queue
		for declared := 0; declared < 256; declared++ {
			// Most lengths deliberately are not a whole number of records.
			// The closed acceptance gate consumes the bounded body untouched.
			wire := bytes.Repeat([]byte{0x1d}, 1+declared)
			wire[0] = byte(declared)
			wire = append(wire, byte(netmsg.MSG_KEEP_ALIVE))
			packet, allocation := guardedPlayerInputR272(wire)
			beforeInput := bytes.Clone(allocation)
			n, err := s.netOnPlayerInput(pl, packet)
			if err != nil || n != 1+declared || packet[n] != byte(netmsg.MSG_KEEP_ALIVE) || *queue != beforeQueue || !bytes.Equal(allocation, beforeInput) {
				t.Fatalf("flags=%x declared=%d: n=%d err=%v queue_changed=%t input_changed=%t", flags, declared, n, err, *queue != beforeQueue, !bytes.Equal(allocation, beforeInput))
			}
		}
	}
}

func TestNetworkPlayerInputAcceptedOrderAndTailR272(t *testing.T) {
	s := new(Server)
	pl := &Player{PlayerInd: 3, Field3680: 0x10}
	queue := s.Players.Control.Player(pl.Index())
	seed := PlayerCtrl{Code: player.CCJump, Data: 0x76543210, Active: true}
	queue.Append([]PlayerCtrl{seed})
	wantDecoded := []PlayerCtrl{seed}
	var body []byte
	for _, spec := range []struct {
		code  byte
		value uint32
	}{{6, 0}, {2, 0x11}, {4, 0x22}, {2, 0x33}, {29, 0xfedcba98}, {255, 0}} {
		record, event := playerInputRecordR272(spec.code, [3]byte{0xa5, 0x5a, 0xff}, spec.value)
		body = append(body, record...)
		wantDecoded = append(wantDecoded, event)
	}
	decoded, err := netDecodePlayerInput(body, []PlayerCtrl{seed})
	if err != nil || !slices.Equal(decoded, wantDecoded) {
		t.Fatalf("ordered decoder = %v err=%v want=%v", decoded, err, wantDecoded)
	}
	wire := append([]byte{byte(netmsg.MSG_PLAYER_INPUT), byte(len(body))}, body...)
	wire = append(wire, byte(netmsg.MSG_KEEP_ALIVE))
	packet, allocation := guardedPlayerInputR272(wire)
	beforeInput := bytes.Clone(allocation)
	n, handled, err := s.OnPacketOpSub(ntype.PlayerInd(pl.PlayerInd), netmsg.MSG_PLAYER_INPUT, packet, pl, nil)
	wantQueued := slices.Clone(wantDecoded)
	// Preserve ctrlBuf.Append's existing last-forward-input deduplication.
	wantQueued[2].Active = false
	if err != nil || !handled || n != 2+len(body) || queue.write != len(wantQueued) || queue.read != 0 || !slices.Equal(queue.events[:queue.write], wantQueued) || !bytes.Equal(allocation, beforeInput) {
		t.Fatalf("accepted packet n=%d handled=%t err=%v queue=%v want=%v input_changed=%t", n, handled, err, queue.events[:queue.write], wantQueued, !bytes.Equal(allocation, beforeInput))
	}
	if n2, handled2, err2 := s.OnPacketOpSub(ntype.PlayerInd(pl.PlayerInd), netmsg.Op(packet[n]), packet[n:], pl, nil); n2 != 1 || !handled2 || err2 != nil {
		t.Fatalf("following packet n=%d handled=%t err=%v", n2, handled2, err2)
	}
	beforeQueue := *queue
	if n, err := s.netOnPlayerInput(pl, []byte{0}); n != 1 || err != nil || *queue != beforeQueue {
		t.Fatalf("accepted zero-length body n=%d err=%v queue_changed=%t", n, err, *queue != beforeQueue)
	}
	// A complete declared body containing a valid record then an incomplete
	// record must not append even the first record to the existing queue.
	badBody := append(bytes.Clone(body), 0x1d, 0, 0)
	badPacket := append([]byte{byte(len(badBody))}, badBody...)
	if n, err := s.netOnPlayerInput(pl, badPacket); n != 0 || err != io.ErrUnexpectedEOF || *queue != beforeQueue {
		t.Fatalf("partial accepted record n=%d err=%v queue_changed=%t", n, err, *queue != beforeQueue)
	}
}

func TestNetworkPlayerInputDispatchBoundsR272(t *testing.T) {
	s := new(Server)
	pl := &Player{PlayerInd: 3, Field3680: 0x10}
	queue := s.Players.Control.Player(pl.Index())
	queue.Append([]PlayerCtrl{{Code: player.CCJump, Active: true}})
	beforeQueue := *queue
	for _, op := range []netmsg.Op{0x26, netmsg.MSG_PLAYER_INPUT} {
		for length := 0; length < 2; length++ {
			packet, allocation := guardedPlayerInputR272([]byte{byte(op)}[:length])
			beforeInput := bytes.Clone(allocation)
			n, handled, err := s.OnPacketOpSub(ntype.PlayerInd(pl.PlayerInd), op, packet, pl, nil)
			if n != 0 || handled || err != io.ErrUnexpectedEOF || *queue != beforeQueue || !bytes.Equal(allocation, beforeInput) {
				t.Fatalf("op=%x length=%d: n=%d handled=%t err=%v queue_changed=%t input_changed=%t", op, length, n, handled, err, *queue != beforeQueue, !bytes.Equal(allocation, beforeInput))
			}
		}
		wire := []byte{byte(op), 0, byte(netmsg.MSG_KEEP_ALIVE)}
		packet, allocation := guardedPlayerInputR272(wire)
		beforeInput := bytes.Clone(allocation)
		n, handled, err := s.OnPacketOpSub(ntype.PlayerInd(pl.PlayerInd), op, packet, pl, nil)
		if n != 2 || !handled || err != nil || *queue != beforeQueue || !bytes.Equal(allocation, beforeInput) {
			t.Fatalf("complete op=%x n=%d handled=%t err=%v queue_changed=%t input_changed=%t", op, n, handled, err, *queue != beforeQueue, !bytes.Equal(allocation, beforeInput))
		}
	}
}
