package server

import (
	"testing"

	"github.com/opennox/libs/object"
)

func TestPlayerDamageGreatSwordMissile4E17B0AlreadyPrefixedNPC(t *testing.T) {
	v, a, w, m, _, ctx, r, _ := damageGreatSwordFixture4E17B0(t, false, false)
	ctx.prefixed = true
	*ctx.marker, *ctx.markerType = 1, 777
	// Availability admission is a callback too. It may change the LIVE
	// projectile type, but must not replay the preceding cached hit marker.
	r.Melee.CanDamageBlockWeapon = func(*Object) bool { m.TypeInd = 999; return true }
	r.ProjectileReflect = func(*Object, *Object) {
		if *ctx.marker != 1 || *ctx.markerType != 777 {
			t.Fatalf("prefix replayed: %d/%d", *ctx.marker, *ctx.markerType)
		}
	}
	if applicable, h, ok := playerDamageGreatSwordMissileBlock4E17B0(v, a, w, ctx, 9, object.DamageFlame, r); !applicable || !h || ok {
		t.Fatal("prefixed NPC block failed")
	}
}
