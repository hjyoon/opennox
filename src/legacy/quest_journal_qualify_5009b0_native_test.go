package legacy

import (
	"bytes"
	"strings"
	"testing"

	"github.com/opennox/opennox/v1/common/memmap"
)

func prepareQuestJournalQualifyNative5009B0(t *testing.T, mapName string, scratchFill byte) ([]byte, []byte) {
	t.Helper()
	if len(mapName) >= questJournalScratchCapacity5009B0 {
		t.Fatalf("test map name length = %d, must leave room for NUL", len(mapName))
	}
	mapSlot := memmap.Slice(questJournalMemoryBase5009B0, questJournalMapNameOffset5009B0)[:questJournalScratchCapacity5009B0]
	scratch := questJournalScratch5009B0()
	savedMap := append([]byte(nil), mapSlot...)
	savedScratch := append([]byte(nil), scratch...)
	t.Cleanup(func() {
		copy(mapSlot, savedMap)
		copy(scratch, savedScratch)
	})
	for i := range mapSlot {
		mapSlot[i] = 0
	}
	copy(mapSlot, mapName)
	for i := range scratch {
		scratch[i] = scratchFill
	}
	return mapSlot, scratch
}

func requireQuestJournalScratch5009B0(t *testing.T, scratch []byte, value string, trailing byte) {
	t.Helper()
	written := len(value) + 1
	if got := string(scratch[:len(value)]); got != value {
		t.Fatalf("scratch value = %q, want %q", got, value)
	}
	if scratch[len(value)] != 0 {
		t.Fatalf("scratch terminator = %#x, want NUL", scratch[len(value)])
	}
	if tail := scratch[written:]; !bytes.Equal(tail, bytes.Repeat([]byte{trailing}, len(tail))) {
		t.Fatalf("scratch tail = %x, want unchanged %#x bytes", tail, trailing)
	}
}

func TestQuestJournalQualifyNative5009B0WritesQualifiedInputAndReturn(t *testing.T) {
	_, scratch := prepareQuestJournalQualifyNative5009B0(t, "must-not-be-read", 0xa5)
	name := "War01a:Count:Extra"
	if got, want := questJournalQualifyNative5009B0(name), uint32(len(name)+1); got != want {
		t.Fatalf("return = %d, want %d", got, want)
	}
	requireQuestJournalScratch5009B0(t, scratch, name, 0xa5)
}

func TestQuestJournalQualifyNative5009B0PrefixesCurrentMapAndReturnsZero(t *testing.T) {
	_, scratch := prepareQuestJournalQualifyNative5009B0(t, "War01a", 0x5a)
	if got := questJournalQualifyNative5009B0("Count"); got != 0 {
		t.Fatalf("return = %d, want zero", got)
	}
	requireQuestJournalScratch5009B0(t, scratch, "War01a:Count", 0x5a)
}

func TestQuestJournalQualifyNative5009B0ColonAndEmptyEdges(t *testing.T) {
	for _, tc := range []struct {
		name    string
		input   string
		mapName string
		want    string
		result  uint32
	}{
		{name: "leading-colon", input: ":Quest", mapName: "War01a", want: ":Quest", result: 7},
		{name: "trailing-colon", input: "Quest:", mapName: "War01a", want: "Quest:", result: 7},
		{name: "empty-input", input: "", mapName: "War01a", want: "War01a:"},
		{name: "empty-map", input: "Quest", mapName: "", want: ":Quest"},
		{name: "both-empty", input: "", mapName: "", want: ":"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, scratch := prepareQuestJournalQualifyNative5009B0(t, tc.mapName, 0xc3)
			if got := questJournalQualifyNative5009B0(tc.input); got != tc.result {
				t.Fatalf("return = %d, want %d", got, tc.result)
			}
			requireQuestJournalScratch5009B0(t, scratch, tc.want, 0xc3)
		})
	}
}

func TestQuestJournalQualifyNative5009B0CapacityPolicy(t *testing.T) {
	t.Run("qualified-exact-fit", func(t *testing.T) {
		_, scratch := prepareQuestJournalQualifyNative5009B0(t, "unused", 0x11)
		name := strings.Repeat("Q", 130) + ":"
		if got := questJournalQualifyNative5009B0(name); got != questJournalScratchCapacity5009B0 {
			t.Fatalf("return = %d, want %d", got, questJournalScratchCapacity5009B0)
		}
		requireQuestJournalScratch5009B0(t, scratch, name, 0x11)
	})

	t.Run("qualified-overflow-is-no-op", func(t *testing.T) {
		_, scratch := prepareQuestJournalQualifyNative5009B0(t, "unused", 0x22)
		before := append([]byte(nil), scratch...)
		name := strings.Repeat("Q", 131) + ":"
		if got, want := questJournalQualifyNative5009B0(name), uint32(len(name)+1); got != want {
			t.Fatalf("return = %d, want %d", got, want)
		}
		if !bytes.Equal(scratch, before) {
			t.Fatalf("overflow changed scratch: got %x, want %x", scratch, before)
		}
	})

	t.Run("prefixed-exact-fit", func(t *testing.T) {
		_, scratch := prepareQuestJournalQualifyNative5009B0(t, "M", 0x33)
		name := strings.Repeat("N", 129)
		if got := questJournalQualifyNative5009B0(name); got != 0 {
			t.Fatalf("return = %d, want zero", got)
		}
		requireQuestJournalScratch5009B0(t, scratch, "M:"+name, 0x33)
	})

	t.Run("prefixed-overflow-is-no-op", func(t *testing.T) {
		_, scratch := prepareQuestJournalQualifyNative5009B0(t, "M", 0x44)
		before := append([]byte(nil), scratch...)
		if got := questJournalQualifyNative5009B0(strings.Repeat("N", 130)); got != 0 {
			t.Fatalf("return = %d, want zero", got)
		}
		if !bytes.Equal(scratch, before) {
			t.Fatalf("overflow changed scratch: got %x, want %x", scratch, before)
		}
	})
}
