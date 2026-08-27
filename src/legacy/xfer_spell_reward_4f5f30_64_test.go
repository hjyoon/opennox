//go:build amd64 || arm64

package legacy

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

func TestSpellRewardNativeLayout4F5F30(t *testing.T) {
	if got := unsafe.Offsetof(server.Object{}.UseData); got != 848 {
		t.Fatalf("Object.UseData offset = %d, want native 848 (PE32 was 736)", got)
	}
	if got := unsafe.Sizeof(server.SpellRewardUseData{}); got != 1 {
		t.Fatalf("SpellRewardUseData size = %d, want 1", got)
	}
}

func TestRWSpellRewardPayload4F5F30ModernRead(t *testing.T) {
	cf := openSpellRewardReadFixture4F5F30(t, spellRewardNameFixture4F5F30("SPELL_BLINK"))
	got, err := rwSpellRewardPayload4F5F30(cf, byte(spell.SPELL_ANCHOR), 60)
	if err != nil {
		t.Fatal(err)
	}
	if got != byte(spell.SPELL_BLINK) {
		t.Fatalf("spell = %d, want %d", got, spell.SPELL_BLINK)
	}
}

func TestRWSpellRewardPayload4F5F30LegacyNamePrecedence(t *testing.T) {
	tests := []struct {
		name  string
		third string
		want  spell.ID
	}{
		{name: "third overrides second", third: "SPELL_BURN", want: spell.SPELL_BURN},
		{name: "unknown third preserves second", third: "SPELL_NOT_IN_NOX", want: spell.SPELL_BLINK},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var payload []byte
			payload = append(payload, spellRewardNameFixture4F5F30("SPELL_ANCHOR")...)
			payload = append(payload, spellRewardNameFixture4F5F30("SPELL_BLINK")...)
			payload = append(payload, spellRewardNameFixture4F5F30(tc.third)...)
			cf := openSpellRewardReadFixture4F5F30(t, payload)
			got, err := rwSpellRewardPayload4F5F30(cf, 0, 40)
			if err != nil {
				t.Fatal(err)
			}
			if got != byte(tc.want) {
				t.Fatalf("spell = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestRWSpellRewardPayload4F5F30LegacyRawVersion10(t *testing.T) {
	cf := openSpellRewardReadFixture4F5F30(t, []byte{0xaa, 0x89, byte(spell.SPELL_BLINK), 0xcc})
	got, err := rwSpellRewardPayload4F5F30(cf, 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if got != byte(spell.SPELL_BLINK) {
		t.Fatalf("spell = %d, want %d", got, spell.SPELL_BLINK)
	}
	if _, err := cf.ReadU8(); err != io.EOF {
		t.Fatalf("version 10 payload left unread bytes: %v", err)
	}
}

func TestRWSpellRewardPayload4F5F30Write(t *testing.T) {
	path := filepath.Join(t.TempDir(), "spell-reward-write.bin")
	cf, err := cryptfile.OpenFile(path, cryptfile.WriteOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := rwSpellRewardPayload4F5F30(cf, byte(spell.SPELL_BLINK), 60); err != nil {
		t.Fatal(err)
	}
	if err := cf.Close(); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := spellRewardNameFixture4F5F30("SPELL_BLINK")
	if !bytes.Equal(got, want) {
		t.Fatalf("payload = %x, want %x", got, want)
	}
}

func TestRWSpellRewardPayload4F5F30RejectsPE32NameOverflow(t *testing.T) {
	payload := append([]byte{spellRewardNameLimit4F5F30}, make([]byte, spellRewardNameLimit4F5F30)...)
	cf := openSpellRewardReadFixture4F5F30(t, payload)
	if _, err := rwSpellRewardPayload4F5F30(cf, 0, 60); err == nil {
		t.Fatal("128-byte PE32 spell-name buffer overflow was accepted")
	}
}

func spellRewardNameFixture4F5F30(name string) []byte {
	return append([]byte{byte(len(name))}, []byte(name)...)
}

func openSpellRewardReadFixture4F5F30(t *testing.T, payload []byte) *cryptfile.CryptFile {
	t.Helper()
	path := filepath.Join(t.TempDir(), "spell-reward-read.bin")
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	cf, err := cryptfile.OpenFile(path, cryptfile.ReadOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cf.Close() })
	return cf
}
