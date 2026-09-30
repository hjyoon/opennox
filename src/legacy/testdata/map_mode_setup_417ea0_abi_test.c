#include "../GAME1.h"
#include "../map_mode_setup_417ea0.h"

typedef int (*map_mode_setup_fn)(void);
typedef bool (*map_flag_scan_fn)(void);
typedef char (*map_flagball_setup_fn)(void);

_Static_assert(_Generic(&nox_xxx_mapInfoSetCapflag_417EA0, map_mode_setup_fn: 1, default: 0),
	"00417EA0 must retain its C int result");
_Static_assert(_Generic(&sub_417EC0, map_flag_scan_fn: 1, default: 0),
	"00417EC0 must retain its C bool result");
_Static_assert(_Generic(&nox_xxx_mapInfoSetFlagball_417F30, map_flagball_setup_fn: 1, default: 0),
	"00417F30 must retain its C char result");
_Static_assert(_Generic(&nox_xxx_mapInfoSetKotr_4180D0, map_mode_setup_fn: 1, default: 0),
	"004180D0 must retain its C int result");

int main(void) {
	return 0;
}
