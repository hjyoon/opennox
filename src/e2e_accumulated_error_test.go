package opennox

import (
	"errors"
	"testing"

	"github.com/opennox/libs/client/seat"
)

func preserveE2EErrorInputState(t *testing.T) {
	t.Helper()
	oldErr, oldFailFast := e2e.err, e2eFailFast
	oldInput, oldOnInput := e2e.input, e2e.onInput
	oldRecording := e2e.recording
	t.Cleanup(func() {
		e2e.err, e2eFailFast = oldErr, oldFailFast
		e2e.input, e2e.onInput = oldInput, oldOnInput
		e2e.recording = oldRecording
	})
	e2e.err, e2eFailFast = nil, false
	e2e.input, e2e.onInput = nil, nil
	e2e.recording = false
}

func TestE2EAccumulatedErrorSurvivesEarlyClose(t *testing.T) {
	for _, tc := range []struct {
		name     string
		schedule func(*e2eScenario)
	}{
		{"quit", func(sc *e2eScenario) { sc.Quit(1) }},
		{"recorded close", func(sc *e2eScenario) { sc.RawInput(1, seat.WindowClosed) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			preserveE2EErrorInputState(t)
			want := errors.New("collected scenario failure")
			var closed bool
			e2e.onInput = []func(seat.InputEvent){func(ev seat.InputEvent) {
				if ev == seat.WindowClosed {
					closed = true
				}
			}}
			e2eError(want)
			var sc e2eScenario
			tc.schedule(&sc)
			defer func() {
				if got := recover(); got != want {
					t.Fatalf("close result = %v, want original failure %v", got, want)
				}
			}()
			// The first close can end the game loop before any later
			// scenario tick. runArgs still defers the shared e2eStop hook.
			sc.steps[0].fnc()
			e2eInputTick()
			if !closed {
				t.Fatal("scenario did not deliver its close event")
			}
			e2eStop()
		})
	}
}

func TestE2ECollectedFailureStillAllowsOrdinaryInput(t *testing.T) {
	preserveE2EErrorInputState(t)
	want := errors.New("collected scenario failure")
	e2eError(want)
	input := &seat.KeyboardEvent{Pressed: true}
	var got seat.InputEvent
	e2e.onInput = []func(seat.InputEvent){func(ev seat.InputEvent) { got = ev }}
	e2eQueueRawInput(input)
	e2eInputTick()
	if got != input || e2e.err != want || len(e2e.input) != 0 {
		t.Fatalf("ordinary input/failure changed: input=%v failure=%v queued=%d", got, e2e.err, len(e2e.input))
	}
}

func TestE2ESuccessfulQuitPreservesCloseInputs(t *testing.T) {
	preserveE2EErrorInputState(t)
	var closed int
	e2e.onInput = []func(seat.InputEvent){func(ev seat.InputEvent) {
		if ev != seat.WindowClosed {
			t.Fatalf("unexpected quit input: %v", ev)
		}
		closed++
	}}
	var sc e2eScenario
	sc.Quit(1)
	for _, step := range sc.steps {
		if step.fnc != nil {
			step.fnc()
		}
		e2eInputTick()
	}
	e2eStop()
	if closed != 2 || e2e.err != nil {
		t.Fatalf("successful quit delivered %d close events with error %v", closed, e2e.err)
	}
}

func TestE2EStopPreservesExistingPanic(t *testing.T) {
	original := errors.New("original runtime failure")
	collected := errors.New("collected scenario failure")
	for _, tc := range []struct {
		name      string
		panicking any
		collected error
		want      any
	}{
		{"runtime failure", original, collected, original},
		{"nonzero exit", ErrExit(7), collected, ErrExit(7)},
		{"successful exit with failure", ErrExit(0), collected, collected},
		{"successful exit", ErrExit(0), nil, ErrExit(0)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			preserveE2EErrorInputState(t)
			e2e.err = tc.collected
			defer func() {
				if got := recover(); got != tc.want {
					t.Fatalf("shutdown panic = %v, want %v", got, tc.want)
				}
			}()
			func() {
				defer e2eStop()
				panic(tc.panicking)
			}()
		})
	}
}
