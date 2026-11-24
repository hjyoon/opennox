package legacy

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
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

func setMapgenDecorKeywords(t *testing.T, offset uintptr, names ...string) {
	t.Helper()
	for i, name := range append(names, "") {
		slot := memmap.PtrPtr(0x587000, offset+4*uintptr(i))
		old := *slot
		t.Cleanup(func() { *slot = old })
		if name == "" {
			*slot = nil
		} else {
			*slot = unsafe.Pointer(alloc.InternCString(name))
		}
	}
}

func TestMapgenDecorConstraintNative520810(t *testing.T) {
	handles.InitHigh()
	t.Cleanup(handles.Release)
	scratch := memmap.Slice(0x5D4594, 2487264)[:256]
	oldScratch := bytes.Clone(scratch)
	t.Cleanup(func() { copy(scratch, oldScratch) })
	setMapgenDecorKeywords(t, mapgenDecorConstraintNames, "One", "Two", "Three")
	for _, tc := range []struct {
		name, text, wantScratch string
		mask                    byte
		ok                      bool
	}{
		{"separator spans", "+one++Two+THREE+\n", "+one\x00+Two\x00THREE\x00\x00", 7, true},
		{"unknown after match", "One+Bad+Three\n", "One\x00Bad\x00Three\x00", 1, false},
		{"none", "nOnE\n", "nOnE\x00", 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := writeMapgenDecorFixture51F9F0(t, "constraint.txt", tc.text)
			mask, ok := mapgenDecorConstraintFixture(path)
			gotScratch := string(scratch[:len(tc.wantScratch)])
			if mask != tc.mask || ok != tc.ok || gotScratch != tc.wantScratch {
				t.Fatalf("native constraint = %#x/%t, scratch = %q", mask, ok, gotScratch)
			}
		})
	}
}

func TestMapgenDecorChoicesNative520380(t *testing.T) {
	handles.InitHigh()
	t.Cleanup(handles.Release)
	setMapgenDecorKeywords(t, mapgenDecorEntryNames, "OBJECT", "AREAMAP", "CLEAR_COLLIDES", "WEAPON", "ARMOR", "SPELL")
	var text strings.Builder
	text.WriteString("20\n")
	for i := 0; i < 34; i++ {
		fmt.Fprintf(&text, "object\nItem%d\n", i)
	}
	text.WriteString("OR\n*\nSPELL\nSpark\nOR\n*\nARMOR\nCoat\nOR\n-3\nWEAPON\nStaff\nEND\n")
	got := mapgenDecorChoicesFixture(writeMapgenDecorFixture51F9F0(t, "choices.txt", text.String()))
	if len(got) != 4 {
		t.Fatalf("choice groups = %d, want 4", len(got))
	}
	if got[0].weight != 20 || got[0].count != 31 || alloc.GoStringS(got[0].items[30].name[:]) != "Item30" ||
		alloc.GoStringS(got[0].items[31].name[:]) != "Item33" {
		t.Fatalf("the kept choices or saturation slot changed, weight=%d count=%d", got[0].weight, got[0].count)
	}
	for i, name := range []string{"Spark", "Coat", "Staff"} {
		group := &got[i+1]
		if group.weight != 42 || group.count != 1 || group.items[0].kind != uint32(5-i) || alloc.GoStringS(group.items[0].name[:]) != name {
			t.Fatalf("choice group %d disagrees with its original wildcard weight, type or name", i+1)
		}
	}
	for _, invalid := range []string{"20\nOR\n*\nUNKNOWN\n", "20\nOR\n*\nOBJECT\n" + strings.Repeat("x", 60) + "\nEND\n"} {
		if got := mapgenDecorChoicesFixture(writeMapgenDecorFixture51F9F0(t, "invalid.txt", invalid)); got != nil {
			t.Fatal("invalid choice published a partially read list")
		}
	}
}

func TestMapgenDecorOccurrenceLimitNative524220(t *testing.T) {
	order, totals := mapgenConsumeDecorLimitFixture()
	if order != [2]bool{true, true} || totals != [6]int32{100, 87, 87, 87, 87, 100} {
		t.Fatalf("occurrence limit order = %v, constraint totals = %v", order, totals)
	}
}

func TestMapgenDecor51F9F0CopiesAndSelectsNativePointerGraph(t *testing.T) {
	handles.InitHigh()
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
	rejectedPath := writeMapgenDecorFixture51F9F0(t, "rejected.txt", `BACKDROP
Rejected
WALL_FLOOR
TemporaryWall
TemporaryFloor
DECOR_SET
0
0
END
COPY
TemplateBase
INVALID_DECOR_RULE
`)
	got := mapgenDecorFixture51F9F0(templatePath, backdropPath, rejectedPath)
	if !got.templateParsed || !got.backdropParsed || !got.settingsValid ||
		!got.selected || !got.cleanupCleared {
		t.Fatalf("parse template=%v backdrop=%v settings=%v selected=%v cleanup=%v",
			got.templateParsed, got.backdropParsed, got.settingsValid,
			got.selected, got.cleanupCleared)
	}
	if !got.rejectedKeptBackdrop {
		t.Fatal("rejected decoration changed the published backdrop list")
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
		if strconv.IntSize == 64 && uint64(addresses[name]) <= 1<<32 {
			t.Fatalf("%s fixture address=%#x, want address above PE32 range",
				name, addresses[name])
		}
		if strconv.IntSize == 64 && uintptr(token) == addresses[name] {
			t.Fatalf("%s token %#x leaked native pointer", name, token)
		}
	}
}
