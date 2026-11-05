package opennox

import (
	"context"
	"fmt"
	"math"
	"unsafe"

	"github.com/opennox/libs/console"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

// The source event is a diagnostic stimulus through the real audio binding.
// SightRange=0 is an ordinary diagnostic map property. Auditory admission and
// reaction must precede any visual enemy; no hearing fields, AI actions,
// sound flags or clocks are set.
type e2eAIHearingFixture struct {
	kind, mode                   string
	host, unit                   *server.Object
	original, sourcePosition     types.Pointf
	sound                        sound.ID
	flags, score, kindCode       int
	armed, emittedFrame          uint32
	hostHP, unitHP               uint16
	faceAtEmit                   server.Dir16
	requested, emitted, queued   bool
	heard, reacted, visual, done bool
	audioEvents                  int
}

func init() {
	if e2ePlay == "" {
		return
	}
	var current *e2eAIHearingFixture
	installed := false
	noxConsole.Register(&console.Command{Token: "e2e-ai-hearing", HelpID: "nohelp", Flags: console.Server | console.NoHelp, Func: func(_ context.Context, _ *console.Console, args []string) bool {
		if len(args) == 1 && args[0] == "finish" {
			if current == nil {
				e2eError(fmt.Errorf("hearing diagnostic was not started"))
			} else {
				current.finish()
			}
			return true
		}
		if len(args) != 3 || args[0] != "start" || (args[1] != "NPC" && args[1] != "Troll") {
			return false
		}
		switch args[2] {
		case "run", "walk", "spell", "private", "friendly-run":
		default:
			return false
		}
		if current != nil && !current.done {
			e2eError(fmt.Errorf("one hearing diagnostic at a time"))
			return true
		}
		current = &e2eAIHearingFixture{kind: args[1], mode: args[2], armed: noxServer.Frame()}
		if !installed {
			installed = true
			noxServer.Audio.OnSound(func(id sound.ID, kind int, source *server.Object, pos types.Pointf) {
				if current != nil && !current.done {
					current.observeAudio(id, kind, source, pos)
				}
			})
			noxServer.TickHook(func() {
				if current != nil && !current.done {
					current.tick()
				}
			})
		}
		return true
	}})
}

func (f *e2eAIHearingFixture) prepare() {
	f.host = noxServer.Players.HostUnit()
	if f.host == nil || f.host.HealthData == nil || f.host.HealthData.Cur == 0 || f.host.Buffs != 0 {
		e2eError(fmt.Errorf("hearing diagnostic needs a living unenchanted host"))
		return
	}
	f.original, f.hostHP = f.host.PosVec, f.host.HealthData.Cur
	f.unit = noxServer.NewObjectByTypeID(f.kind)
	if f.unit == nil || f.unit.UpdateData == nil || f.unit.HealthData == nil {
		e2eError(fmt.Errorf("missing stock hearing body %s", f.kind))
		return
	}
	from, to, err := e2eAIFirstAttackArena(f.host, max(f.host.Shape.Circle.R, f.unit.Shape.Circle.R)+12, false)
	if err != nil {
		e2eError(err)
		return
	}
	asObjectS(f.host).SetPos(to)
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	var owner server.Obj
	if f.mode == "friendly-run" {
		owner = f.host
	}
	noxServer.CreateObjectAt(f.unit, owner, from)
	noxServer.ObjectsAddPending()
	asObjectS(f.unit).SetAggression(0.83)
	d := f.unit.UpdateDataMonster()
	stockSight := d.SightRange
	d.SightRange = 0
	f.unit.SetDir(server.DirFromVec(from.Sub(to)))
	f.unitHP = f.unit.HealthData.Cur
	if d.CurrentEnemy != nil || d.PreferredEnemy != nil || d.Field97 != 0 || d.Field101 != 0 || legacy.Nox_xxx_checkIsKillable_528190(f.unit) == 0 {
		e2eError(fmt.Errorf("stock hearing body not initially clean/killable: %s/%s", f.kind, f.mode))
		return
	}
	for _, pointer := range []unsafe.Pointer{unsafe.Pointer(f.unit), f.unit.UpdateData, unsafe.Pointer(f.host)} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
			e2eError(fmt.Errorf("hearing body/data must have native pointers above4GiB"))
			return
		}
	}
	walk := f.mode == "walk"
	candidates := []sound.ID{sound.SoundRunOnDirt, sound.SoundRunOnStone, sound.SoundRunOnWood, sound.SoundRunOnSnow, sound.SoundRunOnMud, sound.SoundRunOnWater}
	if walk {
		candidates = []sound.ID{sound.SoundWalkOnDirt, sound.SoundWalkOnStone, sound.SoundWalkOnWood, sound.SoundWalkOnSnow, sound.SoundWalkOnMud, sound.SoundWalkOnWater}
	} else if f.mode == "spell" {
		candidates = []sound.ID{sound.SoundCleansingFlameCast}
	}
	for _, id := range candidates {
		flags := noxServer.Audio.Flags(id)
		if noxServer.Audio.Field12(id) <= 0 || flags&0x10 != 0 || (flags&0x20 != 0) != walk || (!walk && flags&0x40 == 0) {
			continue
		}
		score := noxServer.ai.soundFadePerc(id, f.host.PosVec, f.unit.PosVec)
		if score < 60 || score >= 89 {
			continue
		}
		f.sound, f.flags, f.score = id, flags, score
		break
	}
	if f.sound == 0 || noxServer.S().IsEnemyTo(f.unit, f.host) == (f.mode == "friendly-run") {
		e2eError(fmt.Errorf("stock run/walk descriptor or ordinary hostility unavailable: %s/%s", f.kind, f.mode))
		return
	}
	if f.mode == "private" {
		f.kindCode = 2
	}
	f.armed = noxServer.Frame()
	e2eLog.Printf("AI HEARING PREPARED: body=%s mode=%s unit=%p update=%p source=%p from=%v to=%v stock-sight=%g diagnostic-sight=0 aggression=%g sound=%d/%s flags=%#x max-distance=%d score=%d hostile=%t stack=%v heard-fields=untouched", f.kind, f.mode, f.unit, d, f.host, from, to, stockSight, d.Aggression, f.sound, f.sound, f.flags, noxServer.Audio.MaxDist(f.sound), f.score, noxServer.S().IsEnemyTo(f.unit, f.host), d.GetAIStack())
}

func (f *e2eAIHearingFixture) observeAudio(id sound.ID, kind int, source *server.Object, pos types.Pointf) {
	if !f.requested || id != f.sound || source != f.host || kind != f.kindCode {
		return
	}
	f.audioEvents++
	f.emitted, f.emittedFrame, f.sourcePosition = true, noxServer.Frame(), pos
	for event := noxServer.ai.listenHead; event != nil; event = event.next {
		if event.snd == id && event.obj == source && event.pos == pos && event.frame == f.emittedFrame {
			f.queued = true
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(event)) <= math.MaxUint32 {
				e2eError(fmt.Errorf("hearing queue record must preserve native pointers"))
				return
			}
		}
	}
	e2eLog.Printf("AI HEARING AUDIO: body=%s mode=%s id=%d flags=%#x kind=%d actual-event=true actual-listen-queue=%t frame=%d source=%p pos=%v", f.kind, f.mode, id, f.flags, kind, f.queued, f.emittedFrame, source, pos)
}

func (f *e2eAIHearingFixture) tick() {
	if f.unit == nil {
		host := noxServer.Players.HostUnit()
		if host != nil && host.HealthData != nil && host.HealthData.Cur != 0 && host.Buffs == 0 {
			f.prepare()
		} else if noxServer.Frame()-f.armed >= 30 && (noxServer.Frame()-f.armed)%30 == 0 {
			var hp uint16
			var buffs uint32
			if host != nil {
				buffs = host.Buffs
				if host.HealthData != nil {
					hp = host.HealthData.Cur
				}
			}
			e2eLog.Printf("AI HEARING WAIT: body=%s mode=%s host=%p HP=%d buffs=%#x elapsed=%d", f.kind, f.mode, host, hp, buffs, noxServer.Frame()-f.armed)
		}
		return
	}
	if f.host == nil || f.unit.Flags().HasAny(object.FlagDead|object.FlagDestroyed) {
		return
	}
	if !f.requested && noxServer.Frame()-f.armed >= 8 {
		f.requested = true
		f.faceAtEmit = f.unit.Direction1
		noxServer.Audio.EventObj(f.sound, f.host, f.kindCode, f.host.NetCode)
	}
	d := f.unit.UpdateDataMonster()
	if f.emitted && f.mode != "run" && f.mode != "spell" && noxServer.Frame()-f.emittedFrame > 2 {
		// The original event is no longer eligible after two frames. Observe
		// that complete window before a later visual encounter can interfere.
		defer f.finish()
	}
	if !f.reacted {
		f.visual = f.visual || d.CurrentEnemy != nil || d.PreferredEnemy != nil
	}
	if !f.emitted || d.Field101 < f.emittedFrame || d.Field101-f.emittedFrame > 2 || d.Field98 != f.host.NetCode || d.Field99X != f.sourcePosition.X || d.Field99Y != f.sourcePosition.Y {
		return
	}
	if !f.heard {
		f.heard = true
		e2eLog.Printf("AI HEARING DETECTED: body=%s mode=%s sound=%d native-heard=%d source-netcode=%#x score=%d position=%v frame=%d stack=%v", f.kind, f.mode, f.sound, d.Field97, d.Field98, d.Field102, f.sourcePosition, d.Field101, d.GetAIStack())
	}
	for _, action := range d.GetAIStack() {
		if (action.Type() == ai.ACTION_FACE_LOCATION || action.Type() == ai.ACTION_MOVE_TO) && action.ArgPos(0) == f.sourcePosition {
			if !f.reacted {
				f.reacted = true
				e2eLog.Printf("AI HEARING REACTION: body=%s mode=%s action=%s position=%v natural=true current-enemy=%p", f.kind, f.mode, action.Type(), action.ArgPos(0), d.CurrentEnemy)
			}
		}
	}
	if !f.reacted && f.unit.Direction1 != f.faceAtEmit && f.unit.Direction1 == server.DirFromVec(f.sourcePosition.Sub(f.unit.PosVec)) {
		f.reacted = true
		e2eLog.Printf("AI HEARING REACTION: body=%s mode=%s facing=%d->%d position=%v natural=true current-enemy=%p", f.kind, f.mode, f.faceAtEmit, f.unit.Direction1, f.sourcePosition, d.CurrentEnemy)
	}
	if (f.mode == "run" || f.mode == "spell") && f.reacted {
		// Later visual acquisition after turning is a separate natural stage.
		f.finish()
	}
}

func (f *e2eAIHearingFixture) finish() {
	if f.done {
		return
	}
	want := f.mode == "run" || f.mode == "spell"
	if f.unit == nil || f.host == nil || !f.emitted || f.audioEvents != 1 || f.queued != (f.mode != "private") || f.heard != want || f.reacted != want || f.visual || f.host.HealthData.Cur != f.hostHP || f.unit.HealthData.Cur != f.unitHP {
		e2eError(fmt.Errorf("hearing %s/%s failed: emitted=%t events=%d queued=%t heard=%t reacted=%t want=%t visual=%t", f.kind, f.mode, f.emitted, f.audioEvents, f.queued, f.heard, f.reacted, want, f.visual))
		return
	}
	frame := noxServer.Frame()
	path, err := e2eWriteMagicFrame("", noxClient.r.CopyPixBuffer())
	if err != nil {
		e2eError(err)
		return
	}
	e2eLog.Printf("AI HEARING PASS: body=%s mode=%s emitted=true queued=%t heard=%t reaction=%t sight-enemy=false HP=unchanged sound=%d flags=%#x original-score=%d frame=%d path=%s prepared-stock=true earned-campaign-credit=false", f.kind, f.mode, f.queued, f.heard, f.reacted, f.sound, f.flags, f.score, frame, path)
	f.done = true
	noxServer.DelayedDelete(f.unit)
	asObjectS(f.host).SetPos(f.original)
}
