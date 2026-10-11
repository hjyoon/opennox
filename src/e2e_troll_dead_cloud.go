package opennox

import (
	"context"
	"fmt"
	"math"
	"unsafe"

	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/console"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/noxscript/ns/v4/damage"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func e2eTrollCloudLifetime54A270(lifetime float32, fps uint32) (int32, bool) {
	product := float32(float64(lifetime) * float64(int32(fps)))
	if math.IsNaN(float64(product)) || math.IsInf(float64(product), 0) {
		return 0, false
	}
	frames := math.RoundToEven(float64(product))
	if frames < 1 || frames > 1800 {
		return 0, false
	}
	return int32(frames), true
}

type e2eTrollCloudVictim struct {
	name               string
	unit               *server.Object
	hp                 uint16
	applied, firstTick uint32
	dot, client        bool
	audio              int
}

type e2eTrollCloudFixture struct {
	host, source, immune, cloud *server.Object
	victims                     []*e2eTrollCloudVictim
	original, center, direction types.Pointf
	start, prepared, killed     uint32
	created, wire               uint32
	scriptID                    int32
	expected, previous          int32
	phase, flatus               int
	immuneHP                    uint16
	visible, expired, active    bool
}

// The extra command exists only under the existing E2E environment. It lets a
// new YAML use ordinary F1/text/Enter without editing the frozen dispatcher.
func init() {
	if e2ePlay == "" {
		return
	}
	var fixture *e2eTrollCloudFixture
	noxConsole.Register(&console.Command{
		Token: "e2e-troll-cloud", HelpID: "nohelp", Flags: console.Server | console.NoHelp,
		Func: func(_ context.Context, _ *console.Console, args []string) bool {
			if len(args) != 1 {
				return false
			}
			switch args[0] {
			case "start":
				if fixture != nil || noxServer.Players.HostUnit() == nil {
					e2eError(fmt.Errorf("Troll cloud fixture requires one hosted start"))
					return true
				}
				fixture = &e2eTrollCloudFixture{active: true, start: noxServer.Frame()}
				noxServer.Audio.OnSound(fixture.observeSound)
				noxServer.TickHook(fixture.tick)
				e2eLog.Printf("TROLL CLOUD START: frame=%d kill=normal-script-damage callbacks=natural", fixture.start)
			case "finish":
				if fixture == nil {
					e2eError(fmt.Errorf("Troll cloud fixture was never started"))
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

func (f *e2eTrollCloudFixture) prepare() {
	f.host = noxServer.Players.HostUnit()
	f.original = f.host.PosVec
	origin, direction, err := e2eWarriorAbilityArena(f.original, f.host.Shape.Circle.R+40,
		func(from, to types.Pointf) bool { return e2eWarriorLaneClear(f.host, from, to) })
	if err != nil {
		e2eError(err)
		return
	}
	f.direction, f.center = direction, origin.Add(direction.Mul(80))
	f.source, f.immune = noxServer.NewObjectByTypeID("Troll"), noxServer.NewObjectByTypeID("Troll")
	f.victims = []*e2eTrollCloudVictim{
		{name: "player", unit: f.host},
		{name: "monster", unit: noxServer.NewObjectByTypeID("Wolf")},
		{name: "NPC", unit: noxServer.NewObjectByTypeID("NPC")},
	}
	units := []*server.Object{f.source, f.immune, f.victims[1].unit, f.victims[2].unit}
	for i, unit := range units {
		if unit == nil || unit.UpdateData == nil || unit.HealthData == nil || unit.HealthData.Cur == 0 {
			e2eError(fmt.Errorf("Troll cloud stock unit was not initialized: %d", i))
			return
		}
		owner := f.host
		if unit == f.source {
			owner = nil
		}
		noxServer.CreateObjectAt(unit, owner, f.center.Add(direction.Mul(float32(i)*48)))
		// Initial placement and idle AI isolate cloud contact; no health,
		// immunity, RNG, poison, animation or damage result is injected.
		unit.UpdateDataMonster().SetAggression(0)
		unit.ClearActionStack()
		unit.MonsterPushAction(ai.ACTION_WAIT, noxServer.Frame()+250000)
	}
	noxServer.ObjectsAddPending()
	f.immuneHP = f.immune.HealthData.Cur
	if !f.source.SubClass().Has(0x200) || !f.immune.SubClass().Has(0x200) ||
		f.source.UpdateDataMonster().MonsterDef == nil || f.source.UpdateDataMonster().MonsterDef.DeadFunc232 == nil {
		e2eError(fmt.Errorf("stock Troll poison immunity/dead callback is missing"))
		return
	}
	for _, victim := range f.victims {
		if victim.unit.HealthData == nil || victim.unit.Damage == nil ||
			victim.unit.Poison540 != 0 || victim.unit.SubClass().Has(0x200) ||
			!noxServer.S().IsEnemyTo(f.source, victim.unit) {
			e2eError(fmt.Errorf("Troll cloud has no susceptible enemy: %s", victim.name))
			return
		}
		victim.hp = victim.unit.HealthData.Cur
	}
	var ok bool
	f.expected, ok = e2eTrollCloudLifetime54A270(float32(noxServer.Balance.Float("SmallToxicCloudLifetime")), noxServer.TickRate())
	if !ok {
		e2eError(fmt.Errorf("stock SmallToxicCloudLifetime is outside the fixed 1..1800-frame fixture"))
		return
	}
	asObjectS(f.host).SetPos(origin.Sub(direction.Mul(48)))
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, unit := range append(units, f.host) {
			if uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 || uintptr(unit.UpdateData) <= math.MaxUint32 ||
				uintptr(unsafe.Pointer(unit.HealthData)) <= math.MaxUint32 {
				e2eError(fmt.Errorf("Troll cloud fixture is not C-owned native-width: %p", unit))
				return
			}
		}
	}
	f.prepared, f.phase = noxServer.Frame(), 1
	e2eLog.Printf("TROLL CLOUD PREPARED: source=%p update=%p immune=%p stock-immunity=%#x duration=%d player/NPC/monster=stock-HP placement-only", f.source, f.source.UpdateData, f.immune, uint32(f.immune.SubClass()), f.expected)
}

func (f *e2eTrollCloudFixture) observeSound(id sound.ID, kind int, owner *server.Object, _ types.Pointf) {
	if !f.active || f.phase < 2 {
		return
	}
	if id == sound.SoundTrollFlatus && owner == f.source {
		f.flatus++
		f.created = noxServer.Frame()
		if kind != 0 || f.flatus != 1 {
			e2eError(fmt.Errorf("Troll flatus sound count/kind differs: %d/%d", f.flatus, kind))
		}
	}
	if id == sound.SoundPoisonEffect {
		for _, victim := range f.victims {
			if owner == victim.unit {
				victim.audio++
				if kind != 0 || victim.audio != 1 {
					e2eError(fmt.Errorf("Troll poison effect sound differs: %s count=%d kind=%d", victim.name, victim.audio, kind))
				}
			}
		}
	}
}

func (f *e2eTrollCloudFixture) tick() {
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
	if !f.visible && !f.expired && noxClient.Objs.ByNetCode(uint16(f.wire)) != nil {
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

func (f *e2eTrollCloudFixture) discover() {
	if f.flatus != 1 || !e2eObjectInWorld(f.source) {
		return
	}
	for obj, remaining := f.source.Field129, 4096; obj != nil; obj, remaining = obj.Field128, remaining-1 {
		if remaining == 0 {
			e2eError(fmt.Errorf("Troll cloud owner list did not terminate"))
			return
		}
		if obj.ObjectTypeC().ID() != "SmallToxicCloud" {
			continue
		}
		data := (*server.ToxicCloudUpdateData)(obj.UpdateData)
		if f.cloud != nil || !e2eObjectInWorld(obj) || obj.ObjOwner != f.source || data == nil ||
			data.Duration > f.expected || data.Duration < f.expected-1 || obj.PosVec.Sub(f.source.PosVec).Len() > .01 ||
			obj.FindOwnerChainPlayer() != f.source || !f.source.Flags().Has(object.FlagAllowOverlap) {
			e2eError(fmt.Errorf("natural Troll cloud creation/owner/duration differs: %p", obj))
			return
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(obj)) <= math.MaxUint32 || uintptr(obj.UpdateData) <= math.MaxUint32) {
			e2eError(fmt.Errorf("Troll cloud/update pointer is below 4 GiB"))
			return
		}
		f.cloud, f.wire, f.scriptID, f.previous = obj, obj.NetCode, obj.ScriptIDVal, data.Duration
		side := types.Ptf(-f.direction.Y, f.direction.X)
		positions := []types.Pointf{obj.PosVec.Add(f.direction.Mul(25)), obj.PosVec.Sub(f.direction.Mul(25)), obj.PosVec.Add(side.Mul(25))}
		for i, victim := range f.victims {
			asObjectS(victim.unit).SetPos(positions[i])
			victim.unit.VelVec, victim.unit.ForceVec, victim.unit.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
		}
		asObjectS(f.immune).SetPos(obj.PosVec.Sub(side.Mul(25)))
		e2eLog.Printf("TROLL CLOUD CREATED: cloud=%p update=%p source=%p owner-chain=%p duration=%d configured=%d radius=35 sound=1 corpse=natural-overlap", obj, obj.UpdateData, f.source, obj.FindOwnerChainPlayer(), data.Duration, f.expected)
	}
}

func (f *e2eTrollCloudFixture) observeVictim(victim *e2eTrollCloudVictim) {
	unit := victim.unit
	if !e2eObjectInWorld(unit) || unit.HealthData.Cur == 0 {
		e2eError(fmt.Errorf("Troll cloud target disappeared: %s", victim.name))
		return
	}
	if victim.applied == 0 && unit.Poison540 != 0 {
		victim.applied = unit.HealthData.Field16
		var ok bool
		victim.firstTick, ok = e2ePoisonFirstTick(victim.applied, 1)
		if !ok || unit.Poison540 != 1 || unit.HealthData.Cur != victim.hp {
			e2eError(fmt.Errorf("SmallToxicCloud must apply dose 1 without instant HP damage: %s HP=%d/%d poison=%d", victim.name, unit.HealthData.Cur, victim.hp, unit.Poison540))
			return
		}
		e2eLog.Printf("TROLL CLOUD APPLIED: target=%s native=%p HP=%d unchanged dose=1 applied=%d first-DOT=%d", victim.name, unit, victim.hp, victim.applied, victim.firstTick)
	}
	if !victim.dot && unit.HealthData.Cur < victim.hp {
		if victim.applied == 0 || unit.HealthData.Cur != victim.hp-1 || unit.Frame134 != victim.firstTick ||
			unit.Field131 != uint32(object.DamagePoison) || unit.Obj130 != nil {
			e2eError(fmt.Errorf("SmallToxicCloud target lost HP outside the ordinary poison tick: %s HP=%d->%d tick=%d/%d attribution=%p", victim.name, victim.hp, unit.HealthData.Cur, unit.Frame134, victim.firstTick, unit.Obj130))
			return
		}
		victim.dot = true
		e2eLog.Printf("TROLL CLOUD DOT: target=%s HP=%d->%d tick=%d applied=%d source=nil weapon=nil type=POISON", victim.name, victim.hp, unit.HealthData.Cur, unit.Frame134, victim.applied)
	}
	if victim.dot && !victim.client {
		dr := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(unit)))
		if dr == nil {
			return
		}
		if unit == f.host {
			meter, ready := e2eClientHUDMeter(0)
			if !ready || meter.Current != uint32(unit.HealthData.Cur) || !meter.Poisoned {
				return
			}
		} else if delta, ok := legacy.HealthChangeForDrawable(dr.NetCode32); !ok || delta >= 0 {
			return
		}
		victim.client = true
		e2eLog.Printf("TROLL CLOUD CLIENT DAMAGE: target=%s drawable=%p HP=%d->%d poison=%d", victim.name, dr, victim.hp, unit.HealthData.Cur, unit.Poison540)
	}
}

func (f *e2eTrollCloudFixture) finish() {
	if !f.active || f.cloud == nil || !f.visible || !f.expired || f.flatus != 1 ||
		noxClient.Objs.ByNetCode(uint16(f.wire)) != nil {
		e2eError(fmt.Errorf("Troll cloud did not complete natural creation, rendering and expiry"))
		return
	}
	for _, victim := range f.victims {
		if !victim.dot || !victim.client || victim.audio != 1 {
			e2eError(fmt.Errorf("Troll cloud target did not complete poison/DOT/client replay: %s", victim.name))
			return
		}
		noxServer.Server.RemovePoison4EE9D0(victim.unit)
		if victim.unit.Poison540 != 0 || victim.unit.HealthData.Field16 != 0 {
			e2eError(fmt.Errorf("Troll cloud cleanup did not use the ordinary cure service"))
			return
		}
	}
	f.active = false
	path, err := e2eWriteMagicFrame("", noxClient.r.CopyPixBuffer())
	if err != nil {
		e2eError(err)
		return
	}
	e2eLog.Printf("TROLL CLOUD PASS: player/monster/NPC=zero-instant,dose1,natural-DOT,client-replay immune-Troll-HP=%d-unchanged sound=1 natural-death/lifetime/world/client-expiry=true path=%s", f.immuneHP, path)
	for _, unit := range []*server.Object{f.source, f.immune, f.victims[1].unit, f.victims[2].unit} {
		noxServer.DelayedDelete(unit)
	}
	asObjectS(f.host).SetPos(f.original)
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
}
