package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

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
