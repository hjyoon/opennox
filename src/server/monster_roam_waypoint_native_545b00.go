package server

import "unsafe"

// MonsterRoamWaypointRuntime545B00 supplies the original world services.
// Raw entry methods keep nil-record faults and only request a service when
// the original reaches it. In particular, constructing this value is lazy.
type MonsterRoamWaypointRuntime545B00 struct {
	Random   func(int, int) int
	UnitName func(*Object) string
	Frame    func() uint32
	Debug    func(string, uint32, string, int32)
	Pop      func(*Object) int
}

func (waypoint *Waypoint) RoamValid547EE0(mask byte) bool {
	return waypoint != nil && waypoint.Flags&1 != 0 && waypoint.Flags2&mask != 0
}

func (update *MonsterUpdateData) RoamRecordWaypoint545B00(waypoint *Waypoint) {
	update.Field91++
	if int32(update.Field91) >= int32(len(update.Waypoints)) {
		update.Field91 = 0
	}
	index := int(update.Field91)
	update.Waypoints[index] = waypoint
	for i := range update.Waypoints {
		if i != index && update.Waypoints[i] == waypoint {
			update.Waypoints[i] = nil
		}
	}
}

func (update *MonsterUpdateData) RoamPreviousWaypoint545B60(mask byte) *Waypoint {
	for distance := uint32(1); distance < uint32(len(update.Waypoints)); distance++ {
		index := int32(update.Field91 - distance)
		if index < 0 {
			index += int32(len(update.Waypoints))
		}
		waypoint := update.Waypoints[index]
		if waypoint.RoamValid547EE0(mask) {
			return waypoint
		}
	}
	return nil
}

func (update *MonsterUpdateData) RoamNextWaypoint545C60(current *Waypoint, mask byte, runtime MonsterRoamWaypointRuntime545B00) *Waypoint {
	return monsterRoamNextWaypointCount545C60(update, current, mask, runtime.Random, int(current.PointsCnt))
}

func monsterRoamNextWaypointCount545C60(update *MonsterUpdateData, current *Waypoint, mask byte, random func(int, int) int, count int) *Waypoint {
	var choices [32]*Waypoint
	n := 0
	for i := 0; i < count; i++ {
		candidate := current.Points[i].Waypoint
		if !candidate.RoamValid547EE0(mask) {
			continue
		}
		seen := false
		for _, waypoint := range update.Waypoints {
			if candidate == waypoint {
				seen = true
				break
			}
		}
		if !seen {
			// Duplicate links are separate original RNG choices.
			choices[n] = candidate
			n++
		}
	}
	if n != 0 {
		return choices[random(0, n-1)]
	}
	for distance := uint32(0); distance < uint32(len(update.Waypoints)); distance++ {
		index := int32(update.Field91 + distance + 1)
		if index >= int32(len(update.Waypoints)) {
			index -= int32(len(update.Waypoints))
		}
		candidate := update.Waypoints[index]
		if candidate == nil {
			continue
		}
		for i := 0; i < count; i++ {
			if current.Points[i].Waypoint == candidate {
				// The original checks adjacency before dereferencing flags.
				if candidate.RoamValid547EE0(mask) {
					return candidate
				}
				break
			}
		}
	}
	return nil
}

func monsterRoamDeadEndWaypoint545BB0(update *MonsterUpdateData, head *AIStackItem, current *Waypoint, mask byte, random func(int, int) int) bool {
	update.Field2 = 0
	if current.PointsCnt != 0 {
		next := update.RoamNextWaypoint545C60(current, mask, MonsterRoamWaypointRuntime545B00{Random: random})
		head.Args[0] = uintptr(unsafe.Pointer(next))
		if current.PointsCnt != 0 && head.Args[0] != 0 {
			update.RoamRecordWaypoint545B00((*Waypoint)(unsafe.Pointer(head.Args[0])))
			return true
		}
	}
	return false
}

func (unit *Object) RoamDeadEnd545BB0(current *Waypoint, runtime MonsterRoamWaypointRuntime545B00) bool {
	update := (*MonsterUpdateData)(unit.UpdateData)
	head := &update.AIStack[update.AIStackInd]
	mask := byte(head.Args[2])
	if monsterRoamDeadEndWaypoint545BB0(update, head, current, mask, runtime.Random) {
		return true
	}
	// The code word is captured before name lookup; the frame is read after it.
	code := unit.NetCode
	name := runtime.UnitName(unit)
	frame := runtime.Frame()
	runtime.Debug("%d: %s(#%d) Reached dead end, giving up.\n", frame, name, int32(code))
	runtime.Pop(unit)
	return false
}
