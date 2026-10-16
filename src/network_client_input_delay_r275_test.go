package opennox

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"github.com/opennox/libs/env"
	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/platform"
	"github.com/opennox/libs/player"

	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/server"
)

// This unit-test platform overrides only Ticks. All other platform behavior
// remains delegated, and the original owner is restored before the next test.
// These are deterministic unit clocks, not native elapsed-delay observations.
type inputDelayClockR275 struct {
	platform.Platform
	now   uint32
	reads int
}

func (p *inputDelayClockR275) Ticks() time.Duration {
	p.reads++
	if env.IsE2E() {
		return time.Duration(p.now)
	}
	return time.Duration(p.now) * time.Millisecond
}

func inputDelayFixtureR275(t *testing.T, host bool) (*Client, *inputDelayClockR275) {
	t.Helper()
	inputHostModeR273(t, host)
	previousPlatform := platform.Get()
	info := clientPlayerInfo()
	previousInfo := *info
	clock := &inputDelayClockR275{Platform: previousPlatform, now: 1000}
	platform.Set(clock)
	info.PlayerClassVal = byte(player.Conjurer)
	t.Cleanup(func() {
		*info = previousInfo
		platform.Set(previousPlatform)
	})
	return inputClientFixtureR273(t), clock
}

func inputDelaySentinelRingR275(start, count int) CtrlEventHandler {
	ctrl := CtrlEventHandler{
		ticks: 1000, playerDir: 0x51, indA: (start + count) % ctrlEventCap,
		indB: start, indC: 23, indD: 17,
		flags747848: 0x91, flags750956: true, flags754064: 0xabcdef01,
		netBuf: []byte{0xee, 0xff},
	}
	for i := range ctrl.bufA {
		ctrl.bufA[i] = noxCtrlEvent{tick: 0x12345678, PlayerCtrl: server.PlayerCtrl{
			Code: player.CCJump, Data: 0xa5000000 + uint32(i), Active: true,
		}}
		ctrl.bufB[i] = ctrl.bufA[i]
	}
	return ctrl
}

func inputDelayPrepareR275(t *testing.T, c *Client, clock *inputDelayClockR275, wantCursor int, events []noxCtrlEvent) []byte {
	t.Helper()
	before, reads := *c.ctrl, clock.reads
	c.ctrl.writeToNetBuffer()
	wantState, gotState := before, *c.ctrl
	wantState.indB = wantCursor
	wantState.netBuf, gotState.netBuf = nil, nil
	if !reflect.DeepEqual(gotState, wantState) {
		t.Fatalf("producer changed state beyond consumed cursor: cursor %d -> %d, want %d", before.indB, c.ctrl.indB, wantCursor)
	}
	want := inputReferenceFramesR273(events)
	if !bytes.Equal(c.ctrl.netBuf, want) {
		t.Fatalf("prepared wire=%x, want=%x", c.ctrl.netBuf, want)
	}
	if clock.reads != reads+1 {
		t.Fatalf("producer clock reads=%d, want 1", clock.reads-reads)
	}
	return bytes.Clone(c.ctrl.netBuf)
}

func TestClientInputHostPartialDrainAllCursorsR275(t *testing.T) {
	c, clock := inputDelayFixtureR275(t, true)
	cases := 0
	for start := 0; start < ctrlEventCap; start++ {
		for _, count := range []int{0, 1, 2, 51, 64, 126, 127} {
			seen := make(map[int]bool)
			for _, cut := range []int{0, count / 2, max(0, count-1), count} {
				if seen[cut] {
					continue
				}
				seen[cut] = true
				for activity := 0; activity < 3; activity++ {
					clock.now = 1000
					*c.ctrl = inputDelaySentinelRingR275(start, count)
					events := make([]noxCtrlEvent, count)
					for i := range events {
						events[i] = noxCtrlEvent{tick: 1050, PlayerCtrl: server.PlayerCtrl{
							Code: []player.CtrlCode{29, 1, 6}[i%3], Data: 0x89abcd00 + uint32(i),
							Active: activity == 0 || activity == 1 && i%2 == 0,
						}}
						if i == cut {
							events[i].tick = 1051
						} else if i > cut {
							// Older ready entries after a future head cannot bypass it.
							events[i].tick = 999
						}
						c.ctrl.bufA[(start+i)%ctrlEventCap] = events[i]
					}
					cursor := (start + cut) % ctrlEventCap
					inputDelayPrepareR275(t, c, clock, cursor, events[:cut])
					inputDelayPrepareR275(t, c, clock, cursor, nil)
					clock.now = 1001 // The preserved 1051 head is now exactly due.
					inputDelayPrepareR275(t, c, clock, c.ctrl.indA, events[cut:])
					inputDelayPrepareR275(t, c, clock, c.ctrl.indA, nil)
					cases++
				}
			}
		}
	}
	if cases != 8448 {
		t.Fatalf("matrix case count=%d, want 8448", cases)
	}
}

func TestClientInputHostInactiveFutureGateAndAppendR275(t *testing.T) {
	c, clock := inputDelayFixtureR275(t, true)
	*c.ctrl = inputDelaySentinelRingR275(126, 4)
	events := inputRepeatEventsR273(4, player.CCOrientation)
	events[0].tick, events[1].tick, events[2].tick, events[3].tick = 1050, 1051, 1052, 999
	events[1].Active = false
	for i, event := range events {
		c.ctrl.bufA[(126+i)%ctrlEventCap] = event
	}
	inputDelayPrepareR275(t, c, clock, 127, events[:1])
	inputDelayPrepareR275(t, c, clock, 127, nil)
	before := *c.ctrl
	c.ctrl.nox_ctrlevent_action_42E670(player.CCMoveForward, 0x73)
	want := before
	want.bufA[2] = noxCtrlEvent{tick: 1000, PlayerCtrl: server.PlayerCtrl{Code: player.CCMoveForward, Data: 0x73, Active: true}}
	want.indA = 3
	if !reflect.DeepEqual(*c.ctrl, want) {
		t.Fatal("append after partial drain overwrote retained entries or changed unrelated state")
	}
	inputDelayPrepareR275(t, c, clock, 127, nil)
	clock.now = 1001
	// The now-due inactive head is consumed, but the next future entry still
	// blocks its older successor. No inactive-only MSG_PLAYER_INPUT is emitted.
	inputDelayPrepareR275(t, c, clock, 0, nil)
	inputDelayPrepareR275(t, c, clock, 0, nil)
	clock.now = 1002
	tail := append(append([]noxCtrlEvent(nil), events[2:]...), want.bufA[2])
	inputDelayPrepareR275(t, c, clock, 3, tail)
	inputDelayPrepareR275(t, c, clock, 3, nil)
}

func TestClientInputHostLongPrefixTailAndBackpressureR275(t *testing.T) {
	c, clock := inputDelayFixtureR275(t, true)
	*c.ctrl = inputDelaySentinelRingR275(125, 127)
	events := inputRepeatEventsR273(127, 29)
	for i := range events {
		events[i].tick = 1050
		if i == 64 {
			events[i].tick = 1051
		}
		c.ctrl.bufA[(125+i)%ctrlEventCap] = events[i]
	}
	cursor := (125 + 64) % ctrlEventCap
	prepared := inputDelayPrepareR275(t, c, clock, cursor, events[:64])
	if parts := inputFrameEventsR273(t, prepared); len(parts) != 3 || len(prepared) != 518 {
		t.Fatalf("long ready prefix must retain three complete bounded envelopes, got %d bytes", len(prepared))
	}
	filler := bytes.Repeat([]byte{byte(netmsg.MSG_KEEP_ALIVE)}, 2048)
	if !c.srv.NetList.AddToMsgListCli(server.HostPlayerIndex, netlist.Kind0, filler) {
		t.Fatal("could not fill existing legitimate message queue")
	}
	before := *c.ctrl
	list := c.srv.NetList.ByInd(server.HostPlayerIndex, netlist.Kind0)
	if c.clientSendInput(server.HostPlayerIndex) || list.Count() != 1 || list.Size() != len(filler) ||
		!reflect.DeepEqual(*c.ctrl, before) || !bytes.Equal(c.ctrl.netBuf, prepared) {
		t.Fatal("sender backpressure changed retained cursor/prepared prefix or queued a partial message")
	}
	if got := c.srv.NetList.CopyPacketsA(server.HostPlayerIndex, netlist.Kind0); !bytes.Equal(got, filler) {
		t.Fatal("failed send altered existing queue bytes")
	}
	if !c.clientSendInput(server.HostPlayerIndex) || !reflect.DeepEqual(*c.ctrl, before) ||
		!bytes.Equal(c.srv.NetList.CopyPacketsA(server.HostPlayerIndex, netlist.Kind0), prepared) {
		t.Fatal("explicit retry of unchanged prepared prefix failed after queue recovery")
	}
	inputDelayPrepareR275(t, c, clock, cursor, nil)
	if !c.clientSendInput(server.HostPlayerIndex) || list.Count() != 0 {
		t.Fatal("repeated blocked prepare queued an empty header or replayed consumed prefix")
	}
	clock.now = 1001
	tail := inputDelayPrepareR275(t, c, clock, c.ctrl.indA, events[64:])
	if len(inputFrameEventsR273(t, tail)) != 3 || !c.clientSendInput(server.HostPlayerIndex) ||
		!bytes.Equal(c.srv.NetList.CopyPacketsA(server.HostPlayerIndex, netlist.Kind0), tail) {
		t.Fatal("delayed tail did not retain complete framing/order through actual sender")
	}
	inputDelayPrepareR275(t, c, clock, c.ctrl.indA, nil)
	// No receiver storage/admission policy or transport-level retry guarantee
	// is inferred from these producer/sender queue checks.
}

func TestClientInputHostFullGuardAllCursorsR275(t *testing.T) {
	c, _ := inputDelayFixtureR275(t, true)
	for start := 0; start < ctrlEventCap; start++ {
		*c.ctrl = inputDelaySentinelRingR275(start, 127)
		before := *c.ctrl
		c.ctrl.nox_ctrlevent_action_42E670(player.CCOrientation, 0x53)
		if !reflect.DeepEqual(*c.ctrl, before) {
			t.Fatalf("full ring admitted input at read=%d write=%d", start, before.indA)
		}
		*c.ctrl = inputDelaySentinelRingR275(start, 126)
		before = *c.ctrl
		c.ctrl.nox_ctrlevent_action_42E670(player.CCOrientation, 0x71)
		want := before
		want.bufA[before.indA] = noxCtrlEvent{tick: 1000, PlayerCtrl: server.PlayerCtrl{Code: player.CCOrientation, Data: 0x71, Active: true}}
		want.indA = (before.indA + 1) % ctrlEventCap
		if !reflect.DeepEqual(*c.ctrl, want) {
			t.Fatalf("last available slot append changed unrelated state at read=%d write=%d", start, before.indA)
		}
		before = *c.ctrl
		c.ctrl.nox_ctrlevent_action_42E670(player.CCMoveForward, 0x99)
		if !reflect.DeepEqual(*c.ctrl, before) {
			t.Fatalf("newly full ring admitted extra input at read=%d write=%d", start, before.indA)
		}
	}
}

func TestClientInputHostPendingPauseDedupR275(t *testing.T) {
	c, clock := inputDelayFixtureR275(t, true)
	for _, active := range []bool{false, true} {
		*c.ctrl = inputDelaySentinelRingR275(127, 2)
		c.ctrl.bufA[127] = noxCtrlEvent{tick: 1051, PlayerCtrl: server.PlayerCtrl{Code: player.CCAction, Data: 0x61, Active: active}}
		c.ctrl.bufA[0] = noxCtrlEvent{tick: 999, PlayerCtrl: server.PlayerCtrl{Code: player.CCJump, Active: true}}
		before := *c.ctrl
		c.ctrl.nox_ctrlevent_action_42E670(player.CCAction, 0x99)
		if !reflect.DeepEqual(*c.ctrl, before) {
			t.Fatalf("pending pause-mode duplicate changed ring (active=%t)", active)
		}
		inputDelayPrepareR275(t, c, clock, 127, nil)
		before = *c.ctrl
		c.ctrl.nox_ctrlevent_action_42E670(player.CCAction, 0x88)
		if !reflect.DeepEqual(*c.ctrl, before) {
			t.Fatalf("blocked prepare lost pending pause-mode dedup (active=%t)", active)
		}
	}
}

func TestClientInputNonHostIgnoresTickGateR275(t *testing.T) {
	c, clock := inputDelayFixtureR275(t, false)
	*c.ctrl = inputDelaySentinelRingR275(117, 0)
	c.ctrl.indA = 128
	events := inputRepeatEventsR273(128, 29)
	for i := range events {
		events[i].tick = 1051
		events[i].Active = i%3 != 1
		c.ctrl.bufA[i] = events[i]
	}
	wire := inputDelayPrepareR275(t, c, clock, 117, events)
	if !c.clientSendInput(server.HostPlayerIndex) || !bytes.Equal(c.srv.NetList.CopyPacketsA(server.HostPlayerIndex, netlist.Kind0), wire) {
		t.Fatal("non-host active-event serialization or actual sender changed")
	}
	if again := inputDelayPrepareR275(t, c, clock, 117, events); !bytes.Equal(again, wire) {
		t.Fatal("non-host producer unexpectedly gained host cursor-consumption behavior")
	}
}

// uint32 timestamp wrap and the original owner's higher-word arithmetic are
// explicitly outside R275. These tests assert only cursor retention and the
// modulo full guard; no 64-bit timing equivalence is claimed.
