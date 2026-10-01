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

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

func e2ePlayerThrownWeaponType(item string) (string, object.WeaponClass, bool) {
	switch item {
	case "FanChakram":
		return "FanChakramInMotion", object.WeaponShuriken, true
	case "RoundChakram":
		return "RoundChakramInMotion", object.WeaponChakram, true
	default:
		return "", 0, false
	}
}

type e2ePlayerThrownWeaponFixture struct {
	item, projectileType string
	flag                 object.WeaponClass
	unit, weapon, target *server.Object
	original             types.Pointf
	frame, attackFrame   uint32
	charge               uint8
	health               uint16
	lastLog              uint32
}

func (f *e2ePlayerThrownWeaponFixture) currentCharge() uint8 {
	// The reusable RoundChakram has no ammo-use record. Only Shuriken's
	// stock record is meaningful here; do not require or invent one for Round.
	if f.flag != object.WeaponShuriken {
		return 0
	}
	return (*server.AmmoUseData)(f.weapon.UseData.Ptr).Charge1
}

func (f *e2ePlayerThrownWeaponFixture) prepare() {
	f.unit = noxServer.Players.HostUnit()
	if f.unit == nil || f.unit.ControllingPlayer() == nil || f.unit.ControllingPlayer().PlayerClass() != player.Warrior {
		e2eError(fmt.Errorf("player thrown weapon requires a live Warrior"))
		return
	}
	f.weapon = f.unit.UpdateDataPlayer().EquippedWeapon
	if f.weapon == nil || f.weapon.ObjectTypeC().ID() != f.item || !f.weapon.Flags().Has(object.FlagEquipped) ||
		f.unit.ControllingPlayer().WeaponEquip&uint32(f.flag) == 0 || f.weapon.InvHolder != f.unit {
		e2eError(fmt.Errorf("%s was not equipped by actual inventory input: weapon=%p equip=%#x", f.item, f.weapon, f.unit.ControllingPlayer().WeaponEquip))
		return
	}
	if f.weapon.InitData == nil || f.flag == object.WeaponShuriken && f.weapon.UseData.Ptr == nil {
		e2eError(fmt.Errorf("%s lacks initialized stock ammo/modifiers", f.item))
		return
	}
	f.charge = f.currentCharge()
	f.original = f.unit.PosVec
	origin, direction, err := e2eWarriorAbilityArena(f.original, f.unit.Shape.Circle.R+4, func(from, to types.Pointf) bool {
		return e2eWarriorLaneClear(f.unit, from, to)
	})
	if err != nil {
		e2eError(err)
		return
	}
	asObjectS(f.unit).SetPos(origin)
	f.unit.VelVec, f.unit.ForceVec, f.unit.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	f.target = noxServer.NewObjectByTypeID("Troll")
	if f.target == nil {
		e2eError(fmt.Errorf("player thrown weapon has no stock Troll target"))
		return
	}
	noxServer.CreateObjectAt(f.target, nil, origin.Add(direction.Mul(112)))
	noxServer.ObjectsAddPending()
	if f.target.HealthData == nil || f.target.UpdateData == nil || f.target.Damage == nil {
		e2eError(fmt.Errorf("player thrown weapon target not initialized"))
		return
	}
	// Placement, durability and ordinary waiting AI are fixture setup only.
	// Neither attack state, ammo, ownership, projectile nor damage is injected.
	asObjectS(f.target).SetMaxHealth(2000)
	f.target.UpdateDataMonster().SetAggression(0)
	f.target.ClearActionStack()
	f.target.MonsterPushAction(ai.ACTION_WAIT, noxServer.Frame()+250000)
	f.health = f.target.HealthData.Cur
	if unsafe.Sizeof(uintptr(0)) == 8 {
		pointers := []unsafe.Pointer{unsafe.Pointer(f.unit), unsafe.Pointer(f.weapon), f.unit.UpdateData}
		if f.flag == object.WeaponShuriken {
			pointers = append(pointers, f.weapon.UseData.Ptr)
		}
		for _, ptr := range pointers {
			if uintptr(ptr) <= math.MaxUint32 {
				e2eError(fmt.Errorf("player throw native fixture pointer below 4 GiB: %p", ptr))
				return
			}
		}
	}
	e2eLog.Printf("PLAYER THROW PREPARED: item=%s unit=%p weapon=%p equip=%#x flags=%#x charge=%d target=%s health=%d", f.item,
		f.unit, f.weapon, f.unit.ControllingPlayer().WeaponEquip, uint32(f.weapon.ObjFlags), f.charge, f.target.ObjectTypeC().ID(), f.health)
}

func (f *e2ePlayerThrownWeaponFixture) observeLaunch() bool {
	now := noxServer.Frame()
	for _, obj := range noxServer.Objs.AllObjects() {
		if obj.ObjOwner == f.unit && obj.ObjectTypeC().ID() == f.projectileType && !obj.Flags().Has(object.FlagDestroyed) {
			e2eLog.Printf("PLAYER THROW LAUNCHED: item=%s projectile=%p owner=%p pos=%v velocity=%v elapsed=%d", f.item,
				obj, obj.ObjOwner, obj.PosVec, obj.VelVec, now-f.frame)
			return true
		}
	}
	if f.lastLog == 0 || now-f.lastLog >= 10 {
		f.lastLog = now
		update := f.unit.UpdateDataPlayer()
		e2eLog.Printf("PLAYER THROW WAIT: item=%s elapsed=%d state=%d attack-frame=%d->%d anim-frame=%d deadline=%d equipped=%p equip=%#x charge=%d->%d target-health=%d->%d stamina=%d",
			f.item, now-f.frame, update.State, f.attackFrame, f.unit.Field34, update.Field59_0, update.Field0, update.EquippedWeapon,
			f.unit.ControllingPlayer().WeaponEquip, f.charge, f.currentCharge(), f.health, f.target.HealthData.Cur, update.Stamina)
	}
	return false
}

// CheckPlayerThrownWeapon deliberately asserts a functional player launch.
// A timeout is a regression result, not a passing "unsupported" fallback.
// The scenario grants stock inventory and equips through actual UI input;
// this action observes real mouse/button -> network -> player attack -> world.
// Downstream collision/damage/return are not verified if launch fails.
func (sc *e2eScenario) CheckPlayerThrownWeapon(item, name string) {
	projectileType, flag, ok := e2ePlayerThrownWeaponType(item)
	if !ok {
		e2eError(fmt.Errorf("unknown player thrown weapon %q", item))
		return
	}
	f := &e2ePlayerThrownWeaponFixture{item: item, projectileType: projectileType, flag: flag}
	sc.addWhen(0, name+" prepare", 1200, func() bool {
		return noxServer.Players.HostUnit() != nil && noxClient.ClientPlayerUnit() != nil && nox_client_isConnected()
	}, f.prepare)
	sc.Wait(12, name+" publish target and player placement")
	sc.add(0, name+" actual attack button", func() {
		f.frame, f.attackFrame = noxServer.Frame(), f.unit.Field34
		mouse := noxClient.Viewport().ToScreenPos(image.Pt(int(f.target.PosVec.X), int(f.target.PosVec.Y)))
		e2eQueueInput(&seat.MouseMoveEvent{Pos: mouse}, &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true})
		e2eLog.Printf("PLAYER THROW INPUT: item=%s frame=%d target=%v mouse=%v", f.item, f.frame, f.target.PosVec, mouse)
	})
	sc.Input(1, "", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false})
	sc.addWhen(1, name+" actual projectile launch", 90, f.observeLaunch, func() {
		noxServer.DelayedDelete(f.target)
		asObjectS(f.unit).SetPos(f.original)
	})
}
