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
