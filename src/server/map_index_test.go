package server

import (
	"image"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestMapIndexNativeSidecar(t *testing.T) {
	var m serverMap
	m.Init()
	t.Cleanup(m.Free)

	pos := types.Ptf(170, 255)
	a := &Object{ObjClass: object.ClassMissile, ObjFlags: object.FlagActive, NewPos: pos}
	b := &Object{ObjClass: object.ClassMissile, ObjFlags: object.FlagActive, NewPos: pos}
	m.AddObjectToIndex(a)
	m.AddObjectToIndex(b)

	bucket := m.DebugIndex()[image.Pt(2, 3)]
	if len(bucket.List) != 2 || bucket.List[0].Obj != b || bucket.List[1].Obj != a {
		t.Fatalf("base index order = %#v, want b then a", bucket.List)
	}
	if len(m.objectIndex) != 2 {
		t.Fatalf("native sidecars = %d, want 2", len(m.objectIndex))
	}

	m.RemoveObjectFromIndex(b)
	if b.Flags().Has(object.FlagPartitioned) {
		t.Fatal("partition flag was not cleared")
	}
	bucket = m.DebugIndex()[image.Pt(2, 3)]
	if len(bucket.List) != 1 || bucket.List[0].Obj != a {
		t.Fatalf("base index after removal = %#v, want a", bucket.List)
	}

	m.releaseIndexState(b)
	if _, ok := m.objectIndex[b]; ok {
		t.Fatal("released object still has native sidecar")
	}
	m.releaseIndexState(a)
	if _, ok := m.objectIndex[a]; ok {
		t.Fatal("partitioned object still has native sidecar after release")
	}
	if _, ok := m.DebugIndex()[image.Pt(2, 3)]; ok {
		t.Fatal("partitioned object remained linked after release")
	}
}

func TestMapIndexLargeObjectParts(t *testing.T) {
	var m serverMap
	m.Init()
	t.Cleanup(m.Free)

	obj := &Object{ObjFlags: object.FlagActive, NewPos: types.Ptf(170, 170)}
	obj.Shape.Kind = ShapeKindCircle
	obj.Shape.Circle.R = 50
	m.AddObjectToIndex(obj)

	state := m.objectIndex[obj]
	if state == nil || state.Cur != uint32(len(state.Parts)) {
		t.Fatalf("part count = %v, want %d", state, len(state.Parts))
	}
	want := []image.Point{{X: 1, Y: 1}, {X: 2, Y: 1}, {X: 3, Y: 1}, {X: 1, Y: 2}}
	for i, p := range want {
		if got := image.Pt(int(state.Parts[i].X), int(state.Parts[i].Y)); got != p {
			t.Fatalf("part %d = %v, want %v", i, got, p)
		}
	}

	m.RemoveObjectFromIndex(obj)
	if state.Cur != 0 {
		t.Fatalf("part count after removal = %d, want 0", state.Cur)
	}
	for pos, bucket := range m.DebugIndex() {
		if len(bucket.List) != 0 || len(bucket.Parts) != 0 {
			t.Fatalf("bucket %v remains linked: %#v", pos, bucket)
		}
	}
}
