package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

type worldProjectileCase4E0B30 struct {
	name                     string
	ownerClass, missileClass object.Class
	typ                      object.DamageType
	radial                   bool
}

func worldProjectileCases4E0B30() []worldProjectileCase4E0B30 {
	return []worldProjectileCase4E0B30{
		{"ArrowTrap1", object.ClassImmobile | object.ClassVisibleEnable, object.ClassMissile | object.ClassComplex | object.ClassWeapon | object.ClassNotStackable, object.DamageImpale, false},
		{"ArrowTrap2", object.ClassImmobile | object.ClassVisibleEnable, object.ClassMissile | object.ClassWeapon, object.DamageImpale, false},
		{"SkullStrongFireball", object.ClassImmobile | object.ClassVisibleEnable, object.ClassMissile | object.ClassLight | object.ClassComplex, object.DamageFlame, false},
		{"TowerDeathBall", object.ClassImmobile, object.ClassMissile | object.ClassLight | object.ClassSimple, object.DamageCrush, false},
		{"TowerDeathBallFragment", object.ClassImmobile, object.ClassMissile | object.ClassLight | object.ClassSimple, object.DamageCrush, false},
		{"DeathBallNearby", object.ClassImmobile, object.ClassMissile | object.ClassLight | object.ClassSimple, object.DamageCrush, true},
	}
}

func worldProjectileFixture4E0B30(tc worldProjectileCase4E0B30) (source, weapon, missile *Object) {
	owner := &Object{TypeInd: 600, ObjClass: tc.ownerClass, PrevPos: types.Ptf(88, 92)}
	missile = &Object{TypeInd: 601, ObjClass: tc.missileClass, ObjSubClass: 0x10, ObjOwner: owner, PrevPos: types.Ptf(34, 56), InitData: unsafe.Pointer(&ModifierInitData{})}
	if tc.radial {
		return missile, nil, missile
	}
	return owner, missile, missile
}

func TestDefaultDamageWorld4E0B30WorldProjectiles(t *testing.T) {
	for _, tc := range worldProjectileCases4E0B30() {
		for _, kind := range []string{"Player", "Monster", "NPC"} {
			t.Run(tc.name+"/"+kind, func(t *testing.T) {
				v := damageMeleeUnitFixture4E17B0(t, kind == "Player")
				if kind == "Monster" {
					v.ObjSubClass = 0
				}
				a, w, m := worldProjectileFixture4E0B30(tc)
				r := damageMeleeWorldRuntime4E0B30(t)
				r.MonsterHasHitSound = nil // A world source has no Monster update.
				r.FireProtection = func(*Object) float64 { return 0 }
				reason := ""
				r.Unsupported = func(why string, _, _, _ *Object, _ int32, _ object.DamageType) { reason = why }
				if !DefaultDamageWorld4E0B30(v, a, w, 25, tc.typ, r) || v.HealthData.Cur != 175 || reason != "" {
					t.Fatalf("world projectile lost HP tail: HP=200->%d unsupported=%q", v.HealthData.Cur, reason)
				}
				pos := m.PrevPos
				if w != nil && w.Class().Has(object.ClassWeapon) {
					pos = a.PrevPos
				}
				if v.Obj130 != m || v.Pos132 != pos || v.Field131 != uint32(tc.typ) || v.Frame134 != 1400 {
					t.Fatalf("lost attribution: obj=%p pos=%v type=%d frame=%d", v.Obj130, v.Pos132, v.Field131, v.Frame134)
				}
			})
		}
	}
}
