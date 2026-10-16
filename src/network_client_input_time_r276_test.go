package opennox

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/opennox/libs/env"
	"github.com/opennox/libs/platform"
	"github.com/opennox/libs/player"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

// This clock replaces only the existing Platform owner, not any production
// control/queue function. Chosen millisecond values
// fit time.Duration in both ordinary and E2E clock representations.
type inputClockR276 struct {
	platform.Platform
	now   uint64
	reads int
}

func (p *inputClockR276) Ticks() time.Duration {
	p.reads++
	if env.IsE2E() {
		return time.Duration(p.now)
	}
	return time.Duration(p.now) * time.Millisecond
}

func inputClockFixtureR276(t *testing.T, host, online bool) (*Client, *inputClockR276) {
	t.Helper()
	inputHostModeR273(t, host)
	previousOnline := noxflags.HasGame(noxflags.GameOnline)
	previousPlatform := platform.Get()
	previousServer := noxServer
	info := clientPlayerInfo()
	previousInfo := *info
	settings := getServerSettings()
	previousMode, previousDelay := settings.LatencyCompensationA66, settings.LatencyCompensationB70
	previousSprite := legacy.Get_dword_5d4594_1096640()
	clock := &inputClockR276{Platform: previousPlatform, now: 1000}
	t.Cleanup(func() {
		legacy.Set_dword_5d4594_1096640(previousSprite)
		settings.LatencyCompensationA66, settings.LatencyCompensationB70 = previousMode, previousDelay
		*info = previousInfo
		noxServer = previousServer
		platform.Set(previousPlatform)
		if previousOnline {
			noxflags.SetGame(noxflags.GameOnline)
		} else {
			noxflags.UnsetGame(noxflags.GameOnline)
		}
	})
	if online {
		noxflags.SetGame(noxflags.GameOnline)
	} else {
		noxflags.UnsetGame(noxflags.GameOnline)
	}
	platform.Set(clock)
	noxServer = &Server{Server: &server.Server{}}
	info.PlayerClassVal = byte(player.Warrior)
	legacy.Set_dword_5d4594_1096640(nil)
	binary.LittleEndian.PutUint32(settings.LatencyCompensationA66[:], 3)
	binary.LittleEndian.PutUint32(settings.LatencyCompensationB70[:], 0)
	return inputClientFixtureR273(t), clock
}

func inputClockCompensationR276(raw uint32) {
	settings := getServerSettings()
	binary.LittleEndian.PutUint32(settings.LatencyCompensationA66[:], 3)
	binary.LittleEndian.PutUint32(settings.LatencyCompensationB70[:], raw)
}

// Call the actual 42D6B0 owner. A nil binding list skips binding actions, Warrior
// skips the spell-target callback, and a nil sprite avoids sprite transforms.
// The owner still reads the real platform seam, applies configured compensation,
// emits orientation through 42E670, and runs its unchanged history/local tails.
// Reading fields through uint64 keeps this draft compilable against uint32 too.
func inputClockOwnerR276(t *testing.T, c *Client, clock *inputClockR276, now, want uint64) noxCtrlEvent {
	t.Helper()
	*c.ctrl = CtrlEventHandler{}
	clock.now = now
	reads := clock.reads
	c.ctrl.nox_xxx_clientControl_42D6B0(videoGetWindowSize().Div(2), nil)
	if clock.reads != reads+1 {
		t.Fatalf("control owner clock reads=%d, want 1", clock.reads-reads)
	}
	wantHistory := 0
	if !noxflags.HasGame(noxflags.GameHost) {
		wantHistory = 1 // Non-host _C copies this orientation into its history.
	}
	if c.ctrl.indA != 1 || c.ctrl.indB != 0 || c.ctrl.indC != wantHistory || c.ctrl.indD != 0 || len(c.ctrl.netBuf) != 0 {
		t.Fatalf("orientation owner cursor/tail state: read=%d write=%d history=%d/%d net=%x", c.ctrl.indB, c.ctrl.indA, c.ctrl.indC, c.ctrl.indD, c.ctrl.netBuf)
	}
	event := c.ctrl.bufA[0]
	if event.PlayerCtrl != (server.PlayerCtrl{Code: player.CCOrientation, Data: 0, Active: true}) {
		t.Fatalf("actual owner orientation=%+v, want code=1 data=0 active=true", event.PlayerCtrl)
	}
	if uint64(c.ctrl.ticks) != want || uint64(event.tick) != want {
		t.Fatalf("actual owner timestamp now=%#x compensation=%#x: handler=%#x event=%#x, want=%#x", now,
			binary.LittleEndian.Uint32(getServerSettings().LatencyCompensationB70[:]), uint64(c.ctrl.ticks), uint64(event.tick), want)
	}
	return event
}

func inputClockPrepareR276(t *testing.T, c *Client, clock *inputClockR276, now uint64, wantCursor int, events []noxCtrlEvent) {
	t.Helper()
	clock.now = now
	before, reads := *c.ctrl, clock.reads
	c.ctrl.writeToNetBuffer()
	gotState, wantState := *c.ctrl, before
	wantState.indB = wantCursor
	gotState.netBuf, wantState.netBuf = nil, nil
	if !reflect.DeepEqual(gotState, wantState) {
		t.Fatalf("clock=%#x serializer read cursor=%d -> %d, want=%d, or unrelated state changed", now, before.indB, c.ctrl.indB, wantCursor)
	}
	if want := inputReferenceFramesR273(events); !bytes.Equal(c.ctrl.netBuf, want) {
		t.Fatalf("clock=%#x serializer wire=%x, want=%x", now, c.ctrl.netBuf, want)
	}
	if clock.reads != reads+1 {
		t.Fatalf("serializer clock reads=%d, want 1", clock.reads-reads)
	}
}

func TestClientInputClockOwnerHighWordR276(t *testing.T) {
	for _, mode := range []struct {
		name         string
		host, online bool
	}{
		{"host-online", true, true}, {"host-offline", true, false},
		{"client-online", false, true}, {"client-offline", false, false},
	} {
		t.Run(mode.name, func(t *testing.T) {
			c, clock := inputClockFixtureR276(t, mode.host, mode.online)
			for _, now := range []uint64{1<<32 + 100, 3<<32 + 7} {
				t.Run(fmt.Sprintf("ticks-%#x", now), func(t *testing.T) {
					inputClockOwnerR276(t, c, clock, now, now)
				})
			}
		})
	}
}

func TestClientInputClockOwnerPositiveCarryR276(t *testing.T) {
	for _, mode := range []struct {
		name         string
		host, online bool
	}{
		{"host-online", true, true}, {"host-offline", true, false},
		{"client-online", false, true}, {"client-offline", false, false},
	} {
		t.Run(mode.name, func(t *testing.T) {
			c, clock := inputClockFixtureR276(t, mode.host, mode.online)
			for _, spec := range []struct {
				name string
				now  uint64
				raw  uint32
			}{
				{"carry-hundred", math.MaxUint32 - 25, 100},
				{"maximum-positive", 1<<32 + 100, math.MaxInt32},
			} {
				t.Run(spec.name, func(t *testing.T) {
					inputClockCompensationR276(spec.raw)
					want := spec.now
					if mode.host && mode.online {
						want += uint64(spec.raw)
					}
					inputClockOwnerR276(t, c, clock, spec.now, want)
				})
			}
		})
	}
}

func TestClientInputClockOwnerSignedCompensationR276(t *testing.T) {
	c, clock := inputClockFixtureR276(t, true, true)
	for _, spec := range []struct {
		name string
		now  uint64
		raw  uint32
		want uint64
	}{
		// Low-word cases distinguish CDQ sign extension from zero extension even
		// when no high platform word is present. The helper ABI stays uint32.
		{"minus-one-low", 100, 0xffffffff, 99},
		{"minus-one-high", 1<<32 + 100, 0xffffffff, 1<<32 + 99},
		{"minus-hundred-high", 1<<32 + 100, 0xffffff9c, 1 << 32},
		{"minus-hundred-low-word-borrow", 1<<32 + 10, 0xffffff9c, 1<<32 - 90},
		{"minimum-signed-high", 1<<32 + 100, 0x80000000, 1<<31 + 100},
		{"minus-one-wraps-native-qword", 0, 0xffffffff, math.MaxUint64},
	} {
		t.Run(spec.name, func(t *testing.T) {
			inputClockCompensationR276(spec.raw)
			if got := nox_ctrlevent_add_ticks_42E630(); got != spec.raw {
				t.Fatalf("compensation helper ABI raw=%#x, want=%#x", got, spec.raw)
			}
			inputClockOwnerR276(t, c, clock, spec.now, spec.want)
		})
	}
}

func TestClientInputClockDeadlineCarryR276(t *testing.T) {
	c, clock := inputClockFixtureR276(t, true, false)
	for _, start := range []int{0, 126, 127} {
		for _, active := range []bool{false, true} {
			// These untyped timestamps fit both the exact uint32 baseline and the
			// intended uint64 fields; no reflection or alternate implementation.
			*c.ctrl = CtrlEventHandler{indB: start, indA: (start + 2) % ctrlEventCap}
			events := []noxCtrlEvent{
				{tick: 0xffffffe6, PlayerCtrl: server.PlayerCtrl{Code: player.CCMoveForward, Data: 0x71, Active: active}},
				{tick: math.MaxUint32, PlayerCtrl: server.PlayerCtrl{Code: player.CCAction, Data: 0x72, Active: true}},
			}
			for i, event := range events {
				c.ctrl.bufA[(start+i)%ctrlEventCap] = event
			}
			inputClockPrepareR276(t, c, clock, 0xffffffe6, c.ctrl.indA, events)
		}
	}
}

func TestClientInputClockAncientLowWordR276(t *testing.T) {
	c, clock := inputClockFixtureR276(t, true, false)
	for _, start := range []int{0, 127} {
		*c.ctrl = CtrlEventHandler{indB: start, indA: (start + 2) % ctrlEventCap}
		events := []noxCtrlEvent{
			{tick: 1000, PlayerCtrl: server.PlayerCtrl{Code: player.CCMoveForward, Data: 0x81, Active: true}},
			{tick: 1001, PlayerCtrl: server.PlayerCtrl{Code: player.CCAction, Data: 0x82, Active: true}},
		}
		for i, event := range events {
			c.ctrl.bufA[(start+i)%ctrlEventCap] = event
		}
		inputClockPrepareR276(t, c, clock, 1<<32+100, c.ctrl.indA, events)
	}
}

func TestClientInputClockActualOwnerCrossRolloverR276(t *testing.T) {
	c, clock := inputClockFixtureR276(t, true, true)
	// Both fields can represent this production-created event before rollover.
	// Only the later consumer observes the higher word, so its RED is not merely
	// a type/width assertion and does not need a high-valued event initializer.
	event := inputClockOwnerR276(t, c, clock, 1<<32-100, 1<<32-100)
	inputClockPrepareR276(t, c, clock, 1<<32+10, c.ctrl.indA, []noxCtrlEvent{event})
	inputClockPrepareR276(t, c, clock, 1<<32+10, c.ctrl.indA, nil)
}

func TestClientInputClockActualOwnerFutureGateR276(t *testing.T) {
	const now uint64 = 1<<32 + 100
	for _, spec := range []struct {
		name       string
		raw        uint32
		activeHead bool
		olderTail  bool
	}{
		{"equal-active", 50, true, false},
		{"future-active", 51, true, false},
		{"future-inactive", 51, false, false},
		{"equal-active-older-tail", 50, true, true},
		{"equal-inactive-older-tail", 50, false, true},
		{"future-active-older-tail", 51, true, true},
		{"future-inactive-older-tail", 51, false, true},
	} {
		t.Run(spec.name, func(t *testing.T) {
			c, clock := inputClockFixtureR276(t, true, true)
			inputClockCompensationR276(spec.raw)
			head := inputClockOwnerR276(t, c, clock, now, now+uint64(spec.raw))
			// Preserve the actual owner's timestamp; only deactivate the local
			// head when requested. The timestamp gate precedes this Active gate.
			head.Active = spec.activeHead
			c.ctrl.bufA[0].Active = spec.activeHead
			events := []noxCtrlEvent{head}
			if spec.olderTail {
				// A later-enqueued but older active record cannot bypass a future
				// head. Both records use wire-only orientation, with no UI action.
				tail := noxCtrlEvent{tick: 1000, PlayerCtrl: server.PlayerCtrl{
					Code: player.CCOrientation, Data: 0xa7, Active: true,
				}}
				c.ctrl.bufA[1] = tail
				c.ctrl.indA = 2
				events = append(events, tail)
			}
			end := c.ctrl.indA
			if spec.raw == 51 {
				inputClockPrepareR276(t, c, clock, now, 0, nil)
				inputClockPrepareR276(t, c, clock, now, 0, nil)
				// Advancing one millisecond makes the original head exactly equal
				// to the new now+50 deadline; its entire FIFO tail is now eligible.
				inputClockPrepareR276(t, c, clock, now+1, end, events)
				inputClockPrepareR276(t, c, clock, now+1, end, nil)
			} else {
				inputClockPrepareR276(t, c, clock, now, end, events)
				inputClockPrepareR276(t, c, clock, now, end, nil)
				inputClockPrepareR276(t, c, clock, now+1, end, nil)
			}
		})
	}
}
