#include "client__gui__servopts__advserv.h"
#include <string.h>

#include "GAME1.h"
#include "client__gui__window.h"
#include "common__strman.h"
extern nox_window* dword_5d4594_1316972;

//----- (004BE2C0) --------------------------------------------------------
int sub_4BE2C0(int a1) {
	wchar2_t* v1;  // eax
	nox_window* v2; // eax

	uint32_t value = (uint32_t)a1;
	memcpy((uint8_t*)sub_416640() + 74, &value, sizeof(value));
	v1 = nox_strman_loadString_40F1D0("AudCullDesc", 0, "C:\\NoxPost\\src\\client\\Gui\\ServOpts\\advserv.c", 71);
	nox_swprintf((wchar2_t*)getMemAt(0x5D4594, 1316716), v1, a1);
	v2 = nox_xxx_wndGetChildByID_46B0C0(dword_5d4594_1316972, 2120);
	return (int)nox_window_call_field_94(v2, 16385, (uintptr_t)getMemAt(0x5D4594, 1316716), (uintptr_t)-1);
}
