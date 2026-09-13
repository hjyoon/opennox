package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestSpellOvalShield531490NativeRecordLayout(t *testing.T) {
	wantPos, wantTarget, wantFrame := uintptr(28), uintptr(48), uintptr(68)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantPos, wantTarget, wantFrame = 48, 72, 96
	}
	for _, check := range []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Pos", unsafe.Offsetof(DurSpell{}.Pos), wantPos},
		{"Target48", unsafe.Offsetof(DurSpell{}.Target48), wantTarget},
		{"Frame68", unsafe.Offsetof(DurSpell{}.Frame68), wantFrame},
	} {
		if check.got != check.want {
			t.Errorf("%s offset = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestSpellOvalShield531490ServerBindingNativePointersAndPosition(t *testing.T) {
	s := &Server{}
	s.SetTickRate(30)
	s.SetFrame(0xfffffff0)
	target := &Object{ObjClass: object.Class(4)}
	record := &DurSpell{Target48: target, Level: 2, Pos: types.Pointf{X: 3, Y: 4}}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(target)) <= 0xffffffff {
		t.Fatal("expected a target pointer above 4 GiB")
	}
	applyCalls := 0
	removeCalls := 0
	runtime := SpellOvalShieldRuntime531490{
		ApplyBuff: func(got *Object, buff int32, duration int16, power int8) {
			applyCalls++
			if got != target || buff != 27 || duration != 1200 || power != 2 {
				t.Errorf("apply = %p/%d/%d/%d", got, buff, duration, power)
			}
		},
		BuffOff: func(got *Object, buff int32) {
			removeCalls++
			if got != target || buff != 27 {
				t.Errorf("buff-off = %p/%d", got, buff)
			}
		},
	}
	if got := s.SpellOvalShieldCreate531490(record, runtime); got != 0 {
		t.Fatalf("create = %d, want 0", got)
	}
	if applyCalls != 1 || record.Frame68 != 1184 {
		t.Fatalf("create effects = %d calls, frame %d", applyCalls, record.Frame68)
	}
	target.ObjClass = object.Class(2)
	target.PosVec = record.Pos
	if got := s.SpellOvalShieldUpdate5314F0(record); got != 0 {
		t.Fatalf("stationary monster update = %d, want 0", got)
	}
	target.PosVec.X += 5
	if got := s.SpellOvalShieldUpdate5314F0(record); got != 1 {
		t.Fatalf("moved monster update = %d, want 1", got)
	}
	target.ObjClass = object.Class(4)
	target.Buffs = 1 << 8
	if got := s.SpellOvalShieldUpdate5314F0(record); got != 1 {
		t.Fatalf("blocked buff update = %d, want 1", got)
	}
	target.Buffs = 0
	target.ObjFlags = object.Flags(0x8000)
	if got := s.SpellOvalShieldUpdate5314F0(record); got != 1 {
		t.Fatalf("removed target update = %d, want 1", got)
	}
	s.SpellOvalShieldDestroy531560(record, runtime)
	if removeCalls != 1 {
		t.Fatalf("buff-off calls = %d, want 1", removeCalls)
	}
}
