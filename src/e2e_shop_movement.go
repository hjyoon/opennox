package opennox

import (
	"fmt"
	"image"

	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

// Send the same real movement input during and after a stock dialog/shop.
// No freeze/unfreeze, control packet, position or velocity is supplied.
func (sc *e2eScenario) CheckShopMovement(dialog, frozen bool, name string) {
	var unit *server.Object
	var original, start, destination types.Pointf
	var clientStart image.Point
	var mouse image.Point
	sc.addWhen(0, name+" open conversation", 1200, func() bool {
		host := noxServer.Players.HostUnit()
		if host == nil || noxClient.ClientPlayerUnit() == nil {
			return false
		}
		if dialog {
			root := legacy.Get_dword_5d4594_1123524()
			return root != nil && !root.GetFlags().IsHidden()
		}
		return host.UpdateDataPlayer().Trade70 != nil
	}, func() {
		unit = noxServer.Players.HostUnit()
		original, start = unit.PosVec, unit.PosVec
		clientStart = noxClient.ClientPlayerUnit().Pos()
		if unit.Flags().Has(object.FlagNoUpdate) != frozen {
			e2eError(fmt.Errorf("conversation freeze = %t, want %t", unit.Flags().Has(object.FlagNoUpdate), frozen))
			return
		}
		if unit.Flags().Has(object.FlagNoUpdate) != noxClient.ClientPlayerUnit().Flags().Has(object.FlagNoUpdate) {
			e2eError(fmt.Errorf("conversation freeze not synchronized to client"))
			return
		}
		var found bool
		for dir := 0; dir < 256; dir += 32 {
			to := start.Add(server.Dir16(dir).Vec().Mul(80))
			if e2eWarriorLaneClear(unit, start, to) {
				destination, found = to, true
				break
			}
		}
		if !found {
			e2eError(fmt.Errorf("no clear conversation movement lane"))
			return
		}
		mouse = noxClient.Viewport().ToScreenPos(image.Pt(int(destination.X), int(destination.Y)))
		e2eLog.Printf("CONVERSATION LOCK PREPARED: dialog=%t player=%p flags=%x client-flags=%x pos=%v trade=%p dialog-with=%p", dialog, unit, unit.ObjFlags, noxClient.ClientPlayerUnit().ObjFlags, start, unit.UpdateDataPlayer().Trade70, unit.UpdateDataPlayer().DialogWith)
		e2eQueueInput(&seat.MouseMoveEvent{Pos: mouse}, &seat.MouseButtonEvent{Button: seat.MouseButtonRight, Pressed: true})
	})
	sc.add(30, name+" movement remains blocked", func() {
		distance := unit.PosVec.Sub(start).Len()
		if frozen && distance > 0.5 || !frozen && distance < 6 {
			e2eError(fmt.Errorf("conversation movement policy failed: frozen=%t %v -> %v distance=%g velocity=%v", frozen, start, unit.PosVec, distance, unit.VelVec))
			return
		}
		if position := noxClient.ClientPlayerUnit().Pos(); frozen && position != clientStart || !frozen && position == clientStart {
			e2eError(fmt.Errorf("client conversation movement policy failed: frozen=%t %v -> %v server=%v", frozen, clientStart, position, unit.PosVec))
			return
		}
		e2eLog.Printf("CONVERSATION MOVEMENT PASS: dialog=%t frozen=%t held-input=30 frames pos=%v velocity=%v", dialog, frozen, unit.PosVec, unit.VelVec)
	})
	sc.Input(1, name+" release", &seat.MouseButtonEvent{Button: seat.MouseButtonRight, Pressed: false})
	if dialog {
		sc.ClickNPCDialogDone(name + " actual Done")
	} else {
		sc.Click(image.Pt(786, 752), seat.MouseButtonLeft, name+" actual shop exit")
	}
	sc.addWhen(0, name+" ordinary close acknowledgement", 1200, func() bool {
		return !unit.Flags().Has(object.FlagNoUpdate) && !noxClient.ClientPlayerUnit().Flags().Has(object.FlagNoUpdate) && unit.UpdateDataPlayer().Trade70 == nil && unit.UpdateDataPlayer().DialogWith == nil
	}, func() {})
	sc.Wait(10, name+" settle closed UI")
	sc.add(0, name+" same movement input after conversation", func() {
		start = unit.PosVec
		// Non-Coop shops legitimately allow movement while open. The old
		// destination may already have been reached; observe another clear
		// lane from the current position without relocating the player.
		found := false
		for dir := 0; dir < 256; dir += 32 {
			to := start.Add(server.Dir16(dir).Vec().Mul(80))
			if e2eWarriorLaneClear(unit, start, to) {
				destination, found = to, true
				break
			}
		}
		if !found {
			e2eError(fmt.Errorf("no clear post-conversation movement lane"))
			return
		}
		mouse = noxClient.Viewport().ToScreenPos(image.Pt(int(destination.X), int(destination.Y)))
		e2eQueueInput(&seat.MouseMoveEvent{Pos: mouse}, &seat.MouseButtonEvent{Button: seat.MouseButtonRight, Pressed: true})
	})
	sc.add(12, name+" verify movement restored", func() {
		if distance := unit.PosVec.Sub(start).Len(); distance < 6 {
			e2eError(fmt.Errorf("same movement input ineffective after conversation: %v -> %v", start, unit.PosVec))
			return
		}
		e2eLog.Printf("CONVERSATION UNLOCK PASS: dialog=%t pos=%v->%v input-effective=true", dialog, start, unit.PosVec)
	})
	sc.Input(1, name+" release restored movement", &seat.MouseButtonEvent{Button: seat.MouseButtonRight, Pressed: false})
	sc.Wait(12, name+" settle movement")
	sc.add(0, name+" cleanup position", func() { asObjectS(unit).SetPos(original) })
}
