package server

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
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

//go:embed testdata/shop_exit_original_v1.json.gz
var shopExitStock50F4C0 []byte

type originalShopExitCase struct {
	Input struct {
		Name          string
		Players       [2]bool
		OldLinks      [2]*string `json:"old_links"`
		Quest         bool
		PlayerIndices [2]uint8 `json:"player_indices"`
	}
	Observed struct {
		Links       [2]*string
		Status      uint32
		CachedIndex *int `json:"cached_index"`
		Events      []struct {
			Kind     string
			Party    int
			Argument int
			Index    int
			Packet   string
			Related  int
			Remove   int
			Sequence int
			Result   uint32
		}
	}
}

func TestShopExitOriginal50F4C0(t *testing.T) {
	reader, err := gzip.NewReader(bytes.NewReader(shopExitStock50F4C0))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	raw, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(raw)); got != "7e2b0f20776cc59f1ef7aea3759daf1279b84931ed3d7c5d530620ddf38300fa" {
		t.Fatalf("stock fixture digest %s", got)
	}
	var stock struct{ Rows []originalShopExitCase }
	if err := json.Unmarshal(raw, &stock); err != nil {
		t.Fatal(err)
	}
	if len(stock.Rows) != 54 {
		t.Fatalf("original cases %d", len(stock.Rows))
	}
	oldFlags := noxflags.GetGame()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})
	for _, row := range stock.Rows {
		t.Run(row.Input.Name, func(t *testing.T) { checkShopExitOriginal50F4C0(t, row) })
	}
}

func checkShopExitOriginal50F4C0(t *testing.T, row originalShopExitCase) {
	in, want := row.Input, row.Observed
	if in.Quest {
		noxflags.SetGame(noxflags.GameModeQuest)
	} else {
		noxflags.UnsetGame(noxflags.GameModeQuest)
	}
	s := &Server{}
	defer s.tradeNative.close()
	session := s.newTradeSessionNative50E870(nil)
	other := s.newTradeSessionNative50E870(nil)
	session.Field0 = 0x1178ac
	role := func(name *string) *TradeSession {
		if name == nil {
			return nil
		}
		switch *name {
		case "session":
			return session
		case "other":
			return other
		default:
			t.Fatalf("unknown stock session role %q", *name)
			return nil
		}
	}
	var actors [2]*Object
	var updates [2]*PlayerUpdateData
	for i, isPlayer := range in.Players {
		class := object.ClassMonster
		if isPlayer {
			class = object.ClassPlayer
		}
		actors[i], updates[i] = shopExitActorFixture(t, class, in.PlayerIndices[i], role(in.OldLinks[i]))
	}
	session.Field8, session.Field12 = actors[0], actors[1]
	expectedSlots := [questShopSessionSlots50E8F0]*TradeSession{}
	for i := range expectedSlots {
		if i%2 == 0 {
			expectedSlots[i] = other
		}
	}
	s.tradeNative.questShopSession = expectedSlots
	frozen, sent := 0, 0
	unfreeze := func(unit *Object, argument int) {
		frozen++
		if want.Events[0].Kind != "unfreeze" || unit != actors[want.Events[0].Party] || argument != want.Events[0].Argument ||
			session.Field0 != want.Status || frozen != 1 || sent != 0 {
			t.Fatal("unfreeze order or participant")
		}
		for i, update := range updates {
			if update.Trade70 != role(want.Links[i]) {
				t.Fatalf("participant %d link not cleared", i)
			}
		}
	}
	s.NetSendPacketXxx = func(index int, packet []byte, related *Object, remove, sequence int) int {
		sent++
		event := want.Events[1]
		if event.Kind != "send" || event.Index != index || event.Packet != hex.EncodeToString(packet) ||
			frozen != 1 || sent != 1 || related != nil || event.Related != 0 ||
			remove != event.Remove || sequence != event.Sequence {
			t.Fatal("unfreeze and close packet order or payload")
		}
		if s.tradeNative.questShopSession != expectedSlots {
			t.Fatal("cache changed before close packet")
		}
		return int(int32(event.Result))
	}

	if closed := s.CloseShopNative50F4C0(session, unfreeze); !closed || frozen != 1 || sent != 1 {
		t.Fatalf("shop close = %t, unfreeze calls = %d, sends = %d", closed, frozen, sent)
	}
	if want.CachedIndex != nil {
		expectedSlots[*want.CachedIndex] = session
	}
	if s.tradeNative.questShopSession != expectedSlots || s.IsTradeSessionNative(session) != (want.CachedIndex != nil) ||
		!s.IsTradeSessionNative(other) {
		t.Fatal("incorrect cached session or release")
	}
}

func shopExitActorFixture(t *testing.T, class object.Class, index uint8, session *TradeSession) (*Object, *PlayerUpdateData) {
	t.Helper()
	player, freePlayer := alloc.NewHigh(Player{})
	player.PlayerInd = index
	t.Cleanup(freePlayer)
	update, freeUpdate := alloc.NewHigh(PlayerUpdateData{})
	update.Player, update.Trade70 = player, session
	t.Cleanup(freeUpdate)
	unit, freeUnit := alloc.NewHigh(Object{})
	unit.ObjClass, unit.UpdateData = class, unsafe.Pointer(update)
	t.Cleanup(freeUnit)
	return unit, update
}

func TestShopExitKeepsReplacementSession50F4C0(t *testing.T) {
	flags := noxflags.GetGame()
	noxflags.SetGame(noxflags.GameModeQuest)
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(flags)
	})
	s := &Server{}
	defer s.tradeNative.close()
	session := s.newTradeSessionNative50E870(nil)
	original := s.tradeNative.sessions[session]
	unit, update := shopExitActorFixture(t, object.ClassPlayer, 3, session)
	session.Field8 = unit
	replacement := &nativeTradeSessionAllocation{freeSession: original.freeSession, items: original.items}
	s.NetSendPacketXxx = func(_ int, _ []byte, _ *Object, _, _ int) int {
		s.tradeNative.sessions[session] = replacement
		return 1
	}
	if !s.CloseShopNative50F4C0(session, nil) || s.tradeNative.sessions[session] != replacement ||
		s.tradeNative.questShopSession[3] != nil || update.Trade70 != nil {
		t.Fatal("closing notification changed a replacement session")
	}
}
