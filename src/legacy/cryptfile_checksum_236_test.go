package legacy

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	crypt "github.com/opennox/noxcrypt"
	"github.com/opennox/opennox/v1/internal/cryptfile"
)

func TestCryptFileCExportEmptyWriteChecksum236(t *testing.T) {
	testCryptFileEmptyWriteChecksum236(t, cryptFileExportCall236)
}

func testCryptFileEmptyWriteChecksum236(t *testing.T, call func([]byte) uint64) {
	for _, mode := range []cryptfile.Mode{cryptfile.WriteOnly, cryptfile.Append} {
		for _, xor := range []bool{false, true} {
			for _, null := range []bool{false, true} {
				for _, chunks := range [][][]byte{nil, {{0}}, {{1, 2, 3, 4, 5, 6, 7}, {8, 9, 10, 11, 12, 13, 14, 15, 16}}} {
					for repeats := 1; repeats <= 4; repeats++ {
						t.Run(fmt.Sprintf("mode=%d/xor=%t/null=%t/chunks=%d/zeros=%d", mode, xor, null, len(chunks), repeats), func(t *testing.T) {
							path := filepath.Join(t.TempDir(), "owned-checksum.bin")
							if err := os.WriteFile(path, cryptFileWire236(t, make([]byte, 8), crypt.MapKey, false), 0o600); err != nil {
								t.Fatal(err)
							}
							cf := cryptFileGlobal236(t, path, mode, crypt.MapKey)
							if mode == cryptfile.WriteOnly {
								// Reserve the checksum block outside CryptFile's CRC stream.
								if n, err := cf.File.Write(make([]byte, 8)); err != nil || n != 8 {
									t.Fatalf("reserve checksum = %d, %v", n, err)
								}
							}
							cf.SetXOR(xor)
							want := uint32(0xffffffff)
							for _, chunk := range chunks {
								if got := call(chunk); got != 1 {
									t.Fatalf("complete seed item = %d, want 1", got)
								}
								want = cryptFileReferenceCRC236(want, chunk)
							}
							position := cf.File.Written()
							zero := make([]byte, 0, 16)
							if null {
								zero = nil
							}
							for i := 0; i < repeats; i++ {
								if got := call(zero); got != 0 {
									t.Errorf("empty write item = %d, want 0", got)
								}
								want = ^want
								if cf.File.Written() != position {
									t.Fatal("empty write moved logical offset")
								}
							}
							if _, err := cf.File.FileFlush(); err != nil {
								t.Fatal(err)
							}
							if err := cf.WriteChecksumAt(0); err != nil {
								t.Fatal(err)
							}
							if err := cf.Close(); err != nil {
								t.Fatal(err)
							}
							wire, err := os.ReadFile(path)
							if err != nil || len(wire) < 8 {
								t.Fatalf("checksum file = %x, %v", wire, err)
							}
							c, err := crypt.NewCipher(crypt.MapKey)
							if err != nil {
								t.Fatal(err)
							}
							var block [8]byte
							c.Decrypt(block[:], wire[:8])
							if got := binary.LittleEndian.Uint32(block[:4]); got != want || binary.LittleEndian.Uint32(block[4:]) != 0 {
								t.Errorf("persisted CRC block = %x; want %08x/00000000", block, want)
							}
						})
					}
				}
			}
		}
	}
}

func TestCryptFileCExportClosedIO236(t *testing.T) {
	testCryptFileClosedIO236(t, cryptFileExportCall236)
}

func testCryptFileClosedIO236(t *testing.T, call func([]byte) uint64) {
	for _, mode := range []cryptfile.Mode{cryptfile.ReadOnly, cryptfile.WriteOnly} {
		for _, key := range []int{-1, crypt.MapKey} {
			for _, xor := range []bool{false, true} {
				t.Run(fmt.Sprintf("mode=%d/key=%d/xor=%t", mode, key, xor), func(t *testing.T) {
					path := filepath.Join(t.TempDir(), "owned-closed.bin")
					if err := os.WriteFile(path, make([]byte, 16), 0o600); err != nil {
						t.Fatal(err)
					}
					cf, err := cryptfile.OpenFile(path, mode, key)
					if err != nil {
						t.Fatal(err)
					}
					previous := cryptfile.Global()
					cryptfile.SetGlobal(cf)
					t.Cleanup(func() {
						cryptfile.SetGlobal(previous)
						if err := cf.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
							t.Error(err)
						}
					})
					cf.SetXOR(xor)
					if err := cf.File.File.File.Close(); err != nil {
						t.Fatal(err)
					}
					buf := bytes.Repeat([]byte{0xa5}, 16)
					want := byte(0xa5)
					if mode == cryptfile.ReadOnly && xor {
						want ^= 126
					}
					if got := call(buf); got != 0 || !bytes.Equal(buf, bytes.Repeat([]byte{want}, 16)) {
						t.Fatalf("closed I/O = %d, %x; want 0, %02x repeated", got, buf, want)
					}
				})
			}
		}
	}
}

func cryptFileReferenceCRC236(crc uint32, p []byte) uint32 {
	for _, b := range p {
		crc ^= uint32(b)
		for bit := 0; bit < 8; bit++ {
			if crc&1 != 0 {
				crc = crc>>1 ^ 0xedb88320
			} else {
				crc >>= 1
			}
		}
	}
	return ^crc
}
