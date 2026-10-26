//go:build amd64 || arm64

package legacy

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/datapath"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"
	crypt "github.com/opennox/noxcrypt"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// The original CRC complements at each call boundary, including an empty
// write. Use a bitwise reference, not the production table or crc32.Checksum.
func xferCRC237(crc uint32, p []byte) uint32 {
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

func xferStringChunks237(value string, present bool) [][]byte {
	out := [][]byte{{byte(len(value))}}
	if present {
		out = append(out, []byte(value))
	}
	return out
}

func checkXferChecksum237(t *testing.T, xor bool, chunks [][]byte, write func(*cryptfile.CryptFile) error) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "owned-xfer.bin")
	cf, err := cryptfile.OpenFile(path, cryptfile.WriteOnly, crypt.MapKey)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cf.Close() })
	// The checksum block is outside the serialized object's CRC stream.
	if n, err := cf.File.Write(make([]byte, crypt.Block)); err != nil || n != crypt.Block {
		t.Fatalf("reserve checksum = %d, %v", n, err)
	}
	cf.SetXOR(xor)
	if err := write(cf); err != nil {
		t.Fatal(err)
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
	if err != nil || len(wire) < crypt.Block {
		t.Fatalf("persisted file = %x, %v", wire, err)
	}
	cipher, err := crypt.NewCipher(crypt.MapKey)
	if err != nil {
		t.Fatal(err)
	}
	var header [crypt.Block]byte
	cipher.Decrypt(header[:], wire[:crypt.Block])
	gotCRC := binary.LittleEndian.Uint32(header[:4])
	wantCRC := uint32(0xffffffff)
	var want []byte
	for _, chunk := range chunks {
		want = append(want, chunk...)
		wantCRC = xferCRC237(wantCRC, chunk)
	}
	body := append([]byte(nil), wire[crypt.Block:]...)
	if xor {
		for i := range body {
			body[i] ^= 126
		}
	} else {
		if len(body)%crypt.Block != 0 {
			t.Fatalf("unaligned ciphertext: %d", len(body))
		}
		for off := 0; off < len(body); off += crypt.Block {
			cipher.Decrypt(body[off:off+crypt.Block], body[off:off+crypt.Block])
		}
	}
	if len(body) < len(want) || !bytes.Equal(body[:len(want)], want) {
		t.Errorf("serialized payload = %x; want prefix %x", body, want)
	}
	if gotCRC != wantCRC || binary.LittleEndian.Uint32(header[4:]) != 0 {
		t.Errorf("persisted checksum = %x; want %08x/00000000", header, wantCRC)
	}
	// The independent original-instruction audit joins these actual results
	// with the original voice, shop and definition records by test name.
	actual := body
	if len(actual) > len(want) {
		actual = actual[:len(want)]
	}
	record, err := json.Marshal(struct {
		Test    string `json:"test"`
		Payload string `json:"payload"`
		CRC     uint32 `json:"crc"`
	}{t.Name(), hex.EncodeToString(actual), gotCRC})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("R237 actual %s", record)
}

func TestMonsterString8Checksum237(t *testing.T) {
	for _, value := range []string{"", "A", "Apprentice", strings.Repeat("x", 255), strings.Repeat("y", 256)} {
		for _, xor := range []bool{false, true} {
			for _, seeded := range []bool{false, true} {
				t.Run(fmt.Sprintf("bytes=%d/xor=%t/seeded=%t", len(value), xor, seeded), func(t *testing.T) {
					stored := value
					if len(stored) > 255 {
						stored = stored[:255]
					}
					var chunks [][]byte
					if seeded {
						chunks = append(chunks, []byte{1, 2, 3, 4, 5, 6, 7}, []byte{8})
					}
					chunks = append(chunks, xferStringChunks237(stored, true)...)
					chunks = append(chunks, []byte{0x5a})
					checkXferChecksum237(t, xor, chunks, func(cf *cryptfile.CryptFile) error {
						if seeded {
							for _, p := range chunks[:2] {
								if _, err := cf.Write(p); err != nil {
									return err
								}
							}
						}
						got, err := monsterRWString8(cf, value)
						if got != stored || err != nil {
							return fmt.Errorf("string = %q, %v; want %q", got, err, stored)
						}
						return cf.WriteU8(0x5a)
					})
				})
			}
		}
	}
}

func TestMonsterVoiceChecksum237(t *testing.T) {
	previous := soundSetByName
	soundSetByName = make(map[string]*soundSetRecord)
	t.Cleanup(func() { soundSetByName = previous })
	for _, tc := range []struct {
		name    string
		present bool
	}{
		{"", false}, {"", true}, {"NPCMale", true},
	} {
		for _, xor := range []bool{false, true} {
			t.Run(fmt.Sprintf("present=%t/name=%s/xor=%t", tc.present, tc.name, xor), func(t *testing.T) {
				ud := new(server.MonsterUpdateData)
				if tc.present {
					p, free := alloc.Malloc(soundSetRecordSize)
					t.Cleanup(func() {
						delete(soundSetByName, tc.name)
						free()
					})
					ud.SoundSet122 = p
					soundSetByName[tc.name] = &soundSetRecord{name: tc.name, data: p}
				}
				checkXferChecksum237(t, xor, xferStringChunks237(tc.name, tc.present), func(cf *cryptfile.CryptFile) error {
					return monsterXferVoice528DB0(cf, ud)
				})
			})
		}
	}
}

func TestMonsterDefinitionChecksum237(t *testing.T) {
	names := [...]string{"SPELL_INVALID", "SPELL_ANCHOR", "SPELL_ARACHNAPHOBIA", "SPELL_BLIND"}
	for _, name := range []string{"", "Apprentice"} {
		for _, auto := range [][3]uint32{{0, 0, 0}, {0, 1, 0}, {1, 2, 3}} {
			for _, xor := range []bool{false, true} {
				t.Run(fmt.Sprintf("name=%s/auto=%d-%d-%d/xor=%t", name, auto[0], auto[1], auto[2], xor), func(t *testing.T) {
					ud := new(server.MonsterUpdateData)
					copy(unsafe.Slice((*byte)(unsafe.Pointer(&ud.Field341)), 76), name)
					ud.Field511, ud.Field512, ud.Field513 = auto[0], auto[1], auto[2]
					// GAME.EXE 00529058..005296D7, version 64. Each
					// numeric element below is one original write call.
					chunks := [][]byte{{0}}
					for range 6 {
						chunks = append(chunks, make([]byte, 4))
					}
					chunks = append(chunks, xferStringChunks237(name, true)...)
					chunks = append(chunks, make([]byte, 4)) // empty spell list
					for range 10 {
						chunks = append(chunks, make([]byte, 2))
					}
					for range 3 {
						chunks = append(chunks, make([]byte, 4))
					}
					for _, id := range auto {
						chunks = append(chunks, xferStringChunks237(names[id], true)...)
					}
					chunks = append(chunks, xferStringChunks237("ACTION_IDLE", true)...)
					checkXferChecksum237(t, xor, chunks, func(cf *cryptfile.CryptFile) error {
						return monsterXferDefinition528DB0(cf, ud, 64)
					})
				})
			}
		}
	}
}

func TestMonsterShopChecksum237(t *testing.T) {
	oldData := datapath.Data()
	dataDir := t.TempDir()
	datapath.SetData(dataDir)
	t.Cleanup(func() { datapath.SetData(oldData) })
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	for _, name := range []string{"Sword", "Wasp"} {
		if err := srv.Types.ReadObjectType(&things.Thing{Name: name}); err != nil {
			t.Fatal(err)
		}
	}
	// A synthetic, encrypted one-entry registry exercises live modifier-ID
	// resolution without loading or changing a user's modifier.bin/datapath.
	path := filepath.Join(dataDir, "modifier.bin")
	plain := []byte("MATERIAL\nMaterial1\nEND\nEND\n")
	if err := os.WriteFile(path, cryptFileWire236(t, plain, crypt.ModifierBin, false), 0o600); err != nil {
		t.Fatal(err)
	}
	base, err := filepath.Abs(datapath.Data())
	if err != nil {
		t.Fatal(err)
	}
	relative, err := filepath.Rel(base, path)
	if err != nil {
		t.Fatal(err)
	}
	if err := srv.Modif.Nox_xxx_parseModifierBin_412930(relative); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(srv.Modif.Nox_xxx_freeWeaponArmorDefAndModifs_413060)
	modifierID := srv.Modif.Nox_xxx_modifGetIdByName413290("Material1")
	slot, ok := server.EncodeShopkeeperModifierID(modifierID)
	if !ok {
		t.Fatal("fixture modifier did not enter the registry")
	}
	typ := srv.Types.ByID("Sword")
	for _, tc := range []struct {
		kind    string
		param   uint32
		name    string
		present bool
	}{
		{"guide", 0, "", false}, {"guide", uint32(srv.Types.IndByID("Wasp")), "Wasp", true},
		{"spell", 0, "SPELL_INVALID", true}, {"spell", 1, "SPELL_ANCHOR", true},
		{"ability", 0, "ABILITY_INVALID", true}, {"ability", 1, "ABILITY_BERSERKER_CHARGE", true},
	} {
		for _, mask := range []uint32{0, 1, 5, 10, 15} {
			for _, xor := range []bool{false, true} {
				t.Run(fmt.Sprintf("type=%s/param=%s/modifiers=%d/xor=%t", tc.kind, tc.name, mask, xor), func(t *testing.T) {
					typ.Xfer = nil
					switch tc.kind {
					case "guide":
						typ.Xfer = Get_nox_xxx_XFerFieldGuide_4F6390()
					case "ability":
						typ.Xfer = Get_nox_xxx_XFerAbilityReward_4F6240()
					}
					item := server.ShopkeeperItemDefinition{TypeInd: uint32(typ.Ind()), Count: 2, Param: tc.param}
					chunks := [][]byte{{2}}
					chunks = append(chunks, xferStringChunks237("Sword", true)...)
					chunks = append(chunks, xferStringChunks237(tc.name, tc.present)...)
					for i := range item.ModifierSlots {
						name := ""
						present := mask&(1<<i) != 0
						if present {
							item.ModifierSlots[i], name = slot, "Material1"
						}
						chunks = append(chunks, xferStringChunks237(name, present)...)
					}
					checkXferChecksum237(t, xor, chunks, func(cf *cryptfile.CryptFile) error {
						return monsterXferShopItem528DB0(cf, srv, &item, 64)
					})
				})
			}
		}
	}
}

func TestMonsterString8Read237(t *testing.T) {
	for _, value := range []string{"", "A", strings.Repeat("x", 255)} {
		for _, xor := range []bool{false, true} {
			for _, key := range []int{-1, crypt.MapKey} {
				t.Run(fmt.Sprintf("bytes=%d/xor=%t/key=%d", len(value), xor, key), func(t *testing.T) {
					plain := append([]byte{byte(len(value))}, value...)
					plain = append(plain, 0x5a)
					path := filepath.Join(t.TempDir(), "owned-read.bin")
					if err := os.WriteFile(path, cryptFileWire236(t, plain, key, xor), 0o600); err != nil {
						t.Fatal(err)
					}
					cf, err := cryptfile.OpenFile(path, cryptfile.ReadOnly, key)
					if err != nil {
						t.Fatal(err)
					}
					t.Cleanup(func() { _ = cf.Close() })
					cf.SetXOR(xor)
					got, err := monsterRWString8(cf, "ignored write value")
					if err != nil || got != value {
						t.Fatalf("read = %q, %v; want %q", got, err, value)
					}
					if got, err := cf.ReadU8(); err != nil || got != 0x5a {
						t.Fatalf("sentinel = %x, %v", got, err)
					}
				})
			}
		}
	}
}
