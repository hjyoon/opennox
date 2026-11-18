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
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/trade_start_packet_original_v1.json.gz
var tradeStartPacketStockOriginal []byte

type tradeStartPacketParty struct {
	Player    bool
	Name      string
	Display   string
	Type      uint16
	Recipient uint8
	ShopText  string `json:"shop_text"`
}

type tradeStartPacketCase struct {
	Input struct {
		Name      string
		Entry     string
		Recipient int
		Parties   [2]tradeStartPacketParty
	}
	Observed struct {
		Defined string
		Result  uint32
		Events  []struct {
			Kind     string
			Index    int
			Packet   string
			Remove   int
			Related  int
			Sequence int
			Result   uint32
		}
	}
}

func TestTradeStartPacketNativeOriginal(t *testing.T) {
	compressed, err := gzip.NewReader(bytes.NewReader(tradeStartPacketStockOriginal))
	if err != nil {
		t.Fatal(err)
	}
	defer compressed.Close()
	raw, err := io.ReadAll(compressed)
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(raw)); got != "feabc2e3a19c3fe83419a492c0dd889946095a0eaf0a96da95d4b1fb3e26739d" {
		t.Fatalf("source digest %s", got)
	}
	var stock struct{ Rows []tradeStartPacketCase }
	if err := json.Unmarshal(raw, &stock); err != nil {
		t.Fatal(err)
	}
	if len(stock.Rows) != 80 {
		t.Fatalf("fixture cases %d", len(stock.Rows))
	}

	strings := strman.New()
	world := server.NewHigh(nil, nil, strings)
	defer world.Close()
	var entries []strman.Entry
	for i, row := range stock.Rows {
		for side, party := range row.Input.Parties {
			id := strman.ID("NPC:" + tradeStartPacketObjectID(i, side))
			entries = append(entries, strman.Entry{ID: id, Vals: []strman.Variant{{Str: party.Display}}})
		}
	}
	data, err := json.Marshal(struct {
		Entries []strman.Entry `json:"entries"`
	}{Entries: entries})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "trade-start.csf.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	if err := strings.ReadJSON(path); err != nil {
		t.Fatal(err)
	}
	game4ConsumerServices(t, game4TradeServices{world: world})
	for route := range 2 {
		t.Run(fmt.Sprintf("route%d", route), func(t *testing.T) {
			for i, row := range stock.Rows {
				t.Run(row.Input.Name, func(t *testing.T) {
					checkTradeStartPacketCase(t, row, world, i, route)
				})
			}
		})
	}
}

func tradeStartPacketObjectID(caseIndex, side int) string {
	return fmt.Sprintf("TradeStartPacket%dSide%d", caseIndex, side)
}

func checkTradeStartPacketCase(t *testing.T, row tradeStartPacketCase, world *server.Server, index, route int) {
	in, want := row.Input, row.Observed
	event := want.Events[len(want.Events)-1]
	expected, e1 := hex.DecodeString(event.Packet)
	mask, e2 := hex.DecodeString(want.Defined)
	if e1 != nil || e2 != nil || len(mask) != len(expected) || event.Kind != "send" {
		t.Fatal("invalid expected packet")
	}
	for i, defined := range mask {
		if defined == 0 {
			// The original left these positions unwritten. The Go packet uses
			// deterministic zero padding.
			expected[i] = 0
		}
	}

	var participants [2]*server.Object
	for i, party := range in.Parties {
		var unit *server.Object
		if party.Player {
			var player *server.Player
			unit, player = game4ConsumerPlayer(t, 0)
			player.SetName(party.Name)
			player.PlayerInd = party.Recipient
		} else {
			init := game4ConsumerValue(t, server.ShopkeeperInitData{})
			copy(init.ShopText[:], party.ShopText)
			unit = game4ConsumerValue(t, server.Object{ObjClass: object.ClassMonster, InitData: unsafe.Pointer(init)})
		}
		id, free := alloc.CString(tradeStartPacketObjectID(index, i))
		t.Cleanup(free)
		unit.IDPtr = unsafe.Pointer(id)
		unit.TypeInd = party.Type
		participants[i] = unit
	}
	session := game4ConsumerValue(t, server.TradeSession{Field8: participants[0], Field12: participants[1]})
	calls := 0
	world.NetSendPacketXxx = func(index int, packet []byte, related *server.Object, remove, sequence int) int {
		calls++
		if index != event.Index || remove != event.Remove || sequence != event.Sequence || related != nil ||
			event.Related != 0 || !bytes.Equal(packet, expected) {
			t.Fatalf("index %d remove %d sequence %d packet %x, want index %d and %x", index, remove, sequence, packet, event.Index, expected)
		}
		return int(int32(event.Result))
	}
	var shop int
	recipient := participants[in.Recipient]
	if in.Entry == "shop_start" {
		shop = 1
		recipient = nil
	}
	if got := tradeStartPacketFixtureCall(shop, route, recipient, session); got != want.Result || calls != 1 {
		t.Fatalf("result %#x and %d sends, want %#x and one send", got, calls, want.Result)
	}
}
