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

//go:embed testdata/trade_notify_item_original_v1.json.gz
var tradeNotifyItemNativeOriginal []byte

type tradeNotifyItemNativeCase struct {
	Input struct {
		Entry         string
		Name          string
		RecipientRole int    `json:"recipient_role"`
		OfferingRole  int    `json:"offering_role"`
		TypeIndex     uint16 `json:"type_index"`
		NetCode       uint32 `json:"net_code"`
		ClassBits     uint32 `json:"class_bits"`
		Cost          uint32
		Modifiers     [4]*uint32
		PlayerIndices [3]byte `json:"player_indices"`
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

func tradeNotifyItemNativeCases(t *testing.T) []tradeNotifyItemNativeCase {
	t.Helper()
	var cases []tradeNotifyItemNativeCase
	npcVoiceOriginalDecode(t, tradeNotifyItemNativeOriginal, "ad421f1dd206fd446faf21ece6a0065d54e02f963da53c6b794ec670a0e11fd9", "4673c69abfcb692c948079aeac8c764d7cce3ff7e5eba46ac42d141a4a88f0b7", 860, &cases)
	return cases
}

func TestTradeNotifyItemNativeLayout(t *testing.T) {
	var obj server.Object
	var update server.PlayerUpdateData
	var player server.Player
	var attrs server.ModifierInitData
	var modifier server.Modifier
	objectPrefix := unsafe.Offsetof(obj.ScriptPickup) + unsafe.Sizeof(obj.ScriptPickup)
	updatePrefix := unsafe.Offsetof(update.QuestWarpGate) + unsafe.Sizeof(update.QuestWarpGate)
	want := [31]uintptr{unsafe.Sizeof(uintptr(0)), objectPrefix, unsafe.Offsetof(obj.TypeInd), unsafe.Sizeof(obj.TypeInd), unsafe.Offsetof(obj.NetCode), unsafe.Sizeof(obj.NetCode), unsafe.Offsetof(obj.ObjClass), unsafe.Sizeof(obj.ObjClass), unsafe.Offsetof(obj.UpdateData), unsafe.Sizeof(obj.UpdateData), updatePrefix, unsafe.Offsetof(update.Player), unsafe.Sizeof(update.Player), unsafe.Sizeof(player), unsafe.Offsetof(player.PlayerInd), unsafe.Sizeof(player.PlayerInd), unsafe.Offsetof(obj.InitData), unsafe.Sizeof(obj.InitData), unsafe.Sizeof(attrs), unsafe.Offsetof(attrs.Modifiers), unsafe.Sizeof(attrs.Modifiers[0]), unsafe.Offsetof(attrs.Field16), unsafe.Offsetof(modifier.TypeInd), unsafe.Sizeof(modifier.TypeInd), unsafe.Sizeof(uintptr(0)), unsafe.Sizeof(uintptr(0)), unsafe.Sizeof(int32(0)), unsafe.Sizeof(uint32(0)), unsafe.Sizeof(attrs.Modifiers), unsafe.Offsetof(modifier.Name0), unsafe.Sizeof(modifier.Name0)}
	if got := tradeNotifyItemNativeLayoutC(); got != want {
		t.Fatal("native notification/item/player/modifier layouts", got, want)
	}
	if unsafe.Sizeof(obj) != objectPrefix+2*unsafe.Sizeof(uintptr(0)) || unsafe.Sizeof(update) < updatePrefix || unsafe.Sizeof(obj) > 0x1000 || unsafe.Sizeof(update) > 0x800 || unsafe.Sizeof(player) > 0x2000 || unsafe.Sizeof(attrs) > 0x800 || unsafe.Sizeof(server.ModifierEff{}) > 0x200 {
		t.Fatal("shared prefixes, Go extensions or foreign strides exceed owners")
	}
}

type tradeNotifyItemNativeArena struct {
	raw, far          []byte
	before, farBefore []byte
	owners            [3]*server.Object
	updates           [3]*server.PlayerUpdateData
	players           [3]*server.Player
	item              *server.Object
	attrs             *server.ModifierInitData
	mods              [4]*server.ModifierEff
	unused            unsafe.Pointer
}

func tradeNotifyItemNativeArenaAt(raw, far []byte, placement int) *tradeNotifyItemNativeArena {
	a := &tradeNotifyItemNativeArena{raw: raw, far: far, before: make([]byte, len(raw)), farBefore: make([]byte, len(far))}
	for i := range a.owners {
		a.owners[i] = (*server.Object)(unsafe.Pointer(&raw[0x1000+i*0x1000+placement]))
		a.updates[i] = (*server.PlayerUpdateData)(unsafe.Pointer(&raw[0x4000+i*0x800+placement]))
		a.players[i] = (*server.Player)(unsafe.Pointer(&raw[0x6000+i*0x2000+placement]))
	}
	if far != nil {
		a.owners[1] = (*server.Object)(unsafe.Pointer(&far[0x1000+placement]))
	}
	a.item = (*server.Object)(unsafe.Pointer(&raw[0xc000+placement]))
	a.attrs = (*server.ModifierInitData)(unsafe.Pointer(&raw[0xd000+placement]))
	for i := range a.mods {
		a.mods[i] = (*server.ModifierEff)(unsafe.Pointer(&raw[0xd800+i*0x200+placement]))
	}
	a.unused = unsafe.Pointer(&raw[0xf000+placement])
	return a
}

func (a *tradeNotifyItemNativeArena) run(t *testing.T, route int, c tradeNotifyItemNativeCase) {
	t.Helper()
	entry, ok := map[string]int{"nox_xxx_sendEndTradeMsg_50F560": 0, "sub_50F6E0": 1, "sub_50FAE0": 2}[c.Input.Entry]
	if !ok || c.Input.RecipientRole < 0 || c.Input.RecipientRole >= 3 || c.Input.OfferingRole < 0 || c.Input.OfferingRole >= 3 || len(c.Events) != 1 {
		t.Fatal("invalid canonical notification/item event")
	}
	for _, region := range [][]byte{a.raw, a.far} {
		for i := range region {
			region[i] = byte(i*47 + 23)
		}
	}
	for i := range a.owners {
		uiStateTradeStorePointer(unsafe.Pointer(&a.owners[i].UpdateData), unsafe.Pointer(a.updates[i]))
		uiStateTradeStorePointer(unsafe.Pointer(&a.updates[i].Player), unsafe.Pointer(a.players[i]))
		a.players[i].PlayerInd = c.Input.PlayerIndices[i]
	}
	a.item.TypeInd, a.item.NetCode, a.item.ObjClass = c.Input.TypeIndex, c.Input.NetCode, object.Class(c.Input.ClassBits)
	uiStateTradeStorePointer(unsafe.Pointer(&a.item.InitData), unsafe.Pointer(a.attrs))
	for i, mod := range c.Input.Modifiers {
		var pointer unsafe.Pointer
		var index uint32
		if mod != nil {
			pointer, index = unsafe.Pointer(a.mods[i]), *mod
		}
		uiStateTradeStorePointer(unsafe.Pointer(&a.attrs.Modifiers[i]), pointer)
		tradeNotifyItemNativeSetIndex(a.mods[i], index)
	}
	var unused unsafe.Pointer
	if c.Input.UnusedSession != 0 {
		unused = a.unused
	}
	copy(a.before, a.raw)
	copy(a.farBefore, a.far)
	want := c.Events[0]
	payload, err := hex.DecodeString(want.Payload)
	if err != nil || len(payload) != want.Length || want.Service != "nox_xxx_netSendPacket1_4E5390" || want.ReturnWord != c.Input.ServiceReturn {
		t.Fatal("invalid original network event")
	}
	calls := 0
	tradePacketSendForTest = func(index int, packet []byte, related *server.Object, disconnect int) int {
		calls++
		if index != want.Recipient || !bytes.Equal(packet, payload) || related != nil || want.RelatedObject != 0 || disconnect != want.RemoveIfDisconnected {
			t.Errorf("original network event: index %d packet %x related %p disconnect %d; want %+v", index, packet, related, disconnect, want)
		}
		if !bytes.Equal(a.raw, a.before) || !bytes.Equal(a.far, a.farBefore) {
			t.Error("foreign notification/item owner changed before service")
		}
		return int(int32(c.Input.ServiceReturn))
	}
	if got := uint32(tradeNotifyItemNativeCall(route, entry, a.owners[c.Input.RecipientRole], a.owners[c.Input.OfferingRole], unused, a.item, c.Input.Cost)); got != c.ReturnWord || calls != 1 {
		t.Fatalf("original network return/calls: %#x/%d want %#x/1", got, calls, c.ReturnWord)
	}
	if !bytes.Equal(a.raw, a.before) || !bytes.Equal(a.far, a.farBefore) {
		t.Fatal("whole foreign notification/item/modifier owners changed")
	}
}

func TestTradeNotifyItemNativeOriginal(t *testing.T) {
	cases := tradeNotifyItemNativeCases(t)
	oldSend := tradePacketSendForTest
	defer func() { tradePacketSendForTest = oldSend }()
	for route := 0; route < 3; route++ {
		for _, placement := range []int{0, 64, 512} {
			raw, free := alloc.MakeHigh([]byte{}, 65536)
			t.Cleanup(free)
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&raw[0])) <= math.MaxUint32 {
				t.Fatal("foreign notification/item owner must exceed4GiB")
			}
			a := tradeNotifyItemNativeArenaAt(raw, nil, placement)
			for _, c := range cases {
				t.Run(fmt.Sprintf("route%d/place%d/%s-%s", route, placement, c.Input.Entry, c.Input.Name), func(t *testing.T) { a.run(t, route, c) })
			}
		}
	}
}

func TestTradeNotifyItemNativeOwnerIdentity(t *testing.T) {
	near, far, free := tradeNotifyItemNativeOwnerPair()
	defer free()
	a := tradeNotifyItemNativeArenaAt(near, far, 0)
	first, second := uintptr(unsafe.Pointer(a.owners[0])), uintptr(unsafe.Pointer(a.owners[1]))
	if first == second {
		t.Fatal("distinct owner records required")
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && (first <= math.MaxUint32 || uint64(second-first) != uint64(1)<<32 || uint32(first) != uint32(second)) {
		t.Fatal("owned sparse pair must have equal low words and distinct native addresses")
	}
	oldSend := tradePacketSendForTest
	defer func() { tradePacketSendForTest = oldSend }()
	for _, c := range tradeNotifyItemNativeCases(t) {
		if c.Input.Entry == "sub_50FAE0" && c.Input.RecipientRole == 0 && c.Input.OfferingRole == 1 {
			payload, err := hex.DecodeString(c.Events[0].Payload)
			if err != nil || len(payload) != 15 || payload[2] != 0 {
				t.Fatal("original distinct-owner projection")
			}
			for route := 0; route < 3; route++ {
				a.run(t, route, c)
			}
			return
		}
	}
	t.Fatal("missing sealed original distinct-owner case")
}
