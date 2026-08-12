package legacy

import (
	"bytes"
	"testing"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

type netClientSendTestServer struct {
	Server
	srv *server.Server
}

func (s *netClientSendTestServer) S() *server.Server {
	return s.srv
}

func TestImportantPacketUnsequencedWrapperMatchesGAMEEXEContract(t *testing.T) {
	handles.Init()
	t.Cleanup(handles.Release)
	preserveImportantPacketState(t)
	setImportantNativeListC(nil, nil)

	pool := alloc.NewClass("important-unsequenced-wrapper-contract", importantPacketSizeC(), 1)
	setImportantAllocClassC(pool.UPtr())
	setImportantRecipientMaskC(uint32(1) << 6)
	counters := memmap.PtrT[[32]uint16](0x5D4594, 1565524)
	for i := range counters {
		counters[i] = uint16(0x2200 + i)
	}
	wantCounters := *counters
	gameFrameHook = func() uint32 { return 0x56789ABC }
	relatedObject, freeRelatedObject := alloc.Malloc(1)
	t.Cleanup(freeRelatedObject)
	payload := []byte{0x44, 0x54, 0x20, 0xC3, 0x3C, 0x7E}

	if got := sendImportantPacketWrapperC(6, payload, relatedObject, 0x50607080, importantPacketSequenceDisabled); got != 1 {
		t.Fatalf("wrapper return value: got %d, want 1", got)
	}
	first, last := importantNativeListC()
	if first == nil || first != last {
		t.Fatalf("wrapper list: got (%p, %p), want one packet", first, last)
	}
	packet := (*importantPacketLegacy)(first)
	if packet.CreatedFrame != 0x56789ABC || packet.Recipient != 6 {
		t.Errorf("wrapper routing: frame=%#x recipient=%d", packet.CreatedFrame, packet.Recipient)
	}
	if packet.PayloadSize != byte(len(payload)) || !bytes.Equal(packet.Payload[:len(payload)], payload) {
		t.Errorf("wrapper payload: size=%d data=%x, want size=%d data=%x", packet.PayloadSize, packet.Payload[:len(payload)], len(payload), payload)
	}
	if packet.RemoveIfDisconnected != 0x50607080 {
		t.Errorf("wrapper disconnect policy: got %#x, want %#x", packet.RemoveIfDisconnected, uint32(0x50607080))
	}
	if packet.SequenceEnabled != 0 || packet.Sequence != [32]uint16{} {
		t.Errorf("wrapper sequence state: enabled=%d values=%#v, want disabled and zero", packet.SequenceEnabled, packet.Sequence)
	}
	if got := *counters; got != wantCounters {
		t.Errorf("wrapper counters: got %#v, want %#v", got, wantCounters)
	}
	if got := importantPacketRelatedObjectC(first); got != relatedObject {
		t.Errorf("wrapper related object: got %p, want %p", got, relatedObject)
	}
}

func TestNetClientSend2MatchesGAMEEXEContract(t *testing.T) {
	handles.Init()
	t.Cleanup(handles.Release)
	preserveImportantPacketState(t)
	setImportantNativeListC(nil, nil)
	setImportantAllocClassC(nil)

	oldFlags := noxflags.GetGame()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})
	list := netlist.New()
	list.Init()
	t.Cleanup(list.Free)
	oldGetServer := GetServer
	GetServer = func() Server {
		return &netClientSendTestServer{srv: &server.Server{NetList: list}}
	}
	t.Cleanup(func() { GetServer = oldGetServer })

	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameHost)
	hostPayload := []byte{0x71, 0x72, 0x73}
	if got := Nox_xxx_netClientSend2_4E53C0(31, hostPayload, nil, 0x10203040); got != 1 {
		t.Fatalf("host-mode return value: got %d, want 1", got)
	}
	if got := list.CopyPacketsA(31, netlist.Kind0); !bytes.Equal(got, hostPayload) {
		t.Errorf("host-mode queue: got %x, want %x", got, hostPayload)
	}
	if got := Nox_xxx_netClientSend2_4E53C0(31, make([]byte, 2049), nil, 1); got != 1 {
		t.Errorf("host-mode full-queue return value: got %d, want 1", got)
	}
	if first, last := importantNativeListC(); first != nil || last != nil {
		t.Fatalf("host-mode important list: got (%p, %p), want empty", first, last)
	}
	if importantAllocClassC() != nil {
		t.Fatal("host mode allocated an important-packet class")
	}

	noxflags.ResetGame()
	for _, recipient := range []int{255, 0x80, 0x180, -1} {
		if got := Nox_xxx_netClientSend2_4E53C0(recipient, []byte{0x74}, nil, 1); got != 0 {
			t.Errorf("non-host rejected recipient %#x: got %d, want 0", recipient, got)
		}
	}
	if importantAllocClassC() != nil {
		t.Fatal("rejected non-host recipients allocated an important-packet class")
	}

	pool := alloc.NewClass("important-client-send-wrapper-contract", importantPacketSizeC(), 1)
	setImportantAllocClassC(pool.UPtr())
	setImportantRecipientMaskC(uint32(1) << 3)
	counters := memmap.PtrT[[32]uint16](0x5D4594, 1565524)
	for i := range counters {
		counters[i] = uint16(0x3300 + i)
	}
	wantCounters := *counters
	gameFrameHook = func() uint32 { return 0x89ABCDEF }
	nonHostPayload := []byte{0x75, 0x76, 0x77, 0x78}
	relatedObject, freeRelatedObject := alloc.Malloc(1)
	t.Cleanup(freeRelatedObject)
	if got := Nox_xxx_netClientSend2_4E53C0(3, nonHostPayload, AsObjectP(relatedObject), 0x60708090); got != 1 {
		t.Fatalf("non-host forwarded return value: got %d, want 1", got)
	}
	first, last := importantNativeListC()
	if first == nil || first != last {
		t.Fatalf("non-host forwarded list: got (%p, %p), want one packet", first, last)
	}
	packet := (*importantPacketLegacy)(first)
	if packet.CreatedFrame != 0x89ABCDEF || packet.Recipient != 3 || packet.RemoveIfDisconnected != 0x60708090 {
		t.Errorf("non-host forwarded fields: frame=%#x recipient=%d policy=%#x", packet.CreatedFrame, packet.Recipient, packet.RemoveIfDisconnected)
	}
	if packet.PayloadSize != byte(len(nonHostPayload)) || !bytes.Equal(packet.Payload[:len(nonHostPayload)], nonHostPayload) {
		t.Errorf("non-host forwarded payload: size=%d data=%x, want size=%d data=%x", packet.PayloadSize, packet.Payload[:len(nonHostPayload)], len(nonHostPayload), nonHostPayload)
	}
	if packet.SequenceEnabled != 0 || packet.Sequence != [32]uint16{} || *counters != wantCounters {
		t.Errorf("non-host sequence state: enabled=%d values=%#v counters=%#v", packet.SequenceEnabled, packet.Sequence, *counters)
	}
	if got := importantPacketRelatedObjectC(first); got != relatedObject {
		t.Errorf("non-host related object: got %p, want %p", got, relatedObject)
	}
}

func TestImportantPacketReplacementWrapperMatchesGAMEEXEContract(t *testing.T) {
	handles.Init()
	t.Cleanup(handles.Release)

	t.Run("direct-recipient-acknowledgement", func(t *testing.T) {
		preserveImportantPacketState(t)
		setImportantNativeListC(nil, nil)
		pool := alloc.NewClass("important-replacement-direct-contract", importantPacketSizeC(), 6)
		setImportantAllocClassC(pool.UPtr())
		setImportantRecipientMaskC((uint32(1) << 1) | (uint32(1) << 2) | (uint32(1) << 3))
		frame := uint32(0x41000000)
		gameFrameHook = func() uint32 { return frame }
		relatedObject, freeRelatedObject := alloc.New(server.Object{})
		t.Cleanup(freeRelatedObject)
		relatedObject.Field37 = (uint32(1) << 2) | (uint32(1) << 6)
		sendOld := func(recipient int, payload []byte, related *server.Object) {
			t.Helper()
			if got := sendImportantPacketC(recipient, payload, related.CObj(), 0x11111111, false); got != 1 {
				t.Fatalf("old packet recipient %#x: got %d, want 1", recipient, got)
			}
			frame++
		}

		sendOld(255, []byte{0x32, 0xA0}, nil)
		mismatch, _ := importantNativeListC()
		sendOld(255, []byte{0x31, 0xB0}, relatedObject)
		broadcast, _ := importantNativeListC()
		(*importantPacketLegacy)(broadcast).AcknowledgedMask = (uint32(1) << 1) | (uint32(1) << 3)
		sendOld(2, []byte{0x31, 0xC0}, relatedObject)
		sendOld(3, []byte{0x31, 0xD0}, relatedObject)
		directOther, _ := importantNativeListC()
		sendOld(0x81, []byte{0x31, 0xE0}, relatedObject)
		excluded, _ := importantNativeListC()

		payload := []byte{0x31, 0xF0, 0x0F}
		if got := sendImportantPacketWrapperC(2, payload, relatedObject.CObj(), 0x50607080, importantPacketReplaceExisting); got != 1 {
			t.Fatalf("replacement return value: got %d, want 1", got)
		}
		first, last := importantNativeListC()
		if first == nil || last != mismatch {
			t.Fatalf("replacement list ends: got (%p, %p), want (new, %p)", first, last, mismatch)
		}
		if importantPacketNextC(first) != excluded || importantPacketPrevC(excluded) != first {
			t.Fatal("replacement head is not linked to the excluded-recipient survivor")
		}
		if importantPacketNextC(excluded) != directOther || importantPacketPrevC(directOther) != excluded {
			t.Fatal("excluded-recipient survivor is not linked to the other direct survivor")
		}
		if importantPacketNextC(directOther) != mismatch || importantPacketPrevC(mismatch) != directOther {
			t.Fatal("other direct survivor is not linked to the mismatched tail")
		}
		if importantPacketPrevC(first) != nil || importantPacketNextC(mismatch) != nil {
			t.Fatal("replacement list has non-nil outer links")
		}

		packet := (*importantPacketLegacy)(first)
		if packet.CreatedFrame != frame || packet.Recipient != 2 || packet.RemoveIfDisconnected != 0x50607080 {
			t.Errorf("replacement packet fields: frame=%#x recipient=%d policy=%#x", packet.CreatedFrame, packet.Recipient, packet.RemoveIfDisconnected)
		}
		if packet.PayloadSize != byte(len(payload)) || !bytes.Equal(packet.Payload[:len(payload)], payload) {
			t.Errorf("replacement payload: size=%d data=%x, want size=%d data=%x", packet.PayloadSize, packet.Payload[:len(payload)], len(payload), payload)
		}
		if packet.SequenceEnabled != 0 || packet.Sequence != [32]uint16{} {
			t.Errorf("replacement sequence state: enabled=%d values=%#v", packet.SequenceEnabled, packet.Sequence)
		}
		if got := importantPacketRelatedObjectC(first); got != relatedObject.CObj() {
			t.Errorf("replacement related object: got %p, want %p", got, relatedObject.CObj())
		}
		if got := (*importantPacketLegacy)(excluded).AcknowledgedMask; got != uint32(1)<<2 {
			t.Errorf("excluded packet acknowledgement: got %#x, want %#x", got, uint32(1)<<2)
		}
		if got := (*importantPacketLegacy)(directOther).AcknowledgedMask; got != 0 {
			t.Errorf("other direct packet acknowledgement: got %#x, want 0", got)
		}
		if got := relatedObject.Field37; got != uint32(1)<<6 {
			t.Errorf("related-object sync mask: got %#x, want %#x", got, uint32(1)<<6)
		}
	})

	for _, tc := range []struct {
		name      string
		recipient int
	}{
		{name: "broadcast", recipient: 255},
		{name: "high-bit", recipient: 0x82},
	} {
		t.Run(tc.name+"-raw-high-message-byte", func(t *testing.T) {
			preserveImportantPacketState(t)
			setImportantNativeListC(nil, nil)
			pool := alloc.NewClass("important-replacement-high-byte-contract", importantPacketSizeC(), 5)
			setImportantAllocClassC(pool.UPtr())
			setImportantRecipientMaskC((uint32(1) << 0) | (uint32(1) << 1))
			frame := uint32(0x42000000)
			gameFrameHook = func() uint32 { return frame }
			sendOld := func(messageType byte) {
				t.Helper()
				if got := sendImportantPacketC(0, []byte{messageType, 0xA5}, nil, 1, false); got != 1 {
					t.Fatalf("old packet type %#x: got %d, want 1", messageType, got)
				}
				frame++
			}

			sendOld(0xDE)
			sendOld(0x45)
			mismatch, _ := importantNativeListC()
			sendOld(0xDE)
			sendOld(0xDE)
			payload := []byte{0xDE, 0x5A, 0xC3}
			if got := sendImportantPacketWrapperC(tc.recipient, payload, nil, 0x12345678, importantPacketReplaceExisting); got != 1 {
				t.Fatalf("replacement return value: got %d, want 1", got)
			}
			first, last := importantNativeListC()
			if first == nil || first == mismatch || last != mismatch {
				t.Fatalf("replacement list ends: got (%p, %p), want (new, %p)", first, last, mismatch)
			}
			if importantPacketPrevC(first) != nil || importantPacketNextC(first) != mismatch ||
				importantPacketPrevC(mismatch) != first || importantPacketNextC(mismatch) != nil {
				t.Fatal("matching head/middle/tail packets were not removed before insertion")
			}
			packet := (*importantPacketLegacy)(first)
			if packet.CreatedFrame != frame || packet.Recipient != int8(uint8(tc.recipient)) ||
				packet.PayloadSize != byte(len(payload)) || !bytes.Equal(packet.Payload[:len(payload)], payload) {
				t.Errorf("replacement packet: frame=%#x recipient=%#x size=%d payload=%x", packet.CreatedFrame, uint8(packet.Recipient), packet.PayloadSize, packet.Payload[:len(payload)])
			}
		})
	}

	t.Run("masked-variable-shift-count", func(t *testing.T) {
		preserveImportantPacketState(t)
		setImportantNativeListC(nil, nil)
		pool := alloc.NewClass("important-replacement-shift-contract", importantPacketSizeC(), 2)
		setImportantAllocClassC(pool.UPtr())
		setImportantRecipientMaskC((uint32(1) << 0) | (uint32(1) << 1))
		frame := uint32(0x43000000)
		gameFrameHook = func() uint32 { return frame }
		payload := []byte{0x45, 0xAA}
		if got := sendImportantPacketC(255, payload, nil, 1, false); got != 1 {
			t.Fatalf("old broadcast packet: got %d, want 1", got)
		}
		old, _ := importantNativeListC()
		(*importantPacketLegacy)(old).AcknowledgedMask = uint32(1) << 1
		frame++
		if got := sendImportantPacketWrapperC(0x100, payload, nil, 0x10203040, importantPacketReplaceExisting); got != 1 {
			t.Fatalf("replacement return value: got %d, want 1", got)
		}
		first, last := importantNativeListC()
		if first == nil || first != last || importantPacketPrevC(first) != nil || importantPacketNextC(first) != nil {
			t.Fatalf("masked-shift list: got (%p, %p), want one replacement packet", first, last)
		}
		packet := (*importantPacketLegacy)(first)
		if packet.CreatedFrame != frame || packet.Recipient != 0 || packet.AcknowledgedMask != 0 {
			t.Errorf("masked-shift replacement: frame=%#x recipient=%d acknowledged=%#x", packet.CreatedFrame, packet.Recipient, packet.AcknowledgedMask)
		}
	})
}
