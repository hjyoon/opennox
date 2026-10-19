package opennox

import (
	"fmt"
	"image"

	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/legacy"
)

type e2eNamedUseInputTarget struct {
	wire    uint16
	point   image.Point
	hovered bool
	cursor  gui.Cursor
}

// UseNamedObject resolves a map-loaded object each frame and clicks only when
// the normal client selects that exact drawable with its Use cursor. Server
// and drawable pointers remain local to each observation. Static object codes
// pass through the normal client lookup, including their high network bit.
// The action only queues ordinary mouse motion and a left press/release.
func (sc *e2eScenario) UseNamedObject(id string, offset image.Point, name string) {
	sc.useNamedObjectInput(name, func() (e2eNamedUseInputTarget, bool) {
		if noxServer == nil || noxClient == nil || noxClient.ClientPlayerUnit() == nil || noxClient.Viewport() == nil {
			return e2eNamedUseInputTarget{}, false
		}
		obj := noxServer.Objs.GetObjectByID(id)
		if obj == nil || obj.Flags().HasAny(object.FlagDead|object.FlagDestroyed) || !obj.Flags().Has(object.FlagEnabled) {
			return e2eNamedUseInputTarget{}, false
		}
		code := noxServer.GetUnitNetCode(obj)
		if code <= 0 || code > int(^uint16(0)) {
			return e2eNamedUseInputTarget{}, false
		}
		dr := noxClient.Objs.ByNetCode(uint16(code))
		if dr == nil {
			return e2eNamedUseInputTarget{}, false
		}
		point := noxClient.Viewport().ToScreenPos(dr.Pos()).Add(offset)
		if !point.In(noxClient.Viewport().Screen) {
			return e2eNamedUseInputTarget{}, false
		}
		hovered := legacy.Nox_xxx_clientGetSpriteAtCursor_476F90() == dr
		if hovered {
			point = noxClient.Inp.GetMousePos()
		}
		return e2eNamedUseInputTarget{
			wire: uint16(code), point: point, hovered: hovered, cursor: noxClient.Nox_client_getCursorType(),
		}, true
	}, e2eQueueInput, func(target e2eNamedUseInputTarget) {
		e2eLog.Printf("NAMED USE ACTUAL INPUT: id=%q wire=%#x point=%v cursor=%d frame=%d",
			id, target.wire, target.point, target.cursor, noxServer.Frame())
	})
}

func (sc *e2eScenario) useNamedObjectInput(name string, read func() (e2eNamedUseInputTarget, bool), queue func(...seat.InputEvent), clicked func(e2eNamedUseInputTarget)) {
	var target e2eNamedUseInputTarget
	sc.addWhen(0, name+" actual Use cursor", 600, func() bool {
		current, ok := read()
		if !ok || current.wire == 0 {
			return false
		}
		if !current.hovered || current.cursor != gui.CursorUse {
			queue(&seat.MouseMoveEvent{Pos: current.point, Relative: false})
			return false
		}
		target = current
		return true
	}, func() {
		if target.wire == 0 {
			e2eError(fmt.Errorf("named use input lost its observed target"))
			return
		}
		queue(&seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true})
		clicked(target)
	})
	sc.add(1, name+" release actual mouse", func() {
		queue(&seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false})
	})
}
