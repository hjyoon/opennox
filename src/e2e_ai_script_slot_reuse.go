package opennox

import (
	"fmt"
	"math"
	"strings"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

func e2eAIScriptSlotReuseMode(mode string) (kind string, linked, ok bool) {
	kind, path, found := strings.Cut(mode, "/")
	if !found || (kind != "Troll" && kind != "NPC") || (path != "simple" && path != "linked") {
		return "", false, false
	}
	return kind, path == "linked", true
}

type e2eAIScriptSlotReuseFixture struct {
	mode, kind               string
	linked                   bool
	host, unit               *server.Object
	waypoint                 *server.Waypoint
	original, from           types.Pointf
	unitHP, hostHP           uint16
	moveFrame, wanderFrame   uint32
	movePassed, wanderPassed bool
}

func (f *e2eAIScriptSlotReuseFixture) prepare() {
	f.host = noxServer.Players.HostUnit()
	f.original, f.hostHP = f.host.PosVec, f.host.HealthData.Cur
	f.unit = noxServer.NewObjectByTypeID(f.kind)
	if f.unit == nil || f.unit.UpdateData == nil || f.unit.HealthData == nil ||
		(f.kind == "NPC") != f.unit.MonsterClass().Has(object.MonsterNPC) {
		e2eError(fmt.Errorf("script slot reuse requires a stock %s body", f.kind))
		return
	}
	from, to, err := e2eAIFirstAttackArena(f.host, max(f.host.Shape.Circle.R, f.unit.Shape.Circle.R)+12, false)
	if err != nil {
		e2eError(err)
		return
	}
	asObjectS(f.host).SetPos(to)
	// Placement and ordinary ownership are fixture inputs, not combat results.
	// Keep the player beyond the destination so it cannot block the scripted move.
	noxServer.CreateObjectAt(f.unit, f.host, from)
	noxServer.ObjectsAddPending()
	f.from, f.unitHP = f.unit.PosVec, f.unit.HealthData.Cur
	f.waypoint = noxServer.S().NewWaypoint(from.Add(to.Sub(from).Mul(0.65)))
	if f.linked {
		next := noxServer.S().NewWaypoint(from.Add(to.Sub(from).Mul(0.35)))
		f.waypoint.PointsCnt, next.PointsCnt = 1, 1
		f.waypoint.Points[0].Waypoint, next.Points[0].Waypoint = next, f.waypoint
		f.waypoint.Flags2, next.Flags2 = 0xff, 0xff
	}
	for _, ptr := range []unsafe.Pointer{unsafe.Pointer(f.unit), f.unit.UpdateData, unsafe.Pointer(f.host), unsafe.Pointer(f.waypoint)} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ptr) <= math.MaxUint32 {
			e2eError(fmt.Errorf("script slot reuse pointer is not above 4 GiB: %p", ptr))
			return
		}
	}
	// Deliberately prepare a previously used pointer-bearing slot through the
	// production stack API. This tests script dispatch, not autonomous AI choice.
	f.unit.ClearActionStack()
	f.unit.MonsterPushAction(ai.ACTION_WAIT)
	moveIndex := 1
	if f.linked {
		f.unit.MonsterPushAction(ai.ACTION_WAIT)
		moveIndex = 2
	}
	old := f.unit.MonsterPushAction(ai.ACTION_MOVE_TO, f.host.PosVec, f.host)
	if old != &f.unit.UpdateDataMonster().AIStack[moveIndex] || old.ArgObj(2) != f.host {
		e2eError(fmt.Errorf("script slot reuse failed to prepare the native old target"))
		return
	}
	oldTarget := old.Args[2]
	asObjectS(f.unit).Move(f.waypoint)
	head := f.unit.UpdateDataMonster().AIStackHead()
	if head != old || head.Type() != ai.ACTION_FAR_MOVE_TO || head.Args[2] != 0 {
		e2eError(fmt.Errorf("script Move left a stale target: mode=%s head=%v", f.mode, head))
		return
	}
	f.moveFrame = noxServer.Frame()
	e2eLog.Printf("AI SCRIPT REUSE MOVE: mode=%s unit=%p update=%p old-target=%#x cleared-target=%#x waypoint=%p slot=%d frame=%d", f.mode, f.unit, f.unit.UpdateData, oldTarget, head.Args[2], f.waypoint, moveIndex, f.moveFrame)
}

func (f *e2eAIScriptSlotReuseFixture) observeMove() bool {
	if noxServer.Frame()-f.moveFrame < 64 {
		return false
	}
	if !e2eObjectInWorld(f.unit) || f.unit.Flags().Has(object.FlagDead|object.FlagDestroyed) || f.unit.HealthData.Cur != f.unitHP ||
		f.host.HealthData.Cur != f.hostHP || f.unit.PosVec.Sub(f.from).Len() < 32 || noxClient.Objs.ByNetCode(uint16(f.unit.NetCode)) == nil {
		e2eError(fmt.Errorf("script Move did not naturally update/render: mode=%s elapsed=%d position=%v->%v HP=%d/%d", f.mode, noxServer.Frame()-f.moveFrame, f.from, f.unit.PosVec, f.unit.HealthData.Cur, f.unitHP))
		return true
	}
	f.movePassed = true
	e2eLog.Printf("AI SCRIPT REUSE MOVE PASS: mode=%s elapsed=%d movement=%g native-refresh=ordinary-game-loop drawable=present HP=unchanged", f.mode, noxServer.Frame()-f.moveFrame, f.unit.PosVec.Sub(f.from).Len())
	return true
}

func (f *e2eAIScriptSlotReuseFixture) dispatchWander() {
	if !f.movePassed {
		e2eError(fmt.Errorf("script slot reuse must pass actual movement before Wander"))
		return
	}
	f.unit.ClearActionStack()
	f.unit.MonsterPushAction(ai.ACTION_WAIT)
	old := f.unit.MonsterPushAction(ai.ACTION_ROAM, unsafe.Pointer(f.waypoint))
	oldWaypoint := old.Args[0]
	asObjectS(f.unit).Wander()
	head := f.unit.UpdateDataMonster().AIStackHead()
	if head != old || head.Type() != ai.ACTION_ROAM || head.Args[0] != 0 {
		e2eError(fmt.Errorf("script Wander left a stale waypoint: mode=%s head=%v", f.mode, head))
		return
	}
	f.wanderFrame = noxServer.Frame()
	e2eLog.Printf("AI SCRIPT REUSE WANDER: mode=%s old-waypoint=%#x cleared-waypoint=%#x frame=%d", f.mode, oldWaypoint, head.Args[0], f.wanderFrame)
}

func (f *e2eAIScriptSlotReuseFixture) observeWander() bool {
	if noxServer.Frame()-f.wanderFrame < 64 {
		return false
	}
	if !e2eObjectInWorld(f.unit) || f.unit.Flags().Has(object.FlagDead|object.FlagDestroyed) || f.unit.HealthData.Cur != f.unitHP ||
		f.host.HealthData.Cur != f.hostHP || noxClient.Objs.ByNetCode(uint16(f.unit.NetCode)) == nil {
		e2eError(fmt.Errorf("script Wander did not survive normal AI updates: %s", f.mode))
		return true
	}
	f.wanderPassed = true
	e2eLog.Printf("AI SCRIPT REUSE WANDER PASS: mode=%s elapsed=%d native-waypoint-consumer=ordinary-game-loop drawable=present HP=unchanged stack=%v", f.mode, noxServer.Frame()-f.wanderFrame, f.unit.UpdateDataMonster().GetAIStack())
	return true
}

func (f *e2eAIScriptSlotReuseFixture) cleanup() {
	if !f.movePassed || !f.wanderPassed {
		e2eError(fmt.Errorf("incomplete actual script slot reuse lifecycle: %s", f.mode))
		return
	}
	noxServer.DelayedDelete(f.unit)
	asObjectS(f.host).SetPos(f.original)
}

func (sc *e2eScenario) CheckAIScriptSlotReuse(mode, name string) {
	kind, linked, ok := e2eAIScriptSlotReuseMode(mode)
	if !ok {
		e2eError(fmt.Errorf("invalid script slot reuse mode %q", mode))
		return
	}
	f := &e2eAIScriptSlotReuseFixture{mode: mode, kind: kind, linked: linked}
	sc.addWhen(0, name+" prepare native reused target and dispatch Move", 1200, func() bool {
		host := noxServer.Players.HostUnit()
		return nox_client_isConnected() && host != nil && host.HealthData != nil && noxClient.ClientPlayerUnit() != nil && host.Buffs == 0 && host.Poison540 == 0
	}, f.prepare)
	sc.addWhen(1, name+" ordinary Move updates then dispatch Wander", 600, f.observeMove, f.dispatchWander)
	sc.Screen("native_script_slot_reuse_" + kind + fmt.Sprintf("_linked_%t", linked))
	sc.addWhen(1, name+" ordinary Wander updates", 600, f.observeWander, f.cleanup)
	sc.Wait(12, name+" ordinary deletion cleanup")
}
