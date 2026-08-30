package server

import (
	"fmt"
	"reflect"
	"testing"
)

type playerInventoryStrengthTestObject4F8420 struct {
	name     string
	flags    uint32
	next     *playerInventoryStrengthTestObject4F8420
	strength int32
}

type playerInventoryStrengthTestWorld4F8420 struct {
	player      *playerInventoryStrengthTestObject4F8420
	first       *playerInventoryStrengthTestObject4F8420
	strong      *playerInventoryStrengthTestObject4F8420
	weak        *playerInventoryStrengthTestObject4F8420
	replacement *playerInventoryStrengthTestObject4F8420
	events      []string
	faultAt     int
}

func playerInventoryStrengthObjectName4F8420(obj *playerInventoryStrengthTestObject4F8420) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func (w *playerInventoryStrengthTestWorld4F8420) event(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *playerInventoryStrengthTestWorld4F8420) hooks() playerInventoryStrengthHooks4F8420[*playerInventoryStrengthTestObject4F8420] {
	return playerInventoryStrengthHooks4F8420[*playerInventoryStrengthTestObject4F8420]{
		loadInventoryHead: func(player *playerInventoryStrengthTestObject4F8420) *playerInventoryStrengthTestObject4F8420 {
			w.event("head:" + playerInventoryStrengthObjectName4F8420(player))
			return w.first
		},
		loadItemFlags: func(item *playerInventoryStrengthTestObject4F8420) uint32 {
			w.event(fmt.Sprintf("flags:%s:%08x", item.name, item.flags))
			return item.flags
		},
		checkStrength: func(player, item *playerInventoryStrengthTestObject4F8420) int32 {
			w.event("strength:" + player.name + ":" + item.name)
			if item == w.strong {
				// The original reloads strong.next after this callback.
				item.next = w.weak
			}
			return item.strength
		},
		forceDrop: func(player, item *playerInventoryStrengthTestObject4F8420) int32 {
			w.event("drop:" + player.name + ":" + item.name)
			if item == w.weak {
				// The original reloads weak.next after force-drop as well.
				item.next = w.replacement
			}
			return -0x7654321
		},
		loadInventoryNext: func(item *playerInventoryStrengthTestObject4F8420) *playerInventoryStrengthTestObject4F8420 {
			w.event("next:" + item.name)
			return item.next
		},
	}
}

func newPlayerInventoryStrengthTrace4F8420() (
	*playerInventoryStrengthTestWorld4F8420,
	[]string,
) {
	player := &playerInventoryStrengthTestObject4F8420{name: "player"}
	plain := &playerInventoryStrengthTestObject4F8420{
		name:  "plain",
		flags: 0xffffffff &^ playerInventoryEquippedFlag4F8420,
	}
	strong := &playerInventoryStrengthTestObject4F8420{
		name:     "strong",
		flags:    playerInventoryEquippedFlag4F8420,
		strength: -1,
	}
	staleAfterStrength := &playerInventoryStrengthTestObject4F8420{
		name:  "stale-after-strength",
		flags: playerInventoryEquippedFlag4F8420,
	}
	weak := &playerInventoryStrengthTestObject4F8420{
		name:  "weak",
		flags: playerInventoryEquippedFlag4F8420,
	}
	staleAfterDrop := &playerInventoryStrengthTestObject4F8420{
		name:  "stale-after-drop",
		flags: playerInventoryEquippedFlag4F8420,
	}
	replacement := &playerInventoryStrengthTestObject4F8420{
		name:  "replacement",
		flags: 0x80000200,
	}
	plain.next = strong
	strong.next = staleAfterStrength
	weak.next = staleAfterDrop
	w := &playerInventoryStrengthTestWorld4F8420{
		player:      player,
		first:       plain,
		strong:      strong,
		weak:        weak,
		replacement: replacement,
	}
	want := []string{
		"head:player",
		"flags:plain:fffffeff",
		"next:plain",
		"flags:strong:00000100",
		"strength:player:strong",
		"next:strong",
		"flags:weak:00000100",
		"strength:player:weak",
		"drop:player:weak",
		"next:weak",
		"flags:replacement:80000200",
		"next:replacement",
	}
	return w, want
}

func TestPlayerInventoryStrength4F8420ExactTraceAndLiveNext(t *testing.T) {
	w, want := newPlayerInventoryStrengthTrace4F8420()
	playerInventoryStrength4F8420(w.player, w.hooks())
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events =\n%v\nwant =\n%v", w.events, want)
	}
}

func TestPlayerInventoryStrength4F8420FaultPrefixes(t *testing.T) {
	_, want := newPlayerInventoryStrengthTrace4F8420()
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w, _ := newPlayerInventoryStrengthTrace4F8420()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if !reflect.DeepEqual(w.events, want[:faultAt]) {
					t.Fatalf("events = %v, want %v", w.events, want[:faultAt])
				}
			}()
			playerInventoryStrength4F8420(w.player, w.hooks())
		})
	}
}

func TestPlayerInventoryStrength4F8420EmptyInventoryStopsAfterHead(t *testing.T) {
	w := &playerInventoryStrengthTestWorld4F8420{
		player: &playerInventoryStrengthTestObject4F8420{name: "player"},
	}
	playerInventoryStrength4F8420(w.player, w.hooks())
	if want := []string{"head:player"}; !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestPlayerInventoryStrength4F8420DoesNotGuardNilPlayer(t *testing.T) {
	w := &playerInventoryStrengthTestWorld4F8420{}
	w.faultAt = 1
	defer func() {
		if got := recover(); got != "head:nil" {
			t.Fatalf("panic = %v, want head:nil", got)
		}
	}()
	playerInventoryStrength4F8420(w.player, w.hooks())
}

func TestPlayerInventoryStrength4F8420WholeStrengthResult(t *testing.T) {
	for _, strength := range []int32{1, -1, 0x76543210, -0x7654321} {
		item := &playerInventoryStrengthTestObject4F8420{
			name:     "item",
			flags:    playerInventoryEquippedFlag4F8420,
			strength: strength,
		}
		w := &playerInventoryStrengthTestWorld4F8420{
			player: &playerInventoryStrengthTestObject4F8420{name: "player"},
			first:  item,
		}
		playerInventoryStrength4F8420(w.player, w.hooks())
		for _, event := range w.events {
			if event == "drop:player:item" {
				t.Fatalf("strength %#x called drop: events = %v", uint32(strength), w.events)
			}
		}
	}
}
