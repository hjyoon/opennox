package legacy

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	crypt "github.com/opennox/noxcrypt"
	"github.com/opennox/opennox/v1/internal/cryptfile"
)

func TestCryptFileCExportItemBoundaries236(t *testing.T) {
	testCryptFileItemBoundaries236(t, cryptFileExportCall236)
}

func testCryptFileItemBoundaries236(t *testing.T, call func([]byte) uint64) {
	for _, mode := range []cryptfile.Mode{cryptfile.ReadOnly, cryptfile.WriteOnly, cryptfile.Append} {
		for _, key := range []int{-1, crypt.MapKey} {
			for _, xor := range []bool{false, true} {
				for _, null := range []bool{false, true} {
					t.Run(fmt.Sprintf("mode=%d/key=%d/xor=%t/null=%t", mode, key, xor, null), func(t *testing.T) {
						plain := make([]byte, 32)
						for i := range plain {
							plain[i] = byte(i*37 + 11)
						}
						path := filepath.Join(t.TempDir(), "owned-items.bin")
						if err := os.WriteFile(path, cryptFileWire236(t, plain, key, xor), 0o600); err != nil {
							t.Fatal(err)
						}
						cf := cryptFileGlobal236(t, path, mode, key)
						cf.SetXOR(xor)
						before, err := os.ReadFile(path)
						if err != nil {
							t.Fatal(err)
						}
						position := cf.File.Written()
						guard := bytes.Repeat([]byte{0xa5}, 32)
						buf := guard[16:16]
						if null {
							buf = nil
						}
						for repeat := 0; repeat < 4; repeat++ {
							if got := call(buf); got != 0 {
								t.Errorf("empty item returned %d, original fread/fwrite return 0", got)
							}
							if cf.File.Written() != position || !bytes.Equal(guard, bytes.Repeat([]byte{0xa5}, 32)) {
								t.Fatal("zero item changed position or buffer canaries")
							}
						}
						after, err := os.ReadFile(path)
						if err != nil || !bytes.Equal(after, before) {
							t.Fatalf("empty I/O changed file bytes: %x -> %x, %v", before, after, err)
						}
						data := []byte{11, 48, 85}
						if mode == cryptfile.ReadOnly {
							clear(data)
						}
						if got := call(data); got != 1 || !bytes.Equal(data, plain[:3]) {
							t.Fatalf("complete item = %d, bytes %x; want 1, %x", got, data, plain[:3])
						}
						if mode == cryptfile.ReadOnly {
							tail := bytes.Repeat([]byte{0xa5}, 64)
							want := append(append([]byte(nil), plain[3:]...), bytes.Repeat([]byte{0xa5}, 35)...)
							if xor {
								for i := 29; i < len(want); i++ {
									want[i] ^= 126
								}
							}
							if got := call(tail); got != 0 || !bytes.Equal(tail, want) {
								t.Fatalf("partial item = %d, bytes %x; want 0, %x", got, tail, want)
							}
						} else {
							if err := cf.Close(); err != nil {
								t.Fatal(err)
							}
							got, err := os.ReadFile(path)
							want := append([]byte(nil), data...)
							if mode == cryptfile.Append {
								want = append(append([]byte(nil), plain...), data...)
							}
							if err != nil || !bytes.Equal(got, cryptFileWire236(t, want, key, xor)) {
								t.Fatalf("follow-up write bytes = %x, %v", got, err)
							}
						}
					})
				}
			}
		}
	}
}

// Infer C's integer types from the actual exported function signature. This
// exercises its Go body without importing C or depending on generated names.
func cryptFileExportTyped236[B ~uint8, S ~uint32 | ~uint64, Ch ~int8 | ~uint8](fn func(*B, S, *Ch, int32) S, p []byte) uint64 {
	return uint64(fn((*B)(unsafe.Pointer(unsafe.SliceData(p))), S(len(p)), nil, 0))
}

func cryptFileExportCall236(p []byte) uint64 {
	return cryptFileExportTyped236(nox_xxx_fileReadWrite_426AC0_file3_fread_impl, p)
}

func cryptFileGlobal236(t *testing.T, path string, mode cryptfile.Mode, key int) *cryptfile.CryptFile {
	t.Helper()
	cf, err := cryptfile.OpenFile(path, mode, key)
	if err != nil {
		t.Fatal(err)
	}
	previous := cryptfile.Global()
	cryptfile.SetGlobal(cf)
	t.Cleanup(func() {
		cryptfile.SetGlobal(previous)
		if err := cf.Close(); err != nil {
			t.Error(err)
		}
	})
	return cf
}

func cryptFileWire236(t *testing.T, plain []byte, key int, xor bool) []byte {
	t.Helper()
	out := append([]byte(nil), plain...)
	if xor {
		for i := range out {
			out[i] ^= 126
		}
	} else if key >= 0 {
		out = append(out, make([]byte, (crypt.Block-len(out)%crypt.Block)%crypt.Block)...)
		c, err := crypt.NewCipher(key)
		if err != nil {
			t.Fatal(err)
		}
		for off := 0; off < len(out); off += crypt.Block {
			c.Encrypt(out[off:off+crypt.Block], out[off:off+crypt.Block])
		}
	}
	return out
}
