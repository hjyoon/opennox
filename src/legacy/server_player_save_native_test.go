package legacy

import (
	"bytes"
	"path/filepath"
	"testing"

	crypt "github.com/opennox/noxcrypt"

	"github.com/opennox/opennox/v1/internal/cryptfile"
)

func TestServerPlayerSaveWirePreservesAlignedSections(t *testing.T) {
	path := filepath.Join(t.TempDir(), "player.plr")
	cf, err := cryptfile.OpenFile(path, cryptfile.WriteOnly, crypt.SaveKey)
	if err != nil {
		t.Fatal(err)
	}
	defer cf.Close()
	for _, section := range []struct {
		id      uint32
		payload []byte
	}{
		{1, []byte{0xa1, 0xa2, 0xa3}},
		{2, []byte{0xb1, 0xb2, 0xb3, 0xb4, 0xb5}},
		{7, []byte{0xc1, 0xc2}},
		{4, []byte{0xd1, 0xd2, 0xd3, 0xd4}},
		{11, []byte{0xee}},
	} {
		if err := cf.WriteU32(section.id); err != nil {
			t.Fatal(err)
		}
		cf.SectionStart()
		if _, err := cf.Write(section.payload); err != nil {
			t.Fatal(err)
		}
		cf.SectionEnd()
	}
	if err := cf.WriteU32(0); err != nil {
		t.Fatal(err)
	}
	if err := cf.Close(); err != nil {
		t.Fatal(err)
	}

	var want [53]byte
	// The original transfer kept the alignment of the source file after
	// omitting its client-only sections.
	want[0], want[5] = 2, 5
	copy(want[13:], []byte{0xb1, 0xb2, 0xb3, 0xb4, 0xb5})
	want[18], want[24] = 4, 4
	copy(want[32:], []byte{0xd1, 0xd2, 0xd3, 0xd4})
	want[36], want[40], want[48] = 11, 1, 0xee
	got, err := serverPlayerSaveWire(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want[:]) {
		t.Fatalf("legacy save transfer = %x, want %x", got, want)
	}
}
