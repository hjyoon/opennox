package server

import (
	"sync/atomic"
	"testing"

	"github.com/opennox/libs/object"
)

func TestObjectXStatusMatchesGAMEEXEContract(t *testing.T) {
	tests := []struct {
		name      string
		set       bool
		status    uint32
		field5    uint32
		class     object.Class
		typeField uint32
	}{
		{name: "set status one skips sync", set: true, status: 1, field5: 0x20, class: object.ClassPlayer},
		{name: "unset absent skips all writes", status: 2, field5: 0x20, class: object.ClassPlayer},
		{name: "unset status one skips sync", status: 1, field5: 0x21, class: object.ClassPlayer},
		{name: "set special object", set: true, status: 2, field5: 0x20, class: object.ClassClientPersist},
		{name: "unset special object", status: 2, field5: 0x22, class: object.ClassImmobile},
		{name: "set ordinary object matching type", set: true, status: 2, field5: 0x20, class: object.ClassMissile, typeField: 0x22},
		{name: "set ordinary object differing type", set: true, status: 2, field5: 0x20, class: object.ClassMissile, typeField: 0},
		{name: "unset ordinary object matching type", status: 2, field5: 0x22, class: object.ClassMissile, typeField: 0x20},
		{name: "unset ordinary object differing type", status: 2, field5: 0x22, class: object.ClassMissile, typeField: 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := &Object{
				ObjClass: tc.class,
				ObjFlags: object.FlagActive,
				Field5:   tc.field5,
				Field37:  0x80000015,
				Field38:  0x2468ACE0,
			}
			for i := range got.Field140 {
				got.Field140[i] = 0xA5000000 | 0x80000 | 8 | uint32(i<<4)
			}
			if !tc.class.HasAny(object.ClassClientPersist | object.ClassImmobile | object.ClassPlayer) {
				bindStatusTestType(t, got, tc.typeField)
			}
			want := *got
			applyGAMEEXEXStatus(&want, tc.set, tc.status, tc.typeField)

			if tc.set {
				got.SetXStatus(tc.status)
			} else {
				got.UnsetXStatus(tc.status)
			}
			if got.Field5 != want.Field5 || got.Field38 != want.Field38 {
				t.Errorf("status/sync: got (%#08x, %#08x), want (%#08x, %#08x)", got.Field5, got.Field38, want.Field5, want.Field38)
			}
			if got.Field140 != want.Field140 {
				t.Errorf("Field140 differs: got %#v, want %#v", got.Field140, want.Field140)
			}
			if got.Field37 != want.Field37 || got.ObjClass != want.ObjClass || got.ObjFlags != want.ObjFlags {
				t.Fatal("status update overwrote state outside the original function contract")
			}
		})
	}
}

func bindStatusTestType(t *testing.T, obj *Object, field9 uint32) {
	t.Helper()
	h := atomic.AddUintptr(&serverLast, 1)
	s := &Server{handle: h}
	s.Types.byInd = []*ObjectType{nil, {Field9: field9}}
	servers.Store(h, s)
	t.Cleanup(s.Close)
	obj.serverHandle = h
	obj.TypeInd = 1
}

func applyGAMEEXEXStatus(obj *Object, set bool, status, typeField uint32) {
	if set {
		obj.Field5 |= status
	} else {
		if obj.Field5&status == 0 {
			return
		}
		obj.Field5 &^= status
	}
	if status == 1 {
		return
	}
	obj.Field38 = ^uint32(0)
	if obj.Class().HasAny(object.ClassClientPersist | object.ClassImmobile | object.ClassPlayer) {
		for i := range obj.Field140 {
			obj.Field140[i] = obj.Field140[i]&0xFFFFF000 | 0x80000
		}
		return
	}
	changed := typeField != obj.Field5
	for i := range obj.Field140 {
		if changed {
			obj.Field140[i] |= 8
		} else {
			obj.Field140[i] &^= 8
		}
		if obj.Field37&uint32(1<<i) != 0 {
			obj.Field140[i] |= 0x80000
		} else if obj.Field140[i]&8 == 0 {
			obj.Field140[i] &^= 0x80000
		}
	}
}
