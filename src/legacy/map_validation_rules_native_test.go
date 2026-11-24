package legacy

import (
	"bytes"
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/opennox/libs/maps"
	"github.com/opennox/noxcrypt"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
)

func TestMapValidationNative4CF470ChecksAlignedCRC(t *testing.T) {
	t.Chdir(t.TempDir())
	const expected = 0x81dead12
	separator := memmap.Slice(0x587000, 191672)[:2]
	old := bytes.Clone(separator)
	t.Cleanup(func() { copy(separator, old) })
	separator[0], separator[1] = '\\', 0
	if err := os.MkdirAll(filepath.Join("maps", "arena"), 0o700); err != nil {
		t.Fatal(err)
	}
	cipher, err := crypt.NewCipher(crypt.MapKey)
	if err != nil {
		t.Fatal(err)
	}
	write := func(name string, marker, crc uint32) {
		t.Helper()
		var cleartext, encoded [16]byte
		binary.LittleEndian.PutUint32(cleartext[:4], marker)
		binary.LittleEndian.PutUint32(cleartext[8:12], crc)
		for i := 0; i < len(cleartext); i += crypt.Block {
			cipher.Encrypt(encoded[i:i+crypt.Block], cleartext[i:i+crypt.Block])
		}
		if err := os.WriteFile(name, encoded[:], 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join("maps", "arena", "arena.map"), maps.Magic, expected)
	write("old.map", maps.MagicOld, expected)
	for _, row := range []struct {
		name     string
		expected uint32
		status   uint32
	}{
		{"arena.map", expected, 6},
		{"maps/arena/arena.map", expected, 6},
		{"maps\\arena\\arena.map", expected, 6},
		{"arena.map\x00ignored", expected, 6},
		{"arena.map", expected + 1, 2},
		{"./old.map", expected, 2},
		{"absent.map", 0, 6},
		{"absent.map", expected, 0},
	} {
		if got := validateMapFile4CF470(row.name, row.expected); got != row.status {
			t.Errorf("map %q CRC %#08x returned %d, want %d", row.name, row.expected, got, row.status)
		}
	}
}

func TestMapRuleFileNative4D0670KeepsModesAndNarrowBytes(t *testing.T) {
	oldCmd, oldFlags := ExecConsoleCmd, noxflags.GetGame()
	t.Cleanup(func() {
		ExecConsoleCmd = oldCmd
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameModeQuest)
	var got []string
	ExecConsoleCmd = func(ctx context.Context, cmd string) bool {
		got = append(got, cmd)
		return true
	}
	path := filepath.Join(t.TempDir(), "quest.rul")
	content := "first\r\n[DEATHMATCH]\nignore\n[QUEST]\nname \xe9\n[COMMON]\nlast"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	if !readMapRuleFile4D0670(path) || !slices.Equal(got, []string{"first", "name é", "last"}) {
		t.Fatalf("map rule commands = %q", got)
	}
	if readMapRuleFile4D0670(path + ".missing") {
		t.Fatal("missing map rules reported success")
	}
}
