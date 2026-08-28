package server

import (
	"math"
	"reflect"
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
