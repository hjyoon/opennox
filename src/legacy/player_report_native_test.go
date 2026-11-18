package legacy

import (
	"bytes"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/player_self_report_original_v2.json.gz
var playerSelfReportOriginal []byte

//go:embed testdata/player_report_packet_original_v1.json.gz
var playerReportPacketOriginal []byte

type playerSelfReportCase struct {
	Input struct {
		Name   string
		Active bool
		Class  uint32
	}
	Observed struct {
		Events []json.RawMessage
	}
}

func TestPlayerSelfReportNativeCABI4D9900(t *testing.T) {
	rows := tradeOfferOriginalRows[playerSelfReportCase](t, playerSelfReportOriginal,
		"058b055dcc6b6a2bca68f150404c1fd69e683f80d5a439da2a8fe43536c144f1", 96)
	unit, release := alloc.NewHigh(server.Object{})
	t.Cleanup(release)
	if cgoABIPointerSize == 8 && uintptr(unsafe.Pointer(unit)) <= 0xffffffff {
		t.Fatal("player report needs a native object above 4 GiB")
	}
	var calls int
	var got *server.Object
	old := playerObjectSendRuntimeNative518C30
	t.Cleanup(func() { playerObjectSendRuntimeNative518C30 = old })
	playerObjectSendRuntimeNative518C30 = func() server.PlayerObjectSendRuntime518C30 {
		return server.PlayerObjectSendRuntime518C30{Self: func(obj *server.Object) {
			calls++
			got = obj
		}}
	}
	for route := range 2 {
		for _, row := range rows {
			t.Run(fmt.Sprintf("route%d-%s", route, row.Input.Name), func(t *testing.T) {
				unit.ObjClass = object.Class(row.Input.Class)
				before := *unit
				var arg *server.Object
				if row.Input.Active {
					arg = unit
				}
				calls, got = 0, nil
				callPlayerSelfReport4D9900(route, arg)
				if len(row.Observed.Events) == 0 {
					if calls != 0 || got != nil {
						t.Fatal("stock nil or class gate was not preserved")
					}
				} else if calls != 1 || got != unit {
					t.Fatalf("Go sender received %d calls with pointer %p, want one with %p", calls, got, unit)
				}
				if *unit != before {
					t.Fatal("C adapter modified the object")
				}
			})
		}
	}
}

type playerReportPacketCase struct {
	Input struct {
		Name, Entry            string
		Code, Recipient, Flags uint32
		Value                  uint8
		QueueResult            uint32 `json:"queue_result"`
	}
	Observed struct {
		ReturnWord uint32 `json:"return_word"`
		Packet     struct {
			Recipient, Related uint32
			Remove, Sequence   int
			Bytes              string
		}
	}
}

type playerReportPacketServer struct {
	Server
	world *server.Server
}

func (s *playerReportPacketServer) S() *server.Server { return s.world }

func TestPlayerReportPacketNativeCABI(t *testing.T) {
	rows := tradeOfferOriginalRows[playerReportPacketCase](t, playerReportPacketOriginal,
		"f3bab955d8b0a6c207c22e4a7b7c946919d1e8115e7b335d1125a5cc70abf239", 420)
	world := server.NewHigh(nil, nil, strman.New())
	t.Cleanup(world.Close)
	unit, release := alloc.NewHigh(server.Object{})
	t.Cleanup(release)
	if cgoABIPointerSize == 8 && uintptr(unsafe.Pointer(unit)) <= 0xffffffff {
		t.Fatal("packet fixture needs a native object above 4 GiB")
	}
	oldServer, oldFlags := GetServer, noxflags.GetGame()
	t.Cleanup(func() {
		GetServer = oldServer
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})
	host := &playerReportPacketServer{world: world}
	GetServer = func() Server { return host }
	entries := map[string]int{"quest-silver-key-packet": 0, "quest-gold-key-packet": 1, "creature-shield-packet": 2}
	for route := range 2 {
		for _, row := range rows {
			t.Run(fmt.Sprintf("route%d-%s", route, row.Input.Name), func(t *testing.T) {
				entry, known := entries[row.Input.Entry]
				want := row.Observed.Packet
				wire, err := hex.DecodeString(want.Bytes)
				if !known || err != nil || want.Related != 0 {
					t.Fatal("invalid stock packet fixture")
				}
				unit.NetCode = row.Input.Code
				noxflags.ResetGame()
				noxflags.SetGame(noxflags.GameFlag(row.Input.Flags))
				calls := 0
				world.NetSendPacketXxx = func(recipient int, packet []byte, related *server.Object, remove, sequence int) int {
					calls++
					if recipient != int(int32(want.Recipient)) || related != nil || remove != want.Remove ||
						sequence != want.Sequence || !bytes.Equal(packet, wire) {
						t.Fatalf("packet %x to %d with %p/%d/%d, want %x to %d with nil/%d/%d",
							packet, recipient, related, remove, sequence, wire, int32(want.Recipient), want.Remove, want.Sequence)
					}
					return int(int32(row.Input.QueueResult))
				}
				before := *unit
				if got := callPlayerReportPacketNative(route, entry, row.Input.Recipient, unit, row.Input.Value); got != row.Observed.ReturnWord ||
					calls != 1 || *unit != before {
					t.Fatalf("result %#x, calls %d, object unchanged %t, want %#x and one packet without a write",
						got, calls, *unit == before, row.Observed.ReturnWord)
				}
			})
		}
	}
}
