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
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"
	crypt "github.com/opennox/noxcrypt"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// GAME.EXE 00529706..00529A6E: version 61 adds the shop sell
// multiplier, 62 adds buffs, 63 adds wounded-NPC voice, and 64 adds
// poison. Each chunk below is one original file-wrapper call; absent
// versioned fields must neither consume the next record nor alter its CRC.
func monsterTailChunks242(version int, subclass object.MonsterClass, dormant byte) [][]byte {
	u32 := func(v uint32) []byte { return binary.LittleEndian.AppendUint32(nil, v) }
	out := [][]byte{{dormant}}
	if subclass.Has(object.MonsterShopkeeper) {
		out = append(out, u32(math.Float32bits(1.25)))
		if version >= 61 {
			out = append(out, u32(math.Float32bits(2.5)))
		}
		out = append(out, []byte{6}, []byte("market"), []byte{0})
	}
	out = append(out, u32(0xdeadface), u32(0), []byte{50, 0},
		[]byte{1}, []byte{2}, []byte{3}, []byte{4})
	if subclass.Has(object.MonsterFemaleNPC) {
		for i := byte(0); i < 6; i++ {
			out = append(out, []byte{10 + i, 20 + i, 30 + i})
		}
		out = append(out, []byte{0}) // absent female voice
	}
	if version >= 62 {
		out = append(out, []byte{2, 0}, []byte{0}) // version 2, no buffs
	}
	if version >= 63 && subclass.Has(object.MonsterWoundedNPC) {
		out = append(out, []byte{0}) // absent wounded voice
	}
	if version >= 64 {
		out = append(out, []byte{0}) // zero does not clear live poison
	}
	return out
}

func monsterTailObject242(t *testing.T, srv *server.Server, typ *server.ObjectType, subclass object.MonsterClass, mode string, dormant byte) *server.Object {
	t.Helper()
	obj := srv.Objs.NewObject(typ)
	ud, freeUD := alloc.New(server.MonsterUpdateData{})
	health, freeHealth := alloc.New(server.HealthData{})
	shop, freeShop := alloc.New(server.ShopkeeperInitData{})
	// alloc.New takes a type witness, not an initial value.
	*health = server.HealthData{Cur: 50, Field2: 22, Max: 75}
	*shop = server.ShopkeeperInitData{BuyMultiplier: 1.25, SellMultiplier: 2.5}
	t.Cleanup(freeUD)
	t.Cleanup(freeHealth)
	t.Cleanup(freeShop)
	obj.ObjClass, obj.ObjSubClass, obj.ObjFlags = object.ClassMonster, object.SubClass(subclass), 0
	obj.UpdateData, obj.HealthData, obj.InitData = unsafe.Pointer(ud), health, unsafe.Pointer(shop)
	ud.Field0 = 0xdeadface
	ud.Field337, ud.Field335 = 0x11223301, 0x44556602
	ud.Field361, ud.Field509 = 0xaabb0003|uint32(dormant)<<8, 0x77889904
	for i := range ud.Color {
		ud.Color[i] = server.Color3{R: byte(10 + i), G: byte(20 + i), B: byte(30 + i)}
	}
	copy(shop.ShopText[:], "market")
	if mode == "read" {
		health.Cur, shop.SellMultiplier, obj.Poison540 = 99, 99, 9
	}
	return obj
}

func checkMonsterTailState242(t *testing.T, obj *server.Object, version int, subclass object.MonsterClass, mode string, host bool, dormant byte) {
	t.Helper()
	ud := obj.UpdateDataMonster()
	health := obj.HealthData
	wantHealth := [3]uint16{50, 22, 75}
	if mode == "read" && host && dormant != 0 {
		wantHealth = [3]uint16{0, 22, 0}
	}
	if got := [3]uint16{health.Cur, health.Field2, health.Max}; got != wantHealth {
		t.Errorf("health = %v, want %v", got, wantHealth)
	}
	if ud.Field0 != 0xdeadface || ud.Field337 != 0x11223301 || ud.Field335 != 0x44556602 ||
		ud.Field361 != 0xaabb0003|uint32(dormant)<<8 || ud.Field509 != 0x77889904 ||
		obj.ObjSubClass != object.SubClass(subclass) {
		t.Errorf("partial-word definition/subclass mutation: %#x/%#x/%#x/%#x/%#x/%#x",
			ud.Field0, ud.Field337, ud.Field335, ud.Field361, ud.Field509, obj.ObjSubClass)
	}
	wantSell := float32(2.5)
	if mode == "read" && (!subclass.Has(object.MonsterShopkeeper) || version < 61) {
		wantSell = 99
	}
	shop := obj.InitDataShopkeeper()
	if shop.SellMultiplier != wantSell || shop.BuyMultiplier != 1.25 || string(bytes.TrimRight(shop.ShopText[:], "\x00")) != "market" || shop.Count != 0 {
		t.Errorf("shop fields = %+v, want buy=1.25 sell=%v text=market count=0", shop, wantSell)
	}
	wantPoison := byte(0)
	if mode == "read" {
		wantPoison = 9
	}
	if obj.Poison540 != wantPoison {
		t.Errorf("poison = %d, want %d", obj.Poison540, wantPoison)
	}
	for i, color := range ud.Color {
		if color != (server.Color3{R: byte(10 + i), G: byte(20 + i), B: byte(30 + i)}) {
			t.Errorf("color[%d] = %v", i, color)
		}
	}
	row, err := json.Marshal(map[string]any{
		"test": t.Name(), "version": version, "subclass": uint32(subclass), "mode": mode, "host": host, "dormant": dormant,
		"health": [3]uint16{health.Cur, health.Field2, health.Max}, "field0": ud.Field0,
		"field337": ud.Field337, "field335": ud.Field335, "field361": ud.Field361, "field509": ud.Field509,
		"shop_sell_bits": math.Float32bits(shop.SellMultiplier), "subclass_after": uint32(obj.ObjSubClass), "poison": obj.Poison540,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("R242 state %s", row)
}

func TestMonsterTailVersions242V4(t *testing.T) {
	oldFlags := noxflags.GetGame()
	t.Cleanup(func() { noxflags.ResetGame(); noxflags.SetGame(oldFlags) })
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	if !srv.Objs.Init(512) {
		t.Fatal("private object allocator")
	}
	t.Cleanup(srv.Objs.FreeObjects)
	srv.FreeObjectTypes()
	if err := srv.Types.ReadObjectType(&things.Thing{Name: "TailVersionFixture242"}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(srv.FreeObjectTypes)
	typ := srv.Types.ByID("TailVersionFixture242")
	for version := 60; version <= 64; version++ {
		for _, subclass := range []object.MonsterClass{0, object.MonsterShopkeeper, object.MonsterFemaleNPC, object.MonsterWoundedNPC, object.MonsterShopkeeper | object.MonsterFemaleNPC | object.MonsterWoundedNPC} {
			for _, host := range []bool{false, true} {
				for _, dormant := range []byte{0, 1} {
					chunks := monsterTailChunks242(version, subclass, dormant)
					for _, mode := range []string{"read", "write"} {
						encodings := []string{"cipher", "xor"}
						if mode == "read" {
							encodings = append(encodings, "raw")
						}
						for _, encoding := range encodings {
							t.Run(fmt.Sprintf("version=%d/subclass=%x/mode=%s/host=%t/dormant=%d/encoding=%s", version, uint32(subclass), mode, host, dormant, encoding), func(t *testing.T) {
								noxflags.ResetGame()
								if host {
									noxflags.SetGame(noxflags.GameHost)
								}
								obj := monsterTailObject242(t, srv, typ, subclass, mode, dormant)
								xor := encoding == "xor"
								if mode == "write" {
									withNextField := append(append([][]byte(nil), chunks...), []byte{0x5a})
									checkXferChecksum237(t, xor, withNextField, func(cf *cryptfile.CryptFile) error {
										if err := monsterXferTail528DB0(cf, obj, version); err != nil {
											return err
										}
										checkMonsterTailState242(t, obj, version, subclass, mode, host, dormant)
										return cf.WriteU8(0x5a)
									})
									return
								}
								key := crypt.MapKey
								if encoding == "raw" {
									key = -1
								}
								plain := bytes.Join(chunks, nil)
								path := filepath.Join(t.TempDir(), "versioned-tail.bin")
								if err := os.WriteFile(path, cryptFileWire236(t, plain, key, xor), 0600); err != nil {
									t.Fatal(err)
								}
								cf, err := cryptfile.OpenFile(path, cryptfile.ReadOnly, key)
								if err != nil {
									t.Fatal(err)
								}
								t.Cleanup(func() { _ = cf.Close() })
								cf.SetXOR(xor)
								if err := monsterXferTail528DB0(cf, obj, version); err != nil {
									t.Fatal(err)
								}
								position, err := cf.File.Seek(0, io.SeekCurrent)
								if xor {
									// XOR I/O bypasses the cipher reader; query that same
									// raw backend rather than its unused cipher cursor.
									position, err = cf.File.File.Seek(0, io.SeekCurrent)
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
								checkMonsterTailState242(t, obj, version, subclass, mode, host, dormant)
								record, _ := json.Marshal(map[string]any{"test": t.Name(), "payload": hex.EncodeToString(plain), "crc": gotCRC, "bytes": position})
								t.Logf("R242 actual %s", record)
							})
						}
					}
				}
			}
		}
	}
}
