package opennox

import (
	"encoding/binary"
	"fmt"
	"image"
	"math"
	"time"

	"github.com/opennox/libs/client/keybind"
	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/player"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

// The stock ability HUD has five pointer-free 24-byte records after slot zero.
// Observe actual packet consumers; never inject ready/active/cooldown values.
type e2eAbilityHUD struct {
	ID, Ready, Active, Level, Start uint32
}

func e2eDecodeAbilityHUD(data []byte) (e2eAbilityHUD, error) {
	if len(data) != 24 {
		return e2eAbilityHUD{}, fmt.Errorf("ability HUD record size=%d, want 24", len(data))
	}
	return e2eAbilityHUD{
		ID: binary.LittleEndian.Uint32(data), Ready: binary.LittleEndian.Uint32(data[8:]),
		Active: binary.LittleEndian.Uint32(data[12:]), Level: binary.LittleEndian.Uint32(data[16:]),
		Start: binary.LittleEndian.Uint32(data[20:]),
	}, nil
}

func e2eReadAbilityHUD(ability server.Ability) e2eAbilityHUD {
	for slot := uintptr(1); slot <= 5; slot++ {
		record, err := e2eDecodeAbilityHUD(memmap.Slice(0x5D4594, 1047764+24*slot)[:24])
		if err == nil && record.ID == uint32(ability) {
			return record
		}
	}
	return e2eAbilityHUD{}
}

func e2eWarriorAbilityKey(ability server.Ability) (keybind.Key, error) {
	switch ability {
	case server.AbilityBerserk:
		return keybind.KeyA, nil
	case server.AbilityWarcry:
		return keybind.KeyS, nil
	case server.AbilityHarpoon:
		return keybind.KeyD, nil
	case server.AbilityTreadLightly:
		return keybind.KeyF, nil
	case server.AbilityInfravis:
		return keybind.KeyG, nil
	}
	return 0, fmt.Errorf("warrior ability ID=%d, want 1..5", ability)
}

// Find a nearby wall-free lane wide enough for an ordinary player collision.
// This only places the fixture; it neither drives nor bypasses ability movement.
func e2eWarriorAbilityArena(original types.Pointf, radius float32, trace func(types.Pointf, types.Pointf) bool) (types.Pointf, types.Pointf, error) {
	offsets := []types.Pointf{
		types.Ptf(160, 0), types.Ptf(-160, 0), types.Ptf(0, 160), types.Ptf(0, -160),
		types.Ptf(114, 114), types.Ptf(-114, 114), types.Ptf(114, -114), types.Ptf(-114, -114),
	}
	for ring := 0; ring <= 184; ring += 23 {
		for y := -ring; y <= ring; y += 23 {
			for x := -ring; x <= ring; x += 23 {
				if ring != 0 && x != -ring && x != ring && y != -ring && y != ring {
					continue
				}
				candidate := original.Add(types.Ptf(float32(x), float32(y)))
				if candidate != original && !trace(original, candidate) {
					continue
				}
				clear := true
				for direction := 0; direction < 256; direction++ {
					cosine, sine := server.SinCosDir(byte(direction))
					if !trace(candidate, candidate.Add(types.Ptf(radius*cosine, radius*sine))) {
						clear = false
						break
					}
				}
				if !clear {
					continue
				}
				for _, offset := range offsets {
					direction := offset.Normalize()
					side := types.Ptf(-direction.Y*radius, direction.X*radius)
					if trace(candidate, candidate.Add(offset)) &&
						trace(candidate.Add(side), candidate.Add(offset).Add(side)) &&
						trace(candidate.Sub(side), candidate.Add(offset).Sub(side)) {
						return candidate, direction, nil
					}
				}
			}
		}
	}
	return types.Pointf{}, types.Pointf{}, fmt.Errorf("no open warrior target lane near %v (radius=%g)", original, radius)
}

func e2eWarriorLaneMissesCircle(from, to, center types.Pointf, radius float32) bool {
	dx, dy := float64(to.X-from.X), float64(to.Y-from.Y)
	cx, cy := float64(center.X-from.X), float64(center.Y-from.Y)
	projection := 0.0
	if length := dx*dx + dy*dy; length != 0 {
		projection = math.Max(0, math.Min(1, (cx*dx+cy*dy)/length))
	}
	dx, dy = cx-projection*dx, cy-projection*dy
	return dx*dx+dy*dy > float64(radius)*float64(radius)
}

// Fixture lanes also avoid existing solid map props (including the Obelisk at
// the stock spawn). This conservative bound is not a gameplay collision test.
func e2eWarriorLaneClear(unit *server.Object, from, to types.Pointf) bool {
	if !noxServer.MapTraceRay(from, to, server.MapTraceFlag1) {
		return false
	}
	for obj := noxServer.Objs.First(); obj != nil; obj = obj.Next() {
		if obj == unit || obj.Flags().HasAny(object.FlagDestroyed|object.FlagNoCollide) {
			continue
		}
		var radius float32
		switch obj.Shape.Kind {
		case server.ShapeKindCircle:
			radius = obj.Shape.Circle.R
		case server.ShapeKindBox:
			box := obj.Shape.Box
			for _, corner := range []types.Pointf{
				types.Ptf(box.LeftTop, box.LeftBottom), types.Ptf(box.LeftBottom2, box.LeftTop2),
				types.Ptf(box.RightTop, box.RightBottom), types.Ptf(box.RightBottom2, box.RightTop2),
			} {
				radius = max(radius, float32(math.Hypot(float64(corner.X), float64(corner.Y))))
			}
		default:
			continue
		}
		if !e2eWarriorLaneMissesCircle(from, to, obj.PosVec, radius+unit.Shape.Circle.R+4) {
			return false
		}
	}
	return true
}

type e2eWarriorAbilityFixture struct {
	ability       server.Ability
	unit          *server.Object
	targets       []*server.Object
	origin        types.Pointf
	targetOrigin  types.Pointf
	health        uint16
	playerHealth  uint16
	exec          *server.ExecAbilityClass
	deadline      uint32
	startFrame    uint32
	startCooldown int
	retryFrame    uint32
	retryCooldown int
	hudActiveSeen bool
	harpoonSeen   bool
	lastEffectLog uint32
}

func (f *e2eWarriorAbilityFixture) record() *server.ExecAbilityClass {
	for record := noxServer.Abils.ExecHead(); record != nil; record = record.Next {
		if record.Unit == f.unit && record.Abil == f.ability {
			return record
		}
	}
	return nil
}

func e2eObjectInWorld(want *server.Object) bool {
	for obj := noxServer.Objs.First(); obj != nil; obj = obj.Next() {
		if obj == want {
			return true
		}
	}
	return false
}

func (f *e2eWarriorAbilityFixture) target(typeID string, pos types.Pointf) *server.Object {
	obj := noxServer.NewObjectByTypeID(typeID)
	if obj == nil {
		e2eError(fmt.Errorf("warrior %s: stock target %q missing", f.ability, typeID))
		return nil
	}
	noxServer.CreateObjectAt(obj, nil, pos)
	noxServer.ObjectsAddPending()
	if obj.HealthData == nil || obj.UpdateData == nil || obj.HealthData.Cur == 0 || !obj.Class().Has(object.ClassMonster) {
		e2eError(fmt.Errorf("warrior %s: stock target %q not initialized", f.ability, typeID))
		return nil
	}
	// Use ordinary waiting AI: NoUpdate would also make damage return early.
	// Velocity, collision, damage and buff ticks must remain active throughout.
	obj.UpdateDataMonster().SetAggression(0)
	obj.ClearActionStack()
	obj.MonsterPushAction(ai.ACTION_WAIT, noxServer.Frame()+250000)
	f.targets = append(f.targets, obj)
	e2eLog.Printf("WARRIOR TARGET: ability=%s type=%s class=%#x subclass=%#x status=%#x flags=%#x pos=%v", f.ability,
		typeID, uint32(obj.Class()), uint32(obj.SubClass()), uint32(obj.UpdateDataMonster().StatusFlags), uint32(obj.ObjFlags), obj.PosVec)
	return obj
}

func (f *e2eWarriorAbilityFixture) prepare() {
	f.unit = noxServer.Players.HostUnit()
	unit := f.unit
	if unit == nil || unit.UpdateData == nil || unit.HealthData == nil || unit.ControllingPlayer() == nil || unit.ControllingPlayer().PlayerClass() != player.Warrior {
		e2eError(fmt.Errorf("warrior fixture requires a live Warrior player"))
		return
	}
	if unit.Flags().HasAny(object.FlagDead|object.FlagDestroyed|object.FlagNoUpdate) ||
		noxServer.Abils.GetCooldownForUnit(unit, f.ability) != 0 || f.record() != nil {
		e2eError(fmt.Errorf("warrior %s: player not ready: flags=%#x cooldown=%d record=%p", f.ability, unit.ObjFlags, noxServer.Abils.GetCooldownForUnit(unit, f.ability), f.record()))
		return
	}
	if unit.ControllingPlayer().SpellLvl[f.ability] == 0 {
		e2eError(fmt.Errorf("warrior %s was not awarded by normal regular-game initialization", f.ability))
		return
	}
	delay, duration := noxServer.abilities.getDelay(f.ability), noxServer.abilities.getDuration(f.ability)
	if delay <= 0 || delay > 120000 || duration < 0 || duration > 120000 || duration == 0 && f.ability != server.AbilityHarpoon {
		e2eError(fmt.Errorf("warrior %s has invalid stock delay/duration: %d/%d", f.ability, delay, duration))
		return
	}
	// Assign only the quickbar slot through its existing UI setter. Activation
	// below uses real keyboard input and MSG_TRY_ABILITY, never abilities.Do.
	legacy.Nox_xxx_quickBarSetSpell(int(f.ability), int(f.ability)-1)
	original := unit.PosVec
	pos, direction, err := e2eWarriorAbilityArena(original, unit.Shape.Circle.R+4, func(from, to types.Pointf) bool {
		return e2eWarriorLaneClear(unit, from, to)
	})
	if err != nil {
		e2eError(err)
		return
	}
	asObjectS(unit).SetPos(pos)
	unit.VelVec, unit.ForceVec, unit.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	f.origin, f.playerHealth = unit.PosVec, unit.HealthData.Cur
	targetPos := unit.PosVec.Add(direction.Mul(112))
	f.targetOrigin = targetPos
	switch f.ability {
	case server.AbilityBerserk, server.AbilityHarpoon:
		if target := f.target("Troll", targetPos); target != nil {
			// A durable test target isolates ability damage from the death path.
			asObjectS(target).SetMaxHealth(2000)
			f.health, f.targetOrigin = target.HealthData.Cur, target.PosVec
		}
	case server.AbilityWarcry:
		// Most monsters (including Wolf) lack the stock WARCRY_STUN bit.
		// Select an eligible original type without changing its subclass.
		stunnableType := ""
		for _, typ := range noxServer.Types.List() {
			if typ.Class().Has(object.ClassMonster) && typ.SubClass().AsMonster().Has(object.MonsterWarcryStun) {
				stunnableType = typ.ID()
				break
			}
		}
		if stunnableType == "" {
			e2eError(fmt.Errorf("stock Warcry stunnable target missing"))
			return
		}
		caster := f.target("Necromancer", targetPos)
		stunnable := f.target(stunnableType, unit.PosVec.Add(direction.Mul(70)))
		f.target("Necromancer", unit.PosVec.Add(direction.Mul(420)))
		if caster != nil && !caster.UpdateDataMonster().StatusFlags.Has(object.MonStatusCanCastSpells) ||
			stunnable != nil && !stunnable.MonsterClass().Has(object.MonsterWarcryStun) {
			e2eError(fmt.Errorf("stock Warcry targets lack spellcaster/stunnable status"))
			return
		}
	case server.AbilityTreadLightly:
		bomber := f.target("Bomber", targetPos)
		if bomber != nil && (!bomber.MonsterClass().Has(object.MonsterBomber) || !noxServer.CanSee(bomber, unit, 0)) {
			e2eError(fmt.Errorf("Tread Lightly control target cannot initially detect Warrior"))
			return
		}
	case server.AbilityInfravis:
		target := f.target("Spider", targetPos)
		if target != nil {
			asObjectS(target).ApplyEnchant(server.ENCHANT_INVISIBLE, duration+1000, 1)
			if noxServer.CanSee(unit, target, 0) {
				e2eError(fmt.Errorf("Eye of the Wolf control target was already visible"))
				return
			}
		}
	}
	e2eLog.Printf("WARRIOR PREPARED: ability=%s player=%p pos=%v->%v targets=%d delay=%d duration=%d level=%d quickbar=%d HUD=%+v",
		f.ability, unit, original, f.origin, len(f.targets), delay, duration, unit.ControllingPlayer().SpellLvl[f.ability], legacy.Nox_xxx_quickBarSpell(int(f.ability)-1), e2eReadAbilityHUD(f.ability))
}

func (f *e2eWarriorAbilityFixture) aim() {
	mouse := noxClient.Viewport().ToScreenPos(image.Pt(int(f.targetOrigin.X), int(f.targetOrigin.Y)))
	noxClient.ChangeMousePos(mouse, true)
	e2eQueueInput(&seat.MouseMoveEvent{Pos: mouse, Relative: false})
	e2eLog.Printf("WARRIOR AIM: ability=%s target=%v mouse=%v", f.ability, f.targetOrigin, mouse)
}

func (f *e2eWarriorAbilityFixture) observeHUD() e2eAbilityHUD {
	hud := e2eReadAbilityHUD(f.ability)
	if hud.Active&(1<<f.ability) != 0 {
		f.hudActiveSeen = true
	}
	return hud
}

func (f *e2eWarriorAbilityFixture) effectReady() bool {
	hud := f.observeHUD()
	if now := noxServer.Frame(); f.unit != nil && (f.lastEffectLog == 0 || now-f.lastEffectLog >= 30) {
		f.lastEffectLog = now
		var targets []string
		for _, target := range f.targets {
			if e2eObjectInWorld(target) {
				targets = append(targets, fmt.Sprintf("%s health=%d/%d flags=%#x buffs=%#x pos=%v", target.ObjectTypeC().ID(), target.HealthData.Cur, target.HealthData.Max, uint32(target.ObjFlags), target.Buffs, target.PosVec))
			}
		}
		e2eLog.Printf("WARRIOR EFFECT WAIT: ability=%s frame=%d state=%d direction=%d cursor=%v health=%d/%d flags=%#x pos=%v velocity=%v record=%p HUD=%+v targets=%v", f.ability, now,
			f.unit.UpdateDataPlayer().State, f.unit.Direction1, f.unit.ControllingPlayer().CursorVec, f.unit.HealthData.Cur, f.unit.HealthData.Max, uint32(f.unit.ObjFlags), f.unit.PosVec, f.unit.VelVec, f.record(), hud, targets)
	}
	// Stock HarpoonDuration is zero: 004FBB70 allocates no timed record and
	// the HUD reports cooldown, not an active enchant. Observe its live bolt.
	if hud.Ready != 0 || f.ability != server.AbilityHarpoon && !f.hudActiveSeen || f.unit == nil {
		return false
	}
	for _, target := range f.targets {
		if !e2eObjectInWorld(target) {
			e2eError(fmt.Errorf("warrior %s: target disappeared before effect check", f.ability))
			return true
		}
	}
	switch f.ability {
	case server.AbilityBerserk:
		// GAME.EXE 004E8555 jumps over pain/stun for a live unit hit. The
		// self-damage path is for qualifying wall/solid-prop impacts instead.
		return len(f.targets) == 1 && f.targets[0].HealthData.Cur < f.health && f.unit.HealthData.Cur == f.playerHealth &&
			math.Hypot(float64(f.unit.PosVec.X-f.origin.X), float64(f.unit.PosVec.Y-f.origin.Y)) > 10
	case server.AbilityWarcry:
		return len(f.targets) == 3 && f.targets[0].HasEnchant(server.ENCHANT_ANTI_MAGIC) &&
			f.targets[1].HasEnchant(server.ENCHANT_HELD) && !f.targets[2].HasEnchant(server.ENCHANT_ANTI_MAGIC) && !f.targets[2].HasEnchant(server.ENCHANT_HELD)
	case server.AbilityHarpoon:
		if len(f.targets) != 1 {
			return false
		}
		update := f.unit.UpdateDataPlayer()
		if update.HarpoonTarg == f.targets[0] && update.HarpoonBolt != nil && f.targets[0].HealthData.Cur < f.health {
			if !f.harpoonSeen {
				f.harpoonSeen = true
				e2eLog.Printf("WARRIOR HARPOON ATTACHED: owner=%p bolt=%p target=%p health=%d->%d frame=%d", f.unit, update.HarpoonBolt, update.HarpoonTarg, f.health, f.targets[0].HealthData.Cur, noxServer.Frame())
			}
		}
		return f.harpoonSeen && math.Hypot(float64(f.targets[0].PosVec.X-f.targetOrigin.X), float64(f.targets[0].PosVec.Y-f.targetOrigin.Y)) > 8
	case server.AbilityTreadLightly, server.AbilityInfravis:
		enchant := server.ENCHANT_SNEAK
		visible := len(f.targets) == 1 && !noxServer.CanSee(f.targets[0], f.unit, 0)
		if f.ability == server.AbilityInfravis {
			enchant = server.ENCHANT_INFRAVISION
			visible = len(f.targets) == 1 && noxServer.CanSee(f.unit, f.targets[0], 0)
		}
		client := noxClient.ClientPlayerUnit()
		return visible && f.unit.HasEnchant(enchant) && f.unit.EnchantDur(enchant) > 0 &&
			f.unit.EnchantPower(enchant) == int(f.unit.ControllingPlayer().SpellLvl[f.ability]) && client != nil && client.HasEnchant(enchant)
	}
	return false
}

func (f *e2eWarriorAbilityFixture) ended() bool {
	if f.unit == nil || f.record() != nil || f.observeHUD().Active&(1<<f.ability) != 0 {
		return false
	}
	client := noxClient.ClientPlayerUnit()
	switch f.ability {
	case server.AbilityHarpoon:
		update := f.unit.UpdateDataPlayer()
		return update.HarpoonTarg == nil && update.HarpoonBolt == nil
	case server.AbilityTreadLightly:
		return !f.unit.HasEnchant(server.ENCHANT_SNEAK) && client != nil && !client.HasEnchant(server.ENCHANT_SNEAK) &&
			len(f.targets) == 1 && noxServer.CanSee(f.targets[0], f.unit, 0)
	case server.AbilityInfravis:
		return !f.unit.HasEnchant(server.ENCHANT_INFRAVISION) && client != nil && !client.HasEnchant(server.ENCHANT_INFRAVISION) &&
			len(f.targets) == 1 && !noxServer.CanSee(f.unit, f.targets[0], 0)
	}
	return f.unit.UpdateDataPlayer().State != server.PlayerState1
}

// CheckWarriorAbility exercises two complete cycles, including an extra real
// key press during cooldown. Targets/quickbar assignment are explicit fixtures;
// activation, attack frames, collision, effects, expiry and reports are live.
// Tread Lightly's stock duration is 99999; end it with an ordinary attack
// through the real input path, not a timer/deadline or direct cancellation.
func (sc *e2eScenario) CheckWarriorAbility(ability server.Ability, name string) {
	key, err := e2eWarriorAbilityKey(ability)
	if err != nil {
		panic(err)
	}
	for attempt := 1; attempt <= 2; attempt++ {
		f := &e2eWarriorAbilityFixture{ability: ability}
		label := fmt.Sprintf("%s attempt %d", name, attempt)
		sc.addWhen(0, label+" prepare", 1200, func() bool {
			unit := noxServer.Players.HostUnit()
			return unit != nil && unit.HealthData != nil && unit.UpdateData != nil &&
				!unit.HasEnchant(server.ENCHANT_INVULNERABLE) && !unit.HasEnchant(server.ENCHANT_HELD)
		}, f.prepare)
		sc.Wait(3, label+" synchronize fixture position and quickbar")
		sc.add(0, label+" aim by real mouse input", f.aim)
		sc.Wait(3, label+" synchronize aim")
		sc.Screen(label + " prepared")
		sc.Key(key, label+" activate by real keyboard input")
		sc.addWhen(0, label+" wait for server activation", 120, func() bool {
			return f.unit != nil && (f.record() != nil || ability == server.AbilityHarpoon && f.unit.UpdateDataPlayer().HarpoonBolt != nil) &&
				noxServer.Abils.GetCooldownForUnit(f.unit, ability) > 0
		}, func() {
			f.exec, f.startFrame = f.record(), noxServer.Frame()
			if f.exec != nil {
				f.deadline = f.exec.Frame
			}
			f.startCooldown = noxServer.Abils.GetCooldownForUnit(f.unit, ability)
			hud := f.observeHUD()
			duration := noxServer.abilities.getDuration(ability)
			invalidRecord := f.exec == nil || f.exec.Active != 1 || f.deadline-f.startFrame > uint32(duration)
			if duration == 0 && ability == server.AbilityHarpoon {
				invalidRecord = f.exec != nil || f.unit.UpdateDataPlayer().HarpoonBolt == nil || hud.Active&(1<<ability) != 0
			}
			if invalidRecord || f.startCooldown > noxServer.abilities.getDelay(ability) {
				e2eError(fmt.Errorf("warrior %s: bad active record/cooldown: %+v cooldown=%d frame=%d", ability, f.exec, f.startCooldown, f.startFrame))
				return
			}
			e2eLog.Printf("WARRIOR STARTED: ability=%s attempt=%d record=%p frame=%d deadline=%d cooldown=%d HUD=%+v", ability, attempt, f.exec, f.startFrame, f.deadline, f.startCooldown, e2eReadAbilityHUD(ability))
		})
		sc.addWhen(0, label+" wait for real gameplay effect", 240, f.effectReady, func() {
			var targets []string
			for _, target := range f.targets {
				targets = append(targets, fmt.Sprintf("%s health=%d/%d buffs=%#x pos=%v", target.ObjectTypeC().ID(), target.HealthData.Cur, target.HealthData.Max, target.Buffs, target.PosVec))
			}
			e2eLog.Printf("WARRIOR EFFECT: ability=%s attempt=%d frame=%d player-health=%d->%d player-pos=%v->%v targets=%v HUD-active-seen=%t", ability, attempt, noxServer.Frame(), f.playerHealth, f.unit.HealthData.Cur, f.origin, f.unit.PosVec, targets, f.hudActiveSeen)
		})
		sc.Screen(label + " effect")
		sc.add(0, label+" capture cooldown before repeat input", func() {
			f.retryFrame, f.retryCooldown = noxServer.Frame(), noxServer.Abils.GetCooldownForUnit(f.unit, ability)
			if f.retryCooldown <= 8 {
				e2eError(fmt.Errorf("warrior %s: no remaining cooldown for repeat input: %d", ability, f.retryCooldown))
			}
		})
		sc.Key(key, label+" repeat key during cooldown")
		sc.Wait(5, label+" process repeat input")
		sc.add(0, label+" verify cooldown did not restart", func() {
			elapsed := int(noxServer.Frame() - f.retryFrame)
			want, got := f.retryCooldown-elapsed, noxServer.Abils.GetCooldownForUnit(f.unit, ability)
			if got != want || f.record() != nil && (f.record() != f.exec || f.record().Frame != f.deadline) {
				e2eError(fmt.Errorf("warrior %s: repeat input restarted ability: cooldown=%d want=%d elapsed=%d record=%p original=%p", ability, got, want, elapsed, f.record(), f.exec))
				return
			}
			e2eLog.Printf("WARRIOR REPEAT REJECTED: ability=%s attempt=%d cooldown=%d->%d elapsed=%d", ability, attempt, f.retryCooldown, got, elapsed)
		})
		if ability == server.AbilityTreadLightly {
			// The original player update checks cancellation after handling
			// input from walking/running states. Start an ordinary walk first;
			// attacking straight from the idle shield stance is not that path.
			sc.Input(0, label+" walk by real mouse input while sneaking", &seat.MouseButtonEvent{Button: seat.MouseButtonRight, Pressed: true})
			sc.addWhen(0, label+" verify movement preserves sneaking", 120, func() bool {
				return f.unit.UpdateDataPlayer().State == server.PlayerState0 &&
					math.Hypot(float64(f.unit.PosVec.X-f.origin.X), float64(f.unit.PosVec.Y-f.origin.Y)) > 8
			}, func() {
				client := noxClient.ClientPlayerUnit()
				if !f.unit.HasEnchant(server.ENCHANT_SNEAK) || client == nil || !client.HasEnchant(server.ENCHANT_SNEAK) ||
					f.observeHUD().Active&(1<<ability) == 0 || noxServer.CanSee(f.targets[0], f.unit, 0) {
					e2eError(fmt.Errorf("Tread Lightly lost its effect during an ordinary walk"))
				}
				e2eLog.Printf("WARRIOR SNEAK WALK: attempt=%d frame=%d pos=%v->%v state=%d", attempt, noxServer.Frame(), f.origin, f.unit.PosVec, f.unit.UpdateDataPlayer().State)
			})
			sc.Input(0, label+" attack by real mouse input to end sneaking", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true})
			sc.Input(1, "", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false}, &seat.MouseButtonEvent{Button: seat.MouseButtonRight, Pressed: false})
		}
		endTimeout := time.Duration(120100)
		if ability == server.AbilityTreadLightly {
			endTimeout = 240
		}
		sc.addWhen(0, label+" wait for gameplay end", endTimeout, f.ended, func() {
			e2eLog.Printf("WARRIOR ENDED: ability=%s attempt=%d frame=%d state=%d cooldown=%d HUD=%+v", ability, attempt, noxServer.Frame(), f.unit.UpdateDataPlayer().State, noxServer.Abils.GetCooldownForUnit(f.unit, ability), e2eReadAbilityHUD(ability))
		})
		sc.addWhen(0, label+" wait for cooldown ready report", 120100, func() bool {
			return noxServer.Abils.GetCooldownForUnit(f.unit, ability) == 0 && e2eReadAbilityHUD(ability).Ready == 1
		}, func() {
			hud := e2eReadAbilityHUD(ability)
			if hud.Start != 0 || hud.Active&(1<<ability) != 0 || !f.ended() || f.unit.HealthData.Cur == 0 {
				e2eError(fmt.Errorf("warrior %s: not fully reset after cooldown: HUD=%+v", ability, hud))
				return
			}
			for _, target := range f.targets {
				if e2eObjectInWorld(target) {
					noxServer.DelayedDelete(target)
				}
			}
			e2eLog.Printf("WARRIOR READY: ability=%s attempt=%d frame=%d HUD=%+v", ability, attempt, noxServer.Frame(), hud)
		})
		sc.Wait(3, label+" retire only fixture targets")
	}
}
