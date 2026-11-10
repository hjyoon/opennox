package gui

import (
	"reflect"
	"testing"
)

func TestWindowCleanupDeferredFIFOAndOnce(t *testing.T) {
	g := New(nil)
	defer g.DestroyAll()
	w := g.NewWindowRaw(nil, StatusEnabled, 0, 0, 20, 20, nil)
	w.field92 = 0x12345678
	var calls []int
	w.AddCleanup(nil)
	w.AddCleanup(func() {
		if !g.OwnsWindow(w) || w.field92 != 0x12345678 || !w.Flags.Has(StatusDestroyed) {
			t.Error("cleanup did not run while destroyed window storage was still owned")
		}
		calls = append(calls, 1)
	})
	w.AddCleanup(func() { calls = append(calls, 2) })
	w.Destroy()
	w.Destroy()
	w.AddCleanup(func() { calls = append(calls, 99) })
	if len(calls) != 0 {
		t.Fatal("cleanup ran before FreeDestroyed")
	}
	g.FreeDestroyed()
	g.FreeDestroyed()
	if !reflect.DeepEqual(calls, []int{1, 2}) || g.OwnsWindow(w) {
		t.Fatalf("cleanup FIFO/once or pool release: calls=%v", calls)
	}
}

func TestWindowCleanupChildrenAndFreeReentry(t *testing.T) {
	g := New(nil)
	defer g.DestroyAll()
	parent := g.NewWindowRaw(nil, StatusEnabled, 0, 0, 30, 30, nil)
	child := g.NewWindowRaw(parent, StatusEnabled, 0, 0, 10, 10, nil)
	other := g.NewWindowRaw(nil, StatusEnabled, 0, 0, 10, 10, nil)
	var calls []string
	parent.AddCleanup(func() {
		calls = append(calls, "parent1")
		parent.Destroy()
		parent.AddCleanup(func() { calls = append(calls, "late") })
		g.FreeDestroyed()
	})
	parent.AddCleanup(func() { calls = append(calls, "parent2") })
	other.AddCleanup(func() { calls = append(calls, "other") })
	child.AddCleanup(func() {
		calls = append(calls, "child")
		other.Destroy()
		g.FreeDestroyed()
		if !g.OwnsWindow(child) {
			t.Error("recursive free recycled the current child before its cleanup returned")
		}
	})
	parent.Destroy()
	if len(calls) != 0 {
		t.Fatal("child cleanup ran during logical destruction")
	}
	g.FreeDestroyed()
	g.FreeDestroyed()
	if !reflect.DeepEqual(calls, []string{"parent1", "parent2", "child", "other"}) {
		t.Fatalf("child/reentrant cleanup order or repeated call: %v", calls)
	}
	for _, w := range []*Window{parent, child, other} {
		if g.OwnsWindow(w) {
			t.Error("cleaned window remained in the owner registry")
		}
	}
}

func TestWindowCleanupDefersReentrantDestroyAll(t *testing.T) {
	g := New(nil)
	defer g.DestroyAll()
	current := g.NewWindowRaw(nil, StatusEnabled, 0, 0, 10, 10, nil)
	other := g.NewWindowRaw(nil, StatusEnabled, 0, 0, 10, 10, nil)
	current.field92 = 0x76543210
	var late *Window
	var calls []string
	current.AddCleanup(func() {
		calls = append(calls, "current1")
		g.DestroyAll()
		if !g.OwnsWindow(current) || !g.OwnsWindow(other) || current.field92 != 0x76543210 {
			t.Error("DestroyAll recycled or changed storage inside a running cleanup")
		}
		late = g.NewWindowRaw(nil, StatusEnabled, 0, 0, 10, 10, nil)
		late.AddCleanup(func() { calls = append(calls, "late") })
		g.FreeDestroyed()
	})
	current.AddCleanup(func() { calls = append(calls, "current2") })
	other.AddCleanup(func() { calls = append(calls, "other") })
	current.Destroy()
	g.FreeDestroyed()
	if !reflect.DeepEqual(calls, []string{"current1", "current2", "other", "late"}) {
		t.Fatalf("deferred DestroyAll did not drain windows safely: %v", calls)
	}
	for _, w := range []*Window{current, other, late} {
		if g.OwnsWindow(w) {
			t.Error("deferred DestroyAll left an allocated window")
		}
	}
	if g.last != nil || g.free != nil {
		t.Fatal("deferred DestroyAll left a pending window queue")
	}
	g.DestroyAll()
	if len(calls) != 4 {
		t.Fatal("DestroyAll invoked an already consumed cleanup again")
	}
}
