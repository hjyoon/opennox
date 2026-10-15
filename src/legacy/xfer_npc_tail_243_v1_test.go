//go:build amd64 || arm64

package legacy

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	crypt "github.com/opennox/noxcrypt"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// GAME.EXE 0052B81A..0052B9B4: experience starts at 51, voice at 52,
// buffs at 61, and poison at 62. Each chunk is one original file-wrapper
// call. The version-60 Kalen record ends after its voice, without buffs.
func npcTailChunks243(version int, mode string, dormant byte, namedVoice bool) [][]byte {
	u32 := func(v uint32) []byte { return binary.LittleEndian.AppendUint32(nil, v) }
	maximum, subclass, experience := uint32(75), uint32(0x180), uint32(0x47afc800)
	if mode == "read" {
		maximum, subclass, experience = 0xbeef004b, 0x181, 0x7fc12345
	}
	out := [][]byte{{dormant}, u32(0xdeadface), u32(maximum), u32(subclass), {50, 0}}
	if version >= 51 {
		out = append(out, u32(experience))
	}
	if version >= 52 {
		if namedVoice {
			out = append(out, []byte{3}, []byte("NPC"))
		} else {
			out = append(out, []byte{0}) // no body/CRC call for an absent voice
		}
	}
	if version >= 61 {
		out = append(out, []byte{2, 0}, []byte{0}) // empty version-2 buff list
	}
	if version >= 62 {
		out = append(out, []byte{0}) // zero does not clear the live poison value
	}
	return out
}

func npcTailObject243(t *testing.T, mode string, dormant byte, flag8000, namedVoice bool, voice, oldVoice unsafe.Pointer) *server.Object {
	t.Helper()
	obj, freeObj := alloc.New(server.Object{})
	ud, freeUD := alloc.New(server.MonsterUpdateData{})
	health, freeHealth := alloc.New(server.HealthData{})
	t.Cleanup(freeObj)
	t.Cleanup(freeUD)
	t.Cleanup(freeHealth)
	*health = server.HealthData{Cur: 50, Field2: 22, Max: 75}
	*obj = server.Object{
		ObjClass: object.ClassMonster, ObjSubClass: 0x8381, ObjFlags: 0x1000204,
		UpdateData: unsafe.Pointer(ud), HealthData: health, Experience: math.Float32frombits(0x47afc800),
	}
	if flag8000 {
		obj.ObjFlags |= 0x8000
	}
	ud.Field0, ud.Field361 = 0xdeadface, 0xaabb0003|uint32(dormant)<<8
	if namedVoice {
		ud.SoundSet122 = voice
	}
	if mode == "read" {
		obj.ObjSubClass, obj.Experience, obj.Poison540 = 0x8200, 99, 9
		health.Cur, health.Max = 99, 100
		ud.Field0, ud.Field361 = 0xc0ffee00, 0xaabb0003|uint32(dormant+2)<<8
		ud.SoundSet122 = oldVoice
	}
	return obj
}

func checkNPCTailState243(t *testing.T, obj *server.Object, version int, mode string, host bool, dormant byte, flag8000, namedVoice bool, voice, oldVoice unsafe.Pointer) {
	t.Helper()
	ud, health := obj.UpdateDataMonster(), obj.HealthData
	wantHealth := [3]uint16{50, 22, 75}
	if mode == "read" && host && dormant != 0 {
		wantHealth = [3]uint16{0, 22, 0}
	}
	if got := [3]uint16{health.Cur, health.Field2, health.Max}; got != wantHealth {
		t.Errorf("health = %v, want %v", got, wantHealth)
	}
	wantExperience, wantVoice := uint32(0x47afc800), unsafe.Pointer(nil)
	if namedVoice {
		wantVoice = voice
	}
	if mode == "read" {
		wantExperience = 0x42c60000
		if version >= 51 {
			wantExperience = 0x7fc12345
		}
		if version < 52 {
			wantVoice = oldVoice
		}
	}
	if math.Float32bits(obj.Experience) != wantExperience || ud.SoundSet122 != wantVoice {
		t.Errorf("experience bits/voice = %#x/%p, want %#x/%p", math.Float32bits(obj.Experience), ud.SoundSet122, wantExperience, wantVoice)
	}
	wantFlags := object.Flags(0x1000204)
	if flag8000 {
		wantFlags |= 0x8000
		if mode == "read" {
			wantFlags |= 0x40
		}
	}
	wantPoison := byte(0)
	if mode == "read" {
		wantPoison = 9
	}
	if ud.Field0 != 0xdeadface || ud.Field361 != 0xaabb0003|uint32(dormant)<<8 ||
		obj.ObjSubClass != 0x8381 || obj.ObjFlags != wantFlags || obj.Poison540 != wantPoison {
		t.Errorf("NPC state = %#x/%#x/%#x/%#x/%d", ud.Field0, ud.Field361, obj.ObjSubClass, obj.ObjFlags, obj.Poison540)
	}
	voiceName := "nil"
	if ud.SoundSet122 == voice {
		voiceName = "NPC"
	} else if ud.SoundSet122 == oldVoice {
		voiceName = "old"
	}
	row, err := json.Marshal(map[string]any{
		"test": t.Name(), "version": version, "mode": mode, "host": host, "dormant": dormant,
		"flag8000": flag8000, "named_voice": namedVoice,
		"health": [3]uint16{health.Cur, health.Field2, health.Max}, "field0": ud.Field0, "field361": ud.Field361,
		"experience_bits": math.Float32bits(obj.Experience), "subclass_after": uint32(obj.ObjSubClass),
		"flags_after": uint32(obj.ObjFlags), "voice_after": voiceName, "poison": obj.Poison540,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("R243 state %s", row)
}

func TestNPCTailVersions243V1(t *testing.T) {
	oldFlags, oldSets := noxflags.GetGame(), soundSetByName
	t.Cleanup(func() { noxflags.ResetGame(); noxflags.SetGame(oldFlags); soundSetByName = oldSets })
	voice, freeVoice := alloc.Malloc(soundSetRecordSize)
	oldVoice, freeOldVoice := alloc.Malloc(soundSetRecordSize)
	t.Cleanup(freeVoice)
	t.Cleanup(freeOldVoice)
	soundSetByName = map[string]*soundSetRecord{"NPC": {name: "NPC", data: voice}, "old": {name: "old", data: oldVoice}}
	for version := 50; version <= 62; version++ {
		for _, host := range []bool{false, true} {
			for _, dormant := range []byte{0, 1} {
				for _, flag8000 := range []bool{false, true} {
					for _, namedVoice := range []bool{false, true} {
						for _, mode := range []string{"read", "write"} {
							chunks := npcTailChunks243(version, mode, dormant, namedVoice)
							encodings := []string{"cipher", "xor"}
							if mode == "read" {
								encodings = append(encodings, "raw")
							}
							for _, encoding := range encodings {
								t.Run(fmt.Sprintf("version=%d/mode=%s/host=%t/dormant=%d/flag8000=%t/named_voice=%t/encoding=%s", version, mode, host, dormant, flag8000, namedVoice, encoding), func(t *testing.T) {
									noxflags.ResetGame()
									if host {
										noxflags.SetGame(noxflags.GameHost)
									}
									obj := npcTailObject243(t, mode, dormant, flag8000, namedVoice, voice, oldVoice)
									xor := encoding == "xor"
									if mode == "write" {
										withNextField := append(append([][]byte(nil), chunks...), []byte{0x5a})
										checkXferChecksum237(t, xor, withNextField, func(cf *cryptfile.CryptFile) error {
											if err := npcXferTail52ADE0(cf, obj, version); err != nil {
												return err
											}
											checkNPCTailState243(t, obj, version, mode, host, dormant, flag8000, namedVoice, voice, oldVoice)
											return cf.WriteU8(0x5a)
										})
										return
									}
									key := crypt.MapKey
									if encoding == "raw" {
										key = -1
									}
									plain := bytes.Join(chunks, nil)
									path := filepath.Join(t.TempDir(), "versioned-npc-tail.bin")
									if err := os.WriteFile(path, cryptFileWire236(t, plain, key, xor), 0600); err != nil {
										t.Fatal(err)
									}
									cf, err := cryptfile.OpenFile(path, cryptfile.ReadOnly, key)
									if err != nil {
										t.Fatal(err)
									}
									t.Cleanup(func() { _ = cf.Close() })
									cf.SetXOR(xor)
									if err := npcXferTail52ADE0(cf, obj, version); err != nil {
										t.Fatal(err)
									}
									var position int64
									if xor {
										position, err = cf.File.File.Seek(0, io.SeekCurrent)
									} else {
										position, err = cf.File.Seek(0, io.SeekCurrent)
									}
									if err != nil || position != int64(len(plain)) {
										t.Fatalf("tail consumed %d, %v; want exactly %d", position, err, len(plain))
									}
									gotCRC, wantCRC := readStackCRC238(t, cf), uint32(0xffffffff)
									for _, chunk := range chunks {
										wantCRC = xferCRC237(wantCRC, chunk)
									}
									if gotCRC != wantCRC {
										t.Errorf("read CRC = %#x, want %#x", gotCRC, wantCRC)
									}
									checkNPCTailState243(t, obj, version, mode, host, dormant, flag8000, namedVoice, voice, oldVoice)
									record, _ := json.Marshal(map[string]any{"test": t.Name(), "payload": hex.EncodeToString(plain), "crc": gotCRC, "bytes": position})
									t.Logf("R243 actual %s", record)
								})
							}
						}
					}
				}
			}
		}
	}
}
