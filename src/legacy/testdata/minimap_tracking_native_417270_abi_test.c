#include <stdint.h>

#include "../GAME1.h"

typedef int (*minimap_count_fn)(int);
typedef int (*minimap_tracks_fn)(int, nox_object_t*);

_Static_assert(sizeof(void*) == sizeof(uintptr_t),
	"minimap objects must remain native-width");
_Static_assert(_Generic(&sub_417270, minimap_count_fn: 1, default: 0),
	"00417270 must retain its player-index ABI");
_Static_assert(_Generic(&nox_xxx_playerMapTracksObj_4173D0, minimap_tracks_fn: 1, default: 0),
	"004173D0 must accept a native-width object pointer");

int main(void) {
	return 0;
}
