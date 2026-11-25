package legacy

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/internal/cfixture"
)

func init() {
	cfixture.Game2ReferenceAllocate = g2ServerAllocate
	cfixture.Game2ReferenceRelease = g2ServerRelease
}

var (
	game2ServerFixtureOwner        = cfixture.Game2Owner
	game2ServerFixtureSetOwner     = cfixture.Game2SetOwner
	game2ServerFixtureLayout       = cfixture.Game2Layout
	game2ServerFixtureAllocateHigh = cfixture.Game2AllocateHigh
	game2ServerFixtureAlloc        = cfixture.Game2Alloc
	game2ServerFixtureFree         = cfixture.Game2Free
	game2ServerFixtureFirstVisible = cfixture.Game2FirstVisible
	game2ServerFixtureTeam         = cfixture.Game2Team
	game2ServerFixtureAccessEvent  = cfixture.Game2AccessEvent
	game2ServerFixtureDrawable     = cfixture.Game2Drawable
	game2ServerFixtureNoRender     = cfixture.Game2NoRender
	game2ServerFixtureHUD          = cfixture.Game2HUD
	game2ServerFixtureControls     = cfixture.Game2Controls
)

func game2ServerFixtureCall(old bool, entry uint32, p unsafe.Pointer, a, b int32) uintptr {
	previousAlloc, previousFree := game2ServerAllocationService, game2ServerReleaseService
	game2ServerAllocationService, game2ServerReleaseService = game2ServerFixtureAllocateHigh, game2ServerFixtureFree
	defer func() { game2ServerAllocationService, game2ServerReleaseService = previousAlloc, previousFree }()
	return cfixture.Game2Call(old, entry, p, a, b)
}

func game2ServerFixtureHead(team bool) *accessListNode416720 {
	return (*accessListNode416720)(cfixture.Game2Head(team))
}

func game2ServerFixtureClearHead(team bool) {
	h := game2ServerFixtureHead(team)
	for p := accessListNativeNext416720(h); p != nil; {
		next := accessListNativeNext416720(p)
		accessListNativeRemove416720(p)
		game2ServerFixtureFree(unsafe.Pointer(p))
		p = next
	}
	accessListNativeClear416720(h)
}
