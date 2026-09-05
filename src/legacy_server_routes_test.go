package opennox

import (
	"testing"

	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func TestCollisionEnchantLegacyRoute4FDF90UsesLiveRootServer(t *testing.T) {
	oldServer := noxServer
	t.Cleanup(func() { noxServer = oldServer })
	noxServer = &Server{Server: new(server.Server)}

	if legacy.Nox_xxx_collide_4FDF90 == nil {
		t.Fatal("legacy collision-enchant callback is not registered")
	}
	legacy.Nox_xxx_collide_4FDF90(new(server.Object), new(server.Object))
}

func TestCreateSpellProjectileLegacyRoute4FDDA0UsesLiveRootServer(t *testing.T) {
	oldServer := noxServer
	t.Cleanup(func() { noxServer = oldServer })
	noxServer = &Server{Server: new(server.Server)}

	if legacy.Nox_xxx_createSpellFly_4FDDA0 == nil {
		t.Fatal("legacy spell-projectile callback is not registered")
	}
	source := &server.Object{PosVec: types.Pointf{X: -100, Y: -100}}
	if got := legacy.Nox_xxx_createSpellFly_4FDDA0(source, new(server.Object), spell.SPELL_INVALID); got != nil {
		t.Fatalf("out-of-map projectile = %p, want nil", got)
	}
}
