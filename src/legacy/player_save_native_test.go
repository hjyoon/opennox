package legacy

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

func setPlayerSaveTestFlags(t *testing.T, flags noxflags.GameFlag) {
	t.Helper()
	previous := noxflags.GetGame()
	noxflags.ResetGame()
	noxflags.SetGame(flags)
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(previous)
	})
}

func writePlayerSaveTestPayload(t *testing.T, write func(*cryptfile.CryptFile) error) []byte {
	t.Helper()
	path := filepath.Join(t.TempDir(), "player-section.bin")
	cf, err := cryptfile.OpenFile(path, cryptfile.WriteOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	if err := write(cf); err != nil {
		_ = cf.Close()
		t.Fatal(err)
	}
	if err := cf.Close(); err != nil {
		t.Fatal(err)
	}
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return payload
}

func readPlayerSaveTestPayload(t *testing.T, payload []byte, read func(*cryptfile.CryptFile) error) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "player-section.bin")
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	cf, err := cryptfile.OpenFile(path, cryptfile.ReadOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer cf.Close()
	if err := read(cf); err != nil {
		t.Fatal(err)
	}
}

func TestPlayerAttribWriteNative41A590ExactVersion5Payload(t *testing.T) {
	setPlayerSaveTestFlags(t, noxflags.GameModeCoop)
	info := &server.PlayerInfo{}
	info.SetName("AX")
	raw := unsafe.Slice((*byte)(info.C()), playerInfoPayloadSize41A590)
	for i := 50; i < 89; i++ {
		raw[i] = byte(i)
	}
	pl := &server.Player{QuestStage: 0x55667788}
	update := &server.PlayerUpdateData{ExtraLives: 0x11223344, Player: pl}
	unit := &server.Object{UpdateData: unsafe.Pointer(update)}

	got := writePlayerSaveTestPayload(t, func(cf *cryptfile.CryptFile) error {
		return playerAttribWriteNative41A590(cf, unit, info)
	})
	want, err := hex.DecodeString(
		"0500" +
			"01000000" +
			"02" +
			"41005800" +
			"32333435363738393a3b3c3d3e3f404142434445464748494a4b4c4d4e4f505152535455565758" +
			"44332211" +
			"88776655",
	)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("attribute payload = %x, want %x", got, want)
	}
}

func TestPlayerAttribReadNative41A590UsesNativePlayerLinks(t *testing.T) {
	setPlayerSaveTestFlags(t, noxflags.GameModeCoop)
	sourceInfo := &server.PlayerInfo{}
	sourceInfo.SetName("AX")
	sourceRaw := unsafe.Slice((*byte)(sourceInfo.C()), playerInfoPayloadSize41A590)
	for i := 50; i < 89; i++ {
		sourceRaw[i] = byte(i)
	}
	sourcePlayer := &server.Player{QuestStage: 7}
	sourceUpdate := &server.PlayerUpdateData{ExtraLives: 3, Player: sourcePlayer}
	sourceUnit := &server.Object{UpdateData: unsafe.Pointer(sourceUpdate)}
	payload := writePlayerSaveTestPayload(t, func(cf *cryptfile.CryptFile) error {
		return playerAttribWriteNative41A590(cf, sourceUnit, sourceInfo)
	})

	targetInfo := &server.PlayerInfo{}
	targetPlayer := &server.Player{
		PlayerInd:           4,
		ProtPlayerOrigName:  0x10,
		ProtPlayerField2239: 0x20,
		ProtPlayerField2235: 0x30,
		ProtPlayerClass:     0x40,
		QuestStage:          99,
	}
	targetUpdate := &server.PlayerUpdateData{Player: targetPlayer, ExtraLives: 99}
	targetUnit := &server.Object{NetCode: 0x5566, UpdateData: unsafe.Pointer(targetUpdate)}
	var events []string
	readPlayerSaveTestPayload(t, payload, func(cf *cryptfile.CryptFile) error {
		return playerAttribReadNative41A590(cf, targetUnit, targetInfo, playerAttribReadHooks41A590{
			questMode: func() bool { return false },
			disconnect: func(ind byte) {
				events = append(events, fmt.Sprintf("disconnect:%d", ind))
			},
			protectName: func(token uint32, name []byte) {
				events = append(events, fmt.Sprintf("name:%x:%x", token, name))
			},
			protectInt: func(token, value uint32) {
				events = append(events, fmt.Sprintf("int:%x:%x", token, value))
			},
			protectClass: func(token uint32, class byte) {
				events = append(events, fmt.Sprintf("class:%x:%x", token, class))
			},
			initColors: func(netCode uint32) {
				events = append(events, fmt.Sprintf("colors:%x", netCode))
			},
			maxExtraLives: func() int32 {
				events = append(events, "maximum")
				return 9
			},
			resetQuest: func(unit *server.Object) {
				if unit != targetUnit {
					t.Fatalf("reset unit = %p, want %p", unit, targetUnit)
				}
				events = append(events, "reset")
			},
			sendQuestStage: func(player *server.Player) {
				if player != targetPlayer {
					t.Fatalf("stage player = %p, want %p", player, targetPlayer)
				}
				events = append(events, fmt.Sprintf("stage:%d", player.QuestStage))
			},
		})
	})

	targetRaw := unsafe.Slice((*byte)(targetInfo.C()), playerInfoPayloadSize41A590)
	if !bytes.Equal(targetRaw[:89], sourceRaw[:89]) {
		t.Fatalf("attribute payload = %x, want %x", targetRaw[:89], sourceRaw[:89])
	}
	if targetPlayer.Name() != "AX" || targetUpdate.ExtraLives != 3 || targetPlayer.QuestStage != 7 {
		t.Fatalf("player state = name %q lives %d stage %d", targetPlayer.Name(), targetUpdate.ExtraLives, targetPlayer.QuestStage)
	}
	wantEvents := []string{
		"name:10:41005800",
		"int:20:39383736",
		"int:30:35343332",
		"class:40:42",
		"colors:5566",
		"maximum",
		"reset",
		"stage:7",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
}

func TestPlayerJournalWriteNative41BEC0OldestFirst(t *testing.T) {
	setPlayerSaveTestFlags(t, noxflags.GameModeCoop)
	newest := &server.PlayerJournal{Field3: 0x5678}
	copy(newest.EntryBuf[:], "new")
	oldest := &server.PlayerJournal{Field3: 0x1234, Prev: newest}
	copy(oldest.EntryBuf[:], "old")
	newest.Next = oldest
	pl := &server.Player{Journal: newest}
	update := &server.PlayerUpdateData{Player: pl}
	unit := &server.Object{UpdateData: unsafe.Pointer(update)}

	got := writePlayerSaveTestPayload(t, func(cf *cryptfile.CryptFile) error {
		return playerJournalWriteNative41BEC0(cf, unit, &server.PlayerInfo{})
	})
	want, err := hex.DecodeString("0100010200036f6c643412036e65777856")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("journal payload = %x, want %x", got, want)
	}
}

func TestPlayerInventoryQuestLimits41AC30(t *testing.T) {
	if !playerInventoryQuestLimits41AC30(false, func(string) int32 {
		t.Fatal("non-player limit check read inventory")
		return 0
	}, 0) {
		t.Fatal("non-player was rejected")
	}

	counts := make(map[string]int32)
	for _, name := range playerInventoryQuestStackNames41AC30 {
		counts[name] = 9
	}
	counts["InfinitePainWand"] = 3
	count := func(name string) int32 { return counts[name] }
	if !playerInventoryQuestLimits41AC30(true, count, 3) {
		t.Fatal("GAME.EXE boundary inventory was rejected")
	}
	counts["ShieldPotion"] = 10
	if playerInventoryQuestLimits41AC30(true, count, 3) {
		t.Fatal("ten quest potions were accepted")
	}
	counts["ShieldPotion"] = 9
	counts["InfinitePainWand"] = 4
	if playerInventoryQuestLimits41AC30(true, count, 3) {
		t.Fatal("staff count above the balance limit was accepted")
	}
}

func TestPlayerAbilityCooldownStart41B9C0(t *testing.T) {
	if got := playerAbilityCooldownStart41B9C0(false); got != 1 {
		t.Fatalf("inactive Berserker start = %d, want 1", got)
	}
	if got := playerAbilityCooldownStart41B9C0(true); got != 2 {
		t.Fatalf("active Berserker start = %d, want 2", got)
	}
}

func TestPlayerMapNamePayload41C080FixedBuffer(t *testing.T) {
	pl := &server.Player{}
	for i := range pl.SaveNameBuf {
		pl.SaveNameBuf[i] = 0xcc
	}
	got, err := playerMapNamePayload41C080(pl, "War01a")
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{'W', 'a', 'r', '0', '1', 'a', 0, 0xcc, 0xcc, 0xcc, 0xcc, 0xcc}
	if !bytes.Equal(got, want) {
		t.Fatalf("map-name payload = %x, want %x", got, want)
	}
	if len(pl.SaveNameBuf) != 32 {
		t.Fatalf("map-name buffer = %d bytes, want 32", len(pl.SaveNameBuf))
	}
	if delta := unsafe.Offsetof(pl.Field4792) - unsafe.Offsetof(pl.SaveNameBuf); delta != 32 {
		t.Fatalf("Field4792 follows map-name buffer by %d bytes, want 32", delta)
	}
	if _, err := playerMapNamePayload41C080(pl, "12345678901234567"); err == nil {
		t.Fatal("17-byte doubled map name fit in the 32-byte GAME.EXE buffer")
	}
}
