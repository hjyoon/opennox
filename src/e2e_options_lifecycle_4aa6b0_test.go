package opennox

import (
	"image"
	"testing"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/gui"
)

func TestOptionsShow4AA6B0LifecycleObserverReadOnlyAndRejectsInvalidState(t *testing.T) {
	oldGlobal := gui.AnimGlobalState()
	t.Cleanup(func() { gui.SetAnimGlobalState(oldGlobal) })
	g := gui.New(nil)
	t.Cleanup(g.DestroyAll)
	root := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 640, 480, nil)
	other := g.NewWindowRaw(nil, gui.StatusEnabled, 0, 0, 640, 480, nil)
	anim := gui.NewAnim(root, image.Point{}, image.Point{}, image.Point{}, image.Point{})
	t.Cleanup(anim.Free)
	anim.StateID = client.StateOptions
	anim.SetState(gui.AnimInDone)
	snapshot := optionsLifecycleSnapshot4AA6B0{state: client.StateOptions, root: root, animation: anim, global: gui.AnimInDone}
	beforeRoot, beforeAnim, beforeGlobal := *root, *anim, gui.AnimGlobalState()
	if err := snapshot.validate(true); err != nil {
		t.Fatal(err)
	}
	if *root != beforeRoot || *anim != beforeAnim || gui.AnimGlobalState() != beforeGlobal {
		t.Fatal("observer changed native state")
	}
	for _, changed := range []optionsLifecycleSnapshot4AA6B0{
		{state: client.StateMainMenu, root: root, animation: anim, global: gui.AnimInDone},
		{state: client.StateOptions, animation: anim, global: gui.AnimInDone},
		{state: client.StateOptions, root: other, animation: anim, global: gui.AnimInDone},
		{state: client.StateOptions, root: root, global: gui.AnimInDone},
		{state: client.StateOptions, root: root, animation: anim, input: other, global: gui.AnimInDone},
		{state: client.StateOptions, root: root, animation: anim, global: gui.AnimIn},
	} {
		if err := changed.validate(true); err == nil {
			t.Fatalf("accepted invalid open snapshot: %+v", changed)
		}
	}
	anim.SetState(gui.AnimOut)
	if err := snapshot.validate(true); err == nil {
		t.Fatal("accepted unfinished animation")
	}
	anim.SetState(gui.AnimInDone)
	root.SetPos(image.Pt(0, -1))
	if err := snapshot.validate(true); err == nil {
		t.Fatal("accepted offscreen root")
	}
	root.SetPos(image.Point{})
	root.Flags |= gui.StatusHidden
	if err := snapshot.validate(true); err == nil {
		t.Fatal("accepted hidden root")
	}
	closed := optionsLifecycleSnapshot4AA6B0{state: client.StateMainMenu, global: gui.AnimInDone}
	if err := closed.validate(false); err != nil {
		t.Fatal(err)
	}
	for _, changed := range []optionsLifecycleSnapshot4AA6B0{
		{state: client.StateOptions, global: gui.AnimInDone},
		{state: client.StateMainMenu, root: root, global: gui.AnimInDone},
		{state: client.StateMainMenu, animation: anim, global: gui.AnimInDone},
		{state: client.StateMainMenu, input: other, global: gui.AnimInDone},
		{state: client.StateMainMenu, global: gui.AnimOutDone},
	} {
		if err := changed.validate(false); err == nil {
			t.Fatalf("accepted invalid closed snapshot: %+v", changed)
		}
	}
}

func TestOptionsShow4AA6B0LifecycleRejectsInvalidMode(t *testing.T) {
	for _, mode := range []int{-1, 2} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("accepted invalid mode")
				}
			}()
			new(e2eScenario).CheckMainOptionsLifecycle4AA6B0(mode, "invalid test mode")
		}()
	}
}
