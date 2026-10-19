package opennox

import (
	"fmt"
	"image"
	"math"

	"github.com/opennox/libs/object"
)

type e2eMonsterInputIdentity struct {
	wire     uint16
	extent   uint32
	scriptID int32
}

type e2eNearestMonsterInputCandidate struct {
	identity  e2eMonsterInputIdentity
	typeID    string
	target    e2eNamedMonsterInputTarget
	hostile   bool
	available bool
}

type e2eNearestMonsterInputSelection struct {
	identity e2eMonsterInputIdentity
	pinned   bool
}

// StrikeNearestMonster selects an existing visible hostile creature of the
// requested type. A single action pins copied identities while approaching;
// it never names, creates, moves, damages or changes the creature's AI. Only
// ordinary mouse input can cause an attack through the native cursor target.
func (sc *e2eScenario) StrikeNearestMonster(typeID string, offset image.Point, name string) {
	if typeID == "" {
		e2eError(fmt.Errorf("nearest monster input requires an original creature type"))
		return
	}
	var selection e2eNearestMonsterInputSelection
	sc.strikeNamedMonsterInput(name, func() (e2eNamedMonsterInputTarget, bool) {
		if noxServer == nil || noxClient == nil || noxClient.Viewport() == nil || noxClient.ClientPlayerUnit() == nil {
			return e2eNamedMonsterInputTarget{}, false
		}
		unit := noxServer.Players.HostUnit()
		if unit == nil || unit.HealthData == nil || unit.HealthData.Cur == 0 || unit.Flags().HasAny(object.FlagDead|object.FlagDestroyed) {
			return e2eNamedMonsterInputTarget{}, false
		}
		var candidates []e2eNearestMonsterInputCandidate
		for npc := noxServer.Objs.First(); npc != nil; npc = npc.Next() {
			if !npc.Class().Has(object.ClassMonster) || npc.HealthData == nil || npc.UpdateData == nil {
				continue
			}
			typ := noxServer.Types.ByInd(int(npc.TypeInd))
			if typ == nil || typ.ID() != typeID {
				continue
			}
			code := noxServer.GetUnitNetCode(npc)
			if code <= 0 || code > int(^uint16(0)) {
				continue
			}
			target, available := e2eReadMonsterInputTarget(npc, offset, true)
			if available && !target.dead {
				available = noxServer.CanInteract(unit, npc, 0)
			}
			candidates = append(candidates, e2eNearestMonsterInputCandidate{
				identity: e2eMonsterInputIdentity{wire: uint16(code), extent: npc.Extent, scriptID: npc.ScriptIDVal},
				typeID:   typeID, target: target, hostile: noxServer.IsEnemyTo(unit, npc), available: available,
			})
		}
		target, available := selection.observe(typeID, candidates)
		if available {
			target.playerHP = unit.HealthData.Cur
		}
		return target, available
	}, e2eQueueInput, func(target e2eNamedMonsterInputTarget) {
		e2eLog.Printf("NEAREST MONSTER INPUT RESULT: type=%q wire=%#x point=%v distance=%g health=%d player_hp=%d observed_dead=%t observed_absent=%t left_press=%t frame=%d",
			typeID, target.wire, target.point, target.distance, target.health, target.playerHP,
			target.dead, target.absent, !target.dead && !target.absent, noxServer.Frame())
	})
}

func (s *e2eNearestMonsterInputSelection) observe(typeID string, candidates []e2eNearestMonsterInputCandidate) (e2eNamedMonsterInputTarget, bool) {
	if s.pinned {
		for _, candidate := range candidates {
			if candidate.identity != s.identity || candidate.typeID != typeID {
				continue
			}
			if !candidate.hostile {
				return e2eNamedMonsterInputTarget{absent: true}, true
			}
			if !candidate.available || !e2eMonsterInputFiniteDistance(candidate.target.distance) {
				return e2eNamedMonsterInputTarget{}, false
			}
			target := candidate.target
			target.wire = candidate.identity.wire
			return target, true
		}
		// Retirement or a recycled network code cannot authorize another click.
		// Absence is reported separately from an observed original death.
		return e2eNamedMonsterInputTarget{absent: true}, true
	}
	var best *e2eNearestMonsterInputCandidate
	for i := range candidates {
		candidate := &candidates[i]
		if candidate.identity.wire == 0 || candidate.typeID != typeID || !candidate.hostile || !candidate.available ||
			candidate.target.dead || candidate.target.absent || candidate.target.health == 0 ||
			!e2eMonsterInputFiniteDistance(candidate.target.distance) || candidate.target.distance > 512 {
			continue
		}
		if best == nil || candidate.target.distance < best.target.distance ||
			candidate.target.distance == best.target.distance && candidate.identity.wire < best.identity.wire {
			best = candidate
		}
	}
	if best == nil {
		return e2eNamedMonsterInputTarget{absent: true}, true
	}
	s.identity, s.pinned = best.identity, true
	target := best.target
	target.wire = best.identity.wire
	return target, true
}

func e2eMonsterInputFiniteDistance(distance float32) bool {
	return distance >= 0 && !math.IsNaN(float64(distance)) && !math.IsInf(float64(distance), 0)
}
