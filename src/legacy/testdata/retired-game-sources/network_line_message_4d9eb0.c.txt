#include "GAME1.h"
#include "GAME3_2.h"
#include "common__net_list.h"
#include "noxstring.h"
#include "operators.h"

//----- (004D9EB0) --------------------------------------------------------
intptr_t nox_xxx_netSendLineMessage_4D9EB0(nox_object_t* obj, wchar2_t* format, ...) {
	intptr_t result;   // eax
	void* update_data; // esi
	char flags;        // al
	int width;         // eax
	wchar2_t buf[516]; // [esp+4h] [ebp-408h]
	va_list va;        // [esp+418h] [ebp+Ch]

	va_start(va, format);
	result = (intptr_t)obj;
	if (obj && obj->obj_class & 4) {
		update_data = obj->data_update;
		nox_vswprintf(&buf[260], format, va);
		LOBYTE(buf[0]) = -88; // MSG_TEXT_MESSAGE
		*(wchar2_t*)((char*)buf + 1) = 0;
		HIBYTE(buf[1]) = 0;
		if (nox_xxx_cliCanTalkMB_4100F0((short*)&buf[260])) {
			flags = HIBYTE(buf[1]) | 2;
		} else {
			flags = HIBYTE(buf[1]) | 4;
		}
		HIBYTE(buf[1]) = flags;
		buf[2] = 0;
		buf[3] = 0;
		LOBYTE(buf[5]) = 0;
		buf[4] = (unsigned char)(nox_wcslen(&buf[260]) + 1);
		if (buf[1] & 0x400) {
			nox_wcscpy((wchar2_t*)((char*)&buf[5] + 1), &buf[260]);
			width = 2;
		} else {
			nox_sprintf((char*)&buf[5] + 1, "%S", &buf[260]);
			width = 1;
		}
		result = nox_netlist_addToMsgListCli_40EBC0(
			nox_server_playerIndexFromUpdateData_4D9EB0(update_data), 1, (unsigned char*)buf,
			width * LOBYTE(buf[4]) + 11);
	}
	va_end(va);
	return result;
}
