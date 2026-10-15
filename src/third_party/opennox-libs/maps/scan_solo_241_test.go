package maps_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/opennox/libs/maps"
)

func scanFixture241(t *testing.T, root, name string, flags uint32) maps.Info {
	t.Helper()
	dir := filepath.Join(root, name)
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(dir, name+maps.Ext))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	info := maps.MapInfo{Format: 2, Summary: name, Flags: flags, MinPlayers: 2, MaxPlayers: 16}
	w, err := maps.NewWriter(f, maps.Header{Magic: maps.MagicOld})
	if err != nil {
		t.Fatal(err)
	}
	if err := w.WriteSection(&info); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	stat, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	return maps.Info{Filename: name, Size: int(stat.Size()), MapInfo: info}
}

func TestScanSoloOption241(t *testing.T) {
	root := t.TempDir()
	var multiplayer, all maps.MapList
	for _, row := range []struct {
		name  string
		flags uint32
		solo  bool
	}{
		{"z_arena", 4, false},
		{"Arena_war01a", 4, false},
		{"xWIZ01a", 4, false},
		{"xcon01a", 4, false},
		{"War01a", 1, true},
		{"WIZ02b", 1, true},
		{"Con03c", 1, true},
		// Keep the existing prefix rule, even for non-campaign metadata.
		{"warrens", 4, true},
		{"wizard_arena", 4, true},
		{"conclave", 4, true},
	} {
		info := scanFixture241(t, root, row.name, row.flags)
		all = append(all, info)
		if !row.solo {
			multiplayer = append(multiplayer, info)
		}
	}
	// Exercise multiple ReadDir pages without requiring installed game data.
	for i := 0; i < 103; i++ {
		info := scanFixture241(t, root, fmt.Sprintf("Arena%03d", i), 4)
		all = append(all, info)
		multiplayer = append(multiplayer, info)
	}
	if err := os.Mkdir(filepath.Join(root, "Empty"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "NotADirectory.map"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	for _, list := range []maps.MapList{multiplayer, all} {
		sort.Slice(list, func(i, j int) bool { return list[i].Filename < list[j].Filename })
	}
	for _, row := range []struct {
		name string
		opts *maps.ScanOptions
		want maps.MapList
	}{
		{"default", nil, multiplayer},
		{"exclude_solo", &maps.ScanOptions{Solo: false}, multiplayer},
		{"include_all_three_classes", &maps.ScanOptions{Solo: true}, all},
	} {
		t.Run(row.name, func(t *testing.T) {
			got, err := maps.Scan(nil, root, row.opts)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, row.want) {
				t.Fatalf("map catalog differs: got %d entries, want %d; all metadata, names and sort order must match", len(got), len(row.want))
			}
		})
	}
}

func TestScanSoloReadErrors241(t *testing.T) {
	for _, name := range []string{"WarBroken", "WIZBroken", "ConBroken"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			want := maps.MapList{scanFixture241(t, root, "Arena", 4)}
			dir := filepath.Join(root, name)
			if err := os.Mkdir(dir, 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, name+maps.Ext), nil, 0600); err != nil {
				t.Fatal(err)
			}
			for _, row := range []struct {
				name string
				opts *maps.ScanOptions
				bad  bool
			}{
				{"default", nil, false},
				{"exclude_solo", &maps.ScanOptions{}, false},
				{"include_solo", &maps.ScanOptions{Solo: true}, true},
			} {
				t.Run(row.name, func(t *testing.T) {
					var messages bytes.Buffer
					logger := slog.New(slog.NewTextHandler(&messages, nil))
					got, err := maps.Scan(logger, root, row.opts)
					if row.bad {
						if !errors.Is(err, io.EOF) || !strings.Contains(messages.String(), "invalid map file") || !strings.Contains(messages.String(), "name="+name) {
							t.Fatalf("included invalid map must return its read error and warning: err=%v log=%q", err, messages.String())
						}
					} else if err != nil || messages.Len() != 0 {
						t.Fatalf("excluded invalid map was read: err=%v log=%q", err, messages.String())
					}
					if !reflect.DeepEqual(got, want) {
						t.Fatal("valid map was lost after a skipped or invalid solo map")
					}
				})
			}
		})
	}
}

func TestScanMissingRoot241(t *testing.T) {
	root := filepath.Join(t.TempDir(), "absent")
	for _, row := range []struct {
		name string
		opts *maps.ScanOptions
	}{
		{"default", nil},
		{"exclude_solo", &maps.ScanOptions{}},
		{"include_solo", &maps.ScanOptions{Solo: true}},
	} {
		t.Run(row.name, func(t *testing.T) {
			got, err := maps.Scan(nil, root, row.opts)
			if got != nil || !os.IsNotExist(err) {
				t.Fatalf("missing root semantics changed: list=%v err=%v", got, err)
			}
		})
	}
}
