package opennox

import (
	"fmt"
	"image"
	"math"
	"time"
	"unsafe"

	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/player"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"
	ns4 "github.com/opennox/noxscript/ns/v4"
	nsp "github.com/opennox/noxscript/ns/v4/spell"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func e2eSpellUnitMatrixPairs() [6][2]string {
	return [6][2]string{
		{"player", "monster"}, {"monster", "player"},
		{"player", "NPC"}, {"NPC", "player"},
		{"monster", "NPC"}, {"NPC", "monster"},
	}
}

func e2eSpellUnitMatrixID(kind string) (spell.ID, bool) {
	switch kind {
	case "fireball":
		return spell.SPELL_FIREBALL, true
	case "magic-missile":
		return spell.SPELL_MAGIC_MISSILE, true
	case "poison":
		return spell.SPELL_POISON, true
	case "death-ray":
		return spell.SPELL_DEATH_RAY, true
	default:
		_, id, _, ok := e2eMutualStatusMode(kind + "/player-to-npc")
		return id, ok
	}
}

// This is an application matrix, not a player input or autonomous AI test.
// Script casts, ordinary ownership, WAIT, placement and durable starting HP
// are explicit setup. Damage callbacks, missile targets, hit metadata, buffs,
// packets, expiry and rendering are never supplied by the fixture.
type e2eSpellUnitMatrixFixture struct {
	kind, from, to        string
	id                    spell.ID
	host, caster, target  *server.Object
	created               []*server.Object
	original              types.Pointf
	hostHP, hostMax       uint16
	casterHP, spectatorHP uint16
	idleInput             e2eLockIdleInput
	statusReported        bool
	fire                  *e2eFireballUnitFixture
	missiles              *e2eMagicMissileUnitFixture
	poison                *e2ePoisonFixture
	status                *e2eMutualStatusFixture
	ray                   *e2eSpellUnitMatrixRay
}

type e2eSpellUnitMatrixRay struct {
	active, hit, clientHit, clientFX, reported bool
	health                                     uint16
	damage                                     int32
	frame, carry                               uint32
	castAudio, particles                       int
	sparks                                     map[*client.Drawable]uint32
}

func (f *e2eSpellUnitMatrixFixture) prepare() {
	f.host = noxServer.Players.HostUnit()
	if f.host == nil || f.host.HealthData == nil || f.host.ControllingPlayer() == nil || f.host.Buffs != 0 || f.host.Poison540 != 0 {
		e2eError(fmt.Errorf("spell matrix requires a live unenchanted host"))
		return
	}
	f.original, f.hostHP, f.hostMax = f.host.PosVec, f.host.HealthData.Cur, f.host.HealthData.Max
	if f.to == "player" && f.host.UpdateDataPlayer().Player.ArmorEquip&0x3000000 != 0 {
		e2eError(fmt.Errorf("spell matrix HP-loss baseline requires an unshielded player"))
		return
	}
	distance := float32(160)
	origin, direction, err := e2eSpellUnitMatrixArena(f.original, distance, f.from != "player" && f.to != "player", func(a, b types.Pointf) bool { return e2eWarriorLaneClear(f.host, a, b) })
	if err != nil {
		e2eError(err)
		return
	}
	casterPos, targetPos, hostPos := origin, origin.Add(direction.Mul(distance)), origin
	if f.to == "player" {
		hostPos = origin.Add(direction.Mul(distance))
	} else if f.from != "player" {
		// The observer is farther from the caster than its spell target,
		// but beside the lane, not 352 units behind the wandering target.
		// Both sight lines and its stock body clearance are checked below.
		hostPos = e2eSpellUnitMatrixObserver(origin, direction)
	}
	asObjectS(f.host).SetPos(hostPos)
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	f.caster = f.makeUnit(f.from, casterPos)
	f.target = f.makeUnit(f.to, targetPos)
	noxServer.ObjectsAddPending()
	if f.caster == nil || f.target == nil {
		return
	}
	for _, u := range []*server.Object{f.caster, f.target} {
		if !e2eObjectInWorld(u) || u.HealthData == nil || u.UpdateData == nil || u.Buffs != 0 || u.Poison540 != 0 || u.Damage == nil || u.Flags().HasAny(object.FlagDead|object.FlagDestroyed|object.FlagNoUpdate|object.FlagNoCollide) {
			e2eError(fmt.Errorf("spell matrix stock unit is not ready: %s/%s unit=%p", f.from, f.to, u))
			return
		}
		if u != f.host {
			asObjectS(u).SetMaxHealth(2000)
			asObjectS(u).SetAggression(0)
			asObjectS(u).SetRetreatLevel(0)
			u.ClearActionStack()
			u.MonsterPushAction(ai.ACTION_WAIT, noxServer.Frame()+250000)
			if u.UpdateDataMonster().MonsterDef == nil {
				e2eError(fmt.Errorf("spell matrix stock MonsterDef is missing"))
				return
			}
		}
		if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(u)) <= math.MaxUint32 || uintptr(u.UpdateData) <= math.MaxUint32) {
			e2eError(fmt.Errorf("spell matrix unit/update pointer was narrowed"))
			return
		}
	}
	if !noxServer.S().IsEnemyTo(f.caster, f.target) || !noxServer.S().IsEnemyTo(f.target, f.caster) {
		e2eError(fmt.Errorf("spell matrix pair is not normally hostile: %s->%s", f.from, f.to))
		return
	}
	// Keep native callbacks and stock armor; no equipped protection modifiers
	// or immune/buff baseline is admitted into the exact HP-loss matrix.
	if f.target.Class().Has(object.ClassMonster) && uint32(f.target.SubClass())&0x400 != 0 {
		e2eError(fmt.Errorf("spell matrix baseline is fire-immune"))
		return
	}
	if f.kind == "poison" && f.target.Class().Has(object.ClassMonster) && uint32(f.target.SubClass())&0x200 != 0 {
		e2eError(fmt.Errorf("spell matrix poison HP-loss baseline is poison-immune"))
		return
	}
	for it := f.target.FirstItem(); it != nil; it = it.NextItem() {
		if !it.Flags().Has(object.FlagEquipped) || it.InitData == nil || !it.Class().HasAny(object.ClassArmor|object.ClassWeapon|object.ClassWand) {
			continue
		}
		for _, m := range it.InitDataModifier().Modifiers {
			if m != nil && (m.Defend76.Fnc != nil || m.Engage112 != nil) {
				e2eError(fmt.Errorf("spell matrix baseline has an equipped protection modifier"))
				return
			}
		}
	}
	if f.target == f.host && (f.kind == "fireball" || f.kind == "magic-missile" || f.kind == "death-ray") {
		asObjectS(f.host).SetMaxHealth(2000)
	}
	f.casterHP, f.spectatorHP = f.caster.HealthData.Cur, f.host.HealthData.Cur
	switch f.kind {
	case "fireball":
		f.prepareFireball()
	case "magic-missile":
		f.prepareMissiles()
	case "poison":
		f.preparePoison()
	case "death-ray":
		f.prepareDeathRay()
	default:
		f.prepareStatus()
	}
	noxServer.TickHook(f.observeControls)
	e2eLog.Printf("SPELL MATRIX PREPARED: kind=%s direction=%s->%s caster=%p target=%p owner=%p/%p callbacks=%p/%p hostile=true subclass=%x/%x pos=%v/%v", f.kind, f.from, f.to, f.caster, f.target, f.caster.ObjOwner, f.target.ObjOwner, f.caster.Damage, f.target.Damage, uint32(f.caster.SubClass()), uint32(f.target.SubClass()), f.caster.PosVec, f.target.PosVec)
}

// Check the actual three footprints instead of requiring one large empty
// circle. Stock rooms can fit the separated units without an 80-unit circle.
// No walls or props are removed to make a projectile reach its target.
func e2eSpellUnitMatrixArena(original types.Pointf, distance float32, spectator bool, clear func(types.Pointf, types.Pointf) bool) (types.Pointf, types.Pointf, error) {
	for ring := 0; ring <= 368; ring += 23 {
		for y := -ring; y <= ring; y += 23 {
			for x := -ring; x <= ring; x += 23 {
				if ring != 0 && x != -ring && x != ring && y != -ring && y != ring {
					continue
				}
				origin := original.Add(types.Ptf(float32(x), float32(y)))
				if origin != original && !clear(original, origin) {
					continue
				}
				for direction := byte(0); ; direction += 32 {
					cx, cy := server.SinCosDir(direction)
					dir := types.Ptf(cx, cy)
					target := origin.Add(dir.Mul(distance))
					side := types.Ptf(-cy*32, cx*32)
					open := clear(origin, target) && clear(origin.Add(side), target.Add(side)) && clear(origin.Sub(side), target.Sub(side))
					points := []types.Pointf{origin, target}
					if spectator {
						// Monster homing spells select the nearest real enemy.
						// The player observer must be farther than the target.
						observer := e2eSpellUnitMatrixObserver(origin, dir)
						open = open && clear(observer, origin) && clear(observer, target)
						points = append(points, observer)
					}
					for _, pos := range points {
						for _, offset := range []types.Pointf{types.Ptf(32, 0), types.Ptf(-32, 0), types.Ptf(0, 32), types.Ptf(0, -32), types.Ptf(32, 32), types.Ptf(-32, 32), types.Ptf(32, -32), types.Ptf(-32, -32)} {
							if !open || !clear(pos, pos.Add(offset)) {
								open = false
								break
							}
						}
						if !open {
							break
						}
					}
					if open {
						return origin, dir, nil
					}
					if direction == 224 {
						break
					}
				}
			}
		}
	}
	return types.Pointf{}, types.Pointf{}, fmt.Errorf("no open spell matrix lane near %v (distance=%g spectator=%t)", original, distance, spectator)
}

func e2eSpellUnitMatrixObserver(origin, direction types.Pointf) types.Pointf {
	return origin.Add(types.Ptf(-direction.Y, direction.X).Mul(192))
}

func (f *e2eSpellUnitMatrixFixture) makeUnit(kind string, pos types.Pointf) *server.Object {
	if kind == "player" {
		return f.host
	}
	id := "Troll"
	if f.kind == "poison" {
		// Stock Troll has POISON_IMMUNE (0x200). Keep that resistance
		// intact and use an ordinary susceptible monster for DOT checks.
		id = "Urchin"
	}
	if kind == "NPC" {
		id = "NPC"
	}
	u := noxServer.NewObjectByTypeID(id)
	if u == nil {
		e2eError(fmt.Errorf("missing stock spell matrix unit %s", id))
		return nil
	}
	var owner *server.Object
	if kind == "NPC" && (f.from == "monster" || f.to == "monster") {
		// A player-owned NPC is naturally opposed to an unowned monster.
		// For Player<->NPC the unowned NPC is instead the hostile side.
		owner = f.host
	}
	noxServer.CreateObjectAt(u, owner, pos)
	f.created = append(f.created, u)
	if (kind == "NPC") != u.SubClass().AsMonster().Has(object.MonsterNPC) || (kind == "NPC") != (u.Damage == f.host.Damage) {
		e2eError(fmt.Errorf("spell matrix stock damage/class registration changed: %s", kind))
	}
	return u
}

func (f *e2eSpellUnitMatrixFixture) observeControls() {
	active := f.fire != nil && f.fire.active || f.missiles != nil && f.missiles.active || f.poison != nil && f.poison.active || f.status != nil && f.status.active || f.ray != nil && f.ray.active
	if !active {
		return
	}
	if f.caster.HealthData.Cur != f.casterHP || f.caster.Buffs != 0 || f.caster.Poison540 != 0 ||
		f.host != f.target && (f.host.HealthData.Cur != f.spectatorHP || f.host.Buffs != 0 || f.host.Poison540 != 0) {
		e2eError(fmt.Errorf("spell matrix affected its caster/spectator: %s/%s->%s caster HP=%d/%d buffs=%x poison=%d; host HP=%d/%d buffs=%x poison=%d caster damage=%d source=%p frame=%d pos=%v targetPos=%v", f.kind, f.from, f.to,
			f.caster.HealthData.Cur, f.casterHP, f.caster.Buffs, f.caster.Poison540, f.host.HealthData.Cur, f.spectatorHP, f.host.Buffs, f.host.Poison540,
			f.caster.Field131, f.caster.Obj130, f.caster.Frame134, f.caster.PosVec, f.target.PosVec))
	}
}

func (f *e2eSpellUnitMatrixFixture) prepareFireball() {
	p := &e2eFireballUnitFixture{level: 3, actualLevel: 3, direction: f.from + "-to-" + f.to, fromNPC: f.target == f.host,
		host: f.host, caster: f.caster, target: f.target, hostMax: f.hostMax}
	f.fire = p
	noxServer.Audio.OnSound(func(id sound.ID, kind int, owner *server.Object, _ types.Pointf) {
		if !p.active {
			return
		}
		if id == sound.SoundFireballCast && owner == p.caster {
			p.castAudio++
			if kind != 0 || p.castAudio != 1 {
				e2eError(fmt.Errorf("spell matrix Fireball cast event repeated"))
			}
			if p.projectile == nil {
				p.observeProjectile()
			}
		}
		if id == sound.SoundFireballExplode && owner == p.projectile {
			p.explodeAudio++
			if kind != 0 || p.explodeAudio != 1 {
				e2eError(fmt.Errorf("spell matrix Fireball explosion event repeated"))
			}
		}
	})
	noxServer.TickHook(f.observeFireballHit)
}

func e2eSpellUnitMatrixMarker(u *server.Object) (uint32, uint32, uint32) {
	if u.Class().Has(object.ClassPlayer) {
		ud := u.UpdateDataPlayer()
		return ud.Field76, ud.Field75, ud.Field21
	}
	ud := u.UpdateDataMonster()
	return ud.Field547, ud.Field546, ud.Field1
}

func (f *e2eSpellUnitMatrixFixture) prepareDeathRay() {
	p := &e2eSpellUnitMatrixRay{damage: int32(noxServer.Balance.Float("DeathRayDamage"))}
	f.ray = p
	if ext := f.target.GetExt(); f.target.Class().Has(object.ClassMonster) && ext != nil && (ext.HealthRegenToMax > 0 || ext.HealthRegenPerFrame >= 0) {
		e2eError(fmt.Errorf("spell matrix Death Ray requires the stock regeneration baseline"))
		return
	}
	radius := float32(noxServer.Balance.Float("DeathRayOutRadius"))
	if p.damage <= 0 || p.damage >= 2000 || radius <= 0 || float64(radius+f.caster.Shape.Circle.R) >= f.caster.PosVec.Sub(f.target.PosVec).Len() {
		e2eError(fmt.Errorf("spell matrix Death Ray needs positive stock damage and a caster outside its radius: damage=%d radius=%g", p.damage, radius))
		return
	}
	snd := noxServer.Spells.DefByInd(f.id).GetCastSound()
	if snd == 0 {
		e2eError(fmt.Errorf("spell matrix stock Death Ray cast sound missing"))
		return
	}
	noxServer.Audio.OnSound(func(id sound.ID, kind int, owner *server.Object, _ types.Pointf) {
		if p.active && id == snd && owner == f.caster {
			p.castAudio++
			if kind != 0 || p.castAudio != 1 {
				e2eError(fmt.Errorf("spell matrix Death Ray cast event repeated"))
				return
			}
			f.observeDeathRayHit()
		}
	})
}

func (f *e2eSpellUnitMatrixFixture) observeDeathRayHit() {
	p := f.ray
	if !p.active || p.hit {
		return
	}
	marker, kind, carry := e2eSpellUnitMatrixMarker(f.target)
	if f.target.HealthData.Cur != p.health-uint16(p.damage) || f.target.Obj130 != f.caster || f.target.Field131 != uint32(object.DamageZapRay) ||
		f.target.Frame134 != p.frame || marker != 2 || kind != uint32(object.DamageZapRay) || carry != p.carry || f.target.Buffs != 0 {
		e2eError(fmt.Errorf("spell matrix Death Ray raw hit mismatch: %s->%s HP=%d/%d source=%p/%p type=%d marker=%d/%d frame=%d/%d carry=%x/%x",
			f.from, f.to, f.target.HealthData.Cur, p.health-uint16(p.damage), f.target.Obj130, f.caster, f.target.Field131, marker, kind, f.target.Frame134, p.frame, carry, p.carry))
		return
	}
	p.hit = true
	e2eLog.Printf("SPELL MATRIX DEATH RAY HIT: direction=%s->%s HP=%d->%d raw-damage=%d marker=%d/%d carry=%x", f.from, f.to, p.health, f.target.HealthData.Cur, p.damage, marker, kind, carry)
}

func e2eSpellUnitMatrixRayParticles(from, to image.Point) int {
	return (int(math.Hypot(float64(to.X-from.X), float64(to.Y-from.Y)))+1)/2 + 1
}

// Units naturally heal after one second without damage. The
// longest ray particle lasts 40 frames, so its actual expiry can outlast that
// grace period. Count completed stock regeneration ticks; never disable
// regeneration or restore HP to keep the raw-hit assertion green.
func e2eSpellUnitMatrixRayHealth(health, maximum, damage uint16, hitFrame, currentFrame, fps uint32, monster bool) uint16 {
	hp := uint32(health - damage)
	first := hitFrame + fps + 1
	if fps == 0 || maximum == 0 || currentFrame <= first {
		return uint16(hp)
	}
	interval := 300 * fps // sub_4F9ED0: player baseline
	if monster {
		interval = 180 * fps // monsterRegenerateHP, including NPC
	}
	period, regen := interval/uint32(maximum), uint32(1)
	if period == 0 {
		period = 1
		if monster {
			regen = uint32(maximum) / interval
		}
	}
	last := currentFrame - 1
	hp += (last/period - (first-1)/period) * regen
	if hp > uint32(maximum) {
		hp = uint32(maximum)
	}
	return uint16(hp)
}

func (f *e2eSpellUnitMatrixFixture) completeDeathRay() bool {
	p := f.ray
	if !p.hit || p.castAudio != 1 {
		return false
	}
	if !p.clientFX {
		want := e2eSpellUnitMatrixRayParticles(f.caster.PosVec.Point(), f.target.PosVec.Point())
		typ := uint32(noxClient.Things.IndByID("VioletSpark"))
		count := 0
		if !p.reported {
			p.reported = true
			pl := f.host.ControllingPlayer()
			e2eLog.Printf("SPELL MATRIX DEATH RAY CLIENT DIAGNOSTIC: cast-frame=%d server=%d client=%d type=%d cast-audio=%d camera-extent=%d/%d list4=%p", p.frame, noxServer.Frame(), noxClient.srv.Frame(), typ, p.castAudio, pl.Field10, pl.Field12, noxClient.Objs.List4)
		}
		for dr := noxClient.Objs.List4; dr != nil; dr = dr.Field_98 {
			effect := dr.UnionEffect()
			if dr.TypeIDVal != typ || effect.Field_111 < p.frame {
				continue
			}
			life := effect.Field_112 - effect.Field_111
			if !dr.Flags().Has(object.FlagActive) || dr.DrawFuncPtr == nil || effect.Field_111 > p.frame+15 || life < 20 || life > 40 ||
				unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= math.MaxUint32 {
				e2eError(fmt.Errorf("spell matrix Death Ray native FX is invalid: drawable=%p active=%t life=%d", dr, dr.Flags().Has(object.FlagActive), life))
				return true
			}
			p.sparks[dr] = effect.Field_111
			count++
		}
		if count == 0 {
			return false
		}
		if count != want {
			e2eError(fmt.Errorf("spell matrix Death Ray FX count=%d want=%d", count, want))
			return true
		}
		p.clientFX, p.particles = true, count
		path, err := e2eWriteMagicFrame("", noxClient.r.CopyPixBuffer())
		if err != nil {
			e2eError(err)
			return true
		}
		e2eLog.Printf("SPELL MATRIX DEATH RAY FX: direction=%s->%s native-particles=%d lifetime=20..40 frame=%s", f.from, f.to, count, path)
	}
	marker, kind, carry := e2eSpellUnitMatrixMarker(f.target)
	wantHP := e2eSpellUnitMatrixRayHealth(p.health, f.target.HealthData.Max, uint16(p.damage), p.frame, noxServer.Frame(), noxServer.TickRate(), f.target.Class().Has(object.ClassMonster))
	if f.target.HealthData.Cur != wantHP || f.target.Buffs != 0 || f.target.Poison540 != 0 ||
		f.target.Obj130 != f.caster || f.target.Field131 != uint32(object.DamageZapRay) || f.target.Frame134 != p.frame ||
		marker != 2 || kind != uint32(object.DamageZapRay) || carry != p.carry {
		e2eError(fmt.Errorf("spell matrix Death Ray target changed after its verified hit: HP=%d/%d buffs=%x poison=%d type=%d source=%p/%p damage-frame=%d/%d current-frame=%d", f.target.HealthData.Cur, wantHP, f.target.Buffs, f.target.Poison540, f.target.Field131, f.target.Obj130, f.caster, f.target.Frame134, p.frame, noxServer.Frame()))
		return true
	}
	if !p.clientHit {
		if noxServer.Frame() > p.frame+noxServer.TickRate()+1 {
			e2eError(fmt.Errorf("spell matrix Death Ray client HP loss was not observed before natural regeneration"))
			return true
		}
		dr := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.target)))
		if dr == nil {
			return false
		}
		if f.target == f.host {
			meter, ready := e2eClientHUDMeter(0)
			if !ready || meter.Current != uint32(p.health)-uint32(p.damage) || meter.Maximum != uint32(f.hostMax) {
				return false
			}
		} else if delta, ok := legacy.HealthChangeForDrawable(dr.NetCode32); !ok || int32(delta) != -p.damage {
			return false
		}
		p.clientHit = true
		e2eLog.Printf("SPELL MATRIX DEATH RAY CLIENT HIT: direction=%s->%s raw-HP-loss=%d frame=%d", f.from, f.to, p.damage, noxServer.Frame())
	}
	// Keep observing the real client list until every received particle
	// expires. A recycled address with a different birth frame is not the
	// original ray. No FX packet or drawable is constructed by this test.
	for spark, born := range p.sparks {
		for live := noxClient.Objs.List4; live != nil; live = live.Field_98 {
			if live == spark && live.UnionEffect().Field_111 == born {
				return false
			}
		}
	}
	return true
}

func (f *e2eSpellUnitMatrixFixture) observeFireballHit() {
	p := f.fire
	if !p.active || p.hit || p.projectile == nil || p.target.HealthData.Cur == p.health {
		return
	}
	marker, kind, carry := e2eSpellUnitMatrixMarker(p.target)
	if p.target.HealthData.Cur != p.health-uint16(p.damage) || p.target.Obj130 != p.projectile || p.target.Field131 != uint32(object.DamageFlame) ||
		p.target.Frame134 < p.frame || marker != 1 || kind != uint32(p.projectileType) || carry != p.carry ||
		p.target.Class().Has(object.ClassMonster) && !p.target.UpdateDataMonster().StatusFlags.Has(object.MonStatusOnFire) {
		e2eError(fmt.Errorf("spell matrix Fireball hit mismatch: %s->%s HP=%d/%d marker=%d/%d carry=%x/%x", f.from, f.to, p.target.HealthData.Cur, p.health-uint16(p.damage), marker, kind, carry, p.carry))
		return
	}
	p.hit = true
	e2eLog.Printf("SPELL MATRIX FIREBALL HIT: direction=%s->%s HP=%d->%d marker=%d/%d carry=%x", f.from, f.to, p.health, p.target.HealthData.Cur, marker, kind, carry)
}

func (f *e2eSpellUnitMatrixFixture) prepareMissiles() {
	p := &e2eMagicMissileUnitFixture{level: 3, actualLevel: 3, direction: f.from + "-to-" + f.to, fromNPC: f.target == f.host,
		host: f.host, caster: f.caster, target: f.target, hostMax: f.hostMax, projectiles: make(map[*server.Object]e2eMagicMissileIdentity)}
	f.missiles = p
	spl := noxServer.Spells.DefByInd(f.id)
	opts := spl.Def.Missiles.Level(p.actualLevel)
	p.projectileType, p.count = opts.Projectile, opts.Count
	if p.count <= 0 {
		p.count = int(noxServer.Balance.FloatInd("MagicMissileCount", p.actualLevel-1))
	}
	p.direct = int32(math.RoundToEven(float64(float32(noxServer.Balance.Float("MagicMissileDamage")))))
	p.splash = int32(math.RoundToEven(float64(float32(noxServer.Balance.Float("MagicMissileSplashDamage")))))
	p.radius = float32(noxServer.Balance.Float("MagicMissileRange"))
	if p.count <= 0 || p.count > 20 || p.direct <= 0 || p.splash <= 0 || p.radius <= 5 {
		e2eError(fmt.Errorf("spell matrix invalid stock missile definition"))
		return
	}
	noxServer.Audio.OnSound(func(id sound.ID, kind int, owner *server.Object, _ types.Pointf) {
		if !p.active {
			return
		}
		if id == spl.GetCastSound() && owner == p.caster {
			p.castAudio++
			if kind != 0 || p.castAudio != 1 {
				e2eError(fmt.Errorf("spell matrix missile cast event repeated"))
			}
			if len(p.projectiles) == 0 {
				p.observeProjectiles()
			}
		}
		if id == sound.SoundMagicMissileDetonate {
			if _, known := p.projectiles[owner]; known {
				f.observeMissileDetonation(owner, kind)
			}
		}
	})
	noxServer.TickHook(p.observeReports)
}

// DefaultDamage clears the Monster latch at 004E0B8F; nil-weapon splash
// records the raw type at 004E0FC9. PlayerDamage also resets Player/NPC entry.
// Non-immune Troll has no armor absorption/carry processing in DefaultDamage.
func e2eSpellUnitMatrixExplosionMarker(targetKind string, splash bool, missileType uint32) (uint32, uint32) {
	if splash {
		return 2, uint32(object.DamageExplosion)
	}
	return 1, missileType
}

func (f *e2eSpellUnitMatrixFixture) observeMissileDetonation(missile *server.Object, kind int) {
	p := f.missiles
	identity := p.projectiles[missile]
	rawSplash, inRange := e2eMagicMissileSplash(p.splash, p.radius, missile.PosVec, p.target.PosVec)
	armor, carry := p.readArmorCarry()
	direct, nextCarry := e2eMagicMissileDamage(p.direct, p.armor, p.carry)
	splashApplied := inRange && noxServer.MapTraceRay(missile.PosVec, p.target.PosVec, server.MapTraceFlag1)
	splash := int32(0)
	if splashApplied {
		splash, nextCarry = e2eMagicMissileDamage(rawSplash, p.armor, nextCarry)
	}
	damage := direct + splash
	wantMarker, wantType := e2eSpellUnitMatrixExplosionMarker(f.to, splashApplied, uint32(missile.TypeInd))
	marker, markerType, _ := e2eSpellUnitMatrixMarker(p.target)
	if kind != 0 || identity.hit || noxServer.Frame()-missile.Field32 > 3*noxServer.TickRate() ||
		missile.UpdateDataMissile().Target != p.target || missile.UpdateDataMissile().Owner != p.caster ||
		p.lastHP <= uint16(damage) || p.target.HealthData.Cur != p.lastHP-uint16(damage) || armor != p.armor || math.Float32bits(carry) != math.Float32bits(nextCarry) ||
		p.target.Obj130 != missile || p.target.Field131 != uint32(object.DamageExplosion) || p.target.Frame134 != noxServer.Frame() || marker != wantMarker || markerType != wantType ||
		f.to == "monster" && (p.armor != 0 || math.Float32bits(p.carry) != math.Float32bits(carry)) {
		e2eError(fmt.Errorf("spell matrix missile mismatch: %s->%s HP=%d->%d want=%d direct/splash=%d/%d marker=%d/%d want=%d/%d carry=%g/%g", f.from, f.to, p.lastHP, p.target.HealthData.Cur, p.lastHP-uint16(damage), direct, splash, marker, markerType, wantMarker, wantType, carry, nextCarry))
		return
	}
	identity.hit = true
	p.projectiles[missile] = identity
	p.hits++
	p.totalDamage += damage
	p.lastHP, p.carry = p.target.HealthData.Cur, nextCarry
	e2eLog.Printf("SPELL MATRIX MISSILE HIT: direction=%s->%s hit=%d/%d direct=%d splash=%d HP=%d carry=%g", f.from, f.to, p.hits, p.count, direct, splash, p.lastHP, p.carry)
}

func (f *e2eSpellUnitMatrixFixture) preparePoison() {
	p := &e2ePoisonFixture{level: 3, direction: f.from + "-to-" + f.to, fromNPC: f.target == f.host,
		host: f.host, caster: f.caster, target: f.target}
	f.poison = p
	if !noxServer.Spells.HasFlags(f.id, things.SpellTargeted) {
		e2eError(fmt.Errorf("spell matrix Poison is not stock targeted"))
		return
	}
	noxServer.Audio.OnSound(func(id sound.ID, kind int, owner *server.Object, _ types.Pointf) {
		if !p.active {
			return
		}
		if id == sound.SoundPoisonCast && owner == p.caster {
			p.castAudio++
			if kind != 0 || p.castAudio != 1 {
				e2eError(fmt.Errorf("spell matrix Poison cast event repeated"))
			}
			p.observeProjectile()
		}
		if id == sound.SoundPoisonEffect && owner == p.target {
			p.effectAudio++
			if kind != 0 || p.effectAudio != 1 {
				e2eError(fmt.Errorf("spell matrix Poison effect event repeated"))
			}
			p.observeEffect()
		}
	})
	noxServer.TickHook(f.observePoisonHit)
}

func (f *e2eSpellUnitMatrixFixture) observePoisonHit() {
	p := f.poison
	if p.active && !p.effectSeen && (noxServer.Frame()-p.frame == 120 || noxServer.Frame()-p.frame == 240) {
		live := p.magic != nil && e2eFistInWorld(p.magic, p.magicWire, p.magicScript)
		e2eLog.Printf("SPELL MATRIX POISON WAIT: elapsed=%d magic-live=%t dose=%d/%d timer=%d HP=%d effect-audio=%d", noxServer.Frame()-p.frame, live, p.target.Poison540, p.power, p.target.Field542, p.target.HealthData.Cur, p.effectAudio)
		if live {
			e2eLog.Printf("SPELL MATRIX POISON PROJECTILE WAIT: pos=%v velocity=%v target=%p flags=%x target-pos=%v target-flags=%x", p.magic.PosVec, p.magic.VelVec, p.magic.UpdateDataSpellProjectile().Target, uint32(p.magic.Flags()), p.target.PosVec, uint32(p.target.Flags()))
		}
	}
	if !p.active || !p.effectSeen || p.hitSeen || p.target.HealthData.Cur == p.health {
		return
	}
	marker, kind, _ := e2eSpellUnitMatrixMarker(p.target)
	if p.target.HealthData.Cur != p.health-1 || p.target.Poison540 != p.power || p.target.Frame134 != p.firstTick || p.target.HealthData.Field16 != p.applied ||
		p.target.Obj130 != nil || p.target.Field131 != uint32(object.DamagePoison) || p.target.Pos132 != (types.Pointf{}) || marker != 2 || kind != uint32(object.DamagePoison) {
		e2eError(fmt.Errorf("spell matrix Poison DOT mismatch: %s->%s HP=%d->%d frame=%d/%d marker=%d/%d", f.from, f.to, p.health, p.target.HealthData.Cur, p.target.Frame134, p.firstTick, marker, kind))
		return
	}
	p.hitSeen = true
	e2eLog.Printf("SPELL MATRIX POISON HIT: direction=%s->%s HP=%d->%d application=%d first-DOT=%d dose=%d", f.from, f.to, p.health, p.target.HealthData.Cur, p.applied, p.firstTick, p.power)
}

func (f *e2eSpellUnitMatrixFixture) prepareStatus() {
	p := &e2eMutualStatusFixture{kind: f.kind, id: f.id, fromNPC: f.target == f.host, host: f.host, caster: f.caster, target: f.target}
	f.status = p
	key := map[string]string{"confused": "ConfuseEnchantDuration", "stun": "StunEnchantDuration", "slow": "SlowEnchantDuration"}[f.kind]
	balance := float64(0)
	if key != "" {
		balance = noxServer.Balance.Float(key)
	}
	isPlayer := f.target == f.host
	warrior := isPlayer && f.host.ControllingPlayer().PlayerClass() == player.Warrior
	p.buff, p.duration = e2eMutualStatusExpected(f.kind, isPlayer, warrior, f.target.Mass, balance, int(noxServer.TickRate()))
	if p.duration <= 24 || p.duration > 2900 {
		e2eError(fmt.Errorf("spell matrix stock status timer is outside bounds"))
		return
	}
	p.targeted = noxServer.Spells.HasFlags(f.id, things.SpellTargeted)
	p.power = 3
	if p.targeted {
		p.power = int(noxServer.Server.SpellPower4FE7B0(f.id, f.caster))
	}
	p.castSound = noxServer.Spells.DefByInd(f.id).GetAudio(0)
	p.onSound = noxServer.Spells.DefByInd(p.buff.Spell()).GetAudio(1)
	p.offSound = noxServer.Spells.DefByInd(p.buff.Spell()).GetAudio(2)
	noxServer.Audio.OnSound(func(id sound.ID, kind int, owner *server.Object, _ types.Pointf) {
		if !p.active {
			return
		}
		if id == p.castSound && id != 0 && owner == p.caster {
			p.castCount++
			if kind != 0 || p.castCount != 1 {
				e2eError(fmt.Errorf("spell matrix status cast event repeated"))
			}
			if p.targeted {
				p.observeProjectile()
			}
		}
		if id == p.onSound && id != 0 && owner == p.target {
			p.onCount++
			p.observeEffect()
			if kind != 0 || p.onCount != 1 {
				e2eError(fmt.Errorf("spell matrix status on event repeated"))
			}
		}
		if id == p.offSound && id != 0 && owner == p.target {
			p.offCount++
			if kind != 0 || p.offCount != 1 {
				e2eError(fmt.Errorf("spell matrix status off event repeated"))
			}
		}
	})
	noxServer.TickHook(func() {
		if p.active && !p.applied && p.target.HasEnchant(p.buff) {
			p.observeEffect()
		}
	})
}

func (f *e2eSpellUnitMatrixFixture) beginCast() {
	if !noxServer.MapTraceRay(f.caster.PosVec, f.target.PosVec, server.MapTraceFlag1) ||
		noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.target))) == nil ||
		noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.caster))) == nil {
		e2eError(fmt.Errorf("spell matrix units are not published/visible"))
		return
	}
	if f.fire != nil {
		f.fire.beginCast()
		return
	}
	if f.ray != nil {
		// E2E steps run in mainloopPre, before the server resets Kind1's
		// transient message list. An instant ray must be cast by an ordinary
		// script timer inside ActRun, after reset and before client replay.
		if noxServer.noxScriptP().NewTimer(ns4.Frames(1), f.beginDeathRayCast) == nil {
			e2eError(fmt.Errorf("spell matrix Death Ray script timer missing"))
		}
		return
	}
	f.caster.SetDir(server.DirFromVec(f.target.PosVec.Sub(f.caster.PosVec)))
	if p := f.missiles; p != nil {
		p.health, p.lastHP, p.frame = f.target.HealthData.Cur, f.target.HealthData.Cur, noxServer.Frame()
		p.armor, p.carry = p.readArmorCarry()
		sample := unitHealthSampleNative4D8760(f.target, int(f.host.ControllingPlayer().PlayerIndex()))
		if sample == nil || *sample != p.health {
			e2eError(fmt.Errorf("spell matrix missile HP report baseline did not settle"))
			return
		}
		p.reportHP, p.active = *sample, true
		if f.caster == f.host {
			// Normal player cursor/aim, not a manufactured missile target.
			ud := f.host.UpdateDataPlayer()
			ud.CursorObj, ud.Player.Obj3640 = f.target, f.target
			ud.Field55, ud.Field56 = int(f.target.PosVec.X), int(f.target.PosVec.Y)
			ud.Player.CursorVec = image.Pt(int(f.target.PosVec.X), int(f.target.PosVec.Y))
			mouse := noxClient.Viewport().ToScreenPos(ud.Player.CursorVec)
			noxClient.ChangeMousePos(mouse, true)
			e2eQueueInput(&seat.MouseMoveEvent{Pos: mouse})
		}
	}
	if p := f.poison; p != nil {
		p.health, p.frame, p.active = f.target.HealthData.Cur, noxServer.Frame(), true
	}
	if p := f.status; p != nil {
		p.health, p.casterHealth = f.target.HealthData.Cur, f.caster.HealthData.Cur
		p.frame, p.active = noxServer.Frame(), true
		if p.buff == server.ENCHANT_CONFUSED || p.buff == server.ENCHANT_HELD {
			ref := legacy.AsImageRefP(*memmap.PtrPtr(0x5D4594, 1096456))
			if ref == nil || ref.Kind() != 2 || len(ref.Field24ptr().Images()) < 2 {
				e2eError(fmt.Errorf("spell matrix status animation cache missing"))
				return
			}
			p.visual = &e2ePlayerStatusAnimation{kind: p.kind, buff: p.buff, unit: f.target, ref: ref, first: -1}
		}
	}
	api := noxServer.noxScriptP()
	api.CastSpellLvl(nsp.Spell(f.id.String()), 3, api.toObj(f.caster), api.toObj(f.target))
	noxServer.ObjectsAddPending()
	if p := f.missiles; p != nil && len(p.projectiles) == 0 {
		p.observeProjectiles()
	}
}

func (f *e2eSpellUnitMatrixFixture) beginDeathRayCast() {
	p := f.ray
	p.health, p.frame, p.active = f.target.HealthData.Cur, noxServer.Frame(), true
	_, _, p.carry = e2eSpellUnitMatrixMarker(f.target)
	p.sparks = make(map[*client.Drawable]uint32)
	api := noxServer.noxScriptP()
	api.CastSpellLvl(nsp.Spell(f.id.String()), 3, api.toObj(f.caster), api.toObj(f.target))
}

func (f *e2eSpellUnitMatrixFixture) keepAlive() {
	// Each short case must also renew ordinary input activity: resetting a
	// per-case 300-frame timer meant it never fired across 54 short waits.
	// Same-position motion leaves aim, movement and all spell results intact.
	f.idleInput.observe(noxServer.Frame(), noxClient.Inp.GetMousePos(), e2eQueueInput)
}

func (f *e2eSpellUnitMatrixFixture) completeStatus() bool {
	p := f.status
	if !p.applied || noxServer.Frame() < p.appliedFrame+uint32(p.duration)+8 {
		return false
	}
	dr := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.target)))
	cr := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.caster)))
	if !f.statusReported {
		f.statusReported = true
		drawBuffs := func(dr *client.Drawable) uint32 {
			if dr == nil {
				return 0
			}
			return uint32(dr.Buffs)
		}
		e2eLog.Printf("SPELL MATRIX STATUS EXPIRY DIAGNOSTIC: kind=%s direction=%s->%s frame=%d applied=%d duration=%d native-buffs=%x/%x client=%p/%p client-buffs=%x/%x audio=%d/%d/%d expected-audio=%d/%d/%d magic=%p server-live=%t client-live=%t targetPos=%v hostPos=%v",
			f.kind, f.from, f.to, noxServer.Frame(), p.appliedFrame, p.duration, f.target.Buffs, f.caster.Buffs, dr, cr, drawBuffs(dr), drawBuffs(cr),
			p.castCount, p.onCount, p.offCount, p.castSound, p.onSound, p.offSound, p.magic,
			p.magic != nil && e2eFistInWorld(p.magic, p.magicWire, p.magicScript), noxClient.Objs.ByNetCode(uint16(p.magicWire)) != nil, f.target.PosVec, f.host.PosVec)
	}
	if dr == nil || cr == nil || f.target.Buffs != 0 || dr.Buffs != 0 || f.caster.Buffs != 0 || cr.Buffs != 0 || f.target.EnchantDur(p.buff) != 0 ||
		p.targeted && (p.magic == nil || e2eFistInWorld(p.magic, p.magicWire, p.magicScript) || noxClient.Objs.ByNetCode(uint16(p.magicWire)) != nil) {
		return false
	}
	if p.castSound != 0 && p.castCount != 1 || p.onSound != 0 && p.onCount != 1 || p.offSound != 0 && p.offCount != 1 {
		return false
	}
	if f.target.HealthData.Cur != p.health || f.caster.HealthData.Cur != p.casterHealth || f.target == f.host && memmap.Uint32(0x5D4594, 1062540) != 0 {
		e2eError(fmt.Errorf("spell matrix status expiry changed HP or remains in HUD"))
		return true
	}
	if p.visual != nil {
		_, matched, total := e2eMutualStatusBirdiesPixels(p.visual.ref, dr)
		if total >= 10 && matched*100 >= total*80 {
			e2eError(fmt.Errorf("spell matrix status sprite remains after expiry"))
			return true
		}
	}
	return true
}

func (f *e2eSpellUnitMatrixFixture) completePoison() bool {
	p := f.poison
	if !p.hitSeen || p.castAudio != 1 || p.effectAudio != 1 || p.magic == nil || e2eFistInWorld(p.magic, p.magicWire, p.magicScript) || noxClient.Objs.ByNetCode(uint16(p.magicWire)) != nil {
		return false
	}
	if f.target.HealthData.Cur != p.health-1 || f.target.Poison540 != p.power {
		e2eError(fmt.Errorf("spell matrix Poison lost its verified first DOT"))
		return true
	}
	dr := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.target)))
	if dr == nil {
		return false
	}
	if f.target == f.host {
		meter, ready := e2eClientHUDMeter(0)
		if !ready || !meter.Poisoned || !meter.PoisonTubeReady || meter.Current != uint32(p.health-1) || meter.Maximum != uint32(f.hostMax) || f.host.UpdateDataPlayer().Player.Field3680&0x400 == 0 {
			return false
		}
		if _, err := e2eAssertHUDMeterPixels(noxClient.r.CopyPixBuffer(), meter, 'g', "spell matrix Poison"); err != nil {
			e2eError(err)
			return true
		}
	} else if delta, got := legacy.HealthChangeForDrawable(dr.NetCode32); !got || delta != -1 {
		return false
	}
	return true
}

func (f *e2eSpellUnitMatrixFixture) complete() bool {
	f.keepAlive()
	ok := false
	switch {
	case f.fire != nil:
		ok = f.fire.complete()
	case f.missiles != nil:
		ok = f.missiles.complete()
	case f.poison != nil:
		ok = f.completePoison()
	case f.status != nil:
		ok = f.completeStatus()
	case f.ray != nil:
		ok = f.completeDeathRay()
	}
	if !ok {
		return false
	}
	f.observeControls()
	path, err := e2eWriteMagicFrame("", noxClient.r.CopyPixBuffer())
	if err != nil {
		e2eError(err)
		return true
	}
	e2eLog.Printf("SPELL MATRIX PASS: kind=%s direction=%s->%s server/client=verified caster/spectator=unchanged frame=%s", f.kind, f.from, f.to, path)
	return true
}

func (f *e2eSpellUnitMatrixFixture) cleanup() {
	if f.fire != nil {
		f.fire.active = false
	}
	if f.missiles != nil {
		f.missiles.active = false
	}
	if f.poison != nil {
		f.poison.active = false
		// Ordinary cure is cleanup only, after real application/DOT/client HP.
		noxServer.Server.RemovePoison4EE9D0(f.target)
		if f.target.Poison540 != 0 || f.target.HealthData.Field16 != 0 {
			e2eError(fmt.Errorf("spell matrix ordinary poison cure failed"))
			return
		}
	}
	if f.status != nil {
		f.status.active = false
	}
	if f.ray != nil {
		f.ray.active = false
	}
	for _, u := range f.created {
		noxServer.DelayedDelete(u)
	}
	asObjectS(f.host).SetPos(f.original)
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	if f.target == f.host {
		asObjectS(f.host).SetMaxHealth(int(f.hostMax))
		asObjectS(f.host).SetHealth(int(f.hostHP))
	}
}

func (sc *e2eScenario) CheckSpellUnitMatrix(kind, name string) {
	id, ok := e2eSpellUnitMatrixID(kind)
	if !ok {
		e2eError(fmt.Errorf("invalid spell unit matrix kind %q", kind))
		return
	}
	for _, pair := range e2eSpellUnitMatrixPairs() {
		f := &e2eSpellUnitMatrixFixture{kind: kind, id: id, from: pair[0], to: pair[1]}
		label := name + " " + f.from + "->" + f.to
		sc.addWhen(0, label+" prepare", 1200, func() bool {
			u := noxServer.Players.HostUnit()
			return u != nil && u.Buffs == 0 && u.Poison540 == 0 && nox_client_isConnected() && noxClient.ClientPlayerUnit() != nil
		}, f.prepare)
		sc.Wait(12, label+" publish original units")
		sc.add(0, label+" normal script cast", f.beginCast)
		if kind != "fireball" && kind != "magic-missile" && kind != "poison" && kind != "death-ray" {
			for _, sample := range []struct {
				label string
				dt    uint32
			}{{"first", 12}, {"advanced", 24}} {
				sc.addWhen(0, label+" "+sample.label+" visible", 300, func() bool {
					p := f.status
					dr := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.target)))
					return p.applied && noxServer.Frame() >= p.appliedFrame+sample.dt && dr != nil && dr.HasEnchant(p.buff)
				}, func() { f.status.sample(sample.label) })
			}
		}
		timeout := time.Duration(300)
		if kind != "fireball" && kind != "magic-missile" && kind != "poison" && kind != "death-ray" {
			timeout = 3000
		}
		sc.addWhen(1, label+" verify real result/expiry", timeout, f.complete, f.cleanup)
		sc.addWhen(6, label+" ordinary cleanup replay", 120, func() bool {
			if f.target != f.host || f.poison == nil {
				return true
			}
			meter, ready := e2eClientHUDMeter(0)
			return ready && !meter.Poisoned && f.host.UpdateDataPlayer().Player.Field3680&0x400 == 0
		}, func() {})
	}
}
