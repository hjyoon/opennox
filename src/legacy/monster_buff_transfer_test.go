package legacy

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

func TestMonsterBuffTransferUsesOriginalOrder(t *testing.T) {
	obj := &server.Object{Buffs: 1<<server.ENCHANT_INVISIBLE |
		1<<server.ENCHANT_INVULNERABLE | 1<<server.ENCHANT_CHARMING |
		1<<server.ENCHANT_CROWN | 1<<server.ENCHANT_SNEAK}
	obj.BuffsPower[server.ENCHANT_INVULNERABLE] = 0x33
	obj.BuffsDur[server.ENCHANT_INVULNERABLE] = 0x4321
	obj.BuffsPower[server.ENCHANT_INVISIBLE] = 0x44
	obj.BuffsDur[server.ENCHANT_INVISIBLE] = 0x8765

	path := filepath.Join(t.TempDir(), "monster-buffs.bin")
	file, err := cryptfile.OpenFile(path, cryptfile.WriteOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = file.Close() })
	if err := monsterXferBuffs52AAB0(file, obj); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	// Stock saves put Invulnerable before Invisible. The other three
	// active flags above are outside the original 29-entry save list.
	const invulnerable, invisible = "ENCHANT_INVULNERABLE", "ENCHANT_INVISIBLE"
	want := append([]byte{2, 0, 2, byte(len(invulnerable))}, invulnerable...)
	want = append(want, 0x33, 0x21, 0x43, 0, 0, byte(len(invisible)))
	want = append(want, invisible...)
	want = append(want, 0x44, 0x65, 0x87, 0, 0)
	if !bytes.Equal(got, want) {
		t.Fatalf("saved buffs = %x, want %x", got, want)
	}
}
