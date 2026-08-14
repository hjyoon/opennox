package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
)

func TestAwardSpellCollide4EAD20NativeLayout(t *testing.T) {
	type objectLayout struct {
		size        uintptr
		collideData uintptr
	}
	wantByPointerSize := map[uintptr]objectLayout{
		4: {size: 780, collideData: 700},
		8: {size: 928, collideData: 776},
	}
	pointerSize := unsafe.Sizeof(uintptr(0))
	want, ok := wantByPointerSize[pointerSize]
	if !ok {
		t.Fatalf("unsupported pointer size %d", pointerSize)
	}
	got := objectLayout{
		size:        unsafe.Sizeof(Object{}),
		collideData: unsafe.Offsetof(Object{}.CollideData),
	}
	if got != want {
		t.Fatalf("Object layout = %+v, want %+v", got, want)
	}
	if got := unsafe.Sizeof(AwardSpellCollideData{}); got != 4 {
		t.Fatalf("AwardSpellCollideData size = %d, want 4", got)
	}
	if got := unsafe.Offsetof(AwardSpellCollideData{}.Spell); got != 0 {
		t.Fatalf("Spell offset = %d, want 0", got)
	}
}

func TestAwardSpellCollideNative4EAD20PreservesGrantABI(t *testing.T) {
	data := &AwardSpellCollideData{Spell: 0xf1234567}
	replacement := &AwardSpellCollideData{Spell: 0x11223344}
	source := &Object{CollideData: unsafe.Pointer(data)}
	target := &Object{}
	collision := &types.Pointf{X: 17.5, Y: -31.25}
	called := 0

	got := awardSpellCollideNative4EAD20(
		source,
		target,
		collision,
		awardSpellCollideNativeDeps4EAD20{
			grantSpell: func(obj *Object, spell uint32, mode, fourth, fifth int32) int32 {
				called++
				if obj != target {
					t.Fatalf("target = %p, want %p", obj, target)
				}
				if spell != 0xf1234567 || mode != 1 || fourth != 0 || fifth != 0 {
					t.Fatalf("grant args = %#x/%d/%d/%d", spell, mode, fourth, fifth)
				}
				source.CollideData = unsafe.Pointer(replacement)
				return int32(-0x1234567)
			},
		},
	)
	if got != int32(-0x1234567) {
		t.Fatalf("return = %#x", uint32(got))
	}
	if called != 1 {
		t.Fatalf("grant calls = %d, want 1", called)
	}
	if collision.X != 17.5 || collision.Y != -31.25 {
		t.Fatalf("collision changed: %+v", collision)
	}
}

func TestAwardSpellCollideNative4EAD20GuardAndFaultOrder(t *testing.T) {
	t.Run("nil target precedes nil source", func(t *testing.T) {
		got := awardSpellCollideNative4EAD20(
			nil,
			nil,
			nil,
			awardSpellCollideNativeDeps4EAD20{
				grantSpell: func(*Object, uint32, int32, int32, int32) int32 {
					t.Fatal("grant reached")
					return 0
				},
			},
		)
		if got != 0 {
			t.Fatalf("return = %d, want 0", got)
		}
	})

	t.Run("nil source faults after target guard", func(t *testing.T) {
		var recovered any
		func() {
			defer func() { recovered = recover() }()
			awardSpellCollideNative4EAD20(
				nil,
				&Object{},
				nil,
				awardSpellCollideNativeDeps4EAD20{},
			)
		}()
		if recovered == nil {
			t.Fatal("nil source did not fault")
		}
	})

	t.Run("nil data faults before grant", func(t *testing.T) {
		called := false
		var recovered any
		func() {
			defer func() { recovered = recover() }()
			awardSpellCollideNative4EAD20(
				&Object{},
				&Object{},
				nil,
				awardSpellCollideNativeDeps4EAD20{
					grantSpell: func(*Object, uint32, int32, int32, int32) int32 {
						called = true
						return 0
					},
				},
			)
		}()
		if recovered == nil {
			t.Fatal("nil collide data did not fault")
		}
		if called {
			t.Fatal("grant called after data fault")
		}
	})
}

func TestAwardSpellCollide4EAD20ServerBinding(t *testing.T) {
	data := &AwardSpellCollideData{Spell: 0x80000001}
	source := &Object{CollideData: unsafe.Pointer(data)}
	target := &Object{}
	s := &Server{}

	got := s.AwardSpellCollide4EAD20(
		source,
		target,
		nil,
		AwardSpellCollideRuntime4EAD20{
			GrantSpell: func(obj *Object, spell uint32, mode, fourth, fifth int32) int32 {
				if obj != target || spell != 0x80000001 || mode != 1 || fourth != 0 || fifth != 0 {
					t.Fatalf("grant = %p/%#x/%d/%d/%d", obj, spell, mode, fourth, fifth)
				}
				return -31
			},
		},
	)
	if got != -31 {
		t.Fatalf("return = %d, want -31", got)
	}
	if got := s.AwardSpellCollide4EAD20(nil, nil, nil, AwardSpellCollideRuntime4EAD20{}); got != 0 {
		t.Fatalf("nil-target return = %d, want 0", got)
	}
}
