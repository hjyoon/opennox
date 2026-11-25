package legacy

import (
	"testing"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestWOLServiceLookupPreservesRecordAndPortSemantics(t *testing.T) {
	head := wolServiceListHead()
	previousHead := *head
	flags := noxflags.GetGame()
	t.Cleanup(func() {
		*head = previousHead
		noxflags.ResetGame()
		noxflags.SetGame(flags)
	})
	records, free := alloc.NewHigh([4]wolServiceEntry{})
	defer free()
	for i := 0; i+1 < len(records); i++ {
		records[i].Next = &records[i+1]
	}
	entries := []struct{ kind, title, args string }{
		{"lad", "first", "ignored;partial;0"},
		{"LaD", "second", "ignored;;ladder.local;\t-2extra"},
		{"GAM", "gameres server", "ignored;normal.local;65537"},
		{"gam", "QUEST GAMERES SERVER", "ignored;quest.local;12345"},
	}
	for i, entry := range entries {
		copy(records[i].Kind[:], entry.kind)
		copy(records[i].Title[:], entry.title)
		copy(records[i].Args[:], entry.args)
	}
	address := CString("preset___________________________________________________________")
	defer StrFree(address)
	var port wolEndpointPort = 17
	*head = nil
	if r := sub_420230(address, &port); r != 0 || port != 17 || GoString(address)[:6] != "preset" {
		t.Fatalf("missing ladder = %d, address %q, port %d", r, GoString(address), port)
	}
	*head = &records[0]
	if r := sub_420230(address, &port); r != 1 || GoString(address) != "ladder.local" || port != 65534 {
		t.Fatalf("ladder = %d, address %q, port %d", r, GoString(address), port)
	}
	noxflags.ResetGame()
	if r := sub_420360(address, &port); r != 0 || GoString(address) != "" || port != 0 {
		t.Fatalf("disabled results address = %d, address %q, port %d", r, GoString(address), port)
	}
	noxflags.SetGame(noxflags.GameModeCoop)
	if r := sub_420360(address, &port); r != 1 || GoString(address) != "normal.local" || port != 1 {
		t.Fatalf("co-op results address = %d, address %q, port %d", r, GoString(address), port)
	}
	noxflags.SetGame(noxflags.GameModeQuest)
	if r := sub_420360(address, &port); r != 1 || GoString(address) != "quest.local" || port != 12345 {
		t.Fatalf("quest results address = %d, address %q, port %d", r, GoString(address), port)
	}
}

func TestWOLGameAndMessageCleanupFreesNativeOwners(t *testing.T) {
	games, channels := wolGameListHead(), wolChannelHead()
	oldGames, oldChannels := *games, *channels
	oldMode := Get_dword_5d4594_527988()
	Set_dword_5d4594_527988(6)
	t.Cleanup(func() {
		*games, *channels = oldGames, oldChannels
		Set_dword_5d4594_527988(oldMode)
	})
	game, _ := alloc.NewHigh(wolGameListNode{})
	gameData, _ := alloc.NewHigh([64]byte{})
	channel, freeChannel := alloc.NewHigh(wolChannel{})
	defer freeChannel()
	message, _ := alloc.NewHigh(wolChannelMessage{})
	text, _ := alloc.NewHigh([64]byte{})
	owned := []unsafe.Pointer{unsafe.Pointer(game), unsafe.Pointer(gameData), unsafe.Pointer(message), unsafe.Pointer(text)}
	defer func() {
		for _, ptr := range owned {
			alloc.TryFreePtr(ptr)
		}
	}()
	copy(gameData[36:], "Night Arena")
	game.Data = unsafe.Pointer(gameData)
	message.Text = unsafe.Pointer(text)
	channel.Messages, channel.State, channel.Count = message, 1, 1
	*games, *channels = game, channel
	if got := sub_41F800(internCStr("Night Arena")); got != game.Data || sub_41F800(internCStr("night arena")) != nil {
		t.Fatalf("game lookup = %p, want %p", got, game.Data)
	}
	if r := sub_41F4B0(); r != 6 || *games != nil {
		t.Fatalf("game list cleanup = %d, retained list %p", r, *games)
	}
	if r := sub_41EC30(); r != 6 || *channels != channel || channel.Messages != nil || channel.State != 0 || channel.Count != 0 {
		t.Fatalf("channel cleanup = %d, channel %+v", r, *channel)
	}
	for _, ptr := range owned {
		if alloc.TryFreePtr(ptr) {
			t.Fatalf("retained native allocation %p", ptr)
		}
	}
}
