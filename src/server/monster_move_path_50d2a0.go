package server

import (
	"math"

	"github.com/opennox/libs/types"
)

type monsterWaypointPathStatus547F70 uint8

const (
	monsterWaypointPathOK547F70 monsterWaypointPathStatus547F70 = iota
	monsterWaypointPathTooLong547F70
	monsterWaypointPathNotFound547F70
	monsterMoveDistanceBias50D3B0 = 0.009999999776482582
)

// monsterBuildWaypointPath547F70 is the native-width counterpart of
// GAME.EXE 00547F70. The original stores the BFS parent and frontier links in
// two PE32 pointer fields at waypoint offsets 508 and 512. Keeping those links
// in Go maps/slices preserves full pointers on LP64 without changing the map
// waypoint layout.
//
// Each breadth layer is processed in the original order: newly discovered
// nodes are prepended to the next frontier. The PE32 routine also writes a
// seventeenth pointer before noticing that its nominal 16-entry output is
// full. This implementation intentionally preserves the returned prefix and
// status, but never writes beyond out.
func monsterBuildWaypointPath547F70(start, end *Waypoint, out []*Waypoint, mask byte) (int, monsterWaypointPathStatus547F70) {
	if !monsterWaypointValid547EE0(start, mask) || !monsterWaypointValid547EE0(end, mask) {
		return 0, monsterWaypointPathNotFound547F70
	}

	parents := map[*Waypoint]*Waypoint{start: nil}
	frontier := []*Waypoint{start}
	for len(frontier) != 0 {
		discovered := make([]*Waypoint, 0)
		for _, waypoint := range frontier {
			if waypoint == end {
				path := make([]*Waypoint, 0, len(parents))
				for current := waypoint; current != nil; current = parents[current] {
					path = append(path, current)
				}
				for left, right := 0, len(path)-1; left < right; left, right = left+1, right-1 {
					path[left], path[right] = path[right], path[left]
				}
				count := min(len(path), len(out))
				copy(out[:count], path[:count])
				if count != len(path) {
					return count, monsterWaypointPathTooLong547F70
				}
				return count, monsterWaypointPathOK547F70
			}

			count := int(waypoint.PointsCnt)
			if count > len(waypoint.Points) {
				count = len(waypoint.Points)
			}
			for i := 0; i < count; i++ {
				next := waypoint.Points[i].Waypoint
				if _, visited := parents[next]; visited || !monsterWaypointValid547EE0(next, mask) {
					continue
				}
				parents[next] = waypoint
				discovered = append(discovered, next)
			}
		}
		frontier = frontier[:0]
		for i := len(discovered) - 1; i >= 0; i-- {
			frontier = append(frontier, discovered[i])
		}
	}
	return 0, monsterWaypointPathNotFound547F70
}

// monsterBuildMoveWaypointPath50D2A0 preserves the target snapshot and
// waypoint cursor written by GAME.EXE 0050D2A0. findWaypoint corresponds to
// 0050CB20 and is invoked source-first, target-second just like 00547F20.
func monsterBuildMoveWaypointPath50D2A0(unit *Object, target types.Pointf, findWaypoint func(*Object, *types.Pointf) *Waypoint) (int, monsterWaypointPathStatus547F70, bool) {
	if unit == nil || unit.UpdateData == nil {
		return 0, monsterWaypointPathOK547F70, false
	}
	update := unit.UpdateDataMonster()
	update.Field92 = math.Float32bits(target.X)
	update.Field93 = math.Float32bits(target.Y)
	count := 0
	status := monsterWaypointPathOK547F70
	statusSet := false
	if findWaypoint != nil {
		start := findWaypoint(unit, &unit.PosVec)
		end := findWaypoint(unit, &target)
		if start != nil && end != nil && start != end {
			count, status = monsterBuildWaypointPath547F70(start, end, update.Waypoints[:], 0x80)
			statusSet = true
		}
	}
	update.Field74 = uint32(count)
	update.Field91 = 0
	return count, status, statusSet
}

type monsterMovePathHooks50D5A0 struct {
	frame           func() uint32
	trace           func(types.Pointf, types.Pointf, MapTraceFlags) bool
	findWaypoint    func(*Object, *types.Pointf) *Waypoint
	setPathStatus   func(monsterWaypointPathStatus547F70)
	setDetailedPath func(*Object, *types.Pointf)
	actuallyMove    func(*Object) bool
}

// monsterAdvanceWaypointPath50D2E0 advances the coarse waypoint route and
// feeds its current node into the existing detailed pathfinder. Invalid
// native state is cleared rather than indexing a damaged PE32 count/pointer.
func monsterAdvanceWaypointPath50D2E0(unit *Object, hooks monsterMovePathHooks50D5A0) bool {
	if unit == nil || unit.UpdateData == nil {
		return false
	}
	update := unit.UpdateDataMonster()
	if update.Field74 != 0 && update.Field2 == 0 {
		count := int(update.Field74)
		index := int(update.Field91)
		if count > len(update.Waypoints) || index < 0 || index >= count {
			update.Field74 = 0
			return false
		}
		waypoint := update.Waypoints[index]
		if waypoint == nil {
			update.Field74 = 0
			return false
		}
		dx := float64(waypoint.PosVec.X) - float64(unit.PosVec.X)
		dy := float64(waypoint.PosVec.Y) - float64(unit.PosVec.Y)
		distance2 := dx*dx + dy*dy
		// 0050D329 tests only x87 C0 after FCOMP. Unordered values set C0,
		// so a NaN follows the same branch as a waypoint inside eight units.
		if distance2 < 64.0 || math.IsNaN(distance2) {
			if index == count-1 {
				update.Field74 = 0
				return true
			}
			index++
			update.Field91 = uint32(index)
			waypoint = update.Waypoints[index]
			if waypoint == nil {
				update.Field74 = 0
				return false
			}
		}
		if hooks.setDetailedPath != nil {
			hooks.setDetailedPath(unit, &waypoint.PosVec)
		}
		if byte(update.Field71) == 0 {
			pathIndex := int(update.Field2)
			if pathIndex >= 0 && pathIndex < len(update.Path) {
				update.Path[pathIndex] = waypoint.PosVec
			}
		}
	}
	return hooks.actuallyMove != nil && hooks.actuallyMove(unit) && byte(update.Field71) == 2
}

// monsterCreatureSetMovePath50D5A0 restores GAME.EXE 0050D5A0, including
// its blocked-ray waypoint route. The previous adapter skipped 0050D2A0 and
// 0050D2E0 entirely, causing monsters to give up or repeatedly rebuild a
// detailed path when a direct line to the action target was obstructed.
func monsterCreatureSetMovePath50D5A0(unit *Object, target types.Pointf, hooks monsterMovePathHooks50D5A0) bool {
	if unit == nil || unit.UpdateData == nil || hooks.frame == nil || hooks.trace == nil {
		return false
	}
	update := unit.UpdateDataMonster()
	dx := float64(target.X) - float64(unit.PosVec.X)
	dy := float64(target.Y) - float64(unit.PosVec.Y)
	distance := math.Sqrt(dx*dx + dy*dy)
	// 0050D5F3 tests C0|C3 after comparing against eight. The x87
	// unordered result therefore shares the original arrival branch.
	if math.IsNaN(distance) || distance+monsterMoveDistanceBias50D3B0 <= 8.0 {
		return true
	}

	traceFlags := MapTraceFlags((uint32(unit.ObjFlags) >> 12) & 4)
	direct := hooks.trace(unit.PosVec, target, traceFlags)
	pathCount := update.Field2
	frame := hooks.frame()
	if !direct {
		if pathCount == 0 {
			lastWaypointTarget := types.Pointf{
				X: math.Float32frombits(update.Field92),
				Y: math.Float32frombits(update.Field93),
			}
			waypointDX := float64(lastWaypointTarget.X) - float64(target.X)
			waypointDY := float64(lastWaypointTarget.Y) - float64(target.Y)
			if update.Field74 == 0 || frame-update.Field70 > 10 &&
				waypointDX*waypointDX+waypointDY*waypointDY > 10000.0 {
				_, status, statusSet := monsterBuildMoveWaypointPath50D2A0(unit, target, hooks.findWaypoint)
				if statusSet && hooks.setPathStatus != nil {
					hooks.setPathStatus(status)
				}
			}
			if update.Field74 != 0 {
				if monsterAdvanceWaypointPath50D2E0(unit, hooks) {
					update.Field2 = 0
					update.Field74 = 0
					return true
				}
				return false
			}
			if hooks.setDetailedPath != nil {
				hooks.setDetailedPath(unit, &target)
			}
		}
	} else {
		lastDX := float64(update.Field68.X) - float64(target.X)
		lastDY := float64(update.Field68.Y) - float64(target.Y)
		if pathCount == 0 || frame-update.Field70 > 10 &&
			lastDX*lastDX+lastDY*lastDY > 2500.0 {
			if hooks.setDetailedPath != nil {
				hooks.setDetailedPath(unit, &target)
			}
		}
	}

	if update.Field74 != 0 {
		if monsterAdvanceWaypointPath50D2E0(unit, hooks) {
			update.Field2 = 0
			update.Field74 = 0
			return true
		}
		return false
	}
	if update.Field2 != 0 && hooks.actuallyMove != nil && hooks.actuallyMove(unit) {
		update.Field2 = 0
		return true
	}
	return false
}
