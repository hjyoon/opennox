package legacy

import (
	"reflect"
	"testing"
)

type netPlayerStatusPlayer4D8270 struct {
	index byte
}

type netPlayerStatusUpdate4D8270 struct {
	player *netPlayerStatusPlayer4D8270
}

type netPlayerStatusObject4D8270 struct {
	flags  uint32
	update *netPlayerStatusUpdate4D8270
}

func TestNetReportPlayerStatus4D8270ReadOrderAndPacket(t *testing.T) {
	oldPlayer := &netPlayerStatusPlayer4D8270{index: 0xfe}
	oldUpdate := &netPlayerStatusUpdate4D8270{player: oldPlayer}
	obj := &netPlayerStatusObject4D8270{flags: 0x89abcdef, update: oldUpdate}
	var events []string

	got := netReportPlayerStatus4D8270(obj, netPlayerStatusHooks4D8270[
		*netPlayerStatusObject4D8270,
		*netPlayerStatusUpdate4D8270,
		*netPlayerStatusPlayer4D8270,
	]{
		flags: func(obj *netPlayerStatusObject4D8270) uint32 {
			events = append(events, "flags")
			return obj.flags
		},
		updateData: func(obj *netPlayerStatusObject4D8270) *netPlayerStatusUpdate4D8270 {
			events = append(events, "update")
			update := obj.update
			obj.flags = 0
			obj.update = &netPlayerStatusUpdate4D8270{player: &netPlayerStatusPlayer4D8270{index: 1}}
			return update
		},
		player: func(update *netPlayerStatusUpdate4D8270) *netPlayerStatusPlayer4D8270 {
			events = append(events, "player")
			return update.player
		},
		playerInd: func(player *netPlayerStatusPlayer4D8270) byte {
			events = append(events, "index")
			return player.index
		},
		send: func(index byte, packet []byte) int {
			events = append(events, "send")
			if index != 0xfe {
				t.Fatalf("player index = %#x, want cached 0xfe", index)
			}
			if want := []byte{102, 0xef, 0xcd, 0xab, 0x89}; !reflect.DeepEqual(packet, want) {
				t.Fatalf("packet = % x, want % x", packet, want)
			}
			return 0x12345678
		},
	})
	if got != 0x12345678 {
		t.Fatalf("return = %#x, want send result", got)
	}
	if want := []string{"flags", "update", "player", "index", "send"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestNetReportPlayerStatus4D8270FaultOrder(t *testing.T) {
	var events []string
	defer func() {
		if recover() == nil {
			t.Fatal("nil object did not fault")
		}
		if want := []string{"flags"}; !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}()
	netReportPlayerStatus4D8270[*netPlayerStatusObject4D8270, *netPlayerStatusUpdate4D8270, *netPlayerStatusPlayer4D8270](
		nil,
		netPlayerStatusHooks4D8270[*netPlayerStatusObject4D8270, *netPlayerStatusUpdate4D8270, *netPlayerStatusPlayer4D8270]{
			flags: func(obj *netPlayerStatusObject4D8270) uint32 {
				events = append(events, "flags")
				return obj.flags
			},
			updateData: func(obj *netPlayerStatusObject4D8270) *netPlayerStatusUpdate4D8270 {
				events = append(events, "update")
				return obj.update
			},
		},
	)
}

func TestNetReportPlayerStatus4D8270DependentFaultOrder(t *testing.T) {
	for _, tc := range []struct {
		name   string
		obj    *netPlayerStatusObject4D8270
		events []string
	}{
		{
			name:   "update data",
			obj:    &netPlayerStatusObject4D8270{},
			events: []string{"flags", "update", "player"},
		},
		{
			name: "player",
			obj: &netPlayerStatusObject4D8270{
				update: &netPlayerStatusUpdate4D8270{},
			},
			events: []string{"flags", "update", "player", "index"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			defer func() {
				if recover() == nil {
					t.Fatal("missing dependent pointer did not fault")
				}
				if !reflect.DeepEqual(events, tc.events) {
					t.Fatalf("events = %v, want %v", events, tc.events)
				}
			}()
			netReportPlayerStatus4D8270(tc.obj, netPlayerStatusHooks4D8270[
				*netPlayerStatusObject4D8270,
				*netPlayerStatusUpdate4D8270,
				*netPlayerStatusPlayer4D8270,
			]{
				flags: func(obj *netPlayerStatusObject4D8270) uint32 {
					events = append(events, "flags")
					return obj.flags
				},
				updateData: func(obj *netPlayerStatusObject4D8270) *netPlayerStatusUpdate4D8270 {
					events = append(events, "update")
					return obj.update
				},
				player: func(update *netPlayerStatusUpdate4D8270) *netPlayerStatusPlayer4D8270 {
					events = append(events, "player")
					return update.player
				},
				playerInd: func(player *netPlayerStatusPlayer4D8270) byte {
					events = append(events, "index")
					return player.index
				},
			})
		})
	}
}
