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
	"runtime"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/shop_packet_original_v1.json.gz
var shopPacketOriginal []byte

type shopRemovalCase struct {
	Input struct {
		Name       string
		Entry      string
		Recipient  uint8
		Code       uint32
		Items      []int
		Target     int
		PlayerSide int `json:"player_side"`
	}
	Observed struct {
		Result    uint32
		Remaining []int
		Released  int
		Events    []struct {
			Service  string
			Sequence int
			Index    int
			Packet   string
			Related  int
			Remove   int
			Result   uint32
		}
	}
}

func TestShopRemovalNativeOriginal(t *testing.T) {
	z, err := gzip.NewReader(bytes.NewReader(shopPacketOriginal))
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	raw, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(raw)); got != "9122bfc83bb07e482b1b6ff9523686e38929aa2eda8a31a68627aaf3c60d5e37" {
		t.Fatalf("original fixture digest %s", got)
	}
	var original struct{ Rows []shopRemovalCase }
	if err := json.Unmarshal(raw, &original); err != nil {
		t.Fatal(err)
	}
	if len(original.Rows) != 124 {
		t.Fatalf("original cases %d", len(original.Rows))
	}
	entries := map[string]int{"shop_remove": 0, "trade_close": 1, "missing_gold": 2, "shop_unlink": 3}
	for route := 0; route < 2; route++ {
		t.Run(fmt.Sprintf("route%d", route), func(t *testing.T) {
			for _, row := range original.Rows {
				t.Run(row.Input.Name, func(t *testing.T) {
					entry, ok := entries[row.Input.Entry]
					if !ok {
						t.Fatal(row.Input.Entry)
					}
					checkShopRemovalCase(t, row, entry, route)
				})
			}
		})
	}
}

func checkShopRemovalCase(t *testing.T, row shopRemovalCase, entry, route int) {
	in, want := row.Input, row.Observed
	owner, player := game4ConsumerPlayer(t, 0)
	player.PlayerInd = in.Recipient
	item := game4ConsumerValue(t, server.Object{NetCode: in.Code})
	world := &server.Server{}
	var session *server.TradeSession
	var nodes []*server.TradeItem
	var detached *server.Object
	released := false
	if entry == 3 {
		world = server.NewHigh(nil, nil, strman.New())
		if !world.Objs.InitHigh(8) {
			t.Fatal("object pool")
		}
		defer world.Objs.FreeObjects()
		defer world.FreeObjectTypes()
		defer world.Close()
		world.FreeObjectTypes()
		price := 10
		if err := world.Types.ReadObjectType(&things.Thing{Name: "NativeShopRemovalItem", Price: &price}); err != nil {
			t.Fatal(err)
		}
		init := game4ConsumerValue(t, server.ShopkeeperInitData{Count: 1, BuyMultiplier: 1, SellMultiplier: 1})
		init.Items[0].TypeInd = uint32(world.Types.IndByID("NativeShopRemovalItem"))
		init.Items[0].Count = byte(len(in.Items))
		merchant := game4ConsumerValue(t, server.Object{ObjClass: object.ClassMonster, InitData: unsafe.Pointer(init)})
		session = world.NewShopSessionNative50E8F0(owner, merchant)
		if in.PlayerSide != 0 {
			session.Field8, session.Field12 = session.Field12, session.Field8
		}
		if got, complete := world.LoadRegularShopItemsNative50E970(session, server.ShopItemLoadRuntime50E970{}); !complete || got != len(in.Items) {
			t.Fatalf("load shop items: %d, complete=%t", got, complete)
		}
		byIdentity := map[int]*server.Object{}
		allocated := make([]*server.Object, 0, len(in.Items))
		for node := session.Field20; node != nil; node = node.Field8 {
			allocated = append(allocated, node.Item0)
			id := in.Items[len(nodes)]
			node.Item0.NetCode = in.Code
			if reused := byIdentity[id]; reused != nil {
				node.Item0 = reused
			} else {
				byIdentity[id] = node.Item0
			}
			nodes = append(nodes, node)
		}
		if target := byIdentity[in.Target]; target != nil {
			item = target
		}
		if want.Released >= 0 {
			detached = allocated[want.Released]
			defer func() {
				if released {
					world.Objs.FreeObject(detached)
				}
			}()
		}
	}
	checkList := func() {
		var indexes []int
		var previous *server.TradeItem
		for node := session.Field20; node != nil; node = node.Field8 {
			index := slices.Index(nodes, node)
			if index < 0 || slices.Contains(indexes, index) || node.Field12 != previous {
				t.Fatalf("invalid native trade links at node %p", node)
			}
			indexes = append(indexes, index)
			previous = node
		}
		if !slices.Equal(indexes, want.Remaining) {
			t.Fatalf("remaining nodes %v, want %v", indexes, want.Remaining)
		}
	}
	calls := 0
	world.NetSendPacketXxx = func(index int, packet []byte, related *server.Object, remove, seq int) int {
		if calls != 0 || len(want.Events) == 0 || want.Events[0].Service != "send" {
			t.Fatal("unexpected trade send")
		}
		calls++
		released = detached != nil
		event := want.Events[0]
		if index != event.Index || seq != event.Sequence || remove != event.Remove || related != nil || hex.EncodeToString(packet) != event.Packet {
			t.Fatalf("send index %d seq %d remove %d related %p packet %x, want %+v", index, seq, remove, related, packet, event)
		}
		if session != nil {
			checkList()
		}
		return int(int32(event.Result))
	}
	game4ConsumerServices(t, game4TradeServices{world: world})
	if route != 0 && entry == 3 {
		runtime.GC()
	}
	if got := shopRemovalFixtureCall(entry, route, owner, item, session, uint16(in.Code)); got != want.Result {
		t.Fatalf("C result %#x, want %#x", got, want.Result)
	}
	if calls == 0 && len(want.Events) != 0 {
		t.Fatal("original send absent")
	}
	if session != nil {
		checkList()
	}
}
