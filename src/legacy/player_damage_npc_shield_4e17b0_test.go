package legacy

import (
	"fmt"
	"testing"

	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"

	"github.com/opennox/opennox/v1/server"
)

func TestPlayerDamageNPCShieldSourceExclusions4E17B0(t *testing.T) {
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	// Missing IDs must not classify a zero-type object as an excluded spell.
	if playerDamageBlockSourceExcluded4E17B0(srv, &server.Object{}, true) ||
		playerDamageBlockSourceExcluded4E17B0(srv, &server.Object{}, false) {
		t.Fatal("missing spell IDs excluded a zero-type object")
	}
	names := []string{"SmallFist", "MediumFist", "LargeFist", "Meteor", "ToxicCloud", "SmallToxicCloud", "Other"}
	for _, name := range names {
		if err := srv.Types.ReadObjectType(&things.Thing{Name: name}); err != nil {
			t.Fatal(err)
		}
	}
	for i, name := range names {
		ind := srv.Types.IndByID(name)
		if ind == 0 {
			t.Fatalf("synthetic type %q was not loaded", name)
		}
		attack := &server.Object{TypeInd: uint16(ind)}
		for _, hasWeapon := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/weapon-%t", name, hasWeapon), func(t *testing.T) {
				want := i < 4 || (hasWeapon && i < 6)
				if got := playerDamageBlockSourceExcluded4E17B0(srv, attack, hasWeapon); got != want {
					t.Fatalf("production exclusion=%t, want %t", got, want)
				}
			})
		}
	}
}
