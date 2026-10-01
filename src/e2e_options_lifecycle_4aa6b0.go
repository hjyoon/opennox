package opennox

import (
	"fmt"
	"image"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/gui"
)

// Read-only lifecycle observer. In particular, do not dereference the retained
// Options C globals after close: the original close callback leaves them stale.
type optionsLifecycleSnapshot4AA6B0 struct {
	state       gui.StateID
	root, input *gui.Window
	animation   *gui.Anim
	global      gui.AnimState
}

func (s optionsLifecycleSnapshot4AA6B0) validate(open bool) error {
	if s.input != nil {
		return fmt.Errorf("unexpected input window %p", s.input)
	}
	if !open {
		if s.state != client.StateMainMenu || s.root != nil || s.animation != nil || s.global != gui.AnimInDone {
			return fmt.Errorf("options not closed: state=%d root=%p animation=%p global=%d", s.state, s.root, s.animation, s.global)
		}
		return nil
	}
	if s.state != client.StateOptions || !optionsAuditEnabled(s.root) || s.root.Offs() != (image.Point{}) {
		return fmt.Errorf("options not open: state=%d root=%p", s.state, s.root)
	}
	if s.animation == nil || s.animation.StateID != client.StateOptions || s.animation.Window() != s.root || s.animation.State() != gui.AnimInDone || s.global != gui.AnimInDone {
		return fmt.Errorf("options animation incomplete: animation=%p global=%d", s.animation, s.global)
	}
	return nil
}

func (sc *e2eScenario) CheckMainOptionsLifecycle4AA6B0(mode int, name string) {
	if mode != 0 && mode != 1 {
		panic("invalid main options lifecycle mode")
	}
	sc.add(0, name, func() {
		snapshot := optionsLifecycleSnapshot4AA6B0{
			state: noxClient.GameGetStateCode(), root: noxClient.GUI.ChildByID(300), input: noxClient.GUI.ChildByID(900),
			animation: gui.FindAnimForStateID(client.StateOptions), global: gui.AnimGlobalState(),
		}
		if err := snapshot.validate(mode == 1); err != nil {
			e2eError(err)
		}
		e2eLog.Printf("MAIN OPTIONS LIFECYCLE: open=%t state=%d root=%p animation=%p global=%d", mode == 1, snapshot.state, snapshot.root, snapshot.animation, snapshot.global)
	})
}

func (sc *e2eScenario) ClickMainOptionsBack4AA6B0(name string) {
	new(optionsAudit).clickWindow(sc, name, func() *gui.Window { return noxClient.GUI.ChildByID(152) })
}
