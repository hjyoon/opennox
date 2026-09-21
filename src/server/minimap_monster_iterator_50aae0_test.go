package server

import (
	"testing"

	"github.com/opennox/libs/object"
)

func TestMinimapMonsterIterator50AAE0(t *testing.T) {
	last := &Object{ObjClass: object.ClassMonster | object.ClassImmobile}
	middle := &Object{ObjNext: last}
	firstMonster := &Object{ObjClass: object.ClassMonster, ObjNext: middle}
	first := &Object{ObjClass: object.ClassPlayer, ObjNext: firstMonster}

	var it MinimapMonsterIterator50AAE0
	if got := it.First(first); got != firstMonster {
		t.Fatalf("First = %p, want first monster %p", got, firstMonster)
	}
	if got := it.Next(); got != last {
		t.Fatalf("Next = %p, want last monster %p", got, last)
	}
	if got := it.Next(); got != nil {
		t.Fatalf("Next after end = %p, want nil", got)
	}
	if got := it.Next(); got != nil {
		t.Fatalf("repeated Next after end = %p, want nil", got)
	}
	if got := it.First(first); got != firstMonster {
		t.Fatalf("restarted First = %p, want %p", got, firstMonster)
	}
	if got := it.First(nil); got != nil {
		t.Fatalf("First(nil) = %p, want nil", got)
	}
}

func TestMinimapMonsterIterator50AAE0ReadsLiveSuccessor(t *testing.T) {
	originalNext := &Object{ObjClass: object.ClassMonster}
	first := &Object{ObjClass: object.ClassMonster, ObjNext: originalNext}
	replacement := &Object{ObjClass: object.ClassMonster}

	var it MinimapMonsterIterator50AAE0
	if got := it.First(first); got != first {
		t.Fatalf("First = %p, want %p", got, first)
	}
	first.ObjNext = replacement
	if got := it.Next(); got != replacement {
		t.Fatalf("Next after list mutation = %p, want live successor %p", got, replacement)
	}
}
