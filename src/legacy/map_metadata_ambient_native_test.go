package legacy

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/internal/cryptfile"
)

func TestMapMetadataNativeVersionsRetainLegacyFields(t *testing.T) {
	metadata := memmap.PtrT[[1462]byte](0x973F18, 2408)
	fallback1, fallback2 := memmap.PtrUint8(0x5D4594, 741376), memmap.PtrUint8(0x5D4594, 741380)
	original, old1, old2 := *metadata, *fallback1, *fallback2
	t.Cleanup(func() { *metadata, *fallback1, *fallback2 = original, old1, old2 })
	*fallback1, *fallback2 = 'I', 'G'
	for _, version := range []uint16{0, 1, 2, 3} {
		for i := range metadata {
			metadata[i] = 0x93
		}
		want := *metadata
		wire := binary.LittleEndian.AppendUint16(nil, version)
		if version >= 1 {
			for i := range want[:1396] {
				want[i] = byte(i % 251)
			}
			wire = append(wire, want[:1396]...)
			want[1396], want[1397] = 2, 16
		}
		if version == 2 {
			want[1396], want[1397] = 3, 13
			wire = append(wire, 3, 13)
		}
		if version == 3 {
			for _, f := range []struct {
				offset int
				value  string
			}{
				{1398, "QuestIntro"},
				{1430, "QuestGraphics"},
			} {
				wire = append(wire, byte(len(f.value)))
				wire = append(wire, f.value...)
				n := copy(want[f.offset:], f.value)
				want[f.offset+n] = 0
			}
		} else {
			want[1398], want[1430] = 'I', 'G'
		}

		path := filepath.Join(t.TempDir(), "metadata.mapinfo")
		if err := os.WriteFile(path, wire, 0o600); err != nil {
			t.Fatal(err)
		}
		cf, err := cryptfile.OpenFile(path, cryptfile.ReadOnly, -1)
		if err != nil {
			t.Fatal(err)
		}
		if err := mapRWMetadata(cf); err != nil {
			_ = cf.Close()
			t.Fatalf("version %d: %v", version, err)
		}
		if err := cf.Close(); err != nil {
			t.Fatal(err)
		}
		if *metadata != want {
			t.Fatalf("version %d changed unexpected legacy metadata bytes", version)
		}
		if version != 3 {
			continue
		}

		replay := filepath.Join(t.TempDir(), "roundtrip.mapinfo")
		cf, err = cryptfile.OpenFile(replay, cryptfile.WriteOnly, -1)
		if err != nil {
			t.Fatal(err)
		}
		if err := mapRWMetadata(cf); err != nil {
			_ = cf.Close()
			t.Fatal(err)
		}
		if err := cf.Close(); err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(replay)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, wire) {
			t.Fatal("map metadata version 3 did not round-trip byte-for-byte")
		}
	}
}
