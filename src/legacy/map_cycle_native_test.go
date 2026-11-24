package legacy

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/datapath"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestMapCycleNativePreservesFirstLineAndPerModeChecks(t *testing.T) {
	dir := t.TempDir()
	oldRoot := datapath.Data()
	datapath.SetData(dir)
	t.Cleanup(func() { datapath.SetData(oldRoot) })
	p := filepath.Join(dir, "mapcycle.txt")
	wire := "unchecked.map\r\nGood.map\r\nWrong.map\r\nModeB\r\nMissing.map\r\nOther.map\r\nIncomplete.map"
	if err := os.WriteFile(p, []byte(wire), 0o600); err != nil {
		t.Fatal(err)
	}

	for _, region := range [][]byte{
		memmap.Slice(0x5d4594, mapCyclePathOff)[:mapCyclePathSize],
		memmap.Slice(0x5d4594, mapCycleEntriesOff)[:mapCycleModes*mapCycleMapLimit*mapCycleRecordSize],
		memmap.Slice(0x5d4594, mapCycleCountOff)[:4*mapCycleModes],
	} {
		before := bytes.Clone(region)
		t.Cleanup(func() { copy(region, before) })
	}
	wordA := memmap.PtrUint32(0x587000, 191836+8)
	wordB := memmap.PtrUint32(0x587000, 191836+8*2)
	header := memmap.PtrPtr(0x587000, 191832+8*2)
	mode, freeMode := alloc.CStringHigh("ModeB")
	t.Cleanup(freeMode)
	enabled := memmap.PtrUint32(0x5d4594, mapCycleEnabledOff)
	mapFlags := memmap.PtrUint32(0x973f18, 2408+1392)
	oldWordA, oldWordB, oldHeader, oldMode, oldMapFlags := *wordA, *wordB, *header, *enabled, *mapFlags
	t.Cleanup(func() {
		*wordA, *wordB, *header, *enabled, *mapFlags = oldWordA, oldWordB, oldHeader, oldMode, oldMapFlags
	})
	*wordA, *wordB, *header, *enabled = 0x10, 0x40, unsafe.Pointer(mode), 1

	oldCheck, oldFlags := Nox_common_checkMapFile, Nox_mapToGameFlags
	t.Cleanup(func() { Nox_common_checkMapFile, Nox_mapToGameFlags = oldCheck, oldFlags })
	var visits []string
	Nox_common_checkMapFile = func(name string) error {
		visits = append(visits, name)
		switch name {
		case "Good":
			*mapFlags = 0x10
		case "Wrong", "Other":
			*mapFlags = 0x40
		default:
			return os.ErrNotExist
		}
		return nil
	}
	Nox_mapToGameFlags = func(mask int) noxflags.GameFlag { return noxflags.GameFlag(mask) }

	if ret := nox_xxx_loadMapCycle_4D0A30(); ret != nil || *enabled != 1 {
		t.Fatalf("mapcycle returned %p, enable word %d", ret, *enabled)
	}
	if !slices.Equal(visits, []string{"Good", "Wrong", "Missing", "Other"}) {
		t.Fatalf("validated map names = %q", visits)
	}
	for _, tc := range []struct {
		mode int32
		want []string
	}{
		{1, []string{"unchecked.map", "Good.map"}},
		{2, []string{"Other.map"}},
	} {
		if count := *mapCycleModeCount(tc.mode); int(count) != len(tc.want) {
			t.Fatalf("mode %d count = %d, want %d", tc.mode, count, len(tc.want))
		}
		for n, name := range tc.want {
			offset := mapCycleEntriesOff + uintptr(int(tc.mode)*mapCycleMapLimit+n)*mapCycleRecordSize
			if got := alloc.GoStringS(memmap.Slice(0x5d4594, offset)[:mapCycleRecordSize]); got != name {
				t.Fatalf("mode %d map %d = %q, want %q", tc.mode, n, got, name)
			}
		}
	}
	if err := os.Remove(p); err != nil {
		t.Fatal(err)
	}
	nox_xxx_loadMapCycle_4D0A30()
	if *enabled != 0 || *mapCycleModeCount(1) != 0 || *mapCycleModeCount(2) != 0 {
		t.Fatalf("missing cycle kept mode %d and counts %d, %d", *enabled, *mapCycleModeCount(1), *mapCycleModeCount(2))
	}
}

func TestQuestMapPoolNativePreservesSortedNamesGroupsAndHistory(t *testing.T) {
	head := (*nativeListItem)(Get_nox_common_maplist())
	previous := *head
	*head = nativeListItem{next: head, prev: head, head: head}
	t.Cleanup(func() { *head = previous })
	count, clock := questMapFileGlobals4D0F60()
	oldCount, oldClock := *count, *clock
	oldFlags := Nox_mapToGameFlags
	Nox_mapToGameFlags = func(mask int) noxflags.GameFlag { return noxflags.GameFlag(mask) }
	suffix := memmap.Slice(0x587000, 192000)[:5]
	oldSuffix := bytes.Clone(suffix)
	copy(suffix, ".map\x00")
	for i := int32(0); i < 3; i++ {
		entry := questMapEntryNative4D0F60(i)
		before := *entry
		t.Cleanup(func() { *entry = before })
		entry.Uses, entry.LastUsed = uint32(i)+1, uint32(i)+1001
	}
	t.Cleanup(func() { Nox_mapToGameFlags = oldFlags; *count, *clock = oldCount, oldClock; copy(suffix, oldSuffix) })
	for _, tc := range []struct {
		name  string
		flags noxflags.GameFlag
		show  int32
	}{
		{"g_Root02", noxflags.GameModeQuest, 1},
		{"G_Root01", noxflags.GameModeQuest, 1},
		{"Hidden", noxflags.GameModeQuest, 0},
		{"Arena", noxflags.GameModeArena, 1},
		{"Zeta", noxflags.GameModeQuest, 1},
	} {
		item, free := alloc.NewHigh(Nox_map_list_item{})
		t.Cleanup(free)
		copy(item.Name[:], tc.name)
		item.Field_6, item.Field_7 = tc.show, uint32(tc.flags)
		Nox_common_maplist_add_4D0760(item)
	}
	if got := ConsoleMapNames4D09B0(); !slices.Equal(got, []string{"Arena", "G_Root01", "Hidden", "Zeta", "g_Root02"}) {
		t.Fatalf("sorted map list = %q", got)
	}
	if n := nox_xxx_mapSelectFirst_4D0E00(); n != 3 {
		t.Fatalf("quest pool has %d maps", n)
	}
	for i, want := range []string{"G_Root01.map", "Zeta.map", "g_Root02.map"} {
		entry := questMapEntryNative4D0F60(int32(i))
		group := []uint32{1, 2, 1}[i]
		if name := alloc.GoStringS(entry.Name[:]); name != want || entry.Group != group ||
			entry.Uses != uint32(i)+1 || entry.LastUsed != uint32(i)+1001 {
			t.Fatalf("quest pool entry %d = %+v", i, entry)
		}
	}
	sub_4D0F30()
	if *clock != 1000 {
		t.Fatalf("quest clock = %d", *clock)
	}
	for i := int32(0); i < 3; i++ {
		entry := questMapEntryNative4D0F60(i)
		if entry.Uses != 0 || entry.LastUsed != 0 {
			t.Fatalf("quest history %d = %+v", i, entry)
		}
	}
}
