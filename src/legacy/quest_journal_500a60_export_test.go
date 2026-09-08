package legacy

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/internal/cryptfile"
)

func questJournalWriteExportWithGlobal500A60(t *testing.T, cf *cryptfile.CryptFile) int32 {
	t.Helper()
	previous := cryptfile.Global()
	cryptfile.SetGlobal(cf)
	defer cryptfile.SetGlobal(previous)
	return questJournalWriteExportCall500A60()
}

func TestQuestJournalWriteExport500A60ReturnsCanonicalSuccess(t *testing.T) {
	resetQuestJournal500540(t)
	setPlayerSaveTestFlags(t, noxflags.GameModeCoop)
	if questJournalSet500540("War01a:Count", 0, 0x12345678) == nil {
		t.Fatal("cannot create quest-journal entry")
	}

	path := filepath.Join(t.TempDir(), "quest-journal-export.bin")
	cf, err := cryptfile.OpenFile(path, cryptfile.WriteOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	if got := questJournalWriteExportWithGlobal500A60(t, cf); got != 1 {
		_ = cf.Close()
		t.Fatalf("export result = %d, want canonical success 1", got)
	}
	if err := cf.Close(); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{
		0x01, 0x00,
		0x01, 0x00, 0x00, 0x00,
		0x0c, 'W', 'a', 'r', '0', '1', 'a', ':', 'C', 'o', 'u', 'n', 't',
		0x00, 0x00, 0x00, 0x00,
		0x78, 0x56, 0x34, 0x12,
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("export payload = %x, want %x", got, want)
	}
}

func TestQuestJournalWriteExport500A60ReturnsCanonicalFailure(t *testing.T) {
	resetQuestJournal500540(t)

	path := filepath.Join(t.TempDir(), "quest-journal-export-read.bin")
	if err := os.WriteFile(path, []byte{0x02, 0x00}, 0o600); err != nil {
		t.Fatal(err)
	}
	cf, err := cryptfile.OpenFile(path, cryptfile.ReadOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer cf.Close()
	if got := questJournalWriteExportWithGlobal500A60(t, cf); got != 0 {
		t.Fatalf("export result = %d, want canonical failure 0", got)
	}
}
