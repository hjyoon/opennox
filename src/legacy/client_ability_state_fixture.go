package legacy

/*
#include "GAME2_1.h"
*/
import "C"

func clientAbilityActive461160(id int) bool {
	return C.sub_461160(C.int(id)) != 0
}
