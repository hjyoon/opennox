package legacy

import (
	"bytes"
	"encoding/hex"
	"os"
	"path/filepath"
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
