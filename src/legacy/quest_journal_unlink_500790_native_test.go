package legacy

import (
	"math"
	"testing"
	"unsafe"
)

func requireQuestJournalNativeEntry500790[Entry any](t *testing.T, entry *Entry, name string) *Entry {
	t.Helper()
	if entry == nil {
		t.Fatalf("cannot allocate quest-journal entry %q", name)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(entry)) <= math.MaxUint32 {
		t.Fatalf("entry pointer = %p, want native address above 4 GiB", entry)
	}
	return entry
}

func TestQuestJournalDeleteEntry500790NativeInteriorTailAndSingleton(t *testing.T) {
	t.Run("interior", func(t *testing.T) {
		resetQuestJournal500540(t)
		tail := requireQuestJournalNativeEntry500790(t, questJournalSet500540("War01a:Tail", 0, 1), "War01a:Tail")
		entry := requireQuestJournalNativeEntry500790(t, questJournalSet500540("War01a:Entry", 0, 1), "War01a:Entry")
		head := requireQuestJournalNativeEntry500790(t, questJournalSet500540("War01a:Head", 0, 1), "War01a:Head")

		questJournalDeleteEntry500790(entry)

		if questJournalHead500540 != head || head.next != tail || tail.prev != head {
			t.Fatalf("interior unlink = head %p, links %p/%p; want %p, %p/%p", questJournalHead500540, head.next, tail.prev, head, tail, head)
		}
	})

	t.Run("tail", func(t *testing.T) {
		resetQuestJournal500540(t)
		tail := requireQuestJournalNativeEntry500790(t, questJournalSet500540("War01a:Tail", 0, 1), "War01a:Tail")
		head := requireQuestJournalNativeEntry500790(t, questJournalSet500540("War01a:Head", 0, 1), "War01a:Head")

		questJournalDeleteEntry500790(tail)

		if questJournalHead500540 != head || head.next != nil || head.prev != nil {
			t.Fatalf("tail unlink = head %p, links %p/%p; want singleton %p", questJournalHead500540, head.next, head.prev, head)
		}
	})

	t.Run("singleton", func(t *testing.T) {
		resetQuestJournal500540(t)
		entry := requireQuestJournalNativeEntry500790(t, questJournalSet500540("War01a:Entry", 0, 1), "War01a:Entry")

		questJournalDeleteEntry500790(entry)

		if questJournalHead500540 != nil {
			t.Fatalf("singleton unlink head = %p, want nil", questJournalHead500540)
		}
	})
}

func TestQuestJournalDeleteEntry500790NativeHeadIdentityIsIndependentOfPrev(t *testing.T) {
	t.Run("nil-prev-but-not-head", func(t *testing.T) {
		resetQuestJournal500540(t)
		tail := requireQuestJournalNativeEntry500790(t, questJournalSet500540("War01a:Tail", 0, 1), "War01a:Tail")
		entry := requireQuestJournalNativeEntry500790(t, questJournalSet500540("War01a:Entry", 0, 1), "War01a:Entry")
		head := requireQuestJournalNativeEntry500790(t, questJournalSet500540("War01a:Head", 0, 1), "War01a:Head")
		head.next = tail
		tail.prev = head
		entry.prev = nil

		questJournalDeleteEntry500790(entry)

		if questJournalHead500540 != head {
			t.Fatalf("detached unlink head = %p, want unchanged %p", questJournalHead500540, head)
		}
		if tail.prev != nil {
			t.Fatalf("successor previous = %p, want live nil previous", tail.prev)
		}
		tail.prev = head
	})

	t.Run("non-nil-prev-but-is-head", func(t *testing.T) {
		resetQuestJournal500540(t)
		tail := requireQuestJournalNativeEntry500790(t, questJournalSet500540("War01a:Tail", 0, 1), "War01a:Tail")
		entry := requireQuestJournalNativeEntry500790(t, questJournalSet500540("War01a:Entry", 0, 1), "War01a:Entry")
		prev := requireQuestJournalNativeEntry500790(t, questJournalSet500540("War01a:Prev", 0, 1), "War01a:Prev")
		questJournalHead500540 = entry

		questJournalDeleteEntry500790(entry)

		if questJournalHead500540 != tail {
			t.Fatalf("head-identity unlink head = %p, want successor %p", questJournalHead500540, tail)
		}
		if prev.next != tail || tail.prev != prev {
			t.Fatalf("head-identity links = %p/%p, want %p/%p", prev.next, tail.prev, tail, prev)
		}
		questJournalHead500540 = prev
	})
}
