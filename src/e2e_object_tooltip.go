package opennox

import (
	"fmt"
	"image"

	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func e2eObjectTooltipText() string {
	return alloc.GoString16((*uint16)(memmap.PtrOff(0x5D4594, 1096676)))
}

// CheckGroundItemTooltip uses stock server objects and real mouse input. It
// never calls the tooltip builder or seeds the drawable's book cache/text.
func (sc *e2eScenario) CheckGroundItemTooltip(typeID, creature, text string, parameter int, offset image.Point, name string) {
	var wire uint16
	var bookID uint32
	sc.addWhen(0, name+" spawn", 1200, func() bool {
		return noxServer.Players.HostUnit() != nil
	}, func() {
		item := noxServer.NewObjectByTypeID(typeID)
		if item == nil || item.Pickup.Ptr == nil {
			e2eError(fmt.Errorf("tooltip fixture %q is not a stock pickup object: %p", typeID, item))
			return
		}
		sm := noxClient.Strings()
		if item.Class().Has(object.ClassInfoBook) {
			if int(sm.Lang()) != 8 || item.UseData.Ptr == nil {
				e2eError(fmt.Errorf("tooltip book fixture needs Korean strings and stock use data: language=%d object=%p data=%p", sm.Lang(), item, item.UseData.Ptr))
				return
			}
			switch {
			case item.SubClass().AsBook().Has(object.BookSpell):
				if parameter <= 0 || parameter >= 137 {
					e2eError(fmt.Errorf("invalid tooltip spell ID %d", parameter))
					return
				}
				item.UseDataSpellReward().Spell = byte(parameter)
				bookID = uint32(parameter)
				title, ok := legacy.Nox_xxx_spellTitle_424930(parameter)
				if !ok {
					e2eError(fmt.Errorf("missing stock spell title %d", parameter))
					return
				}
				if text == "" {
					text = sm.GetStringInFile("BookOf", `C:\NoxPost\src\client\Gui\ToolTip.c`) + " " + title
				}
			case item.SubClass().AsBook().Has(object.BookFieldGuide):
				bookID = uint32(server.RewardFieldGuideID4F0D20(creature))
				if bookID == 0 || bookID >= 41 {
					e2eError(fmt.Errorf("invalid tooltip creature %q", creature))
					return
				}
				item.UseDataFieldGuide().SetCreature(creature)
				if text == "" {
					text = sm.GetStringInFile(strman.ID("creature:"+creature), `C:\NoxPost\src\common\Magic\ComGuide.c`) + " " +
						sm.GetStringInFile("LoreScroll", `C:\NoxPost\src\client\Gui\ToolTip.c`)
				}
			case item.SubClass().AsBook().Has(object.BookAbility):
				if parameter <= 0 || parameter >= 6 {
					e2eError(fmt.Errorf("invalid tooltip ability ID %d", parameter))
					return
				}
				item.UseDataAbilityReward().Ability = byte(parameter)
				bookID = uint32(parameter)
				if text == "" {
					text = sm.GetStringInFile("BookOf", `C:\NoxPost\src\client\Gui\ToolTip.c`) + " " + legacy.Nox_xxx_abilityGetName_0_425260(parameter)
				}
			}
		}
		if text == "" {
			e2eError(fmt.Errorf("tooltip fixture %q has no expected text", typeID))
			return
		}
		pos := noxServer.Players.HostUnit().Pos().Add(types.Ptf(float32(offset.X), float32(offset.Y)))
		noxServer.CreateObjectAt(item, nil, pos)
		noxServer.ObjectsAddPending()
		code := noxServer.GetUnitNetCode(item)
		if code <= 0 || code > 0xFFFF {
			e2eError(fmt.Errorf("tooltip fixture %q has invalid wire code %#x", typeID, code))
			return
		}
		wire = uint16(code)
		e2eLog.Printf("TOOLTIP FIXTURE: item=%s object=%p wire=%#x book=%d expected=%q", typeID, item, wire, bookID, text)
	})
	sc.addWhen(1, name+" mouse", 1200, func() bool {
		return wire != 0 && noxClient.Objs.ByNetCode(wire) != nil
	}, func() {
		dr := noxClient.Objs.ByNetCode(wire)
		pos := noxClient.Viewport().ToScreenPos(dr.Pos())
		if !pos.In(noxClient.Viewport().Screen) || (bookID != 0 && dr.UnionEffect().Field_108 != 0) {
			e2eError(fmt.Errorf("tooltip fixture was not visible with a cold book cache: item=%s drawable=%p screen=%v cache=%d", typeID, dr, pos, dr.UnionEffect().Field_108))
			return
		}
		e2eLog.Printf("TOOLTIP MOUSE: item=%s drawable=%p wire=%#x screen=%v cold_book=%t", typeID, dr, wire, pos, bookID != 0)
		e2eQueueInput(&seat.MouseMoveEvent{Pos: pos, Relative: false})
	})
	var lastObservation string
	sc.addWhen(1, name+" text", 1200, func() bool {
		dr := noxClient.Objs.ByNetCode(wire)
		target := legacy.Nox_xxx_clientGetSpriteAtCursor_476F90()
		observation := fmt.Sprintf("target=%p cursor=%d text=%q", target, noxClient.Nox_client_getCursorType(), e2eObjectTooltipText())
		if observation != lastObservation {
			e2eLog.Printf("TOOLTIP OBSERVE: item=%s want=%p %s", typeID, dr, observation)
			lastObservation = observation
		}
		return dr != nil && target == dr &&
			noxClient.Nox_client_getCursorType() == gui.CursorPickup && noxClient.Inp.GetDistSlow() &&
			e2eObjectTooltipText() == text && (bookID == 0 || dr.UnionEffect().Field_108 == bookID)
	}, func() {
		dr := noxClient.Objs.ByNetCode(wire)
		e2eLog.Printf("TOOLTIP VERIFIED: item=%s drawable=%p wire=%#x type=%d class=%#x cursor=%d text=%q book_reply=%d", typeID, dr, wire, dr.TypeIDVal, uint32(dr.ObjClass), noxClient.Nox_client_getCursorType(), e2eObjectTooltipText(), bookID)
	})
}

func (sc *e2eScenario) CheckObjectTooltipCleared(name string) {
	sc.Input(0, name+" mouse", &seat.MouseMoveEvent{Pos: image.Pt(10, 10)})
	sc.addWhen(1, name, 600, func() bool {
		return legacy.Nox_xxx_clientGetSpriteAtCursor_476F90() == nil && e2eObjectTooltipText() == ""
	}, func() { e2eLog.Printf("TOOLTIP CLEARED: target=nil text=%q", e2eObjectTooltipText()) })
}

func e2eVisibleHoverType(typeID string) *client.Drawable {
	typ := noxClient.Things.TypeByID(typeID)
	if typ == nil {
		return nil
	}
	for _, dr := range noxClient.Objs.AllList1() {
		if int(dr.TypeIDVal) == typ.Index() && noxClient.Viewport().ToScreenPos(dr.Pos()).In(noxClient.Viewport().Screen) {
			return dr
		}
	}
	return nil
}

func (sc *e2eScenario) CheckObjectHoverCursor(typeID string, cursor int, name string) {
	var target *client.Drawable
	sc.addWhen(0, name+" mouse", 1200, func() bool { return e2eVisibleHoverType(typeID) != nil }, func() {
		target = e2eVisibleHoverType(typeID)
		e2eQueueInput(&seat.MouseMoveEvent{Pos: noxClient.Viewport().ToScreenPos(target.Pos())})
	})
	sc.addWhen(1, name, 600, func() bool {
		return legacy.Nox_xxx_clientGetSpriteAtCursor_476F90() == target && int(noxClient.Nox_client_getCursorType()) == cursor && e2eObjectTooltipText() == ""
	}, func() {
		e2eLog.Printf("OBJECT HOVER VERIFIED: type=%s drawable=%p wire=%#x cursor=%d tooltip=%q", typeID, target, target.NetCode32, cursor, e2eObjectTooltipText())
	})
}
