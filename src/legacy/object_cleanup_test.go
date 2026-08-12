package legacy

import (
	"fmt"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

func TestCleanupObjectsForMapLoad4E5BF0MatchesGAMEEXEContract(t *testing.T) {
	const moonglowTypeInd = 77
	player := &server.Object{ObjClass: object.ClassPlayer}

	tests := []struct {
		name        string
		mode        int
		objects     []*server.Object
		missiles    []*server.Object
		migrating   map[*server.Object]bool
		playerPixie map[*server.Object]bool
		wantEvents  []string
	}{
		{
			name: "empty lists still cache moonglow before either traversal",
			mode: 1,
			wantEvents: []string{
				"moonglow", "first-object", "first-missile",
			},
		},
		{
			name: "mode zero deletes every object after saving its successor",
			objects: []*server.Object{
				{ObjClass: object.ClassPlayer},
				{TypeInd: moonglowTypeInd, ObjOwner: player},
			},
			missiles: []*server.Object{
				{ObjClass: object.ClassMissile},
				{ObjClass: object.ClassMissile},
			},
			wantEvents: []string{
				"moonglow", "first-object",
				"next:object-0", "delete:object-0",
				"next:object-1", "delete:object-1",
				"first-missile",
				"next:missile-0", "delete:missile-0",
				"next:missile-1", "delete:missile-1",
			},
		},
		{
			name: "mode one applies every preservation rule in original order",
			mode: 1,
			objects: []*server.Object{
				{ObjClass: object.ClassPlayer},
				{InvHolder: player},
				{TypeInd: moonglowTypeInd, ObjOwner: player},
				{},
				{},
			},
			missiles: []*server.Object{
				{ObjClass: object.ClassMissile},
				{ObjClass: object.ClassMissile},
			},
			migrating:   map[*server.Object]bool{},
			playerPixie: map[*server.Object]bool{},
			wantEvents: []string{
				"moonglow", "first-object",
				"next:object-0",
				"next:object-1",
				"next:object-2",
				"next:object-3", "migrating:object-3",
				"next:object-4", "migrating:object-4", "delete:object-4",
				"first-missile",
				"next:missile-0", "pixie:missile-0",
				"next:missile-1", "pixie:missile-1", "delete:missile-1",
			},
		},
		{
			name: "mode two keeps normal-list rules but deletes all missiles without pixie tests",
			mode: 2,
			objects: []*server.Object{
				{ObjClass: object.ClassPlayer},
				{},
			},
			missiles: []*server.Object{
				{ObjClass: object.ClassMissile},
				{ObjClass: object.ClassMissile},
			},
			migrating:   map[*server.Object]bool{},
			playerPixie: map[*server.Object]bool{},
			wantEvents: []string{
				"moonglow", "first-object",
				"next:object-0",
				"next:object-1", "migrating:object-1", "delete:object-1",
				"first-missile",
				"next:missile-0", "delete:missile-0",
				"next:missile-1", "delete:missile-1",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			labels := make(map[*server.Object]string)
			for i, obj := range tc.objects {
				labels[obj] = fmt.Sprintf("object-%d", i)
				if i+1 < len(tc.objects) {
					obj.ObjNext = tc.objects[i+1]
				}
			}
			for i, obj := range tc.missiles {
				labels[obj] = fmt.Sprintf("missile-%d", i)
				if i+1 < len(tc.missiles) {
					obj.ObjNext = tc.missiles[i+1]
				}
			}
			if len(tc.objects) >= 4 && tc.migrating != nil {
				tc.migrating[tc.objects[3]] = true
			}
			if len(tc.missiles) != 0 && tc.playerPixie != nil {
				tc.playerPixie[tc.missiles[0]] = true
			}

			var events []string
			first := func(objs []*server.Object) *server.Object {
				if len(objs) == 0 {
					return nil
				}
				return objs[0]
			}
			hooks := objectCleanup4E5BF0Hooks{
				moonglowTypeInd: func() int {
					events = append(events, "moonglow")
					return moonglowTypeInd
				},
				firstObject: func() *server.Object {
					events = append(events, "first-object")
					return first(tc.objects)
				},
				nextObject: func(obj *server.Object) *server.Object {
					events = append(events, "next:"+labels[obj])
					return obj.ObjNext
				},
				firstMissile: func() *server.Object {
					events = append(events, "first-missile")
					return first(tc.missiles)
				},
				nextMissile: func(obj *server.Object) *server.Object {
					events = append(events, "next:"+labels[obj])
					return obj.ObjNext
				},
				isOfflineMigratingMonster: func(obj *server.Object) bool {
					events = append(events, "migrating:"+labels[obj])
					return tc.migrating[obj]
				},
				isCoopPlayerPixie: func(obj *server.Object) bool {
					events = append(events, "pixie:"+labels[obj])
					return tc.playerPixie[obj]
				},
				delayedDelete: func(obj *server.Object) {
					events = append(events, "delete:"+labels[obj])
					obj.ObjNext = nil
				},
			}

			cleanupObjectsForMapLoad_4E5BF0(tc.mode, hooks)
			if fmt.Sprint(events) != fmt.Sprint(tc.wantEvents) {
				t.Fatalf("events: got %v, want %v", events, tc.wantEvents)
			}
		})
	}
}
