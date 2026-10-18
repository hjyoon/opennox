package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

func TestMapgenWaypointLifetime316TransferredCAllocation(t *testing.T) {
	if noxflags.HasGame(1) {
		t.Fatal("fixture requires the isolated non-host test state")
	}
	var s server.Server
	first := mapgenWaypointRaw316()
	second := mapgenWaypointRaw316()
	if first == nil || second == nil {
		t.Fatal("native generated-waypoint allocator returned nil")
	}
	first.Index, second.Index = 101, 102
	first.PosVec = types.Ptf(1079, 1000)
	second.PosVec = types.Ptf(1109, 1000)
	first.PointsCnt, second.PointsCnt = 1, 1
	first.Field348[0], second.Field348[0] = 102, 101
	first.Points[0].Ind, second.Points[0].Ind = 2, 2
	s.Sub_579E90(first)
	s.Sub_579E90(second)
	s.WPs.Sub_579D20()
	if first.Points[0].Waypoint != second || second.Points[0].Waypoint != first {
		t.Fatal("native generated-waypoint links did not resolve")
	}
	s.Nox_xxx_waypoint_5799C0()
	if s.WPs.Pending != nil || s.WPs.ByInd(int(first.Index)) != first || s.WPs.ByInd(int(second.Index)) != second {
		t.Fatal("generated waypoint publication changed identities")
	}
	if first.PosVec != types.Ptf(1079, 1000) || second.PosVec != types.Ptf(1109, 1000) {
		t.Fatal("generated waypoint publication changed positions")
	}
	s.Nox_xxx_waypointDeleteAll_579DD0()
	if s.WPs.List != nil || s.WPs.Pending != nil || s.WPs.CntXxx != 0 {
		t.Fatal("generated waypoint teardown left a list")
	}
}

func TestMapgenWaypointLifetime316MixedAllocators(t *testing.T) {
	if noxflags.HasGame(1) {
		t.Fatal("fixture requires the isolated non-host test state")
	}
	var s server.Server
	ordinary := s.WPs.Nox_xxx_waypointNewNotMap_579970(41, types.Ptf(1, 2))
	generated := mapgenWaypointRaw316()
	if generated == nil {
		t.Fatal("native generated-waypoint allocator returned nil")
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(generated)) <= 0xffffffff {
		t.Fatal("native generated waypoint must retain its address above 4GiB")
	}
	generated.Index = 42
	s.Sub_579E90(generated)
	s.Nox_xxx_waypoint_5799C0()
	if s.WPs.ByInd(41) != ordinary || s.WPs.ByInd(42) != generated {
		t.Fatal("mixed waypoint identities changed")
	}
	s.Nox_xxx_waypointDeleteAll_579DD0()
	// A second empty teardown must not retain or release either old pointer.
	s.Nox_xxx_waypointDeleteAll_579DD0()
}
