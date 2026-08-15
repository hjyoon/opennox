#ifndef NOX_UNDEAD_KILLER_COLLIDE_4EBD40_H
#define NOX_UNDEAD_KILLER_COLLIDE_4EBD40_H

#include <stddef.h>

typedef struct nox_object_t nox_object_t;

// Native-pointer representation of GAME.EXE's one-word collision record.
// TurnUndead stores a server.DurSpell pointer in this sole field.
typedef struct nox_undead_killer_collide_data_t {
	void* spell;
} nox_undead_killer_collide_data_t;

_Static_assert(offsetof(nox_undead_killer_collide_data_t, spell) == 0,
	"wrong offset of UndeadKiller collide spell!");
_Static_assert(sizeof(nox_undead_killer_collide_data_t) == sizeof(void*),
	"wrong size of UndeadKiller collide data!");

void nox_xxx_collideUndeadKiller_4EBD40(
	nox_object_t* source,
	nox_object_t* target,
	float* collision);

#endif // NOX_UNDEAD_KILLER_COLLIDE_4EBD40_H
