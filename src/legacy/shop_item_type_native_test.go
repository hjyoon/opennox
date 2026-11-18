package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"testing"

	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/shop_type_original_v1.json.gz
var shopItemTypeOriginal []byte

func TestShopItemTypeNativeOriginal(t *testing.T) {
	compressed, err := gzip.NewReader(bytes.NewReader(shopItemTypeOriginal))
	if err != nil {
		t.Fatal(err)
	}
	defer compressed.Close()
	raw, err := io.ReadAll(compressed)
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(raw)); got != "a5c04297bd0e8bcf7c320af9335c199363d2cccc7204ad7f44f1e70a3296f00c" {
		t.Fatalf("original fixture digest %s", got)
	}
	var stock struct {
		Rows []struct {
			Input struct {
				Name    string
				Entry   string
				Quest   bool
				TypeInd uint16 `json:"type_ind"`
				Cache   [7]uint32
				IDs     [4]uint32
				Null    bool
			}
			Observed struct {
				Result int
				Cache  [7]uint32
				Calls  []string
			}
		}
	}
	if err := json.Unmarshal(raw, &stock); err != nil {
		t.Fatal(err)
	}
	if len(stock.Rows) != 102 {
		t.Fatalf("original cases %d", len(stock.Rows))
	}
	game := noxflags.GetGame()
	t.Cleanup(func() { noxflags.ResetGame(); noxflags.SetGame(game) })
	cache := memmap.PtrT[[9]uint32](shopTypeCacheBase, questShopTypeOff-4)
	old := *cache
	t.Cleanup(func() { *cache = old })
	worlds := map[[4]uint32]*server.Server{}
	typeNames := [...]string{"Diamond", "Emerald", "Ruby", "AnkhTradable"}
	for _, row := range stock.Rows {
		ids := row.Input.IDs
		if worlds[ids] != nil {
			continue
		}
		world := server.NewHigh(nil, nil, strman.New())
		world.FreeObjectTypes()
		t.Cleanup(func() { world.FreeObjectTypes(); world.Close() })
		names := [...]string{"", "ShopFixtureType1", "ShopFixtureType2", "ShopFixtureType3", "ShopFixtureType4"}
		for i, id := range ids {
			if id != 0 {
				names[id] = typeNames[i]
			}
		}
		for _, name := range names[1:] {
			if err := world.Types.ReadObjectType(&things.Thing{Name: name}); err != nil {
				t.Fatal(err)
			}
		}
		worlds[ids] = world
	}
	for route := 0; route < 2; route++ {
		t.Run(fmt.Sprintf("route%d", route), func(t *testing.T) {
			for _, row := range stock.Rows {
				in, want := row.Input, row.Observed
				t.Run(in.Name, func(t *testing.T) {
					noxflags.ResetGame()
					if in.Quest {
						noxflags.SetGame(noxflags.GameModeQuest)
					}
					gem, flagCalls := 0, 1
					if in.Entry == "is_gem" {
						gem, flagCalls = 1, 0
					}
					var world *server.Server
					if len(want.Calls) > flagCalls {
						world = worlds[in.IDs]
					}
					game4ConsumerServices(t, game4TradeServices{world: world})
					cache[0], cache[8] = 0x359a694e, 0xab00be44
					copy(cache[1:8], in.Cache[:])
					var item *server.Object
					initial := server.Object{TypeInd: in.TypeInd, NetCode: 0xfeed9184, ObjFlags: 0x35000}
					if !in.Null {
						item = game4ConsumerValue(t, initial)
					}
					if got := shopItemTypeFixtureCall(item, gem, route); got != want.Result {
						t.Fatalf("C result %d, want %d", got, want.Result)
					}
					if [7]uint32(cache[1:8]) != want.Cache || cache[0] != 0x359a694e || cache[8] != 0xab00be44 {
						t.Fatalf("native cache %v, want %v with intact guards", cache, want.Cache)
					}
					if item != nil && *item != initial {
						t.Fatal("shop type predicate changed object owner")
					}
				})
			}
		})
	}
}
