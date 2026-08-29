// Compile-only native-width contract for the script-handler transfer restored
// from GAME.EXE 004F5580.
#include "../server__script__script.h"
#include "../GAME4.h"

#include <stddef.h>
#include <stdint.h>

_Static_assert(sizeof(nox_script_callback_t) == 8,
	"script callback size");
_Static_assert(offsetof(nox_script_callback_t, flags) == 0,
	"script callback flags offset");
_Static_assert(offsetof(nox_script_callback_t, func) == 4,
	"script callback function offset");

typedef int32_t (*script_handler_xfer_fn_4F5580)(
	nox_script_callback_t*,
	char*);

_Static_assert(
	_Generic(&nox_xxx_xferReadScriptHandler_4F5580,
		script_handler_xfer_fn_4F5580: 1, default: 0),
	"script-handler transfer signature");

static script_handler_xfer_fn_4F5580 const script_handler_xfer_signature =
	nox_xxx_xferReadScriptHandler_4F5580;

typedef int (*monster_generator_xfer_fn_4F7130)(nox_object_t*);

_Static_assert(offsetof(nox_object_t, field_34) ==
	(sizeof(void*) == 4 ? 136 : 140),
	"MonsterGenerator object field-34 offset");
_Static_assert(offsetof(nox_object_t, data_update) ==
	(sizeof(void*) == 4 ? 748 : 872),
	"MonsterGenerator object update-data offset");
_Static_assert(offsetof(nox_object_t, field_189) ==
	(sizeof(void*) == 4 ? 756 : 888),
	"MonsterGenerator object script-context offset");
_Static_assert(
	_Generic(&nox_xxx_XFerMonsterGen_4F7130,
		monster_generator_xfer_fn_4F7130: 1, default: 0),
	"MonsterGeneratorXfer signature");

static monster_generator_xfer_fn_4F7130 const monster_generator_xfer_signature =
	nox_xxx_XFerMonsterGen_4F7130;

int main(void) {
	return script_handler_xfer_signature == NULL ||
		monster_generator_xfer_signature == NULL;
}
