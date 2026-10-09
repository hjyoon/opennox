package opennox

import (
	"fmt"
	"image"
	"math"
	"unsafe"

	"github.com/opennox/libs/client/keybind"
	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func e2ePlayerPlasmaRecord(unit *server.Object) *server.DurSpell {
	for r := noxServer.Spells.Dur.List; r != nil; r = r.Next {
		if r.Spell == uint32(spell.SPELL_PLASMA) && r.Caster16 == unit {
			return r
		}
	}
	return nil
}

func e2ePlayerPlasmaRay(unit, target *server.Object) bool {
	source, dest := uint16(noxServer.GetUnitNetCode(unit)), uint16(noxServer.GetUnitNetCode(target))
	for _, ray := range noxClient.fxDurationRays {
		if ray.kind == 1 && ray.source == source && ray.target == dest && ray.drawable != nil {
			return true
		}
	}
	return false
}

// The initial inventory may already be open. Observe its settled state before
// sending the real toggle key; never overwrite animation or window state.
func (sc *e2eScenario) playerPlasmaInventory(open bool, name string) {
	wantState, wantOffset := 0, -225
	if open {
		wantState, wantOffset = 2, 0
	}
	sc.addWhen(0, name+" settled state", 120, func() bool {
		state := legacy.Nox_client_inventoryAnimationState()
		return legacy.InventoryWindow() != nil && (state == 0 || state == 2)
	}, func() {
		state := legacy.Nox_client_inventoryAnimationState()
		press := state != wantState
		e2eLog.Printf("PLAYER PLASMA INVENTORY INPUT: open=%t state=%d offset=%d toggle=%t", open, state, legacy.Nox_client_inventoryAnimationOffset(), press)
		if press {
			e2eQueueInput(&seat.KeyboardEvent{Key: keybind.KeyI, Pressed: true})
		}
	})
	sc.Input(1, name+" release inventory key", &seat.KeyboardEvent{Key: keybind.KeyI, Pressed: false})
	sc.addWhen(1, name+" observed visibility", 120, func() bool {
		return legacy.Nox_client_inventoryAnimationState() == wantState && legacy.Nox_client_inventoryAnimationOffset() == wantOffset
	}, nil)
}

// CheckPlayerPlasma uses inventory-equipped stock OblivionOrb and actual
// mouse input. Only the starting arena/target is arranged; no cast, damage,
// duration, charge, ray packet or cancellation result is injected.
func (sc *e2eScenario) CheckPlayerPlasma(name string) {
	sc.playerPlasmaInventory(true, name+" open inventory")
	sc.ClickInventoryItem("OblivionOrb", name+" equip stock wand through inventory input")
	sc.addWhen(1, name+" receive equipped weapon report", 120, func() bool {
		unit := noxServer.Players.HostUnit()
		if unit == nil {
			return false
		}
		weapon := unit.UpdateDataPlayer().EquippedWeapon
		return weapon != nil && weapon.ObjectTypeC().ID() == "OblivionOrb" && weapon.Flags().Has(object.FlagEquipped)
	}, nil)
	sc.playerPlasmaInventory(false, name+" close inventory")
	var unit, weapon, target *server.Object
	var original types.Pointf
	var data *server.WandUseData
	var record *server.DurSpell
	var beforeCharge uint8
	var beforeHP, stoppedHP uint16
	var started, stopped uint32
	sc.addWhen(0, name+" stock equipment", 1200, func() bool {
		return noxServer.Players.HostUnit() != nil && noxClient.ClientPlayerUnit() != nil && nox_client_isConnected()
	}, func() {
		unit = noxServer.Players.HostUnit()
		weapon = unit.UpdateDataPlayer().EquippedWeapon
		if weapon == nil || weapon.ObjectTypeC().ID() != "OblivionOrb" || !weapon.Flags().Has(object.FlagEquipped) {
			e2eError(fmt.Errorf("stock OblivionOrb not equipped through inventory"))
			return
		}
		data = weapon.UseData.AsWand()
		if data == nil || data.Spell != uint32(spell.SPELL_PLASMA) || data.Charge != 250 || data.MaxCharge != 250 {
			e2eError(fmt.Errorf("unexpected stock Plasma wand data: %+v", data))
			return
		}
		for _, obj := range []*server.Object{unit, weapon} {
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(obj)) <= math.MaxUint32 {
				e2eError(fmt.Errorf("Plasma fixture pointer is not above PE32: %p", obj))
				return
			}
		}
		original = unit.PosVec
		origin, direction, err := e2eWarriorAbilityArena(original, unit.Shape.Circle.R+4, func(from, to types.Pointf) bool { return e2eWarriorLaneClear(unit, from, to) })
		if err != nil {
			e2eError(err)
			return
		}
		asObjectS(unit).SetPos(origin)
		unit.VelVec, unit.ForceVec, unit.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
		target = noxServer.NewObjectByTypeID("Troll")
		if target == nil {
			e2eError(fmt.Errorf("missing stock Troll"))
			return
		}
		noxServer.CreateObjectAt(target, nil, origin.Add(direction.Mul(112)))
		noxServer.ObjectsAddPending()
		asObjectS(target).SetMaxHealth(2000)
		target.UpdateDataMonster().SetAggression(0)
		target.ClearActionStack()
		target.MonsterPushAction(ai.ACTION_WAIT, noxServer.Frame()+250000)
		if !noxServer.IsEnemyTo(unit, target) {
			e2eError(fmt.Errorf("stock target is not hostile"))
			return
		}
		e2eLog.Printf("PLAYER PLASMA PREPARED: player=%p weapon=%p target=%p charges=%d/%d spell=%d", unit, weapon, target, data.Charge, data.MaxCharge, data.Spell)
	})
	sc.Wait(12, name+" publish placement")
	for phase := 0; phase < 2; phase++ {
		exhaust := phase == 1
		label := fmt.Sprintf("%s phase %d", name, phase+1)
		sc.add(0, label+" actual held attack", func() {
			beforeCharge, beforeHP, started = data.Charge, target.HealthData.Cur, noxServer.Frame()
			record = nil
			mouse := noxClient.Viewport().ToScreenPos(image.Pt(int(target.PosVec.X), int(target.PosVec.Y)))
			e2eQueueInput(&seat.MouseMoveEvent{Pos: mouse}, &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true})
		})
		sc.addWhen(1, label+" real duration damage charge and client ray", 240, func() bool {
			record = e2ePlayerPlasmaRecord(unit)
			return record != nil && record.Target48 == target && data.Charge <= beforeCharge-8 && target.HealthData.Cur < beforeHP && e2ePlayerPlasmaRay(unit, target)
		}, func() {
			if noxServer.spells.duration.plasmaWeapons[record] != weapon || noxServer.spells.duration.durationRayTargets[record] != target || record.Flags88&2 == 0 || data.Progress != 100*uint32(data.Charge)/uint32(data.MaxCharge) || target.Obj130 != unit {
				e2eError(fmt.Errorf("Plasma native weapon/ray/damage attribution mismatch"))
				return
			}
			e2eLog.Printf("PLAYER PLASMA ACTIVE: phase=%d record=%p caster=%p target=%p charges=%d->%d HP=%d->%d elapsed=%d ray=true state=%d", phase+1, record, unit, target, beforeCharge, data.Charge, beforeHP, target.HealthData.Cur, noxServer.Frame()-started, unit.UpdateDataPlayer().State)
		})
		sc.Screen(label + " natural Plasma ray")
		if exhaust {
			sc.addWhen(1, label+" natural stock charge exhaustion", 900, func() bool { return data.Charge == 0 }, func() {
				if target.HealthData.Cur >= beforeHP || data.Progress != 0 {
					e2eError(fmt.Errorf("exhausted Plasma did not damage target or update progress"))
					return
				}
				e2eLog.Printf("PLAYER PLASMA EXHAUSTED: charges=%d/%d HP=%d->%d elapsed=%d", data.Charge, data.MaxCharge, beforeHP, target.HealthData.Cur, noxServer.Frame()-started)
			})
		}
		sc.Input(0, label+" actual attack release", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false})
		sc.addWhen(1, label+" natural duration and ray cleanup", 120, func() bool {
			return e2ePlayerPlasmaRecord(unit) == nil && !e2ePlayerPlasmaRay(unit, target) && data.Flags&4 == 0 && noxServer.spells.duration.plasmaWeapons[record] == nil && noxServer.spells.duration.durationRayTargets[record] == nil
		}, func() {
			stoppedHP, stopped = target.HealthData.Cur, noxServer.Frame()
			e2eLog.Printf("PLAYER PLASMA STOPPED: phase=%d record=%p frame=%d flags=%#x charges=%d ray=false", phase+1, record, stopped, data.Flags, data.Charge)
		})
		sc.addWhen(1, label+" no damage after natural stop", 60, func() bool {
			if target.HealthData.Cur < stoppedHP || e2ePlayerPlasmaRecord(unit) != nil || e2ePlayerPlasmaRay(unit, target) {
				e2eError(fmt.Errorf("Plasma resumed after release/exhaustion"))
				return false
			}
			return noxServer.Frame()-stopped >= 12
		}, func() {})
		sc.Wait(40, label+" settle attack cooldown")
	}
	sc.Screen(name + " natural completion")
	sc.add(0, name+" cleanup", func() { noxServer.DelayedDelete(target); asObjectS(unit).SetPos(original) })
}
