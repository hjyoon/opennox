package gui

import "testing"

func TestWindowSetEnabledRecursive(t *testing.T) {
	g := New(nil)
	defer g.alloc.Free()

	root := g.NewWindowRaw(nil, StatusEnabled, 0, 0, 100, 100, nil)
	childA := g.NewWindowRaw(root, 0, 0, 0, 20, 20, nil)
	childB := g.NewWindowRaw(root, StatusEnabled, 20, 0, 20, 20, nil)
	grandchild := g.NewWindowRaw(childA, 0, 0, 0, 10, 10, nil)

	if got := root.SetEnabled(true); got != 0 {
		t.Fatalf("SetEnabled(true) = %d, want 0", got)
	}
	for name, win := range map[string]*Window{
		"root": root, "child A": childA, "child B": childB, "grandchild": grandchild,
	} {
		if !win.GetFlags().IsEnabled() {
			t.Errorf("%s was not enabled", name)
		}
	}

	if got := root.SetEnabled(false); got != 0 {
		t.Fatalf("SetEnabled(false) = %d, want 0", got)
	}
	for name, win := range map[string]*Window{
		"root": root, "child A": childA, "child B": childB, "grandchild": grandchild,
	} {
		if win.GetFlags().IsEnabled() {
			t.Errorf("%s was not disabled", name)
		}
	}

	var nilWindow *Window
	if got := nilWindow.SetEnabled(true); got != -2 {
		t.Fatalf("nil SetEnabled(true) = %d, want -2", got)
	}
}
