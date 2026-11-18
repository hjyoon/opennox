package legacy

/*
#include "GAME4_3.h"

typedef int (*monster_spell_enchant_query)(nox_object_t*, int);
typedef int (*monster_spell_id_query)(int);
typedef int (*monster_spell_owner_query)(nox_object_t*);

_Static_assert(_Generic(&sub_540CE0, monster_spell_enchant_query: 1, default: 0), "spell enchant ABI");
_Static_assert(_Generic(&sub_540D20, monster_spell_id_query: 1, default: 0), "summon spell ID ABI");
_Static_assert(_Generic(&sub_540D40, monster_spell_owner_query: 1, default: 0), "summon owner ABI");

static int monster_spell_enchant_fixture(nox_object_t* unit, int id) {
	monster_spell_enchant_query volatile query = sub_540CE0;
	return query(unit, id);
}

static int monster_spell_id_fixture(int id) {
	monster_spell_id_query volatile query = sub_540D20;
	return query(id);
}

static int monster_spell_owner_fixture(nox_object_t* unit) {
	monster_spell_owner_query volatile query = sub_540D40;
	return query(unit);
}
*/
import "C"

import "github.com/opennox/opennox/v1/server"

func monsterSpellEnchantFixture540CE0(unit *server.Object, id int32) int {
	return int(C.monster_spell_enchant_fixture(asObjectC(unit), C.int(id)))
}

func monsterSpellIDFixture540D20(id int32) int {
	return int(C.monster_spell_id_fixture(C.int(id)))
}

func monsterSpellOwnerFixture540D40(unit *server.Object) int {
	return int(C.monster_spell_owner_fixture(asObjectC(unit)))
}
