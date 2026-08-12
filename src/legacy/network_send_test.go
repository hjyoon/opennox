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

	if got := sendImportantPacketWrapperC(6, payload, relatedObject, 0x50607080, false); got != 1 {
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
