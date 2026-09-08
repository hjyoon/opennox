package legacy

import (
	"runtime"
	"testing"
)

func TestQuestJournalDeleteEntryExport500790PreservesNativeLinks(t *testing.T) {
	resetQuestJournal500540(t)
	tail := requireQuestJournalNativeEntry500790(t, questJournalSet500540("War01a:Tail", 0, 1), "War01a:Tail")
	entry := requireQuestJournalNativeEntry500790(t, questJournalSet500540("War01a:Entry", 0, 1), "War01a:Entry")
	head := requireQuestJournalNativeEntry500790(t, questJournalSet500540("War01a:Head", 0, 1), "War01a:Head")

	questJournalDeleteEntryExportCall500790(entry)

	if questJournalHead500540 != head || head.next != tail || tail.prev != head {
		t.Fatalf("export unlink = head %p, links %p/%p; want %p, %p/%p", questJournalHead500540, head.next, tail.prev, head, tail, head)
	}
	runtime.KeepAlive(head)
	runtime.KeepAlive(tail)
}
