package server

import (
	"image"
	"math"
	"reflect"
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestDefaultDamageWorld4E0B30SourceLessLava(t *testing.T) {
	defaultSound := unsafe.Pointer(new(byte))
	target := &Object{
		ObjClass:    object.ClassArmor,
		DamageSound: defaultSound,
		Pos132:      types.Pointf{X: 100, Y: -20},
	}
	var events []string
	runtime := DefaultDamageWorldRuntime4E0B30{
		Frame: func() uint32 { return 8 },
		Audio: func(id int, got *Object) {
			if id != 104 || got != target {
				t.Fatalf("Audio(%d,%p), want (104,%p)", id, got, target)
			}
			events = append(events, "fire-sound")
		},
		FireProtection: func(got *Object) float64 {
			if got != target {
				t.Fatalf("FireProtection(%p), want %p", got, target)
			}
			events = append(events, "fire-protection")
			return 0.25
		},
		BuffOff: func(got *Object, enchant EnchantID) {
			if got != target || enchant != defaultDamageInvisibleEnchant4E0B30 {
				t.Fatalf("BuffOff(%p,%d)", got, enchant)
			}
			events = append(events, "buff-off")
		},
		DefaultDamageSound: func(gotTarget, gotSource *Object) {
			if gotTarget != target || gotSource != nil {
				t.Fatalf("DefaultDamageSound(%p,%p), want (%p,nil)", gotTarget, gotSource, target)
			}
			events = append(events, "damage-sound")
		},
		DamageClear: func(got *Object, damage int32) {
			if got != target || damage != 4 {
				t.Fatalf("DamageClear(%p,%d), want (%p,4)", got, damage, target)
			}
			events = append(events, "damage")
		},
		DefaultDamageSoundC: defaultSound,
		Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
			t.Fatalf("source-less LAVA rejected: %s", reason)
		},
	}
	if !DefaultDamageWorld4E0B30(target, nil, nil, 6, object.DamageLava, runtime) {
		t.Fatal("source-less LAVA returned false")
	}
	wantEvents := []string{"fire-protection", "fire-sound", "buff-off", "damage-sound", "damage"}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	if target.Pos132 != (types.Pointf{}) || target.Obj130 != nil || target.Field131 != uint32(object.DamageLava) || target.Frame134 != 8 {
		t.Fatalf("target metadata = pos:%+v source:%p type:%d frame:%d", target.Pos132, target.Obj130, target.Field131, target.Frame134)
	}
}

func TestDefaultDamageWorld4E0B30LavaMinimumDamage(t *testing.T) {
	var gotDamage int32
	target := &Object{ObjClass: object.ClassArmor}
	runtime := DefaultDamageWorldRuntime4E0B30{
		FireProtection: func(*Object) float64 { return 0.6 },
		BuffOff:        func(*Object, EnchantID) {},
		DamageClear:    func(_ *Object, damage int32) { gotDamage = damage },
	}
	DefaultDamageWorld4E0B30(target, nil, nil, 1, object.DamageLava, runtime)
	if gotDamage != 1 {
		t.Fatalf("minimum LAVA damage = %d, want 1", gotDamage)
	}
}

func TestDefaultDamageWorld4E0B30LavaUsesBinary64BeforeFloatSpill(t *testing.T) {
	var gotDamage int32
	target := &Object{ObjClass: object.ClassArmor}
	protection := math.Float32frombits(0x3defaf0d)
	runtime := DefaultDamageWorldRuntime4E0B30{
		FireProtection: func(*Object) float64 { return float64(protection) },
		BuffOff:        func(*Object, EnchantID) {},
		DamageClear:    func(_ *Object, damage int32) { gotDamage = damage },
	}
	DefaultDamageWorld4E0B30(target, nil, nil, 5413, object.DamageLava, runtime)
	if gotDamage != 4780 {
		t.Fatalf("binary64 LAVA scaling = %d, want 4780", gotDamage)
	}
}

func TestDefaultDamageWorld4E0B30NonUnitExplosion(t *testing.T) {
	for _, tc := range []struct {
		name       string
		withCaster bool
	}{
		{name: "radial missile source"},
		{name: "caster and missile weapon", withCaster: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target := &Object{
				ObjClass:   object.ClassObstacle,
				HealthData: &HealthData{Cur: 20, Max: 20},
			}
			missile := &Object{
				ObjClass: object.ClassMissile,
				PrevPos:  types.Pointf{X: 312, Y: 478},
			}
			source, weapon := missile, (*Object)(nil)
			if tc.withCaster {
				source = &Object{ObjClass: object.ClassPlayer, PrevPos: types.Pointf{X: 8, Y: 9}}
				weapon = missile
			}

			var events []string
			runtime := DefaultDamageWorldRuntime4E0B30{
				Frame:         func() uint32 { return 12 },
				GameplayFlag1: func() bool { return true },
				FireProtection: func(got *Object) float64 {
					if got != target {
						t.Fatalf("FireProtection(%p), want %p", got, target)
					}
					events = append(events, "fire-protection")
					return 0.25
				},
				Audio: func(id int, got *Object) {
					if id != 104 || got != target {
						t.Fatalf("Audio(%d,%p), want (104,%p)", id, got, target)
					}
					events = append(events, "fire-sound")
				},
				BuffOff: func(got *Object, enchant EnchantID) {
					if got != target || enchant != defaultDamageInvisibleEnchant4E0B30 {
						t.Fatalf("BuffOff(%p,%d)", got, enchant)
					}
					events = append(events, "buff-off")
				},
				DefaultDamageSound: func(gotTarget, gotSource *Object) {
					if gotTarget != target || gotSource != missile {
						t.Fatalf("DefaultDamageSound(%p,%p), want (%p,%p)", gotTarget, gotSource, target, missile)
					}
					events = append(events, "damage-sound")
				},
				DamageClear: func(got *Object, damage int32) {
					if got != target || damage != 6 {
						t.Fatalf("DamageClear(%p,%d), want (%p,6)", got, damage, target)
					}
					target.HealthData.Cur -= uint16(damage)
					events = append(events, "damage")
				},
				Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
					t.Fatalf("non-unit explosion rejected: %s", reason)
				},
			}
			if !DefaultDamageWorld4E0B30(target, source, weapon, 8, object.DamageExplosion, runtime) {
				t.Fatal("non-unit explosion returned false")
			}
			wantEvents := []string{"fire-protection", "fire-sound", "buff-off", "damage-sound", "damage"}
			if !reflect.DeepEqual(events, wantEvents) {
				t.Fatalf("events = %v, want %v", events, wantEvents)
			}
			if target.HealthData.Cur != 14 || target.Pos132 != missile.PrevPos || target.Obj130 != missile ||
				target.Field131 != uint32(object.DamageExplosion) || target.Frame134 != 12 {
				t.Fatalf("target state = health:%d pos:%+v source:%p type:%d frame:%d",
					target.HealthData.Cur, target.Pos132, target.Obj130, target.Field131, target.Frame134)
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30MissileImpactDoor(t *testing.T) {
	target := &Object{
		ObjClass:   object.ClassDoor | object.ClassClientPersist,
		HealthData: &HealthData{Cur: 6, Max: 6},
	}
	source := &Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(&MonsterUpdateData{})}
	missile := &Object{ObjClass: object.ClassMissile, PrevPos: types.Pointf{X: 112, Y: 208}}
	var events []string
	runtime := DefaultDamageWorldRuntime4E0B30{
		Frame:         func() uint32 { return 642 },
		GameplayFlag1: func() bool { return true },
		IsEnemy: func(gotTarget, gotSource *Object) bool {
			if gotTarget != target || gotSource != source {
				t.Fatalf("IsEnemy(%p, %p)", gotTarget, gotSource)
			}
			return false
		},
		BuffOff: func(got *Object, enchant EnchantID) {
			if got != target || enchant != defaultDamageInvisibleEnchant4E0B30 {
				t.Fatalf("BuffOff(%p, %d)", got, enchant)
			}
			events = append(events, "buff-off")
		},
		MonsterHasHitSound: func(got *Object) bool {
			if got != source {
				t.Fatalf("MonsterHasHitSound(%p)", got)
			}
			events = append(events, "hit-sound")
			return false
		},
		DefaultDamageSound: func(gotTarget, gotSource *Object) {
			if gotTarget != target || gotSource != missile {
				t.Fatalf("DefaultDamageSound(%p, %p)", gotTarget, gotSource)
			}
			events = append(events, "damage-sound")
		},
		DamageClear: func(got *Object, damage int32) {
			if got != target || damage != 1 {
				t.Fatalf("DamageClear(%p, %d)", got, damage)
			}
			target.HealthData.Cur -= uint16(damage)
			events = append(events, "damage")
		},
		Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
			t.Fatalf("ordinary door impact rejected: %s", reason)
		},
	}
	if !DefaultDamageWorld4E0B30(target, source, missile, 1, object.DamageImpact, runtime) {
		t.Fatal("door impact returned false")
	}
	if target.HealthData.Cur != 5 || target.Pos132 != missile.PrevPos || target.Obj130 != missile ||
		target.Field131 != uint32(object.DamageImpact) || target.Frame134 != 642 {
		t.Fatalf("door state = health:%d pos:%+v source:%p type:%d frame:%d",
			target.HealthData.Cur, target.Pos132, target.Obj130, target.Field131, target.Frame134)
	}
	want := []string{"buff-off", "hit-sound", "damage-sound", "damage"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestDefaultDamageWorld4E0B30BladePointerFields(t *testing.T) {
	defaultSoundMarker := uint32(0x4e0b30)
	defaultSound := unsafe.Pointer(&defaultSoundMarker)
	attrs := &ModifierInitData{}
	target := &Object{ObjClass: object.ClassObstacle, DamageSound: defaultSound}
	source := &Object{ObjClass: object.ClassPlayer, PrevPos: types.Pointf{X: 101.25, Y: -73.5}}
	weapon := &Object{
		ObjClass: object.ClassWeapon,
		PrevPos:  types.Pointf{X: -1, Y: -2},
		InitData: unsafe.Pointer(attrs),
	}
	var events []string
	runtime := DefaultDamageWorldRuntime4E0B30{
		Frame:         func() uint32 { return 0x12345678 },
		GameplayFlag1: func() bool { return true },
		QuestMode:     func() bool { return true },
		BuffOff: func(got *Object, enchant EnchantID) {
			if got != target || enchant != defaultDamageInvisibleEnchant4E0B30 {
				t.Fatalf("BuffOff(%p, %d)", got, enchant)
			}
			events = append(events, "buff-off")
		},
		DefaultDamageSound: func(gotTarget, gotSource *Object) {
			if gotTarget != target || gotSource != weapon {
				t.Fatalf("DefaultDamageSound(%p, %p)", gotTarget, gotSource)
			}
			events = append(events, "sound")
		},
		DamageClear: func(got *Object, gotDamage int32) {
			if got != target || gotDamage != 17 {
				t.Fatalf("DamageClear(%p, %d)", got, gotDamage)
			}
			events = append(events, "damage")
		},
		DefaultDamageSoundC: defaultSound,
		Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
			t.Fatalf("unexpected unsupported branch: %s", reason)
		},
	}

	if !DefaultDamageWorld4E0B30(target, source, weapon, 17, object.DamageBlade, runtime) {
		t.Fatal("DefaultDamageWorld4E0B30 returned false")
	}
	if target.Pos132 != source.PrevPos {
		t.Fatalf("hit position = %+v, want source previous position %+v", target.Pos132, source.PrevPos)
	}
	if target.Obj130 != weapon {
		t.Fatalf("attribution pointer = %p, want full weapon pointer %p", target.Obj130, weapon)
	}
	if target.Field131 != uint32(object.DamageBlade) || target.Frame134 != 0x12345678 {
		t.Fatalf("damage metadata = type %d frame %#x", target.Field131, target.Frame134)
	}
	want := []string{"buff-off", "sound", "damage"}
	if len(events) != len(want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	for i := range want {
		if events[i] != want[i] {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}
}

func TestDefaultDamageWorld4E0B30DeadZombieAttribution(t *testing.T) {
	target := &Object{ObjClass: object.ClassObstacle, ObjFlags: object.FlagDead}
	source := &Object{}
	weapon := &Object{}
	damaged := false
	runtime := DefaultDamageWorldRuntime4E0B30{
		Frame:    func() uint32 { return 73 },
		IsZombie: func(*Object) bool { return true },
		DamageClear: func(*Object, int32) {
			damaged = true
		},
	}
	if !DefaultDamageWorld4E0B30(target, source, weapon, 9, object.DamageBlade, runtime) {
		t.Fatal("dead zombie branch returned false")
	}
	if target.Obj130 != weapon || target.Field131 != 0 || target.Frame134 != 73 {
		t.Fatalf("dead attribution = %p/%d/%d", target.Obj130, target.Field131, target.Frame134)
	}
	if damaged {
		t.Fatal("dead zombie attribution must not apply damage")
	}
}

func TestDefaultDamageWorld4E0B30InvulnerableCadence(t *testing.T) {
	target := &Object{Buffs: uint32(1) << defaultDamageInvulnerableEnchant4E0B30}
	for _, tc := range []struct {
		frame uint32
		want  int
	}{{frame: 8, want: 1}, {frame: 9, want: 0}} {
		calls := 0
		runtime := DefaultDamageWorldRuntime4E0B30{
			Frame: func() uint32 { return tc.frame },
			Audio: func(id int, got *Object) {
				if id != defaultDamageInvulnerableSound4E0B30 || got != target {
					t.Fatalf("Audio(%d, %p)", id, got)
				}
				calls++
			},
		}
		DefaultDamageWorld4E0B30(target, nil, nil, 1, object.DamageBlade, runtime)
		if calls != tc.want {
			t.Fatalf("frame %d audio calls = %d, want %d", tc.frame, calls, tc.want)
		}
	}
}

func TestDefaultDamageWorld4E0B30RejectsModifiedWeaponBeforeMutation(t *testing.T) {
	callbackMarker := uint32(0x4e13b0)
	modifier := &ModifierEff{AttackPreDmg64: ModifierEffFnc{Fnc: unsafe.Pointer(&callbackMarker)}}
	attrs := &ModifierInitData{Modifiers: [4]*ModifierEff{nil, modifier}}
	target := &Object{ObjClass: object.ClassObstacle}
	source := &Object{ObjClass: object.ClassPlayer, PrevPos: types.Pointf{X: 7, Y: 8}}
	weapon := &Object{ObjClass: object.ClassWeapon, InitData: unsafe.Pointer(attrs)}
	var reason string
	runtime := DefaultDamageWorldRuntime4E0B30{
		GameplayFlag1: func() bool { return true },
		CanApplyPreDamage: func(got *ModifierEff) bool {
			if got != modifier {
				t.Fatalf("pre-damage modifier = %p, want %p", got, modifier)
			}
			return false
		},
		ApplyPreDamage: func(*ModifierEff, *Object, *Object, *Object, *int32) {
			t.Fatal("unsupported modifier callback was applied")
		},
		Unsupported: func(got string, _, _, _ *Object, _ int32, _ object.DamageType) {
			reason = got
		},
		DamageClear: func(*Object, int32) {
			t.Fatal("unsupported modifier branch applied damage")
		},
	}
	if !DefaultDamageWorld4E0B30(target, source, weapon, 5, object.DamageBlade, runtime) {
		t.Fatal("unsupported branch must preserve the original success convention")
	}
	if reason != "weapon pre-damage modifiers" {
		t.Fatalf("unsupported reason = %q", reason)
	}
	if target.Obj130 != nil || target.Pos132 != (types.Pointf{}) || target.Frame134 != 0 {
		t.Fatalf("unsupported branch mutated target: attribution=%p pos=%+v frame=%d", target.Obj130, target.Pos132, target.Frame134)
	}
}

func TestDefaultDamageWorld4E0B30AppliesWeaponPreDamageInSlotOrder(t *testing.T) {
	firstMarker := uint32(1)
	secondMarker := uint32(2)
	first := &ModifierEff{AttackPreDmg64: ModifierEffFnc{Fnc: unsafe.Pointer(&firstMarker)}}
	second := &ModifierEff{AttackPreDmg64: ModifierEffFnc{Fnc: unsafe.Pointer(&secondMarker)}}
	attrs := &ModifierInitData{Modifiers: [4]*ModifierEff{nil, first, nil, second}}
	target := &Object{ObjClass: object.ClassObstacle}
	source := &Object{ObjClass: object.ClassPlayer, PrevPos: types.Pointf{X: 7, Y: 8}}
	weapon := &Object{ObjClass: object.ClassWeapon, InitData: unsafe.Pointer(attrs)}
	var events []string
	runtime := DefaultDamageWorldRuntime4E0B30{
		Frame:         func() uint32 { return 123 },
		GameplayFlag1: func() bool { return true },
		CanApplyPreDamage: func(modifier *ModifierEff) bool {
			if target.Obj130 != nil || target.Pos132 != (types.Pointf{}) || target.Frame134 != 0 {
				t.Fatalf("preflight observed mutated target: attribution=%p pos=%+v frame=%d",
					target.Obj130, target.Pos132, target.Frame134)
			}
			switch modifier {
			case first:
				events = append(events, "can:first")
			case second:
				events = append(events, "can:second")
			default:
				t.Fatalf("unexpected modifier %p", modifier)
			}
			return true
		},
		ApplyPreDamage: func(modifier *ModifierEff, gotWeapon, gotSource, gotTarget *Object, damage *int32) {
			if gotWeapon != weapon || gotSource != source || gotTarget != target {
				t.Fatalf("pre-damage objects = %p/%p/%p, want %p/%p/%p",
					gotWeapon, gotSource, gotTarget, weapon, source, target)
			}
			if target.Obj130 != weapon || target.Pos132 != source.PrevPos ||
				target.Field131 != uint32(object.DamageBlade) || target.Frame134 != 123 {
				t.Fatalf("callback attribution = %p/%+v/%d/%d",
					target.Obj130, target.Pos132, target.Field131, target.Frame134)
			}
			switch modifier {
			case first:
				events = append(events, "apply:first")
				*damage += 2
			case second:
				events = append(events, "apply:second")
				*damage *= 3
			default:
				t.Fatalf("unexpected modifier %p", modifier)
			}
		},
		DefaultDamageSound: func(gotTarget, gotSource *Object) {
			if gotTarget != target || gotSource != weapon {
				t.Fatalf("damage sound objects = %p/%p, want %p/%p", gotTarget, gotSource, target, weapon)
			}
			events = append(events, "sound")
		},
		DamageClear: func(gotTarget *Object, damage int32) {
			if gotTarget != target || damage != 21 {
				t.Fatalf("DamageClear(%p, %d), want (%p, 21)", gotTarget, damage, target)
			}
			events = append(events, "damage")
		},
		Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
			t.Fatalf("supported pre-damage branch rejected: %s", reason)
		},
	}
	if !DefaultDamageWorld4E0B30(target, source, weapon, 5, object.DamageBlade, runtime) {
		t.Fatal("DefaultDamageWorld4E0B30 returned false")
	}
	want := []string{"can:first", "can:second", "apply:first", "apply:second", "sound", "damage"}
	if !slices.Equal(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestDefaultDamageWorld4E0B30AllowsInertWeaponModifiers(t *testing.T) {
	modifier := &ModifierEff{}
	attrs := &ModifierInitData{Modifiers: [4]*ModifierEff{modifier}}
	target := &Object{ObjClass: object.ClassObstacle}
	source := &Object{ObjClass: object.ClassPlayer}
	weapon := &Object{ObjClass: object.ClassWeapon, InitData: unsafe.Pointer(attrs)}
	damaged := false
	runtime := DefaultDamageWorldRuntime4E0B30{
		GameplayFlag1: func() bool { return true },
		DamageClear: func(*Object, int32) {
			damaged = true
		},
		Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
			t.Fatalf("inert modifier rejected: %s", reason)
		},
	}
	DefaultDamageWorld4E0B30(target, source, weapon, 5, object.DamageBlade, runtime)
	if !damaged {
		t.Fatal("inert modifier prevented base damage")
	}
}

func TestDefaultDamageWorld4E0B30OrdinaryMonsterBlade(t *testing.T) {
	update := &MonsterUpdateData{Field547: 99}
	target := &Object{
		ObjClass:    object.ClassMonster,
		ObjSubClass: 0x202,
		HealthData:  &HealthData{Cur: 12, Max: 12},
		UpdateData:  unsafe.Pointer(update),
	}
	source := &Object{ObjClass: object.ClassPlayer, PrevPos: types.Pointf{X: 40, Y: 50}}
	weapon := &Object{ObjClass: object.ClassWeapon, InitData: unsafe.Pointer(&ModifierInitData{})}
	var events []string
	runtime := DefaultDamageWorldRuntime4E0B30{
		Frame:         func() uint32 { return 700 },
		GameplayFlag1: func() bool { return true },
		IsEnemy: func(gotTarget, gotSource *Object) bool {
			return gotTarget == target && gotSource == source
		},
		BuffOff: func(got *Object, enchant EnchantID) {
			if got != target || enchant != defaultDamageInvisibleEnchant4E0B30 {
				t.Fatalf("BuffOff(%p, %d)", got, enchant)
			}
			events = append(events, "buff-off")
		},
		DefaultDamageSound: func(gotTarget, gotSource *Object) {
			if gotTarget != target || gotSource != weapon {
				t.Fatalf("DefaultDamageSound(%p, %p)", gotTarget, gotSource)
			}
			events = append(events, "sound")
		},
		AdjustFieldGuide: func(gotSource, gotTarget *Object, gotDamage int32) int32 {
			if gotSource != source || gotTarget != target || gotDamage != 5 {
				t.Fatalf("AdjustFieldGuide(%p, %p, %d)", gotSource, gotTarget, gotDamage)
			}
			events = append(events, "field-guide")
			return 7
		},
		DamageClear: func(gotTarget *Object, gotDamage int32) {
			if gotTarget != target || gotDamage != 7 {
				t.Fatalf("DamageClear(%p, %d)", gotTarget, gotDamage)
			}
			events = append(events, "damage")
		},
		Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
			t.Fatalf("ordinary monster branch rejected: %s", reason)
		},
	}

	if !DefaultDamageWorld4E0B30(target, source, weapon, 5, object.DamageBlade, runtime) {
		t.Fatal("ordinary monster branch returned false")
	}
	if target.Pos132 != source.PrevPos || target.Obj130 != weapon || target.Field131 != 0 || target.Frame134 != 700 {
		t.Fatalf("target metadata = pos:%+v source:%p type:%d frame:%d", target.Pos132, target.Obj130, target.Field131, target.Frame134)
	}
	if !update.StatusFlags.Has(object.MonStatusInjured) || update.Field546 != 0 || update.Field547 != 2 {
		t.Fatalf("monster hit state = status:%#x field546:%d field547:%d", update.StatusFlags, update.Field546, update.Field547)
	}
	want := []string{"buff-off", "sound", "field-guide", "damage"}
	if len(events) != len(want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	for i := range want {
		if events[i] != want[i] {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}
}

func TestDefaultDamageWorld4E0B30PlayerElectricMonster(t *testing.T) {
	for _, typ := range []object.DamageType{object.DamageElectric, object.DamageAirborneElectric} {
		t.Run(typ.String(), func(t *testing.T) {
			update := &MonsterUpdateData{}
			target := &Object{
				ObjClass: object.ClassMonster, ObjSubClass: 0x202,
				HealthData: &HealthData{Cur: 12, Max: 12}, UpdateData: unsafe.Pointer(update),
			}
			source := &Object{ObjClass: object.ClassPlayer, PrevPos: types.Pointf{X: 40, Y: 50}}
			sound, protection, damaged := 0, 0, 0
			runtime := DefaultDamageWorldRuntime4E0B30{
				Frame:         func() uint32 { return 700 },
				GameplayFlag1: func() bool { return true },
				IsEnemy: func(gotTarget, gotSource *Object) bool {
					return gotTarget == target && gotSource == source
				},
				ElectricProtection: func(got *Object) float64 {
					if got != target {
						t.Fatalf("ElectricProtection(%p), want %p", got, target)
					}
					protection++
					return 0.25
				},
				Audio: func(id int, got *Object) {
					if id != 108 || got != target {
						t.Fatalf("Audio(%d,%p), want (108,%p)", id, got, target)
					}
					sound++
				},
				DamageClear: func(got *Object, damage int32) {
					if got != target || damage != 6 {
						t.Fatalf("DamageClear(%p,%d), want (%p,6)", got, damage, target)
					}
					damaged++
				},
				Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
					t.Fatalf("electric branch rejected: %s", reason)
				},
			}
			if !DefaultDamageWorld4E0B30(target, source, nil, 8, typ, runtime) {
				t.Fatal("electric branch returned false")
			}
			if sound != 1 || protection != 1 || damaged != 1 || target.Obj130 != source ||
				target.Field131 != uint32(typ) || target.Frame134 != 700 || target.Pos132 != source.PrevPos ||
				!update.StatusFlags.Has(object.MonStatusInjured) {
				t.Fatalf("electric state: sound=%d protection=%d damage=%d source=%p type=%d frame=%d pos=%+v status=%#x",
					sound, protection, damaged, target.Obj130, target.Field131, target.Frame134, target.Pos132, update.StatusFlags)
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30SourceLessElectricMonster(t *testing.T) {
	for _, typ := range []object.DamageType{object.DamageElectric, object.DamageAirborneElectric} {
		t.Run(typ.String(), func(t *testing.T) {
			update := &MonsterUpdateData{}
			target := &Object{
				ObjClass: object.ClassMonster, ObjSubClass: 0x70001,
				HealthData: &HealthData{Cur: 8, Max: 8}, UpdateData: unsafe.Pointer(update),
				Pos132: types.Pointf{X: 31, Y: 47},
			}
			protection, sound, damage := 0, 0, 0
			runtime := DefaultDamageWorldRuntime4E0B30{
				Frame: func() uint32 { return 1200 },
				IsEnemy: func(*Object, *Object) bool {
					t.Fatal("source-less electric damage tested enemy relation")
					return false
				},
				BuffOff: func(*Object, EnchantID) {
					t.Fatal("source-less electric damage removed invisibility")
				},
				ElectricProtection: func(got *Object) float64 {
					if got != target {
						t.Fatalf("ElectricProtection(%p), want %p", got, target)
					}
					protection++
					return 0.25
				},
				Audio: func(id int, got *Object) {
					if id != 108 || got != target {
						t.Fatalf("Audio(%d,%p), want (108,%p)", id, got, target)
					}
					sound++
				},
				DefaultDamageSound: func(gotTarget, gotSource *Object) {
					if gotTarget != target || gotSource != nil {
						t.Fatalf("DefaultDamageSound(%p,%p)", gotTarget, gotSource)
					}
				},
				DamageClear: func(got *Object, amount int32) {
					if got != target || amount != 38 {
						t.Fatalf("DamageClear(%p,%d), want (%p,38)", got, amount, target)
					}
					damage++
				},
				Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
					t.Fatalf("source-less electric branch rejected: %s", reason)
				},
			}
			if !DefaultDamageWorld4E0B30(target, nil, nil, 50, typ, runtime) {
				t.Fatal("source-less electric branch returned false")
			}
			if protection != 1 || sound != 1 || damage != 1 || target.Obj130 != nil ||
				target.Pos132 != (types.Pointf{}) || target.Field131 != uint32(typ) || target.Frame134 != 1200 ||
				!update.StatusFlags.Has(object.MonStatusInjured) || update.Field546 != uint32(typ) || update.Field547 != 2 {
				t.Fatalf("source-less electric state: protection=%d sound=%d damage=%d source=%p pos=%+v type=%d frame=%d status=%#x field546=%d field547=%d",
					protection, sound, damage, target.Obj130, target.Pos132, target.Field131, target.Frame134,
					update.StatusFlags, update.Field546, update.Field547)
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30MonsterElectricMonster(t *testing.T) {
	for _, typ := range []object.DamageType{object.DamageElectric, object.DamageAirborneElectric} {
		t.Run(typ.String(), func(t *testing.T) {
			targetUpdate := &MonsterUpdateData{}
			target := &Object{
				ObjClass: object.ClassMonster, ObjSubClass: 0x70001,
				HealthData: &HealthData{Cur: 8, Max: 8}, UpdateData: unsafe.Pointer(targetUpdate),
			}
			source := &Object{
				ObjClass: object.ClassMonster, PrevPos: types.Pointf{X: 2365, Y: 3365},
				UpdateData: unsafe.Pointer(&MonsterUpdateData{}),
			}
			protection, damaged := 0, 0
			runtime := DefaultDamageWorldRuntime4E0B30{
				Frame:         func() uint32 { return 1200 },
				GameplayFlag1: func() bool { return true },
				IsEnemy: func(gotTarget, gotSource *Object) bool {
					return gotTarget == target && gotSource == source
				},
				ElectricProtection: func(got *Object) float64 {
					if got != target {
						t.Fatalf("ElectricProtection(%p), want %p", got, target)
					}
					protection++
					return 0.25
				},
				DamageClear: func(got *Object, damage int32) {
					if got != target || damage != 6 {
						t.Fatalf("DamageClear(%p,%d), want (%p,6)", got, damage, target)
					}
					damaged++
				},
				Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
					t.Fatalf("monster electricity rejected: %s", reason)
				},
			}
			if !DefaultDamageWorld4E0B30(target, source, nil, 8, typ, runtime) {
				t.Fatal("monster electricity returned false")
			}
			if protection != 1 || damaged != 1 || target.Obj130 != source ||
				target.Field131 != uint32(typ) || target.Frame134 != 1200 ||
				target.Pos132 != source.PrevPos || !targetUpdate.StatusFlags.Has(object.MonStatusInjured) {
				t.Fatalf("monster electricity state: protection=%d damage=%d source=%p type=%d frame=%d pos=%+v status=%#x",
					protection, damaged, target.Obj130, target.Field131, target.Frame134, target.Pos132, targetUpdate.StatusFlags)
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30SelfSourcedMissileImpactMonster(t *testing.T) {
	targetUpdate := &MonsterUpdateData{Field547: 99}
	target := &Object{
		ObjClass:    object.ClassMonster,
		ObjSubClass: 0x10002,
		HealthData:  &HealthData{Cur: 30, Max: 30},
		UpdateData:  unsafe.Pointer(targetUpdate),
	}
	missile := &Object{ObjClass: object.ClassMissile, PrevPos: types.Pointf{X: 352, Y: 788}}
	var events []string
	runtime := DefaultDamageWorldRuntime4E0B30{
		Frame:         func() uint32 { return 1240 },
		GameplayFlag1: func() bool { return true },
		IsEnemy: func(*Object, *Object) bool {
			t.Fatal("missile impact checked unit allegiance")
			return false
		},
		BuffOff: func(got *Object, enchant EnchantID) {
			if got != target || enchant != defaultDamageInvisibleEnchant4E0B30 {
				t.Fatalf("BuffOff(%p, %d)", got, enchant)
			}
			events = append(events, "buff-off")
		},
		MonsterHasHitSound: func(*Object) bool {
			t.Fatal("missile used monster hit sound")
			return false
		},
		DefaultDamageSound: func(gotTarget, gotSource *Object) {
			if gotTarget != target || gotSource != missile {
				t.Fatalf("DefaultDamageSound(%p, %p)", gotTarget, gotSource)
			}
			events = append(events, "sound")
		},
		AdjustFieldGuide: func(gotSource, gotTarget *Object, damage int32) int32 {
			if gotSource != missile || gotTarget != target || damage != 1 {
				t.Fatalf("AdjustFieldGuide(%p, %p, %d)", gotSource, gotTarget, damage)
			}
			events = append(events, "field-guide")
			return damage
		},
		DamageClear: func(gotTarget *Object, damage int32) {
			if gotTarget != target || damage != 1 {
				t.Fatalf("DamageClear(%p, %d)", gotTarget, damage)
			}
			target.HealthData.Cur -= uint16(damage)
			events = append(events, "damage")
		},
		Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
			t.Fatalf("ordinary monster missile impact rejected: %s", reason)
		},
	}

	if !DefaultDamageWorld4E0B30(target, missile, missile, 1, object.DamageImpact, runtime) {
		t.Fatal("monster missile impact returned false")
	}
	if target.HealthData.Cur != 29 || target.Pos132 != missile.PrevPos || target.Obj130 != missile ||
		target.Field131 != uint32(object.DamageImpact) || target.Frame134 != 1240 {
		t.Fatalf("target state = health:%d pos:%+v source:%p type:%d frame:%d",
			target.HealthData.Cur, target.Pos132, target.Obj130, target.Field131, target.Frame134)
	}
	if !targetUpdate.StatusFlags.Has(object.MonStatusInjured) ||
		targetUpdate.Field546 != uint32(object.DamageImpact) || targetUpdate.Field547 != 2 {
		t.Fatalf("monster hit state = status:%#x type:%d latch:%d",
			targetUpdate.StatusFlags, targetUpdate.Field546, targetUpdate.Field547)
	}
	want := []string{"buff-off", "sound", "field-guide", "damage"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestDefaultDamageWorld4E0B30MonsterFiredMissileImpactMonster(t *testing.T) {
	targetUpdate := &MonsterUpdateData{}
	target := &Object{
		ObjClass:    object.ClassMonster,
		ObjSubClass: 0x10002,
		HealthData:  &HealthData{Cur: 30, Max: 30},
		UpdateData:  unsafe.Pointer(targetUpdate),
	}
	sourceUpdate := &MonsterUpdateData{}
	source := &Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(sourceUpdate)}
	missile := &Object{ObjClass: object.ClassMissile, PrevPos: types.Pointf{X: 419, Y: 795}}
	enemyChecks, hitSounds, damaged := 0, 0, 0
	runtime := DefaultDamageWorldRuntime4E0B30{
		Frame:         func() uint32 { return 1241 },
		GameplayFlag1: func() bool { return true },
		IsEnemy: func(gotTarget, gotSource *Object) bool {
			if gotTarget != target || gotSource != source {
				t.Fatalf("IsEnemy(%p, %p)", gotTarget, gotSource)
			}
			enemyChecks++
			return true
		},
		MonsterHasHitSound: func(got *Object) bool {
			if got != source {
				t.Fatalf("MonsterHasHitSound(%p)", got)
			}
			hitSounds++
			return true
		},
		DefaultDamageSound: func(*Object, *Object) {
			t.Fatal("monster hit sound did not suppress target damage sound")
		},
		DamageClear: func(gotTarget *Object, damage int32) {
			if gotTarget != target || damage != 1 {
				t.Fatalf("DamageClear(%p, %d)", gotTarget, damage)
			}
			target.HealthData.Cur -= uint16(damage)
			damaged++
		},
		Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
			t.Fatalf("monster-fired missile impact rejected: %s", reason)
		},
	}
	if !DefaultDamageWorld4E0B30(target, source, missile, 1, object.DamageImpact, runtime) {
		t.Fatal("monster-fired missile impact returned false")
	}
	if target.HealthData.Cur != 29 || target.Pos132 != missile.PrevPos || target.Obj130 != missile ||
		target.Field131 != uint32(object.DamageImpact) || target.Frame134 != 1241 ||
		!targetUpdate.StatusFlags.Has(object.MonStatusInjured) || targetUpdate.Field546 != uint32(object.DamageImpact) ||
		targetUpdate.Field547 != 2 || sourceUpdate.Field130 != 1241 ||
		enemyChecks != 1 || hitSounds != 1 || damaged != 1 {
		t.Fatalf("missile impact state = health:%d pos:%+v source:%p type:%d frame:%d status:%#x hit type:%d latch:%d source frame:%d enemy checks:%d hit sounds:%d damage calls:%d",
			target.HealthData.Cur, target.Pos132, target.Obj130, target.Field131, target.Frame134,
			targetUpdate.StatusFlags, targetUpdate.Field546, targetUpdate.Field547, sourceUpdate.Field130,
			enemyChecks, hitSounds, damaged)
	}
}

func TestDefaultDamageWorld4E0B30RejectsOtherMonsterMissileDamage(t *testing.T) {
	target := &Object{
		ObjClass:   object.ClassMonster,
		HealthData: &HealthData{Cur: 30, Max: 30},
		UpdateData: unsafe.Pointer(&MonsterUpdateData{}),
	}
	missile := &Object{ObjClass: object.ClassMissile}
	var reason string
	DefaultDamageWorld4E0B30(target, missile, missile, 3, object.DamageExplosion, DefaultDamageWorldRuntime4E0B30{
		GameplayFlag1: func() bool { return true },
		DamageClear:   func(*Object, int32) { t.Fatal("unsupported missile explosion dealt damage") },
		Unsupported: func(got string, _, _, _ *Object, _ int32, _ object.DamageType) {
			reason = got
		},
	})
	if reason != "unsupported monster damage shape" || target.HealthData.Cur != 30 {
		t.Fatalf("unsupported branch = %q, target health = %d", reason, target.HealthData.Cur)
	}
}

func TestDefaultDamageWorld4E0B30MagicMissileExplosionMonster(t *testing.T) {
	player := &Object{ObjClass: object.ClassPlayer}
	missile := &Object{
		ObjClass: object.ClassMissile,
		PrevPos:  types.Pointf{X: 731, Y: 449},
		ObjOwner: player,
	}
	tests := []struct {
		name       string
		source     *Object
		weapon     *Object
		damage     int32
		wantSource *Object
	}{
		{
			name:       "direct",
			source:     player,
			weapon:     missile,
			damage:     8,
			wantSource: missile,
		},
		{
			name:       "splash",
			source:     missile,
			weapon:     nil,
			damage:     5,
			wantSource: missile,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			update := &MonsterUpdateData{Field547: 99}
			target := &Object{
				ObjClass:    object.ClassMonster,
				ObjSubClass: 0x70001,
				HealthData:  &HealthData{Cur: 40, Max: 40},
				UpdateData:  unsafe.Pointer(update),
			}
			var events []string
			runtime := DefaultDamageWorldRuntime4E0B30{
				Frame:         func() uint32 { return 1300 },
				GameplayFlag1: func() bool { return true },
				IsEnemy: func(*Object, *Object) bool {
					t.Fatal("Magic Missile explosion checked unit allegiance")
					return false
				},
				FireProtection: func(got *Object) float64 {
					if got != target {
						t.Fatalf("FireProtection(%p)", got)
					}
					events = append(events, "fire-protection")
					return 0
				},
				BuffOff: func(got *Object, enchant EnchantID) {
					if got != target || enchant != defaultDamageInvisibleEnchant4E0B30 {
						t.Fatalf("BuffOff(%p, %d)", got, enchant)
					}
					events = append(events, "buff-off")
				},
				DefaultDamageSound: func(gotTarget, gotSource *Object) {
					if gotTarget != target || gotSource != tc.wantSource {
						t.Fatalf("DefaultDamageSound(%p, %p)", gotTarget, gotSource)
					}
					events = append(events, "sound")
				},
				AdjustFieldGuide: func(gotSource, gotTarget *Object, damage int32) int32 {
					if gotSource != tc.source || gotTarget != target || damage != tc.damage {
						t.Fatalf("AdjustFieldGuide(%p, %p, %d)", gotSource, gotTarget, damage)
					}
					events = append(events, "field-guide")
					return damage
				},
				DamageClear: func(gotTarget *Object, damage int32) {
					if gotTarget != target || damage != tc.damage {
						t.Fatalf("DamageClear(%p, %d)", gotTarget, damage)
					}
					target.HealthData.Cur -= uint16(damage)
					events = append(events, "damage")
				},
				Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
					t.Fatalf("Magic Missile explosion rejected: %s", reason)
				},
			}

			if !DefaultDamageWorld4E0B30(target, tc.source, tc.weapon, tc.damage, object.DamageExplosion, runtime) {
				t.Fatal("Magic Missile explosion returned false")
			}
			if target.HealthData.Cur != uint16(40-tc.damage) || target.Pos132 != missile.PrevPos ||
				target.Obj130 != tc.wantSource || target.Field131 != uint32(object.DamageExplosion) ||
				target.Frame134 != 1300 {
				t.Fatalf("target state = health:%d pos:%+v source:%p type:%d frame:%d",
					target.HealthData.Cur, target.Pos132, target.Obj130, target.Field131, target.Frame134)
			}
			if !update.StatusFlags.Has(object.MonStatusOnFire|object.MonStatusInjured) ||
				update.Field546 != uint32(object.DamageExplosion) || update.Field547 != 2 {
				t.Fatalf("monster hit state = status:%#x type:%d latch:%d",
					update.StatusFlags, update.Field546, update.Field547)
			}
			wantEvents := []string{"fire-protection", "buff-off", "sound", "field-guide", "damage"}
			if !reflect.DeepEqual(events, wantEvents) {
				t.Fatalf("events = %v, want %v", events, wantEvents)
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30ElectricImmuneMonster(t *testing.T) {
	for _, typ := range []object.DamageType{object.DamageElectric, object.DamageAirborneElectric} {
		t.Run(typ.String(), func(t *testing.T) {
			target := &Object{
				ObjClass: object.ClassMonster, ObjSubClass: 0x800,
				HealthData: &HealthData{Cur: 12, Max: 12}, UpdateData: unsafe.Pointer(&MonsterUpdateData{}),
			}
			source := &Object{ObjClass: object.ClassPlayer}
			if !DefaultDamageWorld4E0B30(target, source, nil, 8, typ,
				DefaultDamageWorldRuntime4E0B30{
					GameplayFlag1: func() bool { return true },
					DamageClear:   func(*Object, int32) { t.Fatal("immune monster damaged") },
				}) {
				t.Fatal("immune branch returned false")
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30PlayerMeleeShapes(t *testing.T) {
	tests := []struct {
		name   string
		weapon *Object
		damage int32
		typ    object.DamageType
	}{
		{
			name:   "wooden staff",
			weapon: &Object{ObjClass: object.ClassWand, InitData: unsafe.Pointer(&ModifierInitData{})},
			damage: 46,
			typ:    object.DamageBlade,
		},
		{
			name:   "unarmed",
			damage: 10,
			typ:    object.DamageClaw,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			update := &MonsterUpdateData{}
			target := &Object{
				ObjClass:    object.ClassMonster,
				ObjSubClass: 0x202,
				HealthData:  &HealthData{Cur: 80, Max: 80},
				UpdateData:  unsafe.Pointer(update),
			}
			source := &Object{ObjClass: object.ClassPlayer, PrevPos: types.Pointf{X: 31, Y: 47}}
			soundSource := source
			if tc.weapon != nil {
				soundSource = tc.weapon
			}
			runtime := DefaultDamageWorldRuntime4E0B30{
				Frame:         func() uint32 { return 911 },
				GameplayFlag1: func() bool { return true },
				IsEnemy: func(gotTarget, gotSource *Object) bool {
					return gotTarget == target && gotSource == source
				},
				DefaultDamageSound: func(gotTarget, gotSource *Object) {
					if gotTarget != target || gotSource != soundSource {
						t.Fatalf("DefaultDamageSound(%p, %p), want (%p, %p)", gotTarget, gotSource, target, soundSource)
					}
				},
				DamageClear: func(gotTarget *Object, gotDamage int32) {
					if gotTarget != target || gotDamage != tc.damage {
						t.Fatalf("DamageClear(%p, %d), want (%p, %d)", gotTarget, gotDamage, target, tc.damage)
					}
					target.HealthData.Cur -= uint16(gotDamage)
				},
				Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
					t.Fatalf("player melee branch rejected: %s", reason)
				},
			}

			if !DefaultDamageWorld4E0B30(target, source, tc.weapon, tc.damage, tc.typ, runtime) {
				t.Fatal("player melee branch returned false")
			}
			if got := target.HealthData.Cur; got != uint16(80-tc.damage) {
				t.Fatalf("target health = %d, want %d", got, 80-tc.damage)
			}
			if target.Pos132 != source.PrevPos || target.Obj130 != soundSource ||
				target.Field131 != uint32(tc.typ) || target.Frame134 != 911 {
				t.Fatalf("target metadata = pos:%+v source:%p type:%d frame:%d",
					target.Pos132, target.Obj130, target.Field131, target.Frame134)
			}
			if !update.StatusFlags.Has(object.MonStatusInjured) || update.Field546 != uint32(tc.typ) || update.Field547 != 2 {
				t.Fatalf("monster hit state = status:%#x field546:%d field547:%d",
					update.StatusFlags, update.Field546, update.Field547)
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30SpiderBitesAirshipCaptain(t *testing.T) {
	targetUpdate := &MonsterUpdateData{Field547: 99}
	target := &Object{
		ObjClass:    object.ClassMonster,
		ObjSubClass: 0x10002,
		HealthData:  &HealthData{Cur: 30, Max: 30},
		UpdateData:  unsafe.Pointer(targetUpdate),
	}
	sourceUpdate := &MonsterUpdateData{}
	source := &Object{
		ObjClass:    object.ClassMonster,
		ObjSubClass: 0x202,
		PrevPos:     types.Pointf{X: 4481.25, Y: 2107.5},
		UpdateData:  unsafe.Pointer(sourceUpdate),
	}
	var events []string
	runtime := DefaultDamageWorldRuntime4E0B30{
		Frame:         func() uint32 { return 863 },
		GameplayFlag1: func() bool { return true },
		IsEnemy: func(gotTarget, gotSource *Object) bool {
			if gotTarget != target || gotSource != source {
				t.Fatalf("IsEnemy(%p, %p)", gotTarget, gotSource)
			}
			return true
		},
		BuffOff: func(got *Object, enchant EnchantID) {
			if got != target || enchant != defaultDamageInvisibleEnchant4E0B30 {
				t.Fatalf("BuffOff(%p, %d)", got, enchant)
			}
			events = append(events, "buff-off")
		},
		MonsterHasHitSound: func(got *Object) bool {
			if got != source {
				t.Fatalf("MonsterHasHitSound(%p)", got)
			}
			events = append(events, "hit-sound")
			return true
		},
		DefaultDamageSound: func(*Object, *Object) {
			t.Fatal("source monster hit sound must suppress the target damage sound")
		},
		AdjustFieldGuide: func(gotSource, gotTarget *Object, gotDamage int32) int32 {
			if gotSource != source || gotTarget != target || gotDamage != 3 {
				t.Fatalf("AdjustFieldGuide(%p, %p, %d)", gotSource, gotTarget, gotDamage)
			}
			events = append(events, "field-guide")
			return gotDamage
		},
		DamageClear: func(gotTarget *Object, gotDamage int32) {
			if gotTarget != target || gotDamage != 3 {
				t.Fatalf("DamageClear(%p, %d)", gotTarget, gotDamage)
			}
			target.HealthData.Cur -= uint16(gotDamage)
			events = append(events, "damage")
		},
		Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
			t.Fatalf("Spider BITE branch rejected: %s", reason)
		},
	}

	if !DefaultDamageWorld4E0B30(target, source, source, 3, object.DamageBite, runtime) {
		t.Fatal("Spider BITE branch returned false")
	}
	if target.HealthData.Cur != 27 {
		t.Fatalf("target health = %d, want 27", target.HealthData.Cur)
	}
	if target.Pos132 != source.PrevPos || target.Obj130 != source || target.Field131 != uint32(object.DamageBite) || target.Frame134 != 863 {
		t.Fatalf("target metadata = pos:%+v source:%p type:%d frame:%d", target.Pos132, target.Obj130, target.Field131, target.Frame134)
	}
	if !targetUpdate.StatusFlags.Has(object.MonStatusInjured) || targetUpdate.Field546 != uint32(object.DamageBite) || targetUpdate.Field547 != 2 {
		t.Fatalf("target hit state = status:%#x field546:%d field547:%d", targetUpdate.StatusFlags, targetUpdate.Field546, targetUpdate.Field547)
	}
	if sourceUpdate.Field130 != 863 {
		t.Fatalf("source combat timestamp = %d, want 863", sourceUpdate.Field130)
	}
	want := []string{"buff-off", "hit-sound", "field-guide", "damage"}
	if len(events) != len(want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	for i := range want {
		if events[i] != want[i] {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}
}

func TestDefaultDamageWorld4E0B30MonsterBiteUsesTargetSoundWithoutHitSound(t *testing.T) {
	targetUpdate := &MonsterUpdateData{}
	target := &Object{
		ObjClass:   object.ClassMonster,
		HealthData: &HealthData{Cur: 12, Max: 12},
		UpdateData: unsafe.Pointer(targetUpdate),
	}
	sourceUpdate := &MonsterUpdateData{}
	source := &Object{
		ObjClass:   object.ClassMonster,
		UpdateData: unsafe.Pointer(sourceUpdate),
	}
	var events []string
	runtime := DefaultDamageWorldRuntime4E0B30{
		Frame:         func() uint32 { return 91 },
		GameplayFlag1: func() bool { return true },
		IsEnemy:       func(*Object, *Object) bool { return true },
		MonsterHasHitSound: func(got *Object) bool {
			if got != source {
				t.Fatalf("MonsterHasHitSound(%p), want source %p", got, source)
			}
			events = append(events, "hit-sound")
			return false
		},
		DefaultDamageSound: func(gotTarget, gotSource *Object) {
			if gotTarget != target || gotSource != source {
				t.Fatalf("DefaultDamageSound(%p, %p), want (%p, %p)", gotTarget, gotSource, target, source)
			}
			events = append(events, "sound")
		},
		DamageClear: func(gotTarget *Object, damage int32) {
			if gotTarget != target || damage != 3 {
				t.Fatalf("DamageClear(%p, %d), want (%p, 3)", gotTarget, damage, target)
			}
			events = append(events, "damage")
		},
		Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
			t.Fatalf("monster BITE without hit sound rejected: %s", reason)
		},
	}

	DefaultDamageWorld4E0B30(target, source, source, 3, object.DamageBite, runtime)
	want := []string{"hit-sound", "sound", "damage"}
	if len(events) != len(want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	for i := range want {
		if events[i] != want[i] {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}
}

func TestDefaultDamageAttackQualifies4E1400(t *testing.T) {
	chargedWandData := &WandUseData{Flags: 2}
	idleWandData := &WandUseData{}
	tests := []struct {
		name   string
		source *Object
		weapon *Object
		want   bool
	}{
		{
			name:   "unarmed player",
			source: &Object{ObjClass: object.ClassPlayer},
			want:   true,
		},
		{
			name:   "eligible unarmed monster",
			source: &Object{ObjClass: object.ClassMonster, ObjSubClass: 0x10},
			want:   true,
		},
		{
			name:   "ordinary unarmed monster",
			source: &Object{ObjClass: object.ClassMonster},
		},
		{
			name:   "basic wand",
			source: &Object{ObjClass: object.ClassPlayer},
			weapon: &Object{ObjClass: object.ClassWand},
			want:   true,
		},
		{
			name:   "charged ranged wand",
			source: &Object{ObjClass: object.ClassPlayer},
			weapon: &Object{
				ObjClass:    object.ClassWand,
				ObjSubClass: 0x00010000,
				UseData:     UseDataPtr{Ptr: unsafe.Pointer(chargedWandData)},
			},
			want: true,
		},
		{
			name:   "idle ranged wand",
			source: &Object{ObjClass: object.ClassPlayer},
			weapon: &Object{
				ObjClass:    object.ClassWand,
				ObjSubClass: 0x00010000,
				UseData:     UseDataPtr{Ptr: unsafe.Pointer(idleWandData)},
			},
		},
		{
			name:   "melee weapon",
			source: &Object{ObjClass: object.ClassPlayer},
			weapon: &Object{ObjClass: object.ClassWeapon},
			want:   true,
		},
		{
			name:   "ranged weapon",
			source: &Object{ObjClass: object.ClassPlayer},
			weapon: &Object{ObjClass: object.ClassWeapon, ObjSubClass: 0x2},
		},
		{
			name:   "monster self attack",
			source: &Object{ObjClass: object.ClassMonster},
			weapon: &Object{ObjClass: object.ClassMonster},
			want:   true,
		},
		{
			name:   "missile",
			source: &Object{ObjClass: object.ClassPlayer},
			weapon: &Object{ObjClass: object.ClassMissile},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := defaultDamageAttackQualifies4E1400(tc.source, tc.weapon); got != tc.want {
				t.Fatalf("defaultDamageAttackQualifies4E1400() = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestDefaultDamageWorld4E0B30ShockRetaliation(t *testing.T) {
	target := &Object{
		ObjClass:   object.ClassMonster,
		Buffs:      uint32(1) << defaultDamageShockEnchant4E0B30,
		HealthData: &HealthData{Cur: 20, Max: 20},
		UpdateData: unsafe.Pointer(&MonsterUpdateData{}),
	}
	source := &Object{ObjClass: object.ClassPlayer}
	weapon := &Object{ObjClass: object.ClassWeapon}
	var events []string
	runtime := DefaultDamageWorldRuntime4E0B30{
		Frame:         func() uint32 { return 77 },
		GameplayFlag1: func() bool { return true },
		IsEnemy: func(gotTarget, gotSource *Object) bool {
			if gotTarget != target || gotSource != source {
				t.Fatalf("IsEnemy(%p, %p), want (%p, %p)", gotTarget, gotSource, target, source)
			}
			return true
		},
		Audio: func(id int, got *Object) {
			if id != defaultDamageShockSound4E0B30 || got != source {
				t.Fatalf("Audio(%d, %p), want (%d, %p)", id, got, defaultDamageShockSound4E0B30, source)
			}
			events = append(events, "shock-audio")
		},
		BuffOff: func(got *Object, enchant EnchantID) {
			if got != target {
				t.Fatalf("BuffOff object = %p, want %p", got, target)
			}
			switch enchant {
			case defaultDamageShockEnchant4E0B30:
				events = append(events, "shock-off")
			case defaultDamageInvisibleEnchant4E0B30:
				events = append(events, "invisibility-off")
			default:
				t.Fatalf("BuffOff enchant = %d", enchant)
			}
			target.Buffs &^= uint32(1) << uint32(enchant)
		},
		BalanceFloatInd: func(key string, index int) float64 {
			if key != defaultDamageShockBalance4E0B30 || index != defaultDamageShockBalanceIndex4E0B30 {
				t.Fatalf("BalanceFloatInd(%q, %d)", key, index)
			}
			events = append(events, "shock-balance")
			return 12.5
		},
		CallDamage: func(gotTarget, gotSource, gotWeapon *Object, gotDamage int32, gotType object.DamageType) bool {
			if gotTarget != source || gotSource != target || gotWeapon != nil ||
				gotDamage != 12 || gotType != object.DamageElectric {
				t.Fatalf("CallDamage(%p, %p, %p, %d, %v)", gotTarget, gotSource, gotWeapon, gotDamage, gotType)
			}
			events = append(events, "shock-damage")
			return true
		},
		PlayerSetState: func(got *Object, state PlayerState) bool {
			if got != source || state != PlayerState23 {
				t.Fatalf("PlayerSetState(%p, %d), want (%p, %d)", got, state, source, PlayerState23)
			}
			events = append(events, "shock-state")
			return true
		},
		DefaultDamageSound: func(gotTarget, gotSource *Object) {
			if gotTarget != target || gotSource != weapon {
				t.Fatalf("DefaultDamageSound(%p, %p), want (%p, %p)", gotTarget, gotSource, target, weapon)
			}
			events = append(events, "damage-sound")
		},
		DamageClear: func(gotTarget *Object, gotDamage int32) {
			if gotTarget != target || gotDamage != 5 {
				t.Fatalf("DamageClear(%p, %d), want (%p, 5)", gotTarget, gotDamage, target)
			}
			events = append(events, "damage")
		},
		Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
			t.Fatalf("Shock melee branch rejected: %s", reason)
		},
	}

	if !DefaultDamageWorld4E0B30(target, source, weapon, 5, object.DamageBlade, runtime) {
		t.Fatal("Shock melee branch returned false")
	}
	want := []string{
		"shock-audio", "shock-off", "shock-balance", "shock-damage", "shock-state",
		"invisibility-off", "damage-sound", "damage",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestDefaultDamageWorld4E0B30ShockDoesNotBlockMagicMissile(t *testing.T) {
	target := &Object{
		ObjClass:   object.ClassMonster,
		Buffs:      uint32(1) << defaultDamageShockEnchant4E0B30,
		HealthData: &HealthData{Cur: 20, Max: 20},
		UpdateData: unsafe.Pointer(&MonsterUpdateData{}),
	}
	source := &Object{ObjClass: object.ClassPlayer}
	missile := &Object{ObjClass: object.ClassMissile}
	damaged := false
	runtime := DefaultDamageWorldRuntime4E0B30{
		GameplayFlag1: func() bool { return true },
		FireProtection: func(got *Object) float64 {
			if got != target {
				t.Fatalf("FireProtection(%p), want %p", got, target)
			}
			return 0
		},
		DamageClear: func(got *Object, damage int32) {
			if got != target || damage != 7 {
				t.Fatalf("DamageClear(%p, %d), want (%p, 7)", got, damage, target)
			}
			damaged = true
		},
		Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
			t.Fatalf("Magic Missile against Shock target rejected: %s", reason)
		},
	}

	if !DefaultDamageWorld4E0B30(target, source, missile, 7, object.DamageExplosion, runtime) {
		t.Fatal("Magic Missile against Shock target returned false")
	}
	if !damaged {
		t.Fatal("Magic Missile damage was not applied")
	}
}

func TestDefaultDamageWorld4E0B30AppliesVampirismBeforeFieldGuide(t *testing.T) {
	target := &Object{
		ObjClass:   object.ClassMonster,
		HealthData: &HealthData{Cur: 20, Max: 20},
		UpdateData: unsafe.Pointer(&MonsterUpdateData{}),
		PosVec:     types.Pointf{X: 20.5, Y: 21.5},
	}
	source := &Object{
		ObjClass: object.ClassPlayer,
		Buffs:    uint32(1) << damageVampirismEnchant4E0B30,
		PosVec:   types.Pointf{X: 10.5, Y: 11.5},
	}
	source.BuffsPower[damageVampirismEnchant4E0B30] = 2
	weapon := &Object{ObjClass: object.ClassWeapon}
	var events []string
	runtime := DefaultDamageWorldRuntime4E0B30{
		Frame:         func() uint32 { return 81 },
		GameplayFlag1: func() bool { return true },
		IsEnemy:       func(*Object, *Object) bool { return true },
		BuffOff: func(got *Object, enchant EnchantID) {
			if got != target || enchant != defaultDamageInvisibleEnchant4E0B30 {
				t.Fatalf("BuffOff(%p, %d), want (%p, %d)", got, enchant, target, defaultDamageInvisibleEnchant4E0B30)
			}
			events = append(events, "buff-off")
		},
		DefaultDamageSound: func(gotTarget, gotWeapon *Object) {
			if gotTarget != target || gotWeapon != weapon {
				t.Fatalf("DefaultDamageSound(%p, %p), want (%p, %p)", gotTarget, gotWeapon, target, weapon)
			}
			events = append(events, "damage-sound")
		},
		Audio: func(id int, got *Object) {
			if id != damageVampirismSound4E0B30 || got != weapon {
				t.Fatalf("Audio(%d, %p), want (%d, %p)", id, got, damageVampirismSound4E0B30, weapon)
			}
			events = append(events, "vampirism-audio")
		},
		BalanceFloatInd: func(key string, index int) float64 {
			if key != damageVampirismBalance4E0B30 || index != 1 {
				t.Fatalf("BalanceFloatInd(%q, %d), want (%q, 1)", key, index, damageVampirismBalance4E0B30)
			}
			events = append(events, "vampirism-balance")
			return 0.5
		},
		AdjustHP: func(got *Object, amount int32) {
			if got != source || amount != 2 {
				t.Fatalf("AdjustHP(%p, %d), want (%p, 2)", got, amount, source)
			}
			events = append(events, "vampirism-heal")
		},
		VampirismFX: func(id int, gotSource, gotTarget image.Point, amount uint16) {
			if id != damageVampirismFX4E0B30 || gotSource != (image.Pt(10, 12)) ||
				gotTarget != (image.Pt(20, 22)) || amount != 2 {
				t.Fatalf("VampirismFX(%d, %v, %v, %d)", id, gotSource, gotTarget, amount)
			}
			events = append(events, "vampirism-fx")
		},
		AdjustFieldGuide: func(gotSource, gotTarget *Object, damage int32) int32 {
			if gotSource != source || gotTarget != target || damage != 5 {
				t.Fatalf("AdjustFieldGuide(%p, %p, %d), want (%p, %p, 5)", gotSource, gotTarget, damage, source, target)
			}
			events = append(events, "field-guide")
			return 9
		},
		DamageClear: func(got *Object, damage int32) {
			if got != target || damage != 9 {
				t.Fatalf("DamageClear(%p, %d), want (%p, 9)", got, damage, target)
			}
			events = append(events, "damage")
		},
		Unsupported: func(reason string, _, _, _ *Object, _ int32, _ object.DamageType) {
			t.Fatalf("Vampirism melee branch rejected: %s", reason)
		},
	}

	if !DefaultDamageWorld4E0B30(target, source, weapon, 5, object.DamageBlade, runtime) {
		t.Fatal("Vampirism melee branch returned false")
	}
	want := []string{
		"buff-off", "damage-sound", "vampirism-audio", "vampirism-balance",
		"vampirism-heal", "vampirism-fx", "field-guide", "damage",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}
