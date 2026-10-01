package opennox

import (
	"fmt"
	"math"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

// Cache only values read from a LIVE stock missile owned by the selected stock
// FlyingGolem. Damage attribution can retain a pointer after delayed deletion;
// that pointer is used only as a map key, never dereferenced by this observer.
type e2eQuestGolemArrow struct {
	typeInd                 uint16
	firstFrame, lastFrame   uint32
	coopDamage, otherDamage int32
}

type e2eQuestGolemPierceFixture struct {
	combat             e2eQuestCombatFixture
	arrows             map[*server.Object]e2eQuestGolemArrow
	arrowType          uint16
	previousHP, lastHP uint16
	lastReportBefore   uint16
	lastHitFrame       uint32
	hits, required     int
	ticks              uint32
	lastDelta          int16
	clientDelta        int16
	hud                e2eHUDMeterState
	pixels             e2eHUDPixelStats
}

func e2eQuestGolemPierceSample(arrow e2eQuestGolemArrow, known bool, arrowType uint16,
	frame, startFrame, previousHit, damageFrame, damageType uint32, before, after uint16,
) bool {
	return known && arrowType != 0 && arrow.typeInd == arrowType && arrow.otherDamage > 0 &&
		arrow.firstFrame >= startFrame && arrow.lastFrame >= arrow.firstFrame &&
		damageFrame >= arrow.lastFrame && damageFrame-arrow.lastFrame <= 2 &&
		damageFrame > previousHit && damageFrame <= frame && frame-damageFrame <= 2 &&
		damageType == uint32(object.DamageImpale) && after > 0 && after < before
}

// GAME.EXE delays damage-number reports until the last damage is more than
// two frames old. Another stock enemy can hit in that interval, so the report
// may include both its damage and the confirmed arrow hit. Require the actual
// recipient cache to advance past the arrow HP, then match that exact report.
func e2eQuestGolemPierceReport(before, reported, arrowHP, current uint16, delta int16) bool {
	return arrowHP > 0 && reported > 0 && reported < before && reported <= arrowHP &&
		current == reported && delta < 0 && delta == int16(reported-before)
}

func (f *e2eQuestGolemPierceFixture) prepare() {
	if f.required < 1 || f.required > 12 {
		e2eError(fmt.Errorf("Quest GolemArrow requires 1..12 observed hits, got %d", f.required))
		return
	}
	f.combat.prepare()
	if f.combat.unit == nil || f.combat.target == nil {
		return
	}
	if !f.combat.target.Class().Has(object.ClassMonster) || f.combat.target.ObjOwner != nil ||
		f.combat.target.UpdateDataMonster().MonsterDef == nil || f.combat.unit.UpdateDataPlayer() == nil {
		e2eError(fmt.Errorf("Quest GolemArrow: stock encounter definition/ownership or player update is missing"))
		return
	}
	typ := noxServer.Types.ByID("GolemArrow")
	if typ == nil || typ.Ind() <= 0 || typ.Ind() > math.MaxUint16 {
		e2eError(fmt.Errorf("Quest GolemArrow: missing stock missile definition"))
		return
	}
	f.arrowType = uint16(typ.Ind())
	f.arrows = make(map[*server.Object]e2eQuestGolemArrow)
	f.previousHP, f.lastHP = f.combat.playerHP, f.combat.playerHP
	f.lastHitFrame = f.combat.startFrame
	e2eLog.Printf("QUEST GOLEM PIERCE APPROACH: frame=%d player=%p stock-encounter=%p arrow-type=%d hp=%d required-hits=%d player-placement-only=true",
		f.combat.startFrame, f.combat.unit, f.combat.target, f.arrowType, f.previousHP, f.required)
}

func (f *e2eQuestGolemPierceFixture) stockGolemReady() bool {
	if f.arrows == nil || !e2eObjectInWorld(f.combat.unit) || f.combat.unit.HealthData.Cur == 0 {
		return false
	}
	var target *server.Object
	distance := float64(640)
	for obj := noxServer.Objs.First(); obj != nil; obj = obj.Next() {
		if obj.ObjectTypeC().ID() != "FlyingGolem" || !obj.Class().Has(object.ClassMonster) ||
			obj.ObjOwner != nil || obj.UpdateData == nil || obj.UpdateDataMonster().MonsterDef == nil ||
			obj.HealthData == nil || obj.HealthData.Cur == 0 || obj.Flags().HasAny(object.FlagDead|object.FlagDestroyed) {
			continue
		}
		d := math.Hypot(float64(obj.PosVec.X-f.combat.unit.PosVec.X), float64(obj.PosVec.Y-f.combat.unit.PosVec.Y))
		if d < distance {
			target, distance = obj, d
		}
	}
	if target == nil {
		return false
	}
	f.combat.target, f.combat.targetHP, f.combat.typeID = target, target.HealthData.Cur, "FlyingGolem"
	f.previousHP = f.combat.unit.HealthData.Cur
	e2eLog.Printf("QUEST GOLEM PIERCE PREPARED: frame=%d player=%p golem=%p hp=%d golem-hp=%d distance=%.3f stock-generation-and-AI=true",
		noxServer.Frame(), f.combat.unit, target, f.previousHP, f.combat.targetHP, distance)
	return true
}

func (f *e2eQuestGolemPierceFixture) observe() bool {
	if f.arrows == nil {
		return false
	}
	unit, target := f.combat.unit, f.combat.target
	if !e2eObjectInWorld(unit) || unit.HealthData == nil || unit.HealthData.Cur == 0 ||
		!e2eObjectInWorld(target) || target.HealthData == nil || target.HealthData.Cur == 0 {
		e2eError(fmt.Errorf("Quest GolemArrow: player or stock golem died/was removed before verification"))
		return false
	}
	frame := noxServer.Frame()
	for _, missile := range noxServer.Objs.AllMissiles() {
		if missile.TypeInd != f.arrowType || missile.ObjOwner != target {
			// Pool reuse by another live missile invalidates the old identity.
			delete(f.arrows, missile)
			continue
		}
		if missile.Flags().HasAny(object.FlagDead|object.FlagDestroyed) || missile.CollideData == nil {
			continue
		}
		data := (*server.MonsterArrowCollideData)(missile.CollideData)
		arrow, known := f.arrows[missile]
		if !known || frame-arrow.lastFrame > 2 {
			arrow = e2eQuestGolemArrow{typeInd: missile.TypeInd, firstFrame: frame,
				coopDamage: data.CoopDamage, otherDamage: data.OtherDamage}
			e2eLog.Printf("QUEST GOLEM ARROW LIVE: frame=%d arrow=%p owner=%p type=%d coop-damage=%d other-damage=%d pos=%v",
				frame, missile, target, missile.TypeInd, arrow.coopDamage, arrow.otherDamage, missile.PosVec)
		}
		arrow.lastFrame = frame
		f.arrows[missile] = arrow
	}
	currentHP := unit.HealthData.Cur
	arrow, known := f.arrows[unit.Obj130]
	if e2eQuestGolemPierceSample(arrow, known, f.arrowType, frame, f.combat.startFrame,
		f.lastHitFrame, unit.Frame134, unit.Field131, f.previousHP, currentHP) {
		f.hits++
		f.lastHitFrame, f.lastHP = unit.Frame134, currentHP
		f.lastDelta = int16(currentHP) - int16(f.previousHP)
		if pl := unit.ControllingPlayer(); pl != nil {
			if sample := unitHealthSampleNative4D8760(unit, int(pl.PlayerIndex())); sample != nil {
				f.lastReportBefore = *sample
			}
		}
		ud := unit.UpdateDataPlayer()
		e2eLog.Printf("QUEST GOLEM PIERCE HIT: hit=%d frame=%d damage-frame=%d arrow-last-live=%d arrow=%p hp=%d->%d damage-type=%d raw-damage=%d armor=%g carry=%g marker=%d/%d",
			f.hits, frame, unit.Frame134, arrow.lastFrame, unit.Obj130, f.previousHP, currentHP, unit.Field131,
			arrow.otherDamage, math.Float32frombits(ud.Field57), math.Float32frombits(ud.Field21), ud.Field75, ud.Field76)
	}
	f.previousHP = currentHP
	f.ticks++
	if f.ticks%120 == 0 {
		e2eLog.Printf("QUEST GOLEM PIERCE WAIT: frame=%d hits=%d/%d hp=%d damage-type=%d damage-frame=%d cached-arrows=%d",
			frame, f.hits, f.required, currentHP, unit.Field131, unit.Frame134, len(f.arrows))
		f.combat.logState()
	}
	return true
}

func (f *e2eQuestGolemPierceFixture) ready() bool {
	if !f.observe() || f.hits < f.required ||
		noxServer.Frame()-f.lastHitFrame > 30 {
		return false
	}
	drawable := noxClient.ClientPlayerUnit()
	if drawable == nil || drawable.NetCode32 != uint32(noxServer.GetUnitNetCode(f.combat.unit)) {
		return false
	}
	delta, ok := legacy.HealthChangeForDrawable(drawable.NetCode32)
	hud, hudOK := e2eClientHUDMeter(0)
	pl := f.combat.unit.ControllingPlayer()
	if pl == nil {
		return false
	}
	sample := unitHealthSampleNative4D8760(f.combat.unit, int(pl.PlayerIndex()))
	if sample == nil || !ok || !e2eQuestGolemPierceReport(f.lastReportBefore, *sample, f.lastHP, f.combat.unit.HealthData.Cur, delta) ||
		!hudOK || hud.Current != uint32(*sample) || hud.Maximum != uint32(f.combat.unit.HealthData.Max) {
		return false
	}
	pixels, err := e2eAssertHUDMeterPixels(noxClient.r.CopyPixBuffer(), hud, 'r', "stock GolemArrow PIERCE")
	if err != nil {
		return false
	}
	f.hud, f.pixels, f.clientDelta = hud, pixels, delta
	return true
}

func (sc *e2eScenario) CheckQuestGolemPierce(count int, name string) {
	f := &e2eQuestGolemPierceFixture{combat: e2eQuestCombatFixture{typeID: "Necromancer"}, required: count}
	sc.add(0, name+" place only player beside stock encounter", f.prepare)
	sc.addWhen(0, name+" wait for stock FlyingGolem generation", 600, f.stockGolemReady, nil)
	sc.addWhen(0, name+" wait for stock GolemArrow hits and client display", 1800, f.ready, func() {
		e2eLog.Printf("QUEST GOLEM PIERCE VERIFIED: frame=%d hits=%d/%d hp=%d->%d last-arrow-delta=%d HUD=%d/%d report-before=%d client-aggregate-delta=%d fill-height=%d red-filled=%d red-empty=%d pixel-rect=%v stock-AI-and-damage=true",
			noxServer.Frame(), f.hits, f.required, f.combat.playerHP, f.lastHP, f.lastDelta, f.hud.Current, f.hud.Maximum,
			f.lastReportBefore, f.clientDelta, f.pixels.FillHeight, f.pixels.Filled, f.pixels.Empty, f.pixels.Rect)
	})
	sc.Screen(name + " health and damage display")
}
