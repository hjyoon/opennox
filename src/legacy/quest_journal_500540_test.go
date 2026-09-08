package legacy

import (
	"bytes"
	"encoding/hex"
	"math"
	"os"
	"path/filepath"
	"strings"
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

func TestQuestJournalSet500540OriginalReturnContract(t *testing.T) {
	resetQuestJournal500540(t)

	first, result := questJournalSetResult500540("War01a:First", 0, 3)
	if first == nil {
		t.Fatal("first insertion did not allocate an entry")
	}
	if result != nil {
		t.Fatalf("first insertion result = %p, want nil previous head", result)
	}
	if questJournalHead500540 != first || first.next != nil || first.prev != nil {
		t.Fatal("first insertion did not establish a singleton list")
	}

	second, result := questJournalSetResult500540("War01a:Second", 1, 5)
	if second == nil {
		t.Fatal("second insertion did not allocate an entry")
	}
	if result != first {
		t.Fatalf("second insertion result = %p, want previous head %p", result, first)
	}
	if questJournalHead500540 != second || second.next != first || second.prev != nil || first.prev != second {
		t.Fatal("second insertion did not prepend the original doubly linked list")
	}

	updated, result := questJournalSetResult500540("war01A:first", 1, 7)
	if updated != first || result != first {
		t.Fatalf("update result = entry %p/raw %p, want existing entry %p", updated, result, first)
	}
	if uint32(first.kind) != 0 || uint32(first.value) != 7 {
		t.Fatalf("updated entry = kind %d/value %d, want 0/7", uint32(first.kind), uint32(first.value))
	}
	if questJournalHead500540 != second {
		t.Fatal("updating an existing entry changed the list head")
	}
}

func TestQuestJournalFind5005E0UsesBytewiseCInsensitiveComparison(t *testing.T) {
	resetQuestJournal500540(t)
	entry := questJournalSet500540("War01a:K", 0, 1)
	if entry == nil {
		t.Fatal("cannot allocate quest-journal entry")
	}
	if got := questJournalFind5005E0("war01A:k"); got != entry {
		t.Fatalf("ASCII case-insensitive lookup = %p, want %p", got, entry)
	}
	if got := questJournalFind5005E0("War01a:\u212a"); got != nil {
		t.Fatalf("Unicode-fold lookup = %p, want nil bytewise C comparison", got)
	}
}

func TestQuestJournalGetInt500750ReturnsExactDwordBits(t *testing.T) {
	resetQuestJournal500540(t)
	if got := QuestJournalGetInt500750("War01a:Missing"); got != 0 {
		t.Fatalf("missing value = %d, want 0", got)
	}

	entry := questJournalSet500540("War01a:Value", 1, 0)
	if entry == nil {
		t.Fatal("cannot allocate quest-journal entry")
	}
	for _, value := range []uint32{0, 1, 0x7fffffff, 0x80000000, 0xffffffff} {
		if got := questJournalSet500540("WAR01A:VALUE", 0, value); got != entry {
			t.Fatalf("update for value %#08x returned %p, want %p", value, got, entry)
		}
		if got, want := QuestJournalGetInt500750("war01A:value"), int32(value); got != want {
			t.Fatalf("value bits %#08x = %d, want %d", value, got, want)
		}
	}
	if uint32(entry.kind) != 1 {
		t.Fatalf("getter changed or rejected Boolean entry kind %d", uint32(entry.kind))
	}
}

func TestQuestJournalGetFloat500770MatchesX87FloatRoundTrip(t *testing.T) {
	resetQuestJournal500540(t)
	if got := math.Float32bits(QuestJournalGetFloat500770("War01a:Missing")); got != 0 {
		t.Fatalf("missing value bits = %#08x, want 0", got)
	}

	entry := questJournalSet500540("War01a:Value", 1, 0)
	if entry == nil {
		t.Fatal("cannot allocate quest-journal entry")
	}
	tests := []struct {
		stored uint32
		want   uint32
	}{
		{0x00000000, 0x00000000},
		{0x80000000, 0x80000000},
		{0x00000001, 0x00000001},
		{0x007fffff, 0x007fffff},
		{0x3f800000, 0x3f800000},
		{0x7f7fffff, 0x7f7fffff},
		{0x7f800000, 0x7f800000},
		{0xff800000, 0xff800000},
		{0x7fc12345, 0x7fc12345},
		{0xffc12345, 0xffc12345},
		{0x7f812345, 0x7fc12345},
		{0xff812345, 0xffc12345},
	}
	for _, tc := range tests {
		if got := questJournalSet500540("WAR01A:VALUE", 0, tc.stored); got != entry {
			t.Fatalf("update for bits %#08x returned %p, want %p", tc.stored, got, entry)
		}
		got := math.Float32bits(QuestJournalGetFloat500770("war01A:value"))
		if got != tc.want {
			t.Fatalf("stored bits %#08x = %#08x, want %#08x", tc.stored, got, tc.want)
		}
	}
	if uint32(entry.kind) != 1 {
		t.Fatalf("getter changed or rejected Boolean entry kind %d", uint32(entry.kind))
	}
}

func TestQuestJournalSet500540RejectsOnlyUnsafeNameBoundary(t *testing.T) {
	resetQuestJournal500540(t)
	valid := strings.Repeat("v", 130) + ":"
	if len(valid) != 131 {
		t.Fatalf("valid fixture length = %d, want 131", len(valid))
	}
	entry, _ := questJournalSetResult500540(valid, 0, 1)
	if entry == nil {
		t.Fatal("131-byte journal name was rejected")
	}

	invalid := strings.Repeat("x", 131) + ":"
	if len(invalid) != 132 {
		t.Fatalf("invalid fixture length = %d, want 132", len(invalid))
	}
	head := questJournalHead500540
	entry, result := questJournalSetResult500540(invalid, 0, 2)
	if entry != nil || result != nil {
		t.Fatalf("132-byte journal name result = entry %p/raw %p, want nil/nil", entry, result)
	}
	if questJournalHead500540 != head {
		t.Fatal("rejected journal name changed the list head")
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
