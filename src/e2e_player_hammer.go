package opennox

import (
	"fmt"
	"image"
	"math"
	"unsafe"

	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/player"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

type e2ePlayerHammerOutcome struct {
	before, after           [2]uint16
	attributed              [2]bool
	started, animation      bool
	advanced, quake         bool
	completed, held, stable bool
	sounds                  int
}

func (o e2ePlayerHammerOutcome) validate() error {
	for i := range o.before {
		if o.before[i] == 0 || o.after[i] == 0 || o.after[i] >= o.before[i] || !o.attributed[i] {
			return fmt.Errorf("hammer target %d: health=%d->%d attributed=%t", i, o.before[i], o.after[i], o.attributed[i])
		}
	}
	if !o.started || !o.animation || !o.advanced || !o.quake || o.sounds != 1 || !o.completed || !o.held || !o.stable {
		return fmt.Errorf("hammer lifecycle: started=%t animation=%t advanced=%t quake=%t sounds=%d completed=%t held=%t stable=%t",
			o.started, o.animation, o.advanced, o.quake, o.sounds, o.completed, o.held, o.stable)
	}
	return nil
}

func e2ePlayerHammerTargetDistances(unitRadius float32, targetRadii [2]float32, attackRange float32) ([2]float32, error) {
	// Keep the two ordinary colliders separate along the same aiming lane.
	// A lateral pair lets the normal aim lock turn one out of the front sector.
	near := max(float32(30), unitRadius+targetRadii[1]+4)
	distances := [2]float32{near + targetRadii[0] + targetRadii[1] + 4, near}
	for i, distance := range distances {
		if attackRange <= 0 || float32(math.Abs(float64(distance-35)))-targetRadii[i] >= attackRange {
			return [2]float32{}, fmt.Errorf("hammer target %d does not fit stock range %g", i, attackRange)
		}
	}
	return distances, nil
}

type e2ePlayerHammerFixture struct {
	unit, weapon                *server.Object
	targets                     [2]*server.Object
	original, aim               types.Pointf
	frame, attackFrame, lastLog uint32
	firstClientFrame            uint32
	cycle                       int
	active                      bool
	hit                         [2]uint16
	outcome                     e2ePlayerHammerOutcome
}

func (f *e2ePlayerHammerFixture) held() bool {
	return f.weapon.InvHolder == f.unit && f.weapon.Flags().Has(object.FlagEquipped) &&
		f.unit.UpdateDataPlayer().EquippedWeapon == f.weapon &&
		f.unit.ControllingPlayer().WeaponEquip&uint32(object.WeaponHammer) != 0
}

func (f *e2ePlayerHammerFixture) prepare() {
	f.unit = noxServer.Players.HostUnit()
	if f.unit == nil || f.unit.UpdateData == nil || f.unit.ControllingPlayer() == nil ||
		f.unit.ControllingPlayer().PlayerClass() != player.Warrior {
		e2eError(fmt.Errorf("player hammer requires a live Warrior"))
		return
	}
	f.weapon = f.unit.UpdateDataPlayer().EquippedWeapon
	if f.weapon == nil || f.weapon.ObjectTypeC().ID() != "WarHammer" || f.weapon.InitData == nil || !f.held() {
		e2eError(fmt.Errorf("WarHammer not equipped through actual inventory input"))
		return
	}
	modifier := noxServer.Modif.Nox_xxx_getProjectileClassById413250(int(f.weapon.TypeInd))
	if modifier == nil || modifier.Range68 <= 0 {
		e2eError(fmt.Errorf("WarHammer lacks its stock range record"))
		return
	}
	for i, name := range []string{"Troll", "NPC"} {
		f.targets[i] = noxServer.NewObjectByTypeID(name)
		if target := f.targets[i]; target == nil || target.Shape.Kind != server.ShapeKindCircle {
			e2eError(fmt.Errorf("hammer stock %s target missing circle shape", name))
			return
		}
	}
	f.original = f.unit.PosVec
	distances, err := e2ePlayerHammerTargetDistances(f.unit.Shape.Circle.R,
		[2]float32{f.targets[0].Shape.Circle.R, f.targets[1].Shape.Circle.R}, modifier.Range68)
	if err != nil {
		e2eError(err)
		return
	}
	laneRadius := max(f.unit.Shape.Circle.R+4, max(f.targets[0].Shape.Circle.R, f.targets[1].Shape.Circle.R)+4)
	origin, direction, err := e2eWarriorAbilityArena(f.original, laneRadius, func(from, to types.Pointf) bool {
		return e2eWarriorLaneClear(f.unit, from, to)
	})
	if err != nil {
		e2eError(err)
		return
	}
	asObjectS(f.unit).SetPos(origin)
	f.unit.VelVec, f.unit.ForceVec, f.unit.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	f.aim = origin.Add(direction.Mul(60))
	for i, target := range f.targets {
		pos := origin.Add(direction.Mul(distances[i]))
		noxServer.CreateObjectAt(target, nil, pos)
		noxServer.ObjectsAddPending()
		if target.HealthData == nil || target.UpdateData == nil || target.Damage == nil || !target.Class().Has(object.ClassMonster) ||
			i == 1 && !target.SubClass().AsMonster().Has(object.MonsterNPC) {
			e2eError(fmt.Errorf("hammer stock target %d not initialized", i))
			return
		}
		// Placement, durability and ordinary waiting AI are fixture setup.
		// Do not inject attack state, hit callbacks, damage, buffs or NoUpdate.
		asObjectS(target).SetMaxHealth(2000)
		target.UpdateDataMonster().SetAggression(0)
		target.ClearActionStack()
		target.MonsterPushAction(ai.ACTION_WAIT, noxServer.Frame()+250000)
		if unsafe.Sizeof(uintptr(0)) == 8 {
			for _, ptr := range []unsafe.Pointer{unsafe.Pointer(f.unit), unsafe.Pointer(f.weapon), f.unit.UpdateData,
				unsafe.Pointer(target), target.UpdateData} {
				if uintptr(ptr) <= math.MaxUint32 {
					e2eError(fmt.Errorf("hammer fixture pointer below 4 GiB: %p", ptr))
					return
				}
			}
		}
		e2eLog.Printf("PLAYER HAMMER TARGET: type=%s pos=%v health=%d flags=%#x subclass=%#x", target.ObjectTypeC().ID(),
			target.PosVec, target.HealthData.Cur, uint32(target.ObjFlags), uint32(target.ObjSubClass))
	}
	noxServer.Audio.OnSound(func(id sound.ID, kind int, unit *server.Object, _ types.Pointf) {
		if f.active && id == sound.SoundHammerMissing && kind == 0 && unit == f.unit {
			f.outcome.sounds++
		}
	})
	e2eLog.Printf("PLAYER HAMMER PREPARED: unit=%p weapon=%p equip=%#x range=%g aim=%v", f.unit,
		f.weapon, f.unit.ControllingPlayer().WeaponEquip, modifier.Range68, f.aim)
}

func (f *e2ePlayerHammerFixture) input() {
	if !f.held() {
		e2eError(fmt.Errorf("hammer lost before actual attack input"))
		return
	}
	f.cycle++
	f.frame, f.attackFrame, f.lastLog = noxServer.Frame(), f.unit.Field34, 0
	f.outcome = e2ePlayerHammerOutcome{}
	for i, target := range f.targets {
		if !e2eObjectInWorld(target) || target.HealthData == nil || target.HealthData.Cur == 0 {
			e2eError(fmt.Errorf("hammer durable target %d disappeared before attack", i))
			return
		}
		f.outcome.before[i] = target.HealthData.Cur
	}
	f.active = true
	mouse := noxClient.Viewport().ToScreenPos(image.Pt(int(f.aim.X), int(f.aim.Y)))
	e2eQueueInput(&seat.MouseMoveEvent{Pos: mouse}, &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true})
	e2eLog.Printf("PLAYER HAMMER INPUT: cycle=%d frame=%d health=%v mouse=%v stamina=%d cost=%d", f.cycle, f.frame,
		f.outcome.before, mouse, f.unit.UpdateDataPlayer().Stamina,
		server.WeaponStaminaByType4F7E80(f.unit.ControllingPlayer().WeaponEquip))
}

func (f *e2ePlayerHammerFixture) observeHit() bool {
	now := noxServer.Frame()
	for i, target := range f.targets {
		if !e2eObjectInWorld(target) || target.HealthData == nil || target.HealthData.Cur == 0 {
			e2eError(fmt.Errorf("hammer durable target %d disappeared during attack", i))
			return true
		}
		f.outcome.after[i] = target.HealthData.Cur
		f.outcome.attributed[i] = target.Obj130 == f.weapon && target.Field131 == uint32(object.DamageCrush) && target.Frame134 >= f.frame
	}
	if f.unit.Field34 != f.attackFrame && noxServer.PlayerActionState4FA2B0(f.unit) == 39 && f.unit.UpdateDataPlayer().Field0 != 0 {
		f.outcome.started = true
	}
	if drawable := noxClient.ClientPlayerUnit(); drawable != nil && drawable.AnimInd == 39 {
		if !f.outcome.animation {
			f.firstClientFrame = drawable.AnimFrameSlave
		}
		f.outcome.animation = true
		f.outcome.advanced = f.outcome.advanced || drawable.AnimFrameSlave != f.firstClientFrame
	}
	f.outcome.quake = f.outcome.quake || noxClient.Viewport().Jiggle12 != 0
	if f.outcome.after[0] < f.outcome.before[0] && f.outcome.after[1] < f.outcome.before[1] &&
		f.outcome.started && f.outcome.animation && f.outcome.advanced && f.outcome.quake && f.outcome.sounds == 1 {
		f.hit = f.outcome.after
		e2eLog.Printf("PLAYER HAMMER HIT: cycle=%d health=%v->%v attributed=%v animation=39 advanced=%t quake=%t sounds=%d elapsed=%d",
			f.cycle, f.outcome.before, f.hit, f.outcome.attributed, f.outcome.advanced, f.outcome.quake, f.outcome.sounds, now-f.frame)
		return true
	}
	if f.lastLog == 0 || now-f.lastLog >= 10 {
		f.lastLog = now
		e2eLog.Printf("PLAYER HAMMER WAIT: cycle=%d elapsed=%d health=%v->%v started=%t animation=%t advanced=%t quake=%t sounds=%d state=%d anim-frame=%d deadline=%d",
			f.cycle, now-f.frame, f.outcome.before, f.outcome.after, f.outcome.started, f.outcome.animation, f.outcome.advanced,
			f.outcome.quake, f.outcome.sounds, f.unit.UpdateDataPlayer().State, f.unit.UpdateDataPlayer().Field59_0, f.unit.UpdateDataPlayer().Field0)
		cosine, sine := server.SinCosDir(byte(f.unit.Direction1))
		center := f.unit.PosVec.Add(types.Ptf(35*cosine, 35*sine))
		for _, target := range f.targets {
			e2eLog.Printf("PLAYER HAMMER GEOMETRY: player=%v direction=%d center=%v target=%s/%v/r=%g bounds=%v/%v facing=%d trace=%t source=%p damage-type=%d damage-frame=%d",
				f.unit.PosVec, f.unit.Direction1, center, target.ObjectTypeC().ID(), target.PosVec, target.Shape.Circle.R,
				target.CollideP1, target.CollideP2, legacy.Nox_server_testTwoPointsAndDirection_4E6E50(f.unit.PosVec, int16(f.unit.Direction1), target.PosVec),
				noxServer.MapTraceRay(f.unit.PosVec, target.PosVec, 5), target.Obj130, target.Field131, target.Frame134)
		}
	}
	return false
}

func (f *e2ePlayerHammerFixture) observeCompletion() bool {
	drawable := noxClient.ClientPlayerUnit()
	if noxServer.PlayerActionState4FA2B0(f.unit) == 39 || drawable == nil || drawable.AnimInd == 39 || noxClient.Viewport().Jiggle12 != 0 {
		return false
	}
	f.outcome.completed, f.outcome.held, f.outcome.stable = true, f.held(), true
	for i, target := range f.targets {
		if !e2eObjectInWorld(target) || target.HealthData == nil {
			e2eError(fmt.Errorf("hammer target %d disappeared before completion", i))
			return true
		}
		f.outcome.after[i] = target.HealthData.Cur
		f.outcome.stable = f.outcome.stable && f.outcome.after[i] == f.hit[i]
	}
	if err := f.outcome.validate(); err != nil {
		e2eError(err)
		return true
	}
	f.active = false
	e2eLog.Printf("PLAYER HAMMER COMPLETED: cycle=%d health=%v->%v held=%t sounds=%d elapsed=%d", f.cycle,
		f.outcome.before, f.outcome.after, f.outcome.held, f.outcome.sounds, noxServer.Frame()-f.frame)
	return true
}

// CheckPlayerHammer observes two complete attacks initiated only by actual
// mouse input. Stock inventory is equipped through the preceding GUI steps.
// Missing damage, animation, sound, quake or completion is a bounded failure.
func (sc *e2eScenario) CheckPlayerHammer(name string) {
	f := &e2ePlayerHammerFixture{}
	sc.addWhen(0, name+" prepare", 1200, func() bool {
		return noxServer.Players.HostUnit() != nil && noxClient.ClientPlayerUnit() != nil && nox_client_isConnected()
	}, f.prepare)
	for cycle := 1; cycle <= 2; cycle++ {
		prefix := fmt.Sprintf("%s cycle %d", name, cycle)
		// A hammer consumes all 100 stamina. Attack animation completion is
		// earlier than natural stamina recovery; do not inject or bypass it.
		sc.addWhen(12, prefix+" naturally recovered attack stamina", 120, func() bool {
			return f.held() && int32(f.unit.UpdateDataPlayer().Stamina) >=
				server.WeaponStaminaByType4F7E80(f.unit.ControllingPlayer().WeaponEquip)
		}, func() {})
		sc.add(0, prefix+" actual attack button", f.input)
		sc.Input(1, "", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false})
		sc.addWhen(1, prefix+" actual monster and NPC damage and client effects", 120, f.observeHit, func() {})
		sc.addWhen(0, prefix+" natural attack completion and equipment", 360, f.observeCompletion, func() {})
	}
	sc.add(0, name+" cleanup fixture", func() {
		f.active = false
		for _, target := range f.targets {
			noxServer.DelayedDelete(target)
		}
		asObjectS(f.unit).SetPos(f.original)
	})
}
