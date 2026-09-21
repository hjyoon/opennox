package legacy

import (
	"math"
	"reflect"
	"testing"
)

type spellResultClient4FB0B0 struct {
	Client
	statuses []uint32
}

func (c *spellResultClient4FB0B0) SpellResult4FB0B0(status uint32) bool {
	c.statuses = append(c.statuses, status)
	return status < 18
}

func TestSpellResultExport4FB0B0PreservesCompleteUint32Status(t *testing.T) {
	fake := new(spellResultClient4FB0B0)
	oldGetClient := GetClient
	GetClient = func() Client { return fake }
	t.Cleanup(func() { GetClient = oldGetClient })

	want := []uint32{0, 3, 17, 0x80000000, math.MaxUint32}
	for _, status := range want {
		spellResultExportCall4FB0B0(status)
	}
	if !reflect.DeepEqual(fake.statuses, want) {
		t.Fatalf("statuses = %#v, want %#v", fake.statuses, want)
	}
}

func TestSpellResultPacketInform4C9BF0DecodesCompleteLittleEndianStatus(t *testing.T) {
	fake := new(spellResultClient4FB0B0)
	oldGetClient := GetClient
	GetClient = func() Client { return fake }
	t.Cleanup(func() { GetClient = oldGetClient })

	want := []uint32{3, 0x01020304, 0x80000000, math.MaxUint32}
	for _, status := range want {
		if got := spellResultPacketInformCall4C9BF0(status); got != 6 {
			t.Fatalf("MSG_INFORM status %#x consumed %d bytes, want 6", status, got)
		}
	}
	if !reflect.DeepEqual(fake.statuses, want) {
		t.Fatalf("MSG_INFORM statuses = %#v, want %#v", fake.statuses, want)
	}
}
