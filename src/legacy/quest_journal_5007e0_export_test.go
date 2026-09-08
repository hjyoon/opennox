package legacy

import (
	"runtime"
	"testing"
)

func TestQuestJournalDeletePatternExport5007E0PreservesNativeList(t *testing.T) {
	resetQuestJournal500540(t)
	tail := requireQuestJournalNativeEntry500790(t, questJournalSet500540("War02a:Count", 0, 1), "War02a:Count")
	requireQuestJournalNativeEntry500790(t, questJournalSet500540("War01a:Open", 0, 1), "War01a:Open")
	requireQuestJournalNativeEntry500790(t, questJournalSet500540("War01a:Count", 0, 1), "War01a:Count")

	questJournalDeletePatternExportCall5007E0("war01A:*")

	if questJournalHead500540 != tail || tail.next != nil || tail.prev != nil {
		t.Fatalf("export pattern delete = head %p, links %p/%p; want singleton %p", questJournalHead500540, tail.next, tail.prev, tail)
	}
	if questJournalFind5005E0("War01a:Count") != nil || questJournalFind5005E0("War01a:Open") != nil {
		t.Fatal("export pattern delete left a matching entry")
	}
	runtime.KeepAlive(tail)
}
