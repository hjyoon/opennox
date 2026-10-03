package opennox

import (
	"fmt"
	"image"
	"math"
	"strings"

	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func e2eGreatSwordMissileMode(level int, mode string) (fromNPC, front, ok bool) {
	direction, facing, found := strings.Cut(mode, "/")
	fromNPC, ok = e2eFireballUnitMode(level, direction)
	front = facing == "front"
	return fromNPC, front, found && ok && (front || facing == "rear")
}

type e2eGreatSwordMissileFixture struct {
	e2eFireballUnitFixture
	front, stateObserved, diagnosed bool
	sword                           *server.Object
	swordHP                         uint16
	swordCarry                      float32
	percent                         float64
	blocks                          int
	hud                             e2eHUDMeterState
}

func (f *e2eGreatSwordMissileFixture) prepare() {
	f.e2eFireballUnitFixture.prepare()
	if f.npc == nil || f.target == nil {
		return
	}
	if f.level == 1 && f.front {
		e2eLog.Printf("GREATSWORD EQUIPMENT: direction=%s state=%d weapon=%#x armor=%#x",
			f.direction, f.host.UpdateDataPlayer().State, f.host.ControllingPlayer().WeaponEquip, f.host.ControllingPlayer().ArmorEquip)
		for item := f.host.FirstItem(); item != nil; item = item.NextItem() {
			if item.Class().HasAny(object.ClassWeapon | object.ClassArmor) {
				e2eLog.Printf("GREATSWORD INVENTORY: type=%s netcode=%d flags=%#x", item.ObjectTypeC().ID(), item.NetCode, uint32(item.Flags()))
			}
		}
	}
	if f.fromNPC {
		// Host equipment must have arrived through the public inventory input.
		f.sword = f.host.UpdateDataPlayer().EquippedWeapon
		if f.host.ControllingPlayer().ArmorEquip&0x3000000 != 0 {
			e2eError(fmt.Errorf("GreatSword baseline still has the ordinary inventory shield equipped"))
			return
		}
	} else {
		f.sword = noxServer.NewObjectByTypeID("GreatSword")
		if f.sword == nil {
			e2eError(fmt.Errorf("GreatSword stock item missing"))
			return
		}
		noxServer.CreateObjectAt(f.sword, nil, f.npc.PosVec)
		noxServer.ObjectsAddPending()
		// Ordinary pickup/equip calls are fixture setup, not raw equip flags.
		if !asObjectS(f.npc).Equip(f.sword) {
			e2eError(fmt.Errorf("stock NPC could not equip GreatSword"))
			return
		}
	}
	if !f.held() || f.sword.HealthData == nil || f.sword.UpdateData == nil || f.sword.InitData == nil {
		e2eError(fmt.Errorf("GreatSword not held through actual equip path"))
		return
	}
	for _, modifier := range f.sword.InitDataModifier().Modifiers {
		if modifier != nil && (modifier.Defend76.Fnc != nil || modifier.Engage112 != nil) {
			e2eError(fmt.Errorf("GreatSword baseline has a protection modifier"))
			return
		}
	}
	noxServer.Audio.OnSound(f.observeAudio)
	noxServer.TickHook(f.observeState)
}

func (f *e2eGreatSwordMissileFixture) held() bool {
	if f.sword == nil || f.sword.ObjectTypeC().ID() != "GreatSword" ||
		f.sword.InvHolder != f.target || !f.sword.Flags().Has(object.FlagEquipped) {
		return false
	}
	if f.fromNPC {
		return f.target.UpdateDataPlayer().EquippedWeapon == f.sword &&
			f.target.ControllingPlayer().WeaponEquip&0x400 != 0
	}
	return f.target.NPCEquippedWeapon538960() == f.sword &&
		f.target.UpdateDataMonster().WeaponEquipFlags&0x400 != 0
}

func (f *e2eGreatSwordMissileFixture) aim() {
	vector := f.caster.PosVec.Sub(f.target.PosVec)
	if !f.front {
		vector = vector.Mul(-1)
	}
	f.target.SetDir(server.DirFromVec(vector))
	if f.fromNPC {
		aim := f.target.PosVec.Add(vector.Mul(0.6))
		mouse := noxClient.Viewport().ToScreenPos(image.Pt(int(aim.X), int(aim.Y)))
		e2eQueueInput(&seat.MouseMoveEvent{Pos: mouse, Relative: false})
	}
}

func (f *e2eGreatSwordMissileFixture) beginCast() {
	if !f.held() || f.sword.Flags().Has(object.FlagDestroyed) {
		e2eError(fmt.Errorf("GreatSword lost before casting"))
		return
	}
	if f.fromNPC && f.target.UpdateDataPlayer().State != server.PlayerState13 {
		e2eError(fmt.Errorf("GreatSword target did not naturally enter idle: %d", f.target.UpdateDataPlayer().State))
		return
	}
	facing := legacy.Nox_server_testTwoPointsAndDirection_4E6E50(
		f.target.PosVec, int16(f.target.Direction1), f.caster.PosVec,
	)&1 != 0
	if facing != f.front {
		e2eError(fmt.Errorf("GreatSword ordinary aim did not settle: front=%t direction=%d", f.front, f.target.Direction1))
		return
	}
	f.swordHP = f.sword.HealthData.Cur
	f.swordCarry = math.Float32frombits(f.sword.UpdateDataWeaponArmor().Field0)
	f.percent = noxServer.Balance.Float("ItemDamageFromBlockPercentage")
	if f.fromNPC {
		var ready bool
		f.hud, ready = e2eClientHUDMeter(0)
		if !ready {
			e2eError(fmt.Errorf("GreatSword baseline HUD unavailable"))
			return
		}
	}
	f.e2eFireballUnitFixture.beginCast()
}

func (f *e2eGreatSwordMissileFixture) observeAudio(id sound.ID, kind int, owner *server.Object, _ types.Pointf) {
	if f.active && id == sound.SoundFireballExplode && owner == f.projectile {
		// Detonation follows the complete block tail in the same collision.
		// Observe the live action head before ordinary AI can finish the block.
		f.observeState()
	}
	if !f.active || id != sound.SoundGreatSwordReflect || owner != f.target {
		return
	}
	f.blocks++
	projectile := f.projectile
	if !f.front || kind != 0 || f.blocks != 1 || projectile == nil || projectile.ObjOwner != f.target ||
		f.target.HealthData.Cur != f.health {
		e2eError(fmt.Errorf("GreatSword block prefix: front=%t kind/count=%d/%d projectile=%p HP=%d/%d",
			f.front, kind, f.blocks, projectile, f.target.HealthData.Cur, f.health))
		return
	}
	towardCaster := f.caster.PosVec.Sub(f.target.PosVec)
	if projectile.VelVec.X*towardCaster.X+projectile.VelVec.Y*towardCaster.Y <= 0 {
		e2eError(fmt.Errorf("GreatSword did not reverse the real incoming velocity"))
		return
	}
	e2eLog.Printf("GREATSWORD REFLECT: direction=%s level=%d owner=defender velocity=%v HP=%d",
		f.direction, f.actualLevel, projectile.VelVec, f.health)
}

func (f *e2eGreatSwordMissileFixture) observeState() {
	if !f.active || f.blocks == 0 {
		return
	}
	if f.fromNPC {
		state := f.target.UpdateDataPlayer().State
		f.stateObserved = f.stateObserved || state == server.PlayerState18 || state == server.PlayerState19 || state == server.PlayerState20
	} else {
		f.stateObserved = f.stateObserved || e2eGreatSwordNPCBlockState(f.target.UpdateDataMonster())
	}
}

func e2eGreatSwordNPCBlockState(ud *server.MonsterUpdateData) bool {
	if ud == nil {
		return false
	}
	// HasScheduledAction deliberately excludes the executing stack head.
	// A queued block below another action is not the observed block state.
	head := ud.AIStackHead()
	return head != nil && head.Type() == ai.ACTION_WEAPON_BLOCK
}

func (f *e2eGreatSwordMissileFixture) complete() bool {
	if f.active && !f.diagnosed && noxServer.Frame()-f.frame >= 120 {
		f.diagnosed = true
		e2eLog.Printf("GREATSWORD WAIT: front=%t block/state/cast/explode=%d/%t/%d/%d HP=%d/%d sword=%d/%d in-world=%t client-missile=%t",
			f.front, f.blocks, f.stateObserved, f.castAudio, f.explodeAudio, f.target.HealthData.Cur, f.health,
			f.sword.HealthData.Cur, f.swordHP, e2eFistInWorld(f.projectile, f.wire, f.script),
			noxClient.Objs.ByNetCode(uint16(f.wire)) != nil)
	}
	if !f.front {
		if !f.e2eFireballUnitFixture.complete() {
			return false
		}
		if f.blocks != 0 || f.sword.HealthData.Cur != f.swordHP ||
			math.Float32frombits(f.sword.UpdateDataWeaponArmor().Field0) != f.swordCarry {
			e2eError(fmt.Errorf("rear Fireball incorrectly blocked or wore GreatSword"))
			return false
		}
	} else {
		if f.blocks != 1 || !f.stateObserved || f.castAudio != 1 || f.explodeAudio != 1 ||
			f.level == 0 && (!f.natural || f.npc.UpdateDataMonster().HasAction(ai.ACTION_CAST_SPELL_ON_OBJECT)) {
			return false
		}
		// Fireball's stock collision ignores the PlayerDamage return value:
		// a GreatSword block protects direct HP, but still detonates/deletes.
		if e2eFistInWorld(f.projectile, f.wire, f.script) || noxClient.Objs.ByNetCode(uint16(f.wire)) != nil {
			return false
		}
		amount := float32(f.percent*float64(f.damage)) + f.swordCarry
		wear := int32(math.RoundToEven(float64(amount)))
		if wear <= 0 || wear >= int32(f.swordHP) || f.sword.HealthData.Cur != f.swordHP-uint16(wear) ||
			f.sword.UpdateDataWeaponArmor().Field0 != math.Float32bits(amount-float32(wear)) ||
			f.target.HealthData.Cur != f.health || !f.held() {
			e2eError(fmt.Errorf("GreatSword real HP/wear mismatch: target=%d/%d sword=%d/%d wear=%d",
				f.target.HealthData.Cur, f.health, f.swordHP, f.sword.HealthData.Cur, wear))
			return false
		}
		var marker, markerType, carry uint32
		if f.fromNPC {
			ud := f.target.UpdateDataPlayer()
			marker, markerType, carry = ud.Field76, ud.Field75, ud.Field21
			meter, ready := e2eClientHUDMeter(0)
			if !ready || meter.Current != f.hud.Current || meter.Maximum != f.hud.Maximum {
				return false
			}
		} else {
			ud := f.target.UpdateDataMonster()
			marker, markerType, carry = ud.Field547, ud.Field546, ud.Field1
		}
		if marker != 1 || markerType != uint32(f.projectileType) || carry != f.carry {
			e2eError(fmt.Errorf("GreatSword live missile markers/carry mismatch"))
			return false
		}
	}
	e2eLog.Printf("GREATSWORD COMPLETE: direction=%s level=%d front=%t blocks=%d HP=%d->%d sword=%d->%d",
		f.direction, f.actualLevel, f.front, f.blocks, f.health, f.target.HealthData.Cur, f.swordHP, f.sword.HealthData.Cur)
	return true
}

// Only placement, stock equipment, waiting AI and script cast requests are
// setup. Damage, block states/actions, owner/velocity, wear and client results
// come from the real game path; no observed result is supplied by this fixture.
func (sc *e2eScenario) CheckGreatSwordMissileDefense(level int, mode, name string) {
	fromNPC, front, ok := e2eGreatSwordMissileMode(level, mode)
	if !ok {
		e2eError(fmt.Errorf("invalid GreatSword missile fixture: %s/%d", mode, level))
		return
	}
	direction, _, _ := strings.Cut(mode, "/")
	f := &e2eGreatSwordMissileFixture{e2eFireballUnitFixture: e2eFireballUnitFixture{
		level: level, direction: direction, fromNPC: fromNPC,
	}, front: front}
	sc.addWhen(0, name+" prepare", 1200, func() bool {
		host := noxServer.Players.HostUnit()
		return host != nil && host.Buffs == 0 && noxClient.ClientPlayerUnit() != nil && nox_client_isConnected()
	}, f.prepare)
	sc.Wait(12, name+" publish units")
	sc.add(0, name+" ordinary aim", f.aim)
	sc.Wait(12, name+" settle aim")
	sc.add(0, name+" actual cast", f.beginCast)
	sc.addWhen(1, name+" actual block or hit", 300, f.complete, f.cleanup)
	sc.Wait(3, name+" cleanup")
}
