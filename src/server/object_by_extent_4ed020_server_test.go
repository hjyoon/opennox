package server

import (
	"testing"
	"unsafe"
)

func TestObjectByExtentNativeLayout4ED020(t *testing.T) {
	wantFlags := uintptr(16)
	wantExtent := uintptr(40)
	wantNext := uintptr(444)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantFlags = 20
		wantExtent = 44
		wantNext = 448
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantFlags},
		{"Object.Extent", unsafe.Offsetof(Object{}.Extent), wantExtent},
		{"Object.ObjNext", unsafe.Offsetof(Object{}.ObjNext), wantNext},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s offset = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestObjectByExtentNativeSearch4ED020(t *testing.T) {
	const wanted = uint32(0xffffffff)
	dead := &Object{ObjFlags: 0x80000020, Extent: wanted}
	other := &Object{Extent: wanted - 1}
	match := &Object{ObjFlags: 0x80000000, Extent: wanted}
	dead.ObjNext = other
	other.ObjNext = match
	s := &Server{}
	s.Objs.List = dead

	if got := s.ObjectByExtent4ED020(wanted); got != match {
		t.Fatalf("unsigned result = %p, want %p", got, match)
	}
	if got := s.Objs.GetObjectByInd(-1); got != match {
		t.Fatalf("legacy signed result = %p, want %p", got, match)
	}
	if got := s.ObjectByExtent4ED020(wanted - 2); got != nil {
		t.Fatalf("missing result = %p, want nil", got)
	}
}
