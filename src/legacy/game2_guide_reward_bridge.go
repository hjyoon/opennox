package legacy

/*
#include "game2_book_native.h"
*/
import "C"

// Keep the reward-kind word in C across the historical pointer/tag ABI.
//
//export nox_xxx_bookGuideRewardCli_native_45D140
func nox_xxx_bookGuideRewardCli_native_45D140(guide C.int) {
	C.game2_book_reward_scalar(4, guide, 0)
}
