package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

// The unchanged PE32 root/carry/owner/FISTP produced this independent seal.
// Armor wear, DefaultDamage, facing, flags, buffs and audio are declared
// services. This compares the prefix outputs, not whole Windows gameplay.
func TestPlayerDamageSelfArrow4E17B0OriginalPrefixTranscript(t *testing.T) {
	got := nonunitDamagePrefixTranscript4E17B0(t, 0)
	const want = "fdc72e86f8333ca03df633733ce67e8231a18fe9b994caf9ed5320c0a1cb4f28"
	if sum := fmt.Sprintf("%x", sha256.Sum256(got)); sum != want {
		t.Fatalf("1164 original prefix rows SHA=%s want=%s", sum, want)
	}
	t.Logf("1164 unchanged PE32 prefix rows bytes=%d SHA=%s", len(got), want)
}

func nonunitDamagePrefixTranscript4E17B0(t *testing.T, shape uint32) []byte {
	t.Helper()
	armors, carries := []float32{0, 0.25, 0.5, 1}, []float32{-0.25, 0, 0.25, 0.5}
	damages := []int32{-9, 0, 1, 3, 64, 200}
	var output bytes.Buffer
	encode := json.NewEncoder(&output)
	run := func(params [8]uint32) {
		s, player, armor, carry, dmg, quest, god, defense := params[0], params[1] != 0, params[2], params[3], params[4], params[5], params[6] != 0, params[7]
		v, arrow := selfArrowFixture4E17B0(t, player)
		a, w, typ := arrow, arrow, object.DamageImpale
		if s == 1 {
			a = &Object{TypeInd: 734, ObjClass: object.Class(4194816), ObjFlags: object.Flags(525)}
			w = &Object{TypeInd: 1217, ObjClass: object.Class(1048576), ObjFlags: object.Flags(25166348), PrevPos: arrow.PrevPos, PosVec: arrow.PosVec}
			typ = object.DamageCrush
		} else if s == 2 {
			a = &Object{TypeInd: 1399, ObjFlags: object.Flags(16777796)}
			w = &Object{TypeInd: 695, ObjClass: object.Class(2621441), ObjSubClass: 1, ObjFlags: object.Flags(553665028), PrevPos: arrow.PrevPos, PosVec: arrow.PosVec}
			typ = object.DamageFlame
		}
		var marker, kind, carryBits *uint32
		if player {
			u := v.UpdateDataPlayer()
			u.Field57 = math.Float32bits(armors[armor])
			u.Field21, u.Field76, u.Field75 = math.Float32bits(carries[carry]), 77, 88
			marker, kind, carryBits = &u.Field76, &u.Field75, &u.Field21
		} else {
			u := v.UpdateDataMonster()
			u.Field518 = math.Float32bits(armors[armor])
			u.Field1, u.Field547, u.Field546 = math.Float32bits(carries[carry]), 77, 88
			marker, kind, carryBits = &u.Field547, &u.Field546, &u.Field1
		}
		if defense == 1 {
			v.Buffs = 1 << playerDamageInvulnerableEnchant4E17B0
		} else if defense == 2 {
			v.Buffs = 1 << playerDamageReflectEnchant4E17B0
		} else if defense == 3 {
			if player {
				u := v.UpdateDataPlayer()
				u.State, u.Player.ArmorEquip = PlayerState16, 0x1000000
			} else {
				u := v.UpdateDataMonster()
				u.AIStack[0].Action, u.ArmorEquipFlags = uint32(ai.ACTION_BLOCK_ATTACK), 0x1000000
			}
		} else if defense == 4 {
			v.ObjFlags |= object.FlagNoUpdate
		} else if defense == 5 {
			a.ObjOwner = v
		} else if defense == 6 {
			v.ObjFlags |= object.FlagDead
		}
		r := damageMeleeRuntimeFixture4E17B0(t)
		r.Frame = func() uint32 { return 1400 }
		r.CoopMode = func() bool { return defense == 5 }
		r.GodMode = func() bool { return god }
		r.QuestMode = func() bool { return quest != 0 }
		r.QuestDamageScale = func() float32 { return []float32{1, 0, 0.5}[quest] }
		r.BlockSourceExcluded = func(*Object) bool { return s == 1 }
		amount, calls, audio, reflects, directions := uint32(0xfeedface), uint32(0), uint32(0), uint32(0), uint32(0)
		r.BlockDirection = func(*Object, types.Pointf) bool { directions++; return defense == 2 || defense == 3 }
		r.ProjectileReflect = func(*Object, *Object) { reflects++ }
		r.ClearOwner = func(o *Object) { o.ObjOwner = nil }
		r.SetOwner = func(owner, o *Object) { o.ObjOwner = owner }
		r.Audio = func(n int, _ *Object) { audio = uint32(n) }
		r.BlockDamagePercent = func() float64 { return 0.25 }
		r.DamageBlockItem = func(*Object, *Object, *Object, *Object, float32, object.DamageType) bool { return true }
		r.DefaultDamage = func(target, source, weapon *Object, d int32, k object.DamageType) bool {
			if target != v || source != a || weapon != w || k != typ {
				t.Fatal("prefix default identity/type lost")
			}
			amount, calls = uint32(d), calls+1
			return true
		}
		h, result := PlayerDamageNative4E17B0(v, a, w, damages[dmg], typ, r)
		if !h {
			t.Fatalf("prefix not admitted: %v", params)
		}
		bool32 := func(value bool) uint32 {
			if value {
				return 1
			}
			return 0
		}
		row := append(params[:], bool32(result), *marker, *kind, *carryBits, amount, calls, audio, reflects, bool32(w.ObjOwner == v), directions)
		if err := encode.Encode(row); err != nil {
			t.Fatal(err)
		}
	}
	for player := range uint32(2) {
		for armor := range uint32(4) {
			for carry := range uint32(4) {
				for dmg := range uint32(6) {
					for quest := range uint32(3) {
						for god := range uint32(2) {
							run([8]uint32{shape, player, armor, carry, dmg, quest, god, 0})
						}
					}
				}
			}
		}
	}
	dmg := []uint32{3, 5, 4}[shape]
	for player := range uint32(2) {
		for defense := uint32(1); defense < 7; defense++ {
			run([8]uint32{shape, player, 1, 2, dmg, 0, 0, defense})
		}
	}
	return output.Bytes()
}
