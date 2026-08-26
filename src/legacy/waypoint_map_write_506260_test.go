package legacy

import (
	"encoding/hex"
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
