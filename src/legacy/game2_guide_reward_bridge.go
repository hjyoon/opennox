package legacy

/*
#include "game2_book_native.h"
*/
import "C"

//export nox_xxx_bookGuideRewardCli_native_45D140
func nox_xxx_bookGuideRewardCli_native_45D140(guide C.int) {
	nox_xxx_bookRewardCli_499CF0(4, guide, 0)
}
