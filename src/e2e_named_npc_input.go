package opennox

import (
	"fmt"
	"image"

	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/legacy"
)

type e2eNamedNPCInputTarget struct {
	wire    uint16
	point   image.Point
	hovered bool
}

// InteractNamedNPC follows a map-loaded NPC on screen using ordinary mouse
// input. It never changes world, NPC, player or camera state, starts a dialog,
// calls a script, or supplies a cursor target. The normal client must select
// the named drawable at the real mouse position before the single click.
func (sc *e2eScenario) InteractNamedNPC(id string, offset image.Point, name string) {
	sc.interactNamedNPCInput(name, func() (e2eNamedNPCInputTarget, bool) {
		if noxServer == nil || noxClient == nil || noxClient.ClientPlayerUnit() == nil || noxClient.Viewport() == nil {
			return e2eNamedNPCInputTarget{}, false
		}
		npc := noxServer.Objs.GetObjectByID(id)
		if npc == nil || !npc.Class().Has(object.ClassMonster) || npc.UpdateData == nil ||
			npc.Flags().HasAny(object.FlagDead|object.FlagDestroyed) || !npc.Flags().Has(object.FlagEnabled) {
			return e2eNamedNPCInputTarget{}, false
		}
		if npc.UpdateDataMonster().DialogStartFunc < 0 {
			return e2eNamedNPCInputTarget{}, false
		}
		code := noxServer.GetUnitNetCode(npc)
		if code <= 0 || code > int(^uint16(0)) {
			return e2eNamedNPCInputTarget{}, false
		}
		dr := noxClient.Objs.ByNetCode(uint16(code))
		if dr == nil {
			return e2eNamedNPCInputTarget{}, false
		}
		point := noxClient.Viewport().ToScreenPos(dr.Pos()).Add(offset)
		if !point.In(noxClient.Viewport().Screen) {
			return e2eNamedNPCInputTarget{}, false
		}
		hovered := legacy.Nox_xxx_clientGetSpriteAtCursor_476F90() == dr
		if hovered {
			point = noxClient.Inp.GetMousePos()
		}
		return e2eNamedNPCInputTarget{wire: uint16(code), point: point, hovered: hovered}, true
	}, e2eQueueInput, func(target e2eNamedNPCInputTarget) {
		e2eLog.Printf("NAMED NPC ACTUAL INPUT: id=%q wire=%#x point=%v frame=%d", id, target.wire, target.point, noxServer.Frame())
	})
}

func (sc *e2eScenario) interactNamedNPCInput(name string, read func() (e2eNamedNPCInputTarget, bool), queue func(...seat.InputEvent), clicked func(e2eNamedNPCInputTarget)) {
	var target e2eNamedNPCInputTarget
	sc.addWhen(0, name+" actual mouse target", 600, func() bool {
		current, ok := read()
		if !ok {
			return false
		}
		if !current.hovered {
			queue(&seat.MouseMoveEvent{Pos: current.point, Relative: false})
			return false
		}
		target = current
		return true
	}, func() {
		if target.wire == 0 {
			e2eError(fmt.Errorf("named NPC input lost its observed target"))
			return
		}
		queue(&seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true})
		clicked(target)
	})
	sc.add(1, name+" release actual mouse", func() {
		queue(&seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false})
	})
}
