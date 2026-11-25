#include "GAME3_2.h"
#include "noxstring.h"

intptr_t nox_send_line_message_text_4D9EB0(void* player_update, wchar2_t* text);
int nox_line_message_player_update_4D9EB0(nox_object_t* obj, uintptr_t* update);

intptr_t nox_xxx_netSendLineMessage_4D9EB0(nox_object_t* obj, wchar2_t* format, ...) {
	uintptr_t update;
	if (!nox_line_message_player_update_4D9EB0(obj, &update)) {
		return (intptr_t)obj;
	}
	wchar2_t text[256];
	va_list args;
	va_start(args, format);
	nox_vsnwprintf(text, 255, format, args);
	va_end(args);
	text[255] = 0;
	return nox_send_line_message_text_4D9EB0((void*)update, text);
}
