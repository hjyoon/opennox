#include "GAME2.h"
#include "common__strman.h"

//----- (004AA650) --------------------------------------------------------
void sub_4AA650() {
	uint32_t* counter;
	uint32_t index;
	char* v2; // ecx
	void* v3; // [esp+0h] [ebp-4h]

	if (!sub_44D930()) {
		counter = getMemU32Ptr(0x5D4594, 1309744);
		index = *counter;
		v2 = (char*)getMemPtr(0x587000, 172892 + 4 * index);
		*counter = index + 1;
		nox_strman_loadString_40F1D0(v2, &v3, "C:\\NoxPost\\src\\client\\shell\\Options.c", 131);
		*counter = (uint32_t)((int32_t)*counter % 3);
		if (v3) {
			nox_xxx_playDialogFile_44D900((unsigned char*)v3, 100);
		}
	}
}
