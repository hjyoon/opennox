package legacy

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
)

func writeMapgenDecorFixture51F9F0(t *testing.T, name, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestMapgenDecor51F9F0CopiesAndSelectsNativePointerGraph(t *testing.T) {
	handles.Init()
	t.Cleanup(handles.Release)
	templatePath := writeMapgenDecorFixture51F9F0(t, "template.txt", `TEMPLATE
TemplateBase
WALL_FLOOR
StoneWall
StoneFloor
DECOR_SET
0
0
END
END
`)
	backdropPath := writeMapgenDecorFixture51F9F0(t, "backdrop.txt", `BACKDROP
BackdropCopy
COPY
TemplateBase
OCCUR_CONSTRAINT
NONE
FREQUENCY
RARE
ROOM_SIZE_CONSTRAINT
2
6
DOOR
DoorType
DOUBLE_DOOR
DoubleDoorType
END
`)
	got := mapgenDecorFixture51F9F0(templatePath, backdropPath)
	if !got.templateParsed || !got.backdropParsed || !got.settingsValid ||
		!got.selected || !got.cleanupCleared {
		t.Fatalf("parse template=%v backdrop=%v settings=%v selected=%v cleanup=%v",
			got.templateParsed, got.backdropParsed, got.settingsValid,
			got.selected, got.cleanupCleared)
	}
	if got.templateName != "TemplateBase" || got.backdropName != "BackdropCopy" ||
		got.wallName != "StoneWall" || got.floorName != "StoneFloor" ||
		got.doorName != "DoorType" || got.doubleDoorName != "DoubleDoorType" {
		t.Fatalf("names template=%q backdrop=%q wall=%q floor=%q door=%q double=%q",
			got.templateName, got.backdropName, got.wallName, got.floorName,
			got.doorName, got.doubleDoorName)
	}
	if got.frequency != 100 || got.minRoomSize != 2 || got.maxRoomSize != 6 {
		t.Fatalf("frequency=%d room size=%d..%d, want 100 and 2..6",
			got.frequency, got.minRoomSize, got.maxRoomSize)
	}
	if got.templateWallCount != 1 || got.backdropWallCount != 1 ||
		got.templateSetCount != 1 || got.backdropSetCount != 1 {
		t.Fatalf("wall counts=%d/%d set counts=%d/%d, want all one",
			got.templateWallCount, got.backdropWallCount,
			got.templateSetCount, got.backdropSetCount)
	}
	if got.templateSetSharesEntries || !got.backdropSetSharesEntries {
		t.Fatalf("set ownership template=%v backdrop=%v, want false/true",
			got.templateSetSharesEntries, got.backdropSetSharesEntries)
	}
	if got.templateWallAddress == got.backdropWallAddress ||
		got.templateSetAddress == got.backdropSetAddress {
		t.Fatalf("COPY reused source allocations: wall=%#x/%#x set=%#x/%#x",
			got.templateWallAddress, got.backdropWallAddress,
			got.templateSetAddress, got.backdropSetAddress)
	}
	if got.selectedToken != got.backdropToken || got.selectedAddress != got.backdropAddress {
		t.Fatalf("selection token=%#x address=%#x, want backdrop %#x/%#x",
			got.selectedToken, got.selectedAddress, got.backdropToken, got.backdropAddress)
	}

	tokens := map[string]uint32{
		"theme":         got.themeToken,
		"template":      got.templateToken,
		"backdrop":      got.backdropToken,
		"template wall": got.templateWallToken,
		"backdrop wall": got.backdropWallToken,
		"template set":  got.templateSetToken,
		"backdrop set":  got.backdropSetToken,
	}
	addresses := map[string]uintptr{
		"theme":         got.themeAddress,
		"template":      got.templateAddress,
		"backdrop":      got.backdropAddress,
		"template wall": got.templateWallAddress,
		"backdrop wall": got.backdropWallAddress,
		"template set":  got.templateSetAddress,
		"backdrop set":  got.backdropSetAddress,
	}
	for name, token := range tokens {
		if token == 0 || addresses[name] == 0 {
			t.Fatalf("%s token=%#x address=%#x, want both nonzero",
				name, token, addresses[name])
		}
		if strconv.IntSize == 64 && addresses[name] <= 1<<32 {
			t.Fatalf("%s fixture address=%#x, want address above PE32 range",
				name, addresses[name])
		}
		if strconv.IntSize == 64 && uintptr(token) == addresses[name] {
			t.Fatalf("%s token %#x leaked native pointer", name, token)
		}
	}
}
