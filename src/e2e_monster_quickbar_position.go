package opennox

import (
	"github.com/opennox/libs/client/keybind"
	"github.com/opennox/libs/object"
)

// This observer owns copied identities and frame words, with no native
// object/drawable pointer or input-event sink.
type e2eMonsterQuickbarPositionWatch struct {
	start, last, limit uint32
	playerScriptID     int32
	stopped            bool
}

func (w *e2eMonsterQuickbarPositionWatch) liveFrame(frame uint32) bool {
	if w.stopped {
		return false
	}
	if frame-w.start > w.limit {
		w.stopped = true
		return false
	}
	return true
}

func (w *e2eMonsterQuickbarPositionWatch) observe(sample e2eMonsterQuickbarSample, found bool) bool {
	if !w.liveFrame(sample.frame) {
		return false
	}
	if !found || !sample.hasPositions || sample.playerScriptID != w.playerScriptID {
		w.stopped = true
		return false
	}
	if sample.frame == w.last {
		return false
	}
	w.last = sample.frame
	return true
}

func e2eLogMonsterQuickbarPosition(sample e2eMonsterQuickbarSample, identity e2eMonsterInputIdentity, key keybind.Key, castFrame uint32, queued bool) {
	e2eLog.Printf("MONSTER QUICKBAR POSITION: {\"frame\":%d,\"cast_frame\":%d,\"player_script_id\":%d,\"script_id\":%d,\"extent\":%d,\"wire\":%d,\"key\":%d,\"key_queued\":%t,\"player\":[%g,%g],\"target\":[%g,%g]}",
		sample.frame, castFrame, sample.playerScriptID, identity.scriptID, identity.extent, identity.wire,
		key, queued, sample.playerPosition[0], sample.playerPosition[1], sample.targetPosition[0], sample.targetPosition[1])
}

func e2eReadMonsterQuickbarPosition(typeID string, identity e2eMonsterInputIdentity) (e2eMonsterQuickbarSample, bool) {
	sample := e2eMonsterQuickbarSample{frame: noxServer.Frame()}
	unit := noxServer.Players.HostUnit()
	if unit == nil || unit.HealthData == nil || unit.HealthData.Cur == 0 || unit.Flags().HasAny(object.FlagDead|object.FlagDestroyed) {
		return sample, false
	}
	sample.playerScriptID = unit.ScriptIDVal
	sample.playerPosition = [2]float32{unit.PosVec.X, unit.PosVec.Y}
	for npc := noxServer.Objs.First(); npc != nil; npc = npc.Next() {
		if npc.ScriptIDVal != identity.scriptID || npc.Extent != identity.extent ||
			!npc.Class().Has(object.ClassMonster) || npc.HealthData == nil || npc.HealthData.Cur == 0 ||
			npc.Flags().HasAny(object.FlagDead|object.FlagDestroyed) || noxServer.GetUnitNetCode(npc) != int(identity.wire) {
			continue
		}
		typ := noxServer.Types.ByInd(int(npc.TypeInd))
		if typ == nil || typ.ID() != typeID {
			continue
		}
		sample.targetPosition, sample.hasPositions = [2]float32{npc.PosVec.X, npc.PosVec.Y}, true
		return sample, true
	}
	return sample, false
}

func e2eWatchMonsterQuickbarPositions(typeID string, identity e2eMonsterInputIdentity, playerID int32, key keybind.Key, castFrame, frame uint32) {
	world := noxServer
	if world == nil {
		return
	}
	watch := e2eMonsterQuickbarPositionWatch{start: frame, last: frame, limit: 2 * world.TickRate(), playerScriptID: playerID}
	world.TickHook(func() {
		if !watch.liveFrame(world.Frame()) || world.Frame() == watch.last {
			return
		}
		if noxServer != world {
			watch.stopped = true
			return
		}
		sample, found := e2eReadMonsterQuickbarPosition(typeID, identity)
		if watch.observe(sample, found) {
			e2eLogMonsterQuickbarPosition(sample, identity, key, castFrame, true)
			e2eObserveMonsterQuickbarSwapTarget(sample.frame, castFrame, sample.playerScriptID, identity, key)
		}
	})
}
