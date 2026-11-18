package legacy

/*
#include "GAME4_3.h"

typedef int (*bow_shoot_fn)(nox_object_t*, nox_object_t*);
typedef void (*bow_projectile_fn)(nox_object_t*, nox_object_t*, nox_object_t*, uint32_t);
typedef int (*bow_effect_fn)(nox_object_t*, nox_object_t*, nox_object_t*);

_Static_assert(_Generic(&nox_xxx_shootBowCrossbow1_539BD0, bow_shoot_fn: 1, default: 0), "bow shoot ABI");
_Static_assert(_Generic(&nox_xxx_shootBowCrossbow2_539D80, bow_projectile_fn: 1, default: 0), "bow projectile ABI");
_Static_assert(_Generic(&nox_xxx_shootApplyEffects_539F40, bow_effect_fn: 1, default: 0), "bow effects ABI");

static int bow_shoot_native_fixture(nox_object_t* owner, nox_object_t* weapon) {
	bow_shoot_fn volatile shoot = nox_xxx_shootBowCrossbow1_539BD0;
	return shoot(owner, weapon);
}

static void bow_projectile_native_fixture(nox_object_t* owner, nox_object_t* quiver, nox_object_t* weapon, uint32_t flags) {
	bow_projectile_fn volatile shoot = nox_xxx_shootBowCrossbow2_539D80;
	shoot(owner, quiver, weapon, flags);
}

static int bow_effect_native_fixture(nox_object_t* owner, nox_object_t* weapon, nox_object_t* projectile) {
	bow_effect_fn volatile apply = nox_xxx_shootApplyEffects_539F40;
	return apply(owner, weapon, projectile);
}
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func bowShootStockFixture539BD0(owner, weapon *server.Object) int {
	return int(C.bow_shoot_native_fixture(asObjectC(owner), asObjectC(weapon)))
}

func bowProjectileStockFixture539D80(owner, quiver, weapon *server.Object, flags uint32) {
	C.bow_projectile_native_fixture(asObjectC(owner), asObjectC(quiver), asObjectC(weapon), C.uint32_t(flags))
}

func bowEffectsStockFixture539F40(owner, weapon, projectile *server.Object) int {
	return int(C.bow_effect_native_fixture(asObjectC(owner), asObjectC(weapon), asObjectC(projectile)))
}
