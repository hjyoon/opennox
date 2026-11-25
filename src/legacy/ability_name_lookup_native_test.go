package legacy

import (
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

func TestAbilityNameLookupNativeStopsAtTableEnd(t *testing.T) {
	var previous [server.AbilityMax]unsafe.Pointer
	for index, name := range server.AbilityNames {
		address := memmap.PtrPtr(0x587000, 69736+4*uintptr(index))
		previous[index] = *address
		*address = unsafe.Pointer(internCStr(name))
	}
	t.Cleanup(func() {
		for index, pointer := range previous {
			*memmap.PtrPtr(0x587000, 69736+4*uintptr(index)) = pointer
		}
	})
	for index, name := range server.AbilityNames {
		if got := abilityIDBySaveName(name); got != server.Ability(index) {
			t.Fatalf("ability %q returned %d, want %d", name, got, index)
		}
	}
	if name := strings.ToLower(server.AbilityNames[server.AbilityWarcry]); abilityIDBySaveName(name) != 0 {
		t.Fatalf("lower-case ability %q matched a saved name", name)
	}
	if got := abilityIDBySaveName("ABILITY_UNKNOWN"); got != 0 {
		t.Fatalf("unknown ability returned %d", got)
	}
}
