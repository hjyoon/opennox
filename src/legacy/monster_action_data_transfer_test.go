package legacy

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/opennox/libs/datapath"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestMonsterActionDataLegacyVersionAlignment(t *testing.T) {
	// These records come from the stock 529CE0 execution with the
	// historical version selected by the original file service.
	cases := []struct {
		name      string
		loadFrame uint32
		wire      string
		deadlines [3]uint32
	}{
		{
			name: "v1", loadFrame: 0xdaf48c12,
			wire: "0100e99c355900000000000000005976ee154b79b44d8b5c5f22f1398c36db0000000014884c404e" +
				"d4260cefe3e600270a1977ceb84443612ea30a006519476a273f0f1fc7cd581fb2f9b7000000000d" +
				"fe8f9c56b3356894a160db00000000c2491c3e6711e610ff0000000000e6d6afe294ab3cb8c187c7" +
				"e4a58e0e54656d706c617465416374696f6e",
			deadlines: [3]uint32{0x00000001, 0x1e4eed36, 0x00000001},
		},
		{
			name: "v2", loadFrame: 0x89c12ebc,
			wire: "02000175059468000000000000000021b3e51da54b8b46acc994859f819cb2d2000000005f8aaefe" +
				"1dd47e68c4a56eabf2f0a6c778221c9e843e1d9000178bf8a68319e624bf9896e2323de200000000" +
				"3edfc72da7e90041a35b17d800000000034834a545168355ff000000000045c90536722eca30a2b9" +
				"498f5c430e54656d706c617465416374696f6e",
			deadlines: [3]uint32{0x00000001, 0x4ef50885, 0x76b03f8c},
		},
	}
	oldData := datapath.Data()
	datapath.SetData(t.TempDir())
	t.Cleanup(func() { datapath.SetData(oldData) })
	srv := server.NewHigh(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	if !srv.Objs.InitHigh(8) {
		t.Fatal("init object allocator")
	}
	t.Cleanup(srv.Objs.FreeObjects)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	if err := srv.Types.ReadObjectType(&things.Thing{Name: "TestActionData"}); err != nil {
		t.Fatal(err)
	}
	obj := srv.Objs.NewObject(srv.Types.ByID("TestActionData"))
	obj.ObjClass = object.ClassMonster
	update, release := alloc.New(server.MonsterUpdateData{})
	t.Cleanup(release)
	obj.UpdateData = unsafe.Pointer(update)
	for _, row := range cases {
		t.Run(row.name, func(t *testing.T) {
			stream, err := hex.DecodeString(row.wire)
			if err != nil {
				t.Fatal(err)
			}
			filename := filepath.Join(t.TempDir(), "action-data.bin")
			if err := os.WriteFile(filename, append(stream, 0x5a), 0o600); err != nil {
				t.Fatal(err)
			}
			cf, err := cryptfile.OpenFile(filename, cryptfile.ReadOnly, -1)
			if err != nil {
				t.Fatal(err)
			}
			defer cf.Close()
			*update = server.MonsterUpdateData{AIStackInd: -1}
			srv.SetFrame(row.loadFrame)
			var refs server.MonsterActionRefs
			if err := monsterXferActionData529CE0(cf, obj, &refs, false); err != nil {
				t.Fatal(err)
			}
			if got, err := cf.ReadU8(); err != nil || got != 0x5a {
				t.Fatalf("next field = %02x, %v; want 5a", got, err)
			}
			if got := [3]uint32{update.Field70, update.Field124, update.Field135}; got != row.deadlines {
				t.Fatalf("deadlines = %08x; want %08x", got, row.deadlines)
			}
		})
	}
}
