package legacy

import (
	"testing"

	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
)

func TestImportantFreeSlotsMatchesGAMEEXEContract(t *testing.T) {
	handles.Init()
	t.Cleanup(handles.Release)

	for _, tc := range []struct {
		name        string
		capacity    uint32
		packetCount int
		want        uint32
	}{
		{name: "empty-zero", capacity: 0, want: 0},
		{name: "empty-high-bit", capacity: 0x80000000, want: 0x80000000},
		{name: "three-of-five", capacity: 5, packetCount: 3, want: 2},
		{name: "unsigned-underflow", capacity: 1, packetCount: 3, want: 0xFFFFFFFE},
	} {
		t.Run(tc.name, func(t *testing.T) {
			specs := make([]importantResendPacketSpec, tc.packetCount)
			for i := range specs {
				specs[i] = importantResendPacketSpec{recipient: -1, payload: []byte{byte(0x80 + i)}}
			}
			_, packets, _ := setupImportantResendContract(t, 0, specs)
			setImportantCapacityC(tc.capacity)
			before := make([]importantPacketLegacy, len(packets))
			for i, packet := range packets {
				before[i] = *(*importantPacketLegacy)(packet)
			}

			if got := importantFreeSlotsC(); got != tc.want {
				t.Errorf("free slots: got %#x, want %#x", got, tc.want)
			}
			if got := importantCapacityC(); got != tc.capacity {
				t.Errorf("capacity changed: got %#x, want %#x", got, tc.capacity)
			}
			for i, packet := range packets {
				if got := *(*importantPacketLegacy)(packet); got != before[i] {
					t.Errorf("packet %d changed while counting", i)
				}
			}
			assertImportantResendList(t, packets)
		})
	}
}

func TestImportantNoopMatchesGAMEEXEContract(t *testing.T) {
	handles.Init()
	t.Cleanup(handles.Release)
	const capacity = uint32(0x89ABCDEF)
	_, packets, _ := setupImportantResendContract(t, 0, []importantResendPacketSpec{{
		recipient: -1, payload: []byte{0x91, 0x92, 0x93},
	}})
	setImportantCapacityC(capacity)
	before := *(*importantPacketLegacy)(packets[0])

	importantNoopC(packets[0])

	if got := importantCapacityC(); got != capacity {
		t.Errorf("capacity changed: got %#x, want %#x", got, capacity)
	}
	if got := *(*importantPacketLegacy)(packets[0]); got != before {
		t.Error("no-op changed the packet")
	}
	assertImportantResendList(t, packets)
}
