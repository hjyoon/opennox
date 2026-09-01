package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

func TestSkeletonDamage4E23C0InvulnerabilityCadence(t *testing.T) {
	target := &Object{Buffs: uint32(1) << ENCHANT_INVULNERABLE}
	for _, test := range []struct {
		frame     uint32
		wantAudio bool
	}{{frame: 12, wantAudio: true}, {frame: 13, wantAudio: false}} {
		var sounds []int
		defaultCalled := false
		got := SkeletonDamage4E23C0(target, nil, nil, 7, object.DamageBlade, SkeletonDamageRuntime4E23C0{
			Frame: func() uint32 { return test.frame },
			Audio: func(id int, got *Object) {
				if got != target {
					t.Fatalf("audio target = %p, want %p", got, target)
				}
				sounds = append(sounds, id)
			},
			Default: func(_, _, _ *Object, _ int32, _ object.DamageType) bool {
				defaultCalled = true
				return false
			},
		})
		if !got || defaultCalled {
			t.Fatalf("frame %d result/default = %t/%t, want true/false", test.frame, got, defaultCalled)
		}
		want := []int(nil)
		if test.wantAudio {
			want = []int{skeletonDamageInvulnerableSound4E23C0}
		}
		if !reflect.DeepEqual(sounds, want) {
			t.Fatalf("frame %d sounds = %v, want %v", test.frame, sounds, want)
		}
	}
}

func TestSkeletonDamage4E23C0BlockUsesWeaponPreviousPosition(t *testing.T) {
	update := &MonsterUpdateData{AIStackInd: 0, Field120_0: 100, Field120_1: 51}
	update.AIStack[0].Action = uint32(ai.ACTION_BLOCK_ATTACK)
	target := &Object{
		ObjClass:   object.ClassMonster,
		PosVec:     types.Ptf(11, 12),
		Direction1: 73,
		UpdateData: unsafe.Pointer(update),
	}
	source := &Object{PrevPos: types.Ptf(21, 22)}
	weapon := &Object{PrevPos: types.Ptf(31, 32)}
	var sounds []int
	defaultCalled := false
	got := SkeletonDamage4E23C0(target, source, weapon, -19, object.DamageElectric, SkeletonDamageRuntime4E23C0{
		Direction: func(pos types.Pointf, direction int16, attackPos types.Pointf) int32 {
			if pos != target.PosVec || direction != int16(target.Direction1) || attackPos != weapon.PrevPos {
				t.Fatalf("direction args = %+v/%d/%+v", pos, direction, attackPos)
			}
			return 1
		},
		Audio: func(id int, got *Object) {
			if got != target {
				t.Fatalf("audio target = %p, want %p", got, target)
			}
			sounds = append(sounds, id)
		},
		Default: func(_, _, _ *Object, _ int32, _ object.DamageType) bool {
			defaultCalled = true
			return false
		},
	})
	if !got || defaultCalled || !reflect.DeepEqual(sounds, []int{skeletonDamageBlockSound4E23C0}) {
		t.Fatalf("result/default/sounds = %t/%t/%v", got, defaultCalled, sounds)
	}
}

func TestSkeletonDamage4E23C0FallsThroughWithExactArguments(t *testing.T) {
	update := &MonsterUpdateData{AIStackInd: 0, Field120_0: 100, Field120_1: 50}
	update.AIStack[0].Action = uint32(ai.ACTION_BLOCK_ATTACK)
	target := &Object{ObjClass: object.ClassMonster, UpdateData: unsafe.Pointer(update)}
	source := &Object{}
	weapon := &Object{}
	called := false
	got := SkeletonDamage4E23C0(target, source, weapon, -19, object.DamageElectric, SkeletonDamageRuntime4E23C0{
		Direction: func(types.Pointf, int16, types.Pointf) int32 { return 1 },
		Default: func(gotTarget, gotSource, gotWeapon *Object, damage int32, typ object.DamageType) bool {
			called = true
			if gotTarget != target || gotSource != source || gotWeapon != weapon || damage != -19 || typ != object.DamageElectric {
				t.Fatalf("default args = %p/%p/%p/%d/%d", gotTarget, gotSource, gotWeapon, damage, typ)
			}
			return true
		},
	})
	if !got || !called {
		t.Fatalf("result/called = %t/%t, want true/true", got, called)
	}
}

func TestSpecialDamageTransforms4E24B0Through4E2560(t *testing.T) {
	target, source, weapon := &Object{}, &Object{}, &Object{}
	tests := []struct {
		name       string
		call       func(object.DamageType, DamageFunc) bool
		typ        object.DamageType
		wantDamage int32
		wantCall   bool
		wantResult bool
	}{
		{name: "stone", typ: object.DamagePoison, wantDamage: 17, wantCall: true, wantResult: true, call: func(typ object.DamageType, fallback DamageFunc) bool {
			return StoneDamage4E24B0(target, source, weapon, 17, typ, fallback)
		}},
		{name: "mech electric", typ: object.DamageElectric, wantDamage: 34, wantCall: true, wantResult: true, call: func(typ object.DamageType, fallback DamageFunc) bool {
			return MechGolemDamage4E24E0(target, source, weapon, 17, typ, fallback)
		}},
		{name: "mech airborne electric", typ: object.DamageAirborneElectric, wantDamage: 34, wantCall: true, wantResult: true, call: func(typ object.DamageType, fallback DamageFunc) bool {
			return MechGolemDamage4E24E0(target, source, weapon, 17, typ, fallback)
		}},
		{name: "mech blade", typ: object.DamageBlade, wantDamage: 17, wantCall: true, wantResult: true, call: func(typ object.DamageType, fallback DamageFunc) bool {
			return MechGolemDamage4E24E0(target, source, weapon, 17, typ, fallback)
		}},
		{name: "flammable explosion", typ: object.DamageExplosion, wantDamage: 9999999, wantCall: true, wantResult: true, call: func(typ object.DamageType, fallback DamageFunc) bool {
			return FlammableDamage4E2520(target, source, weapon, 17, typ, fallback)
		}},
		{name: "flammable poison", typ: object.DamagePoison, wantDamage: 17, wantCall: true, wantResult: true, call: func(typ object.DamageType, fallback DamageFunc) bool {
			return FlammableDamage4E2520(target, source, weapon, 17, typ, fallback)
		}},
		{name: "black powder crush", typ: object.DamageCrush, wantDamage: 999999, wantCall: true, wantResult: true, call: func(typ object.DamageType, fallback DamageFunc) bool {
			return BlackPowderDamage4E2560(target, source, weapon, 17, typ, fallback)
		}},
		{name: "black powder poison", typ: object.DamagePoison, wantDamage: 0, wantCall: false, wantResult: false, call: func(typ object.DamageType, fallback DamageFunc) bool {
			return BlackPowderDamage4E2560(target, source, weapon, 17, typ, fallback)
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			called := false
			fallback := func(gotTarget, gotSource, gotWeapon *Object, damage int32, typ object.DamageType) bool {
				called = true
				if gotTarget != target || gotSource != source || gotWeapon != weapon || damage != test.wantDamage || typ != test.typ {
					t.Fatalf("default args = %p/%p/%p/%d/%d", gotTarget, gotSource, gotWeapon, damage, typ)
				}
				return true
			}
			if got := test.call(test.typ, fallback); got != test.wantResult || called != test.wantCall {
				t.Fatalf("result/called = %t/%t, want %t/%t", got, called, test.wantResult, test.wantCall)
			}
		})
	}
}
