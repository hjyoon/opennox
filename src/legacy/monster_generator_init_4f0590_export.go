package legacy

/*
#include "monster_generator_init_4f0590.h"

int nox_xxx_getQuestStage_51A930(void);
*/
import "C"

//export nox_xxx_unitInitGenerator_4F0590
func nox_xxx_unitInitGenerator_4F0590(unit *C.nox_object_t) C.int32_t {
	result := monsterGeneratorInitCall4F0590(asObjectS((*nox_object_t)(unit)), func() uint32 {
		return uint32(C.nox_xxx_getQuestStage_51A930())
	})
	return C.int32_t(result)
}
