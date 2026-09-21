package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
)

func moveWaypoint50D2A0(pos types.Pointf) *Waypoint {
	return &Waypoint{PosVec: pos, Flags: 1, Flags2: 0x80}
}

func connectMoveWaypoints50D2A0(from *Waypoint, to ...*Waypoint) {
	from.PointsCnt = uint8(len(to))
	for i, waypoint := range to {
		from.Points[i].Waypoint = waypoint
	}
}

func TestMonsterBuildWaypointPath547F70PreservesFrontierOrder(t *testing.T) {
	start := moveWaypoint50D2A0(types.Ptf(0, 0))
	left := moveWaypoint50D2A0(types.Ptf(10, 0))
	right := moveWaypoint50D2A0(types.Ptf(0, 10))
	end := moveWaypoint50D2A0(types.Ptf(10, 10))
	connectMoveWaypoints50D2A0(start, left, right)
	connectMoveWaypoints50D2A0(left, end)
	connectMoveWaypoints50D2A0(right, end)

	var out [16]*Waypoint
	count, status := monsterBuildWaypointPath547F70(start, end, out[:], 0x80)
	if status != monsterWaypointPathOK547F70 || count != 3 {
		t.Fatalf("path result = %d/%d, want 3/success", count, status)
	}
	want := []*Waypoint{start, right, end}
	for i := range want {
		if out[i] != want[i] {
			t.Fatalf("path[%d] = %p, want %p", i, out[i], want[i])
		}
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(out[1])) <= uintptr(math.MaxUint32) {
		t.Fatalf("native waypoint pointer = %#x, want value above PE32 range", uintptr(unsafe.Pointer(out[1])))
	}
}

func TestMonsterBuildWaypointPath547F70CapsOriginalOverflow(t *testing.T) {
	const nodes = 18
	waypoints := make([]Waypoint, nodes)
	for i := range waypoints {
		waypoints[i].Flags = 1
		waypoints[i].Flags2 = 0x80
		if i != 0 {
			connectMoveWaypoints50D2A0(&waypoints[i-1], &waypoints[i])
		}
	}
	guard := moveWaypoint50D2A0(types.Ptf(-1, -1))
	bounded := struct {
		out   [16]*Waypoint
		guard *Waypoint
	}{guard: guard}
	count, status := monsterBuildWaypointPath547F70(&waypoints[0], &waypoints[nodes-1], bounded.out[:], 0x80)
	if count != len(bounded.out) || status != monsterWaypointPathTooLong547F70 {
		t.Fatalf("long path result = %d/%d, want %d/too-long", count, status, len(bounded.out))
	}
	if bounded.guard != guard {
		t.Fatalf("output overflow changed guard to %p", bounded.guard)
	}
	for i := range bounded.out {
		if bounded.out[i] != &waypoints[i] {
			t.Fatalf("path[%d] = %p, want %p", i, bounded.out[i], &waypoints[i])
		}
	}
}

func TestMonsterBuildWaypointPath547F70RejectsInvalidAndDisconnected(t *testing.T) {
	start := moveWaypoint50D2A0(types.Pointf{})
	end := moveWaypoint50D2A0(types.Ptf(10, 0))
	var out [16]*Waypoint
	if count, status := monsterBuildWaypointPath547F70(start, end, out[:], 0x80); count != 0 || status != monsterWaypointPathNotFound547F70 {
		t.Fatalf("disconnected path = %d/%d", count, status)
	}
	end.Flags2 = 0x40
	connectMoveWaypoints50D2A0(start, end)
	if count, status := monsterBuildWaypointPath547F70(start, end, out[:], 0x80); count != 0 || status != monsterWaypointPathNotFound547F70 {
		t.Fatalf("masked path = %d/%d", count, status)
	}
}

func TestMonsterBuildMoveWaypointPath50D2A0StoresRawTargetAndPointers(t *testing.T) {
	unit := moveToMonsterTestObject5443F0(t)
	unit.PosVec = types.Ptf(12, 34)
	update := unit.UpdateDataMonster()
	update.Field91 = 9
	start := moveWaypoint50D2A0(unit.PosVec)
	end := moveWaypoint50D2A0(types.Ptf(100, 200))
	connectMoveWaypoints50D2A0(start, end)
	target := types.Pointf{
		X: math.Float32frombits(0x80000000),
		Y: math.Float32frombits(0x7fc12345),
	}
	var calls []types.Pointf
	find := func(got *Object, pos *types.Pointf) *Waypoint {
		if got != unit {
			t.Fatalf("find unit = %p, want %p", got, unit)
		}
		calls = append(calls, *pos)
		if len(calls) == 1 {
			return start
		}
		return end
	}
	count, status, statusSet := monsterBuildMoveWaypointPath50D2A0(unit, target, find)
	if count != 2 || status != monsterWaypointPathOK547F70 || !statusSet {
		t.Fatalf("waypoint result = %d/%d/%t, want 2/success/set", count, status, statusSet)
	}
	if len(calls) != 2 || calls[0] != unit.PosVec || math.Float32bits(calls[1].X) != math.Float32bits(target.X) || math.Float32bits(calls[1].Y) != math.Float32bits(target.Y) {
		t.Fatalf("find calls = %#v", calls)
	}
	if update.Field92 != math.Float32bits(target.X) || update.Field93 != math.Float32bits(target.Y) {
		t.Fatalf("target bits = %#x/%#x", update.Field92, update.Field93)
	}
	if update.Field74 != 2 || update.Field91 != 0 || update.Waypoints[0] != start || update.Waypoints[1] != end {
		t.Fatalf("route state = count:%d index:%d path:%p/%p", update.Field74, update.Field91, update.Waypoints[0], update.Waypoints[1])
	}
}

func TestMonsterBuildMoveWaypointPath50D2A0PreservesStatusWithoutSearch(t *testing.T) {
	unit := moveToMonsterTestObject5443F0(t)
	count, status, statusSet := monsterBuildMoveWaypointPath50D2A0(unit, types.Ptf(100, 200), func(*Object, *types.Pointf) *Waypoint {
		return nil
	})
	if count != 0 || status != monsterWaypointPathOK547F70 || statusSet {
		t.Fatalf("waypoint result = %d/%d/%t, want 0/success/unset", count, status, statusSet)
	}
}

func TestMonsterAdvanceWaypointPath50D2E0(t *testing.T) {
	unit := moveToMonsterTestObject5443F0(t)
	unit.PosVec = types.Ptf(0, 0)
	update := unit.UpdateDataMonster()
	first := moveWaypoint50D2A0(types.Ptf(1, 1))
	second := moveWaypoint50D2A0(types.Ptf(40, 50))
	update.Waypoints[0], update.Waypoints[1] = first, second
	update.Field74 = 2
	update.Field2 = 0
	setCalls, moveCalls := 0, 0
	hooks := monsterMovePathHooks50D5A0{
		setDetailedPath: func(got *Object, pos *types.Pointf) {
			setCalls++
			if got != unit || pos != &second.PosVec {
				t.Fatalf("detailed target = %p/%p, want %p/%p", got, pos, unit, &second.PosVec)
			}
			update.Field2 = 1
			update.Field71 = 0
		},
		actuallyMove: func(got *Object) bool {
			moveCalls++
			return false
		},
	}
	if monsterAdvanceWaypointPath50D2E0(unit, hooks) {
		t.Fatal("unfinished route reported completion")
	}
	if update.Field91 != 1 || setCalls != 1 || moveCalls != 1 || update.Path[1] != second.PosVec {
		t.Fatalf("advance state = index:%d set:%d move:%d append:%v", update.Field91, setCalls, moveCalls, update.Path[1])
	}

	unit.PosVec = second.PosVec
	update.Field2 = 0
	if !monsterAdvanceWaypointPath50D2E0(unit, hooks) {
		t.Fatal("last waypoint did not complete route")
	}
	if update.Field74 != 0 || setCalls != 1 || moveCalls != 1 {
		t.Fatalf("completed state = count:%d set:%d move:%d", update.Field74, setCalls, moveCalls)
	}
}

func TestMonsterAdvanceWaypointPath50D2E0RejectsDamagedNativeState(t *testing.T) {
	unit := moveToMonsterTestObject5443F0(t)
	update := unit.UpdateDataMonster()
	update.Field74 = uint32(len(update.Waypoints) + 1)
	if monsterAdvanceWaypointPath50D2E0(unit, monsterMovePathHooks50D5A0{}) {
		t.Fatal("damaged route reported completion")
	}
	if update.Field74 != 0 {
		t.Fatalf("damaged route count = %d, want cleared", update.Field74)
	}
}

func TestMonsterAdvanceWaypointPath50D2E0PreservesUnorderedBranch(t *testing.T) {
	unit := moveToMonsterTestObject5443F0(t)
	unit.PosVec = types.Ptf(float32(math.NaN()), 0)
	update := unit.UpdateDataMonster()
	first := moveWaypoint50D2A0(types.Ptf(10, 0))
	second := moveWaypoint50D2A0(types.Ptf(20, 0))
	update.Waypoints[0], update.Waypoints[1] = first, second
	update.Field74 = 2

	var detailedTarget *types.Pointf
	monsterAdvanceWaypointPath50D2E0(unit, monsterMovePathHooks50D5A0{
		setDetailedPath: func(_ *Object, target *types.Pointf) {
			detailedTarget = target
		},
	})
	if update.Field91 != 1 || detailedTarget != &second.PosVec {
		t.Fatalf("unordered branch = index:%d target:%p, want 1/%p", update.Field91, detailedTarget, &second.PosVec)
	}
}

func TestMonsterCreatureSetMovePath50D5A0UsesBlockedWaypointRoute(t *testing.T) {
	unit := moveToMonsterTestObject5443F0(t)
	unit.PosVec = types.Ptf(10, 20)
	unit.ObjFlags |= 0x4000
	update := unit.UpdateDataMonster()
	start := moveWaypoint50D2A0(unit.PosVec)
	end := moveWaypoint50D2A0(types.Ptf(300, 400))
	connectMoveWaypoints50D2A0(start, end)
	target := end.PosVec
	findCalls, detailedCalls, moveCalls := 0, 0, 0
	hooks := monsterMovePathHooks50D5A0{
		frame: func() uint32 { return 100 },
		trace: func(from, to types.Pointf, flags MapTraceFlags) bool {
			if from != unit.PosVec || to != target || flags != 4 {
				t.Fatalf("trace = %v -> %v flags %d", from, to, flags)
			}
			return false
		},
		findWaypoint: func(got *Object, pos *types.Pointf) *Waypoint {
			findCalls++
			if got != unit {
				t.Fatalf("find unit = %p, want %p", got, unit)
			}
			if findCalls == 1 {
				return start
			}
			return end
		},
		setDetailedPath: func(got *Object, pos *types.Pointf) {
			detailedCalls++
			if got != unit || pos != &end.PosVec {
				t.Fatalf("detailed target = %p/%p, want %p/%p", got, pos, unit, &end.PosVec)
			}
			update.Field2 = 1
		},
		actuallyMove: func(*Object) bool {
			moveCalls++
			return false
		},
	}
	if monsterCreatureSetMovePath50D5A0(unit, target, hooks) {
		t.Fatal("new blocked route reported completion")
	}
	if findCalls != 2 || detailedCalls != 1 || moveCalls != 1 || update.Field74 != 2 || update.Field91 != 1 {
		t.Fatalf("blocked route = find:%d detailed:%d move:%d count:%d index:%d", findCalls, detailedCalls, moveCalls, update.Field74, update.Field91)
	}

	unit.PosVec = end.PosVec
	update.Field2 = 0
	if !monsterCreatureSetMovePath50D5A0(unit, target, hooks) {
		t.Fatal("target arrival did not complete")
	}
	if update.Field74 != 2 {
		// The eight-unit arrival test returns before mutating the coarse route,
		// exactly as 0050D5A0 does.
		t.Fatalf("arrival mutated coarse route count to %d", update.Field74)
	}
}

func TestMonsterCreatureSetMovePath50D5A0WaypointFailureFallsBack(t *testing.T) {
	unit := moveToMonsterTestObject5443F0(t)
	unit.PosVec = types.Ptf(0, 0)
	target := types.Ptf(200, 0)
	update := unit.UpdateDataMonster()
	detailedCalls := 0
	hooks := monsterMovePathHooks50D5A0{
		frame: func() uint32 { return 50 },
		trace: func(types.Pointf, types.Pointf, MapTraceFlags) bool { return false },
		findWaypoint: func(*Object, *types.Pointf) *Waypoint {
			return nil
		},
		setDetailedPath: func(got *Object, pos *types.Pointf) {
			detailedCalls++
			if got != unit || *pos != target {
				t.Fatalf("fallback target = %p/%v", got, *pos)
			}
			update.Field2 = 1
		},
		actuallyMove: func(*Object) bool { return false },
	}
	if monsterCreatureSetMovePath50D5A0(unit, target, hooks) {
		t.Fatal("fallback path reported completion")
	}
	if detailedCalls != 1 || update.Field74 != 0 || update.Field92 != math.Float32bits(target.X) || update.Field93 != math.Float32bits(target.Y) {
		t.Fatalf("fallback state = calls:%d count:%d target:%#x/%#x", detailedCalls, update.Field74, update.Field92, update.Field93)
	}
}

func TestMonsterCreatureSetMovePath50D5A0ReportsWaypointFailure(t *testing.T) {
	unit := moveToMonsterTestObject5443F0(t)
	unit.PosVec = types.Ptf(0, 0)
	target := types.Ptf(200, 0)
	start := moveWaypoint50D2A0(unit.PosVec)
	end := moveWaypoint50D2A0(target)
	findCalls := 0
	statusCalls := 0
	status := monsterWaypointPathOK547F70
	hooks := monsterMovePathHooks50D5A0{
		frame: func() uint32 { return 50 },
		trace: func(types.Pointf, types.Pointf, MapTraceFlags) bool { return false },
		findWaypoint: func(*Object, *types.Pointf) *Waypoint {
			findCalls++
			if findCalls == 1 {
				return start
			}
			return end
		},
		setPathStatus: func(got monsterWaypointPathStatus547F70) {
			statusCalls++
			status = got
		},
		setDetailedPath: func(*Object, *types.Pointf) {},
	}
	if monsterCreatureSetMovePath50D5A0(unit, target, hooks) {
		t.Fatal("disconnected waypoint route reported completion")
	}
	if findCalls != 2 || statusCalls != 1 || status != monsterWaypointPathNotFound547F70 {
		t.Fatalf("waypoint status = find:%d calls:%d value:%d", findCalls, statusCalls, status)
	}
}

func TestMonsterCreatureSetMovePath50D5A0UsesOriginalDistanceBias(t *testing.T) {
	unit := moveToMonsterTestObject5443F0(t)
	unit.PosVec = types.Ptf(0, 0)
	target := types.Ptf(7.995, 0)
	detailedCalls := 0
	hooks := monsterMovePathHooks50D5A0{
		frame: func() uint32 { return 1 },
		trace: func(types.Pointf, types.Pointf, MapTraceFlags) bool { return true },
		setDetailedPath: func(*Object, *types.Pointf) {
			detailedCalls++
		},
	}
	if monsterCreatureSetMovePath50D5A0(unit, target, hooks) {
		t.Fatal("7.995-unit target incorrectly passed the biased eight-unit arrival boundary")
	}
	if detailedCalls != 1 {
		t.Fatalf("detailed path calls = %d, want 1", detailedCalls)
	}
}

func TestMonsterCreatureSetMovePath50D5A0PreservesUnorderedArrival(t *testing.T) {
	unit := moveToMonsterTestObject5443F0(t)
	unit.PosVec = types.Ptf(float32(math.NaN()), 0)
	traceCalls := 0
	hooks := monsterMovePathHooks50D5A0{
		frame: func() uint32 { return 1 },
		trace: func(types.Pointf, types.Pointf, MapTraceFlags) bool {
			traceCalls++
			return true
		},
	}
	if !monsterCreatureSetMovePath50D5A0(unit, types.Ptf(100, 0), hooks) {
		t.Fatal("unordered x87 distance did not follow the original arrival branch")
	}
	if traceCalls != 0 {
		t.Fatalf("arrival branch performed %d ray traces", traceCalls)
	}
}

func TestMonsterCreatureSetMovePath50D5A0DirectRefreshAndCompletion(t *testing.T) {
	unit := moveToMonsterTestObject5443F0(t)
	unit.PosVec = types.Ptf(0, 0)
	target := types.Ptf(100, 0)
	update := unit.UpdateDataMonster()
	update.Field2 = 1
	update.Field68 = types.Ptf(0, 0)
	update.Field70 = math.MaxUint32 - 5
	detailedCalls, moveCalls := 0, 0
	hooks := monsterMovePathHooks50D5A0{
		frame: func() uint32 { return 6 },
		trace: func(from, to types.Pointf, flags MapTraceFlags) bool {
			if from != unit.PosVec || to != target || flags != 0 {
				t.Fatalf("trace = %v -> %v flags %d", from, to, flags)
			}
			return true
		},
		setDetailedPath: func(*Object, *types.Pointf) {
			detailedCalls++
			update.Field2 = 1
		},
		actuallyMove: func(got *Object) bool {
			moveCalls++
			return got == unit
		},
	}
	if !monsterCreatureSetMovePath50D5A0(unit, target, hooks) {
		t.Fatal("completed detailed path was not reported")
	}
	if detailedCalls != 1 || moveCalls != 1 || update.Field2 != 0 {
		t.Fatalf("direct state = detailed:%d move:%d count:%d", detailedCalls, moveCalls, update.Field2)
	}
}

func TestMonsterCreatureActuallyMove50D3B0RunningVelocity(t *testing.T) {
	unit := moveToMonsterTestObject5443F0(t)
	unit.PosVec = types.Ptf(0, 0)
	unit.SpeedCur = 2
	update := unit.UpdateDataMonster()
	update.Field2 = 2
	update.Field67 = 0
	update.Path[0] = types.Ptf(0, 0)
	update.Path[1] = types.Ptf(30, 40)
	update.StatusFlags = 0x4000
	update.MonsterDef = &MonsterDef{RunMultiplier96: 1.5}
	traceCalls := 0
	if monsterCreatureActuallyMove50D3B0(unit, func(from, to types.Pointf, flags MapTraceFlags) bool {
		traceCalls++
		if from != unit.PosVec || flags != 132 {
			t.Fatalf("trace = %v -> %v flags %d", from, to, flags)
		}
		return true
	}) {
		t.Fatal("unfinished path reported completion")
	}
	distance := float32(50 + monsterMoveDistanceBias50D3B0)
	wantX := float32(float64(3) * 30 / float64(distance))
	wantY := float32(float64(3) * 40 / float64(distance))
	if traceCalls != 2 || update.Field67 != 1 || unit.ForceVec.X != wantX || unit.ForceVec.Y != wantY {
		t.Fatalf("movement = traces:%d index:%d force:%v, want force {%g %g}", traceCalls, update.Field67, unit.ForceVec, wantX, wantY)
	}
}
