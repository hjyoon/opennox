#ifndef NOX_PORT_SPELL_ACCEPT_4FD400_H
#define NOX_PORT_SPELL_ACCEPT_4FD400_H

#include <stddef.h>
#include <stdint.h>

typedef struct nox_object_t nox_object_t;

typedef struct nox_spell_accept_arg_t {
	nox_object_t* obj;
	float x;
	float y;
} nox_spell_accept_arg_t;

_Static_assert(offsetof(nox_spell_accept_arg_t, obj) == 0,
	"wrong offset of SpellAcceptArg object");
_Static_assert(offsetof(nox_spell_accept_arg_t, x) == sizeof(void*),
	"wrong offset of SpellAcceptArg X coordinate");
_Static_assert(offsetof(nox_spell_accept_arg_t, y) == sizeof(void*) + sizeof(float),
	"wrong offset of SpellAcceptArg Y coordinate");
_Static_assert(sizeof(nox_spell_accept_arg_t) == sizeof(void*) + 2 * sizeof(float),
	"wrong native size of SpellAcceptArg");

int32_t nox_xxx_spellAccept_4FD400(
	int32_t spell_id,
	nox_object_t* second,
	nox_object_t* third,
	nox_object_t* fourth,
	nox_spell_accept_arg_t* arg,
	int32_t level);

#endif // NOX_PORT_SPELL_ACCEPT_4FD400_H
