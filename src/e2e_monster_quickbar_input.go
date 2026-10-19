package opennox

import (
	"fmt"
	"image"
	"math"

	"github.com/opennox/libs/client/keybind"
	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/legacy"
)

type e2eMonsterQuickbarSample struct {
	target    e2eNamedMonsterInputTarget
	frame     uint32
	mouse     image.Point
	competing bool
}

// Only copied input state survives a frame. The existing selection also pins
// copied wire/extent/script identities, never native object/drawable pointers.
type e2eMonsterQuickbarTracker struct {
	key       keybind.Key
	ticks     uint32
	reads     int
	aimFrame  uint32
	castFrame uint32
	probe     int
	aiming    bool
	started   bool
	pressed   bool
	finished  bool
}

func e2eMonsterQuickbarConfig(typeID string, slot, count int) (keybind.Key, uint32, error) {
	if typeID == "" {
		return 0, 0, fmt.Errorf("quickbar input requires an original monster type")
	}
	if slot == 0 {
		slot = 1
	}
	if count == 0 {
		count = 18
	}
	if slot < 1 || slot > 5 || count < 1 || count > 120 {
		return 0, 0, fmt.Errorf("quickbar input requires slot1..5 and tracking1..120 native frames")
	}
	keys := [...]keybind.Key{keybind.KeyA, keybind.KeyS, keybind.KeyD, keybind.KeyF, keybind.KeyG}
	return keys[slot-1], uint32(count), nil
}

func e2eMonsterQuickbarProbe(index int) image.Point {
	points := [...]image.Point{{}, {-8, 0}, {8, 0}, {0, 8}, {0, -8}, {-12, 4}, {12, 4}, {0, 12}}
	return points[index%len(points)]
}

// step emits ordinary mouse motion and one key press/release. It cannot cast
// a spell or select a target directly. Nil native hover is allowed for ordinary
// position aiming; a different selected object never authorizes the key press.
func (st *e2eMonsterQuickbarTracker) step(sample e2eMonsterQuickbarSample, available bool) ([]seat.InputEvent, bool) {
	if st.finished {
		return nil, true
	}
	st.reads++
	unusable := !available || sample.target.absent || sample.target.dead || sample.target.health == 0 ||
		sample.target.playerHP == 0 || !e2eMonsterInputFiniteDistance(sample.target.distance) || sample.target.distance > 300
	var events []seat.InputEvent
	if st.pressed && (unusable || sample.frame != st.castFrame || st.reads >= 600) {
		events = append(events, &seat.KeyboardEvent{Key: st.key, Pressed: false})
		st.pressed = false
	}
	if unusable || st.reads >= 600 {
		st.finished = true
		return events, true
	}
	events = append(events, &seat.MouseMoveEvent{Pos: sample.target.point, Relative: false})
	close := math.Hypot(float64(sample.mouse.X)-float64(sample.target.point.X), float64(sample.mouse.Y)-float64(sample.target.point.Y)) <= 8
	if sample.competing && close {
		st.probe++
	}
	if !st.started {
		if !st.aiming {
			st.aiming, st.aimFrame = true, sample.frame
			return events, false
		}
		if sample.frame == st.aimFrame || sample.competing || !close {
			return events, false
		}
		events = append(events, &seat.KeyboardEvent{Key: st.key, Pressed: true})
		st.started, st.pressed, st.castFrame = true, true, sample.frame
		return events, false
	}
	if sample.frame-st.castFrame >= st.ticks {
		st.finished = true
	}
	return events, st.finished
}

func e2eReadMonsterQuickbarSample(typeID string, offset image.Point, selection *e2eNearestMonsterInputSelection) (e2eMonsterQuickbarSample, bool) {
	if noxServer == nil || noxClient == nil || noxClient.Viewport() == nil || noxClient.ClientPlayerUnit() == nil {
		return e2eMonsterQuickbarSample{}, false
	}
	unit := noxServer.Players.HostUnit()
	if unit == nil || !unit.Class().Has(object.ClassPlayer) || unit.HealthData == nil || unit.UpdateData == nil ||
		unit.HealthData.Cur == 0 || unit.Flags().HasAny(object.FlagDead|object.FlagDestroyed) {
		return e2eMonsterQuickbarSample{}, false
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
			available = target.distance <= 300 && noxServer.CanInteract(unit, npc, 0)
		}
		candidates = append(candidates, e2eNearestMonsterInputCandidate{
			identity: e2eMonsterInputIdentity{wire: uint16(code), extent: npc.Extent, scriptID: npc.ScriptIDVal},
			typeID:   typeID, target: target, hostile: noxServer.IsEnemyTo(unit, npc), available: available,
		})
	}
	target, available := selection.observe(typeID, candidates)
	sample := e2eMonsterQuickbarSample{target: target, frame: noxServer.Frame(), mouse: noxClient.Inp.GetMousePos()}
	sample.target.playerHP = unit.HealthData.Cur
	if !available || target.dead || target.absent {
		return sample, available
	}
	if dr := legacy.Nox_xxx_clientGetSpriteAtCursor_476F90(); dr != nil && !target.hovered {
		sample.competing = true
	}
	if cursor := unit.UpdateDataPlayer().CursorObj; cursor != nil {
		sample.competing = sample.competing || noxServer.GetUnitNetCode(cursor) != int(selection.identity.wire) ||
			cursor.Extent != selection.identity.extent || cursor.ScriptIDVal != selection.identity.scriptID
	}
	return sample, true
}

// CastAtNearestMonsterInput observes current viewport/sprite geometry, moves
// the real mouse and presses a normal quickbar key. It does not approach,
// bind/learn a spell, change cursor pointers or call a casting/target function.
// The original game chooses the target and decides whether the cast succeeds.
func (sc *e2eScenario) CastAtNearestMonsterInput(typeID string, offset image.Point, slot, count int, name string) {
	key, ticks, err := e2eMonsterQuickbarConfig(typeID, slot, count)
	if err != nil {
		e2eError(err)
		return
	}
	tracker := e2eMonsterQuickbarTracker{key: key, ticks: ticks}
	var selection e2eNearestMonsterInputSelection
	var last e2eMonsterQuickbarSample
	sc.addWhen(0, name+" actual mouse and quickbar tracking", 1200, func() bool {
		sample, available := e2eReadMonsterQuickbarSample(typeID, offset.Add(e2eMonsterQuickbarProbe(tracker.probe)), &selection)
		last = sample
		events, done := tracker.step(sample, available)
		e2eQueueInput(events...)
		return done
	}, func() {
		e2eLog.Printf("MONSTER QUICKBAR ACTUAL INPUT: type=%q wire=%#x extent=%d script_id=%d key=%d key_queued=%t point=%v distance=%g hovered=%t competing=%t observed_dead=%t observed_absent=%t player_hp=%d reads=%d frame=%d",
			typeID, selection.identity.wire, selection.identity.extent, selection.identity.scriptID, key,
			tracker.started, last.target.point, last.target.distance, last.target.hovered, last.competing,
			last.target.dead, last.target.absent, last.target.playerHP, tracker.reads, last.frame)
	})
}
