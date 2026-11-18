package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/shop_stock_original_v1.json.gz
var shopStockOriginal []byte

type shopStockOriginalCase struct {
	Input struct {
		Name      string
		Entry     string
		Recipient uint8
		Items     []struct {
			Class    uint32 `json:"class_"`
			Subclass uint32
			Cost     uint32
			TypeID   uint16 `json:"type_id"`
			Code     uint32
			Health   []uint16
			Mods     []*uint32
		}
	}
	Observed struct {
		Result         uint32
		ReturnsSession bool `json:"returns_session"`
		Events         []struct {
			Index    int
			Packet   string
			Sequence int
			Related  int
			Remove   int
			Result   uint32
		}
	}
}

func TestShopStockNativeOriginal(t *testing.T) {
	z, err := gzip.NewReader(bytes.NewReader(shopStockOriginal))
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	raw, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(raw)); got != "e39a7e56800cfc630b7f7e02799ab419192520ee4a4a9b8eef21034ee7332eda" {
		t.Fatalf("original fixture digest %s", got)
	}
	var original struct{ Rows []shopStockOriginalCase }
	if err := json.Unmarshal(raw, &original); err != nil {
		t.Fatal(err)
	}
	if len(original.Rows) != 105 {
		t.Fatalf("original cases %d", len(original.Rows))
	}
	entries := map[string]int{"sort_key": 0, "stock_packet": 1, "stock_list": 2}
	for route := range 2 {
		t.Run(fmt.Sprintf("route%d", route), func(t *testing.T) {
			for _, row := range original.Rows {
				t.Run(row.Input.Name, func(t *testing.T) {
					entry, ok := entries[row.Input.Entry]
					if !ok {
						t.Fatal(row.Input.Entry)
					}
					checkShopStockOriginalCase(t, row, entry, route)
				})
			}
		})
	}
}

func checkShopStockOriginalCase(t *testing.T, row shopStockOriginalCase, entry, route int) {
	in, want := row.Input, row.Observed
	session := game4ConsumerValue(t, server.TradeSession{})
	nodes := make([]*server.TradeItem, len(in.Items))
	values := make([]server.TradeItem, len(in.Items))
	for i, input := range in.Items {
		item := game4ConsumerValue(t, server.Object{
			ObjClass:    object.Class(input.Class),
			ObjSubClass: object.SubClass(input.Subclass),
			TypeInd:     input.TypeID,
			NetCode:     input.Code,
		})
		if input.Health != nil {
			item.HealthData = game4ConsumerValue(t, server.HealthData{Cur: input.Health[0], Field2: input.Health[1]})
		}
		if input.Mods != nil {
			mods := game4ConsumerValue(t, server.ModifierInitData{})
			item.InitData = unsafe.Pointer(mods)
			for slot, index := range input.Mods {
				if index != nil {
					mod := game4ConsumerValue(t, server.ModifierEff{})
					// The native index follows the leading name pointer.
					*(*uint32)(unsafe.Add(mod.C(), unsafe.Sizeof(uintptr(0)))) = *index
					mods.Modifiers[slot] = mod
				}
			}
		}
		node := game4ConsumerValue(t, server.TradeItem{Item0: item, Cost4: input.Cost})
		nodes[i] = node
		if i == 0 {
			session.Field20 = node
		} else {
			nodes[i-1].Field8 = node
			node.Field12 = nodes[i-1]
		}
	}
	for i, node := range nodes {
		values[i] = *node
	}
	originalSession := *session
	var recipient *server.Object
	if len(want.Events) != 0 {
		owner, player := game4ConsumerPlayer(t, 0)
		player.PlayerInd = in.Recipient
		recipient = owner
	}
	var calls int
	world := &server.Server{}
	world.NetSendPacketXxx = func(index int, packet []byte, related *server.Object, remove, sequence int) int {
		if calls >= len(want.Events) {
			t.Fatalf("unexpected send %d, packet %x", calls, packet)
		}
		event := want.Events[calls]
		calls++
		if index != event.Index || sequence != event.Sequence || remove != event.Remove ||
			related != nil || event.Related != 0 || hex.EncodeToString(packet) != event.Packet {
			t.Fatalf("index %d seq %d remove %d related %p packet %x, want %+v", index, sequence, remove, related, packet, event)
		}
		return int(int32(event.Result))
	}
	game4ConsumerServices(t, game4TradeServices{world: world})
	result := uintptr(want.Result)
	if want.ReturnsSession {
		result = uintptr(unsafe.Pointer(session))
	}
	if got := shopStockFixtureCall(entry, route, recipient, session); got != result {
		t.Fatalf("result %#x, want %#x", got, result)
	}
	if calls != len(want.Events) || *session != originalSession {
		t.Fatalf("send count %d want %d, session links changed %t", calls, len(want.Events), *session != originalSession)
	}
	for i, node := range nodes {
		if *node != values[i] {
			t.Fatalf("node %d changed", i)
		}
	}
}
