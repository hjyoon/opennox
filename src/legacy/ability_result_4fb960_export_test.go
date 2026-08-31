package legacy

import (
	"math"
	"reflect"
	"testing"
)

type abilityResultClient4FB960 struct {
	Client
	statuses []uint32
}

func (c *abilityResultClient4FB960) AbilityResult4FB960(status uint32) bool {
	c.statuses = append(c.statuses, status)
	return status < 8
}

func TestAbilityResultExport4FB960PreservesCompleteUint32Status(t *testing.T) {
	fake := new(abilityResultClient4FB960)
	oldGetClient := GetClient
	GetClient = func() Client { return fake }
	t.Cleanup(func() { GetClient = oldGetClient })

	want := []uint32{0, 7, 0x80000000, math.MaxUint32}
	for _, status := range want {
		abilityResultExportCall4FB960(status)
	}
	if !reflect.DeepEqual(fake.statuses, want) {
		t.Fatalf("statuses = %#v, want %#v", fake.statuses, want)
	}
}
