package legacy

import (
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

const clientGuideCount45D140 = 41

type clientGuideRewardHooks45D140 struct {
	currentPlayer   func() *server.Player
	relatedGuides   func(int) []int
	setBookModes    func()
	sortGuideList   func(uint8)
	findGuidePage   func(int) (int, bool)
	hideBook        func(int)
	moveBookToPage  func(int)
	openBook        func(int)
	showGuideReward func(int)
}

func clientGuideRewardNative45D140(
	guide int,
	notify bool,
	hooks clientGuideRewardHooks45D140,
) bool {
	player := hooks.currentPlayer()
	if player == nil || guide <= 0 || guide >= len(player.BeastScrollLvl) {
		return false
	}
	player.BeastScrollLvl[guide] = 1
	for _, related := range hooks.relatedGuides(guide) {
		if related > 0 && related < len(player.BeastScrollLvl) {
			player.BeastScrollLvl[related] = 1
		}
	}

	hooks.setBookModes()
	hooks.sortGuideList(uint8(player.PlayerClass()))
	if !notify {
		return true
	}
	page, ok := hooks.findGuidePage(guide)
	if !ok {
		return true
	}
	hooks.hideBook(0)
	hooks.moveBookToPage(page)
	hooks.openBook(0)
	hooks.showGuideReward(guide)
	return true
}

func clientGuideRelationsNative45D140(guide int) []int {
	const (
		blob       = uintptr(0x587000)
		groupTable = uintptr(132124)
	)
	var related []int
	for groupIndex := uintptr(0); groupIndex < clientGuideCount45D140; groupIndex++ {
		group := *memmap.PtrPtr(blob, groupTable+4*groupIndex)
		if group == nil {
			break
		}
		groupBlob, groupOff := memmap.BlobByPtr(group)
		if groupBlob == nil || int(memmap.Uint32(groupBlob.Addr, groupOff)) != guide {
			continue
		}
		for memberIndex := uintptr(1); memberIndex < clientGuideCount45D140; memberIndex++ {
			member := int(memmap.Uint32(groupBlob.Addr, groupOff+4*memberIndex))
			if member == 0 {
				break
			}
			related = append(related, member)
		}
	}
	return related
}

func clientGuidePageNative45D140(guide int) (int, bool) {
	const (
		blob = uintptr(0x5D4594)
		list = uintptr(1046960)
	)
	for page := 0; page < clientGuideCount45D140; page++ {
		if int(memmap.Uint32(blob, list+4*uintptr(page))) == guide {
			return page, true
		}
	}
	return 0, false
}
