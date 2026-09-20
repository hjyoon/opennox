package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/opennox/libs/types"
)

type makeScorchTestWorld537AF0 struct {
	object  string
	quest   bool
	fps     uint32
	seconds int
	events  []string
}

func (w *makeScorchTestWorld537AF0) hooks() makeScorchHooks537AF0[string] {
	return makeScorchHooks537AF0[string]{
		newObject: func(typeID string) string {
			w.events = append(w.events, "new:"+typeID)
			return w.object
		},
		createObjectAt: func(obj string, pos types.Pointf) {
			w.events = append(w.events, fmt.Sprintf("create:%s:%v", obj, pos))
		},
		gameFlag: func(flag uint32) int32 {
			w.events = append(w.events, fmt.Sprintf("flag:%04x", flag))
			if w.quest {
				return 1
			}
			return 0
		},
		randomInt: func(minimum, maximum int) int {
			w.events = append(w.events, fmt.Sprintf("random:%d:%d", minimum, maximum))
			return w.seconds
		},
		loadGameFPS: func() uint32 {
			w.events = append(w.events, fmt.Sprintf("fps:%d", w.fps))
			return w.fps
		},
		setDecay: func(obj string, delay uint32) {
			w.events = append(w.events, fmt.Sprintf("decay:%s:%08x", obj, delay))
		},
	}
}

func TestMakeScorch537AF0TypeAndNormalLifetime(t *testing.T) {
	for kind, typeID := range scorchTypeIDs537AF0 {
		t.Run(typeID, func(t *testing.T) {
			w := &makeScorchTestWorld537AF0{object: "scorch", fps: 30, seconds: 17}
			makeScorch537AF0(types.Ptf(12.5, 34.25), kind, w.hooks())
			want := []string{
				"random:0:0",
				"new:" + typeID,
				"create:scorch:{12.5 34.25}",
				"flag:1000",
				"random:10:20",
				"fps:30",
				"decay:scorch:000001fe",
			}
			if !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events = %v, want %v", w.events, want)
			}
		})
	}
}

func TestMakeScorch537AF0QuestLifetimeAndUint32Wrap(t *testing.T) {
	w := &makeScorchTestWorld537AF0{
		object:  "scorch",
		quest:   true,
		fps:     math.MaxUint32,
		seconds: 8,
	}
	makeScorch537AF0(types.Pointf{}, 0, w.hooks())
	want := []string{
		"random:0:0",
		"new:ScorchMarkFloorSmallA",
		"create:scorch:{0 0}",
		"flag:1000",
		"random:5:8",
		fmt.Sprintf("fps:%d", uint32(math.MaxUint32)),
		"decay:scorch:fffffff8",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestMakeScorch537AF0RejectsInvalidKindAndHandlesMissingType(t *testing.T) {
	for _, kind := range []int{-1, len(scorchTypeIDs537AF0)} {
		w := &makeScorchTestWorld537AF0{object: "scorch"}
		makeScorch537AF0(types.Pointf{}, kind, w.hooks())
		if len(w.events) != 0 {
			t.Fatalf("kind %d events = %v, want none", kind, w.events)
		}
	}

	w := &makeScorchTestWorld537AF0{}
	makeScorch537AF0(types.Pointf{}, 1, w.hooks())
	want := []string{"random:0:0", "new:ScorchMarkFloorMediumA"}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("missing-type events = %v, want %v", w.events, want)
	}
}
