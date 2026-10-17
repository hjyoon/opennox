package opennox

import (
	"strings"
	"testing"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestE2EWaitClientConnectedRejectsIncompleteBinding(t *testing.T) {
	unit, free := alloc.New(client.Drawable{})
	*unit = client.Drawable{}
	connected := memmap.PtrUint32(0x5D4594, 815764)
	drawable := memmap.PtrPtr(0x852978, 8)
	oldClient, oldConnected, oldDrawable, oldCode := noxClient, *connected, *drawable, legacy.ClientPlayerNetCode()
	t.Cleanup(func() {
		noxClient, *connected, *drawable = oldClient, oldConnected, oldDrawable
		legacy.ClientSetPlayerNetCode(oldCode)
		free()
	})
	var sc e2eScenario
	sc.WaitClientConnected("remote player binding")
	step := sc.steps[0]
	for _, tc := range []struct {
		name                      string
		missingClient, noDrawable bool
		connected                 uint32
		code                      int
		drawableCode              uint32
		want                      bool
	}{
		{name: "missing client", missingClient: true, connected: 1, code: 185, drawableCode: 185},
		{name: "not connected", code: 185, drawableCode: 185},
		{name: "no local drawable", noDrawable: true, connected: 1, code: 185},
		{name: "zero player netcode", connected: 1, drawableCode: 185},
		{name: "mismatched netcodes", connected: 1, code: 185, drawableCode: 184},
		{name: "matching native binding", connected: 1, code: 185, drawableCode: 185, want: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			noxClient, *connected, *drawable = &Client{}, tc.connected, unit.C()
			if tc.missingClient {
				noxClient = nil
			}
			if tc.noDrawable {
				*drawable = nil
			}
			legacy.ClientSetPlayerNetCode(tc.code)
			unit.NetCode32 = tc.drawableCode
			if got := step.ready(); got != tc.want {
				t.Fatalf("production readiness=%v, want %v", got, tc.want)
			}
		})
	}
}

func TestE2EWaitClientConnectedTimesOutWithoutFalseSuccess(t *testing.T) {
	preserveE2EErrorInputState(t)
	oldClient, oldPlatform, oldSteps, oldDone, oldSlow := noxClient, e2e.p, e2e.steps, e2e.done, e2e.slow
	oldTrace := e2eCombatTraceEnabled
	t.Cleanup(func() {
		noxClient, e2e.p, e2e.steps, e2e.done, e2e.slow = oldClient, oldPlatform, oldSteps, oldDone, oldSlow
		e2eCombatTraceEnabled = oldTrace
	})
	noxClient, e2e.p, e2e.done, e2e.slow = nil, &platformE2E{}, nil, 0
	e2eCombatTraceEnabled = false
	var sc e2eScenario
	sc.WaitClientConnected("remote player binding")
	called := false
	sc.steps[0].fnc = func() { called = true }
	e2e.steps = sc.steps
	for i := 0; i < 499; i++ {
		e2eRun()
	}
	if called || e2e.err != nil || len(e2e.steps) != 1 || e2e.steps[0].waited != 499 {
		t.Fatalf("gate did not retain its iteration budget: called=%v error=%v steps=%v", called, e2e.err, e2e.steps)
	}
	e2eRun()
	if called || len(e2e.steps) != 0 || e2e.err == nil || !strings.Contains(e2e.err.Error(), "timed out after 500 ticks waiting for remote player binding") {
		t.Fatalf("timeout became success or lost its failure: called=%v error=%v steps=%v", called, e2e.err, e2e.steps)
	}
}
