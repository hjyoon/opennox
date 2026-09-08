package legacy

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/opennox/opennox/v1/internal/cryptfile"
)

func questJournalReadExportWithGlobal500B70(t *testing.T, cf *cryptfile.CryptFile) int32 {
	t.Helper()
	previous := cryptfile.Global()
	cryptfile.SetGlobal(cf)
	defer cryptfile.SetGlobal(previous)
	return questJournalReadExportCall500B70()
}

func TestQuestJournalReadExport500B70ReturnsCanonicalSuccess(t *testing.T) {
	resetQuestJournal500540(t)
	payload, err := hex.DecodeString(
		"0100" +
			"01000000" +
			"0c5761723031613a436f756e740000000078563412",
	)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "quest-journal-reader-export.bin")
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	cf, err := cryptfile.OpenFile(path, cryptfile.ReadOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer cf.Close()

	if got := questJournalReadExportWithGlobal500B70(t, cf); got != 1 {
		t.Fatalf("export result = %d, want canonical success 1", got)
	}
	if got := QuestJournalGetInt500750("War01a:Count"); got != 0x12345678 {
		t.Fatalf("export restored value = %#08x, want 0x12345678", uint32(got))
	}
}

func TestQuestJournalReadExport500B70ReturnsCanonicalFailure(t *testing.T) {
	resetQuestJournal500540(t)
	path := filepath.Join(t.TempDir(), "quest-journal-reader-export-newer.bin")
	if err := os.WriteFile(path, []byte{0x02, 0x00}, 0o600); err != nil {
		t.Fatal(err)
	}
	cf, err := cryptfile.OpenFile(path, cryptfile.ReadOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer cf.Close()

	if got := questJournalReadExportWithGlobal500B70(t, cf); got != 0 {
		t.Fatalf("export result = %d, want canonical failure 0", got)
	}
}
