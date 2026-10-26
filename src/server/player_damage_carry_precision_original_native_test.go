package server

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"math"
	"os"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

type originalNPCEnvironmentCarryCase struct {
	Name           string            `json:"name"`
	DamageType     object.DamageType `json:"damage_type"`
	Raw            int32             `json:"raw"`
	Absorption     uint32            `json:"absorption_bits32"`
	Carry          uint32            `json:"carry_bits32"`
	Quest          bool              `json:"quest"`
	Scale          uint32            `json:"quest_scale_bits32"`
	ExpectedCarry  uint32            `json:"expected_carry_bits32"`
	ExpectedWear   uint32            `json:"expected_wear_bits32"`
	ExpectedDamage uint32            `json:"expected_damage_bits32"`
}

type originalPlayerDamageCarryFixture struct {
	Environment []originalNPCEnvironmentCarryCase `json:"environment_cases"`
	Carry       []struct {
		Name            string `json:"name"`
		Class           uint32 `json:"class_bits"`
		Amount          uint32 `json:"amount_bits32"`
		Carry           uint32 `json:"carry_bits32"`
		TargetNil       bool   `json:"target_nil"`
		OutputNil       bool   `json:"output_nil"`
		ExpectedOutput  uint32 `json:"expected_output_bits32"`
		ExpectedMonster uint32 `json:"expected_monster_carry_bits32"`
		ExpectedPlayer  uint32 `json:"expected_player_carry_bits32"`
	} `json:"carry_cases"`
}

func loadOriginalPlayerDamageCarryFixture(t *testing.T) originalPlayerDamageCarryFixture {
	t.Helper()
	b, err := os.ReadFile("testdata/player_damage_carry_precision_original.json")
	if err != nil {
		t.Fatal(err)
	}
	var f originalPlayerDamageCarryFixture
	if err = json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	if len(f.Environment) != 624 || len(f.Carry) != 704 {
		t.Fatal("incomplete independently executed original NPC contexts")
	}
	return f
}

func TestPlayerDamageCarryPrecisionOriginal4E20F0Native(t *testing.T) {
	f := loadOriginalPlayerDamageCarryFixture(t)
	objectSize := uintptr(780)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		objectSize = 928
	}
	if unsafe.Sizeof(Object{}) != objectSize {
		t.Fatal("changed native object layout")
	}
	for _, c := range f.Carry {
		t.Run(c.Name, func(t *testing.T) {
			obj, freeObj := alloc.Make([]byte(nil), 64+objectSize+64)
			defer freeObj()
			udSize := max(unsafe.Sizeof(PlayerUpdateData{}), unsafe.Sizeof(MonsterUpdateData{}))
			update, freeUpdate := alloc.Make([]byte(nil), int(128+udSize))
			defer freeUpdate()
			output, freeOutput := alloc.Make([]byte(nil), 12)
			defer freeOutput()
			for _, b := range [][]byte{obj, update, output} {
				for i := range b {
					b[i] = 0xa5
				}
			}
			v := (*Object)(unsafe.Pointer(&obj[64]))
			v.ObjClass = object.Class(c.Class)
			v.UpdateData = unsafe.Pointer(&update[64])
			monster := (*MonsterUpdateData)(v.UpdateData)
			player := (*PlayerUpdateData)(v.UpdateData)
			monster.Field1, player.Field21 = c.Carry, c.Carry
			binary.LittleEndian.PutUint32(output[4:8], 0x6b6b6b6b)
			for _, ptr := range []unsafe.Pointer{unsafe.Pointer(v), v.UpdateData, unsafe.Pointer(&output[4])} {
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ptr) <= 0xffffffff {
					t.Fatal("C-owned native frame not above4GiB")
				}
			}
			beforeObj := append([]byte(nil), obj...)
			beforeUpdate := append([]byte(nil), update...)
			beforeOutput := append([]byte(nil), output...)
			out := (*int32)(unsafe.Pointer(&output[4]))
			if c.OutputNil {
				out = nil
			}
			target := v
			if c.TargetNil {
				target = nil
			}
			playerDamageCarry4E20F0(target, out, math.Float32frombits(c.Amount))
			if binary.LittleEndian.Uint32(output[4:8]) != c.ExpectedOutput || monster.Field1 != c.ExpectedMonster || player.Field21 != c.ExpectedPlayer {
				t.Fatalf("actual native carry output/monster/player=%08x/%08x/%08x want=%08x/%08x/%08x", binary.LittleEndian.Uint32(output[4:8]), monster.Field1, player.Field21, c.ExpectedOutput, c.ExpectedMonster, c.ExpectedPlayer)
			}
			if !c.TargetNil && !c.OutputNil {
				copy(beforeOutput[4:8], output[4:8])
				var off uintptr
				switch {
				case c.Class&4 != 0:
					off = 64 + unsafe.Offsetof(player.Field21)
				case c.Class&2 != 0:
					off = 64 + unsafe.Offsetof(monster.Field1)
				}
				if off != 0 {
					copy(beforeUpdate[off:off+4], update[off:off+4])
				}
			}
			if !bytes.Equal(beforeObj, obj) || !bytes.Equal(beforeUpdate, update) || !bytes.Equal(beforeOutput, output) {
				t.Fatal("native carry nonoutput or guard bytes changed")
			}
		})
	}
}

// The complete original PlayerDamage/carry/FISTP/Quest-getter/action-getter
// bodies execute under the actual startup control word. Armor and HP effects
// remain external boundaries; this comparison uses empty native inventory.
func TestPlayerDamageNPCEnvironmentPrecisionOriginal4E17B0Native(t *testing.T) {
	f := loadOriginalPlayerDamageCarryFixture(t)
	for _, c := range f.Environment {
		t.Run(c.Name, func(t *testing.T) {
			v, source, w := worldFlameFixture4E17B0(t, "self", "NPC")
			w.ObjClass = object.ClassDangerous | object.ClassImmobile
			if c.DamageType == object.DamageLava {
				source, w = nil, nil
				v.Buffs |= 1 << playerDamageReflectEnchant4E17B0
			}
			u := v.UpdateDataMonster()
			u.Field518, u.Field1 = c.Absorption, c.Carry
			u.Field547, u.Field546 = 99, 77
			frame := func(ptr unsafe.Pointer, n uintptr) []byte { return unsafe.Slice((*byte)(ptr), int(n)) }
			vb := append([]byte(nil), frame(unsafe.Pointer(v), unsafe.Sizeof(*v))...)
			ub := append([]byte(nil), frame(unsafe.Pointer(u), unsafe.Sizeof(*u))...)
			hp := *v.HealthData
			defaults, exclusions, faces, gods, quests, scales := 0, 0, 0, 0, 0, 0
			r := PlayerDamageRuntime4E17B0{
				BlockSourceExcluded: func(got *Object) bool {
					if got != w || u.Field547 != 0 {
						t.Fatal("cached exclusion prefix")
					}
					exclusions++
					return false
				},
				BlockDirection: func(got *Object, pos types.Pointf) bool {
					if got != v || w == nil || pos != w.PrevPos {
						t.Fatal("changed original facing inputs")
					}
					faces++
					return true
				},
				GodMode: func() bool {
					if u.Field547 != 2 || u.Field546 != uint32(c.DamageType) {
						t.Fatal("mode before original marker")
					}
					gods++
					return true
				},
				QuestMode:        func() bool { quests++; return c.Quest },
				QuestDamageScale: func() float32 { scales++; return math.Float32frombits(c.Scale) },
				DefaultDamage: func(got, s, weapon *Object, amount int32, typ object.DamageType) bool {
					if got != v || s != source || weapon != w || typ != c.DamageType {
						t.Fatal("changed original damage identities")
					}
					if uint32(amount) != c.ExpectedDamage || u.Field1 != c.ExpectedCarry {
						t.Fatalf("actual native NPC environment damage/carry=%08x/%08x want=%08x/%08x", uint32(amount), u.Field1, c.ExpectedDamage, c.ExpectedCarry)
					}
					defaults++
					return true
				},
				Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) { t.Fatal(reason) },
			}
			h, ok := PlayerDamageNative4E17B0(v, source, w, c.Raw, c.DamageType, r)
			wantExclusions := 0
			if c.DamageType == object.DamageImpale {
				wantExclusions = 1
			}
			wantScales := 0
			if c.Quest {
				wantScales = 1
			}
			if !h || !ok || defaults != 1 || exclusions != wantExclusions || faces != wantExclusions || gods != 1 || quests != 1 || scales != wantScales || u.Field1 != c.ExpectedCarry {
				t.Fatal("changed original service counts or carry")
			}
			if !bytes.Equal(vb, frame(unsafe.Pointer(v), unsafe.Sizeof(*v))) || *v.HealthData != hp {
				t.Fatal("native target/HP frame changed at excluded HP boundary")
			}
			after := frame(unsafe.Pointer(u), unsafe.Sizeof(*u))
			for _, off := range []uintptr{unsafe.Offsetof(u.Field1), unsafe.Offsetof(u.Field546), unsafe.Offsetof(u.Field547)} {
				copy(ub[off:off+4], after[off:off+4])
			}
			if !bytes.Equal(ub, after) {
				t.Fatal("native update nonoutput bytes changed")
			}
		})
	}
}
