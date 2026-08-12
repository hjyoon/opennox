package opennox

import (
	"fmt"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

func TestDelayedDeleteObject4E5CC0Guards(t *testing.T) {
	delayedDeleteObject_4E5CC0(nil, delayedDeleteObject4E5CC0Hooks{})

	oldNext := &server.Object{}
	obj := &server.Object{
		ObjFlags:    object.FlagActive | object.FlagDestroyed,
		DeletedNext: oldNext,
		DeletedAt:   0x89ABCDEF,
	}
	delayedDeleteObject_4E5CC0(obj, delayedDeleteObject4E5CC0Hooks{})
	if obj.ObjFlags != object.FlagActive|object.FlagDestroyed || obj.DeletedNext != oldNext ||
		obj.DeletedAt != 0x89ABCDEF {
		t.Fatal("nil or already-destroyed guard changed object state")
	}
}

func TestDelayedDeleteObject4E5CC0MatchesGAMEEXEOrder(t *testing.T) {
	const (
		initialDeletedAt = uint32(0xAAAAAAAA)
		frame            = uint32(0xFEDCBA98)
		netCode          = uint32(0x87654321)
	)
	oldOwner := &server.Object{ObjClass: object.ClassPlayer}
	newOwner := &server.Object{ObjClass: object.ClassPlayer}
	holder := &server.Object{}
	oldHead := &server.Object{}
	staleNext := &server.Object{}
	obj := &server.Object{
		ObjClass:    object.ClassMonster,
		ObjSubClass: object.SubClass(object.MonsterMigrate),
		ObjFlags:    object.FlagActive,
		NetCode:     netCode,
		InvHolder:   holder,
		ObjOwner:    oldOwner,
		DeletedNext: staleNext,
		DeletedAt:   initialDeletedAt,
		TeamVal:     server.ObjectTeam{ID: server.TeamID(3)},
	}
	listHead := oldHead
	var events []string
	hooks := delayedDeleteObject4E5CC0Hooks{
		isCreatureMonitored: func(owner, gotObj *server.Object) bool {
			events = append(events, "monitored")
			if owner != oldOwner || gotObj != obj {
				t.Fatal("monitor check received wrong objects")
			}
			obj.ObjOwner = newOwner
			return false
		},
		removeMonsterMonitors: func(owner, gotObj *server.Object) {
			events = append(events, "remove-monitors")
			if owner != newOwner || gotObj != obj {
				t.Fatal("monitor removal did not reload owner after monitor check")
			}
		},
		removeFromInventory: func(gotHolder, gotObj *server.Object) {
			events = append(events, "inventory")
			if gotHolder != holder || gotObj != obj {
				t.Fatal("inventory removal received wrong objects")
			}
		},
		cancelPlayerSpells: func(gotObj *server.Object) {
			events = append(events, "cancel-spells")
			if gotObj != obj {
				t.Fatal("spell cancellation received wrong object")
			}
		},
		questMode: func() bool {
			events = append(events, "quest-mode")
			return true
		},
		questDeleteMonster: func(gotObj *server.Object) {
			events = append(events, "quest-monster")
			if gotObj != obj || !obj.Class().Has(object.ClassMonster) {
				t.Fatal("quest monster cleanup received wrong object state")
			}
			obj.ObjClass = object.ClassPlayer
		},
		deletePlayer: func(gotObj *server.Object) {
			events = append(events, "player")
			if gotObj != obj {
				t.Fatal("player cleanup received wrong object")
			}
		},
		deletedList: func() *server.Object {
			events = append(events, "deleted-list")
			if !obj.Flags().Has(object.FlagDestroyed) || obj.DeletedNext != staleNext {
				t.Fatal("deleted-list head was read at the wrong mutation point")
			}
			return listHead
		},
		frame: func() uint32 {
			events = append(events, "frame")
			if listHead != oldHead || obj.DeletedNext != oldHead || obj.DeletedAt != initialDeletedAt {
				t.Fatal("frame was read at the wrong list mutation point")
			}
			return frame
		},
		setDeletedList: func(gotObj *server.Object) {
			events = append(events, "set-deleted-list")
			if gotObj != obj || listHead != oldHead || obj.DeletedAt != initialDeletedAt {
				t.Fatal("deleted-list head was written at the wrong mutation point")
			}
			listHead = gotObj
		},
		changeTeam: func(team *server.ObjectTeam, gotNetCode uint32) {
			events = append(events, "team")
			if team != obj.TeamPtr() || gotNetCode != netCode || listHead != obj || obj.DeletedAt != frame {
				t.Fatal("team change received wrong final deletion state")
			}
		},
	}

	delayedDeleteObject_4E5CC0(obj, hooks)
	wantEvents := []string{
		"monitored", "remove-monitors", "inventory", "cancel-spells", "quest-mode",
		"quest-monster", "player", "deleted-list", "frame", "set-deleted-list", "team",
	}
	if fmt.Sprint(events) != fmt.Sprint(wantEvents) {
		t.Fatalf("events: got %v, want %v", events, wantEvents)
	}
	if obj.ObjFlags != object.FlagActive|object.FlagDestroyed || obj.DeletedNext != oldHead ||
		obj.DeletedAt != frame || listHead != obj || obj.ObjOwner != newOwner {
		t.Fatal("final delayed-delete state does not match GAME.EXE")
	}

	before := len(events)
	delayedDeleteObject_4E5CC0(obj, hooks)
	if len(events) != before {
		t.Fatal("second deletion did not short-circuit on the destroyed flag")
	}
}

func TestDelayedDeleteObject4E5CC0ShortCircuits(t *testing.T) {
	player := &server.Object{ObjClass: object.ClassPlayer}
	tests := []struct {
		name      string
		obj       *server.Object
		monitored bool
		quest     bool
		want      []string
	}{
		{
			name: "non-player owner skips monitor path",
			obj: &server.Object{
				ObjClass:    object.ClassMonster,
				ObjSubClass: object.SubClass(object.MonsterMigrate),
				ObjOwner:    &server.Object{ObjClass: object.ClassMonster},
			},
			want: []string{"cancel-spells", "quest-mode", "deleted-list", "frame", "set-deleted-list"},
		},
		{
			name: "monitored migrating monster skips removal",
			obj: &server.Object{
				ObjClass:    object.ClassMonster,
				ObjSubClass: object.SubClass(object.MonsterMigrate),
				ObjOwner:    player,
			},
			monitored: true,
			want:      []string{"monitored", "cancel-spells", "quest-mode", "deleted-list", "frame", "set-deleted-list"},
		},
		{
			name: "subclass is tested after monitor callback",
			obj: &server.Object{
				ObjClass: object.ClassMonster,
				ObjOwner: player,
			},
			want: []string{"monitored", "cancel-spells", "quest-mode", "deleted-list", "frame", "set-deleted-list"},
		},
		{
			name:  "quest monster cleanup follows mode check",
			obj:   &server.Object{ObjClass: object.ClassMonster},
			quest: true,
			want:  []string{"cancel-spells", "quest-mode", "quest-monster", "deleted-list", "frame", "set-deleted-list"},
		},
		{
			name: "player cleanup follows quest check",
			obj:  &server.Object{ObjClass: object.ClassPlayer},
			want: []string{"cancel-spells", "quest-mode", "player", "deleted-list", "frame", "set-deleted-list"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			hooks := delayedDeleteObject4E5CC0Hooks{
				isCreatureMonitored: func(_, _ *server.Object) bool {
					events = append(events, "monitored")
					return tc.monitored
				},
				removeMonsterMonitors: func(_, _ *server.Object) {
					events = append(events, "remove-monitors")
				},
				removeFromInventory: func(_, _ *server.Object) {
					events = append(events, "inventory")
				},
				cancelPlayerSpells: func(_ *server.Object) {
					events = append(events, "cancel-spells")
				},
				questMode: func() bool {
					events = append(events, "quest-mode")
					return tc.quest
				},
				questDeleteMonster: func(_ *server.Object) {
					events = append(events, "quest-monster")
				},
				deletePlayer: func(_ *server.Object) {
					events = append(events, "player")
				},
				deletedList: func() *server.Object {
					events = append(events, "deleted-list")
					return nil
				},
				frame: func() uint32 {
					events = append(events, "frame")
					return 1
				},
				setDeletedList: func(_ *server.Object) {
					events = append(events, "set-deleted-list")
				},
				changeTeam: func(_ *server.ObjectTeam, _ uint32) {
					events = append(events, "team")
				},
			}

			delayedDeleteObject_4E5CC0(tc.obj, hooks)
			if fmt.Sprint(events) != fmt.Sprint(tc.want) {
				t.Fatalf("events: got %v, want %v", events, tc.want)
			}
		})
	}
}
