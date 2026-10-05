package gui

import (
	"fmt"
	"reflect"
	"testing"
)

func TestNPCDialogProc479B00NonClick(t *testing.T) {
	for _, event := range []int{0, 1, 2, 21, 23, 0x4000, 0x4006, 0x4008} {
		t.Run(fmt.Sprintf("event_%x", event), func(t *testing.T) {
			// Neither a window nor any service is needed before the event gate.
			if got := NPCDialogProc479B00(event, nil, NPCDialogRuntime479B00{}); got != 0 {
				t.Fatalf("return=%d", got)
			}
		})
	}
}

func TestNPCDialogProc479B00ButtonsAndPause(t *testing.T) {
	g := New(nil)
	t.Cleanup(g.DestroyAll)
	button := g.NewWindowRaw(nil, StatusEnabled, 0, 0, 80, 20, nil)
	for _, id := range []uint{3906, 3907, 3908, 3909, 0, ^uint(0)} {
		for _, paused := range []int{0, 1, -1} {
			t.Run(fmt.Sprintf("button_%d_pause_%d", id, paused), func(t *testing.T) {
				button.SetID(id)
				var calls []string
				r := NPCDialogRuntime479B00{
					Paused:     func() int { calls = append(calls, "pause"); return paused },
					ClickSound: func() { calls = append(calls, "sound") },
					Done:       func() { calls = append(calls, "done") },
					Repeat:     func() { calls = append(calls, "repeat") },
					Answer:     func(v byte) { calls = append(calls, fmt.Sprintf("answer_%d", v)) },
				}
				got := NPCDialogProc479B00(0x4007, button, r)
				want := []string{"pause"}
				if paused == 0 {
					want = append(want, "sound")
					switch id {
					case 3906:
						want = append(want, "done")
					case 3907:
						want = append(want, "repeat")
					case 3908:
						want = append(want, "answer_1")
					case 3909:
						want = append(want, "answer_2")
					}
				}
				if got != 0 || !reflect.DeepEqual(calls, want) {
					t.Fatalf("return=%d calls=%v, want zero/%v", got, calls, want)
				}
			})
		}
	}
}

func TestNPCDialogProc479B00CachedButtonBeforeCallbacks(t *testing.T) {
	for _, id := range []uint{3906, 3907, 3908, 3909} {
		t.Run(fmt.Sprintf("button_%d", id), func(t *testing.T) {
			g := New(nil)
			t.Cleanup(g.DestroyAll)
			button := g.NewWindowRaw(nil, StatusEnabled, 0, 0, 80, 20, nil)
			button.SetID(id)
			var selected uint
			r := NPCDialogRuntime479B00{
				Paused:     func() int { button.SetID(3907); return 0 },
				ClickSound: func() { button.SetID(3909) },
				Done:       func() { selected = 3906 },
				Repeat:     func() { selected = 3907 },
				Answer:     func(v byte) { selected = 3907 + uint(v) },
			}
			if got := NPCDialogProc479B00(0x4007, button, r); got != 0 || selected != id || button.ID() != 3909 {
				t.Fatalf("return=%d selected=%d live ID=%d, want cached %d", got, selected, button.ID(), id)
			}
		})
	}
}

func TestNPCDialogProc479B00NullWindowUsesOriginalZeroID(t *testing.T) {
	var calls []string
	r := NPCDialogRuntime479B00{
		Paused:     func() int { calls = append(calls, "pause"); return 0 },
		ClickSound: func() { calls = append(calls, "sound") },
	}
	if got := NPCDialogProc479B00(0x4007, nil, r); got != 0 || !reflect.DeepEqual(calls, []string{"pause", "sound"}) {
		t.Fatalf("return=%d calls=%v", got, calls)
	}
}
