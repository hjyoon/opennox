package opennox

import (
	"fmt"
	"math"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func e2eRestoreSpellMode(level int, mode string) (spell.ID, bool, bool) {
	if level < 1 || level > 5 {
		return 0, false, false
	}
	switch mode {
	case "health-player":
		return spell.SPELL_RESTORE_HEALTH, false, true
	case "health-npc":
		return spell.SPELL_RESTORE_HEALTH, true, true
	case "wink-player":
		return spell.SPELL_WINK, false, true
	case "wink-npc":
		return spell.SPELL_WINK, true, true
	case "mana-player":
		return spell.SPELL_RESTORE_MANA, false, true
	case "mana-npc":
		return spell.SPELL_RESTORE_MANA, true, true
	default:
		return 0, false, false
	}
}

type e2eRestoreSpellFixture struct {
	id             spell.ID
	mode           string
	level          int
	npc            bool
	host, target   *server.Object
	original       types.Pointf
	health, mana   uint16
	maxHP, maxMana uint16
	active         bool
	audio          int
	frame          uint32
}

func (f *e2eRestoreSpellFixture) prepare() {
	f.host = noxServer.Players.HostUnit()
	if f.host == nil || f.host.ControllingPlayer() == nil || f.host.UpdateData == nil ||
		f.host.HealthData == nil || f.host.HealthData.Cur <= 7 || f.host.Buffs != 0 || f.host.Poison540 != 0 ||
		f.host.Flags().HasAny(object.FlagDead|object.FlagDestroyed|object.FlagNoUpdate) {
		e2eError(fmt.Errorf("restore fixture requires a live unenchanted host"))
		return
	}
	f.original, f.target = f.host.PosVec, f.host
	if f.npc {
		origin, direction, err := e2eWarriorAbilityArena(f.original, f.host.Shape.Circle.R+24,
			func(from, to types.Pointf) bool { return e2eWarriorLaneClear(f.host, from, to) })
		if err != nil {
			e2eError(err)
			return
		}
		asObjectS(f.host).SetPos(origin)
		f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
		f.target = noxServer.NewObjectByTypeID("NPC")
		if f.target == nil {
			e2eError(fmt.Errorf("restore fixture requires stock NPC"))
			return
		}
		noxServer.CreateObjectAt(f.target, f.host, origin.Add(direction.Mul(64)))
		noxServer.ObjectsAddPending()
		if f.target.UpdateData == nil || f.target.HealthData == nil || f.target.HealthData.Cur <= 7 ||
			!f.target.Class().Has(object.ClassMonster) || !f.target.SubClass().AsMonster().Has(object.MonsterNPC) {
			e2eError(fmt.Errorf("restore NPC was not normally initialized"))
			return
		}
		// Ordinary placement, ownership and waiting AI are inputs only.
		f.target.UpdateDataMonster().SetAggression(0)
		f.target.ClearActionStack()
		f.target.MonsterPushAction(ai.ACTION_WAIT, noxServer.Frame()+250000)
	}
	f.maxHP, f.maxMana = f.target.HealthData.Max, f.host.UpdateDataPlayer().ManaMax
	if f.id == spell.SPELL_RESTORE_MANA && !f.npc {
		before := f.host.UpdateDataPlayer().ManaCur
		if before <= 7 {
			e2eError(fmt.Errorf("restore mana requires stock player mana above 7"))
			return
		}
		legacy.Nox_xxx_playerManaSub_4EEBF0(f.host, 7)
		if f.host.UpdateDataPlayer().ManaCur != before-7 {
			e2eError(fmt.Errorf("ordinary mana consumption did not reach the expected input"))
			return
		}
	} else {
		before := f.target.HealthData.Cur
		if !asObjectS(f.target).DoDamage(nil, 7, object.DamageTrue) || f.target.HealthData.Cur != before-7 {
			e2eError(fmt.Errorf("ordinary damage did not reach the expected injury input"))
			return
		}
	}
	// Never write a recovery result, resource maximum, frame/RNG, packet,
	// audio event or client drawable. Injury is published before dispatch.
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(f.host), unsafe.Pointer(f.target), f.target.UpdateData, unsafe.Pointer(f.target.HealthData)} {
			if uintptr(ptr) <= math.MaxUint32 {
				e2eError(fmt.Errorf("restore fixture pointer below 4 GiB: %p", ptr))
				return
			}
		}
	}
	noxServer.Audio.OnSound(f.observeSound)
}

func (f *e2eRestoreSpellFixture) observeSound(id sound.ID, kind int, target *server.Object, _ types.Pointf) {
	if !f.active || id != sound.SoundRestoreHealth && id != sound.SoundRestoreMana {
		return
	}
	want := sound.SoundRestoreHealth
	if f.id == spell.SPELL_RESTORE_MANA {
		want = sound.SoundRestoreMana
	}
	if f.npc && f.id == spell.SPELL_RESTORE_MANA || id != want || kind != 0 || target != f.target ||
		target.HealthData == nil || f.id != spell.SPELL_RESTORE_MANA && target.HealthData.Cur != f.maxHP ||
		f.id == spell.SPELL_RESTORE_MANA && target.UpdateDataPlayer().ManaCur != f.maxMana {
		e2eError(fmt.Errorf("restore audio target/resource/order mismatch: mode=%s sound=%d target=%p", f.mode, id, target))
		return
	}
	f.audio++
}

func (f *e2eRestoreSpellFixture) cast() {
	if !e2eObjectInWorld(f.target) || noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.target))) == nil {
		e2eError(fmt.Errorf("restore target was not published"))
		return
	}
	arg, free := e2ePlayerStatusSpellArg(f.target)
	defer free()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(arg)) <= math.MaxUint32 {
		e2eError(fmt.Errorf("restore acceptance record below 4 GiB: %p", arg))
		return
	}
	before := *arg
	f.health, f.mana = f.target.HealthData.Cur, f.host.UpdateDataPlayer().ManaCur
	hostHP := f.host.HealthData.Cur
	f.frame, f.active = noxServer.Frame(), true
	got := noxServer.SpellAccept4FD400(f.id, f.host, f.host, f.host, arg, int32(f.level))
	wantHP, wantMana := f.health, f.mana
	if f.id != spell.SPELL_RESTORE_MANA {
		wantHP = f.maxHP
	} else if !f.npc {
		wantMana = f.maxMana
	}
	if got != 1 || *arg != before || f.target.HealthData.Cur != wantHP ||
		f.host.UpdateDataPlayer().ManaCur != wantMana || f.target.HealthData.Max != f.maxHP ||
		f.host.UpdateDataPlayer().ManaMax != f.maxMana || f.npc && f.host.HealthData.Cur != hostHP {
		e2eError(fmt.Errorf("restore dispatch mismatch: mode=%s level=%d result=%d HP=%d->%d/%d mana=%d->%d/%d arg-changed=%t",
			f.mode, f.level, got, f.health, f.target.HealthData.Cur, f.maxHP, f.mana, f.host.UpdateDataPlayer().ManaCur, f.maxMana, *arg != before))
		return
	}
	e2eLog.Printf("RESTORE CAST: mode=%s level=%d caster=%p target=%p arg=%p result=1 HP=%d->%d/%d mana=%d->%d/%d arg-unchanged=true",
		f.mode, f.level, f.host, f.target, arg, f.health, f.target.HealthData.Cur, f.maxHP, f.mana, f.host.UpdateDataPlayer().ManaCur, f.maxMana)
}

func (f *e2eRestoreSpellFixture) complete() bool {
	if !f.active || noxServer.Frame()-f.frame < 2 {
		return false
	}
	wantAudio := 1
	if f.npc && f.id == spell.SPELL_RESTORE_MANA {
		wantAudio = 0
	}
	if f.audio != wantAudio {
		return false
	}
	if !f.npc {
		index, maximum := 0, f.maxHP
		if f.id == spell.SPELL_RESTORE_MANA {
			index, maximum = 1, f.maxMana
		}
		meter, ready := e2eClientHUDMeter(index)
		if !ready || meter.Current != uint32(maximum) || meter.Maximum != uint32(maximum) {
			return false
		}
	} else if f.id != spell.SPELL_RESTORE_MANA {
		drawable := noxClient.Objs.ByNetCode(uint16(noxServer.GetUnitNetCode(f.target)))
		if drawable == nil {
			return false
		}
		delta, ready := legacy.HealthChangeForDrawable(drawable.NetCode32)
		if !ready || delta <= 0 {
			return false
		}
	}
	e2eLog.Printf("RESTORE COMPLETE: mode=%s level=%d audio=%d target=%p elapsed=%d client-resource-observed=%t",
		f.mode, f.level, f.audio, f.target, noxServer.Frame()-f.frame, !(f.npc && f.id == spell.SPELL_RESTORE_MANA))
	return true
}

func (f *e2eRestoreSpellFixture) cleanup() {
	f.active = false
	if f.npc {
		noxServer.DelayedDelete(f.target)
		asObjectS(f.host).SetPos(f.original)
		f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	}
}

// This is real server instant dispatch after ordinary injury/mana consumption,
// with live resource/audio/client observations. It does not certify player
// incantation, mana cost, spell-projectile travel or callback-time mutation.
func (sc *e2eScenario) CheckRestoreSpell(level int, mode, name string) {
	id, npc, ok := e2eRestoreSpellMode(level, mode)
	if !ok {
		e2eError(fmt.Errorf("invalid restore spell mode/level %s/%d", mode, level))
		return
	}
	f := &e2eRestoreSpellFixture{id: id, npc: npc, mode: mode, level: level}
	sc.addWhen(0, name+" prepare", 1200, func() bool {
		host := noxServer.Players.HostUnit()
		// The normal host spawn protection must expire naturally before
		// injury, just as in the other unenchanted-host spell observers.
		return host != nil && host.ControllingPlayer() != nil && host.UpdateData != nil &&
			host.HealthData != nil && host.HealthData.Cur > 7 && host.Buffs == 0 && host.Poison540 == 0 &&
			!host.Flags().HasAny(object.FlagDead|object.FlagDestroyed|object.FlagNoUpdate) &&
			noxClient.ClientPlayerUnit() != nil && nox_client_isConnected()
	}, f.prepare)
	sc.Wait(12, name+" publish injury")
	sc.add(0, name+" instant dispatch", f.cast)
	sc.addWhen(1, name+" actual audio and client resource", 180, f.complete, func() {})
	sc.CaptureMagicFrame(name + " actual recovery frame")
	sc.add(0, name+" cleanup", f.cleanup)
	sc.Wait(3, name+" cleanup complete")
}
