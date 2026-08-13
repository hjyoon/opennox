package legacy

import (
	"reflect"
	"testing"

	"github.com/opennox/opennox/v1/server"
)

func TestUnitPostCreateNotifyNative4E7F10(t *testing.T) {
	created := &server.Object{Field35: 0xffffffff, Field36: 0xaaaaaaaa, Field37: 0x11223344}
	nonHostileUnit := &server.Object{TypeInd: 17}
	hostileUnit := &server.Object{TypeInd: 18}
	second := &server.Player{PlayerInd: 0x22, PlayerUnit: hostileUnit}
	first := &server.Player{PlayerInd: 0x21, PlayerUnit: nonHostileUnit}
	var events []string
	got := unitPostCreateNotifyNative4E7F10(created, unitPostCreateNotifyNativeDeps4E7F10{
		firstPlayer: func() *server.Player {
			events = append(events, "first")
			return first
		},
		nextPlayer: func(player *server.Player) *server.Player {
			switch player {
			case first:
				events = append(events, "next:first")
				return second
			case second:
				events = append(events, "next:second")
				return nil
			default:
				t.Fatalf("unexpected next player %p", player)
				return nil
			}
		},
		isHostile: func(unit, obj *server.Object) int32 {
			events = append(events, "hostile")
			if obj != created {
				t.Fatalf("created object = %p, want %p", obj, created)
			}
			if unit == hostileUnit {
				return 1
			}
			return 0
		},
	})
	if got != nil {
		t.Fatalf("return = %p, want nil", got)
	}
	if created.Field35 != 4 || created.Field36 != 4 {
		t.Fatalf("masks = (%#x, %#x), want (4, 4)", created.Field35, created.Field36)
	}
	if created.Field37 != 0x11223344 || first.PlayerInd != 0x21 || second.PlayerInd != 0x22 {
		t.Fatal("native adapter changed an unrelated field")
	}
	wantEvents := []string{"first", "hostile", "next:first", "hostile", "next:second"}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
}

func TestUnitPostCreateNotifyNative4E7F10NilFaultsBeforeFirstPlayer(t *testing.T) {
	firstCalled := false
	defer func() {
		if recover() == nil {
			t.Fatal("nil object returned without a panic")
		}
		if firstCalled {
			t.Fatal("player list was read before the first object store fault")
		}
	}()
	unitPostCreateNotifyNative4E7F10(nil, unitPostCreateNotifyNativeDeps4E7F10{
		firstPlayer: func() *server.Player {
			firstCalled = true
			return nil
		},
	})
}
