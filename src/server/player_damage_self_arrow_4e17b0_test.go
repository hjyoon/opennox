package server

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
)

func selfArrowFixture4E17B0(t *testing.T, player bool) (*Object, *Object) {
	t.Helper()
	v := damageMeleeUnitFixture4E17B0(t, player)
	v.TypeInd, v.ObjFlags = 713, object.Flags(16777732)
	if player {
		v.ObjClass = object.Class(2621444)
	}
	w := &Object{TypeInd: 529, ObjClass: object.Class(85983233), ObjSubClass: 16, ObjFlags: object.Flags(16794116),
		PrevPos: types.Ptf(17, 9), PosVec: types.Ptf(19, 11), InitData: unsafe.Pointer(&ModifierInitData{})}
	return v, w
}

func TestPlayerDamageSelfArrow4E17B0ArmorAndDefaultHP(t *testing.T) {
	for _, player := range []bool{false, true} {
		for _, friendly := range []bool{false, true} {
			t.Run(fmt.Sprintf("player-%t/friendly-%t", player, friendly), func(t *testing.T) {
				v, w := selfArrowFixture4E17B0(t, player)
				armor := damageMeleeArmorFixture4E17B0(v, 0.25, 0)
				r := damageMeleeRuntimeFixture4E17B0(t)
				damageMeleeArmorRuntime4E17B0(&r, armor, 0.25)
				world := damageMeleeWorldRuntime4E0B30(t)
				world.IsEnemy = func(target, source *Object) bool {
					if target != v || source != w {
						t.Fatal("self-arrow identity lost")
					}
					return !friendly
				}
				r.DefaultDamage = func(v, a, w *Object, d int32, k object.DamageType) bool {
					return DefaultDamageWorld4E0B30(v, a, w, d, k, world)
				}
				before := *w
				for hit, hp := range []uint16{198, 196, 193} {
					if h, ok := PlayerDamageNative4E17B0(v, w, w, 3, object.DamageImpale, r); !h || !ok || v.HealthData.Cur != hp {
						t.Fatalf("self-arrow hit=%d handled/result=%t/%t HP=%d want=%d", hit, h, ok, v.HealthData.Cur, hp)
					}
					marker, kind, carry := damageMeleeMarker4E17B0(v)
					if marker != 2 || kind != 3 || carry != []float32{0.25, 0.5, -0.25}[hit] ||
						v.Obj130 != w || v.Pos132 != w.PrevPos || v.Field131 != 3 || *w != before {
						t.Fatalf("self-arrow marker=%d/%d carry=%g", marker, kind, carry)
					}
				}
				if armor.HealthData.Cur != 23 {
					t.Fatalf("armor HP=%d want=23", armor.HealthData.Cur)
				}
			})
		}
	}
}

func TestPlayerDamageSelfArrow4E17B0DefensesAndCoop(t *testing.T) {
	for _, player := range []bool{false, true} {
		for _, defense := range []string{"Reflect", "Shield", "GreatSword", "Invulnerable", "NoUpdate", "Dead", "CoopOwner"} {
			t.Run(fmt.Sprintf("player-%t/%s", player, defense), func(t *testing.T) {
				v, w := selfArrowFixture4E17B0(t, player)
				r := damageMeleeRuntimeFixture4E17B0(t)
				r.BlockDirection = func(*Object, types.Pointf) bool { return true }
				reflections, owners, sound, wear := 0, 0, 0, float32(0)
				r.ProjectileReflect = func(a, target *Object) {
					if a != w || target != v {
						t.Fatal("reflection identity")
					}
					reflections++
				}
				r.ClearOwner = func(a *Object) { owners++; a.ObjOwner = nil }
				r.SetOwner = func(owner, a *Object) { owners++; a.ObjOwner = owner }
				r.Audio = func(n int, _ *Object) { sound = n }
				r.Frame = func() uint32 { return 1400 }
				r.BlockDamagePercent = func() float64 { return 0.25 }
				r.CanDamageBlockItem = func(*Object) bool { return true }
				r.Melee.CanDamageBlockWeapon = r.CanDamageBlockItem
				blockWear := func(_, target, source, weapon *Object, amount float32, typ object.DamageType) bool {
					if target != v || source != w || weapon != w || typ != object.DamageImpale {
						t.Fatal("blocked arrow identity")
					}
					wear = amount
					return true
				}
				r.DamageBlockItem, r.Melee.DamageBlockWeapon = blockWear, blockWear
				r.Melee.RandomInt = func(int, int) int { return 19 }
				r.PlayerSetState = func(*Object, PlayerState) bool { return true }
				r.Melee.MonsterBlockAction = func(*Object) {}
				r.Melee.MonsterPopBlockAction = func(*Object) {}
				if defense == "Reflect" {
					v.Buffs = 1 << playerDamageReflectEnchant4E17B0
				} else if defense == "Invulnerable" {
					v.Buffs = 1 << playerDamageInvulnerableEnchant4E17B0
				} else if defense == "Dead" {
					v.ObjFlags |= object.FlagDead
				} else if defense == "NoUpdate" {
					v.ObjFlags |= object.FlagNoUpdate
				} else if defense == "CoopOwner" {
					w.ObjOwner = v
					r.CoopMode = func() bool { return true }
				} else {
					item := &Object{ObjClass: object.ClassArmor, ObjSubClass: 2, ObjFlags: object.FlagEquipped, HealthData: &HealthData{Cur: 25, Max: 25}}
					v.InvFirstItem = item
					if defense == "GreatSword" {
						item.ObjClass, item.ObjSubClass = object.ClassWeapon, 0x400
					}
					if player {
						u := v.UpdateDataPlayer()
						if defense == "Shield" {
							u.State, u.Player.ArmorEquip = PlayerState16, 0x1000000
						} else {
							u.Player.WeaponEquip = 0x400
						}
					} else {
						u := v.UpdateDataMonster()
						if defense == "Shield" {
							u.AIStack[0].Action, u.ArmorEquipFlags = uint32(ai.ACTION_BLOCK_ATTACK), 0x1000000
						} else {
							u.AIStack[0].Action, u.WeaponEquipFlags = uint32(ai.ACTION_GUARD), 0x400
						}
					}
				}
				calls := 0
				r.DefaultDamage = func(*Object, *Object, *Object, int32, object.DamageType) bool { calls++; return true }
				h, ok := PlayerDamageNative4E17B0(v, w, w, 3, object.DamageImpale, r)
				wantCall := defense == "CoopOwner" && !player
				wantResult := defense == "Invulnerable" || wantCall
				if !h || ok != wantResult || (calls != 0) != wantCall {
					t.Fatalf("defense handled/result=%t/%t calls=%d", h, ok, calls)
				}
				if defense == "Shield" && (sound != 878 || reflections != 0 || owners != 0 || wear != 0.75) {
					t.Fatalf("subclass-16 arrow shield sound=%d reflect=%d owner=%d wear=%g", sound, reflections, owners, wear)
				}
				if (defense == "Reflect" || defense == "GreatSword") && (reflections != 1 || owners != 2 || w.ObjOwner != v) {
					t.Fatal("self-arrow reflection/ownership lost")
				}
			})
		}
	}
}

func TestPlayerDamageSelfArrowShape4E17B0Boundaries(t *testing.T) {
	for _, tc := range []struct {
		name  string
		class object.Class
		sub   object.SubClass
		typ   object.DamageType
		want  bool
	}{
		{"stock", object.Class(85983233), 16, object.DamageImpale, true},
		{"different type ID", object.ClassMissile | object.ClassWeapon, 16, object.DamageImpale, true},
		{"weapon without missile", object.ClassWeapon, 16, object.DamageImpale, false},
		{"missile without weapon", object.ClassMissile, 16, object.DamageImpale, false},
		{"unit missile", object.ClassMissile | object.ClassWeapon | object.ClassMonster, 16, object.DamageImpale, false},
		{"wand missile", object.ClassMissile | object.ClassWeapon | object.ClassWand, 16, object.DamageImpale, false},
		{"melee qualifier", object.ClassMissile | object.ClassWeapon, 0, object.DamageImpale, false},
		{"other damage", object.Class(85983233), 16, object.DamageBlade, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := &Object{TypeInd: 9999, ObjClass: tc.class, ObjSubClass: tc.sub}
			if got := playerDamageWorldProjectileShape4E17B0(w, w, tc.typ); got != tc.want {
				t.Fatalf("shape=%t want=%t class=%x subclass=%x", got, tc.want, tc.class, tc.sub)
			}
		})
	}
	if playerDamageWorldProjectileShape4E17B0(nil, nil, object.DamageImpale) {
		t.Fatal("nil admitted")
	}
}
