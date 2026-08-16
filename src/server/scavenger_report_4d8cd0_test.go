package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type scavengerReportTestPlayer4D8CD0 struct {
	name             string
	count, maximum   uint32
	afterCountLoaded func()
}

type scavengerReportTestUpdate4D8CD0 struct {
	name   string
	player *scavengerReportTestPlayer4D8CD0
}

type scavengerReportTestObject4D8CD0 struct {
	name     string
	classLow uint8
	update   *scavengerReportTestUpdate4D8CD0
}

type scavengerReportTestWorld4D8CD0 struct {
	owner, ownerArg *scavengerReportTestObject4D8CD0
	unitCode        uint32
	sendResult      int32
	events          []string
	faultAt         int
	afterUnitCode   func(*scavengerReportTestWorld4D8CD0)
	packet          [7]byte
}

func newScavengerReportTestWorld4D8CD0() *scavengerReportTestWorld4D8CD0 {
	player := &scavengerReportTestPlayer4D8CD0{
		name: "player-a", count: 0x12345678, maximum: 0x789abcde,
	}
	update := &scavengerReportTestUpdate4D8CD0{name: "update-a", player: player}
	owner := &scavengerReportTestObject4D8CD0{name: "owner-a", classLow: 4, update: update}
	return &scavengerReportTestWorld4D8CD0{
		owner: owner, ownerArg: owner, unitCode: 0xfeed1234, sendResult: -17,
	}
}

func (w *scavengerReportTestWorld4D8CD0) event(name string) {
	w.events = append(w.events, name)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(name)
	}
}

func scavengerReportObjectName4D8CD0(obj *scavengerReportTestObject4D8CD0) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func scavengerReportPlayerName4D8CD0(player *scavengerReportTestPlayer4D8CD0) string {
	if player == nil {
		return "nil"
	}
	return player.name
}

func (w *scavengerReportTestWorld4D8CD0) hooks() scavengerReportHooks4D8CD0[
	*scavengerReportTestObject4D8CD0,
	*scavengerReportTestUpdate4D8CD0,
	*scavengerReportTestPlayer4D8CD0,
] {
	return scavengerReportHooks4D8CD0[
		*scavengerReportTestObject4D8CD0,
		*scavengerReportTestUpdate4D8CD0,
		*scavengerReportTestPlayer4D8CD0,
	]{
		loadOwnerArg: func() *scavengerReportTestObject4D8CD0 {
			w.event("owner-arg:" + scavengerReportObjectName4D8CD0(w.ownerArg))
			return w.ownerArg
		},
		loadClassLow: func(owner *scavengerReportTestObject4D8CD0) uint8 {
			w.event(fmt.Sprintf("class:%s:%#x", owner.name, owner.classLow))
			return owner.classLow
		},
		loadUpdate: func(owner *scavengerReportTestObject4D8CD0) *scavengerReportTestUpdate4D8CD0 {
			w.event("update:" + owner.name + ":" + owner.update.name)
			return owner.update
		},
		unitCode: func(owner *scavengerReportTestObject4D8CD0) uint32 {
			w.event(fmt.Sprintf("unit-code:%s:%#x", owner.name, w.unitCode))
			value := w.unitCode
			if w.afterUnitCode != nil {
				w.afterUnitCode(w)
			}
			return value
		},
		loadPlayer: func(update *scavengerReportTestUpdate4D8CD0) *scavengerReportTestPlayer4D8CD0 {
			w.event("player:" + update.name + ":" + scavengerReportPlayerName4D8CD0(update.player))
			return update.player
		},
		loadCount: func(player *scavengerReportTestPlayer4D8CD0) uint16 {
			w.event(fmt.Sprintf("count:%s:%#x", player.name, player.count))
			value := uint16(player.count)
			if player.afterCountLoaded != nil {
				player.afterCountLoaded()
			}
			return value
		},
		loadMaximum: func(player *scavengerReportTestPlayer4D8CD0) uint16 {
			w.event(fmt.Sprintf("maximum:%s:%#x", player.name, player.maximum))
			return uint16(player.maximum)
		},
		sendPacket: func(recipient int32, packet [7]byte, related *scavengerReportTestObject4D8CD0, remove int32) int32 {
			w.event(fmt.Sprintf("send:%d:%x:%s:%d", recipient, packet, scavengerReportObjectName4D8CD0(related), remove))
			w.packet = packet
			return w.sendResult
		},
	}
}

func scavengerReportSuccessEvents4D8CD0() []string {
	return []string{
		"owner-arg:owner-a", "class:owner-a:0x4", "update:owner-a:update-a",
		"unit-code:owner-a:0xfeed1234", "player:update-a:player-a",
		"count:player-a:0x12345678", "player:update-a:player-a",
		"maximum:player-a:0x789abcde", "send:255:5534127856debc:nil:1",
	}
}

func verifyScavengerReportFaultPrefixes4D8CD0(t *testing.T, want []string) {
	t.Helper()
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := newScavengerReportTestWorld4D8CD0()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if !reflect.DeepEqual(w.events, want[:faultAt]) {
					t.Fatalf("events = %#v, want %#v", w.events, want[:faultAt])
				}
			}()
			scavengerReport4D8CD0(w.hooks())
		})
	}
}

func TestScavengerReport4D8CD0ExactPacketTraceAndFaultPrefixes(t *testing.T) {
	w := newScavengerReportTestWorld4D8CD0()
	got := scavengerReport4D8CD0(w.hooks())
	if got.kind != scavengerReportSendResult4D8CD0 || got.send != -17 {
		t.Fatalf("result = %#v", got)
	}
	wantEvents := scavengerReportSuccessEvents4D8CD0()
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events = %#v, want %#v", w.events, wantEvents)
	}
	wantPacket := [7]byte{85, 0x34, 0x12, 0x78, 0x56, 0xde, 0xbc}
	if w.packet != wantPacket {
		t.Fatalf("packet = % x, want % x", w.packet, wantPacket)
	}
	verifyScavengerReportFaultPrefixes4D8CD0(t, wantEvents)
}

func TestScavengerReport4D8CD0NonPlayerStillLoadsUpdate(t *testing.T) {
	w := newScavengerReportTestWorld4D8CD0()
	w.owner.classLow = 0xf8
	got := scavengerReport4D8CD0(w.hooks())
	if got.kind != scavengerReportOwnerResult4D8CD0 || got.owner != w.owner {
		t.Fatalf("result = %#v, want owner %p", got, w.owner)
	}
	want := []string{
		"owner-arg:owner-a", "class:owner-a:0xf8", "update:owner-a:update-a",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %#v, want %#v", w.events, want)
	}
}

func TestScavengerReport4D8CD0CachesUpdateAndReloadsPlayer(t *testing.T) {
	w := newScavengerReportTestWorld4D8CD0()
	first := w.owner.update.player
	second := &scavengerReportTestPlayer4D8CD0{name: "player-b", maximum: 0xaabbccdd}
	replacement := &scavengerReportTestUpdate4D8CD0{
		name: "update-b", player: &scavengerReportTestPlayer4D8CD0{name: "unread", maximum: 1},
	}
	originalUpdate := &scavengerReportTestUpdate4D8CD0{name: "update-a", player: first}
	w.owner.update = originalUpdate
	w.afterUnitCode = func(w *scavengerReportTestWorld4D8CD0) {
		w.owner.update = replacement
		originalUpdate.player = first
	}
	first.afterCountLoaded = func() {
		originalUpdate.player = second
	}

	got := scavengerReport4D8CD0(w.hooks())
	if got.kind != scavengerReportSendResult4D8CD0 || got.send != -17 {
		t.Fatalf("result = %#v", got)
	}
	if !containsScavengerReportEvent4D8CD0(w.events, "player:update-a:player-b") ||
		!containsScavengerReportEvent4D8CD0(w.events, "maximum:player-b:0xaabbccdd") {
		t.Fatalf("events = %#v", w.events)
	}
	if containsScavengerReportEvent4D8CD0(w.events, "player:update-b:unread") {
		t.Fatalf("replacement update was used: %#v", w.events)
	}
}

func TestScavengerReport4D8CD0ForwardsFullSendResult(t *testing.T) {
	for _, result := range []int32{0, 1, -1, math.MinInt32, math.MaxInt32} {
		w := newScavengerReportTestWorld4D8CD0()
		w.sendResult = result
		got := scavengerReport4D8CD0(w.hooks())
		if got.kind != scavengerReportSendResult4D8CD0 || got.send != result {
			t.Fatalf("send result %d: got %#v", result, got)
		}
	}
}

func containsScavengerReportEvent4D8CD0(events []string, want string) bool {
	for _, event := range events {
		if event == want {
			return true
		}
	}
	return false
}
