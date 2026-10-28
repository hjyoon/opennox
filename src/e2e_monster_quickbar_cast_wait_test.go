package opennox

import (
	"reflect"
	"testing"

	"github.com/opennox/libs/client/keybind"
	"github.com/opennox/libs/client/seat"
)

func TestMonsterQuickbarQueuedCastWaitsThroughLostLOS(t *testing.T) {
	st := e2eMonsterQuickbarTracker{key: keybind.KeyF, ticks: 60}
	st.step(e2eTestQuickbarSample(584), true)
	events, done := st.step(e2eTestQuickbarSample(585), true)
	if done || !reflect.DeepEqual(e2eTestQuickbarKeys(events), []seat.KeyboardEvent{{Key: keybind.KeyF, Pressed: true}}) {
		t.Fatal("queue the original F input once")
	}
	st.step(e2eTestQuickbarSample(586), true)
	// The retained failing campaign loses availability at589, saves at606,
	// then changes position at640. Input completion must wait until645.
	for frame := uint32(589); frame <= 645; frame++ {
		sample := e2eTestQuickbarSample(frame)
		sample.target.point.X, sample.target.point.Y = 0, 0
		events, done = st.step(sample, false)
		if done != (frame == 645) || len(events) != 0 {
			t.Fatalf("frame%d: unavailable aim after queued F must wait without input; done=%t events=%v", frame, done, events)
		}
	}
	if events, done = st.step(e2eTestQuickbarSample(646), true); !done || len(events) != 0 {
		t.Fatal("completed action cannot press F again or follow a replacement target")
	}
}

func TestMonsterQuickbarQueuedCastWaitsAfterTargetDeath(t *testing.T) {
	st := e2eMonsterQuickbarTracker{key: keybind.KeyA, ticks: 18}
	st.step(e2eTestQuickbarSample(1), true)
	st.step(e2eTestQuickbarSample(2), true)
	var keys []seat.KeyboardEvent
	for frame := uint32(3); frame <= 20; frame++ {
		sample := e2eTestQuickbarSample(frame)
		sample.target.dead, sample.target.health = true, 0
		events, done := st.step(sample, true)
		keys = append(keys, e2eTestQuickbarKeys(events)...)
		if done != (frame == 20) {
			t.Fatalf("frame%d: queued input's requested native observation window ended early", frame)
		}
		if len(events) != len(e2eTestQuickbarKeys(events)) {
			t.Fatal("dead target must receive no further mouse aiming")
		}
	}
	if !reflect.DeepEqual(keys, []seat.KeyboardEvent{{Key: keybind.KeyA, Pressed: false}}) {
		t.Fatal("release the held original key exactly once")
	}
}
