package server

import (
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/balance"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestCastStunNative52C2C0PlayerLinksAndMass(t *testing.T) {
	for _, tc := range []struct {
		class       object.Class
		playerClass uint8
		mass        float32
		buff        int32
	}{
		{object.ClassPlayer, 0, 1000, 4}, {object.ClassPlayer, 1, 0, 5},
		{object.ClassPlayer, 2, 0, 5}, {object.ClassPlayer, 255, 0, 5},
		{object.ClassPlayer | object.ClassMonster, 1, 1000, 5},
		{object.ClassMonster, 0, 15, 5}, {object.ClassMonster, 0, math.Nextafter32(15, 16), 4},
		{object.ClassMonster, 0, float32(math.NaN()), 5}, {object.Class(0x10000), 0, 1000, 5},
	} {
		t.Run(fmt.Sprintf("class=%x/player=%d/mass=%g", tc.class, tc.playerClass, tc.mass), func(t *testing.T) {
			s := &Server{}
			s.Balance.file = &balance.File{Global: balance.Config{"stunenchantduration": balance.Float(2.50000001)}}
			source, target, reloaded := new(Object), new(Object), new(Object)
			p, update := new(Player), new(PlayerUpdateData)
			p.Info().playerClass, update.Player = tc.playerClass, p
			*target = Object{ObjClass: tc.class, Mass: tc.mass, ZSize2: 1000}
			if tc.class&object.ClassPlayer != 0 {
				target.UpdateData = unsafe.Pointer(update)
			}
			arg := &SpellAcceptArg{Obj: target, Pos: types.Pointf{X: -4.5, Y: 10.25}}
			if unsafe.Sizeof(uintptr(0)) > 4 {
				for _, ptr := range []unsafe.Pointer{unsafe.Pointer(source), unsafe.Pointer(target), unsafe.Pointer(reloaded), unsafe.Pointer(arg), unsafe.Pointer(p), unsafe.Pointer(update)} {
					if uintptr(ptr) <= math.MaxUint32 {
						t.Fatalf("native pointer = %p, want above 4 GiB", ptr)
					}
				}
			}
			applyCalls, attributeCalls := 0, 0
			r := CastStunRuntime52C2C0{
				BuffApply: func(unit *Object, buff int32, duration int16, power int8) {
					applyCalls++
					if unit != target || buff != tc.buff || duration != 2 || power != -128 {
						t.Fatalf("buff = %p/%d/%d/%d", unit, buff, duration, power)
					}
					arg.Obj = reloaded
				},
				Attribution: func(caster, unit *Object) {
					attributeCalls++
					if caster != source || unit != reloaded {
						t.Fatalf("attribution = %p/%p", caster, unit)
					}
				},
			}
			if got := s.CastStun52C2C0(source, arg, 0x180, r); got != 1 || applyCalls != 1 || attributeCalls != 1 {
				t.Fatalf("result/calls = %d/%d/%d", got, applyCalls, attributeCalls)
			}
			if arg.Obj != reloaded || arg.Pos != (types.Pointf{X: -4.5, Y: 10.25}) {
				t.Fatal("unexpected spell argument change")
			}
			arg.Obj = nil
			if got := s.CastStun52C2C0(source, arg, 1, CastStunRuntime52C2C0{}); got != 0 {
				t.Fatalf("nil target = %d", got)
			}
		})
	}
}

func TestCastStunNative52C2C0RequiredLinkFaults(t *testing.T) {
	for _, tc := range []struct {
		name string
		arg  *SpellAcceptArg
	}{
		{"nil-argument", nil},
		{"nil-player-update", &SpellAcceptArg{Obj: &Object{ObjClass: object.ClassPlayer}}},
		{"nil-player", &SpellAcceptArg{Obj: &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(new(PlayerUpdateData))}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			r := CastStunRuntime52C2C0{
				BuffApply: func(*Object, int32, int16, int8) { calls++ }, Attribution: func(*Object, *Object) { calls++ },
			}
			func() {
				defer func() {
					if recover() == nil {
						t.Fatal("invalid required link did not fault")
					}
				}()
				(&Server{}).CastStun52C2C0(nil, tc.arg, 1, r)
			}()
			if calls != 0 {
				t.Fatalf("calls after fault = %d", calls)
			}
		})
	}
	wantMass := uintptr(120)
	if unsafe.Sizeof(uintptr(0)) > 4 {
		wantMass = 124
	}
	if got := unsafe.Offsetof(Object{}.Mass); got != wantMass {
		t.Fatalf("native Mass offset = %d, want %d", got, wantMass)
	}
}
