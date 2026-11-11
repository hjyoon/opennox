package opennox

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet"
	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestShopSaleRejectionLocalizedLine5109C0(t *testing.T) {
	for _, tc := range []struct {
		name, quest, item string
	}{
		{"English", "You cannot sell a quest item.", "I am not interested in buying that."},
		{"Korean", "퀘스트 아이템은 팔 수 없습니다.", "이 아이템은 살 수 없습니다."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, unit, _, _ := basicSenderProductionFixture5187E0(t)
			player, freePlayer := alloc.NewHigh(server.Player{})
			t.Cleanup(freePlayer)
			*player = server.Player{PlayerInd: 3}
			update, freeUpdate := alloc.NewHigh(server.PlayerUpdateData{})
			t.Cleanup(freeUpdate)
			*update = server.PlayerUpdateData{Player: player}
			*unit = server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
			data, err := json.Marshal(struct {
				Entries []strman.Entry `json:"entries"`
			}{Entries: []strman.Entry{
				{ID: "Trade.c:CantSellQuestItem", Vals: []strman.Variant{{Str: tc.quest}}},
				{ID: "Trade.c:CantSellItem", Vals: []strman.Variant{{Str: tc.item}}},
			}})
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(t.TempDir(), "shop.csf.json")
			if err := os.WriteFile(path, data, 0600); err != nil {
				t.Fatal(err)
			}
			if err := s.Strings().ReadJSON(path); err != nil {
				t.Fatal(err)
			}
			runtime := s.shopSellRuntime5109C0()
			for _, rejection := range []struct {
				name string
				send func(*server.Object)
				want string
			}{
				{"quest item", runtime.ReportCannotSellQuest, tc.quest},
				{"other unsellable item", runtime.ReportCannotSellItem, tc.item},
			} {
				t.Run(rejection.name, func(t *testing.T) {
					s.NetList.ResetAll()
					rejection.send(unit)
					packet := s.NetList.CopyPacketsA(3, netlist.Kind1)
					if len(packet) == 0 || packet[0] != byte(netmsg.MSG_TEXT_MESSAGE) {
						t.Fatalf("sale rejection packet = %x, want original localized line message", packet)
					}
					var message noxnet.MsgText
					n, err := message.Decode(packet[1:])
					if err != nil || n != len(packet)-1 || message.NetCode != 0 || message.Text() != rejection.want {
						t.Fatalf("sale rejection = %#v, bytes=%d err=%v; want %q", message, n, err, rejection.want)
					}
					if rest := s.NetList.CopyPacketsB(3); len(rest) != 0 {
						t.Fatalf("unexpected additional rejection packets: %x", rest)
					}
				})
			}
		})
	}
}
