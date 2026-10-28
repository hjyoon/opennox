package server

import (
	"image"
	"testing"

	"github.com/opennox/libs/types"
)

func TestBlowRawPointerBoundaries53C160(t *testing.T) {
	for _, row := range []struct {
		name              string
		root              bool
		candidate, source *Object
		fault             bool
	}{
		{name: "nil-root", root: true, fault: true},
		{name: "nil-candidate", source: &Object{}, fault: true},
		{name: "eligible-nil-source", candidate: &Object{}, fault: true},
		{name: "dead-nil-source", candidate: &Object{ObjFlags: 0x8000}, fault: true},
		{name: "destroyed-nil-source", candidate: &Object{ObjFlags: 0x20}},
		{name: "immobile-nil-source", candidate: &Object{ObjClass: 0x400000}},
		{name: "destroyed-immobile-nil-source", candidate: &Object{ObjFlags: 0x20, ObjClass: 0x400000}},
		{name: "disabled-root", root: true, source: &Object{}},
	} {
		t.Run(row.name, func(t *testing.T) {
			defer func() {
				if got := recover() != nil; got != row.fault {
					t.Fatalf("required-pointer fault %v, original requires %v", got, row.fault)
				}
			}()
			if row.root {
				BlowUpdateRaw53C160(row.source, BlowUpdateRuntime53C160{})
			} else {
				BlowCandidateRaw53C240(row.candidate, row.source, BlowUpdateRuntime53C160{})
			}
		})
	}
}

func TestBlowGuardedNilInputs53C160(t *testing.T) {
	var world *Server
	world.BlowUpdate53C160(nil)
	blowUpdateNative53C160(nil, blowUpdateDeps53C160{})
	blowUpdateNative53C160(&Object{ObjFlags: 0x1000000}, blowUpdateDeps53C160{
		indexedDirection: func(int16) image.Point { return image.Pt(1, 0) },
		eachInRect: func(_ types.Rectf, fn func(*Object) bool) {
			if !fn(nil) {
				t.Fatal("guarded nil candidate stopped the scan")
			}
		},
		canInteract:     func(*Object, *Object) bool { t.Fatal("nil candidate interaction"); return false },
		directionVector: func(byte) (float32, float32) { t.Fatal("nil candidate vector"); return 0, 0 },
	})
}
