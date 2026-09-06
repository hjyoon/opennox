package legacy

import (
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

func TestNPCWeaponDequipItemEligible53A030(t *testing.T) {
	tests := []struct {
		name string
		item *server.Object
		want bool
	}{
		{
			name: "sword",
			item: &server.Object{ObjClass: object.ClassWeapon, ObjSubClass: object.SubClass(0x100), ObjFlags: object.FlagEquipped},
			want: true,
		},
		{
			name: "bow",
			item: &server.Object{ObjClass: object.ClassWeapon, ObjSubClass: object.SubClass(0x4), ObjFlags: object.FlagEquipped},
			want: true,
		},
		{
			name: "wand",
			item: &server.Object{ObjClass: object.ClassWand, ObjFlags: object.FlagEquipped},
			want: true,
		},
		{
			name: "unequipped sword",
			item: &server.Object{ObjClass: object.ClassWeapon, ObjSubClass: object.SubClass(0x100)},
		},
		{
			name: "equipped non-weapon",
			item: &server.Object{ObjClass: object.ClassFood, ObjFlags: object.FlagEquipped},
		},
		{name: "nil"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := npcWeaponDequipItemEligible53A030(tc.item); got != tc.want {
				t.Fatalf("eligible = %v, want %v", got, tc.want)
			}
		})
	}
}
