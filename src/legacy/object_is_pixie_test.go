package legacy

import (
	"fmt"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

func TestObjectIsCoopPlayerPixie4E5B80MatchesGAMEEXEContract(t *testing.T) {
	const pixieTypeInd = 77
	player := &server.Object{ObjClass: object.ClassPlayer}
	intermediate := &server.Object{ObjClass: object.ClassMonster, ObjOwner: player}
	playerPixie := &server.Object{
		TypeInd:  pixieTypeInd,
		ObjClass: object.ClassMissile,
		ObjOwner: intermediate,
		ObjFlags: object.FlagDestroyed,
		Field1_2: 0xCAFE,
		Field37:  0xDEADBEEF,
		Field38:  0x12345678,
	}

	tests := []struct {
		name           string
		obj            *server.Object
		coop           bool
		forceNilParent bool
		want           bool
		wantEvents     []string
	}{
		{
			name:       "pixie lookup precedes nil rejection",
			wantEvents: []string{"pixie"},
		},
		{
			name:       "non missile rejects before mode check",
			obj:        &server.Object{TypeInd: pixieTypeInd, ObjClass: object.ClassMonster, ObjOwner: player},
			coop:       true,
			wantEvents: []string{"pixie"},
		},
		{
			name:       "non coop rejects before type and parent",
			obj:        playerPixie,
			wantEvents: []string{"pixie", "coop"},
		},
		{
			name:       "wrong type rejects before parent lookup",
			obj:        &server.Object{TypeInd: pixieTypeInd - 1, ObjClass: object.ClassMissile, ObjOwner: player},
			coop:       true,
			wantEvents: []string{"pixie", "coop"},
		},
		{
			name:           "nil parent result rejects",
			obj:            playerPixie,
			coop:           true,
			forceNilParent: true,
			wantEvents:     []string{"pixie", "coop", "parent"},
		},
		{
			name: "owner chain without player rejects",
			obj: &server.Object{
				TypeInd:  pixieTypeInd,
				ObjClass: object.ClassMissile,
				ObjOwner: &server.Object{ObjClass: object.ClassMonster},
			},
			coop:       true,
			wantEvents: []string{"pixie", "coop", "parent"},
		},
		{
			name:       "nested player owner accepts",
			obj:        playerPixie,
			coop:       true,
			want:       true,
			wantEvents: []string{"pixie", "coop", "parent"},
		},
		{
			name: "bit tests accept additional class flags",
			obj: &server.Object{
				TypeInd:  pixieTypeInd,
				ObjClass: object.ClassMissile | object.ClassSimple,
				ObjOwner: &server.Object{ObjClass: object.ClassPlayer | object.ClassMonster},
			},
			coop:       true,
			want:       true,
			wantEvents: []string{"pixie", "coop", "parent"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			hooks := objectIsCoopPlayerPixieHooks{
				pixieTypeInd: func() int {
					events = append(events, "pixie")
					return pixieTypeInd
				},
				gameCoop: func() bool {
					events = append(events, "coop")
					return tc.coop
				},
				findParentChainPlayer: func(obj *server.Object) *server.Object {
					events = append(events, "parent")
					if tc.forceNilParent {
						return nil
					}
					return obj.FindOwnerChainPlayer()
				},
			}

			var before struct {
				typeInd  uint16
				class    object.Class
				owner    *server.Object
				flags    object.Flags
				field1_2 uint16
				field37  uint32
				field38  uint32
			}
			if tc.obj != nil {
				before.typeInd = tc.obj.TypeInd
				before.class = tc.obj.ObjClass
				before.owner = tc.obj.ObjOwner
				before.flags = tc.obj.ObjFlags
				before.field1_2 = tc.obj.Field1_2
				before.field37 = tc.obj.Field37
				before.field38 = tc.obj.Field38
			}
			if got := objectIsCoopPlayerPixie_4E5B80(tc.obj, hooks); got != tc.want {
				t.Fatalf("result: got %v, want %v", got, tc.want)
			}
			if fmt.Sprint(events) != fmt.Sprint(tc.wantEvents) {
				t.Fatalf("events: got %v, want %v", events, tc.wantEvents)
			}
			if tc.obj != nil && (tc.obj.TypeInd != before.typeInd || tc.obj.ObjClass != before.class ||
				tc.obj.ObjOwner != before.owner || tc.obj.ObjFlags != before.flags ||
				tc.obj.Field1_2 != before.field1_2 || tc.obj.Field37 != before.field37 ||
				tc.obj.Field38 != before.field38) {
				t.Fatal("predicate changed observed input fields")
			}
		})
	}
}
