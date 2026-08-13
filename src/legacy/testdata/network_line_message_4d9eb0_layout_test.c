#include "../GAME3_2.h"

#ifdef NOX_LINE_MESSAGE_4D9EB0_NATIVE_LAYOUT
// Native probes suppress unrelated Win32-only assertions while defs.h is
// parsed, then re-enable the exact object fields consumed by 004D9EB0.
#undef _Static_assert
_Static_assert(offsetof(nox_object_t, obj_class) == (sizeof(void*) == 4 ? 8 : 12), "object class offset");
_Static_assert(offsetof(nox_object_t, data_update) == (sizeof(void*) == 4 ? 748 : 872), "object update data offset");

static intptr_t (*const nox_line_message_signature_4d9eb0)(nox_object_t*, wchar2_t*, ...) =
	nox_xxx_netSendLineMessage_4D9EB0;
static uint8_t (*const nox_player_index_signature_4d9eb0)(void*) =
	nox_server_playerIndexFromUpdateData_4D9EB0;
#endif
