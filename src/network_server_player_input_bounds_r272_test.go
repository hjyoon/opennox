package opennox

import (
	"bytes"
	"encoding/binary"
	"image"
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/player"
	"github.com/opennox/libs/strman"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// This fixture reaches the actual receive dispatcher, without opening a
// socket, launching a game, replacing a handler, or supplying a reply.
func TestServerRawPlayerInputBoundsR272(t *testing.T) {
	base := server.New(nil, nil, strman.New())
	t.Cleanup(base.Close)
	s := &Server{Server: base}
	const index = ntype.PlayerInd(3)
	pl := s.Players.ResetInd(index)
	unit, freeUnit := alloc.New(server.Object{})
	t.Cleanup(freeUnit)
	update, freeUpdate := alloc.New(server.PlayerUpdateData{})
	t.Cleanup(freeUpdate)
	*unit = server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	*update = server.PlayerUpdateData{Player: pl}
	pl.PlayerUnit = unit
	if pl.PlayerUnit != unit || unit.UpdateDataPlayer() != update || update.Player != pl || unit.ObjClass != object.ClassPlayer {
		t.Fatal("native player/object/update links were not initialized")
	}
	t.Cleanup(func() {
		pl.PlayerUnit = nil
		unit.UpdateData = nil
		update.Player = nil
	})
	pl.Field3680 = 0x10
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(pl)) <= math.MaxUint32 ||
		uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 || uintptr(unsafe.Pointer(update)) <= math.MaxUint32) {
		t.Fatalf("receive fixture did not exercise native high pointers: player=%p unit=%p update=%p", pl, unit, update)
	}
	// No replay writer is configured by a parser-only fixture. Preserve every
	// engine flag and restore the replay flag on exit if a caller had set it.
	replayWrite := noxflags.HasEngine(noxflags.EngineReplayWrite)
	noxflags.UnsetEngine(noxflags.EngineReplayWrite)
	t.Cleanup(func() {
		if replayWrite {
			noxflags.SetEngine(noxflags.EngineReplayWrite)
		}
	})

	queue := s.Players.Control.Player(int(index))
	seed := server.PlayerCtrl{Code: player.CCJump, Data: 0x76543210, Active: true}
	origin := image.Pt(11, 22)
	destination := image.Pt(0x1234, 0xabcd)
	// A real following MSG_MOUSE proves that successful parsing consumed only
	// the declared input body. KEEP_ALIVE then exercises another raw-loop step.
	tail := []byte{byte(netmsg.MSG_MOUSE), 0x34, 0x12, 0xcd, 0xab, byte(netmsg.MSG_KEEP_ALIVE)}
	reset := func() {
		queue.Reset()
		queue.Append([]server.PlayerCtrl{seed})
		pl.SetCursorPos(origin)
		pl.Frame3596 = 7
		s.SetFrame(12345)
	}
	receive := func(t *testing.T, packet []byte) (valid bool) {
		t.Helper()
		defer func() {
			if fault := recover(); fault != nil {
				t.Fatalf("actual onPacketRaw panicked for %x: %v", packet, fault)
			}
		}()
		return s.onPacketRaw(index, packet)
	}
	guard := func(wire []byte) ([]byte, []byte) {
		allocation := bytes.Repeat([]byte{0xa5}, 16+len(wire)+512)
		copy(allocation[16:], wire)
		return allocation[16 : 16+len(wire)], allocation
	}
	// State the wire widths independently of the decoder's DataSize method.
	record := func(code byte, value uint32) ([]byte, server.PlayerCtrl) {
		width := 0
		switch code {
		case 1, 2, 3, 4, 5, 28:
			width = 1
		case 29, 30:
			width = 4
		}
		wire := make([]byte, 4+width)
		copy(wire, []byte{code, 0xa5, 0x5a, 0xff})
		var payload [4]byte
		binary.LittleEndian.PutUint32(payload[:], value)
		copy(wire[4:], payload[:width])
		if width == 0 {
			value = 0
		} else if width == 1 {
			value &= 0xff
		}
		return wire, server.PlayerCtrl{Code: player.CtrlCode(code), Data: value, Active: true}
	}
	reject := func(t *testing.T, wire []byte) {
		t.Helper()
		reset()
		packet, allocation := guard(wire)
		beforeInput := bytes.Clone(allocation)
		beforeQueue := *queue
		if receive(t, packet) || *queue != beforeQueue || pl.Frame3596 != 7 || pl.CursorVec != origin || !bytes.Equal(allocation, beforeInput) {
			t.Fatalf("malformed raw packet %x: queue_changed=%t frame=%d cursor=%v input_changed=%t", wire, *queue != beforeQueue, pl.Frame3596, pl.CursorVec, !bytes.Equal(allocation, beforeInput))
		}
	}

	t.Run("truncated envelope", func(t *testing.T) {
		reject(t, []byte{byte(netmsg.MSG_PLAYER_INPUT)})
		for declared := 1; declared <= 8; declared++ {
			for actual := 0; actual < declared; actual++ {
				wire := append([]byte{byte(netmsg.MSG_PLAYER_INPUT), byte(declared)}, bytes.Repeat([]byte{0xa5}, actual)...)
				reject(t, wire)
			}
		}
	})

	t.Run("partial record is atomic", func(t *testing.T) {
		prefix, _ := record(6, 0)
		for code := 0; code < 256; code++ {
			whole, _ := record(byte(code), 0x89abcdef)
			for cut := 1; cut < len(whole); cut++ {
				for _, validPrefix := range [][]byte{nil, prefix} {
					body := append(bytes.Clone(validPrefix), whole[:cut]...)
					wire := append([]byte{byte(netmsg.MSG_PLAYER_INPUT), byte(len(body))}, body...)
					wire = append(wire, tail...)
					reject(t, wire)
				}
			}
		}
	})

	t.Run("valid input and following packets", func(t *testing.T) {
		reset()
		var body []byte
		want := []server.PlayerCtrl{seed}
		for _, spec := range []struct {
			code  byte
			value uint32
		}{{255, 0}, {6, 0}, {2, 0x11}, {29, 0x89abcdef}, {2, 0x22}} {
			wire, event := record(spec.code, spec.value)
			body = append(body, wire...)
			if spec.code != 2 || spec.value != 0x11 {
				want = append(want, event)
			}
		}
		wire := append([]byte{byte(netmsg.MSG_PLAYER_INPUT), byte(len(body))}, body...)
		wire = append(wire, tail...)
		packet, allocation := guard(wire)
		beforeInput := bytes.Clone(allocation)
		if !receive(t, packet) || pl.Frame3596 != s.Frame() || pl.CursorVec != destination || !bytes.Equal(allocation, beforeInput) {
			t.Fatalf("valid raw dispatch: frame=%d cursor=%v input_changed=%t", pl.Frame3596, pl.CursorVec, !bytes.Equal(allocation, beforeInput))
		}
		var got []server.PlayerCtrl
		for event := queue.First(); event != nil; event = queue.Next() {
			got = append(got, *event)
		}
		if !slices.Equal(got, want) {
			t.Fatalf("actual queued control order/dedup = %v, want %v", got, want)
		}
		reset()
		beforeQueue := *queue
		if !receive(t, []byte{byte(netmsg.MSG_PLAYER_INPUT), 0, byte(netmsg.MSG_KEEP_ALIVE)}) || *queue != beforeQueue || pl.Frame3596 != s.Frame() || pl.CursorVec != origin {
			t.Fatalf("empty declared body changed the queue/cursor or failed to advance the heartbeat")
		}
	})

	t.Run("inactive player bounded body skip", func(t *testing.T) {
		defer func() { pl.Field3680 = 0x10 }()
		for _, flags := range []uint32{0, 1, 0xffffffef} {
			pl.Field3680 = flags
			for declared := 0; declared <= 255; declared++ {
				reset()
				beforeQueue := *queue
				wire := append([]byte{byte(netmsg.MSG_PLAYER_INPUT), byte(declared)}, bytes.Repeat([]byte{0x1d}, declared)...)
				wire = append(wire, tail...)
				packet, allocation := guard(wire)
				beforeInput := bytes.Clone(allocation)
				if !receive(t, packet) || *queue != beforeQueue || pl.Frame3596 != s.Frame() || pl.CursorVec != destination || !bytes.Equal(allocation, beforeInput) {
					t.Fatalf("inactive flags=%x declared=%d: queue_changed=%t frame=%d cursor=%v input_changed=%t", flags, declared, *queue != beforeQueue, pl.Frame3596, pl.CursorVec, !bytes.Equal(allocation, beforeInput))
				}
			}
		}
	})

	t.Run("auxiliary opcode exact boundary", func(t *testing.T) {
		reject(t, []byte{0x26})
		reset()
		beforeQueue := *queue
		wire := append([]byte{0x26, 0xff}, tail...)
		packet, allocation := guard(wire)
		beforeInput := bytes.Clone(allocation)
		if !receive(t, packet) || *queue != beforeQueue || pl.Frame3596 != s.Frame() || pl.CursorVec != destination || !bytes.Equal(allocation, beforeInput) {
			t.Fatalf("opcode 0x26 did not preserve its two-byte boundary")
		}
	})
}
