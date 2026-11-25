#include "GAME3_2.h"
#include "GAME3_3.h"
#include "GAME4.h"
#include "GAME5.h"

//----- (0054DFA0) --------------------------------------------------------
// This is the only direct GAME.EXE caller of 004E7470. Keeping the object and
// position as native pointers prevents the decompiler's int temporaries from
// truncating the live path on 64-bit hosts.
void nox_xxx_dieBarrel_54DFA0(nox_object_t* obj) {
	int breaking_type = *getMemU32Ptr(0x5D4594, 2491696);
	if (!breaking_type) {
		breaking_type = nox_xxx_getNameId_4E3AA0("BarrelBreaking");
		*getMemU32Ptr(0x5D4594, 2491696) = breaking_type;
	}
	nox_object_t* effect = nox_xxx_newObjectWithTypeInd_4E3450(breaking_type);
	if (effect) {
		nox_xxx_createAt_4DAA50(effect, NULL, obj->x, obj->y);
	}
	nox_xxx_aud_501960(286, obj, 0, 0);
	nox_xxx_spawnSomeBarrel_4E7470(obj, (float2*)(void*)&obj->x);
	nox_xxx_delayedDeleteObject_4E5CC0(obj);
}
