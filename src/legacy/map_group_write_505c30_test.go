package legacy

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

type mapGroupReadLegacyServer505C30 struct {
	Server
	srv *server.Server
}

func (s *mapGroupReadLegacyServer505C30) S() *server.Server {
	return s.srv
}

func TestMapGroupNext57C090Nil(t *testing.T) {
	if got := mapGroupNext57C090(nil); got != nil {
		t.Fatalf("nil next = %p, want nil", got)
	}
}

func openMapGroupReadFile505C30(t *testing.T, write func(*cryptfile.CryptFile)) *cryptfile.CryptFile {
	t.Helper()
	path := filepath.Join(t.TempDir(), "groups-read.bin")
	cf, err := cryptfile.OpenFile(path, cryptfile.WriteOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	write(cf)
	if err := cf.Close(); err != nil {
		t.Fatal(err)
	}
	cf, err = cryptfile.OpenFile(path, cryptfile.ReadOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := cf.Close(); err != nil {
			t.Errorf("close map-group fixture: %v", err)
		}
	})
	return cf
}

func writeMapGroupRecord505C30(t *testing.T, cf *cryptfile.CryptFile, name string, kind byte, index uint32, items ...mapGroupItemRecord505C30) {
	t.Helper()
	if len(name) > 0xfe {
		t.Fatalf("fixture name too long: %d", len(name))
	}
	if err := cf.WriteU8(byte(len(name) + 1)); err != nil {
		t.Fatal(err)
	}
	if _, err := cf.Write([]byte(name)); err != nil {
		t.Fatal(err)
	}
	if err := cf.WriteU8(0); err != nil {
		t.Fatal(err)
	}
	if err := cf.WriteU8(kind); err != nil {
		t.Fatal(err)
	}
	if err := cf.WriteU32(index); err != nil {
		t.Fatal(err)
	}
	if err := cf.WriteU32(uint32(len(items))); err != nil {
		t.Fatal(err)
	}
	for _, item := range items {
		if err := cf.WriteU32(item.raw0); err != nil {
			t.Fatal(err)
		}
		if server.MapGroupKind(kind) == server.MapGroupWalls {
			if err := cf.WriteU32(item.raw4); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestMapWriteGroupRecords505C30(t *testing.T) {
	records := []mapGroupRecord505C30{
		{
			kind:  server.MapGroupObjects,
			index: 11,
			name:  "A",
			items: []mapGroupItemRecord505C30{{raw0: 0x11223344}},
		},
		{
			kind:  server.MapGroupWalls,
			index: 22,
			items: []mapGroupItemRecord505C30{{raw0: 1, raw4: 2}, {raw0: 3, raw4: 4}},
		},
	}
	path := filepath.Join(t.TempDir(), "groups.bin")
	cf, err := cryptfile.OpenFile(path, cryptfile.WriteOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	if err := mapWriteGroupRecords505C30(cf, records); err != nil {
		t.Fatal(err)
	}
	if err := cf.Close(); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want, err := hex.DecodeString(
		"0300" +
			"02000000" +
			"024100000b0000000100000044332211" +
			"010002160000000200000001000000020000000300000004000000",
	)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatalf("serialized groups = %x, want %x", got, want)
	}
}

func TestMapReadGroups505C30Version3AllKinds(t *testing.T) {
	want := []mapGroupRecord505C30{
		{kind: server.MapGroupObjects, index: 11, name: "objects", items: []mapGroupItemRecord505C30{{raw0: 0x11223344}}},
		{kind: server.MapGroupWaypoints, index: 12, name: "waypoints", items: []mapGroupItemRecord505C30{{raw0: 7}}},
		{kind: server.MapGroupWalls, index: 13, name: "walls", items: []mapGroupItemRecord505C30{{raw0: 8, raw4: 9}}},
		{kind: server.MapGroupGroups, index: 14, name: "groups", items: []mapGroupItemRecord505C30{{raw0: 10}}},
	}
	cf := openMapGroupReadFile505C30(t, func(cf *cryptfile.CryptFile) {
		if err := mapWriteGroupRecords505C30(cf, want); err != nil {
			t.Fatal(err)
		}
	})
	var got []mapGroupRecord505C30
	err := mapReadGroups505C30(cf, mapGroupReadHooks505C30{
		addGroup: func(name string, index uint32, kind server.MapGroupKind) {
			got = append(got, mapGroupRecord505C30{name: name, index: index, kind: kind})
		},
		addItem: func(_ uint32, _ server.MapGroupKind, item mapGroupItemRecord505C30) {
			got[len(got)-1].items = append(got[len(got)-1].items, item)
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("decoded groups = %#v, want %#v", got, want)
	}
}

func TestNoxServerMapRWGroupData505C30BindsNativeGroups(t *testing.T) {
	srv := new(server.Server)
	srv.MapGroups.Init()
	t.Cleanup(srv.MapGroups.Free)

	oldGetServer := GetServer
	GetServer = func() Server { return &mapGroupReadLegacyServer505C30{srv: srv} }
	t.Cleanup(func() { GetServer = oldGetServer })

	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameHost)
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})

	wallFlags := memmap.PtrUint32(0x5D4594, 739992)
	oldWallFlags := *wallFlags
	*wallFlags = 0
	t.Cleanup(func() { *wallFlags = oldWallFlags })

	cf := openMapGroupReadFile505C30(t, func(cf *cryptfile.CryptFile) {
		if err := mapWriteGroupRecords505C30(cf, []mapGroupRecord505C30{{
			kind:  server.MapGroupWalls,
			index: 29,
			name:  "native-walls",
			items: []mapGroupItemRecord505C30{{raw0: 0x11223344, raw4: 0x55667788}},
		}}); err != nil {
			t.Fatal(err)
		}
	})
	if err := Nox_server_mapRWGroupData_505C30(cf, nil); err != nil {
		t.Fatal(err)
	}
	g := srv.MapGroups.GroupByInd(29)
	if g == nil {
		t.Fatal("native map group was not added")
	}
	if g.ID() != "native-walls" || g.GroupType() != server.MapGroupWalls {
		t.Fatalf("native map group = (%q, %d), want (native-walls, %d)", g.ID(), g.GroupType(), server.MapGroupWalls)
	}
	item := g.First()
	if item == nil || item.Raw0 != 0x11223344 || item.Raw4 != 0x55667788 {
		t.Fatalf("native map group item = %#v", item)
	}
}

func TestMapReadGroups505C30LegacyNames(t *testing.T) {
	tests := []struct {
		name       string
		version    uint16
		stored     string
		currentMap string
		want       string
	}{
		{name: "version one", version: 1, stored: "target", currentMap: "castle", want: "castle.map:target"},
		{name: "version two", version: 2, stored: "old.map::target:tail", currentMap: "castle", want: "castle:target"},
		{name: "signed negative version", version: 0xffff, stored: "target", currentMap: "castle", want: "castle.map:target"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cf := openMapGroupReadFile505C30(t, func(cf *cryptfile.CryptFile) {
				if err := cf.WriteU16(tt.version); err != nil {
					t.Fatal(err)
				}
				if err := cf.WriteU32(1); err != nil {
					t.Fatal(err)
				}
				writeMapGroupRecord505C30(t, cf, tt.stored, byte(server.MapGroupObjects), 7)
			})
			var got string
			err := mapReadGroups505C30(cf, mapGroupReadHooks505C30{
				currentMap: tt.currentMap,
				addGroup: func(name string, _ uint32, _ server.MapGroupKind) {
					got = name
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("qualified name = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMapReadGroups505C30OriginalControlFlow(t *testing.T) {
	t.Run("negative count succeeds", func(t *testing.T) {
		cf := openMapGroupReadFile505C30(t, func(cf *cryptfile.CryptFile) {
			if err := cf.WriteU16(3); err != nil {
				t.Fatal(err)
			}
			if err := cf.WriteU32(0xffffffff); err != nil {
				t.Fatal(err)
			}
		})
		if err := mapReadGroups505C30(cf, mapGroupReadHooks505C30{}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("skip still parses", func(t *testing.T) {
		cf := openMapGroupReadFile505C30(t, func(cf *cryptfile.CryptFile) {
			if err := cf.WriteU16(3); err != nil {
				t.Fatal(err)
			}
			if err := cf.WriteU32(1); err != nil {
				t.Fatal(err)
			}
			writeMapGroupRecord505C30(t, cf, "walls", byte(server.MapGroupWalls), 9, mapGroupItemRecord505C30{raw0: 1, raw4: 2})
		})
		var calls int
		err := mapReadGroups505C30(cf, mapGroupReadHooks505C30{
			skip:     true,
			addGroup: func(string, uint32, server.MapGroupKind) { calls++ },
			addItem:  func(uint32, server.MapGroupKind, mapGroupItemRecord505C30) { calls++ },
		})
		if err != nil {
			t.Fatal(err)
		}
		if calls != 0 {
			t.Fatalf("skip hooks called %d times", calls)
		}
	})

	t.Run("invalid type applies group before failing item", func(t *testing.T) {
		cf := openMapGroupReadFile505C30(t, func(cf *cryptfile.CryptFile) {
			if err := cf.WriteU16(3); err != nil {
				t.Fatal(err)
			}
			if err := cf.WriteU32(1); err != nil {
				t.Fatal(err)
			}
			writeMapGroupRecord505C30(t, cf, "bad", 4, 19, mapGroupItemRecord505C30{raw0: 1})
		})
		var groups, items int
		err := mapReadGroups505C30(cf, mapGroupReadHooks505C30{
			addGroup: func(string, uint32, server.MapGroupKind) { groups++ },
			addItem:  func(uint32, server.MapGroupKind, mapGroupItemRecord505C30) { items++ },
		})
		if err == nil || !strings.Contains(err.Error(), "invalid map group type") {
			t.Fatalf("error = %v, want invalid map group type", err)
		}
		if groups != 1 || items != 0 {
			t.Fatalf("hook calls = groups %d, items %d; want 1, 0", groups, items)
		}
	})

	t.Run("invalid type with no items succeeds", func(t *testing.T) {
		cf := openMapGroupReadFile505C30(t, func(cf *cryptfile.CryptFile) {
			if err := cf.WriteU16(3); err != nil {
				t.Fatal(err)
			}
			if err := cf.WriteU32(1); err != nil {
				t.Fatal(err)
			}
			writeMapGroupRecord505C30(t, cf, "empty-bad", 4, 20)
		})
		if err := mapReadGroups505C30(cf, mapGroupReadHooks505C30{}); err != nil {
			t.Fatal(err)
		}
	})
}

func TestMapReadGroups505C30RejectsMalformedInput(t *testing.T) {
	t.Run("future version", func(t *testing.T) {
		cf := openMapGroupReadFile505C30(t, func(cf *cryptfile.CryptFile) {
			if err := cf.WriteU16(4); err != nil {
				t.Fatal(err)
			}
		})
		if err := mapReadGroups505C30(cf, mapGroupReadHooks505C30{}); err == nil {
			t.Fatal("future version succeeded")
		}
	})

	t.Run("version two without second token", func(t *testing.T) {
		cf := openMapGroupReadFile505C30(t, func(cf *cryptfile.CryptFile) {
			if err := cf.WriteU16(2); err != nil {
				t.Fatal(err)
			}
			if err := cf.WriteU32(1); err != nil {
				t.Fatal(err)
			}
			writeMapGroupRecord505C30(t, cf, "one-token", byte(server.MapGroupObjects), 1)
		})
		if err := mapReadGroups505C30(cf, mapGroupReadHooks505C30{currentMap: "map"}); err == nil {
			t.Fatal("malformed version two name succeeded")
		}
	})

	t.Run("oversized name", func(t *testing.T) {
		cf := openMapGroupReadFile505C30(t, func(cf *cryptfile.CryptFile) {
			if err := cf.WriteU16(3); err != nil {
				t.Fatal(err)
			}
			if err := cf.WriteU32(1); err != nil {
				t.Fatal(err)
			}
			if err := cf.WriteU8(mapGroupNameCapacity505C30 + 1); err != nil {
				t.Fatal(err)
			}
			if _, err := cf.Write([]byte(strings.Repeat("x", mapGroupNameCapacity505C30+1))); err != nil {
				t.Fatal(err)
			}
		})
		if err := mapReadGroups505C30(cf, mapGroupReadHooks505C30{}); err == nil {
			t.Fatal("oversized name succeeded")
		}
	})

	t.Run("full native buffer without terminator", func(t *testing.T) {
		cf := openMapGroupReadFile505C30(t, func(cf *cryptfile.CryptFile) {
			if err := cf.WriteU16(3); err != nil {
				t.Fatal(err)
			}
			if err := cf.WriteU32(1); err != nil {
				t.Fatal(err)
			}
			if err := cf.WriteU8(mapGroupNameCapacity505C30); err != nil {
				t.Fatal(err)
			}
			if _, err := cf.Write([]byte(strings.Repeat("x", mapGroupNameCapacity505C30))); err != nil {
				t.Fatal(err)
			}
		})
		if err := mapReadGroups505C30(cf, mapGroupReadHooks505C30{}); err == nil {
			t.Fatal("unterminated native-capacity name succeeded")
		}
	})

	t.Run("truncated wall", func(t *testing.T) {
		cf := openMapGroupReadFile505C30(t, func(cf *cryptfile.CryptFile) {
			if err := cf.WriteU16(3); err != nil {
				t.Fatal(err)
			}
			if err := cf.WriteU32(1); err != nil {
				t.Fatal(err)
			}
			if err := cf.WriteU8(2); err != nil {
				t.Fatal(err)
			}
			if _, err := cf.Write([]byte{'w', 0}); err != nil {
				t.Fatal(err)
			}
			if err := cf.WriteU8(byte(server.MapGroupWalls)); err != nil {
				t.Fatal(err)
			}
			if err := cf.WriteU32(5); err != nil {
				t.Fatal(err)
			}
			if err := cf.WriteU32(1); err != nil {
				t.Fatal(err)
			}
			if err := cf.WriteU32(7); err != nil {
				t.Fatal(err)
			}
		})
		if err := mapReadGroups505C30(cf, mapGroupReadHooks505C30{}); err == nil {
			t.Fatal("truncated wall succeeded")
		}
	})
}

func TestMapWriteGroupRecords505C30RejectsNativeNameOverflow(t *testing.T) {
	path := filepath.Join(t.TempDir(), "groups-long.bin")
	cf, err := cryptfile.OpenFile(path, cryptfile.WriteOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cf.Close() })
	err = mapWriteGroupRecords505C30(cf, []mapGroupRecord505C30{{
		kind: server.MapGroupObjects,
		name: strings.Repeat("x", mapGroupNameCapacity505C30),
	}})
	if err == nil {
		t.Fatal("native name overflow succeeded")
	}
}

func TestMapWriteGroupRecords505C30AllowsNativeNameLimit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "groups-max-name.bin")
	cf, err := cryptfile.OpenFile(path, cryptfile.WriteOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cf.Close() })
	err = mapWriteGroupRecords505C30(cf, []mapGroupRecord505C30{{
		kind: server.MapGroupObjects,
		name: strings.Repeat("x", mapGroupNameCapacity505C30-1),
	}})
	if err != nil {
		t.Fatal(err)
	}
}
