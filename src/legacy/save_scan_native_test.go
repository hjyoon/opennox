package legacy

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/opennox/libs/common"
	"github.com/opennox/libs/datapath"
	crypt "github.com/opennox/noxcrypt"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/legacy/internal/cfixture"
)

func TestCountSaveFilesNativeCEntry4DC550(t *testing.T) {
	var allSlots []string
	for i := 0; i < 16; i++ {
		allSlots = append(allSlots, fmt.Sprintf(common.SaveFormat, i))
	}
	for _, tc := range []struct {
		name  string
		auto  bool
		slots []string
		want  int
	}{
		{name: "no saves"},
		{name: "gap ends manual count", slots: []string{"SAVE0000", "SAVE0001", "SAVE0003"}, want: 2},
		{name: "autosave starts manual count at one", auto: true, slots: []string{"SAVE0000", "SAVE0002"}, want: 1},
		{name: "manual slots count is bounded", slots: allSlots, want: 13},
		{name: "autosave plus thirteen manual slots", auto: true, slots: allSlots, want: 14},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := setSaveScanDataRoot(t)
			slots := append([]string(nil), tc.slots...)
			if tc.auto {
				slots = append(slots, common.SaveAuto)
			}
			for _, slot := range slots {
				dir := filepath.Join(root, "Save", slot)
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "Player.plr"), nil, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if got := cfixture.CountSaveFilesC(); got != tc.want {
				t.Fatalf("C entry counted %d save slots, want %d", got, tc.want)
			}
			if dir, err := os.Stat(filepath.Join(root, "Save")); err != nil || !dir.IsDir() {
				t.Fatalf("expected Save directory, got %v, %v", dir, err)
			}
		})
	}
}

func TestCountPlayerFilesNativeCEntry4DC630(t *testing.T) {
	root := setSaveScanDataRoot(t)
	// Initialize the two table-presence words checked by the header reader.
	first := memmap.PtrUint32(0x587000, 55936)
	next := memmap.PtrUint32(0x587000, 55948)
	oldFirst, oldNext := *first, *next
	*first, *next = 1, 1
	t.Cleanup(func() { *first, *next = oldFirst, oldNext })
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})
	t.Chdir(t.TempDir())
	if got := cfixture.CountPlayerFilesC(); got != 0 {
		t.Fatalf("empty Save directory has %d characters", got)
	}
	dir := filepath.Join(root, "Save")
	if err := os.Mkdir(filepath.Join(dir, "folder.plr"), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, flags := range map[string]uint32{
		"coop.plr": 1, "online.plr": 2, "mixed.plr": 6, "quest.plr": 4, "ignored.bin": 2,
	} {
		writePlayerScanHeader(t, filepath.Join(dir, name), flags)
	}
	if got := cfixture.CountPlayerFilesC(); got != 2 {
		t.Fatalf("C entry counted %d online characters, want 2", got)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.Stat(cwd)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.Stat(root)
	if err != nil || !os.SameFile(got, want) {
		t.Fatalf("working directory is %q, want %q", cwd, root)
	}
}

func setSaveScanDataRoot(t *testing.T) string {
	t.Helper()
	old := datapath.Data()
	root := t.TempDir()
	datapath.SetData(root)
	t.Cleanup(func() { datapath.SetData(old) })
	return root
}

func writePlayerScanHeader(t *testing.T, path string, flags uint32) {
	t.Helper()
	f, err := cryptfile.OpenFile(path, cryptfile.WriteOnly, crypt.SaveKey)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := f.WriteU32(1); err != nil {
		t.Fatal(err)
	}
	f.SectionStart()
	var header [49]byte
	binary.LittleEndian.PutUint16(header[:2], 10)
	binary.LittleEndian.PutUint32(header[2:6], flags)
	if _, err := f.Write(header[:]); err != nil {
		t.Fatal(err)
	}
	f.SectionEnd()
	if err := f.WriteU32(0); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}
