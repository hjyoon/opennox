package legacy

import (
	"testing"

	"github.com/opennox/opennox/v1/server"
)

func TestUnitWasDamagedRecently4E6BD0HealthGate(t *testing.T) {
	frameCalls := 0
	frame := func() uint32 {
		frameCalls++
		return 0
	}
	if got := unitWasDamagedRecently_4E6BD0(&server.Object{}, frame); got {
		t.Fatal("unit without health data reported recent damage")
	}
	if frameCalls != 0 {
		t.Fatalf("frame calls = %d, want 0", frameCalls)
	}
}

func TestUnitWasDamagedRecently4E6BD0FrameBoundaries(t *testing.T) {
	tests := []struct {
		name    string
		frame   uint32
		damaged uint32
		want    bool
	}{
		{name: "same frame", frame: 100, damaged: 100, want: true},
		{name: "one frame", frame: 100, damaged: 99, want: true},
		{name: "two frames", frame: 100, damaged: 98},
		{name: "wrap by one", frame: 0, damaged: ^uint32(0), want: true},
		{name: "future frame wraps far", frame: 0, damaged: 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			unit := &server.Object{HealthData: &server.HealthData{}, Frame134: tc.damaged}
			calls := 0
			got := unitWasDamagedRecently_4E6BD0(unit, func() uint32 {
				calls++
				return tc.frame
			})
			if got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
			if calls != 1 {
				t.Fatalf("frame calls = %d, want 1", calls)
			}
		})
	}
}

func TestUnitWasDamagedRecently4E6BD0ReadsTimestampAfterFrame(t *testing.T) {
	health := &server.HealthData{}
	unit := &server.Object{HealthData: health, Frame134: 200}
	got := unitWasDamagedRecently_4E6BD0(unit, func() uint32 {
		unit.HealthData = nil
		unit.Frame134 = 99
		return 100
	})
	if !got {
		t.Fatal("timestamp mutation made by frame callback was not observed")
	}
}

func TestUnitWasDamagedRecently4E6BD0NilUnitPanicsBeforeFrame(t *testing.T) {
	frameCalls := 0
	defer func() {
		if recover() == nil {
			t.Fatal("nil unit did not panic")
		}
		if frameCalls != 0 {
			t.Fatalf("frame calls = %d, want 0", frameCalls)
		}
	}()
	unitWasDamagedRecently_4E6BD0(nil, func() uint32 {
		frameCalls++
		return 0
	})
}
