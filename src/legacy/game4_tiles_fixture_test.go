package legacy

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/internal/cfixture"
)

type game4TilesTestBank struct {
	*cfixture.Game4TilesBank
	releaseMemory func()
}

func game4TilesTestBegin() *game4TilesTestBank {
	release := game4SupportHighFixtureMemory()
	bank := cfixture.NewGame4TilesBank()
	if bank == nil {
		release()
		return nil
	}
	return &game4TilesTestBank{Game4TilesBank: bank, releaseMemory: release}
}

func (b *game4TilesTestBank) close() {
	b.Close()
	b.releaseMemory()
}

func game4TilesTestServices() game4TilesServices {
	controls := cfixture.Game4TileReferenceControls()
	var services game4TilesServices
	return game4TilesServices{
		snapshot:  controls.Snapshot,
		restore:   controls.Restore,
		image:     controls.Image,
		variation: controls.Variation,
		mode:      controls.Mode,
		base:      game4TilesPointCallback(services.base, controls.Base),
		byte3:     controls.Byte3,
		byte4:     controls.Byte4,
		subtile:   game4TilesPointCallback(services.subtile, controls.Subtile),
	}
}

func game4TilesPointCallback[T any](_ func(*T), callback func(unsafe.Pointer)) func(*T) {
	return func(p *T) { callback(unsafe.Pointer(p)) }
}

func game4TilesTestApplyBrush(x, y, layer, merge int32, brush *mapgenFloorBrush) int32 {
	return cfixture.Game4TilesApplyBrush(x, y, layer, merge, unsafe.Pointer(brush))
}

func game4TilesTestApplyBrushToken(x, y, layer int32, brush *mapgenFloorBrush) int32 {
	return cfixture.Game4TilesApplyBrushToken(x, y, layer, unsafe.Pointer(brush))
}

func game4TilesTestApplyFlood(point *[2]int32, brush *mapgenFloorBrush) int32 {
	return cfixture.Game4TilesApplyFlood(point, unsafe.Pointer(brush))
}
