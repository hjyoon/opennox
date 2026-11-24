package legacy

import (
	"reflect"
	"testing"
)

func TestClientSequenceQueuePreservesWrapDeduplicationAndTimeout(t *testing.T) {
	ticks := uint64(500)
	oldTicks := PlatformTicks
	PlatformTicks = func() uint64 { return ticks }
	t.Cleanup(func() { PlatformTicks = oldTicks })

	var queue clientSequenceQueue
	expected := uint16(0xfffe)
	queued := []byte{3}
	queue.add(expected, 0, queued)
	queued[0] = 99
	queue.add(expected, 0xffff, []byte{2})
	queue.add(expected, 0xfffe, []byte{1})
	queue.add(expected, 0xffff, []byte{99})
	queue.add(expected, 0xfffd, []byte{99})

	var got []byte
	deliver := func(data []byte) { got = append(got, data...) }
	if result := queue.drain(&expected, deliver); result != 0 || expected != 1 || !reflect.DeepEqual(got, []byte{1, 2, 3}) {
		t.Fatalf("wrapped queue: result=%d expected=%d payload=%v", result, expected, got)
	}

	queue.add(expected, 3, []byte{6})
	queue.add(expected, 1, []byte{4})
	if result := queue.drain(&expected, deliver); result != 3 || expected != 2 {
		t.Fatalf("missing sequence: result=%d expected=%d", result, expected)
	}
	ticks += 30_000
	if result := queue.drain(&expected, deliver); result != 30_000 || expected != 2 {
		t.Fatalf("timeout boundary: result=%d expected=%d", result, expected)
	}
	ticks++
	if result := queue.drain(&expected, deliver); result != 0 || expected != 2 {
		t.Fatalf("idle drain kept the previous stopping point: result=%d expected=%d", result, expected)
	}
	queue.add(expected, 2, []byte{5})
	if result := queue.drain(&expected, deliver); result != 3 || expected != 4 || !reflect.DeepEqual(got, []byte{1, 2, 3, 4, 5, 6}) {
		t.Fatalf("filled gap: result=%d expected=%d payload=%v", result, expected, got)
	}

	queue.add(expected, 6, []byte{8})
	queue.add(expected, 4, []byte{7})
	queue.drain(&expected, deliver)
	ticks += 30_001
	if result := queue.drain(&expected, deliver); result != 6 || expected != 7 || !reflect.DeepEqual(got, []byte{1, 2, 3, 4, 5, 6, 7, 8}) {
		t.Fatalf("timeout recovery: result=%d expected=%d payload=%v", result, expected, got)
	}
}
