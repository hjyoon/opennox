package cryptfile

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"testing"

	crypt "github.com/opennox/noxcrypt"
)

var readInitials235 = []uint32{0xffffffff, 0, 1, 0x12345678, 0x80000000, 0xdeadbeef}

// The original 00426AC0 checks fread's completed-item count at 00426BA5
// before calling 00426BD0. A partial item changes the buffer, but not CRC.
// XOR at 0056FDD0 still covers the entire requested buffer on failure.
func TestCryptFileReadChecksumItemBoundary235(t *testing.T) {
	states := sha256.New()
	for _, xor := range []bool{false, true} {
		for _, rw := range []bool{false, true} {
			for _, initial := range readInitials235 {
				for _, requested := range []int{0, 1, 7, 8, 9, 16, 31, 255, 256} {
					for _, delivered := range readPrefixes235(requested) {
						t.Run(fmt.Sprintf("xor=%t/rw=%t/crc=%08x/request=%d/prefix=%d", xor, rw, initial, requested, delivered), func(t *testing.T) {
							payload := readPayload235(delivered)
							cf := openReadFixture235(t, payload, -1)
							cf.SetXOR(xor)
							cf.crcSum = initial
							backing := append(bytes.Repeat([]byte{0x3c}, 16), bytes.Repeat([]byte{0xa5}, requested)...)
							backing = append(backing, bytes.Repeat([]byte{0xc3}, 16)...)
							buf := backing[16 : 16+requested]
							n, err := readCall235(cf, buf, rw)
							var wantErr error
							if requested > 0 && delivered == 0 {
								wantErr = io.EOF
							}
							if n != delivered || !errors.Is(err, wantErr) {
								t.Fatalf("Read = %d, %v; want %d, %v", n, err, delivered, wantErr)
							}
							want := append(append([]byte(nil), payload...), bytes.Repeat([]byte{0xa5}, requested-delivered)...)
							if xor {
								for i := range want {
									want[i] ^= 126
								}
							}
							if !bytes.Equal(buf, want) || !bytes.Equal(backing[:16], bytes.Repeat([]byte{0x3c}, 16)) || !bytes.Equal(backing[16+requested:], bytes.Repeat([]byte{0xc3}, 16)) {
								t.Fatalf("buffer/canaries = %x; want payload %x", backing, want)
							}
							wantCRC := initial
							if requested > 0 && delivered == requested {
								wantCRC = readReferenceCRC235(initial, want)
							}
							var state [4]byte
							binary.LittleEndian.PutUint32(state[:], cf.crcSum)
							states.Write(state[:])
							if cf.crcSum != wantCRC {
								t.Errorf("CRC = %08x, want %08x (completed item=%t)", cf.crcSum, wantCRC, requested > 0 && delivered == requested)
							}
							tail := bytes.Repeat([]byte{0xa5}, 3)
							n, err = readCall235(cf, tail, rw)
							wantTail := byte(0xa5)
							if xor {
								wantTail ^= 126
							}
							if n != 0 || !errors.Is(err, io.EOF) || !bytes.Equal(tail, bytes.Repeat([]byte{wantTail}, 3)) || cf.crcSum != wantCRC {
								t.Errorf("EOF = %d, %v, %x, CRC %08x; want 0, EOF, %02x repeated, CRC %08x", n, err, tail, cf.crcSum, wantTail, wantCRC)
							}
						})
					}
				}
			}
		}
	}
	// Independent execution of unchanged GAME.EXE wrapper/XOR/CRC bodies:
	// 912 controlled fread boundaries, 192 CRC calls, 456 XOR calls.
	const originalStates = "f4fbfee05b3bc636c5e419cb6f3aa7429ae2a5e66f33cb4c3d737704dff13e6a"
	if got := fmt.Sprintf("%x", states.Sum(nil)); got != originalStates {
		t.Errorf("read CRC states SHA256 = %s, original = %s", got, originalStates)
	}
}

func TestCryptFileReadTruncatedCipherChecksum235(t *testing.T) {
	plain := readPayload235(16)
	ciphertext := readEncrypt235(t, plain)
	for _, rw := range []bool{false, true} {
		for _, initial := range readInitials235 {
			for cut := 0; cut <= len(ciphertext); cut++ {
				t.Run(fmt.Sprintf("rw=%t/crc=%08x/file=%d", rw, initial, cut), func(t *testing.T) {
					cf := openReadFixture235(t, ciphertext[:cut], crypt.MapKey)
					cf.crcSum = initial
					buf := bytes.Repeat([]byte{0xa5}, 24)
					n, err := readCall235(cf, buf, rw)
					wantN := cut / crypt.Block * crypt.Block
					wantErr := io.EOF
					if cut%crypt.Block != 0 {
						wantErr = io.ErrUnexpectedEOF
					}
					want := append(append([]byte(nil), plain[:wantN]...), bytes.Repeat([]byte{0xa5}, len(buf)-wantN)...)
					if n != wantN || !errors.Is(err, wantErr) || !bytes.Equal(buf, want) || cf.crcSum != initial {
						t.Fatalf("truncated read = %d, %v, %x, CRC %08x; want %d, %v, %x, CRC %08x", n, err, buf, cf.crcSum, wantN, wantErr, want, initial)
					}
				})
			}
		}
	}
}

func TestCryptFileReadCompleteCipherChecksum235(t *testing.T) {
	plain := readPayload235(32)
	ciphertext := readEncrypt235(t, plain)
	for _, rw := range []bool{false, true} {
		for _, initial := range readInitials235 {
			for _, requested := range []int{0, 1, 7, 8, 9, 16, 31, 32} {
				t.Run(fmt.Sprintf("rw=%t/crc=%08x/request=%d", rw, initial, requested), func(t *testing.T) {
					cf := openReadFixture235(t, ciphertext, crypt.MapKey)
					cf.crcSum = initial
					wantCRC := initial
					for _, span := range [][2]int{{0, requested}, {requested, len(plain)}} {
						buf := make([]byte, span[1]-span[0])
						n, err := readCall235(cf, buf, rw)
						if len(buf) > 0 {
							wantCRC = readReferenceCRC235(wantCRC, plain[span[0]:span[1]])
						}
						if n != len(buf) || err != nil || !bytes.Equal(buf, plain[span[0]:span[1]]) || cf.crcSum != wantCRC {
							t.Fatalf("complete read = %d, %v, %x, CRC %08x; want %x, CRC %08x", n, err, buf, cf.crcSum, plain[span[0]:span[1]], wantCRC)
						}
					}
					n, err := readCall235(cf, make([]byte, 1), rw)
					if n != 0 || !errors.Is(err, io.EOF) || cf.crcSum != wantCRC {
						t.Fatalf("EOF = %d, %v, CRC %08x; want unchanged %08x", n, err, cf.crcSum, wantCRC)
					}
				})
			}
		}
	}
}

func readCall235(cf *CryptFile, p []byte, rw bool) (int, error) {
	if rw {
		return cf.ReadWrite(p)
	}
	return cf.Read(p)
}

func openReadFixture235(t *testing.T, data []byte, key int) *CryptFile {
	t.Helper()
	path := filepath.Join(t.TempDir(), "owned-read.bin")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	cf, err := OpenFile(path, ReadOnly, key)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := cf.Close(); err != nil {
			t.Error(err)
		}
	})
	return cf
}

func readPayload235(n int) []byte {
	data := make([]byte, n)
	for i := range data {
		data[i] = byte(i*37 + 11)
	}
	return data
}

func readPrefixes235(n int) []int {
	values := []int{0, min(1, n), n / 2, max(0, n-1), n}
	sort.Ints(values)
	unique := values[:0]
	for _, v := range values {
		if len(unique) == 0 || unique[len(unique)-1] != v {
			unique = append(unique, v)
		}
	}
	return unique
}

func readReferenceCRC235(crc uint32, p []byte) uint32 {
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

func readEncrypt235(t *testing.T, data []byte) []byte {
	t.Helper()
	c, err := crypt.NewCipher(crypt.MapKey)
	if err != nil {
		t.Fatal(err)
	}
	if len(data)%crypt.Block != 0 {
		t.Fatal("fixture must contain complete cipher blocks")
	}
	out := make([]byte, len(data))
	for off := 0; off < len(data); off += crypt.Block {
		c.Encrypt(out[off:off+crypt.Block], data[off:off+crypt.Block])
	}
	return out
}
