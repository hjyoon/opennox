package legacy

import (
	"testing"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

type importantPacketContractRef struct {
	c      *importantPacketC
	legacy *importantPacketLegacy
}

func setupImportantAcknowledgementPacket(
	t *testing.T, recipient int, recipientMask uint32, messageType byte, relatedObject *server.Object,
) importantPacketContractRef {
	t.Helper()
	preserveImportantPacketState(t)
	setImportantNativeListC(nil, nil)
	pool := alloc.NewClass("important-acknowledgement-contract", importantPacketSizeC(), 1)
	setImportantAllocClassC(pool.UPtr())
	setImportantRecipientMaskC(recipientMask)
	gameFrameHook = func() uint32 { return 0x54000000 }
	var got int
	if relatedObject == nil {
		got = sendImportantPacketC(recipient, []byte{messageType, 0xA5}, nil, 1, false)
	} else {
		got = sendImportantPacketC(recipient, []byte{messageType, 0xA5}, relatedObject.CObj(), 1, false)
	}
	if got != 1 {
		t.Fatalf("packet creation: got %d, want 1", got)
	}
	first, last := importantNativeListC()
	if first == nil || first != last {
		t.Fatalf("packet list: got (%p, %p), want one packet", first, last)
	}
	return importantPacketContractRef{
		c:      (*importantPacketC)(first),
		legacy: (*importantPacketLegacy)(first),
	}
}

func assertImportantPacketSurvives(t *testing.T, packet *importantPacketC) {
	t.Helper()
	first, last := importantNativeListC()
	if first == nil || (*importantPacketC)(first) != packet || (*importantPacketC)(last) != packet {
		t.Fatalf("packet list: got (%p, %p), want surviving packet %p", first, last, packet)
	}
	if importantPacketPrevC(first) != nil || importantPacketNextC(first) != nil {
		t.Fatal("surviving packet has non-nil links")
	}
}

func assertImportantPacketRemoved(t *testing.T) {
	t.Helper()
	if first, last := importantNativeListC(); first != nil || last != nil {
		t.Fatalf("packet list: got (%p, %p), want empty", first, last)
	}
}

func TestImportantPacketAcknowledgementRelatedObjectMatchesGAMEEXEContract(t *testing.T) {
	handles.Init()
	t.Cleanup(handles.Release)
	const (
		initialField37 = uint32(0xA5A5FFFF)
		clientMask     = uint32(0x80000020)
	)
	for _, tc := range []struct {
		name        string
		messageType byte
		related     bool
		clearsMask  bool
	}{
		{name: "before-range", messageType: 0x30, related: true},
		{name: "range-start", messageType: 0x31, related: true, clearsMask: true},
		{name: "range-middle", messageType: 0x32, related: true, clearsMask: true},
		{name: "range-end", messageType: 0x33, related: true, clearsMask: true},
		{name: "after-range", messageType: 0x34, related: true},
		{name: "nil-related-object", messageType: 0x31},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var relatedObject *server.Object
			if tc.related {
				var free func()
				relatedObject, free = alloc.New(server.Object{})
				t.Cleanup(free)
				relatedObject.Field37 = initialField37
			}
			packet := setupImportantAcknowledgementPacket(t, 7, 1<<7, tc.messageType, relatedObject)
			acknowledgeImportantPacketC(clientMask, packet.c, 6)
			assertImportantPacketSurvives(t, packet.c)
			if packet.legacy.AcknowledgedMask != 0 {
				t.Errorf("direct packet acknowledgement: got %#x, want 0", packet.legacy.AcknowledgedMask)
			}
			if relatedObject != nil {
				want := initialField37
				if tc.clearsMask {
					want &^= clientMask
				}
				if relatedObject.Field37 != want {
					t.Errorf("related-object sync mask: got %#x, want %#x", relatedObject.Field37, want)
				}
			}
		})
	}
}

func TestImportantPacketAcknowledgementRecipientModesMatchGAMEEXEContract(t *testing.T) {
	handles.Init()
	t.Cleanup(handles.Release)

	t.Run("broadcast", func(t *testing.T) {
		for _, tc := range []struct {
			name                string
			currentMask         uint32
			initialAcknowledged uint32
			clientMask          uint32
			removed             bool
			wantAcknowledged    uint32
		}{
			{name: "partial", currentMask: 0x15, initialAcknowledged: 0x01, clientMask: 0x04, wantAcknowledged: 0x05},
			{name: "current-mask-complete", currentMask: 0x05, initialAcknowledged: 0x01, clientMask: 0x04, removed: true},
			{name: "no-active-recipients", currentMask: 0, removed: true},
		} {
			t.Run(tc.name, func(t *testing.T) {
				packet := setupImportantAcknowledgementPacket(t, 255, 0x15, 0x40, nil)
				packet.legacy.AcknowledgedMask = tc.initialAcknowledged
				setImportantRecipientMaskC(tc.currentMask)
				acknowledgeImportantPacketC(tc.clientMask, packet.c, 4)
				if tc.removed {
					assertImportantPacketRemoved(t)
					return
				}
				assertImportantPacketSurvives(t, packet.c)
				if packet.legacy.AcknowledgedMask != tc.wantAcknowledged {
					t.Errorf("acknowledged mask: got %#x, want %#x", packet.legacy.AcknowledgedMask, tc.wantAcknowledged)
				}
			})
		}
	})

	t.Run("excluded-recipient", func(t *testing.T) {
		for _, tc := range []struct {
			name                string
			recipient           int
			recipientMask       uint32
			initialAcknowledged uint32
			clientMask          uint32
			removed             bool
			wantAcknowledged    uint32
		}{
			{name: "partial", recipient: 0x82, recipientMask: 0x07, clientMask: 0x02, wantAcknowledged: 0x02},
			{name: "complete", recipient: 0x82, recipientMask: 0x07, initialAcknowledged: 0x01, clientMask: 0x02, removed: true},
			{name: "masked-exclusion-index", recipient: 0xA1, recipientMask: 0x03, initialAcknowledged: 0x01, removed: true},
		} {
			t.Run(tc.name, func(t *testing.T) {
				packet := setupImportantAcknowledgementPacket(t, tc.recipient, tc.recipientMask, 0x40, nil)
				packet.legacy.AcknowledgedMask = tc.initialAcknowledged
				acknowledgeImportantPacketC(tc.clientMask, packet.c, 5)
				if tc.removed {
					assertImportantPacketRemoved(t)
					return
				}
				assertImportantPacketSurvives(t, packet.c)
				if packet.legacy.AcknowledgedMask != tc.wantAcknowledged {
					t.Errorf("acknowledged mask: got %#x, want %#x", packet.legacy.AcknowledgedMask, tc.wantAcknowledged)
				}
			})
		}
	})

	t.Run("direct-recipient", func(t *testing.T) {
		for _, tc := range []struct {
			name        string
			playerIndex int
			removed     bool
		}{
			{name: "exact", playerIndex: 7, removed: true},
			{name: "same-low-byte", playerIndex: 0x107},
		} {
			t.Run(tc.name, func(t *testing.T) {
				packet := setupImportantAcknowledgementPacket(t, 7, 1<<7, 0x40, nil)
				acknowledgeImportantPacketC(1<<7, packet.c, tc.playerIndex)
				if tc.removed {
					assertImportantPacketRemoved(t)
				} else {
					assertImportantPacketSurvives(t, packet.c)
					if packet.legacy.AcknowledgedMask != 0 {
						t.Errorf("direct acknowledged mask: got %#x, want 0", packet.legacy.AcknowledgedMask)
					}
				}
			})
		}
	})
}

func TestImportantPacketFrameAcknowledgementMatchesGAMEEXEContract(t *testing.T) {
	handles.Init()
	t.Cleanup(handles.Release)

	t.Run("empty-list", func(t *testing.T) {
		preserveImportantPacketState(t)
		setImportantNativeListC(nil, nil)
		if got := acknowledgeImportantFrameC(0, 0xFEDCBA98); got != 0 {
			t.Errorf("return value: got %d, want 0", got)
		}
	})

	t.Run("head-middle-tail-removal", func(t *testing.T) {
		preserveImportantPacketState(t)
		setImportantNativeListC(nil, nil)
		pool := alloc.NewClass("important-frame-acknowledgement-contract", importantPacketSizeC(), 5)
		setImportantAllocClassC(pool.UPtr())
		setImportantRecipientMaskC((uint32(1) << 0) | (uint32(1) << 1))
		gameFrameHook = func() uint32 { return 0x55000000 }
		const acknowledgedFrame = uint32(0xFEDCBA98)
		create := func(matches bool) *importantPacketC {
			t.Helper()
			if got := sendImportantPacketC(0, []byte{0x40}, nil, 1, false); got != 1 {
				t.Fatalf("packet creation: got %d, want 1", got)
			}
			first, _ := importantNativeListC()
			packet := (*importantPacketLegacy)(first)
			packet.LastSendFrame[0] = acknowledgedFrame - 1
			packet.LastSendFrame[1] = acknowledgedFrame
			if matches {
				packet.LastSendFrame[0] = acknowledgedFrame
			}
			return (*importantPacketC)(first)
		}

		create(true)
		tailSurvivor := create(false)
		create(true)
		headSurvivor := create(false)
		create(true)
		if got := acknowledgeImportantFrameC(0, acknowledgedFrame); got != 0 {
			t.Errorf("return value: got %d, want 0", got)
		}
		first, last := importantNativeListC()
		if (*importantPacketC)(first) != headSurvivor || (*importantPacketC)(last) != tailSurvivor {
			t.Fatalf("survivor list: got (%p, %p), want (%p, %p)", first, last, headSurvivor, tailSurvivor)
		}
		if importantPacketPrevC(first) != nil || importantPacketNextC(first) != last ||
			importantPacketPrevC(last) != first || importantPacketNextC(last) != nil {
			t.Fatal("survivor links are inconsistent after matched removals")
		}
	})

	t.Run("player-index-31", func(t *testing.T) {
		preserveImportantPacketState(t)
		setImportantNativeListC(nil, nil)
		pool := alloc.NewClass("important-frame-acknowledgement-index-contract", importantPacketSizeC(), 1)
		setImportantAllocClassC(pool.UPtr())
		setImportantRecipientMaskC(uint32(1) << 31)
		gameFrameHook = func() uint32 { return 0x55000001 }
		if got := sendImportantPacketC(31, []byte{0x40}, nil, 1, false); got != 1 {
			t.Fatalf("packet creation: got %d, want 1", got)
		}
		first, _ := importantNativeListC()
		(*importantPacketLegacy)(first).LastSendFrame[31] = 0x80000001
		if got := acknowledgeImportantFrameC(31, 0x80000001); got != 0 {
			t.Errorf("return value: got %d, want 0", got)
		}
		assertImportantPacketRemoved(t)
	})
}
