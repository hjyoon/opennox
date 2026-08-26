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
)

func resetQuestJournal500540(t *testing.T) {
	t.Helper()
	questJournalDelete5007E0("*:*")
	t.Cleanup(func() {
		questJournalDelete5007E0("*:*")
	})
}

func readQuestJournalTestPayload(t *testing.T, payload []byte) error {
	t.Helper()
	path := filepath.Join(t.TempDir(), "quest-journal.bin")
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	cf, err := cryptfile.OpenFile(path, cryptfile.ReadOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer cf.Close()
	return questJournalReadNative500B70(cf)
}

func TestQuestJournalNativeLayout500540(t *testing.T) {
	resetQuestJournal500540(t)
	entry := questJournalSet500540("War01a:Layout", 0, 0)
	if entry == nil {
		t.Fatal("cannot allocate quest-journal entry")
	}

	wantNext, wantPrev, wantSize := uintptr(140), uintptr(144), uintptr(148)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantNext, wantPrev, wantSize = 144, 152, 160
	}
	if got := unsafe.Offsetof(entry.kind); got != 132 {
		t.Fatalf("kind offset = %d, want 132", got)
	}
	if got := unsafe.Offsetof(entry.value); got != 136 {
		t.Fatalf("value offset = %d, want 136", got)
	}
	if got := unsafe.Offsetof(entry.next); got != wantNext {
		t.Fatalf("next offset = %d, want %d", got, wantNext)
	}
	if got := unsafe.Offsetof(entry.prev); got != wantPrev {
		t.Fatalf("prev offset = %d, want %d", got, wantPrev)
	}
	if got := unsafe.Sizeof(*entry); got != wantSize {
		t.Fatalf("native entry size = %d, want %d", got, wantSize)
	}
}

func TestQuestJournalWriteNative500A60ExactVersion1Payload(t *testing.T) {
	resetQuestJournal500540(t)
	setPlayerSaveTestFlags(t, noxflags.GameModeCoop)
	if questJournalSet500540("War01a:Count", 0, 0x89abcdef) == nil {
		t.Fatal("cannot create numeric entry")
	}
	if questJournalSet500540("War01a:Open", 1, 1) == nil {
		t.Fatal("cannot create boolean entry")
	}

	got := writePlayerSaveTestPayload(t, questJournalWriteNative500A60)
	want, err := hex.DecodeString(
		"0100" +
			"02000000" +
			"0b5761723031613a4f70656e0100000001000000" +
			"0c5761723031613a436f756e7400000000efcdab89",
	)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("quest-journal payload = %x, want %x", got, want)
	}
}

func TestQuestJournalSet500540PreservesEntryKind(t *testing.T) {
	resetQuestJournal500540(t)
	setPlayerSaveTestFlags(t, noxflags.GameModeCoop)
	if questJournalSet500540("War01a:StableKind", 0, 3) == nil {
		t.Fatal("cannot create quest-journal entry")
	}
	if questJournalSet500540("war01A:stablekind", 1, 7) == nil {
		t.Fatal("cannot update quest-journal entry")
	}

	got := writePlayerSaveTestPayload(t, questJournalWriteNative500A60)
	want, err := hex.DecodeString(
		"010001000000" +
			"115761723031613a537461626c654b696e64" +
			"0000000007000000",
	)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("updated quest-journal payload = %x, want %x", got, want)
	}
}

func TestQuestJournalReadNative500B70RoundTrip(t *testing.T) {
	resetQuestJournal500540(t)
	setPlayerSaveTestFlags(t, noxflags.GameModeCoop)
	payload, err := hex.DecodeString(
		"0100" +
			"02000000" +
			"0c5761723031613a436f756e74000000002a000000" +
			"0b5761723031613a4f70656e0100000001000000",
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := readQuestJournalTestPayload(t, payload); err != nil {
		t.Fatal(err)
	}
	if questJournalFind5005E0("war01A:COUNT") == nil {
		t.Fatal("numeric entry was not restored case-insensitively")
	}
	if questJournalFind5005E0("WAR01A:open") == nil {
		t.Fatal("boolean entry was not restored case-insensitively")
	}

	// GAME.EXE inserts each restored entry at the list head, so a subsequent
	// write has the same values but the opposite entry order.
	got := writePlayerSaveTestPayload(t, questJournalWriteNative500A60)
	want, err := hex.DecodeString(
		"0100" +
			"02000000" +
			"0b5761723031613a4f70656e0100000001000000" +
			"0c5761723031613a436f756e74000000002a000000",
	)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("round-trip payload = %x, want %x", got, want)
	}
}

func TestQuestJournalDelete5007E0Wildcards(t *testing.T) {
	resetQuestJournal500540(t)
	add := func(name string) {
		t.Helper()
		if questJournalSet500540(name, 0, 1) == nil {
			t.Fatalf("cannot create %q", name)
		}
	}
	add("War01a:Count")
	add("War01a:Open")
	add("War02a:Count")

	questJournalDelete5007E0("War01a:*")
	if questJournalFind5005E0("War01a:Count") != nil || questJournalFind5005E0("War01a:Open") != nil {
		t.Fatal("trailing-star deletion left a War01a entry")
	}
	if questJournalFind5005E0("War02a:Count") == nil {
		t.Fatal("trailing-star deletion removed a different map entry")
	}

	add("War01a:Count")
	add("War01a:Open")
	questJournalDelete5007E0("*:Count")
	if questJournalFind5005E0("War01a:Count") != nil || questJournalFind5005E0("War02a:Count") != nil {
		t.Fatal("leading-star deletion left a Count entry")
	}
	if questJournalFind5005E0("War01a:Open") == nil {
		t.Fatal("leading-star deletion removed a different entry")
	}

	// GAME.EXE uses case-insensitive prefix matching but case-sensitive
	// strstr matching for wildcard suffixes.
	add("War01a:UpperTail")
	questJournalDelete5007E0("war01A:*Tail")
	if questJournalFind5005E0("War01a:UpperTail") != nil {
		t.Fatal("mixed-case map prefix did not match the original wildcard rule")
	}
	add("War01a:UpperTail")
	questJournalDelete5007E0("war01A:*tail")
	if questJournalFind5005E0("War01a:UpperTail") == nil {
		t.Fatal("case-sensitive wildcard suffix matched different casing")
	}

	questJournalDelete5007E0("*:*")
	if questJournalHead500540 != nil {
		t.Fatal("global wildcard did not clear the journal")
	}
}

func TestQuestJournalWriteNative500A60OmitsEntriesOutsideCoop(t *testing.T) {
	resetQuestJournal500540(t)
	setPlayerSaveTestFlags(t, 0)
	if questJournalSet500540("War01a:Count", 0, 1) == nil {
		t.Fatal("cannot create quest-journal entry")
	}
	got := writePlayerSaveTestPayload(t, questJournalWriteNative500A60)
	want, err := hex.DecodeString("010000000000")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("non-coop payload = %x, want %x", got, want)
	}
}
