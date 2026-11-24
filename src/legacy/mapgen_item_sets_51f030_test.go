package legacy

import (
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
)

func initMapgenItemSetHandles51F030(t *testing.T) {
	t.Helper()
	handles.InitHigh()
	t.Cleanup(handles.Release)
}

func writeMapgenItemSetFixture51F030(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "item-set.txt")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func mapgenItemSetModifierNames51F030(got mapgenItemSetsResult51F030, entry, slot int) []string {
	count := int(got.modifierCounts[entry][slot])
	names := make([]string, count)
	copy(names, got.modifierNames[entry][slot][:count])
	return names
}

func checkMapgenItemSetNativeAddresses51F030(t *testing.T, got mapgenItemSetsResult51F030) {
	t.Helper()
	if strconv.IntSize != 64 {
		return
	}
	addresses := map[string]uintptr{"theme": got.themeAddress}
	for entry, address := range got.entryAddresses {
		if address == 0 {
			continue
		}
		addresses["entry "+strconv.Itoa(entry)] = address
		for slot, modifierAddress := range got.modifierAddresses[entry] {
			if modifierAddress != 0 {
				addresses["entry "+strconv.Itoa(entry)+" modifier "+strconv.Itoa(slot)] = modifierAddress
			}
		}
	}
	for name, address := range addresses {
		if uint64(address) <= 1<<32 {
			t.Fatalf("%s fixture address=%#x, want address above PE32 range", name, address)
		}
	}
}

func TestMapgenWeaponSets51F030PreserveNativePointersAndTemplateSemantics(t *testing.T) {
	initMapgenItemSetHandles51F030(t)
	path := writeMapgenItemSetFixture51F030(t, `WEAPON
TEMPLATE
EFFECTIVENESS
BaseEffect
DropMe
END
MATERIAL
BaseMaterial
END
PRIMARY_ENCHANTMENT
BasePrimary
END
SECONDARY_ENCHANTMENT
BaseSecondary
END
END
WEAPON
Longsword
SwordKey
EFFECTIVENESS
-DropMe
NewEffect
BaseEffect
END
MATERIAL
NewMaterial
END
PRIMARY_ENCHANTMENT
END
SECONDARY_ENCHANTMENT
END
END
WEAPON
Bow
BowKey
END
END
`)
	got := mapgenItemSetsFixture51F030(path, false)
	if !got.result || got.count != 2 {
		t.Fatalf("parse result=%v count=%d, want true and two definitions", got.result, got.count)
	}
	if !got.templateCleared || !got.cleanupCleared {
		t.Fatalf("template cleared=%v cleanup cleared=%v, want both true",
			got.templateCleared, got.cleanupCleared)
	}
	if got.lookupNames != [2]string{"SwordKey", "BowKey"} ||
		got.objectNames != [2]string{"Longsword", "Bow"} {
		t.Fatalf("names lookup=%q object=%q", got.lookupNames, got.objectNames)
	}
	if got.entryTokens[0] == 0 || got.entryTokens[1] == 0 ||
		got.nextAddresses[0] != got.entryAddresses[1] || got.nextAddresses[1] != 0 {
		t.Fatalf("entry tokens=%#x next=%#x addresses=%#x",
			got.entryTokens, got.nextAddresses, got.entryAddresses)
	}

	want := [2][4][]string{
		{
			{"BaseEffect", "NewEffect"},
			{"BaseMaterial", "NewMaterial"},
			{"BasePrimary"},
			{"BaseSecondary"},
		},
		{
			{"BaseEffect", "DropMe"},
			{"BaseMaterial"},
			{"BasePrimary"},
			{"BaseSecondary"},
		},
	}
	for entry := range want {
		for slot := range want[entry] {
			actual := mapgenItemSetModifierNames51F030(got, entry, slot)
			if !reflect.DeepEqual(actual, want[entry][slot]) {
				t.Fatalf("entry %d slot %d modifiers=%q, want %q",
					entry, slot, actual, want[entry][slot])
			}
			if got.modifierTokens[entry][slot] == 0 || got.modifierAddresses[entry][slot] == 0 {
				t.Fatalf("entry %d slot %d token=%#x address=%#x, want both nonzero",
					entry, slot, got.modifierTokens[entry][slot], got.modifierAddresses[entry][slot])
			}
		}
	}
	checkMapgenItemSetNativeAddresses51F030(t, got)
}

func TestMapgenArmorSets51F640PreserveNativePointersWithoutTemplate(t *testing.T) {
	initMapgenItemSetHandles51F030(t)
	path := writeMapgenItemSetFixture51F030(t, `ARMOR
PlateArmor
PlateKey
QUALITY
QualityOne
END
END
END
`)
	got := mapgenItemSetsFixture51F030(path, true)
	if !got.result || got.count != 1 {
		t.Fatalf("parse result=%v count=%d, want true and one definition", got.result, got.count)
	}
	if got.lookupNames[0] != "PlateKey" || got.objectNames[0] != "PlateArmor" {
		t.Fatalf("names lookup=%q object=%q", got.lookupNames[0], got.objectNames[0])
	}
	if actual := mapgenItemSetModifierNames51F030(got, 0, 0); !reflect.DeepEqual(actual, []string{"QualityOne"}) {
		t.Fatalf("quality modifiers=%q", actual)
	}
	if got.modifierCounts[0] != [4]uint32{1, 0, 0, 0} {
		t.Fatalf("modifier counts=%v, want [1 0 0 0]", got.modifierCounts[0])
	}
	if !got.templateCleared || !got.cleanupCleared {
		t.Fatalf("template cleared=%v cleanup cleared=%v, want both true",
			got.templateCleared, got.cleanupCleared)
	}
	checkMapgenItemSetNativeAddresses51F030(t, got)
}

func TestMapgenItemSetFailureKeepsEarlierEntries(t *testing.T) {
	initMapgenItemSetHandles51F030(t)
	const prefix = "WEAPON\nTEMPLATE\nMATERIAL\nBase\nEND\nEND\nWEAPON\nBow\nBowKey\nEND\nWEAPON\nStaff\nStaffKey\n"
	for _, tc := range []struct{ name, suffix string }{
		{"modifier count", "MATERIAL\n" + strings.Repeat("Material\n", mapgenMaxItemModifiers+1) + "END\nEND\nEND\n"},
		{"inherited count", "EFFECTIVENESS\nApplied\nEND\nMATERIAL\n" + itemModifierNamesAtCapacity() + "END\nEND\nEND\n"},
		{"truncated block", "MATERIAL\nPartial\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := mapgenItemSetsFixture51F030(writeMapgenItemSetFixture51F030(t, prefix+tc.suffix), false)
			if got.result || got.count != 1 || !got.templateCleared || !got.cleanupCleared || got.lookupNames[0] != "BowKey" {
				t.Fatalf("failure changed the earlier entry or leaked the template: %+v", got)
			}
			if names := mapgenItemSetModifierNames51F030(got, 0, 1); !reflect.DeepEqual(names, []string{"Base"}) {
				t.Fatalf("earlier entry's modifiers = %q", names)
			}
		})
	}
}

func itemModifierNamesAtCapacity() string {
	var names strings.Builder
	for i := 0; i < mapgenMaxItemModifiers; i++ {
		names.WriteString("Material" + strconv.Itoa(i) + "\n")
	}
	return names.String()
}
