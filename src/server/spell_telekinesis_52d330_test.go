package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/sound"
)

type telekinesisCastWorld52D330 struct {
	target, hand uint64
	class        uint32
	fps          uint32
	events       []string
	mutate       func(string)
	failAt       string
}

func (w *telekinesisCastWorld52D330) event(event string) {
	w.events = append(w.events, event)
	if w.mutate != nil {
		w.mutate(event)
	}
	if event == w.failAt {
		panic(event)
	}
}

func (w *telekinesisCastWorld52D330) hooks() telekinesisCastHooks52D330[uint64] {
	return telekinesisCastHooks52D330[uint64]{
		target:    func() uint64 { v := w.target; w.event(fmt.Sprintf("target:%x", v)); return v },
		class:     func(v uint64) uint32 { w.event(fmt.Sprintf("class:%x", v)); return w.class },
		newObject: func(name string) uint64 { w.event("new:" + name); return w.hand },
		positionY: func(v uint64) float32 { w.event(fmt.Sprintf("y:%x", v)); return math.Float32frombits(0x80000000) },
		positionX: func(v uint64) float32 { w.event(fmt.Sprintf("x:%x", v)); return math.Float32frombits(0x7fc01234) },
		createAt: func(hand, target uint64, point types.Pointf) {
			w.event(fmt.Sprintf("create:%x:%x:%08x:%08x", hand, target, math.Float32bits(point.X), math.Float32bits(point.Y)))
		},
		fps: func() uint32 { w.event("fps"); return w.fps },
		apply: func(v uint64, buff int32, duration int16, power int8) {
			w.event(fmt.Sprintf("apply:%x:%d:%d:%d", v, buff, duration, power))
		},
		cancel:  func(id int32, v uint64) { w.event(fmt.Sprintf("cancel:%d:%x", id, v)) },
		onSound: func(id int32) sound.ID { w.event(fmt.Sprintf("sound:%d", id)); return sound.ID(777) },
		audio: func(snd sound.ID, v uint64, kind int, code uint32) {
			w.event(fmt.Sprintf("audio:%d:%x:%d:%d", snd, v, kind, code))
		},
	}
}

func newTelekinesisCastWorld52D330() *telekinesisCastWorld52D330 {
	return &telekinesisCastWorld52D330{target: 0x100000001, hand: 0x200000002, class: 4, fps: 30}
}

func TestTelekinesisCast52D330ExactOriginalTrace(t *testing.T) {
	w := newTelekinesisCastWorld52D330()
	want := []string{
		"target:100000001", "class:100000001", "new:TelekinesisHand",
		"target:100000001", "y:100000001", "x:100000001", "create:200000002:100000001:7fc01234:80000000",
		"target:100000001", "fps", "apply:100000001:24:600:3",
		"target:100000001", "cancel:24:100000001", "target:100000001", "cancel:43:100000001",
		"target:100000001", "sound:123", "audio:777:100000001:0:0",
	}
	if got := telekinesisCast52D330(123, 3, w.hooks()); got != 1 || !reflect.DeepEqual(w.events, want) {
		t.Fatalf("result=%d\ntrace=%q\nwant=%q", got, w.events, want)
	}
}

func TestTelekinesisCast52D330Gates(t *testing.T) {
	for _, tc := range []struct {
		name         string
		target, hand uint64
		class        uint32
		result       int32
		want         []string
	}{
		{"nil target", 0, 2, 4, 0, []string{"target:0"}},
		{"monster", 0x100000001, 2, 2, 1, []string{"target:100000001", "class:100000001"}},
		{"upper byte is not player", 0x100000001, 2, 0x400, 1, []string{"target:100000001", "class:100000001"}},
		{"allocation failure", 0x100000001, 0, 4, 1, []string{"target:100000001", "class:100000001", "new:TelekinesisHand"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := newTelekinesisCastWorld52D330()
			w.target, w.hand, w.class = tc.target, tc.hand, tc.class
			if got := telekinesisCast52D330(math.MinInt32, math.MaxInt32, w.hooks()); got != tc.result || !reflect.DeepEqual(w.events, tc.want) {
				t.Fatalf("result=%d events=%q", got, w.events)
			}
		})
	}
}

func TestTelekinesisCast52D330LiveTargetsAndCachedArguments(t *testing.T) {
	w := newTelekinesisCastWorld52D330()
	w.mutate = func(event string) {
		switch event {
		case "new:TelekinesisHand":
			w.target = 0x300000003
		case "create:200000002:300000003:7fc01234:80000000":
			w.target = 0x400000004
		case "fps":
			w.target = 0x500000005
		case "apply:400000004:24:600:3":
			w.target = 0x600000006
		case "cancel:24:600000006":
			w.target = 0x700000007
		case "cancel:43:700000007":
			w.target = 0x800000008
		case "sound:123":
			w.target = 0x900000009
		}
	}
	want := []string{
		"target:100000001", "class:100000001", "new:TelekinesisHand",
		"target:300000003", "y:300000003", "x:300000003", "create:200000002:300000003:7fc01234:80000000",
		"target:400000004", "fps", "apply:400000004:24:600:3",
		"target:600000006", "cancel:24:600000006", "target:700000007", "cancel:43:700000007",
		"target:800000008", "sound:123", "audio:777:800000008:0:0",
	}
	if got := telekinesisCast52D330(123, 3, w.hooks()); got != 1 || !reflect.DeepEqual(w.events, want) {
		t.Fatalf("result=%d\ntrace=%q\nwant=%q", got, w.events, want)
	}
}

func TestTelekinesisCast52D330FaultPrefixes(t *testing.T) {
	baseline := newTelekinesisCastWorld52D330()
	telekinesisCast52D330(123, 3, baseline.hooks())
	for i, event := range baseline.events {
		// Repeated target loads have dedicated mutation coverage above.
		if event == "target:100000001" && i != 0 {
			continue
		}
		t.Run(fmt.Sprintf("%02d %s", i, event), func(t *testing.T) {
			w := newTelekinesisCastWorld52D330()
			w.failAt = event
			func() {
				defer func() {
					if recover() != event {
						t.Fatal("required fault did not occur")
					}
				}()
				telekinesisCast52D330(123, 3, w.hooks())
			}()
			if !reflect.DeepEqual(w.events, baseline.events[:i+1]) {
				t.Fatalf("fault prefix=%q", w.events)
			}
		})
	}
}

func TestTelekinesisCast52D330OriginalIntegerWidths(t *testing.T) {
	for _, fps := range []uint32{0, 1, 30, 1638, 1639, 32768, 65535, 65536, math.MaxUint32} {
		for _, power := range []int32{math.MinInt32, -129, -128, -1, 0, 1, 127, 128, 255, math.MaxInt32} {
			t.Run(fmt.Sprintf("fps%d power%d", fps, power), func(t *testing.T) {
				w := newTelekinesisCastWorld52D330()
				w.fps = fps
				h := w.hooks()
				h.apply = func(v uint64, buff int32, duration int16, gotPower int8) {
					// Independent low-WORD multiplication and signed-BYTE decode.
					word := (uint64(fps) * 20) & 65535
					wantDuration := int64(word)
					if word >= 32768 {
						wantDuration -= 65536
					}
					bytePower := int64(uint32(power) & 255)
					if bytePower >= 128 {
						bytePower -= 256
					}
					if v != w.target || buff != 24 || int64(duration) != wantDuration || int64(gotPower) != bytePower {
						t.Fatalf("apply=%x/%d/%d/%d", v, buff, duration, gotPower)
					}
				}
				if got := telekinesisCast52D330(math.MinInt32, power, h); got != 1 {
					t.Fatalf("result=%d", got)
				}
			})
		}
	}
}

func TestTelekinesisCastNative52D330NilArgumentFaultsBeforeServices(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("nil required acceptance argument was silently guarded")
		}
	}()
	new(Server).CastTelekinesis52D330(0, nil, nil, nil, nil, 0, TelekinesisCastRuntime52D330{})
}
