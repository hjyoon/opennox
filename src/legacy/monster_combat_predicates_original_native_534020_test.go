package legacy

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"math"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

type combatPredicateInput534020 struct {
	Class          uint32    `json:"cls"`
	Subclass       uint32    `json:"subclass"`
	Status         uint32    `json:"status"`
	Weapons        uint32    `json:"weapons"`
	Armor          uint32    `json:"armor"`
	RangeBits      uint32    `json:"range_bits"`
	MissileFirst   uint8     `json:"missile_first"`
	SpeedBits      uint32    `json:"speed_bits"`
	AggressionBits uint32    `json:"aggression_bits"`
	Head           int8      `json:"head"`
	Action         uint32    `json:"action"`
	Prefix528      uint32    `json:"prefix528"`
	Direction      uint16    `json:"direction"`
	Vector         [2]uint32 `json:"vector"`
}

func combatPredicateSnapshot534020(ptr unsafe.Pointer, size uintptr) []byte {
	return append([]byte(nil), unsafe.Slice((*byte)(ptr), int(size))...)
}

func TestMonsterCombatPredicatesOriginalActualC534020(t *testing.T) {
	path := os.Getenv("NOX_PREDICATE_REFERENCE")
	if path == "" {
		path = "../server/testdata/monster_combat_predicates_original_534020.json.gz"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(data)
	var expectedControl uint16
	switch hex.EncodeToString(digest[:]) {
	case "3e22140d34890c7309037a368e91a10d5cf580a3868c28bbac491ad55b776bc3":
		expectedControl = 0xe3f
	case "be58d4b70d2045260e7025f6aeefa58a8ba44a306216a911a979539fa784858b":
		expectedControl = 0x23f
	default:
		t.Fatal("unsealed original predicate reference")
	}
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	var reference struct {
		Schema      string `json:"schema"`
		NativeRange struct {
			RunID       int64  `json:"ci_run_id"`
			Bytes       int    `json:"original_bytes"`
			Contexts    int    `json:"range_contexts"`
			Native32SHA string `json:"native32_result_sha256"`
			Native64SHA string `json:"native64_result_sha256"`
		} `json:"native_range_reference"`
		OriginalSHA string `json:"original_sha256"`
		Control     uint16 `json:"control_word"`
		Rows        []struct {
			Name      string                     `json:"name"`
			Entry     uint32                     `json:"entry"`
			Input     combatPredicateInput534020 `json:"input"`
			Return    uint32                     `json:"return_u32"`
			ReturnF64 uint64                     `json:"return_f64_bits"`
		} `json:"rows"`
	}
	if err := json.NewDecoder(reader).Decode(&reference); err != nil {
		t.Fatal(err)
	}
	if reference.OriginalSHA != "0040e2c0683b4d73a5fb976e400d5087dca680df2b195c9e27f8edbda2d4974a" ||
		reference.Schema != "original-monster-combat-predicates-534020-v2" ||
		reference.Control != expectedControl || len(reference.Rows) != 11835 ||
		reference.NativeRange.RunID != 41108651935 || reference.NativeRange.Bytes != 20 ||
		reference.NativeRange.Contexts != 119 ||
		reference.NativeRange.Native32SHA != "f2e8bf24caaa0815c4f10c455b427f0b0a371886d56b2bf806e30b679c5914bb" ||
		reference.NativeRange.Native64SHA != "eaea14051d55b57e644db0ac7673eb87c1745c1c6f408bc290e6ab90fc3938c8" {
		t.Fatal("original predicate reference metadata")
	}
	var directions [2048]byte
	for i := 0; i < 256; i++ {
		x, y := server.SinCosDir(byte(i))
		binary.LittleEndian.PutUint32(directions[i*8:], math.Float32bits(x))
		binary.LittleEndian.PutUint32(directions[i*8+4:], math.Float32bits(y))
	}
	directionHash := sha256.Sum256(directions[:])
	if hex.EncodeToString(directionHash[:]) != "fbfdbd07ab3c6fbec2eb3768e69f1c32eb63c77b6e80e9957b9567c78d61bfcf" {
		t.Fatal("compiled Go direction table differs from unchanged GAME.EXE")
	}
	type unitGuard struct {
		Before [64]byte
		Value  server.Object
		After  [64]byte
	}
	type updateGuard struct {
		Before [64]byte
		Value  server.MonsterUpdateData
		After  [64]byte
	}
	type definitionGuard struct {
		Before [64]byte
		Value  server.MonsterDef
		After  [64]byte
	}
	type vectorGuard struct {
		Before [64]byte
		Value  types.Pointf
		After  [64]byte
	}
	ug, freeU := alloc.New(unitGuard{})
	t.Cleanup(freeU)
	udg, freeUD := alloc.New(updateGuard{})
	t.Cleanup(freeUD)
	dg, freeD := alloc.New(definitionGuard{})
	t.Cleanup(freeD)
	vg, freeV := alloc.New(vectorGuard{})
	t.Cleanup(freeV)
	for _, region := range []struct {
		ptr  unsafe.Pointer
		size uintptr
	}{{unsafe.Pointer(ug), unsafe.Sizeof(*ug)}, {unsafe.Pointer(udg), unsafe.Sizeof(*udg)},
		{unsafe.Pointer(dg), unsafe.Sizeof(*dg)}, {unsafe.Pointer(vg), unsafe.Sizeof(*vg)}} {
		for i := range unsafe.Slice((*byte)(region.ptr), int(region.size)) {
			unsafe.Slice((*byte)(region.ptr), int(region.size))[i] = byte(i*31 + 19)
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(region.ptr) <= math.MaxUint32 {
			t.Fatal("actual C predicate records must exceed4GiB")
		}
	}
	unit, update, def, vec := &ug.Value, &udg.Value, &dg.Value, &vg.Value
	controlOnly := os.Getenv("NOX_PREDICATE_REFERENCE_CONTROL_ONLY") == "true"
	entries := make(map[uint32]int)
	for _, row := range reference.Rows {
		t.Run(row.Name, func(t *testing.T) {
			c := row.Input
			*unit = server.Object{ObjClass: object.Class(c.Class), ObjSubClass: object.SubClass(c.Subclass),
				UpdateData: unsafe.Pointer(update), SpeedCur: math.Float32frombits(c.SpeedBits), Direction1: server.Dir16(c.Direction)}
			*update = server.MonsterUpdateData{MonsterDef: def, Aggression: math.Float32frombits(c.AggressionBits),
				StatusFlags: object.MonsterStatus(c.Status), WeaponEquipFlags: c.Weapons, ArmorEquipFlags: c.Armor,
				AIStackInd: c.Head, Field132: c.Prefix528}
			if c.Head >= 0 {
				update.AIStack[c.Head].Action = c.Action
			}
			*def = server.MonsterDef{MeleeAttackRange112: math.Float32frombits(c.RangeBits)}
			def.MissileName148[0] = c.MissileFirst
			*vec = types.Ptf(math.Float32frombits(c.Vector[0]), math.Float32frombits(c.Vector[1]))
			regions := []struct {
				ptr  unsafe.Pointer
				size uintptr
			}{{unsafe.Pointer(ug), unsafe.Sizeof(*ug)}, {unsafe.Pointer(udg), unsafe.Sizeof(*udg)},
				{unsafe.Pointer(dg), unsafe.Sizeof(*dg)}, {unsafe.Pointer(vg), unsafe.Sizeof(*vg)}}
			before := make([][]byte, len(regions))
			for i, region := range regions {
				before[i] = combatPredicateSnapshot534020(region.ptr, region.size)
			}
			target := combatPredicateOriginalPublicC534020(row.Entry)
			if target == nil {
				t.Fatalf("missing actual public C entry %#x", row.Entry)
			}
			entries[row.Entry]++
			if !controlOnly {
				if row.Entry == 0x534470 {
					got := math.Float64bits(combatPredicateOriginalDouble534470(target, unit))
					if got != row.ReturnF64 {
						t.Errorf("original range bits got=%016x original=%016x", got, row.ReturnF64)
					}
				} else {
					var got int
					if row.Entry == 0x534120 {
						got = ccall.CallIntPtr2(target, unit.CObj(), unsafe.Pointer(vec))
					} else {
						got = ccall.CallIntPtr(target, unit.CObj())
					}
					if uint32(got) != row.Return {
						t.Errorf("original predicate %#x got=%08x original=%08x", row.Entry, uint32(got), row.Return)
					}
					if row.Entry == 0x534300 && Nox_xxx_monsterCanCast_534300(unit) != (row.Return != 0) {
						t.Error("existing production Go→C caster bridge differs from original")
					}
				}
			}
			for i, region := range regions {
				if !bytes.Equal(before[i], combatPredicateSnapshot534020(region.ptr, region.size)) {
					t.Errorf("pure predicate changed native input/guard region%d", i)
				}
			}
		})
	}
	if len(entries) != 13 {
		t.Fatal("original predicate root coverage")
	}
	t.Logf("original actual C contexts=%d entries=%d control=%#x preparation-control-only=%t", len(reference.Rows), len(entries), reference.Control, controlOnly)
}
