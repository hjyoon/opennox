package server

import (
	"fmt"
	"reflect"
	"testing"
)

type objectPlayerMasksClearObject4E80C0 struct {
	name             string
	field35, field36 uint32
	next             *objectPlayerMasksClearObject4E80C0
}

func objectPlayerMasksClearTestHooks4E80C0(
	events *[]string,
	first **objectPlayerMasksClearObject4E80C0,
) objectPlayerMasksClearHooks4E80C0[*objectPlayerMasksClearObject4E80C0] {
	return objectPlayerMasksClearHooks4E80C0[*objectPlayerMasksClearObject4E80C0]{
		firstObject: func() *objectPlayerMasksClearObject4E80C0 {
			*events = append(*events, "first")
			return *first
		},
		loadField36: func(obj *objectPlayerMasksClearObject4E80C0) uint32 {
			*events = append(*events, "load36:"+obj.name)
			return obj.field36
		},
		loadField35: func(obj *objectPlayerMasksClearObject4E80C0) uint32 {
			*events = append(*events, "load35:"+obj.name)
			return obj.field35
		},
		storeField36: func(obj *objectPlayerMasksClearObject4E80C0, value uint32) {
			*events = append(*events, fmt.Sprintf("store36:%s:%#x", obj.name, value))
			obj.field36 = value
		},
		storeField35: func(obj *objectPlayerMasksClearObject4E80C0, value uint32) {
			*events = append(*events, fmt.Sprintf("store35:%s:%#x", obj.name, value))
			obj.field35 = value
		},
		nextObject: func(obj *objectPlayerMasksClearObject4E80C0) *objectPlayerMasksClearObject4E80C0 {
			*events = append(*events, "next:"+obj.name)
			return obj.next
		},
	}
}

func TestObjectPlayerMasksClear4E80C0EmptyListStopsAfterHead(t *testing.T) {
	var first *objectPlayerMasksClearObject4E80C0
	var events []string
	got := objectPlayerMasksClear4E80C0(7, objectPlayerMasksClearTestHooks4E80C0(&events, &first))
	if got != nil {
		t.Fatalf("result = %p, want nil", got)
	}
	if want := []string{"first"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestObjectPlayerMasksClear4E80C0MasksShiftCount(t *testing.T) {
	tests := []struct {
		playerInd uint32
		want      uint32
	}{
		{playerInd: 0, want: 0xfffffffe},
		{playerInd: 31, want: 0x7fffffff},
		{playerInd: 32, want: 0xfffffffe},
		{playerInd: 34, want: 0xfffffffb},
		{playerInd: 63, want: 0x7fffffff},
		{playerInd: ^uint32(0), want: 0x7fffffff},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("index_%08x", tc.playerInd), func(t *testing.T) {
			obj := &objectPlayerMasksClearObject4E80C0{name: "obj", field35: ^uint32(0), field36: ^uint32(0)}
			first := obj
			var events []string
			got := objectPlayerMasksClear4E80C0(tc.playerInd, objectPlayerMasksClearTestHooks4E80C0(&events, &first))
			if got != nil {
				t.Fatalf("result = %p, want nil", got)
			}
			if obj.field35 != tc.want || obj.field36 != tc.want {
				t.Fatalf("masks = (%#x, %#x), want both %#x", obj.field35, obj.field36, tc.want)
			}
		})
	}
}

func TestObjectPlayerMasksClear4E80C0CachesLoadsAndUsesLiveSuccessor(t *testing.T) {
	stale := &objectPlayerMasksClearObject4E80C0{name: "stale", field35: 0xffffffff, field36: 0xffffffff}
	replacement := &objectPlayerMasksClearObject4E80C0{name: "replacement", field35: 0x16, field36: 0x1e}
	obj := &objectPlayerMasksClearObject4E80C0{name: "obj", field35: 0x0f, field36: 0x1f, next: stale}
	first := obj
	var events []string
	hooks := objectPlayerMasksClearTestHooks4E80C0(&events, &first)
	baseStore36 := hooks.storeField36
	hooks.storeField36 = func(current *objectPlayerMasksClearObject4E80C0, value uint32) {
		baseStore36(current, value)
		if current == obj {
			current.field35 = 0xffffffff
			current.next = replacement
		}
	}

	got := objectPlayerMasksClear4E80C0(3, hooks)
	if got != nil {
		t.Fatalf("result = %p, want nil", got)
	}
	if obj.field36 != 0x17 || obj.field35 != 0x07 {
		t.Fatalf("first masks = (%#x, %#x), want (0x7, 0x17)", obj.field35, obj.field36)
	}
	if replacement.field36 != 0x16 || replacement.field35 != 0x16 {
		t.Fatalf("replacement masks = (%#x, %#x), want both 0x16", replacement.field35, replacement.field36)
	}
	if stale.field36 != 0xffffffff || stale.field35 != 0xffffffff {
		t.Fatalf("stale successor was visited: masks = (%#x, %#x)", stale.field35, stale.field36)
	}
	want := []string{
		"first",
		"load36:obj", "load35:obj", "store36:obj:0x17", "store35:obj:0x7", "next:obj",
		"load36:replacement", "load35:replacement", "store36:replacement:0x16", "store35:replacement:0x16", "next:replacement",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}
