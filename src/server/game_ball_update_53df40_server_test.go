package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func defaultGameBallUpdateNativeDeps53DF40() gameBallUpdateNativeDeps53DF40 {
	return gameBallUpdateNativeDeps53DF40{
		ticks:     func() uint64 { return 1000 },
		resetBall: func(*Object) {},
		carrierState: func(ball, carrier *Object) {
			(*GameBallUpdateData4EA800)(ball.UpdateData).Carrier = carrier
		},
		changeTeam: func(*Object) {},
		ballStatus: func(uint8, uint16) {},
		move:       func(*Object, types.Pointf) {},
		applyForce: func(*Object, types.Pointf, float32) {},
		clearOwner: func(ball *Object) { ball.ObjOwner = nil },
		trace:      func(types.Pointf, types.Pointf, MapTraceFlags) bool { return false },
		randomInt:  func(int32, int32) int32 { return 0 },
		audio:      func(uint32, *Object) {},
		frame:      func() uint32 { return 100 },
	}
}

func TestGameBallUpdateNative53DF40DeadCarrierThenTimeout(t *testing.T) {
	carrier := &Object{ObjFlags: object.Flags(0x12340020)}
	update := &GameBallUpdateData4EA800{Carrier: carrier, Ticks: 9}
	ball := &Object{UpdateData: unsafe.Pointer(update)}
	deps := defaultGameBallUpdateNativeDeps53DF40()
	events := make([]string, 0, 4)
	deps.ticks = func() uint64 { return 20010 }
	deps.carrierState = func(gotBall, gotCarrier *Object) {
		events = append(events, "carrier")
		if gotBall != ball || gotCarrier != nil {
			t.Fatalf("carrier state = (%p,%p)", gotBall, gotCarrier)
		}
		update.Carrier = nil
	}
	deps.changeTeam = func(got *Object) {
		events = append(events, "team")
		if got != ball {
			t.Fatalf("team ball = %p, want %p", got, ball)
		}
	}
	deps.ballStatus = func(state uint8, netCode uint16) {
		events = append(events, "status")
		if state != 1 || netCode != 0 {
			t.Fatalf("status = %d/%d", state, netCode)
		}
	}
	deps.resetBall = func(got *Object) {
		events = append(events, "reset")
		if got != ball || update.Carrier != nil {
			t.Fatalf("reset ball/carrier = (%p,%p)", got, update.Carrier)
		}
	}

	gameBallUpdateNative53DF40(ball, deps)
	if bits := math.Float32bits(ball.Float28); bits != gameBallUpdateDragBits53DF40 {
		t.Fatalf("drag bits = %#x, want %#x", bits, gameBallUpdateDragBits53DF40)
	}
	if !reflect.DeepEqual(events, []string{"carrier", "team", "status", "reset"}) {
		t.Fatalf("events = %#v", events)
	}
}

func TestGameBallUpdateNative53DF40FollowsNativeOwner(t *testing.T) {
	owner := &Object{PosVec: types.Pointf{X: 30, Y: -20}, Direction1: 32}
	owner.Shape.Circle.R = 2
	update := &GameBallUpdateData4EA800{
		Carrier:            owner,
		Ticks:              900,
		CarrierFrame:       95,
		PossessionDuration: 10,
	}
	ball := &Object{
		ObjFlags:   object.Flags(0x100),
		ObjOwner:   owner,
		UpdateData: unsafe.Pointer(update),
	}
	ball.Shape.Circle.R = 3
	deps := defaultGameBallUpdateNativeDeps53DF40()
	ticks := []uint64{1000, 1001}
	deps.ticks = func() uint64 {
		value := ticks[0]
		ticks = ticks[1:]
		return value
	}
	deps.frame = func() uint32 { return 100 }
	var tracedFrom, tracedTo, movedTo types.Pointf
	deps.trace = func(from, to types.Pointf, flags MapTraceFlags) bool {
		tracedFrom, tracedTo = from, to
		if flags != 5 {
			t.Fatalf("trace flags = %d", flags)
		}
		return true
	}
	deps.move = func(got *Object, destination types.Pointf) {
		if got != ball {
			t.Fatalf("move ball = %p, want %p", got, ball)
		}
		movedTo = destination
	}

	gameBallUpdateNative53DF40(ball, deps)
	cosine, sine := SinCosDir(32)
	wantTo := types.Pointf{
		X: float32(float64(15)*float64(cosine) + 30),
		Y: float32(float64(15)*float64(sine) - 20),
	}
	if tracedFrom != owner.PosVec || tracedTo != wantTo || movedTo != wantTo {
		t.Fatalf("trace/move = (%+v,%+v,%+v), want (%+v,%+v,%+v)",
			tracedFrom, tracedTo, movedTo, owner.PosVec, wantTo, wantTo)
	}
	if update.Ticks != 1001 || ball.ObjFlags != object.Flags(0x140) {
		t.Fatalf("ticks/flags = %d/%#x, want 1001/0x140", update.Ticks, ball.ObjFlags)
	}
	if len(ticks) != 0 {
		t.Fatalf("unused tick values = %#v", ticks)
	}
}

func TestGameBallUpdateNative53DF40BlockedOwnerDropsBall(t *testing.T) {
	owner := &Object{ObjFlags: object.Flags(0x8000)}
	update := &GameBallUpdateData4EA800{Carrier: owner, Ticks: 1000, CarrierFrame: 100}
	ball := &Object{ObjOwner: owner, UpdateData: unsafe.Pointer(update)}
	deps := defaultGameBallUpdateNativeDeps53DF40()
	events := make([]string, 0, 4)
	deps.clearOwner = func(got *Object) {
		events = append(events, "clear")
		got.ObjOwner = nil
	}
	deps.carrierState = func(gotBall, carrier *Object) {
		events = append(events, "carrier")
		if gotBall != ball || carrier != nil {
			t.Fatalf("carrier args = (%p,%p)", gotBall, carrier)
		}
		update.Carrier = nil
	}
	deps.changeTeam = func(*Object) { events = append(events, "team") }
	deps.ballStatus = func(uint8, uint16) { events = append(events, "status") }

	gameBallUpdateNative53DF40(ball, deps)
	if ball.ObjOwner != nil || update.Carrier != nil {
		t.Fatalf("owner/carrier = (%p,%p), want nil", ball.ObjOwner, update.Carrier)
	}
	if !reflect.DeepEqual(events, []string{"clear", "carrier", "team", "status"}) {
		t.Fatalf("events = %#v", events)
	}
}

func TestGameBallUpdateNative53DF40ReleasesAfterPossessionWindow(t *testing.T) {
	owner := &Object{Direction1: Dir16(1)}
	marker := &Object{}
	update := &GameBallUpdateData4EA800{
		Carrier:            owner,
		Ticks:              1000,
		CarrierFrame:       10,
		PossessionDuration: 20,
	}
	ball := &Object{
		ObjFlags:   object.Flags(0x140),
		PosVec:     types.Pointf{X: 50, Y: 75},
		ObjOwner:   owner,
		Obj130:     marker,
		UpdateData: unsafe.Pointer(update),
	}
	deps := defaultGameBallUpdateNativeDeps53DF40()
	deps.frame = func() uint32 { return 31 }
	deps.randomInt = func(minimum, maximum int32) int32 {
		if minimum != -32 || maximum != 32 {
			t.Fatalf("random range = %d..%d", minimum, maximum)
		}
		return -32
	}
	var forceOrigin types.Pointf
	deps.applyForce = func(got *Object, origin types.Pointf, force float32) {
		if got != ball || force != 30 {
			t.Fatalf("force args = (%p,%v)", got, force)
		}
		forceOrigin = origin
	}
	statuses, sounds := 0, 0
	deps.ballStatus = func(state uint8, netCode uint16) {
		statuses++
		if state != 1 || netCode != 0 {
			t.Fatalf("status = %d/%d", state, netCode)
		}
	}
	deps.audio = func(id uint32, got *Object) {
		sounds++
		if id != 926 || got != ball {
			t.Fatalf("audio = %d/%p", id, got)
		}
	}

	gameBallUpdateNative53DF40(ball, deps)
	cosine, sine := SinCosDir(225)
	wantOrigin := types.Pointf{
		X: float32(50 - float64(cosine)*20),
		Y: float32(75 - float64(sine)*20),
	}
	if forceOrigin != wantOrigin {
		t.Fatalf("force origin = %+v, want %+v", forceOrigin, wantOrigin)
	}
	if ball.ObjOwner != nil || ball.Obj130 != nil || ball.ObjFlags != object.Flags(0x100) {
		t.Fatalf("released ball = owner:%p obj130:%p flags:%#x", ball.ObjOwner, ball.Obj130, ball.ObjFlags)
	}
	if statuses != 1 || sounds != 1 {
		t.Fatalf("status/audio calls = %d/%d", statuses, sounds)
	}
}

func TestGameBallUpdateNative53DF40FreeBallClearsCarrierBelowThreshold(t *testing.T) {
	carrier := &Object{}
	update := &GameBallUpdateData4EA800{
		Carrier:       carrier,
		Ticks:         1000,
		ResetVelocity: 5,
	}
	ball := &Object{
		ObjFlags:   object.Flags(0x140),
		VelVec:     types.Pointf{X: 3, Y: 4},
		UpdateData: unsafe.Pointer(update),
	}
	deps := defaultGameBallUpdateNativeDeps53DF40()

	gameBallUpdateNative53DF40(ball, deps)
	if update.Carrier != carrier {
		t.Fatal("equal reset velocity cleared carrier")
	}
	if ball.ObjFlags != object.Flags(0x100) {
		t.Fatalf("flags = %#x, want 0x100", ball.ObjFlags)
	}

	update.ResetVelocity = math.Nextafter32(5, 6)
	gameBallUpdateNative53DF40(ball, deps)
	if update.Carrier != nil {
		t.Fatalf("carrier = %p, want nil", update.Carrier)
	}
}
