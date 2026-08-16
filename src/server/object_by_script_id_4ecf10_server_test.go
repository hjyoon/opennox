package server

import (
	"testing"
	"unsafe"
)

func TestObjectByScriptIDNativeLayout4ECF10(t *testing.T) {
	wantFlags := uintptr(16)
	wantScriptID := uintptr(44)
	wantNext := uintptr(444)
	wantNextItem := uintptr(496)
	wantFirstItem := uintptr(504)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantFlags = 20
		wantScriptID = 48
		wantNext = 448
		wantNextItem = 528
		wantFirstItem = 544
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantFlags},
		{"Object.ScriptIDVal", unsafe.Offsetof(Object{}.ScriptIDVal), wantScriptID},
		{"Object.ObjNext", unsafe.Offsetof(Object{}.ObjNext), wantNext},
		{"Object.InvNextItem", unsafe.Offsetof(Object{}.InvNextItem), wantNextItem},
		{"Object.InvFirstItem", unsafe.Offsetof(Object{}.InvFirstItem), wantFirstItem},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s offset = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestObjectByScriptIDNativeSearchDomains4ECF10(t *testing.T) {
	s := &Server{}
	active := &Object{ScriptIDVal: 11}
	inventory := &Object{ScriptIDVal: -22}
	pending := &Object{ScriptIDVal: 33}
	missile := &Object{ScriptIDVal: -2147483648}
	active.InvFirstItem = inventory
	s.Objs.List = active
	s.Objs.Pending = pending
	s.Objs.MissileList = missile

	for _, tc := range []struct {
		name     string
		scriptID int32
		want     *Object
	}{
		{"active", 11, active},
		{"inventory", -22, inventory},
		{"pending", 33, pending},
		{"missile-full-signed-width", -2147483648, missile},
		{"missing", 44, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := s.ObjectByScriptID4ECF10(tc.scriptID); got != tc.want {
				t.Fatalf("result = %p, want %p", got, tc.want)
			}
		})
	}
}

func TestObjectByScriptIDNativeDeadObjectsDoNotMatch4ECF10(t *testing.T) {
	const wanted = int32(77)
	deadActive := &Object{ObjFlags: 0x20, ScriptIDVal: wanted}
	deadItem := &Object{ObjFlags: 0x20, ScriptIDVal: wanted}
	deadPending := &Object{ObjFlags: 0x20, ScriptIDVal: wanted}
	liveMissile := &Object{ObjFlags: 0x100, ScriptIDVal: wanted}
	deadActive.InvFirstItem = deadItem
	s := &Server{}
	s.Objs.List = deadActive
	s.Objs.Pending = deadPending
	s.Objs.MissileList = liveMissile

	if got := s.ObjectByScriptID4ECF10(wanted); got != liveMissile {
		t.Fatalf("result = %p, want live missile %p", got, liveMissile)
	}
}
