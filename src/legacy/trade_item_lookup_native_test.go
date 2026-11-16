package legacy

import (
	"bytes"
	_ "embed"
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/trade_item_lookup_original_50ffe0_v1.json.gz
var tradeItemLookupNativeOriginal []byte

type tradeItemLookupNativeCase struct {
	Input struct {
		Name  string
		Chain []int
		Key   uint32
		Codes [8]uint32
	}
	ReturnSlot int    `json:"return_slot"`
	ReturnWord uint32 `json:"return_u32"`
}

func TestTradeItemLookupNativeLayout(t *testing.T) {
	var node server.TradeItem
	var obj server.Object
	want := [13]uintptr{unsafe.Sizeof(uintptr(0)), unsafe.Sizeof(node), unsafe.Offsetof(node.Item0), unsafe.Offsetof(node.Cost4), unsafe.Offsetof(node.Field8), unsafe.Offsetof(node.Field12), unsafe.Sizeof(node.Item0), unsafe.Sizeof(node.Cost4), unsafe.Sizeof(node.Field8), unsafe.Offsetof(obj.NetCode), unsafe.Sizeof(obj.NetCode), unsafe.Sizeof(uintptr(0)), unsafe.Sizeof(int32(0))}
	if got := tradeItemLookupNativeLayoutC(); got != want {
		t.Fatal("native list/object/net-code layout", got, want)
	}
	if unsafe.Sizeof(node) > 0x100 || unsafe.Sizeof(obj) > 0x800 {
		t.Fatal("foreign node/object stride too small")
	}
}

func TestTradeItemLookupNativeOriginal(t *testing.T) {
	var cases []tradeItemLookupNativeCase
	npcVoiceOriginalDecode(t, tradeItemLookupNativeOriginal, "984463bd2661f1023db7207a6bd9e2e63dd1ee190fd2ef6cd9fcfc4f1a8983ad", "7f5c0673ad0bd829a585bbe4e168e23f8452fc8e52288ea80587f3231b722922", 291, &cases)
	for route := 0; route < 3; route++ {
		for _, placement := range []int{0, 64, 512} {
			raw, free := alloc.MakeHigh([]byte{}, 65536)
			t.Cleanup(free)
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(&raw[0])) <= math.MaxUint32 {
				t.Fatal("foreign list/object owner must exceed4GiB")
			}
			var nodes [8]*server.TradeItem
			var objects [8]*server.Object
			for i := range nodes {
				nodes[i] = (*server.TradeItem)(unsafe.Pointer(&raw[0x1000+i*0x100+placement]))
				objects[i] = (*server.Object)(unsafe.Pointer(&raw[0x4000+i*0x800+placement]))
			}
			expected := make([]byte, len(raw))
			for _, c := range cases {
				t.Run(fmt.Sprintf("route%d/place%d/%s", route, placement, c.Input.Name), func(t *testing.T) {
					for i := range raw {
						raw[i] = byte(i*31 + 19)
					}
					for i, node := range nodes {
						// These are seeded foreign fields: use C pointer stores so
						// Go's deletion barrier never shades old canary bits.
						uiStateTradeStorePointer(unsafe.Pointer(&node.Item0), unsafe.Pointer(objects[i]))
						uiStateTradeStorePointer(unsafe.Pointer(&node.Field8), nil)
						objects[i].NetCode = c.Input.Codes[i]
					}
					var head *server.TradeItem
					var seen [8]bool
					for at, slot := range c.Input.Chain {
						if slot < 0 || slot >= len(nodes) || seen[slot] {
							t.Fatal("invalid canonical acyclic list", slot)
						}
						seen[slot] = true
						if at == 0 {
							head = nodes[slot]
						}
						if at+1 < len(c.Input.Chain) {
							next := c.Input.Chain[at+1]
							if next < 0 || next >= len(nodes) {
								t.Fatal("unowned canonical link", next)
							}
							uiStateTradeStorePointer(unsafe.Pointer(&nodes[slot].Field8), unsafe.Pointer(nodes[next]))
						}
					}
					var want *server.TradeItem
					if c.ReturnSlot != -1 {
						if c.ReturnSlot < 0 || c.ReturnSlot >= len(nodes) || !seen[c.ReturnSlot] || c.ReturnWord != 0x21001000+uint32(c.ReturnSlot)*0x100 {
							t.Fatal("unowned original return identity")
						}
						want = nodes[c.ReturnSlot]
					} else if c.ReturnWord != 0 {
						t.Fatal("original nil return word")
					}
					copy(expected, raw)
					if got := tradeItemLookupNativeCall(route, head, c.Input.Key); got != want {
						t.Fatalf("original first-match identity: got %p want %p", got, want)
					}
					if !bytes.Equal(raw, expected) {
						t.Fatal("unexpected write to whole64KiB foreign list/object owner")
					}
				})
			}
		}
	}
}
