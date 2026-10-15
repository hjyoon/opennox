//go:build amd64 || arm64

package legacy

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unsafe"

	crypt "github.com/opennox/noxcrypt"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

var stackPatterns238 = [3][4]uint32{
	{},
	{0x3f800000, 0xc0200000, 0x41200000, 0x80000001},
	{0x7fc12345, 0x80000000, 0xff812345, 1},
}

func setupStackMetadata238(t *testing.T) {
	t.Helper()
	// Unit tests do not run the game's blobdata.Init startup. Load only
	// its integer argument metadata, then restore the exact prior bytes.
	data, err := os.ReadFile(filepath.Join("..", "common", "memmap", "nox", "blobdata", "blob_587000.dat"))
	if err != nil {
		t.Fatal(err)
	}
	const offset, size = 255600, 72 * 16
	if len(data) < offset+size {
		t.Fatal("shipped argument metadata is truncated")
	}
	table := data[offset : offset+size]
	if fmt.Sprintf("%x", sha256.Sum256(table)) != "ef279ed82cc7e33ded38130998e24b6a3df1e75c2066bcef3c446346a6cc17fd" {
		t.Fatal("shipped argument metadata differs from the original PE table")
	}
	destination := unsafe.Slice((*byte)(memmap.PtrOff(0x587000, offset)), size)
	previous := append([]byte(nil), destination...)
	t.Cleanup(func() { copy(destination, previous) })
	copy(destination, table)
}

func stackKinds238(action uint32) []uint32 {
	n := int(memmap.Uint8(0x587000, uintptr(255604+16*action)))
	out := make([]uint32, n)
	for i := range out {
		out[i] = memmap.Uint32(0x587000, uintptr(255608+4*(i+4*int(action))))
	}
	return out
}

func stackChunks238(action uint32, pattern int, refs bool) ([][]byte, int, int) {
	name := ai.ActionType(action).String()
	if action == 39 {
		// The original stack name table has an unnamed slot at 39. Keep
		// this wire representation separate from String's diagnostic text.
		name = ""
	}
	kinds := stackKinds238(action)
	out := append(xferStringChunks237(name, true), []byte{byte(len(kinds))})
	vectorIndex, vectorOffset := -1, -1
	for i, kind := range kinds {
		value := stackPatterns238[pattern][2*i]
		size := 4
		switch kind {
		case 0:
			vectorIndex = i
			vectorOffset = len(bytes.Join(out, nil))
			size = 8
		case 1:
			value = 0
			if refs {
				value = 0x11223344
			}
		case 2:
			value = 0
			if refs {
				value = 0x55667788
			}
		case 7:
			size = 1
		}
		var buf [8]byte
		binary.LittleEndian.PutUint32(buf[:4], value)
		binary.LittleEndian.PutUint32(buf[4:], stackPatterns238[pattern][2*i+1])
		out = append(out, append([]byte(nil), buf[:size]...))
	}
	out = append(out, []byte{0xba, 0xdc, 0xcd, 0xab})
	return out, vectorIndex, vectorOffset
}

func stackWriteItem238(action uint32, pattern int, refs bool, obj *server.Object, wp *server.Waypoint) server.AIStackItem {
	item := server.AIStackItem{Action: action, Field5: 0xabcddcba}
	for i, value := range stackPatterns238[pattern] {
		item.Args[i] = uintptr(value) | uintptr(0xaabbccdd)<<32
	}
	for i, kind := range stackKinds238(action) {
		switch kind {
		case 1:
			item.Args[2*i] = 0
			if refs {
				item.Args[2*i] = uintptr(unsafe.Pointer(obj))
			}
		case 2:
			item.Args[2*i] = 0
			if refs {
				item.Args[2*i] = uintptr(unsafe.Pointer(wp))
			}
		}
	}
	return item
}

func stackExpectedRead238(action uint32, pattern int, refs bool, beforeArgument int) (server.AIStackItem, [2]int32, [2]uint32) {
	item := server.AIStackItem{Action: action, Field5: 0xabcddcba}
	var objects [2]int32
	var waypoints [2]uint32
	for i, kind := range stackKinds238(action) {
		if beforeArgument >= 0 && i >= beforeArgument {
			break
		}
		value := stackPatterns238[pattern][2*i]
		switch kind {
		case 0:
			item.Args[2*i+1] = uintptr(stackPatterns238[pattern][2*i+1])
		case 1:
			value = 0
			if refs {
				objects[i] = 0x11223344
			}
		case 2:
			value = 0
			if refs {
				waypoints[i] = 0x55667788
			}
		case 5:
			shifted := int32(value + 7)
			value = 1
			if shifted > 0 {
				value = uint32(shifted)
			}
		case 7:
			value &= 0xff
		}
		item.Args[2*i] = uintptr(value)
	}
	return item, objects, waypoints
}

// Observe a read-mode CRC through the public checksum-writing API using a
// separate owned backend. Restore the reader immediately; neither its mode,
// checksum state nor file position is changed to make the assertion pass.
func readStackCRC238(t *testing.T, cf *cryptfile.CryptFile) uint32 {
	t.Helper()
	path := filepath.Join(t.TempDir(), "owned-read-checksum.bin")
	out, err := cryptfile.OpenFile(path, cryptfile.WriteOnly, crypt.MapKey)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = out.Close() })
	if n, err := out.File.Write(make([]byte, crypt.Block)); err != nil || n != crypt.Block {
		t.Fatalf("reserve read checksum = %d, %v", n, err)
	}
	reader := cf.File
	cf.File = out.File
	err = cf.WriteChecksumAt(0)
	cf.File = reader
	if err != nil {
		t.Fatal(err)
	}
	if err := out.Close(); err != nil {
		t.Fatal(err)
	}
	wire, err := os.ReadFile(path)
	if err != nil || len(wire) < crypt.Block {
		t.Fatalf("read checksum file = %x, %v", wire, err)
	}
	cipher, err := crypt.NewCipher(crypt.MapKey)
	if err != nil {
		t.Fatal(err)
	}
	var plain [crypt.Block]byte
	cipher.Decrypt(plain[:], wire[:crypt.Block])
	if binary.LittleEndian.Uint32(plain[4:]) != 0 {
		t.Fatalf("read checksum padding = %x", plain)
	}
	return binary.LittleEndian.Uint32(plain[:4])
}

func logStackRead238(t *testing.T, payload []byte, crc uint32) {
	t.Helper()
	row, err := json.Marshal(struct {
		Test    string `json:"test"`
		Payload string `json:"payload"`
		CRC     uint32 `json:"crc"`
	}{t.Name(), hex.EncodeToString(payload), crc})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("R238 actual %s", row)
}

func TestMonsterAIStackItems238(t *testing.T) {
	setupStackMetadata238(t)
	obj, freeObject := alloc.New(server.Object{ScriptIDVal: 0x11223344})
	wp, freeWaypoint := alloc.New(server.Waypoint{Index: 0x55667788})
	// alloc.New's value specifies only the allocation's type and size.
	obj.ScriptIDVal = 0x11223344
	wp.Index = 0x55667788
	t.Cleanup(freeObject)
	t.Cleanup(freeWaypoint)
	if uintptr(unsafe.Pointer(obj)) <= 0xffffffff || uintptr(unsafe.Pointer(wp)) <= 0xffffffff {
		t.Fatal("references must exercise native-width pointers")
	}
	for action := uint32(0); action < 72; action++ {
		for pattern := range stackPatterns238 {
			for _, refs := range []bool{false, true} {
				for _, xor := range []bool{false, true} {
					for _, mode := range []string{"read", "write"} {
						t.Run(fmt.Sprintf("mode=%s/action=%d/pattern=%d/refs=%t/xor=%t", mode, action, pattern, refs, xor), func(t *testing.T) {
							chunks, _, _ := stackChunks238(action, pattern, refs)
							if mode == "write" {
								item := stackWriteItem238(action, pattern, refs, obj, wp)
								// The following byte exercises a real next-field CRC
								// boundary; the original audit applies that same call.
								chunks = append(chunks, []byte{0x5a})
								checkXferChecksum237(t, xor, chunks, func(cf *cryptfile.CryptFile) error {
									var objects [2]int32
									var waypoints [2]uint32
									if err := monsterXferAIStackItem52A440(cf, &item, &objects, &waypoints, 7); err != nil {
										return err
									}
									for i, kind := range stackKinds238(action) {
										if (kind == 1 || kind == 2) && item.Args[2*i] != stackWriteItem238(action, pattern, refs, obj, wp).Args[2*i] {
											t.Fatal("write truncated a native reference")
										}
									}
									return cf.WriteU8(0x5a)
								})
								return
							}
							plain := bytes.Join(chunks, nil)
							path := filepath.Join(t.TempDir(), "owned-stack-read.bin")
							if err := os.WriteFile(path, cryptFileWire236(t, append(append([]byte(nil), plain...), 0x5a), crypt.MapKey, xor), 0o600); err != nil {
								t.Fatal(err)
							}
							cf, err := cryptfile.OpenFile(path, cryptfile.ReadOnly, crypt.MapKey)
							if err != nil {
								t.Fatal(err)
							}
							t.Cleanup(func() { _ = cf.Close() })
							cf.SetXOR(xor)
							item := server.AIStackItem{Args: [4]uintptr{0xfeed, 0xbeef, 0xface, 0xcafe}}
							var objects [2]int32
							var waypoints [2]uint32
							if err := monsterXferAIStackItem52A440(cf, &item, &objects, &waypoints, 7); err != nil {
								t.Fatal(err)
							}
							wantItem, wantObjects, wantWaypoints := stackExpectedRead238(action, pattern, refs, -1)
							if item != wantItem || objects != wantObjects || waypoints != wantWaypoints {
								t.Errorf("decoded = %v/%v/%v; want %v/%v/%v", item, objects, waypoints, wantItem, wantObjects, wantWaypoints)
							}
							gotCRC := readStackCRC238(t, cf)
							wantCRC := uint32(0xffffffff)
							for _, chunk := range chunks {
								if len(chunk) != 0 { // an empty fread completes no item
									wantCRC = xferCRC237(wantCRC, chunk)
								}
							}
							if gotCRC != wantCRC {
								t.Errorf("read CRC = %08x; want %08x", gotCRC, wantCRC)
							}
							logStackRead238(t, plain, gotCRC)
							if got, err := cf.ReadU8(); err != nil || got != 0x5a {
								t.Fatalf("next field = %x, %v", got, err)
							}
							if err := cf.Close(); err != nil {
								t.Fatal(err)
							}
						})
					}
				}
			}
		}
	}
}

func TestMonsterAIStackPartialVector238(t *testing.T) {
	setupStackMetadata238(t)
	for action := uint32(0); action < 72; action++ {
		for pattern := range stackPatterns238 {
			chunks, vectorIndex, vectorOffset := stackChunks238(action, pattern, false)
			if vectorIndex < 0 {
				continue
			}
			for available := 0; available < 8; available++ {
				for _, xor := range []bool{false, true} {
					t.Run(fmt.Sprintf("action=%d/pattern=%d/available=%d/xor=%t", action, pattern, available, xor), func(t *testing.T) {
						plain := bytes.Join(chunks, nil)[:vectorOffset+available]
						path := filepath.Join(t.TempDir(), "owned-truncated-stack.bin")
						if err := os.WriteFile(path, cryptFileWire236(t, plain, -1, xor), 0o600); err != nil {
							t.Fatal(err)
						}
						cf, err := cryptfile.OpenFile(path, cryptfile.ReadOnly, -1)
						if err != nil {
							t.Fatal(err)
						}
						t.Cleanup(func() { _ = cf.Close() })
						cf.SetXOR(xor)
						item := server.AIStackItem{Field5: 0xfeedface, Args: [4]uintptr{0xfeed, 0xbeef, 0xface, 0xcafe}}
						var objects [2]int32
						var waypoints [2]uint32
						err = monsterXferAIStackItem52A440(cf, &item, &objects, &waypoints, 7)
						word := 0
						if available >= 4 {
							word = 1
						}
						cause := io.ErrUnexpectedEOF
						if available == 0 || available == 4 {
							cause = io.EOF
						}
						if !errors.Is(err, cause) || !strings.Contains(fmt.Sprint(err), fmt.Sprintf("argument %d.%d:", vectorIndex, word)) {
							t.Errorf("partial error = %v; want argument %d.%d wrapping %v", err, vectorIndex, word, cause)
						}
						wantItem, wantObjects, wantWaypoints := stackExpectedRead238(action, pattern, false, vectorIndex)
						wantItem.Field5 = 0xfeedface
						if available >= 4 {
							wantItem.Args[2*vectorIndex] = uintptr(stackPatterns238[pattern][2*vectorIndex])
						}
						if item != wantItem || objects != wantObjects || waypoints != wantWaypoints {
							t.Errorf("partial state = %v/%v/%v; want %v/%v/%v", item, objects, waypoints, wantItem, wantObjects, wantWaypoints)
						}
						if offset, err := cf.File.Seek(0, io.SeekCurrent); err != nil || offset != int64(len(plain)) {
							t.Errorf("partial file offset = %d, %v; want %d", offset, err, len(plain))
						}
						wantCRC, consumed := uint32(0xffffffff), 0
						for _, chunk := range chunks {
							if consumed+len(chunk) > vectorOffset {
								break
							}
							wantCRC = xferCRC237(wantCRC, chunk)
							consumed += len(chunk)
						}
						gotCRC := readStackCRC238(t, cf)
						if gotCRC != wantCRC {
							t.Errorf("partial read CRC = %08x; want %08x", gotCRC, wantCRC)
						}
						logStackRead238(t, plain, gotCRC)
					})
				}
			}
		}
	}
}
