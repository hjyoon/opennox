package legacy

import (
	"bytes"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func TestGUISettingsDeltaNative4DF020(t *testing.T) {
	world := objectStatusTestServer(t, 2)
	first := world.Players.NewRaw(123)
	host := world.Players.ByIndRaw(server.HostPlayerIndex)
	host.Active, host.PlayerInd = 1, server.HostPlayerIndex
	setImportantRecipientMaskC(1<<first.PlayerInd | 1<<host.PlayerInd)
	last := memmap.Slice(0x5D4594, 1563214)[:58]
	previous := bytes.Clone(last)
	t.Cleanup(func() { copy(last, previous) })
	clear(last)
	settings := &server.Settings2{Field52: 0x400, Field54: 25, Field56: 2, Field57: 1}
	copy(settings.Field0[:], "estate")
	if got := reportGUISettingsIfChanged4DF020(settings, -1); got != 0 {
		t.Fatalf("changed settings returned %d", got)
	}
	want := []byte{0xb1, 1}
	wire := unsafe.Slice((*byte)(unsafe.Pointer(settings)), 58)
	want = append(want, wire...)
	packet, _ := importantNativeListC()
	checkObjectStatusPacket(t, packet, first.PlayerInd, want, 1, 0)
	if importantPacketNextC(packet) != nil || !bytes.Equal(last, wire) {
		t.Fatal("settings report included the host or failed to update its snapshot")
	}
	if got := reportGUISettingsIfChanged4DF020(settings, -1); got != -1 {
		t.Fatalf("same settings returned %d", got)
	}
	if same, _ := importantNativeListC(); same != packet {
		t.Fatal("same settings enqueued a second report")
	}
}
