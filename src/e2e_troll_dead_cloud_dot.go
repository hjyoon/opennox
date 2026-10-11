package opennox

import (
	"context"
	"fmt"

	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/console"
	"github.com/opennox/libs/object"
	"github.com/opennox/noxscript/ns/v4/damage"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

// The original observer and its full-health HUD diagnostic remain frozen.
// This independent version uses every same stock setup/poison/DOT/expiry
// assertion, and samples the partial-meter pixels only after actual server DOT
// and client HP replay. It never supplies health, poison or rendering results.
type e2eTrollCloudDOTFixture struct {
	e2eTrollCloudFixture
}

func init() {
	if e2ePlay == "" {
		return
	}
	var fixture *e2eTrollCloudDOTFixture
	noxConsole.Register(&console.Command{
		Token: "e2e-troll-cloud-dot", HelpID: "nohelp", Flags: console.Server | console.NoHelp,
		Func: func(_ context.Context, _ *console.Console, args []string) bool {
			if len(args) != 1 {
				return false
			}
			switch args[0] {
			case "start":
				if fixture != nil || noxServer.Players.HostUnit() == nil {
					e2eError(fmt.Errorf("Troll cloud DOT fixture requires one hosted start"))
					return true
				}
				fixture = &e2eTrollCloudDOTFixture{e2eTrollCloudFixture: e2eTrollCloudFixture{active: true, start: noxServer.Frame()}}
				noxServer.Audio.OnSound(fixture.observeSound)
				noxServer.TickHook(fixture.tick)
				e2eLog.Printf("TROLL CLOUD START: frame=%d kill=normal-script-damage callbacks=natural pixel-sampling=after-server-DOT-and-client-replay", fixture.start)
			case "finish":
				if fixture == nil {
					e2eError(fmt.Errorf("Troll cloud DOT fixture was never started"))
				} else {
					fixture.finish()
				}
			default:
				return false
			}
			return true
		},
	})
}

func (f *e2eTrollCloudDOTFixture) tick() {
	if !f.active {
		return
	}
	frame := noxServer.Frame()
	// Normal physical motion events keep the online idle observer away.
	// Never write input clocks, observer flags or player status.
	if frame%60 == 0 {
		e2eQueueInput(&seat.MouseMoveEvent{Pos: noxClient.Inp.GetMousePos()})
	}
	if f.phase == 0 {
		host := noxServer.Players.HostUnit()
		if host != nil && host.HealthData != nil && host.HealthData.Cur > 1 && host.Buffs == 0 &&
			noxClient.ClientPlayerUnit() != nil && nox_client_isConnected() {
			f.prepare()
		} else if frame-f.start > 1200 {
			e2eError(fmt.Errorf("Troll cloud host readiness exceeded the fixed deadline"))
		}
		return
	}
	if f.phase == 1 {
		if frame-f.prepared < 12 {
			return
		}
		f.killed, f.phase = frame, 2
		ns := noxServer.noxScriptP()
		// Use the same normal script Damage path as campaign triggers.
		// Health, dying/dead actions and DEAD_FUNCTION are not supplied.
		hp := f.source.HealthData.Cur
		ns.toObj(f.source).Damage(nil, int(hp)+1, damage.BLADE)
		head := f.source.UpdateDataMonster().AIStackHead()
		if f.source.HealthData.Cur != 0 || !f.source.Flags().Has(object.FlagDead) ||
			head == nil || head.Type() != ai.ACTION_DYING {
			e2eError(fmt.Errorf("normal script damage did not start Troll death: HP=%d->%d head=%v", hp, f.source.HealthData.Cur, head))
			return
		}
		e2eLog.Printf("TROLL CLOUD KILLED: source=%p HP=%d->0 action=DYING frame=%d", f.source, hp, f.killed)
		return
	}
	if f.cloud == nil {
		f.discover()
		if f.cloud == nil {
			if frame-f.killed > 300 {
				e2eError(fmt.Errorf("natural Troll death did not create its SmallToxicCloud"))
			}
			return
		}
	}
	if e2eFistInWorld(f.cloud, f.wire, f.scriptID) {
		remaining := (*server.ToxicCloudUpdateData)(f.cloud.UpdateData).Duration
		if remaining < 0 || remaining > f.previous || f.previous-remaining > 1 {
			e2eError(fmt.Errorf("SmallToxicCloud lifetime is not a natural countdown: %d->%d", f.previous, remaining))
			return
		}
		f.previous = remaining
	} else if !f.expired {
		if f.previous > 1 {
			e2eError(fmt.Errorf("SmallToxicCloud disappeared early at duration=%d", f.previous))
			return
		}
		f.expired = true
		e2eLog.Printf("TROLL CLOUD EXPIRED: cloud=%p last-duration=%d age=%d configured=%d", f.cloud, f.previous, frame-f.created, f.expected)
	}
	if !e2eObjectInWorld(f.immune) || f.immune.Poison540 != 0 || f.immune.HealthData.Cur != f.immuneHP {
		e2eError(fmt.Errorf("SmallToxicCloud damaged the naturally poison-immune Troll"))
		return
	}
	for _, victim := range f.victims {
		f.observeVictim(victim)
	}
	if !f.visible && !f.expired && f.victims[0].dot && f.victims[0].client && noxClient.Objs.ByNetCode(uint16(f.wire)) != nil {
		meter, ready := e2eClientHUDMeter(0)
		if ready && meter.Poisoned && meter.PoisonTubeReady {
			pixels, err := e2eAssertHUDMeterPixels(noxClient.r.CopyPixBuffer(), meter, 'g', "Troll small cloud poison")
			if err != nil {
				e2eError(err)
				return
			}
			path, err := e2eWriteMagicFrame("", noxClient.r.CopyPixBuffer())
			if err != nil {
				e2eError(err)
				return
			}
			f.visible = true
			e2eLog.Printf("TROLL CLOUD VISIBLE: cloud=%p drawable=%p green-pixels=%d frame=%d path=%s", f.cloud, noxClient.Objs.ByNetCode(uint16(f.wire)), pixels.Filled, frame, path)
		}
	}
}
