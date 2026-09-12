package noxscriptqueue

import (
	"strconv"
	"testing"
)

type triple struct {
	block, caller, trigger uintptr
}

func queueToken(n uint64) uintptr {
	return uintptr((uint64(0x12345678) << 32) | n)
}

func TestAppendPreservesNativeTriplesAndCapacity(t *testing.T) {
	var queue []triple
	for i := uint64(0); i <= Capacity; i++ {
		queue = Append(queue, triple{queueToken(i + 1), queueToken(i + 101), queueToken(i + 201)})
	}
	if got := len(queue); got != Capacity {
		t.Fatalf("queue length = %d, want %d", got, Capacity)
	}
	for i, got := range queue {
		want := triple{queueToken(uint64(i) + 1), queueToken(uint64(i) + 101), queueToken(uint64(i) + 201)}
		if got != want {
			t.Fatalf("entry %d = %+v, want %+v", i, got, want)
		}
		if strconv.IntSize == 64 && uint64(got.block) <= uint64(^uint32(0)) {
			t.Fatalf("native block token = %#x, want above PE32 range", got.block)
		}
	}
}

func TestRemoveRequiresExactTripleAndSkipsShiftedMatch(t *testing.T) {
	match := triple{queueToken(1), queueToken(2), queueToken(3)}
	nearBlock := triple{queueToken(4), match.caller, match.trigger}
	nearCaller := triple{match.block, queueToken(5), match.trigger}
	nearTrigger := triple{match.block, match.caller, queueToken(6)}
	queue := make([]triple, 6, 6)
	copy(queue, []triple{match, match, nearBlock, nearCaller, nearTrigger, match})

	queue = Remove(queue, match)
	want := []triple{match, nearBlock, nearCaller, nearTrigger}
	if len(queue) != len(want) {
		t.Fatalf("queue length = %d, want %d", len(queue), len(want))
	}
	for i, got := range queue {
		if got != want[i] {
			t.Fatalf("entry %d = %+v, want %+v", i, got, want[i])
		}
	}
	if got := queue[:cap(queue)][5]; got != match {
		t.Fatalf("unused tail = %+v, want original last entry %+v", got, match)
	}
}

func TestRemoveLeavesNonmatchingAndEmptyQueuesAlone(t *testing.T) {
	entry := triple{queueToken(1), queueToken(2), queueToken(3)}
	queue := []triple{entry}
	if got := Remove(queue, triple{}); len(got) != 1 || got[0] != entry {
		t.Fatalf("nonmatching removal changed queue: %+v", got)
	}
	if got := Remove([]triple(nil), entry); got != nil {
		t.Fatalf("empty removal = %+v, want nil", got)
	}
}
