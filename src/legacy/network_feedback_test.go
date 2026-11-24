package legacy

import (
	"bytes"
	"math"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestEarthquakeJigglePacketsUseNativeDistanceAndQueue(t *testing.T) {
	srv := objectStatusTestServer(t, 1)
	srv.NetList.Init()
	t.Cleanup(srv.NetList.Free)
	center := types.Pointf{}
	if result := earthquakePacketNativeCall(nil, -30); result != 0 {
		t.Fatalf("empty native roster returned %d", result)
	}

	rows := []struct {
		pos     types.Pointf
		hasUnit bool
		payload []byte
	}{
		{types.Pointf{}, true, []byte{0x97, 0xe2}},
		{types.Pointf{}, false, nil},
		{types.Pointf{X: 150, Y: 150}, true, []byte{0x97, 0xf2}},
		{types.Pointf{X: 300}, true, nil},
		{types.Pointf{X: 299}, true, []byte{0x97, 0}},
		{types.Pointf{X: float32(math.NaN())}, true, []byte{0x97, 0}},
	}
	for i, row := range rows {
		player := srv.Players.NewRaw(101 + i)
		player.Pos3632Vec = row.pos
		if row.hasUnit {
			unit, freeUnit := alloc.NewHigh(server.Object{})
			t.Cleanup(freeUnit)
			update, freeUpdate := alloc.NewHigh(server.PlayerUpdateData{})
			t.Cleanup(freeUpdate)
			update.Player = player
			*unit = server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
			player.PlayerUnit = unit
		}
	}
	for _, test := range []struct {
		name string
		send func(*testing.T)
	}{
		{"native", func(t *testing.T) {
			if result := earthquakePacketNativeCall(&center, -30); result != 0 {
				t.Fatalf("native earthquake returned %d", result)
			}
		}},
		{"Go", func(*testing.T) { srv.Nox_xxx_earthquakeSend_4D9110(center, -30) }},
	} {
		t.Run(test.name, func(t *testing.T) {
			test.send(t)
			for i, row := range rows {
				if got := srv.NetList.CopyPacketsA(ntype.PlayerInd(i), netlist.Kind1); !bytes.Equal(got, row.payload) {
					t.Errorf("player %d jitter packet %x, want %x", i, got, row.payload)
				}
			}
		})
	}
	if result := jigglePacketNativeCall(2, -125); result != 1 {
		t.Fatalf("native queue success returned %d", result)
	}
	if got := srv.NetList.CopyPacketsA(2, netlist.Kind1); !bytes.Equal(got, []byte{0x97, 0x83}) {
		t.Fatalf("direct screen-shake packet %x", got)
	}
	if !srv.NetList.AddToMsgListCli(2, netlist.Kind1, make([]byte, 2048)) {
		t.Fatal("failed to fill the client message buffer")
	}
	if result := jigglePacketNativeCall(2, 1); result != 0 {
		t.Fatalf("full native queue returned %d", result)
	}
}

func TestLatencyReportNativeRotatesAndSamplesAgain(t *testing.T) {
	srv := objectStatusTestServer(t, 1)
	srv.NetList.Init()
	t.Cleanup(srv.NetList.Free)
	first, second := srv.Players.NewRaw(0x12345), srv.Players.NewRaw(0x23456)
	host := srv.Players.ListSlots()[server.HostPlayerIndex]
	*host = server.Player{PlayerInd: byte(server.HostPlayerIndex), NetCodeVal: 0x34567, Active: 1}
	previous := Sub_554240
	t.Cleanup(func() { Sub_554240 = previous })
	var calls []ntype.PlayerInd
	Sub_554240 = func(index ntype.PlayerInd) int {
		calls = append(calls, index)
		if index == second.PlayerIndex() {
			if len(calls) == 2 {
				return 1
			}
			return 0x12345
		}
		return 0
	}
	checkReport := func(want []byte, recipients ...*server.Player) {
		t.Helper()
		if !latencyReportNativeCall() {
			t.Error("native latency entry returned a roster member")
		}
		for _, recipient := range recipients {
			if got := srv.NetList.CopyPacketsA(recipient.PlayerIndex(), netlist.Kind1); !bytes.Equal(got, want) {
				t.Errorf("player %d latency packet %x, want %x", recipient.PlayerInd, got, want)
			}
		}
	}
	checkReport([]byte{0xd7, 0x56, 0x34, 0x45, 0x23}, first, second, host)
	if want := []ntype.PlayerInd{0, 1, 1}; !slices.Equal(calls, want) {
		t.Fatalf("sample calls %v, want %v", calls, want)
	}
	calls = calls[:0]
	checkReport([]byte{0xd7, 0x67, 0x45, 0, 0}, first, second, host)
	if want := []ntype.PlayerInd{server.HostPlayerIndex}; !slices.Equal(calls, want) {
		t.Fatalf("host sample calls %v, want %v", calls, want)
	}

	second.Active, host.Active = 0, 0
	calls = calls[:0]
	checkReport([]byte{0xd7, 0x45, 0x23, 0, 0}, first)
	if len(calls) != 33 {
		t.Fatalf("zero-latency roster sampled %d times, want 32 checks and one report", len(calls))
	}
	for _, recipient := range []*server.Player{second, host} {
		if got := srv.NetList.CopyPacketsA(recipient.PlayerIndex(), netlist.Kind1); len(got) != 0 {
			t.Errorf("inactive player %d received %x", recipient.PlayerInd, got)
		}
	}
	first.Active = 0
	calls = calls[:0]
	checkReport(nil, first)
	if len(calls) != 0 {
		t.Fatalf("empty roster sampled latency %v", calls)
	}
}

func TestGauntletStartPacketPreservesQueueOptions(t *testing.T) {
	objectStatusTestServer(t, 1)
	setImportantRecipientMaskC(1 << 6)
	if result := gauntletStartNativeCall(6); result != 1 {
		t.Fatalf("native gauntlet start returned %d", result)
	}
	first, _ := importantNativeListC()
	checkObjectStatusPacket(t, first, 6, []byte{0xf0, 0}, 1, 1)
}
