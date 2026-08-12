package legacy

import (
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

func TestObjectIsOfflineMigratingMonster4E5B50(t *testing.T) {
	tests := []struct {
		name       string
		obj        *server.Object
		gameOnline bool
		want       bool
	}{
		{name: "online short-circuits before object access", gameOnline: true},
		{
			name:       "online rejects matching monster",
			obj:        &server.Object{ObjClass: object.ClassMonster, ObjSubClass: object.SubClass(object.MonsterMigrate)},
			gameOnline: true,
		},
		{
			name: "migrate subclass without monster class",
			obj:  &server.Object{ObjClass: object.ClassPlayer, ObjSubClass: object.SubClass(object.MonsterMigrate)},
		},
		{
			name: "monster without migrate subclass",
			obj:  &server.Object{ObjClass: object.ClassMonster, ObjSubClass: object.SubClass(object.MonsterNPC)},
		},
		{
			name: "exact migrating monster",
			obj:  &server.Object{ObjClass: object.ClassMonster, ObjSubClass: object.SubClass(object.MonsterMigrate)},
			want: true,
		},
		{
			name: "bit tests accept additional class and subclass flags",
			obj: &server.Object{
				ObjClass:    object.ClassMonster | object.ClassClientPersist,
				ObjSubClass: object.SubClass(object.MonsterMigrate | object.MonsterNPC),
			},
			want: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := objectIsOfflineMigratingMonster_4E5B50(tc.obj, tc.gameOnline); got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}
