package opennox

import (
	"fmt"
	"time"

	"github.com/opennox/libs/client/keybind"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/platform"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"
	ns "github.com/opennox/noxscript/ns/v4"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

// The E2E clock stays deterministic. Measure real elapsed time separately:
// changing the virtual clock or sleeping in the observer could hide this bug.
func e2eAwardFramePacingError(elapsed time.Duration, loops, draws int, beforeFrame, frame uint32, beforePos, pos types.Pointf) error {
	if loops != 15 || draws < 14 || draws > 16 {
		return fmt.Errorf("award render loop did not continue: loops=%d draws=%d", loops, draws)
	}
	if frame != beforeFrame || pos != beforePos {
		return fmt.Errorf("award advanced paused gameplay: frame=%d->%d pos=%v->%v", beforeFrame, frame, beforePos, pos)
	}
	if elapsed < 450*time.Millisecond || float64(draws)/elapsed.Seconds() > 34 {
		return fmt.Errorf("award rendering was not paced: elapsed=%s loops=%d draws=%d", elapsed, loops, draws)
	}
	return nil
}

// Exercise genuine XP, spell-book and NoxScript halberd awards. Only award
// inputs and the user's frame-limit setting are supplied; pause flags, clocks,
// server frames, animation, knowledge and equipment outcomes are never written.
func (sc *e2eScenario) CheckAwardFramePacing(name string) {
	sc.playerPlasmaInventory(false, name+" close inventory through input")
	var unit *server.Object
	var oldLimit bool
	var oldSlow time.Duration
	sc.addWhen(0, name+" connected Solo player", 1200, func() bool {
		return noxServer.Players.HostUnit() != nil && noxClient.ClientPlayerUnit() != nil && legacy.Get_dword_8531A0_2576() != nil && nox_client_isConnected()
	}, func() {
		unit = noxServer.Players.HostUnit()
		if !noxflags.HasGame(noxflags.GameHost) || !noxflags.HasGame(noxflags.GameClient) || !noxflags.HasGame(noxflags.GameFlag29) || !noxflags.HasGame(noxflags.GameModeCoop) || noxflags.HasGame(noxflags.GamePause) || noxflags.HasEngine(noxflags.EngineNoRendering) {
			e2eError(fmt.Errorf("award pacing requires the real rendered Solo synchronization path"))
			return
		}
		oldLimit, oldSlow = useFrameLimit, e2e.slow
		// Remove the driver's artificial delay. The live limiter alone must
		// supply the measured wait; natural FX expiry then runs normally.
		e2e.slow = 0
		setEnableFrameLimit(false)
	})

	awards := []struct {
		kind string
		item string
	}{
		{kind: "level-up"},
		{kind: "new-spell"},
		{kind: "halberd", item: "OblivionHalberd"},
		{kind: "halberd", item: "OblivionHeart"},
		{kind: "halberd", item: "OblivionWierdling"},
		{kind: "halberd", item: "OblivionOrb"},
	}
	for i, award := range awards {
		label := name + " " + award.kind
		if award.item != "" {
			label += " " + award.item
		}
		var beforeLevel uint8
		var beforeBook [25][2]uint32
		var started time.Time
		var ticks time.Duration
		var frame uint32
		var draws int
		var pos types.Pointf
		sc.addWhen(0, label+" actual award input", 1200, func() bool {
			return !noxflags.HasGame(noxflags.GamePause) && legacy.Nox_xxx_get_57AF20() == 0 && legacy.Sub_45D9B0() == 0
		}, func() {
			switch award.kind {
			case "level-up":
				beforeLevel = unit.ControllingPlayer().Level
				threshold := float32(noxServer.Balance.FloatInd("XPTable", int(int8(beforeLevel))+1))
				amount := threshold - unit.Experience
				if amount <= 0 {
					e2eError(fmt.Errorf("no positive XP award available: level=%d XP=%g threshold=%g", beforeLevel, unit.Experience, threshold))
					return
				}
				legacy.Nox_xxx_plyrGiveExp_4EF3A0_exp_level(unit, amount)
				e2eLog.Printf("AWARD PACING XP INPUT: level=%d threshold=%g award=%g", beforeLevel, threshold, amount)
			case "new-spell":
				var ok bool
				beforeBook, _, ok = legacy.ClientQuickbarSnapshot()
				if !ok || unit.ControllingPlayer().SpellLvl[spell.SPELL_HASTE] != 0 || legacy.Get_dword_8531A0_2576().SpellLvl[spell.SPELL_HASTE] != 0 {
					e2eError(fmt.Errorf("stock Haste is not an unlearned spell"))
					return
				}
				if got := legacy.Nox_xxx_spellGrantToPlayer_4FB550(unit, spell.SPELL_HASTE, 1, 1, 1); got != 1 {
					e2eError(fmt.Errorf("real Haste book award failed: result=%d", got))
					return
				}
			case "halberd":
				noxServer.noxScriptP().SetHalberd(ns.HalberdLevel(i - 2))
			}
			e2eLog.Printf("AWARD PACING INPUT: kind=%s item=%s real_award_path=true", award.kind, award.item)
		})
		sc.addWhen(1, label+" naturally paused rendering", 1200, func() bool {
			return noxflags.HasGame(noxflags.GamePause) && legacy.Nox_xxx_get_57AF20() != 0 && (award.kind != "new-spell" || legacy.Sub_45D9B0() != 0)
		}, func() {
			started, ticks = time.Now(), platform.Ticks()
			frame, draws, pos = noxServer.Frame(), noxClient.Debug.DrawCnt, unit.PosVec
			setEnableFrameLimit(true)
		})
		sc.add(15, label+" measure live render pacing", func() {
			elapsed := time.Since(started)
			loops, rendered := int(platform.Ticks()-ticks), noxClient.Debug.DrawCnt-draws
			if !noxflags.HasGame(noxflags.GamePause) || legacy.Nox_xxx_get_57AF20() == 0 || !useFrameLimit {
				e2eError(fmt.Errorf("award pause ended before pacing was observed"))
				return
			}
			if err := e2eAwardFramePacingError(elapsed, loops, rendered, frame, noxServer.Frame(), pos, unit.PosVec); err != nil {
				e2eError(err)
				return
			}
			e2eLog.Printf("AWARD FRAME PACING VERIFIED: kind=%s item=%s elapsed=%s loops=%d draws=%d render_fps=%.2f frozen_server_frame=%d", award.kind, award.item, elapsed, loops, rendered, float64(rendered)/elapsed.Seconds(), frame)
			// E2E normally runs uncapped. Restore that driver setting between
			// samples, without changing the actual 5000-tick expiry condition.
			setEnableFrameLimit(false)
		})
		sc.Screen(label + " actual paused framebuffer")
		sc.addWhen(1, label+" natural award completion", 6200, func() bool {
			return legacy.Nox_xxx_get_57AF20() == 0 && legacy.Sub_45D9B0() == 0
		}, nil)
		if award.kind == "new-spell" {
			sc.Key(keybind.KeyB, label+" close book through B input")
		}
		sc.addWhen(30, label+" real gameplay resume and award outcome", 1200, func() bool {
			return !noxflags.HasGame(noxflags.GamePause) && noxServer.Frame() >= frame+12
		}, func() {
			switch award.kind {
			case "level-up":
				if unit.ControllingPlayer().Level != beforeLevel+1 || legacy.Get_dword_8531A0_2576().Level != beforeLevel+1 {
					e2eError(fmt.Errorf("level award did not reach both server and client"))
					return
				}
			case "new-spell":
				entries, _, ok := legacy.ClientQuickbarSnapshot()
				found := false
				for n, entry := range entries {
					if entry[0] == uint32(spell.SPELL_HASTE) && beforeBook[n][0] == 0 && entry[1]&1 != 0 {
						found = true
					}
				}
				if !ok || !found || unit.ControllingPlayer().SpellLvl[spell.SPELL_HASTE] != 1 || legacy.Get_dword_8531A0_2576().SpellLvl[spell.SPELL_HASTE] != 1 {
					e2eError(fmt.Errorf("natural new spell knowledge/quickbar reward missing"))
					return
				}
			case "halberd":
				weapon := unit.UpdateDataPlayer().EquippedWeapon
				if weapon == nil || weapon.ObjectTypeC().ID() != award.item || !weapon.Flags().Has(object.FlagEquipped) {
					e2eError(fmt.Errorf("natural halberd upgrade/equip outcome missing: want=%s weapon=%p", award.item, weapon))
					return
				}
			}
			e2eLog.Printf("AWARD NATURAL COMPLETION VERIFIED: kind=%s item=%s server_frame=%d->%d actual_award_preserved=true", award.kind, award.item, frame, noxServer.Frame())
		})
	}
	sc.add(0, name+" restore driver settings", func() {
		setEnableFrameLimit(oldLimit)
		e2e.slow = oldSlow
	})
}
