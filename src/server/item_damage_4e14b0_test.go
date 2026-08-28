package server

import (
	"math"
	"testing"

	"github.com/opennox/libs/object"
)

func TestBallDamage4E14A0AlwaysRejects(t *testing.T) {
	if BallDamage4E14A0(&Object{}, &Object{}, &Object{}, 17, object.DamageBlade) {
		t.Fatal("BallDamage4E14A0 returned true")
	}
}

func TestWeaponDamage4E14B0OracleBranches(t *testing.T) {
	holder := &Object{}
	tests := []struct {
		name   string
		target *Object
		typ    object.DamageType
		want   bool
	}{
		{name: "nil target", target: nil, typ: object.DamageLava},
		{name: "wrong class", target: &Object{ObjClass: object.ClassArmor, InvHolder: holder}, typ: object.DamageBlade},
		{name: "loose weapon non-lava", target: &Object{ObjClass: object.ClassWeapon}, typ: object.DamageBlade},
		{name: "held weapon", target: &Object{ObjClass: object.ClassWeapon, InvHolder: holder}, typ: object.DamageBlade, want: true},
		{name: "held wand", target: &Object{ObjClass: object.ClassWand, InvHolder: holder}, typ: object.DamageElectric, want: true},
		{name: "loose weapon lava", target: &Object{ObjClass: object.ClassWeapon}, typ: object.DamageLava, want: true},
		{name: "loose wand lava", target: &Object{ObjClass: object.ClassWand}, typ: object.DamageLava, want: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			got := WeaponDamage4E14B0(tc.target, nil, nil, 2, tc.typ, func(
				target, source, weapon *Object, damage int32, typ object.DamageType,
			) bool {
				calls++
				if target != tc.target || source != nil || weapon != nil || damage != 2 || typ != tc.typ {
					t.Fatalf("default args = (%p, %p, %p, %d, %v)", target, source, weapon, damage, typ)
				}
				return true
			})
			if got != tc.want {
				t.Fatalf("result = %v, want %v", got, tc.want)
			}
			wantCalls := 0
			if tc.want {
				wantCalls = 1
			}
			if calls != wantCalls {
				t.Fatalf("default calls = %d, want %d", calls, wantCalls)
			}
		})
	}
}

func TestWeaponDamage4E14B0PropagatesDefaultResult(t *testing.T) {
	target := &Object{ObjClass: object.ClassWeapon, InvHolder: &Object{}}
	if WeaponDamage4E14B0(target, nil, nil, 1, object.DamageBlade, nil) {
		t.Fatal("nil default callback returned true")
	}
	if WeaponDamage4E14B0(target, nil, nil, 1, object.DamageBlade, func(
		_, _, _ *Object, _ int32, _ object.DamageType,
	) bool {
		return false
	}) {
		t.Fatal("false default result was not propagated")
	}
}

func TestArmorDamage4E1500OracleBranches(t *testing.T) {
	holder := &Object{}
	tests := []struct {
		name       string
		target     *Object
		damage     int32
		typ        object.DamageType
		want       bool
		wantDamage int32
	}{
		{name: "nil target", target: nil, damage: 2, typ: object.DamageLava},
		{name: "wrong class", target: &Object{ObjClass: object.ClassWeapon, InvHolder: holder}, damage: 2, typ: object.DamageCrush},
		{name: "loose armor non-lava", target: &Object{ObjClass: object.ClassArmor}, damage: 2, typ: object.DamageBlade},
		{name: "zero", target: &Object{ObjClass: object.ClassArmor, InvHolder: holder}, damage: 0, typ: object.DamageBlade},
		{name: "held non-metal crush", target: &Object{ObjClass: object.ClassArmor, InvHolder: holder}, damage: 7, typ: object.DamageCrush, want: true, wantDamage: 7},
		{name: "held metal blade", target: &Object{ObjClass: object.ClassArmor, Material: uint16(object.MaterialMetal), InvHolder: holder}, damage: 7, typ: object.DamageBlade, want: true, wantDamage: 7},
		{name: "held metal crush", target: &Object{ObjClass: object.ClassArmor, Material: uint16(object.MaterialMetal), InvHolder: holder}, damage: 7, typ: object.DamageCrush, want: true, wantDamage: 14},
		{name: "loose armor lava", target: &Object{ObjClass: object.ClassArmor}, damage: 2, typ: object.DamageLava, want: true, wantDamage: 2},
		{name: "crush wrap to zero", target: &Object{ObjClass: object.ClassArmor, Material: uint16(object.MaterialMetal), InvHolder: holder}, damage: math.MinInt32, typ: object.DamageCrush},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			source := &Object{}
			weapon := &Object{}
			calls := 0
			got := ArmorDamage4E1500(tc.target, source, weapon, tc.damage, tc.typ, func(
				target, gotSource, gotWeapon *Object, damage int32, typ object.DamageType,
			) bool {
				calls++
				if target != tc.target || gotSource != source || gotWeapon != weapon || damage != tc.wantDamage || typ != tc.typ {
					t.Fatalf("default args = (%p, %p, %p, %d, %v), want damage %d", target, gotSource, gotWeapon, damage, typ, tc.wantDamage)
				}
				return true
			})
			if got != tc.want {
				t.Fatalf("result = %v, want %v", got, tc.want)
			}
			wantCalls := 0
			if tc.want {
				wantCalls = 1
			}
			if calls != wantCalls {
				t.Fatalf("default calls = %d, want %d", calls, wantCalls)
			}
		})
	}
}

func TestArmorDamage4E1500PropagatesDefaultResult(t *testing.T) {
	target := &Object{ObjClass: object.ClassArmor, InvHolder: &Object{}}
	if ArmorDamage4E1500(target, nil, nil, 1, object.DamageBlade, nil) {
		t.Fatal("nil default callback returned true")
	}
	if ArmorDamage4E1500(target, nil, nil, 1, object.DamageBlade, func(
		_, _, _ *Object, _ int32, _ object.DamageType,
	) bool {
		return false
	}) {
		t.Fatal("false default result was not propagated")
	}
}

func TestItemDamage4E14B0OracleConstants(t *testing.T) {
	if got := uint32(object.ClassWeapon | object.ClassWand); got != 0x01001000 {
		t.Fatalf("weapon/wand class mask = %#x", got)
	}
	if got := uint32(object.ClassArmor); got != 0x02000000 {
		t.Fatalf("armor class mask = %#x", got)
	}
	if got := uint32(object.MaterialMetal); got != 0x10 {
		t.Fatalf("metal material mask = %#x", got)
	}
	if object.DamageCrush != 2 || object.DamageLava != 12 {
		t.Fatalf("damage IDs = crush %d, lava %d", object.DamageCrush, object.DamageLava)
	}
}
