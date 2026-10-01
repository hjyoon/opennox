package opennox

import (
	"fmt"
	"image"
	"math"
	"unsafe"

	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

// A generous phase-independent bound on the ordinary player regeneration in
// sub_4F9ED0. Exceeding this bound proves the equipped modifier contributed;
// merely seeing any HP increase would also pass with a broken item callback.
func e2ePlayerPassiveHealingBound(elapsed, fps, interval uint32) uint32 {
	if elapsed <= fps {
		return 0
	}
	return uint32((uint64(elapsed-fps)+uint64(interval)-1)/uint64(interval)) + 1
}

type e2ePlayerRegenerationFixture struct {
	unit, item              *server.Object
	modifier                *server.ModifierEff
	frame, fps, interval    uint32
	passiveInterval, window uint32
	maximum, before, saved  uint16
	keepaliveFrame          uint32
	keepalivePressed        bool
}

func (f *e2ePlayerRegenerationFixture) prepare() {
	f.unit, f.item, f.modifier = noxServer.Players.HostUnit(), e2e.engageItem, e2e.engageModifier
	if f.unit == nil || f.unit != e2e.engageOwner || f.unit.HealthData == nil || f.unit.Flags().HasAny(object.FlagDead|object.FlagDestroyed) ||
		f.item == nil || f.item.InvHolder != f.unit || !f.item.Flags().Has(object.FlagEquipped) || f.unit.UpdateDataPlayer().EquippedWeapon != f.item ||
		f.modifier == nil || f.modifier.Name() != "Regeneration1" || f.modifier.Update100.Fnc == nil || f.unit.Field110&0x20 == 0 {
		e2eError(fmt.Errorf("regeneration requires the live stock Regeneration1 sword equipped through actual UI input"))
		return
	}
	attrs := f.item.InitDataModifier()
	if attrs == nil || attrs.Modifiers[2] != f.modifier || attrs.Modifiers[3] != nil || f.item.Class().Has(object.ClassArmor) {
		e2eError(fmt.Errorf("regeneration fixture modifier identity or item class changed"))
		return
	}
	f.maximum, f.saved = f.unit.HealthData.Max, f.unit.HealthData.Cur
	f.fps = noxServer.TickRate()
	if f.maximum == 0 || f.fps == 0 || f.modifier.Update100.Val <= 0 {
		e2eError(fmt.Errorf("invalid regeneration fixture HP/FPS/rate: %d/%d/%d", f.maximum, f.fps, f.modifier.Update100.Val))
		return
	}
	f.interval = uint32(f.modifier.Update100.Val) * f.fps / uint32(f.maximum)
	f.passiveInterval = 300 * f.fps / uint32(f.maximum)
	if f.interval == 0 || f.passiveInterval == 0 || uint64(f.fps)+3*uint64(f.interval) > 12000 {
		e2eError(fmt.Errorf("stock regeneration cannot be observed in the bounded fixture: item/passive interval=%d/%d", f.interval, f.passiveInterval))
		return
	}
	f.window = f.fps + 3*f.interval
	injury := e2ePlayerPassiveHealingBound(f.window, f.fps, f.passiveInterval) + 8
	if injury >= uint32(f.maximum) {
		e2eError(fmt.Errorf("regeneration fixture HP %d cannot sustain injury %d for the stock rate", f.maximum, injury))
		return
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{f.unit.CObj(), f.item.CObj(), f.modifier.C(), unsafe.Pointer(attrs), unsafe.Pointer(f.unit.HealthData)} {
			if uintptr(ptr) <= math.MaxUint32 {
				e2eError(fmt.Errorf("regeneration fixture requires native high pointers, got %p", ptr))
				return
			}
		}
	}
	// Only the initial injury and its cooldown timestamp are fixture setup.
	// Healing, modifier invocation and equipment state are never injected.
	f.before = f.maximum - uint16(injury)
	legacy.Nox_xxx_unitSetHP_4E4560(f.unit, f.before)
	f.frame = noxServer.Frame()
	f.keepaliveFrame = f.frame
	f.unit.Frame134 = f.frame
	if f.unit.HealthData.Cur != f.before {
		e2eError(fmt.Errorf("regeneration fixture injury did not reach HP %d", f.before))
		return
	}
	e2eLog.Printf("PLAYER REGEN PREPARED: owner=%p item=%p modifier=%p holder=%p callback=%p health=%d/%d frame=%d rate=%d interval=%d passive-interval=%d window=%d",
		f.unit, f.item, f.modifier, f.item.InvHolder, f.modifier.Update100.Fnc, f.before, f.maximum, f.frame, f.modifier.Update100.Val,
		f.interval, f.passiveInterval, f.window)
}

func (f *e2ePlayerRegenerationFixture) observeHealing() bool {
	now := noxServer.Frame()
	elapsed := now - f.frame
	// This long stock interval exceeds the client's inactivity threshold.
	// Keep it active with ordinary one-frame movement input, never by changing
	// spectator flags, clocks, damage state, healing or modifier callbacks.
	if f.keepalivePressed {
		e2eQueueInput(&seat.MouseButtonEvent{Button: seat.MouseButtonRight, Pressed: false})
		f.keepalivePressed = false
	} else if now-f.keepaliveFrame >= 300 {
		offset := 16
		if elapsed/300&1 != 0 {
			offset = -offset
		}
		mouse := noxClient.Viewport().ToScreenPos(image.Pt(int(f.unit.PosVec.X)+offset, int(f.unit.PosVec.Y)))
		e2eQueueInput(&seat.MouseMoveEvent{Pos: mouse}, &seat.MouseButtonEvent{Button: seat.MouseButtonRight, Pressed: true})
		f.keepaliveFrame, f.keepalivePressed = now, true
		e2eLog.Printf("PLAYER REGEN ACTIVE INPUT: elapsed=%d health=%d/%d status=%#x pos=%v mouse=%v", elapsed,
			f.unit.HealthData.Cur, f.unit.HealthData.Max, f.unit.ControllingPlayer().Field3680, f.unit.PosVec, mouse)
	}
	if elapsed < f.window {
		return false
	}
	if f.unit != noxServer.Players.HostUnit() || f.unit.HealthData == nil || f.unit.HealthData.Max != f.maximum ||
		f.unit.Frame134 != f.frame || f.unit.Flags().HasAny(object.FlagDead|object.FlagDestroyed) || f.unit.ControllingPlayer().Field3680&1 != 0 ||
		f.item.InvHolder != f.unit || !f.item.Flags().Has(object.FlagEquipped) || f.unit.UpdateDataPlayer().EquippedWeapon != f.item {
		e2eError(fmt.Errorf("regeneration owner, injury timestamp or equipped item changed during observation"))
		return true
	}
	health := f.unit.HealthData.Cur
	bound := e2ePlayerPassiveHealingBound(elapsed, f.fps, f.passiveInterval)
	if health <= f.before || uint32(health-f.before) <= bound {
		e2eError(fmt.Errorf("equipped Regeneration1 did not contribute healing beyond passive bound %d: HP %d->%d after %d frames", bound, f.before, health, elapsed))
		return true
	}
	e2eLog.Printf("PLAYER REGEN HEALED: owner=%p item=%p health=%d->%d/%d gain=%d passive-bound=%d elapsed=%d interval=%d",
		f.unit, f.item, f.before, health, f.maximum, health-f.before, bound, elapsed, f.interval)
	return true
}

// CheckPlayerRegeneration observes the real updateUnitsAAA -> equipped-item
// dispatcher -> stock C callback -> HP adjustment path. A timeout or passive
// regeneration alone is a failure; the action never directly calls an effect.
func (sc *e2eScenario) CheckPlayerRegeneration(name string) {
	f := &e2ePlayerRegenerationFixture{}
	sc.addWhen(0, name+" initial injury", 1200, func() bool {
		return noxServer.Players.HostUnit() != nil && noxClient.ClientPlayerUnit() != nil && nox_client_isConnected()
	}, f.prepare)
	sc.Wait(2, name+" publish initial injury")
	sc.Screen(name + " injured")
	sc.addWhen(0, name+" natural equipped-item healing", 15000, f.observeHealing, func() {})
	sc.Wait(2, name+" publish healed HP")
	sc.Screen(name + " healed")
	sc.add(0, name+" restore fixture HP", func() {
		e2eQueueInput(&seat.MouseButtonEvent{Button: seat.MouseButtonRight, Pressed: false})
		legacy.Nox_xxx_unitSetHP_4E4560(f.unit, f.saved)
	})
}
