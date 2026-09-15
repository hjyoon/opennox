package legacy

import (
	"encoding/binary"
	"encoding/hex"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

func TestWaypointNext579870(t *testing.T) {
	second := &server.Waypoint{Index: 22}
	first := &server.Waypoint{Index: 11, WpNext: second}
	if got := waypointNext579870(first); got != second {
		t.Fatalf("next = %p, want %p", got, second)
	}
	if got := waypointNext579870(nil); got != nil {
		t.Fatalf("nil next = %p, want nil", got)
	}
}

func readWaypointFixture506260(t *testing.T, data []byte, hooks mapWaypointReadHooks506260) error {
	t.Helper()
	path := filepath.Join(t.TempDir(), "waypoints-read.bin")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	cf, err := cryptfile.OpenFile(path, cryptfile.ReadOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer cf.Close()
	return mapReadWaypoints506260(cf, hooks)
}

func TestMapReadWaypoints506260Version4(t *testing.T) {
	data, err := hex.DecodeString(
		"0400" +
			"01000000" +
			"443322110000c03f000010c005737461727403000000011600000007",
	)
	if err != nil {
		t.Fatal(err)
	}
	var allocated []*server.Waypoint
	err = readWaypointFixture506260(t, data, mapWaypointReadHooks506260{
		adjust: func(pos types.Pointf) types.Pointf {
			return types.Ptf(pos.X+10, pos.Y-20)
		},
		newWaypoint: func(index int, pos types.Pointf) *server.Waypoint {
			wp := &server.Waypoint{Index: uint32(index), PosVec: pos}
			allocated = append(allocated, wp)
			return wp
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(allocated) != 1 {
		t.Fatalf("allocated %d waypoints, want 1", len(allocated))
	}
	wp := allocated[0]
	if wp.Index != 0x11223344 || wp.PosVec != types.Ptf(11.5, -22.25) || wp.Name() != "start" ||
		wp.Flags != 3 || wp.PointsCnt != 1 || wp.Field348[0] != 22 || wp.Points[0].Ind != 7 {
		t.Fatalf("decoded waypoint = %+v name=%q field348=%d point=%+v", wp, wp.Name(), wp.Field348[0], wp.Points[0])
	}
}

func TestMapReadWaypoints506260LegacyVersion1(t *testing.T) {
	data := make([]byte, 2+4+4+4+4+4+4+4)
	binary.LittleEndian.PutUint16(data[0:2], 1)
	binary.LittleEndian.PutUint32(data[2:6], 1)
	binary.LittleEndian.PutUint32(data[6:10], 7)
	binary.LittleEndian.PutUint32(data[10:14], 123)
	binary.LittleEndian.PutUint32(data[14:18], 456)
	binary.LittleEndian.PutUint32(data[18:22], 1)
	binary.LittleEndian.PutUint32(data[22:26], 1)
	binary.LittleEndian.PutUint32(data[26:30], 99)
	var wp *server.Waypoint
	err := readWaypointFixture506260(t, data, mapWaypointReadHooks506260{
		newWaypoint: func(index int, pos types.Pointf) *server.Waypoint {
			wp = &server.Waypoint{Index: uint32(index), PosVec: pos}
			return wp
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if wp == nil || wp.Index != 7 || wp.PosVec != types.Ptf(123, 456) || wp.Name() != "" ||
		wp.Flags != 1 || wp.PointsCnt != 1 || wp.Field348[0] != 99 || wp.Points[0].Ind != 2 {
		t.Fatalf("decoded legacy waypoint = %+v", wp)
	}
}

func TestMapReadWaypoints506260RejectsOversizedConnections(t *testing.T) {
	data := make([]byte, 2+4+4+4+4+1+4+1)
	binary.LittleEndian.PutUint16(data[0:2], 4)
	binary.LittleEndian.PutUint32(data[2:6], 1)
	binary.LittleEndian.PutUint32(data[6:10], 7)
	binary.LittleEndian.PutUint32(data[10:14], math.Float32bits(1))
	binary.LittleEndian.PutUint32(data[14:18], math.Float32bits(2))
	// Byte 18 is the empty name, bytes 19..22 are flags, and byte 23 is count.
	data[23] = 33
	err := readWaypointFixture506260(t, data, mapWaypointReadHooks506260{
		newWaypoint: func(index int, pos types.Pointf) *server.Waypoint {
			return &server.Waypoint{Index: uint32(index), PosVec: pos}
		},
	})
	if err == nil {
		t.Fatal("oversized waypoint connection list was accepted")
	}
}

func TestMapWriteWaypoints506260(t *testing.T) {
	second := &server.Waypoint{Index: 22, PosVec: types.Ptf(3.25, 4.5)}
	first := &server.Waypoint{
		Index:     11,
		PosVec:    types.Ptf(1.5, -2.25),
		PointsCnt: 1,
		Flags:     3,
		WpNext:    second,
	}
	first.SetName("start")
	first.Points[0] = server.WaypointSub{Waypoint: second, Ind: 7}

	path := filepath.Join(t.TempDir(), "waypoints.bin")
	cf, err := cryptfile.OpenFile(path, cryptfile.WriteOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	if err := mapWriteWaypoints506260(cf, first, func(*server.Waypoint) bool { return true }); err != nil {
		t.Fatal(err)
	}
	if err := cf.Close(); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want, err := hex.DecodeString(
		"0400" +
			"02000000" +
			"0b0000000000c03f000010c005737461727401000000011600000007" +
			"160000000000504000009040000000000000",
	)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatalf("serialized waypoints = %x, want %x", got, want)
	}
}
