package legacy

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
)

func mapIntroWire505080(version uint16, size uint32, payload []byte) []byte {
	wire := make([]byte, 6, 6+len(payload))
	binary.LittleEndian.PutUint16(wire[0:2], version)
	binary.LittleEndian.PutUint32(wire[2:6], size)
	return append(wire, payload...)
}

func openMapIntroRead505080(t *testing.T, wire []byte) *cryptfile.CryptFile {
	t.Helper()
	path := filepath.Join(t.TempDir(), "map-intro-read.bin")
	if err := os.WriteFile(path, wire, 0o600); err != nil {
		t.Fatal(err)
	}
	cf, err := cryptfile.OpenFile(path, cryptfile.ReadOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cf.Close() })
	return cf
}

func openMapIntroWrite505080(t *testing.T) (*cryptfile.CryptFile, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "map-intro-write.bin")
	cf, err := cryptfile.OpenFile(path, cryptfile.WriteOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cf.Close() })
	return cf, path
}

func TestMapReadIntro505080StoresNativeWidthPayload(t *testing.T) {
	clearMapIntro505080()
	storeMapIntro505080([]byte("stale"))
	payload := []byte("Horvath defeats the urchins with lightning.")
	cf := openMapIntroRead505080(t, mapIntroWire505080(1, uint32(len(payload)), payload))
	if err := mapRWMapIntro505080(cf, "unused", mapIntroHooks505080{
		clear: clearMapIntro505080,
		store: storeMapIntro505080,
	}); err != nil {
		t.Fatal(err)
	}
	if got := string(mapIntroSnapshot505080()); got != string(payload) {
		t.Fatalf("stored payload = %q, want %q", got, payload)
	}
	mapIntroState505080.RLock()
	ptr := uintptr(unsafe.Pointer(&mapIntroState505080.data[0]))
	mapIntroState505080.RUnlock()
	if unsafe.Sizeof(uintptr(0)) == 8 && ptr <= math.MaxUint32 {
		t.Fatalf("payload address = %#x, want native-width address", ptr)
	}
}

func TestMapReadIntro505080ParseOnlyTakesPrecedence(t *testing.T) {
	payload := []byte("skip this payload")
	cf := openMapIntroRead505080(t, mapIntroWire505080(1, uint32(len(payload)), payload))
	created := false
	stored := false
	if err := mapRWMapIntro505080(cf, "unused", mapIntroHooks505080{
		external:  true,
		parseOnly: true,
		create: func(string) (io.WriteCloser, error) {
			created = true
			return nil, errors.New("unexpected create")
		},
		store: func([]byte) { stored = true },
	}); err != nil {
		t.Fatal(err)
	}
	if created || stored {
		t.Fatalf("parse-only created=%t stored=%t, want both false", created, stored)
	}
	pos, err := cf.File.Seek(0, io.SeekCurrent)
	if err != nil {
		t.Fatal(err)
	}
	if want := int64(6 + len(payload)); pos != want {
		t.Fatalf("cursor = %d, want %d", pos, want)
	}
}

func TestMapReadIntro505080ExternalFile(t *testing.T) {
	payload := []byte{0, 1, 2, 3, 0xff}
	cf := openMapIntroRead505080(t, mapIntroWire505080(1, uint32(len(payload)), payload))
	outPath := filepath.Join(t.TempDir(), "intro.txt")
	const logicalPath = "logical-map-intro"
	if err := mapRWMapIntro505080(cf, logicalPath, mapIntroHooks505080{
		external: true,
		create: func(path string) (io.WriteCloser, error) {
			if path != logicalPath {
				t.Fatalf("create path = %q, want %q", path, logicalPath)
			}
			return os.Create(outPath)
		},
		store: func([]byte) { t.Fatal("external read stored an in-memory payload") },
	}); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(payload) {
		t.Fatalf("external payload = %x, want %x", got, payload)
	}
}

func TestMapReadIntro505080VersionsAndSignedLengths(t *testing.T) {
	for _, tc := range []struct {
		name    string
		version uint16
		wantErr bool
	}{
		{name: "zero", version: 0, wantErr: true},
		{name: "current", version: 1},
		{name: "future-signed", version: 0x7fff},
		{name: "negative", version: 0x8000, wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cf := openMapIntroRead505080(t, mapIntroWire505080(tc.version, 0, nil))
			err := mapRWMapIntro505080(cf, "unused", mapIntroHooks505080{})
			if (err != nil) != tc.wantErr {
				t.Fatalf("error = %v, wantErr=%t", err, tc.wantErr)
			}
		})
	}

	cf := openMapIntroRead505080(t, mapIntroWire505080(1, 0x80000001, nil))
	stored := false
	if err := mapRWMapIntro505080(cf, "unused", mapIntroHooks505080{
		store: func([]byte) { stored = true },
	}); err != nil {
		t.Fatal(err)
	}
	if stored {
		t.Fatal("signed-negative length stored a payload")
	}
	pos, err := cf.File.Seek(0, io.SeekCurrent)
	if err != nil {
		t.Fatal(err)
	}
	if pos != 6 {
		t.Fatalf("cursor = %d, want 6", pos)
	}
}

func TestMapReadIntro505080TruncatedPayloadClearsOldData(t *testing.T) {
	storeMapIntro505080([]byte("stale"))
	cf := openMapIntroRead505080(t, mapIntroWire505080(1, 8, []byte("short")))
	err := mapRWMapIntro505080(cf, "unused", mapIntroHooks505080{
		clear: clearMapIntro505080,
		store: storeMapIntro505080,
	})
	if err == nil {
		t.Fatal("truncated payload succeeded")
	}
	if got := mapIntroSnapshot505080(); got != nil {
		t.Fatalf("payload after failed read = %x, want nil", got)
	}
}

func TestMapWriteIntro505080ExternalFile(t *testing.T) {
	payload := []byte("external map intro")
	source := filepath.Join(t.TempDir(), "intro.txt")
	if err := os.WriteFile(source, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	cf, path := openMapIntroWrite505080(t)
	const logicalPath = "logical-map-intro"
	cleared := false
	if err := mapRWMapIntro505080(cf, logicalPath, mapIntroHooks505080{
		external: true,
		clear:    func() { cleared = true },
		open: func(path string) (mapIntroReadSeekCloser505080, error) {
			if path != logicalPath {
				t.Fatalf("open path = %q, want %q", path, logicalPath)
			}
			return os.Open(source)
		},
	}); err != nil {
		t.Fatal(err)
	}
	if !cleared {
		t.Fatal("writer did not clear the previous in-memory payload")
	}
	if err := cf.Close(); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := mapIntroWire505080(1, uint32(len(payload)), payload)
	if string(got) != string(want) {
		t.Fatalf("serialized intro = %x, want %x", got, want)
	}
}

func TestMapWriteIntro505080AbsentOrDisabledExternalFile(t *testing.T) {
	for _, tc := range []struct {
		name     string
		external bool
	}{
		{name: "disabled"},
		{name: "missing", external: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cf, path := openMapIntroWrite505080(t)
			opened := false
			err := mapRWMapIntro505080(cf, "missing", mapIntroHooks505080{
				external: tc.external,
				open: func(string) (mapIntroReadSeekCloser505080, error) {
					opened = true
					return nil, os.ErrNotExist
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			if err := cf.Close(); err != nil {
				t.Fatal(err)
			}
			got, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			want := mapIntroWire505080(1, 0, nil)
			if string(got) != string(want) {
				t.Fatalf("serialized intro = %x, want %x", got, want)
			}
			if opened != tc.external {
				t.Fatalf("opened = %t, want %t", opened, tc.external)
			}
		})
	}
}

type oversizedMapIntro505080 struct {
	size   int64
	closed bool
}

func (f *oversizedMapIntro505080) Read([]byte) (int, error) { return 0, io.EOF }

func (f *oversizedMapIntro505080) Seek(_ int64, whence int) (int64, error) {
	if whence == io.SeekEnd {
		return f.size, nil
	}
	return 0, nil
}

func (f *oversizedMapIntro505080) Close() error {
	f.closed = true
	return nil
}

func TestMapWriteIntro505080RejectsUnsignedLength(t *testing.T) {
	cf, _ := openMapIntroWrite505080(t)
	in := &oversizedMapIntro505080{size: math.MaxInt32 + 1}
	err := mapRWMapIntro505080(cf, "oversized", mapIntroHooks505080{
		external: true,
		open: func(string) (mapIntroReadSeekCloser505080, error) {
			return in, nil
		},
	})
	if err == nil {
		t.Fatal("unsigned map-intro length succeeded")
	}
	if !in.closed {
		t.Fatal("oversized map-intro source was not closed")
	}
}

func TestMapIntroPath505080(t *testing.T) {
	root := filepath.Join("root", "data")
	got := mapIntroPath505080(root, "wizard1a")
	want := filepath.Join(root, "maps", "wizard1a", "wizard1a.txt")
	if got != want {
		t.Fatalf("path = %q, want %q", got, want)
	}
}

func TestMapRWIntro505080RejectsNilCryptFile(t *testing.T) {
	cleared := false
	if err := mapRWMapIntro505080(nil, "unused", mapIntroHooks505080{
		clear: func() { cleared = true },
	}); err == nil {
		t.Fatal("nil crypt file succeeded")
	}
	if cleared {
		t.Fatal("nil crypt file changed payload state")
	}
}
