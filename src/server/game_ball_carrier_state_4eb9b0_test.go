package server

import (
	"fmt"
	"reflect"
	"testing"
)

type gameBallCarrierStateTestData4EB9B0 struct {
	name    string
	carrier *gameBallCarrierStateTestObject4EB9B0
	teamID  uint32
	frame   uint32
}

type gameBallCarrierStateTestObject4EB9B0 struct {
	name  string
	class uint8
	team  uint8
	data  *gameBallCarrierStateTestData4EB9B0
}

type gameBallCarrierStateTestFixture4EB9B0 struct {
	events         []string
	found          *gameBallCarrierStateTestObject4EB9B0
	frame          uint32
	onFind         func()
	onStoreCarrier func()
	onStoreTeam    func()
}

func gameBallCarrierStateObjectName4EB9B0(obj *gameBallCarrierStateTestObject4EB9B0) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func gameBallCarrierStateDataName4EB9B0(data *gameBallCarrierStateTestData4EB9B0) string {
	if data == nil {
		return "nil"
	}
	return data.name
}

func (f *gameBallCarrierStateTestFixture4EB9B0) event(format string, args ...any) {
	f.events = append(f.events, fmt.Sprintf(format, args...))
}

func (f *gameBallCarrierStateTestFixture4EB9B0) hooks() gameBallCarrierStateHooks4EB9B0[
	*gameBallCarrierStateTestObject4EB9B0,
	*gameBallCarrierStateTestData4EB9B0,
] {
	return gameBallCarrierStateHooks4EB9B0[
		*gameBallCarrierStateTestObject4EB9B0,
		*gameBallCarrierStateTestData4EB9B0,
	]{
		loadUpdateData: func(obj *gameBallCarrierStateTestObject4EB9B0) *gameBallCarrierStateTestData4EB9B0 {
			if obj == nil {
				f.event("update:nil")
				panic("nil ball UpdateData")
			}
			data := obj.data
			f.event("update:%s=%s", obj.name, gameBallCarrierStateDataName4EB9B0(data))
			return data
		},
		findPlayer: func(obj *gameBallCarrierStateTestObject4EB9B0) *gameBallCarrierStateTestObject4EB9B0 {
			f.event("find:%s=%s", gameBallCarrierStateObjectName4EB9B0(obj), gameBallCarrierStateObjectName4EB9B0(f.found))
			if f.onFind != nil {
				f.onFind()
			}
			return f.found
		},
		loadClassLow: func(obj *gameBallCarrierStateTestObject4EB9B0) uint8 {
			f.event("class:%s=%#x", gameBallCarrierStateObjectName4EB9B0(obj), obj.class)
			return obj.class
		},
		storeCarrier: func(data *gameBallCarrierStateTestData4EB9B0, carrier *gameBallCarrierStateTestObject4EB9B0) {
			f.event("carrier:%s=%s", gameBallCarrierStateDataName4EB9B0(data), gameBallCarrierStateObjectName4EB9B0(carrier))
			if data == nil {
				panic("nil GameBall UpdateData carrier")
			}
			data.carrier = carrier
			if f.onStoreCarrier != nil {
				f.onStoreCarrier()
			}
		},
		loadTeamID: func(obj *gameBallCarrierStateTestObject4EB9B0) uint8 {
			f.event("team:%s=%#x", gameBallCarrierStateObjectName4EB9B0(obj), obj.team)
			return obj.team
		},
		storeTeamID: func(data *gameBallCarrierStateTestData4EB9B0, teamID uint32) {
			f.event("store-team:%s=%#x", gameBallCarrierStateDataName4EB9B0(data), teamID)
			if data == nil {
				panic("nil GameBall UpdateData team")
			}
			data.teamID = teamID
			if f.onStoreTeam != nil {
				f.onStoreTeam()
			}
		},
		loadFrame: func() uint32 {
			f.event("frame=%#x", f.frame)
			return f.frame
		},
		storeFrame: func(data *gameBallCarrierStateTestData4EB9B0, frame uint32) {
			f.event("store-frame:%s=%#x", gameBallCarrierStateDataName4EB9B0(data), frame)
			if data == nil {
				panic("nil GameBall UpdateData frame")
			}
			data.frame = frame
		},
	}
}

func assertGameBallCarrierStateEvents4EB9B0(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("event order mismatch\n got: %#v\nwant: %#v", got, want)
	}
}

func TestGameBallCarrierState4EB9B0NilTargetCachesUpdateAndRetainsFrame(t *testing.T) {
	data := &gameBallCarrierStateTestData4EB9B0{
		name: "entry", carrier: &gameBallCarrierStateTestObject4EB9B0{name: "old"},
		teamID: 0xaabbccdd, frame: 0x11223344,
	}
	ball := &gameBallCarrierStateTestObject4EB9B0{name: "ball", data: data}
	f := &gameBallCarrierStateTestFixture4EB9B0{}
	got := gameBallCarrierState4EB9B0(ball, nil, f.hooks())
	if got != nil {
		t.Fatalf("result = %s, want nil", gameBallCarrierStateObjectName4EB9B0(got))
	}
	assertGameBallCarrierStateEvents4EB9B0(t, f.events, []string{
		"update:ball=entry", "carrier:entry=nil", "store-team:entry=0x0",
	})
	if data.carrier != nil || data.teamID != 0 || data.frame != 0x11223344 {
		t.Fatalf("cleared state = carrier %s, team %#x, frame %#x", gameBallCarrierStateObjectName4EB9B0(data.carrier), data.teamID, data.frame)
	}
}

func TestGameBallCarrierState4EB9B0LookupFailuresPreserveExactResult(t *testing.T) {
	for _, tc := range []struct {
		name      string
		found     *gameBallCarrierStateTestObject4EB9B0
		wantEvent []string
	}{
		{
			name: "nil result",
			wantEvent: []string{
				"update:ball=data", "find:target=nil", "carrier:data=nil", "store-team:data=0x0",
			},
		},
		{
			name:  "non-player terminal",
			found: &gameBallCarrierStateTestObject4EB9B0{name: "terminal", class: 0x80},
			wantEvent: []string{
				"update:ball=data", "find:target=terminal", "class:terminal=0x80", "carrier:data=nil", "store-team:data=0x0",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			oldCarrier := &gameBallCarrierStateTestObject4EB9B0{name: "old"}
			data := &gameBallCarrierStateTestData4EB9B0{name: "data", carrier: oldCarrier, teamID: 9, frame: 77}
			ball := &gameBallCarrierStateTestObject4EB9B0{name: "ball", data: data}
			target := &gameBallCarrierStateTestObject4EB9B0{name: "target"}
			f := &gameBallCarrierStateTestFixture4EB9B0{found: tc.found, frame: 123}
			got := gameBallCarrierState4EB9B0(ball, target, f.hooks())
			if got != tc.found {
				t.Fatalf("result = %s, want %s", gameBallCarrierStateObjectName4EB9B0(got), gameBallCarrierStateObjectName4EB9B0(tc.found))
			}
			assertGameBallCarrierStateEvents4EB9B0(t, f.events, tc.wantEvent)
			if data.carrier != nil || data.teamID != 0 || data.frame != 77 {
				t.Fatalf("failure state = carrier %s, team %#x, frame %#x", gameBallCarrierStateObjectName4EB9B0(data.carrier), data.teamID, data.frame)
			}
		})
	}
}

func TestGameBallCarrierState4EB9B0SuccessUsesCachedDataAndLiveValues(t *testing.T) {
	entry := &gameBallCarrierStateTestData4EB9B0{name: "entry", frame: 7}
	replacement := &gameBallCarrierStateTestData4EB9B0{name: "replacement", frame: 8}
	ball := &gameBallCarrierStateTestObject4EB9B0{name: "ball", data: entry}
	target := &gameBallCarrierStateTestObject4EB9B0{name: "target"}
	carrier := &gameBallCarrierStateTestObject4EB9B0{name: "carrier", class: 0x84, team: 1}
	f := &gameBallCarrierStateTestFixture4EB9B0{found: carrier, frame: 0x11111111}
	f.onFind = func() {
		ball.data = replacement
	}
	f.onStoreCarrier = func() {
		carrier.team = 0xab
	}
	f.onStoreTeam = func() {
		f.frame = 0x89abcdef
	}

	got := gameBallCarrierState4EB9B0(ball, target, f.hooks())
	if got != carrier {
		t.Fatalf("result = %s, want carrier", gameBallCarrierStateObjectName4EB9B0(got))
	}
	assertGameBallCarrierStateEvents4EB9B0(t, f.events, []string{
		"update:ball=entry",
		"find:target=carrier",
		"class:carrier=0x84",
		"carrier:entry=carrier",
		"team:carrier=0xab",
		"store-team:entry=0xab",
		"frame=0x89abcdef",
		"store-frame:entry=0x89abcdef",
	})
	if entry.carrier != carrier || entry.teamID != 0xab || entry.frame != 0x89abcdef {
		t.Fatalf("entry state = carrier %s, team %#x, frame %#x", gameBallCarrierStateObjectName4EB9B0(entry.carrier), entry.teamID, entry.frame)
	}
	if replacement.carrier != nil || replacement.teamID != 0 || replacement.frame != 8 {
		t.Fatalf("replacement data changed: %+v", *replacement)
	}
}

func TestGameBallCarrierState4EB9B0NilUpdateFaultIsDelayedUntilStore(t *testing.T) {
	ball := &gameBallCarrierStateTestObject4EB9B0{name: "ball"}
	target := &gameBallCarrierStateTestObject4EB9B0{name: "target"}
	carrier := &gameBallCarrierStateTestObject4EB9B0{name: "carrier", class: 4, team: 3}
	f := &gameBallCarrierStateTestFixture4EB9B0{found: carrier, frame: 9}
	defer func() {
		if recover() == nil {
			t.Fatal("nil update-data did not fault")
		}
		assertGameBallCarrierStateEvents4EB9B0(t, f.events, []string{
			"update:ball=nil", "find:target=carrier", "class:carrier=0x4", "carrier:nil=carrier",
		})
	}()
	gameBallCarrierState4EB9B0(ball, target, f.hooks())
}
