package legacy

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

func TestMapGroupNext57C090Nil(t *testing.T) {
	if got := mapGroupNext57C090(nil); got != nil {
		t.Fatalf("nil next = %p, want nil", got)
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
