#include "GAME3_2.h"
#include "noxstring.h"

intptr_t nox_send_line_message_text_4D9EB0(void* player_update, wchar2_t* text);

intptr_t nox_xxx_netSendLineMessage_4D9EB0(nox_object_t* obj, wchar2_t* format, ...) {
	if (!obj || !(obj->obj_class & UINT32_C(4))) {
		return (intptr_t)obj;
	}
	void* update = obj->data_update;
	wchar2_t text[256];
	va_list args;
	va_start(args, format);
	nox_vsnwprintf(text, 255, format, args);
	va_end(args);
	text[255] = 0;
	return nox_send_line_message_text_4D9EB0(update, text);
}
