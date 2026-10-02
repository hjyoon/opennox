package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/balance"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestCastShockNative52C5A0BalanceAndLiveDamageRegistry(t *testing.T) {
	s := new(Server)
	s.Balance.file = &balance.File{Global: balance.Config{
		"shocktrapdamage": balance.Array{1, 2.50000001, 3},
	}}
	cache := uint32(0)
	owner, caster, context := new(Object), new(Object), new(Object)
	caster.ObjOwner, context.TypeInd = owner, 0x8123
	target, live := new(Object), new(Object)
	health := &HealthData{Cur: 31, Max: 31}
	callback := unsafe.Pointer(new(byte))
	*live = Object{ObjClass: object.Class(0xfedcba98), ObjFlags: object.Flags(0xffffffff), HealthData: health, Damage: callback}
	arg := &SpellAcceptArg{Obj: target, Pos: types.Ptf(-4.5, 10.25)}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(owner), unsafe.Pointer(caster), unsafe.Pointer(context), unsafe.Pointer(target), unsafe.Pointer(live), unsafe.Pointer(arg)} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatalf("native Shock pointer=%p, want above 4 GiB", ptr)
			}
		}
	}
	calls := 0
	objDamage.Register(callback, func(gotTarget, source, weapon *Object, damage int32, typ object.DamageType) bool {
		calls++
		if gotTarget != live || source != caster || weapon != caster || damage != 2 || typ != object.DamageElectric {
			t.Fatalf("damage=%p/%p/%p/%d/%d", gotTarget, source, weapon, damage, typ)
		}
		gotTarget.HealthData.Cur -= uint16(damage)
		arg.Obj = nil
		return false
	})
	r := CastShockRuntime52C5A0{
		GlyphTypeCache: &cache,
		LookupTypeID: func(name string) uint32 {
			if name != "Glyph" {
				t.Fatal(name)
			}
			return 0x8123
		},
		BalanceFloatInd: func(key string, index int32) float64 {
			if key != "ShockTrapDamage" || index != 1 {
				t.Fatalf("table=%s/%d", key, index)
			}
			arg.Obj = live
			return s.Balance.FloatInd(key, int(index))
		},
	}
	if got := CastShock52C5A0(caster, context, arg, 2, r); got != 1 || calls != 1 || cache != 0x8123 || health.Cur != 29 || arg.Obj != nil {
		t.Fatalf("result/calls/cache/HP/target=%d/%d/%x/%d/%p", got, calls, cache, health.Cur, arg.Obj)
	}
	if arg.Pos != types.Ptf(-4.5, 10.25) {
		t.Fatal("Shock changed acceptance position")
	}
}

func TestCastShockNative52C5A0BuffWidthsAndWholeDwordCache(t *testing.T) {
	s := new(Server)
	s.Balance.file = &balance.File{Global: balance.Config{
		"shockenchantduration": balance.Float(32767.5),
	}}
	cache := uint32(0x18123) // Must not compare as uint16(0x8123).
	context := &Object{TypeInd: 0x8123}
	initial, live := new(Object), new(Object)
	arg := &SpellAcceptArg{Obj: initial}
	calls := 0
	r := CastShockRuntime52C5A0{
		GlyphTypeCache: &cache,
		BalanceFloat: func(key string) float64 {
			if key != "ShockEnchantDuration" {
				t.Fatal(key)
			}
			arg.Obj = live
			return s.Balance.Float(key)
		},
		BuffApply: func(unit *Object, buff int32, duration int16, power int8) {
			calls++
			if unit != live || buff != 22 || duration != math.MinInt16 || power != -128 {
				t.Fatalf("buff=%p/%d/%d/%d", unit, buff, duration, power)
			}
			arg.Obj = nil
		},
	}
	if got := CastShock52C5A0(nil, context, arg, 0x180, r); got != 1 || calls != 1 || cache != 0x18123 || arg.Obj != nil {
		t.Fatalf("result/calls/cache/target=%d/%d/%x/%p", got, calls, cache, arg.Obj)
	}
}

func TestCastShockNative52C5A0NilTargetNeedsNoRuntime(t *testing.T) {
	if got := CastShock52C5A0(nil, &Object{}, &SpellAcceptArg{}, math.MinInt32, CastShockRuntime52C5A0{}); got != 0 {
		t.Fatalf("nil target=%d", got)
	}
}

func TestCastShockNative52C5A0RequiredLinkFaults(t *testing.T) {
	for _, name := range []string{"argument", "cache", "damage-callback", "reloaded-target"} {
		t.Run(name, func(t *testing.T) {
			cache := uint32(17)
			arg := &SpellAcceptArg{Obj: new(Object)}
			r := CastShockRuntime52C5A0{
				GlyphTypeCache: &cache,
				BalanceFloatInd: func(string, int32) float64 {
					if name == "reloaded-target" {
						arg.Obj = nil
					}
					return 2
				},
			}
			if name == "argument" {
				arg = nil
			}
			if name == "cache" {
				r.GlyphTypeCache = nil
			}
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				CastShock52C5A0(nil, &Object{TypeInd: 17}, arg, 1, r)
			}()
			if recovered == nil {
				t.Fatal("invalid required Shock link did not fault")
			}
		})
	}
	cache := uint32(17)
	arg := &SpellAcceptArg{Obj: new(Object)}
	calls := 0
	if got := CastShock52C5A0(nil, nil, arg, 1, CastShockRuntime52C5A0{
		GlyphTypeCache: &cache,
		BalanceFloat:   func(string) float64 { arg.Obj = nil; return 2 },
		BuffApply: func(unit *Object, buff int32, duration int16, power int8) {
			calls++
			if unit != nil || buff != 22 || duration != 2 || power != 1 {
				t.Fatal("nil live target was masked")
			}
		},
	}); got != 1 || calls != 1 {
		t.Fatalf("live nil buff result/calls=%d/%d", got, calls)
	}
}
