package legacy

import (
	"bytes"
	_ "embed"
	"encoding/hex"
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/trade_packets_original_v1.json.gz
var tradePacketNativeOriginal []byte

type tradePacketNativeCase struct {
	Input struct {
		Entry         string
		Name          string
		RecipientRole int  `json:"recipient_role"`
		OwnersAlias   bool `json:"owners_alias"`
		Flags         [2]uint32
		Classes       [2]uint32
		Gold          [2]uint32
		Values        [2]uint32
		PlayerIndices [3]byte `json:"player_indices"`
		NetCode       uint32  `json:"net_code"`
		UnusedSession uint32  `json:"unused_session"`
		ServiceReturn uint32  `json:"service_return"`
	}
	Events []struct {
		Service              string
		Recipient            int
		Payload              string `json:"payload_hex"`
		Length               int
		RelatedObject        uint32 `json:"related_object"`
		RemoveIfDisconnected int    `json:"remove_if_disconnected"`
		ReturnWord           uint32 `json:"return_u32"`
	}
	ReturnWord uint32 `json:"return_u32"`
}

func TestTradePacketNativeLayout(t *testing.T) {
	var session server.TradeSession
	var obj server.Object
	var update server.PlayerUpdateData
	var player server.Player
	// C exposes the shared prefixes; Go retains its existing trailing fields.
	objectPrefix := unsafe.Offsetof(obj.ScriptPickup) + unsafe.Sizeof(obj.ScriptPickup)
	updatePrefix := unsafe.Offsetof(update.QuestWarpGate) + unsafe.Sizeof(update.QuestWarpGate)
	want := [31]uintptr{unsafe.Sizeof(uintptr(0)), unsafe.Sizeof(session), unsafe.Offsetof(session.Field8), unsafe.Offsetof(session.Field12), unsafe.Offsetof(session.Field24), unsafe.Offsetof(session.Field28), unsafe.Offsetof(session.Field40), unsafe.Offsetof(session.Field44), unsafe.Offsetof(session.Field48), unsafe.Offsetof(session.Field52), objectPrefix, unsafe.Offsetof(obj.ObjClass), unsafe.Offsetof(obj.UpdateData), unsafe.Sizeof(obj.UpdateData), unsafe.Offsetof(obj.InitData), unsafe.Sizeof(obj.InitData), unsafe.Offsetof(obj.NetCode), unsafe.Sizeof(obj.NetCode), updatePrefix, unsafe.Offsetof(update.Player), unsafe.Sizeof(update.Player), unsafe.Sizeof(player), unsafe.Offsetof(player.PlayerInd), unsafe.Sizeof(player.PlayerInd), unsafe.Sizeof(int32(0)), unsafe.Sizeof(uintptr(0)), unsafe.Sizeof(uintptr(0)), unsafe.Sizeof(session.Field24), unsafe.Sizeof(session.Field40), unsafe.Sizeof(session.Field8), unsafe.Sizeof(uint32(0))}
	if got := tradePacketNativeLayoutC(); got != want {
		t.Fatal("native trade/object/player/network layout", got, want)
	}
	if unsafe.Sizeof(obj) != objectPrefix+2*unsafe.Sizeof(uintptr(0)) || unsafe.Sizeof(update) < updatePrefix {
		t.Fatal("shared prefixes or existing Go extensions exceed their owners")
	}
	if unsafe.Sizeof(session) > 0x800 || unsafe.Sizeof(obj) > 0x1000 || unsafe.Sizeof(update) > 0x800 || unsafe.Sizeof(player) > 0x2000 {
		t.Fatal("foreign record stride too small")
	}
}

func TestTradePacketNativeOriginal(t *testing.T) {
	var cases []tradePacketNativeCase
	npcVoiceOriginalDecode(t, tradePacketNativeOriginal, "8d3556313566615f3ef816e5710f16358e149a566559df8118b3386bc896cd73", "1ddb1809698f8a23371c94994a40f4aabb7fe8b45bff8f7a6b533435ef54e9ac", 508, &cases)
	entries := map[string]int{"sub_50F720": 0, "nox_xxx_tradeP2PUpdStuff_50FA00": 1, "sub_50FF90": 2}
	oldSend := tradePacketSendForTest
	defer func() { tradePacketSendForTest = oldSend }()
	for route := 0; route < 3; route++ {
		for _, placement := range []int{0, 64, 512} {
			raw, free := alloc.MakeHigh([]byte{}, 65536)
			t.Cleanup(free)
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&raw[0])) <= math.MaxUint32 {
				t.Fatal("foreign trade/object/player owner must exceed4GiB")
			}
			session := (*server.TradeSession)(unsafe.Pointer(&raw[0x100+placement]))
			var owners [3]*server.Object
			var updates [3]*server.PlayerUpdateData
			var players [3]*server.Player
			var golds [2]*server.Object
			for i := range owners {
				owners[i] = (*server.Object)(unsafe.Pointer(&raw[0x1000+i*0x1000+placement]))
				updates[i] = (*server.PlayerUpdateData)(unsafe.Pointer(&raw[0x4000+i*0x800+placement]))
				players[i] = (*server.Player)(unsafe.Pointer(&raw[0x8000+i*0x2000+placement]))
			}
			for i := range golds {
				golds[i] = (*server.Object)(unsafe.Pointer(&raw[0x6000+i*0x1000+placement]))
			}
			item := (*server.Object)(unsafe.Pointer(&raw[0xf000+placement]))
			expected := make([]byte, len(raw))
			for _, c := range cases {
				t.Run(fmt.Sprintf("route%d/place%d/%s-%s", route, placement, c.Input.Entry, c.Input.Name), func(t *testing.T) {
					entry, ok := entries[c.Input.Entry]
					if !ok || c.Input.RecipientRole < 0 || c.Input.RecipientRole >= len(owners) || len(c.Events) != 1 {
						t.Fatal("invalid canonical entry/recipient/service event")
					}
					for i := range raw {
						raw[i] = byte(i*43 + 71)
					}
					for i := range owners {
						class := c.Input.Classes[1]
						if i == 0 {
							class = c.Input.Classes[0]
						}
						owners[i].ObjClass = object.Class(class)
						uiStateTradeStorePointer(unsafe.Pointer(&owners[i].UpdateData), unsafe.Pointer(updates[i]))
						uiStateTradeStorePointer(unsafe.Pointer(&updates[i].Player), unsafe.Pointer(players[i]))
						players[i].PlayerInd = c.Input.PlayerIndices[i]
					}
					uiStateTradeStorePointer(unsafe.Pointer(&session.Field8), unsafe.Pointer(owners[0]))
					second := owners[1]
					if c.Input.OwnersAlias {
						second = owners[0]
					}
					uiStateTradeStorePointer(unsafe.Pointer(&session.Field12), unsafe.Pointer(second))
					session.Field24, session.Field28 = c.Input.Flags[0], c.Input.Flags[1]
					session.Field40, session.Field44 = c.Input.Values[0], c.Input.Values[1]
					uiStateTradeStorePointer(unsafe.Pointer(&session.Field48), unsafe.Pointer(golds[0]))
					uiStateTradeStorePointer(unsafe.Pointer(&session.Field52), unsafe.Pointer(golds[1]))
					for i := range golds {
						uiStateTradeStorePointer(unsafe.Pointer(&golds[i].InitData), unsafe.Pointer(&raw[0xe000+i*0x100+placement]))
						golds[i].InitDataGold().Amount = c.Input.Gold[i]
					}
					item.NetCode = c.Input.NetCode
					recipient := owners[c.Input.RecipientRole]
					argument := unsafe.Pointer(session)
					if entry == 2 {
						// The original second word is unused. Its native projection
						// is nil or an opaque full-width canary pointer, never an
						// invented Go pointer made from the original scalar bits.
						argument = nil
						if c.Input.UnusedSession != 0 {
							argument = unsafe.Pointer(&raw[0xe800+placement])
						}
					}
					copy(expected, raw)
					calls := 0
					want := c.Events[0]
					payload, err := hex.DecodeString(want.Payload)
					if err != nil || len(payload) != want.Length || want.Service != "nox_xxx_netSendPacket1_4E5390" || want.ReturnWord != c.Input.ServiceReturn {
						t.Fatal("invalid original service event")
					}
					tradePacketSendForTest = func(index int, packet []byte, related *server.Object, disconnect int) int {
						calls++
						if index != want.Recipient || !bytes.Equal(packet, payload) || related != nil || want.RelatedObject != 0 || disconnect != want.RemoveIfDisconnected {
							t.Errorf("original network event: index %d packet %x related %p disconnect %d; want %+v", index, packet, related, disconnect, want)
						}
						if !bytes.Equal(raw, expected) {
							t.Error("unexpected foreign-owner write before network service")
						}
						return int(int32(c.Input.ServiceReturn))
					}
					if got := uint32(tradePacketNativeCall(route, entry, recipient, argument, item)); got != c.ReturnWord || calls != 1 {
						t.Fatalf("original network return/calls: %#x/%d want %#x/1", got, calls, c.ReturnWord)
					}
					if !bytes.Equal(raw, expected) {
						t.Fatal("unexpected write to whole64KiB foreign trade/object/player owner")
					}
				})
			}
		}
	}
}
