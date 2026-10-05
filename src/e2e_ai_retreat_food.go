package opennox

import (
	"fmt"
	"math"
	"strings"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

func e2eAIRetreatFoodMode(mode string) (kind, item string, descending, ok bool) {
	parts := strings.Split(mode, "/")
	if len(parts) != 3 || parts[0] != "Troll" && parts[0] != "NPC" || parts[2] != "RedApple" && parts[2] != "Meat" {
		return "", "", false, false
	}
	switch parts[1] {
	case "ascending":
	case "descending":
		descending = true
	default:
		return "", "", false, false
	}
	return parts[0], parts[2], descending, true
}

// Observe the complete live follow-up, including all three cached native food
// identities. Never select food, push/pop an action or change a target here.
func e2eAIRetreatFoodMoveStack(update *server.MonsterUpdateData, food *server.Object, pos types.Pointf) bool {
	if update == nil || food == nil || update.AIStackInd < 5 || int(update.AIStackInd) >= len(update.AIStack) {
		return false
	}
	stack := update.GetAIStack()
	base := len(stack) - 6
	want := [...]ai.ActionType{ai.ACTION_RETREAT, ai.DEPENDENCY_NOT_HEALTHY, ai.DEPENDENCY_NO_VISIBLE_ENEMY,
		ai.DEPENDENCY_OBJECT_AT_VISIBLE_LOCATION, ai.ACTION_PICKUP_OBJECT, ai.ACTION_MOVE_TO}
	for i, action := range want {
		if stack[base+i].Type() != action {
			return false
		}
	}
	visible, pickup, move := &stack[base+3], &stack[base+4], &stack[base+5]
	return visible.ArgObj(2) == food && visible.ArgPos(0) == pos &&
		pickup.ArgObj(0) == food && move.ArgObj(2) == food && move.ArgPos(0) == pos
}

// Model only the unchanged classic regeneration on a bounded live-world
// probe. Script DamageTrue does not set the ordinary injury-pause timestamp.
// No health/action service is called and no observed HP is used as an input.
func e2eAIRetreatFoodRegenerationHP(start, frame, injury, fps uint32, maximum, initial uint16) (uint16, bool) {
	elapsed := frame - start
	if elapsed > 600 || fps == 0 || maximum == 0 || initial > maximum {
		return 0, false
	}
	health := int32(initial)
	for offset := uint32(1); offset <= elapsed; offset++ {
		health += e2eMeteorShowerRegenAmount(start+offset, injury, fps, int32(maximum), health)
	}
	return uint16(health), true
}

type e2eAIRetreatFoodFixture struct {
	mode, kind, item               string
	descending                     bool
	host, unit, food               *server.Object
	original, origin, foodPos      types.Pointf
	hostHP, hostMax, initialHP     uint16
	injuredHP, expectedHP          uint16
	unitWire, foodWire             uint16
	start, moveFrame, eatFrame     uint32
	lastLog                        uint32
	travelSquared                  float64
	sounds                         int
	active, moving, moved, ate     bool
	unitDrawn, foodDrawn, complete bool
}

func (f *e2eAIRetreatFoodFixture) prepare() {
	f.host = noxServer.Players.HostUnit()
	if f.host == nil || f.host.HealthData == nil || f.host.Buffs != 0 || f.host.Poison540 != 0 ||
		noxflags.HasGame(noxflags.GameModeCoop|noxflags.GameModeQuest) {
		e2eError(fmt.Errorf("RETREAT food needs an unenchanted regular-game host"))
		return
	}
	f.original, f.hostHP, f.hostMax = f.host.PosVec, f.host.HealthData.Cur, f.host.HealthData.Max
	f.unit, f.food = noxServer.NewObjectByTypeID(f.kind), noxServer.NewObjectByTypeID(f.item)
	handler, hasPickup := server.ObjectPickupHandler("FoodPickup")
	if f.unit == nil || f.unit.UpdateData == nil || f.unit.HealthData == nil || f.food == nil ||
		!f.food.Class().Has(object.ClassFood) || !hasPickup || handler.Ptr == nil || f.food.Pickup.Ptr != handler.Ptr ||
		f.food.UseData.Ptr == nil || f.food.Use.Get() == nil {
		e2eError(fmt.Errorf("RETREAT food stock body/food definition missing: %s", f.mode))
		return
	}
	from, to, err := e2eAIFirstAttackArena(f.host, max(f.host.Shape.Circle.R, f.unit.Shape.Circle.R)+12, f.descending)
	if err != nil {
		e2eError(err)
		return
	}
	asObjectS(f.host).SetPos(to)
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
	// Ordinary ownership prevents the host from becoming a sight enemy. Low
	// aggression alone only excludes periodic nearby-food consumption.
	noxServer.CreateObjectAt(f.unit, f.host, from)
	noxServer.CreateObjectAt(f.food, nil, to)
	noxServer.ObjectsAddPending()
	// Stock hosted-game food cannot be picked up by incidental contact. Keep
	// its flags unchanged and require the real inventory-placement callback.
	if !f.food.Flags().Has(object.FlagNoCollide) {
		e2eError(fmt.Errorf("RETREAT stock food unexpectedly collides: %s", f.mode))
		return
	}
	f.origin, f.foodPos = f.unit.PosVec, f.food.PosVec
	f.unitWire = uint16(noxServer.GetUnitNetCode(f.unit))
	f.foodWire = uint16(noxServer.GetUnitNetCode(f.food))
	for _, ptr := range []unsafe.Pointer{unsafe.Pointer(f.unit), f.unit.UpdateData, unsafe.Pointer(f.food)} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ptr) <= math.MaxUint32 {
			e2eError(fmt.Errorf("RETREAT food object/update is not above 4 GiB: %p", ptr))
			return
		}
	}
	// These are normal map/script properties on stock configurable bodies.
	// Passive aggression isolates RETREAT's PICKUP from the separate periodic
	// nearby-food branch (which requires aggression >= 0.08). No AI result is
	// supplied. This is not a claim about default Troll/NPC configuration.
	asObjectS(f.unit).SetAggression(0.01)
	asObjectS(f.unit).SetRetreatLevel(0.98)
	asObjectS(f.unit).SetRegroupLevel(1)
	update := f.unit.UpdateDataMonster()
	f.initialHP = f.unit.HealthData.Cur
	loss := min(10, int(f.food.UseDataConsume().Value))
	if f.initialHP != f.unit.HealthData.Max || int(f.initialHP) <= loss || loss <= 0 ||
		f.unit.Owner() != f.host || f.unit.IsEnemyTo(f.host) || f.host.IsEnemyTo(f.unit) ||
		update.CurrentEnemy != nil || update.PreferredEnemy != nil || update.HasAction(ai.ACTION_RETREAT) ||
		!noxServer.S().CanInteract(f.unit, f.food, 0) {
		e2eError(fmt.Errorf("RETREAT food initial stock state invalid: %s", f.mode))
		return
	}
	// Ordinary script damage only arms the injury. The production AI must
	// enter RETREAT, select/move/pick up the food and heal without further input.
	if !asObjectS(f.unit).DoDamage(nil, loss, object.DamageTrue) {
		e2eError(fmt.Errorf("RETREAT food ordinary script injury rejected: %s", f.mode))
		return
	}
	f.injuredHP = f.unit.HealthData.Cur
	f.expectedHP = uint16(min(int(f.injuredHP)+int(f.food.UseDataConsume().Value), int(f.unit.HealthData.Max)))
	if int(f.injuredHP) != int(f.initialHP)-loss || update.CurrentEnemy != nil || update.PreferredEnemy != nil {
		e2eError(fmt.Errorf("RETREAT food injury/absence of enemy not observed: %s", f.mode))
		return
	}
	f.start, f.active = noxServer.Frame(), true
	noxServer.Audio.OnSound(f.observeSound)
	noxServer.TickHook(f.tick)
	e2eLog.Printf("AI RETREAT FOOD PREPARED: mode=%s frame=%d native=%p/%p food=%p wire=%d/%d pos=%v->%v HP=%d->%d stock-heal=%d aggression=%g threshold=%g/%g ordinary-owner=%p stack=unforced injury=script", f.mode, f.start, f.unit, update, f.food, f.unitWire, f.foodWire, f.origin, f.foodPos, f.initialHP, f.injuredHP, f.food.UseDataConsume().Value, update.Aggression, update.RetreatLevel, update.ResumeLevel, f.unit.Owner())
}

func (f *e2eAIRetreatFoodFixture) tick() {
	if !f.active {
		return
	}
	if !e2eObjectInWorld(f.unit) || f.unit.Flags().HasAny(object.FlagDead|object.FlagDestroyed) {
		e2eError(fmt.Errorf("RETREAT food unit disappeared: %s", f.mode))
		return
	}
	update := f.unit.UpdateDataMonster()
	if update.CurrentEnemy != nil || update.PreferredEnemy != nil || f.unit.Buffs != 0 ||
		update.Aggression >= 0.08 || f.host.HealthData.Cur != f.hostHP || f.host.HealthData.Max != f.hostMax {
		e2eError(fmt.Errorf("RETREAT food received unrelated combat/effect: %s current=%p preferred=%p host=%p buffs=%x aggression=%g host-HP=%d/%d expected=%d/%d stack=%v", f.mode, update.CurrentEnemy, update.PreferredEnemy, f.host, f.unit.Buffs, update.Aggression, f.host.HealthData.Cur, f.host.HealthData.Max, f.hostHP, f.hostMax, update.GetAIStack()))
		return
	}
	dx, dy := float64(f.unit.PosVec.X)-float64(f.origin.X), float64(f.unit.PosVec.Y)-float64(f.origin.Y)
	f.travelSquared = max(f.travelSquared, dx*dx+dy*dy)
	if e2eAIRetreatFoodMoveStack(update, f.food, f.foodPos) {
		if !f.moving {
			f.moveFrame = noxServer.Frame()
			e2eLog.Printf("AI RETREAT FOOD MOVE: mode=%s frame=%d target=%p pos=%v stack=%v", f.mode, f.moveFrame, f.food, f.unit.PosVec, update.GetAIStack())
		}
		f.moving = true
		f.moved = f.moved || f.travelSquared > 16*16
	}
	f.unitDrawn = f.unitDrawn || noxClient.Objs.ByNetCode(f.unitWire) != nil
	f.foodDrawn = f.foodDrawn || noxClient.Objs.ByNetCode(f.foodWire) != nil
	ext := f.unit.GetExt()
	wantHP, validHP := e2eAIRetreatFoodRegenerationHP(f.start, noxServer.Frame(), f.unit.Frame134, noxServer.TickRate(), f.initialHP, f.injuredHP)
	if f.unit.HealthData.Max != f.initialHP || ext.HealthRegenToMax > 0 || ext.HealthRegenPerFrame >= 0 ||
		!f.ate && (!validHP || f.unit.HealthData.Cur != wantHP || !e2eObjectInWorld(f.food)) {
		e2eError(fmt.Errorf("RETREAT food HP/definition/deletion differs from classic regeneration or observed PICKUP: %s frame=%d injury=%d HP=%d expected=%d valid=%t food-live=%t ate=%t stack=%v", f.mode, noxServer.Frame(), f.unit.Frame134, f.unit.HealthData.Cur, wantHP, validHP, e2eObjectInWorld(f.food), f.ate, update.GetAIStack()))
		return
	}
	if !f.ate {
		before, validBefore := e2eAIRetreatFoodRegenerationHP(f.start, noxServer.Frame()-1, f.unit.Frame134, noxServer.TickRate(), f.initialHP, f.injuredHP)
		if validBefore && wantHP > before {
			e2eLog.Printf("AI RETREAT FOOD REGEN: mode=%s HP=%d->%d frame=%d injury=%d", f.mode, before, wantHP, noxServer.Frame(), f.unit.Frame134)
		}
	}
	if noxServer.Frame()-f.lastLog >= 30 {
		f.lastLog = noxServer.Frame()
		e2eLog.Printf("AI RETREAT FOOD TICK: mode=%s elapsed=%d pos=%v HP=%d moving=%t moved=%t ate=%t draw=%t/%t stack=%v", f.mode, noxServer.Frame()-f.start, f.unit.PosVec, f.unit.HealthData.Cur, f.moving, f.moved, f.ate, f.unitDrawn, f.foodDrawn, update.GetAIStack())
	}
}

func (f *e2eAIRetreatFoodFixture) observeSound(id sound.ID, kind int, owner *server.Object, _ types.Pointf) {
	if !f.active || id != sound.SoundMonsterEatFood || owner != f.unit {
		return
	}
	update := owner.UpdateDataMonster()
	head := update.AIStackHead()
	dx, dy := float64(owner.PosVec.X)-float64(f.foodPos.X), float64(owner.PosVec.Y)-float64(f.foodPos.Y)
	if kind != 0 || !f.moving || !f.moved || !e2eObjectInWorld(f.food) ||
		head == nil || head.Type() != ai.ACTION_PICKUP_OBJECT || head.ArgObj(0) != f.food ||
		!update.HasAction(ai.ACTION_RETREAT) || !(dx*dx+dy*dy < 75*75) || owner.HealthData.Cur != f.expectedHP {
		e2eError(fmt.Errorf("RETREAT food consumption lacks live movement/PICKUP/heal: %s HP=%d/%d head=%v moved=%t", f.mode, owner.HealthData.Cur, f.expectedHP, head, f.moved))
		return
	}
	f.sounds++
	if f.sounds != 1 {
		e2eError(fmt.Errorf("RETREAT food consumed more than once: %s", f.mode))
		return
	}
	f.eatFrame, f.ate = noxServer.Frame(), true
	e2eLog.Printf("AI RETREAT FOOD PICKUP: mode=%s frame=%d native-target=%p HP=%d->%d pos=%v distance=%g sound=%d/%d", f.mode, f.eatFrame, head.ArgObj(0), f.injuredHP, owner.HealthData.Cur, owner.PosVec, math.Sqrt(dx*dx+dy*dy), id, kind)
}

func (f *e2eAIRetreatFoodFixture) observe() bool {
	if !f.active || !f.ate || !f.unitDrawn || !f.foodDrawn || e2eObjectInWorld(f.food) ||
		noxClient.Objs.ByNetCode(f.foodWire) != nil || f.unit.UpdateDataMonster().HasAction(ai.ACTION_RETREAT) {
		return false
	}
	dr := noxClient.Objs.ByNetCode(f.unitWire)
	if dr == nil || f.unit.HealthData.Cur != f.expectedHP || f.sounds != 1 {
		return false
	}
	path, err := e2eWriteMagicFrame("", noxClient.r.CopyPixBuffer())
	if err != nil {
		e2eError(err)
		return true
	}
	f.complete, f.active = true, false
	e2eLog.Printf("AI RETREAT FOOD PASS: mode=%s elapsed=%d move/eat=%d/%d HP=%d->%d->%d travel=%g sound=1 native-food=%p client-unit=%p food-deleted=server/client RETREAT=resumed path=%s", f.mode, noxServer.Frame()-f.start, f.moveFrame, f.eatFrame, f.initialHP, f.injuredHP, f.unit.HealthData.Cur, math.Sqrt(f.travelSquared), f.food, dr, path)
	return true
}

func (f *e2eAIRetreatFoodFixture) cleanup() {
	if !f.complete {
		e2eError(fmt.Errorf("RETREAT food incomplete: %s", f.mode))
		return
	}
	noxServer.DelayedDelete(f.unit)
	asObjectS(f.host).SetPos(f.original)
	f.host.VelVec, f.host.ForceVec, f.host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
}

func (sc *e2eScenario) CheckAIRetreatFood(mode, name string) {
	kind, item, descending, ok := e2eAIRetreatFoodMode(mode)
	if !ok {
		e2eError(fmt.Errorf("invalid RETREAT food mode %q", mode))
		return
	}
	f := &e2eAIRetreatFoodFixture{mode: mode, kind: kind, item: item, descending: descending}
	sc.addWhen(0, name+" prepare distant food and ordinary script injury", 1200, func() bool {
		unit := noxServer.Players.HostUnit()
		return nox_client_isConnected() && unit != nil && noxClient.ClientPlayerUnit() != nil && unit.Buffs == 0 && unit.Poison540 == 0
	}, f.prepare)
	sc.addWhen(1, name+" observe unforced RETREAT movement PICKUP consumption", 600, f.observe, f.cleanup)
	sc.Wait(12, name+" finalize ordinary unit deletion")
}
