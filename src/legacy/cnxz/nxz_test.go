package cnxz

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/opennox/libs/ifs"
	"github.com/opennox/libs/noxtest"
	"github.com/stretchr/testify/require"
)

func TestDecompress(t *testing.T) {
	maps := noxtest.DataPath(t, "maps")
	files, err := os.ReadDir(maps)
	require.NoError(t, err)
	checked := 0
	for _, fi := range files {
		mname := filepath.Join(maps, fi.Name(), fi.Name()+".map")
		zname := filepath.Join(maps, fi.Name(), fi.Name()+".nxz")
		if _, err = ifs.Stat(zname); err != nil {
			continue
		}
		checked++
		t.Run(fi.Name(), func(t *testing.T) {
			mexp, mexpN := hashFile(t, mname)
			gotc, gotcN := decompress(t, zname)
			require.Equal(t, mexpN, gotcN)
			require.Equal(t, mexp, gotc)
		})
	}
	requireStrictOracleMapCount(t, checked)
}

func TestCompress(t *testing.T) {
	maps := noxtest.DataPath(t, "maps")
	files, err := os.ReadDir(maps)
	require.NoError(t, err)
	checked := 0
	for _, fi := range files {
		mname := filepath.Join(maps, fi.Name(), fi.Name()+".map")
		zname := filepath.Join(maps, fi.Name(), fi.Name()+".nxz")
		if _, err = ifs.Stat(zname); err != nil {
			continue
		}
		checked++
		t.Run(fi.Name(), func(t *testing.T) {
			mexp, mexpN := hashFile(t, mname)
			dir := t.TempDir()
			compressed := filepath.Join(dir, fi.Name()+".nxz")
			decoded := filepath.Join(dir, fi.Name()+".map")
			require.NoError(t, CompressFile(mname, compressed))
			require.NoError(t, DecompressFile(compressed, decoded))
			got, gotN := hashFile(t, decoded)
			require.Equal(t, mexpN, gotN)
			require.Equal(t, mexp, got)
		})
	}
	requireStrictOracleMapCount(t, checked)
}

func TestPortableRoundTrip(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "input.map")
	compressed := filepath.Join(dir, "input.nxz")
	decoded := filepath.Join(dir, "decoded.map")
	// Exercise all literals, a window wrap, and data beyond the legacy
	// compressor's 500,000-byte chunk boundary.
	want := bytes.Repeat([]byte{
		0x00, 0x01, 0x02, 0x20, 0x30, 0x7f, 0x80, 0xfe, 0xff,
	}, 60_000)
	require.NoError(t, os.WriteFile(src, want, 0o600))
	require.NoError(t, CompressFile(src, compressed))
	require.NoError(t, DecompressFile(compressed, decoded))
	got, err := os.ReadFile(decoded)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func requireStrictOracleMapCount(t testing.TB, got int) {
	t.Helper()
	if os.Getenv("NOX_ORACLE_STRICT") != "1" {
		return
	}
	const want = 50
	require.Equal(t, want, got, "sealed nox-2023-1003-01 map-pair count")
}

func decompress(t testing.TB, path string) (string, int) {
	out, err := os.CreateTemp("", "nxzmap_*.map")
	require.NoError(t, err)
	defer func() {
		out.Close()
		_ = os.Remove(out.Name())
	}()
	err = DecompressFile(path, out.Name())
	require.NoError(t, err)
	return hashFile(t, out.Name())
}

func hashReader(t testing.TB, r io.Reader) (string, int) {
	h := sha256.New()
	n, err := io.Copy(h, r)
	require.NoError(t, err)
	return hex.EncodeToString(h.Sum(nil)), int(n)
}

func hashFile(t testing.TB, path string) (string, int) {
	f, err := ifs.Open(path)
	require.NoError(t, err)
	defer f.Close()
	return hashReader(t, f)
}
