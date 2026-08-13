#include "../defs.h"

#ifdef NOX_DIRECTION_4E7680_NATIVE_LAYOUT
// Native 64-bit probes suppress unrelated Win32-only assertions while the
// header is parsed, then re-enable only the object fields consumed here.
#undef _Static_assert
_Static_assert(sizeof(((nox_object_t*)0)->direction1) == 2, "object direction1 width");
_Static_assert(offsetof(nox_object_t, direction1) == (sizeof(void*) == 4 ? 124 : 128), "object direction1 offset");
_Static_assert(offsetof(nox_object_t, direction2) == (sizeof(void*) == 4 ? 126 : 130), "object direction2 offset");
#endif
