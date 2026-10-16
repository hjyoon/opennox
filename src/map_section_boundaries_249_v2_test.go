package opennox

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	crypt "github.com/opennox/noxcrypt"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

// Stock section names and callback order are also read from the unchanged
// GAME.EXE table by the independent PE32 execution probe.
var mapSectionNames249 = []string{
	"MapInfo", "WallMap", "FloorMap", "SecretWalls", "DestructableWalls",
	"WayPoints", "DebugData", "WindowWalls", "GroupData", "ScriptObject",
	"AmbientData", "Polygons", "MapIntro", "ScriptData", "ObjectTOC", "ObjectData",
}

func TestMapSectionWriterNativeBoundaries249V2(t *testing.T) {
	context, free := alloc.New([4]uint64{})
	defer free()
	*context = [4]uint64{0x1122334455667788, 0x99aabbccddeeff00, 0x0123456789abcdef, 0xfedcba9876543210}
	before := *context
	for _, mode := range []cryptfile.Mode{cryptfile.WriteOnly, cryptfile.Append} {
		for _, xor := range []bool{false, true} {
			for _, seed := range []bool{false, true} {
				for _, payload := range []string{"none", "empty", "typed"} {
					for _, fail := range []int{-1, 0, 7, 15} {
						for _, cEntry := range []bool{false, true} {
							for _, null := range []bool{true, false} {
								name := fmt.Sprintf("mode%d/xor%t/seed%t/payload-%s/fail%d/C%t/null%t", mode, xor, seed, payload, fail, cEntry, null)
								t.Run(name, func(t *testing.T) {
									arg := unsafe.Pointer(context)
									if null {
										arg = nil
									}
									testMapSectionWriter249(t, mode, xor, seed, payload, fail, cEntry, null, arg)
								})
							}
						}
					}
				}
			}
		}
	}
	if *context != before {
		t.Fatal("native callback context/canary changed")
	}
}

func mapSectionChunks249(index int, payload string) [][]byte {
	switch payload {
	case "none":
		return nil
	case "empty":
		return [][]byte{nil}
	case "typed":
		word, dword, vector := make([]byte, 2), make([]byte, 4), make([]byte, 8)
		binary.LittleEndian.PutUint16(word, uint16(index+1))
		binary.LittleEndian.PutUint32(dword, 0x80000000|uint32(index))
		binary.LittleEndian.PutUint32(vector, 0x7f801234+uint32(index))
		binary.LittleEndian.PutUint32(vector[4:], 0xff001234+uint32(index))
		return [][]byte{word, dword, vector, {byte(index), 0x7e, 0xff}}
	default:
		panic("unknown fixture payload")
	}
}

// CRC is inverted after every semantic I/O item, not after each cipher block.
func mapSectionReferenceCRC249(crc uint32, data []byte) uint32 {
	for _, b := range data {
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

func testMapSectionWriter249(t *testing.T, mode cryptfile.Mode, xor, seed bool, payload string, fail int, cEntry, null bool, arg unsafe.Pointer) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "sections.bin")
	cipher, err := crypt.NewCipher(crypt.MapKey)
	if err != nil {
		t.Fatal(err)
	}
	if mode == cryptfile.Append {
		var initial, encrypted [8]byte
		cipher.Encrypt(encrypted[:], initial[:])
		if err := os.WriteFile(path, encrypted[:], 0o600); err != nil {
			t.Fatal(err)
		}
	}
	cf, err := cryptfile.OpenFile(path, mode, crypt.MapKey)
	if err != nil {
		t.Fatal(err)
	}
	defer cf.Close()
	if mode == cryptfile.WriteOnly {
		// The reserved checksum block is not part of the semantic CRC stream.
		if n, err := cf.File.Write(make([]byte, 8)); err != nil || n != 8 {
			t.Fatalf("reserve checksum = %d, %v", n, err)
		}
	}
	cf.SetXOR(xor)
	oldFile, oldSections := cryptfile.Global(), noxMapSections
	cryptfile.SetGlobal(cf)
	defer cryptfile.SetGlobal(oldFile)
	defer func() { noxMapSections = oldSections }()
	wantCRC := uint32(0xffffffff)
	seedData := []byte{1, 2, 3, 4, 5, 6, 7}
	if seed {
		if n, err := cf.Write(seedData); err != nil || n != len(seedData) {
			t.Fatalf("seed write = %d, %v", n, err)
		}
		wantCRC = mapSectionReferenceCRC249(wantCRC, seedData)
	}
	var callbacks []string
	fault := errors.New("controlled section failure")
	noxMapSections = make([]mapSection, len(mapSectionNames249))
	for index, name := range mapSectionNames249 {
		noxMapSections[index] = mapSection{Name: name, Fnc: func(actual *cryptfile.CryptFile, context unsafe.Pointer) error {
			if actual != cf || context != arg {
				t.Errorf("native file/context identity changed: file=%p/%p context=%p/%p", actual, cf, context, arg)
			}
			callbacks = append(callbacks, name)
			for _, chunk := range mapSectionChunks249(index, payload) {
				if n, err := actual.Write(chunk); err != nil || n != len(chunk) {
					return fmt.Errorf("controlled payload write = %d, %v", n, err)
				}
			}
			if index == fail {
				return fault
			}
			return nil
		}}
	}
	var result int
	if cEntry {
		result = ccall.CallIntPtr(legacy.Get_nox_xxx_mapWriteSectionsMB_426E20(), arg)
	} else {
		result = nox_xxx_mapWriteSectionsMB_426E20(arg)
	}
	position := cf.File.Written()
	if xor {
		position, err = cf.File.File.Seek(0, io.SeekCurrent)
		if err != nil {
			t.Fatal(err)
		}
	}
	count, wantResult := len(mapSectionNames249), 1
	if fail >= 0 {
		count, wantResult = fail+1, 0
	}
	if result != wantResult || len(callbacks) != count {
		t.Errorf("section result/callback count = %d/%d, want %d/%d", result, len(callbacks), wantResult, count)
	}
	for index, name := range mapSectionNames249[:count] {
		if callbacks[index] != name {
			t.Errorf("callback %d = %q, want %q", index, callbacks[index], name)
		}
		wantCRC = mapSectionReferenceCRC249(wantCRC, []byte{byte(len(name) + 1)})
		wantCRC = mapSectionReferenceCRC249(wantCRC, append([]byte(name), 0))
		for _, chunk := range mapSectionChunks249(index, payload) {
			wantCRC = mapSectionReferenceCRC249(wantCRC, chunk)
		}
	}
	if _, err = cf.File.FileFlush(); err != nil {
		t.Fatal(err)
	}
	if err = cf.WriteChecksumAt(0); err != nil {
		t.Fatal(err)
	}
	if err = cf.Close(); err != nil {
		t.Fatal(err)
	}
	wire, err := os.ReadFile(path)
	if err != nil || len(wire) < 8 {
		t.Fatalf("native output = %x, %v", wire, err)
	}
	var header [8]byte
	cipher.Decrypt(header[:], wire[:8])
	crc := binary.LittleEndian.Uint32(header[:4])
	if crc != wantCRC || binary.LittleEndian.Uint32(header[4:]) != 0 {
		t.Errorf("persisted semantic CRC = %x; want %08x/00000000", header, wantCRC)
	}
	row, err := json.Marshal(map[string]any{"mode": mode, "xor": xor, "seed": seed, "payload": payload, "fail": fail, "C_entry": cEntry, "null_context": null, "result": result, "crc": crc, "position": position, "callbacks": callbacks, "pointer_bytes": unsafe.Sizeof(uintptr(0))})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("R249 SECTION: %s", row)
	plain := append([]byte(nil), wire...)
	if xor {
		for i := 8; i < len(plain); i++ {
			plain[i] ^= 126
		}
	} else {
		for off := 0; off+8 <= len(wire); off += 8 {
			cipher.Decrypt(plain[off:off+8], wire[off:off+8])
		}
	}
	checkMapSectionWire249(t, plain, wire, position, seed, payload, fail, xor, count)
}

func checkMapSectionWire249(t *testing.T, plain, wire []byte, position int64, seed bool, payload string, fail int, xor bool, count int) {
	t.Helper()
	off := 8
	if seed {
		if len(plain) < off+7 || !bytes.Equal(plain[off:off+7], []byte{1, 2, 3, 4, 5, 6, 7}) {
			t.Fatal("seed bytes changed")
		}
		off += 7
	}
	for index, name := range mapSectionNames249[:count] {
		wantName := append([]byte{byte(len(name) + 1)}, append([]byte(name), 0)...)
		if off+len(wantName) > len(plain) || !bytes.Equal(plain[off:off+len(wantName)], wantName) {
			t.Fatalf("section %d name framing changed at %d", index, off)
		}
		off += len(wantName)
		headerBytes := 4
		if !xor {
			off = (off + 7) &^ 7
			headerBytes = 8
		}
		if off+headerBytes > len(plain) {
			t.Fatalf("section %d missing length header", index)
		}
		var data []byte
		for _, chunk := range mapSectionChunks249(index, payload) {
			data = append(data, chunk...)
		}
		wantSize := uint32(len(data))
		if index == fail {
			wantSize = 0
			if xor {
				// The raw placeholder initially stores its own file position.
				wantSize = uint32(off)
			}
		}
		if !xor && index == fail {
			// GAME.EXE 00409156..00409177 reserves eight raw zero bytes.
			// A failed callback never reaches the encrypted length backpatch.
			if !bytes.Equal(wire[off:off+headerBytes], make([]byte, headerBytes)) {
				t.Errorf("section %d unclosed reservation is not raw zero bytes", index)
			}
		} else {
			if got := binary.LittleEndian.Uint32(plain[off:]); got != wantSize {
				t.Errorf("section %d length/placeholder = %d, want %d", index, got, wantSize)
			}
			if !xor && binary.LittleEndian.Uint32(plain[off+4:]) != 0 {
				t.Errorf("section %d length high DWORD changed", index)
			}
		}
		off += headerBytes
		if off+len(data) > len(plain) || !bytes.Equal(plain[off:off+len(data)], data) {
			t.Fatalf("section %d payload changed: want %s", index, hex.EncodeToString(data))
		}
		off += len(data)
	}
	if int64(off) != position {
		t.Errorf("section end position = %d, want %d; seek patch appended data", position, off)
	}
}
