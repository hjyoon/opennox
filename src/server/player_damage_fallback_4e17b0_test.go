package server

import (
	"math"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestPlayerDamageFallbackNativeMarkers4E17B0(t *testing.T) {
	for _, tc := range []struct {
		name      string
		source    *Object
		typ       object.DamageType
		amount    int32
		wear      int32
		wearCalls int
		want      int32
	}{
		{"weaponless-projectile", &Object{ObjClass: object.ClassMissile}, object.DamageImpale, 9, 5, 1, 4},
		{"reflect-without-attack", nil, object.DamageZapRay, 9, 0, 0, 9},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target := damageMeleeUnitFixture4E17B0(t, true)
			u := target.UpdateDataPlayer()
			u.Field57, u.Field21, u.Field76, u.Field75 = math.Float32bits(.5), math.Float32bits(.25), 99, 77
			if tc.source == nil {
				target.Buffs = 1 << playerDamageReflectEnchant4E17B0
			}
			forwarded, armor := 0, 0
			r := PlayerDamageRuntime4E17B0{
				BlockSourceOnlyExcluded: func(*Object) bool { return false },
				BlockDirection:          func(*Object, types.Pointf) bool { return false },
				DefaultDamage: func(v, source, weapon *Object, amount int32, typ object.DamageType) bool {
					if v != target || source != tc.source || weapon != nil || amount != tc.want || typ != tc.typ {
						t.Fatal("default damage arguments changed")
					}
					forwarded++
					return true
				},
			}
			wear := PlayerDamageWear4E17B0{Armor: func(v, source, weapon *Object, amount int32, typ object.DamageType) {
				if v != target || source != tc.source || weapon != nil || amount != tc.wear || typ != tc.typ {
					t.Fatal("armor wear arguments changed")
				}
				armor++
			}}
			ok := PlayerDamageFallback4E17B0(target, tc.source, nil, tc.amount, tc.typ, r, wear)
			if !ok || forwarded != 1 || armor != tc.wearCalls || u.Field76 != 2 || u.Field75 != uint32(tc.typ) {
				t.Fatalf("result %t, default calls %d, wear calls %d, marker %d/%d", ok, forwarded, armor, u.Field76, u.Field75)
			}
		})
	}
}
