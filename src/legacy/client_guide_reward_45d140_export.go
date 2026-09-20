package legacy

/*
#include <stdint.h>

#include "GAME2.h"

extern uint32_t dword_5d4594_1046868;
extern uint32_t dword_5d4594_1046864;
extern uint32_t dword_5d4594_1046872;
*/
import "C"

var clientGuideRewardCall45D140 = func(guide, notify int32) {
	clientGuideRewardNative45D140(
		int(guide),
		notify != 0,
		clientGuideRewardHooks45D140{
			currentPlayer: Get_dword_8531A0_2576,
			relatedGuides: clientGuideRelationsNative45D140,
			setBookModes: func() {
				C.dword_5d4594_1046868 = 1
				C.dword_5d4594_1046872 = 1
			},
			sortGuideList: func(class uint8) {
				C.nox_xxx_guiSpellSortList_45ADF0(C.int(class))
			},
			findGuidePage: clientGuidePageNative45D140,
			hideBook: func(value int) {
				C.nox_xxx_bookHideMB_45ACA0(C.int(value))
			},
			moveBookToPage: func(page int) {
				C.nox_xxx_bookMoveToPage_45B930(C.int(page))
			},
			openBook: func(value int) {
				C.nox_xxx_book_45B010(C.int(value))
			},
			showGuideReward: func(guide int) {
				C.nox_xxx_bookGuideRewardCli_native_45D140(C.int(guide))
			},
		},
	)
}

func clientGuideRewardExportCall45D140(guide, notify int32) {
	C.nox_xxx_netGuideRewardCli_45D140(C.int(guide), C.int(notify))
}

// Nox_client_guideRewardState45D140 exposes the client-only guide-book state
// needed by the headless acquisition regression. The Player pointer remains
// native width; page lookup reads the sorted guide list produced by 45D140.
func Nox_client_guideRewardState45D140(guide int) (level uint32, guideMode, bookOpen bool, page int, found bool) {
	player := Get_dword_8531A0_2576()
	if player == nil || guide <= 0 || guide >= len(player.BeastScrollLvl) {
		return 0, C.dword_5d4594_1046872 != 0, C.dword_5d4594_1046864 != 0, 0, false
	}
	page, found = clientGuidePageNative45D140(guide)
	return player.BeastScrollLvl[guide], C.dword_5d4594_1046872 != 0, C.dword_5d4594_1046864 != 0, page, found
}

//export nox_xxx_netGuideRewardCli_45D140
func nox_xxx_netGuideRewardCli_45D140(guide, notify C.int) {
	clientGuideRewardCall45D140(int32(guide), int32(notify))
}
