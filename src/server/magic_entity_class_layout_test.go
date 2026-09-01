package server

import (
	"testing"
	"unsafe"
)

func TestMagicEntityClassNativeLayout(t *testing.T) {
	type layout struct {
		size, object, spells, spellIndex, trap, phonemeRoot uintptr
		phonemeIndex, frame, delay, targetMode, next, prev  uintptr
	}
	wantByPointerSize := map[uintptr]layout{
		4: {
			size: 60, object: 4, spells: 8, spellIndex: 28, trap: 29,
			phonemeRoot: 32, phonemeIndex: 36, frame: 40, delay: 44,
			targetMode: 48, next: 52, prev: 56,
		},
		8: {
			size: 80, object: 8, spells: 16, spellIndex: 36, trap: 37,
			phonemeRoot: 40, phonemeIndex: 48, frame: 52, delay: 56,
			targetMode: 60, next: 64, prev: 72,
		},
	}
	pointerSize := unsafe.Sizeof(uintptr(0))
	want, ok := wantByPointerSize[pointerSize]
	if !ok {
		t.Fatalf("unsupported pointer size %d", pointerSize)
	}
	got := layout{
		size:         unsafe.Sizeof(MagicEntityClass{}),
		object:       unsafe.Offsetof(MagicEntityClass{}.Obj4),
		spells:       unsafe.Offsetof(MagicEntityClass{}.Spells8),
		spellIndex:   unsafe.Offsetof(MagicEntityClass{}.SpellInd28),
		trap:         unsafe.Offsetof(MagicEntityClass{}.Field29),
		phonemeRoot:  unsafe.Offsetof(MagicEntityClass{}.Field32),
		phonemeIndex: unsafe.Offsetof(MagicEntityClass{}.Field36),
		frame:        unsafe.Offsetof(MagicEntityClass{}.Frame40),
		delay:        unsafe.Offsetof(MagicEntityClass{}.Field44),
		targetMode:   unsafe.Offsetof(MagicEntityClass{}.Field48),
		next:         unsafe.Offsetof(MagicEntityClass{}.Next52),
		prev:         unsafe.Offsetof(MagicEntityClass{}.Prev56),
	}
	if got != want {
		t.Fatalf("MagicEntityClass layout = %+v, want %+v", got, want)
	}
	if got := unsafe.Sizeof(MagicEntityClass{}.Spells8[0]); got != 4 {
		t.Fatalf("spell ID width = %d, want 4", got)
	}
}
