package legacy

import (
	"bytes"
	"fmt"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

type importantResendPacketSpec struct {
	recipient            int
	payload              []byte
	sequenceEnabled      bool
	sequence             uint16
	acknowledgedMask     uint32
	sentMask             uint32
	retryDelay           byte
	sendCount            byte
	removeIfDisconnected uint32
	relatedObject        *server.Object
}

type importantResendSendCall struct {
	playerIndex uint8
	messageKind int
	data        []byte
}

func setupImportantResendContract(
	t *testing.T, playerIndex uint8, specs []importantResendPacketSpec,
) (*[32]importantRateControl, []unsafe.Pointer, *[154]byte) {
	t.Helper()
	preserveImportantPacketState(t)

	records := memmap.PtrT[[32]importantRateControl](0x5D4594, 1565124)
	scratch := memmap.PtrT[[154]byte](0x5D4594, 1564964)
	scratchBefore := memmap.PtrUint8(0x5D4594, 1564963)
	scratchAfter := memmap.PtrUint8(0x5D4594, 1565118)
	oldRecords := *records
	oldScratch := *scratch
	oldScratchBefore := *scratchBefore
	oldScratchAfter := *scratchAfter
	oldFPS := gameFPSHook
	oldPlayerByInd := importantPlayerByIndHook
	oldGameHost := importantGameHostHook
	oldReplayRead := importantReplayReadHook
	oldRateGet := importantRateGetHook
	oldRateAdjust := importantRateAdjustHook
	oldNetSend := importantNetSendHook
	t.Cleanup(func() {
		*records = oldRecords
		*scratch = oldScratch
		*scratchBefore = oldScratchBefore
		*scratchAfter = oldScratchAfter
		gameFPSHook = oldFPS
		importantPlayerByIndHook = oldPlayerByInd
		importantGameHostHook = oldGameHost
		importantReplayReadHook = oldReplayRead
		importantRateGetHook = oldRateGet
		importantRateAdjustHook = oldRateAdjust
		importantNetSendHook = oldNetSend
	})

	for i := range records {
		records[i] = importantRateControl{
			ResendsPerUpdate: byte(0x20 + i),
			ResendInterval:   byte(0x40 + i),
			UpdateRate:       byte(0x60 + i),
			Reserved3:        byte(0x80 + i),
			Threshold:        0xA0000000 | uint32(i),
			LowerThreshold:   0xB0000000 | uint32(i),
		}
	}
	records[playerIndex] = importantRateControl{
		ResendsPerUpdate: 1,
		ResendInterval:   5,
		UpdateRate:       3,
		Reserved3:        0xA5,
		Threshold:        50,
		LowerThreshold:   40,
	}
	for i := range scratch {
		scratch[i] = byte(0x70 + i*13)
	}
	*scratchBefore = 0xC3
	*scratchAfter = 0xD4
	setImportantRecipientMaskC(^uint32(0))
	setImportantNativeListC(nil, nil)
	setImportantAllocClassC(nil)
	gameFPSHook = func() uint32 { return 30 }
	importantPlayerByIndHook = func(ntype.PlayerInd) *server.Player { return nil }
	importantGameHostHook = func() bool { return false }
	importantReplayReadHook = func() bool { return false }
	importantRateGetHook = func() uint32 { return 3 }
	importantRateAdjustHook = func(uint8) {}
	importantNetSendHook = func(uint8, int, []byte) bool { return true }

	packets := make([]unsafe.Pointer, len(specs))
	if len(specs) == 0 {
		return records, packets, scratch
	}
	pool := alloc.NewClass("important-resend-contract", importantPacketSizeC(), len(specs))
	setImportantAllocClassC(pool.UPtr())
	for i, spec := range specs {
		packets[i] = pool.NewObject()
		if packets[i] == nil {
			t.Fatalf("packet %d allocation failed", i)
		}
		packet := (*importantPacketLegacy)(packets[i])
		packet.Recipient = int8(uint8(spec.recipient))
		packet.PayloadSize = byte(len(spec.payload))
		copy(packet.Payload[:], spec.payload)
		packet.SequenceEnabled = byte(bool2int(spec.sequenceEnabled))
		packet.Sequence[playerIndex] = spec.sequence
		packet.AcknowledgedMask = spec.acknowledgedMask
		packet.SentMask = spec.sentMask
		packet.RetryDelay[playerIndex] = spec.retryDelay
		packet.SendCount = spec.sendCount
		packet.RemoveIfDisconnected = spec.removeIfDisconnected
		if spec.relatedObject != nil {
			setImportantPacketRelatedObjectC(packets[i], spec.relatedObject.CObj())
		}
	}
	for i, packet := range packets {
		if i > 0 {
			setImportantPacketPrevC(packet, packets[i-1])
		}
		if i+1 < len(packets) {
			setImportantPacketNextC(packet, packets[i+1])
		}
	}
	setImportantNativeListC(packets[0], packets[len(packets)-1])
	return records, packets, scratch
}

func assertImportantResendList(t *testing.T, packets []unsafe.Pointer) {
	t.Helper()
	first, last := importantNativeListC()
	if len(packets) == 0 {
		if first != nil || last != nil {
			t.Fatalf("empty list ends: got (%p, %p), want (nil, nil)", first, last)
		}
		return
	}
	if first != packets[0] || last != packets[len(packets)-1] {
		t.Fatalf("list ends: got (%p, %p), want (%p, %p)", first, last, packets[0], packets[len(packets)-1])
	}
	for i, packet := range packets {
		var wantPrev, wantNext unsafe.Pointer
		if i > 0 {
			wantPrev = packets[i-1]
		}
		if i+1 < len(packets) {
			wantNext = packets[i+1]
		}
		if got := importantPacketPrevC(packet); got != wantPrev {
			t.Errorf("packet %d prev: got %p, want %p", i, got, wantPrev)
		}
		if got := importantPacketNextC(packet); got != wantNext {
			t.Errorf("packet %d next: got %p, want %p", i, got, wantNext)
		}
	}
}

func TestImportantResendMatchesGAMEEXEContract(t *testing.T) {
	handles.Init()
	t.Cleanup(handles.Release)

	t.Run("player-and-host-gate", func(t *testing.T) {
		for _, tc := range []struct {
			name       string
			player     *server.Player
			host       bool
			wantEvents string
		}{
			{name: "present-host-active-is-blocked", player: &server.Player{}, host: true, wantEvents: "[lookup host]"},
			{name: "missing-player-processes", wantEvents: "[lookup send:aa]"},
			{name: "non-host-processes", player: &server.Player{}, wantEvents: "[lookup host send:aa]"},
			{name: "status-bit-processes", player: &server.Player{Field3680: 0x10}, host: true, wantEvents: "[lookup host send:aa]"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				_, packets, _ := setupImportantResendContract(t, 7, []importantResendPacketSpec{{recipient: 7, payload: []byte{0x40}}})
				before := *(*importantPacketLegacy)(packets[0])
				var events []string
				importantPlayerByIndHook = func(ind ntype.PlayerInd) *server.Player {
					events = append(events, "lookup")
					if ind != 7 {
						t.Errorf("lookup index: got %d, want 7", ind)
					}
					return tc.player
				}
				importantGameHostHook = func() bool {
					events = append(events, "host")
					return tc.host
				}
				importantNetSendHook = func(_ uint8, _ int, data []byte) bool {
					events = append(events, fmt.Sprintf("send:%x", data))
					return false
				}

				sendQueuedImportantC(7, 1)
				if got := fmt.Sprint(events); got != tc.wantEvents {
					t.Errorf("events: got %s, want %s", got, tc.wantEvents)
				}
				if got := *(*importantPacketLegacy)(packets[0]); got != before {
					t.Error("gate/header failure changed the packet")
				}
				assertImportantResendList(t, packets)
			})
		}
	})

	t.Run("tail-to-head-routing-and-sequence-wire-format", func(t *testing.T) {
		const (
			playerIndex = uint8(7)
			playerMask  = uint32(1) << playerIndex
		)
		records, packets, scratch := setupImportantResendContract(t, playerIndex, []importantResendPacketSpec{
			{recipient: 1, payload: []byte{0x10}},
			{recipient: -1, payload: []byte{0x11}, acknowledgedMask: playerMask},
			{recipient: -1, payload: []byte{0x41, 0xA1}},
			{recipient: 0x82, payload: []byte{0x42, 0xB2, 0xC2}, sequenceEnabled: true, sequence: 0x1234},
			{recipient: 7, payload: []byte{0x43, 0xD3}},
		})
		beforeRecords := *records
		beforeScratch := *scratch
		beforePackets := make([]importantPacketLegacy, len(packets))
		for i, packet := range packets {
			beforePackets[i] = *(*importantPacketLegacy)(packet)
		}
		var events []string
		var calls []importantResendSendCall
		player := &server.Player{}
		importantPlayerByIndHook = func(ind ntype.PlayerInd) *server.Player {
			events = append(events, fmt.Sprintf("lookup:%d", ind))
			return player
		}
		hostCalls := 0
		importantGameHostHook = func() bool {
			hostCalls++
			events = append(events, fmt.Sprintf("host:%d", hostCalls))
			return false
		}
		importantNetSendHook = func(index uint8, kind int, data []byte) bool {
			copyData := append([]byte(nil), data...)
			calls = append(calls, importantResendSendCall{playerIndex: index, messageKind: kind, data: copyData})
			events = append(events, fmt.Sprintf("send:%x", data))
			return true
		}
		importantRateGetHook = func() uint32 {
			events = append(events, "rate")
			return 3
		}
		gameFPSHook = func() uint32 {
			events = append(events, "fps")
			return 30
		}
		frames := []uint32{0x10203040, 0x10203041, 0x10203042}
		gameFrameHook = func() uint32 {
			value := frames[0]
			frames = frames[1:]
			events = append(events, fmt.Sprintf("frame:%x", value))
			return value
		}
		importantReplayReadHook = func() bool {
			events = append(events, "replay")
			return false
		}

		sendQueuedImportantC(playerIndex, 1)
		wantCalls := []importantResendSendCall{
			{playerIndex: 7, messageKind: 1, data: []byte{0xAA}},
			{playerIndex: 7, messageKind: 1, data: []byte{0x43, 0xD3}},
			{playerIndex: 7, messageKind: 1, data: []byte{0xCC, 0x34, 0x12, 0x03, 0x42, 0xB2, 0xC2}},
			{playerIndex: 7, messageKind: 1, data: []byte{0x41, 0xA1}},
		}
		if got, want := fmt.Sprint(calls), fmt.Sprint(wantCalls); got != want {
			t.Errorf("send calls: got %s, want %s", got, want)
		}
		wantEvents := "[lookup:7 host:1 send:aa send:43d3 rate fps frame:10203040 replay send:cc34120342b2c2 rate fps frame:10203041 replay send:41a1 rate fps frame:10203042 replay host:2]"
		if got := fmt.Sprint(events); got != wantEvents {
			t.Errorf("events: got %s, want %s", got, wantEvents)
		}
		wantPackets := beforePackets
		for i := 2; i < len(wantPackets); i++ {
			wantPackets[i].SentMask |= playerMask
			wantPackets[i].RetryDelay[playerIndex] = 50
			wantPackets[i].LastSendFrame[playerIndex] = uint32(0x10203042 - (i - 2))
		}
		for i, packetPtr := range packets {
			if got := *(*importantPacketLegacy)(packetPtr); got != wantPackets[i] {
				t.Errorf("packet %d changed outside the original send-state contract", i)
			}
		}
		if got := *records; got != beforeRecords {
			t.Error("transmission changed rate-control records")
		}
		wantScratch := beforeScratch
		copy(wantScratch[:], []byte{0xCC, 0x34, 0x12, 0x03, 0x42, 0xB2, 0xC2})
		if got := *scratch; got != wantScratch {
			t.Error("sequence scratch buffer does not match the original little-endian wire layout")
		}
		if *memmap.PtrUint8(0x5D4594, 1564963) != 0xC3 || *memmap.PtrUint8(0x5D4594, 1565118) != 0xD4 {
			t.Error("sequence scratch guards changed")
		}
		assertImportantResendList(t, packets)
	})

	t.Run("frame-header-format", func(t *testing.T) {
		for _, tc := range []struct {
			name        string
			playerIndex uint8
			messageKind int
		}{
			{name: "zero-kind-player", playerIndex: 7},
			{name: "reserved-index-always-includes-frame", playerIndex: 31, messageKind: 1},
		} {
			t.Run(tc.name, func(t *testing.T) {
				_, packets, _ := setupImportantResendContract(t, tc.playerIndex, []importantResendPacketSpec{{recipient: int(tc.playerIndex), payload: []byte{0x40}}})
				before := *(*importantPacketLegacy)(packets[0])
				var calls [][]byte
				gameFrameHook = func() uint32 { return 0x89ABCDEF }
				importantNetSendHook = func(index uint8, kind int, data []byte) bool {
					if index != tc.playerIndex || kind != tc.messageKind {
						t.Errorf("route: got (%d, %d), want (%d, %d)", index, kind, tc.playerIndex, tc.messageKind)
					}
					calls = append(calls, append([]byte(nil), data...))
					return len(calls) == 1
				}
				sendQueuedImportantC(tc.playerIndex, tc.messageKind)
				want := [][]byte{{0xAA, 0xEF, 0xCD, 0xAB, 0x89}, {0x40}}
				if got, wantText := fmt.Sprint(calls), fmt.Sprint(want); got != wantText {
					t.Errorf("calls: got %s, want %s", got, wantText)
				}
				if got := *(*importantPacketLegacy)(packets[0]); got != before {
					t.Error("failed payload send changed the packet")
				}
			})
		}
	})

	t.Run("failed-header-keeps-resend-preparation", func(t *testing.T) {
		const (
			playerIndex = uint8(7)
			playerMask  = uint32(1) << playerIndex
		)
		_, packets, _ := setupImportantResendContract(t, playerIndex, []importantResendPacketSpec{{
			recipient: -1, payload: []byte{0x50}, sentMask: playerMask, sendCount: 0xFF,
		}})
		before := *(*importantPacketLegacy)(packets[0])
		var calls [][]byte
		importantNetSendHook = func(_ uint8, _ int, data []byte) bool {
			calls = append(calls, append([]byte(nil), data...))
			return false
		}

		sendQueuedImportantC(playerIndex, 1)
		if got, want := fmt.Sprint(calls), "[[170]]"; got != want {
			t.Errorf("calls: got %s, want %s", got, want)
		}
		want := before
		want.SentMask &^= playerMask
		want.SendCount = 0
		if got := *(*importantPacketLegacy)(packets[0]); got != want {
			t.Error("header failure did not preserve the original resend-preparation mutations")
		}
		assertImportantResendList(t, packets)
	})

	t.Run("retry-delay-resend-cap-and-byte-wrap", func(t *testing.T) {
		const (
			playerIndex = uint8(7)
			playerMask  = uint32(1) << playerIndex
		)
		records, packets, _ := setupImportantResendContract(t, playerIndex, []importantResendPacketSpec{
			{recipient: -1, payload: []byte{0x51}},
			{recipient: -1, payload: []byte{0x52}, sentMask: playerMask},
			{recipient: -1, payload: []byte{0x53}, sentMask: playerMask, sendCount: 0xFF},
			{recipient: -1, payload: []byte{0x54}, sentMask: playerMask, retryDelay: 2},
		})
		records[playerIndex].ResendsPerUpdate = 1
		beforePackets := make([]importantPacketLegacy, len(packets))
		for i, packet := range packets {
			beforePackets[i] = *(*importantPacketLegacy)(packet)
		}
		var calls [][]byte
		importantNetSendHook = func(_ uint8, _ int, data []byte) bool {
			calls = append(calls, append([]byte(nil), data...))
			return len(calls) != 2
		}
		gameFrameHook = func() uint32 { return 0x55667788 }

		sendQueuedImportantC(playerIndex, 1)
		wantCalls := [][]byte{{0xAA}, {0x53}, {0x51}}
		if got, want := fmt.Sprint(calls), fmt.Sprint(wantCalls); got != want {
			t.Errorf("calls: got %s, want %s", got, want)
		}
		wantPackets := beforePackets
		wantPackets[3].RetryDelay[playerIndex] = 1
		wantPackets[2].SendCount = 0
		wantPackets[2].SentMask = 0
		wantPackets[0].SentMask = playerMask
		wantPackets[0].RetryDelay[playerIndex] = 50
		wantPackets[0].LastSendFrame[playerIndex] = 0x55667788
		for i, packet := range packets {
			if got := *(*importantPacketLegacy)(packet); got != wantPackets[i] {
				t.Errorf("packet %d changed outside the retry/resend contract", i)
			}
		}
		assertImportantResendList(t, packets)
	})

	t.Run("retry-delay-uses-wrapped-uint32-product", func(t *testing.T) {
		const playerIndex = uint8(7)
		records, packets, _ := setupImportantResendContract(t, playerIndex, []importantResendPacketSpec{{
			recipient: 7, payload: []byte{0x55},
		}})
		records[playerIndex].ResendInterval = 3
		importantRateGetHook = func() uint32 { return 1 }
		gameFPSHook = func() uint32 { return 0x80000001 }
		gameFrameHook = func() uint32 { return 0x11223344 }
		sendQueuedImportantC(playerIndex, 1)
		packet := (*importantPacketLegacy)(packets[0])
		if packet.RetryDelay[playerIndex] != 3 {
			t.Errorf("retry delay: got %#x, want low byte 3 after uint32 multiplication wrap", packet.RetryDelay[playerIndex])
		}
	})

	t.Run("related-object-and-disconnect-termination", func(t *testing.T) {
		t.Run("destroyed-related-object-is-cleared-before-routing", func(t *testing.T) {
			related, free := alloc.New(server.Object{})
			t.Cleanup(free)
			related.ObjFlags = 0x20
			_, packets, _ := setupImportantResendContract(t, 7, []importantResendPacketSpec{{
				recipient: 7, payload: []byte{0x60}, relatedObject: related,
			}})
			importantNetSendHook = func(_ uint8, _ int, _ []byte) bool { return false }
			sendQueuedImportantC(7, 1)
			if got := importantPacketRelatedObjectC(packets[0]); got != nil {
				t.Errorf("related object: got %p, want nil", got)
			}
			assertImportantResendList(t, packets)
		})

		t.Run("missing-related-sync-bit-acknowledges-and-returns", func(t *testing.T) {
			related, free := alloc.New(server.Object{})
			t.Cleanup(free)
			related.Field37 = 0
			_, packets, _ := setupImportantResendContract(t, 7, []importantResendPacketSpec{
				{recipient: -1, payload: []byte{0x61}},
				{recipient: 7, payload: []byte{0x62}, relatedObject: related},
			})
			importantNetSendHook = func(_ uint8, _ int, _ []byte) bool {
				t.Fatal("invalid related-object packet attempted a send")
				return false
			}
			sendQueuedImportantC(7, 1)
			assertImportantResendList(t, packets[:1])
		})

		t.Run("disconnected-recipient-acknowledges-and-returns", func(t *testing.T) {
			_, packets, _ := setupImportantResendContract(t, 7, []importantResendPacketSpec{
				{recipient: -1, payload: []byte{0x63}},
				{recipient: 7, payload: []byte{0x64}, removeIfDisconnected: 1},
			})
			setImportantRecipientMaskC(0)
			sendQueuedImportantC(7, 1)
			assertImportantResendList(t, packets[:1])
		})
	})

	t.Run("replay-read-acknowledges-after-success", func(t *testing.T) {
		_, _, _ = setupImportantResendContract(t, 7, []importantResendPacketSpec{{recipient: 7, payload: []byte{0x70}}})
		var events []string
		importantNetSendHook = func(_ uint8, _ int, data []byte) bool {
			events = append(events, fmt.Sprintf("send:%x", data))
			return true
		}
		importantRateGetHook = func() uint32 {
			events = append(events, "rate")
			return 3
		}
		gameFPSHook = func() uint32 {
			events = append(events, "fps")
			return 30
		}
		gameFrameHook = func() uint32 {
			events = append(events, "frame")
			return 0x778899AA
		}
		importantReplayReadHook = func() bool {
			events = append(events, "replay")
			return true
		}
		sendQueuedImportantC(7, 1)
		if got, want := fmt.Sprint(events), "[send:aa send:70 rate fps frame replay]"; got != want {
			t.Errorf("events: got %s, want %s", got, want)
		}
		assertImportantResendList(t, nil)
	})

	t.Run("periodic-rate-adjustment", func(t *testing.T) {
		for _, tc := range []struct {
			name       string
			frame      uint32
			host       bool
			wantEvents string
		}{
			{name: "period-boundary", frame: 120, host: true, wantEvents: "[lookup host frame fps adjust:7]"},
			{name: "non-boundary", frame: 121, host: true, wantEvents: "[lookup host frame fps]"},
			{name: "non-host", frame: 120, wantEvents: "[lookup host]"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				records, _, _ := setupImportantResendContract(t, 7, nil)
				records[7].ResendInterval = 3
				var events []string
				importantPlayerByIndHook = func(ntype.PlayerInd) *server.Player {
					events = append(events, "lookup")
					return nil
				}
				importantGameHostHook = func() bool {
					events = append(events, "host")
					return tc.host
				}
				gameFrameHook = func() uint32 {
					events = append(events, "frame")
					return tc.frame
				}
				gameFPSHook = func() uint32 {
					events = append(events, "fps")
					return 10
				}
				importantRateAdjustHook = func(ind uint8) {
					events = append(events, fmt.Sprintf("adjust:%d", ind))
				}
				sendQueuedImportantC(7, 1)
				if got := fmt.Sprint(events); got != tc.wantEvents {
					t.Errorf("events: got %s, want %s", got, tc.wantEvents)
				}
			})
		}
	})

	t.Run("sequence-scratch-tail-is-preserved", func(t *testing.T) {
		_, _, scratch := setupImportantResendContract(t, 7, []importantResendPacketSpec{{
			recipient: 7, payload: []byte{1, 2, 3, 4, 5}, sequenceEnabled: true, sequence: 0xBEEF,
		}})
		before := append([]byte(nil), scratch[:]...)
		importantNetSendHook = func(_ uint8, _ int, data []byte) bool {
			return !bytes.Equal(data, []byte{0xCC, 0xEF, 0xBE, 5, 1, 2, 3, 4, 5})
		}
		sendQueuedImportantC(7, 1)
		want := append([]byte(nil), before...)
		copy(want, []byte{0xCC, 0xEF, 0xBE, 5, 1, 2, 3, 4, 5})
		if !bytes.Equal(scratch[:], want) {
			t.Error("sequence scratch tail changed outside the encoded message")
		}
	})
}
