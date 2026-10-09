package opennox

import (
	"fmt"
	"image"
	"unsafe"

	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

// Observe a map-loaded NPC's stock greeting through real mouse input and the
// normal server audio path. Only the player's setup position is changed; no
// sound, AI action, cursor target, or NPC property is supplied by the test.
func (sc *e2eScenario) CheckNPCHoverSound(id, name string) {
	var npc, host *server.Object
	var original types.Pointf
	var wire uint16
	var greeting sound.ID
	var sounds, heldFrames int
	var holding, done bool
	sc.addWhen(0, name+" stock setup", 1200, func() bool {
		return noxServer.Players.HostUnit() != nil && noxServer.Objs.GetObjectByID(id) != nil
	}, func() {
		npc, host = noxServer.Objs.GetObjectByID(id), noxServer.Players.HostUnit()
		if !npc.Class().Has(object.ClassMonster) || npc.UpdateData == nil || npc.Field5&0x10 == 0 ||
			npc.Flags().HasAny(object.FlagDead|object.FlagDestroyed) ||
			host.UpdateDataPlayer().Trade70 != nil || host.UpdateDataPlayer().DialogWith != nil {
			e2eError(fmt.Errorf("stock hover NPC %q is unavailable: %v", id, npc))
			return
		}
		set := npc.UpdateDataMonster().SoundSet122
		if set == nil {
			e2eError(fmt.Errorf("stock hover NPC %q has no sound set", id))
			return
		}
		greeting = sound.ID(*(*uint32)(unsafe.Add(set, 4)))
		if greeting == 0 || greeting >= 1024 || noxServer.Audio.Field12(greeting) <= 0 {
			e2eError(fmt.Errorf("stock hover NPC %q has no enabled greeting: %d", id, greeting))
			return
		}
		code := noxServer.GetUnitNetCode(npc)
		if code <= 0 || code > int(^uint16(0)) {
			e2eError(fmt.Errorf("stock hover NPC %q has invalid wire code: %d", id, code))
			return
		}
		wire, original = uint16(code), host.PosVec
		noxServer.Audio.OnSound(func(snd sound.ID, _ int, source *server.Object, _ types.Pointf) {
			if !done && source == npc && snd == greeting {
				sounds++
				e2eLog.Printf("NPC HOVER GREETING: id=%s sound=%s/%d frame=%d count=%d", id, greeting, greeting, noxServer.Frame(), sounds)
			}
		})
		noxServer.TickHook(func() {
			if done || !holding {
				return
			}
			dr := noxClient.Objs.ByNetCode(wire)
			if dr == nil {
				return
			}
			pos := noxClient.Viewport().ToScreenPos(dr.Pos())
			if pos != noxClient.Inp.GetMousePos() {
				e2eQueueInput(&seat.MouseMoveEvent{Pos: pos})
			}
			if legacy.Nox_xxx_clientGetSpriteAtCursor_476F90() == dr && npc.UpdateDataMonster().HasAction(ai.DEPENDENCY_UNDER_CURSOR) {
				heldFrames++
			}
		})
		asObjectS(host).SetPos(npc.PosVec.Sub(types.Ptf(48, 0)))
		e2eQueueInput(&seat.MouseMoveEvent{Pos: image.Pt(10, 10)})
		e2eLog.Printf("NPC HOVER STOCK SETUP: id=%s npc=%p host=%p wire=%#x sound=%s/%d stack=%v", id, npc, host, wire, greeting, greeting, npc.UpdateDataMonster().GetAIStack())
	})
	sc.Wait(90, name+" settle real map and viewport")
	sc.add(0, name+" verify no greeting away from NPC", func() {
		if npc == nil || sounds != 0 || npc.UpdateDataMonster().HasAction(ai.DEPENDENCY_UNDER_CURSOR) {
			e2eError(fmt.Errorf("hover setup started with a conversation: NPC=%p sounds=%d", npc, sounds))
		}
	})
	for entry := 1; entry <= 2; entry++ {
		sc.addWhen(0, fmt.Sprintf("%s entry %d mouse", name, entry), 1200, func() bool {
			dr := noxClient.Objs.ByNetCode(wire)
			return dr != nil && noxClient.Viewport().ToScreenPos(dr.Pos()).In(noxClient.Viewport().Screen)
		}, func() {
			heldFrames, holding = 0, true
			dr := noxClient.Objs.ByNetCode(wire)
			e2eQueueInput(&seat.MouseMoveEvent{Pos: noxClient.Viewport().ToScreenPos(dr.Pos())})
		})
		sc.addWhen(1, fmt.Sprintf("%s entry %d actual target", name, entry), 600, func() bool {
			dr := noxClient.Objs.ByNetCode(wire)
			return dr != nil && legacy.Nox_xxx_clientGetSpriteAtCursor_476F90() == dr &&
				npc.UpdateDataMonster().HasAction(ai.DEPENDENCY_UNDER_CURSOR)
		}, func() {})
		sc.Wait(180, fmt.Sprintf("%s entry %d held mouse", name, entry))
		sc.add(0, fmt.Sprintf("%s entry %d one greeting", name, entry), func() {
			if sounds != entry || heldFrames < 180 || !npc.UpdateDataMonster().HasAction(ai.ACTION_WAIT_RELATIVE) {
				e2eError(fmt.Errorf("NPC %q hover entry %d: sounds=%d held=%d stack=%v", id, entry, sounds, heldFrames, npc.UpdateDataMonster().GetAIStack()))
				return
			}
			e2eLog.Printf("NPC HOVER HOLD VERIFIED: id=%s entry=%d total-greetings=%d held-frames=%d no-replay=true stack=%v", id, entry, sounds, heldFrames, npc.UpdateDataMonster().GetAIStack())
			holding = false
			e2eQueueInput(&seat.MouseMoveEvent{Pos: image.Pt(10, 10)})
		})
		sc.Wait(90, fmt.Sprintf("%s entry %d leave mouse", name, entry))
		sc.add(0, fmt.Sprintf("%s entry %d natural wait completion", name, entry), func() {
			if sounds != entry || npc.UpdateDataMonster().HasAction(ai.DEPENDENCY_UNDER_CURSOR) || npc.UpdateDataMonster().HasAction(ai.ACTION_WAIT_RELATIVE) {
				e2eError(fmt.Errorf("NPC %q leave entry %d: sounds=%d stack=%v", id, entry, sounds, npc.UpdateDataMonster().GetAIStack()))
				return
			}
			e2eLog.Printf("NPC HOVER LEAVE VERIFIED: id=%s entry=%d no-greeting-away=true wait-expired-naturally=true", id, entry)
		})
	}
	sc.add(0, name+" restore setup position", func() {
		done, holding = true, false
		asObjectS(host).SetPos(original)
	})
}
