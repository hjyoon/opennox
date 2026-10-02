package opennox

import (
	"fmt"
	"image"
	"math"

	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/player"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func e2eShockRetaliationMode(mode string) bool {
	return mode == "to-npc" || mode == "to-player"
}

// The naked stock fixture has no armor/protection modifiers. Independently
// reproduce the original binary32 spills and nearest/even fractional HP carry.
func e2eShockRetaliationAmount(balance float64, carry float32) (uint16, bool) {
	raw := math.RoundToEven(float64(float32(balance)))
	if math.IsNaN(raw) || math.IsInf(raw, 0) || raw <= 0 || raw > math.MaxUint16 ||
		math.IsNaN(float64(carry)) || math.IsInf(float64(carry), 0) || math.Abs(float64(carry)) > 0.5 {
		return 0, false
	}
	amount := math.RoundToEven(float64(float32(raw) + carry))
	if amount == 0 {
		amount = 1
	}
	if amount > math.MaxUint16 {
		return 0, false
	}
	return uint16(amount), true
}

type e2eShockRetaliationFixture struct {
	mode                      string
	spell                     e2eShockFixture
	attacker, weapon          *server.Object
	armor                     []*server.Object
	active, hit, held         bool
	frame, struckFrame        uint32
	health, damage, hitHealth uint16
	incomingHealth            uint16
	shockedAudio              int
}

func (f *e2eShockRetaliationFixture) prepare() {
	spellMode := "player"
	if f.mode == "to-player" {
		spellMode = "npc"
	}
	f.spell = e2eShockFixture{mode: spellMode, level: 3}
	f.spell.prepare()
	host, npc := f.spell.host, f.spell.npc
	if host.ControllingPlayer().PlayerClass() != player.Warrior ||
		host.UpdateDataPlayer().EquippedWeapon == nil ||
		host.UpdateDataPlayer().EquippedWeapon.ObjectTypeC().ID() != "Longsword" ||
		!noxServer.IsEnemyTo(host, npc) || !noxServer.IsEnemyTo(npc, host) {
		e2eError(fmt.Errorf("Shock retaliation needs a stock Warrior Longsword and hostile NPC"))
		return
	}
	// Use ordinary inventory services for fixture equipment, not equip flags
	// or armor coefficient stores. Restore the host's equipment at the end.
	for item := host.InvFirstItem; item != nil; item = item.InvNextItem {
		if item.Class().Has(object.ClassArmor) && item.Flags().Has(object.FlagEquipped) {
			if !asObjectS(host).Unequip(item) {
				e2eError(fmt.Errorf("Shock fixture could not unequip stock armor"))
				return
			}
			f.armor = append(f.armor, item)
		}
	}
	direction := npc.PosVec.Sub(host.PosVec).Normalize()
	distance := host.Shape.Circle.R + npc.Shape.Circle.R + 8
	asObjectS(npc).SetPos(host.PosVec.Add(direction.Mul(distance)))
	npc.LookAt(host.PosVec)
	f.attacker, f.weapon = host, host.UpdateDataPlayer().EquippedWeapon
	if f.mode == "to-npc" {
		f.attacker = npc
		f.weapon = noxServer.NewObjectByTypeID("Longsword")
		if f.weapon == nil {
			e2eError(fmt.Errorf("Shock fixture has no stock Longsword"))
			return
		}
		noxServer.CreateObjectAt(f.weapon, nil, npc.PosVec)
		noxServer.ObjectsAddPending()
		if !asObjectS(npc).Equip(f.weapon) || npc.NPCEquippedWeapon538960() != f.weapon {
			e2eError(fmt.Errorf("Shock fixture could not normally equip NPC sword"))
			return
		}
	}
	if !f.weapon.Flags().Has(object.FlagEquipped) || f.weapon.InvHolder != f.attacker {
		e2eError(fmt.Errorf("Shock attacker has no real equipped weapon"))
		return
	}
	noxServer.Audio.OnSound(func(id sound.ID, kind int, owner *server.Object, _ types.Pointf) {
		if !f.active || id != sound.SoundShocked || owner != f.attacker {
			return
		}
		f.shockedAudio++
		f.struckFrame = noxServer.Frame()
		if kind != 0 || f.shockedAudio != 1 || !f.spell.target.HasEnchant(server.ENCHANT_SHOCK) {
			e2eError(fmt.Errorf("Shock retaliation audio/order mismatch"))
			return
		}
		if f.mode == "to-npc" {
			if f.attacker.MonsterActionGet50A020() != ai.ACTION_MELEE_ATTACK || f.attacker.NPCEquippedWeapon538960() != f.weapon {
				e2eError(fmt.Errorf("Shock did not use the real NPC melee animation"))
				return
			}
			e2eLog.Printf("SHOCK NPC MELEE FRAME: elapsed=%d stored-frame=%d weapon=%p", noxServer.Frame()-f.attacker.Field34, uint8(f.attacker.UpdateDataMonster().Field481), f.weapon)
		}
	})
	noxServer.TickHook(f.observe)
}

func (f *e2eShockRetaliationFixture) attack() {
	if !f.spell.applied || f.spell.onAudio != 1 || f.spell.offAudio != 0 || f.attacker.Buffs != 0 {
		e2eError(fmt.Errorf("Shock retaliation placement/cast did not settle"))
		return
	}
	var carry float32
	if f.attacker.Class().Has(object.ClassPlayer) {
		carry = math.Float32frombits(f.attacker.UpdateDataPlayer().Field21)
	} else {
		carry = math.Float32frombits(f.attacker.UpdateDataMonster().Field1)
	}
	var ok bool
	f.damage, ok = e2eShockRetaliationAmount(noxServer.Balance.FloatInd("ShockDamage", 4), carry)
	if !ok || f.damage >= f.attacker.HealthData.Cur {
		e2eError(fmt.Errorf("Shock retaliation is not a survivable stock fixture"))
		return
	}
	f.frame, f.health, f.active = noxServer.Frame(), f.attacker.HealthData.Cur, true
	if f.mode == "to-npc" {
		f.attacker.UpdateDataMonster().CurrentEnemy = f.spell.target
		f.attacker.MonsterPushAction(ai.ACTION_MELEE_ATTACK)
	} else {
		pos := noxClient.Viewport().ToScreenPos(image.Pt(int(f.spell.target.PosVec.X), int(f.spell.target.PosVec.Y)))
		e2eQueueInput(&seat.MouseMoveEvent{Pos: pos}, &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true})
		f.held = true
	}
	e2eLog.Printf("SHOCK RETALIATION ATTACK: mode=%s attacker=%p enchanted=%p weapon=%p HP=%d expected-electric=%d", f.mode, f.attacker, f.spell.target, f.weapon, f.health, f.damage)
}

func (f *e2eShockRetaliationFixture) release() {
	if f.held {
		e2eQueueInput(&seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false})
		f.held = false
	}
}

func (f *e2eShockRetaliationFixture) observe() {
	if !f.active || f.hit || f.shockedAudio == 0 || noxServer.Frame() < f.struckFrame {
		return
	}
	attacker, victim := f.attacker, f.spell.target
	if attacker.Frame134 != f.struckFrame || victim.Frame134 != f.struckFrame {
		return
	}
	var marker, rawType uint32
	if attacker.Class().Has(object.ClassPlayer) {
		ud := attacker.UpdateDataPlayer()
		marker, rawType = ud.Field76, ud.Field75
		if ud.State != server.PlayerState23 || ud.Field40_0 != 2 {
			e2eError(fmt.Errorf("Shock player hit reaction/electric state=%d/%d HP=%d marker=%d/%d", ud.State, ud.Field40_0, attacker.HealthData.Cur, marker, rawType))
			return
		}
	} else {
		ud := attacker.UpdateDataMonster()
		marker, rawType = ud.Field547, ud.Field546
		// The attacker is still inside its own melee action when retaliation
		// lands. The outer AI update consumes/clears MonStatusInjured before
		// RunTickHooks, unlike the immediate C->damage callback unit test.
		if ud.Field523_2 != 2 {
			e2eError(fmt.Errorf("Shock NPC electric state=%d HP=%d marker=%d/%d", ud.Field523_2, attacker.HealthData.Cur, marker, rawType))
			return
		}
	}
	if attacker.HealthData.Cur != f.health-f.damage || attacker.Obj130 != victim || attacker.Field131 != uint32(object.DamageElectric) ||
		marker != 2 || rawType != uint32(object.DamageElectric) || victim.HealthData.Cur >= f.spell.health ||
		victim.Obj130 != f.weapon || victim.Field131 != uint32(object.DamageBlade) ||
		victim.HasEnchant(server.ENCHANT_SHOCK) || victim.EnchantDur(server.ENCHANT_SHOCK) != 0 || f.spell.offAudio != 1 {
		e2eError(fmt.Errorf("Shock real retaliation mismatch mode=%s HP=%d->%d expected=%d source=%p marker=%d/%d incoming=%d->%d/%p on/off=%d/%d", f.mode, f.health, attacker.HealthData.Cur, f.damage, attacker.Obj130, marker, rawType, f.spell.health, victim.HealthData.Cur, victim.Obj130, f.spell.onAudio, f.spell.offAudio))
		return
	}
	f.hit, f.hitHealth, f.incomingHealth = true, attacker.HealthData.Cur, victim.HealthData.Cur
	e2eLog.Printf("SHOCK RETALIATION HIT: mode=%s HP=%d->%d electric=%d original-melee=%d->%d frame=%d nil-weapon-source=%p", f.mode, f.health, f.hitHealth, f.damage, f.spell.health, f.incomingHealth, f.struckFrame, attacker.Obj130)
}

func (f *e2eShockRetaliationFixture) complete() bool {
	f.observe()
	if !f.hit {
		if noxServer.Frame()-f.frame > 120 {
			e2eError(fmt.Errorf("Shock retaliation did not damage both units: mode=%s on/off/shocked=%d/%d/%d HP=%d/%d", f.mode, f.spell.onAudio, f.spell.offAudio, f.shockedAudio, f.attacker.HealthData.Cur, f.spell.target.HealthData.Cur))
			return true
		}
		return false
	}
	host, npc := f.spell.host, f.spell.npc
	for _, unit := range []*server.Object{host, npc} {
		dr := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(unit)))
		if dr == nil || dr.Buffs != unit.Buffs || dr.HasEnchant(server.ENCHANT_SHOCK) {
			return false
		}
	}
	meter, ready := e2eClientHUDMeter(0)
	if !ready || meter.Current != uint32(host.HealthData.Cur) || memmap.Uint32(0x5D4594, 1062540) != host.Buffs {
		return false
	}
	damage := f.damage
	if f.mode == "to-player" {
		damage = f.spell.health - f.incomingHealth
	}
	dr := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(npc)))
	if delta, ok := legacy.HealthChangeForDrawable(dr.NetCode32); !ok || delta != -int16(damage) {
		return false
	}
	if f.shockedAudio != 1 || f.spell.onAudio != 1 || f.spell.offAudio != 1 || f.attacker.HealthData.Cur != f.hitHealth || f.spell.target.HealthData.Cur != f.incomingHealth {
		e2eError(fmt.Errorf("Shock retaliation changed while replaying real HP/buff packets"))
		return true
	}
	e2eLog.Printf("SHOCK RETALIATION COMPLETE: mode=%s actual-HP=verified client-HUD=verified NPC-delta=%d Shock-consumed=verified audio-on/off/shocked=1/1/1", f.mode, -int16(damage))
	return true
}

func (f *e2eShockRetaliationFixture) cleanup() {
	f.active, f.spell.active = false, false
	f.release()
	noxServer.DelayedDelete(f.spell.npc)
	for _, item := range f.armor {
		if !asObjectS(f.spell.host).Equip(item) {
			e2eError(fmt.Errorf("Shock fixture could not restore host equipment"))
			return
		}
	}
	asObjectS(f.spell.host).SetPos(f.spell.original)
	f.spell.host.VelVec, f.spell.host.ForceVec, f.spell.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
}

// Normal stock creation/equipment, pose, waiting/melee AI and real mouse input
// are fixture inputs. HP, damage, buffs, hit reaction, timers, audio, packets,
// drawables and framebuffer pixels are only observed, never supplied here.
func (sc *e2eScenario) CheckShockRetaliation(mode, name string) {
	if !e2eShockRetaliationMode(mode) {
		e2eError(fmt.Errorf("invalid Shock retaliation mode %q", mode))
		return
	}
	f := &e2eShockRetaliationFixture{mode: mode}
	sc.addWhen(0, name+" prepare", 1200, func() bool {
		host := noxServer.Players.HostUnit()
		return host != nil && host.Buffs == 0 && noxClient.ClientPlayerUnit() != nil && nox_client_isConnected()
	}, f.prepare)
	sc.Wait(12, name+" publish placement")
	sc.add(0, name+" cast", f.spell.begin)
	sc.addWhen(0, name+" visible", 180, func() bool { return f.spell.visible("first", 12) }, func() {})
	sc.Screen(name + " enchanted")
	sc.add(0, name+" actual attack", f.attack)
	sc.add(1, name+" release input", f.release)
	sc.addWhen(1, name+" complete", 180, f.complete, func() {})
	sc.Screen(name + " consumed and damaged")
	sc.add(0, name+" cleanup", f.cleanup)
}
