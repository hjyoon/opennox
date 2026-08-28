#include "../object_read_old_4f4170.h"

typedef int32_t (*object_read_old_fn_4F4170)(nox_object_t*, int32_t, int32_t);

_Static_assert(_Generic(&nox_xxx_readObjectOldVer_4F4170, object_read_old_fn_4F4170: 1, default: 0),
	"old-version object loader must preserve its native object pointer and exact int32 scalars");

static object_read_old_fn_4F4170 object_read_old_symbol_4F4170 =
	nox_xxx_readObjectOldVer_4F4170;

int main(void) { return object_read_old_symbol_4F4170 == 0; }
