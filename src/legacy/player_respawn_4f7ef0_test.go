package legacy

import (
	"testing"

	"github.com/opennox/opennox/v1/common/memmap"
)

func TestRespawnPlayerCorpseTables53FBC0(t *testing.T) {
	const (
		direction = int32(3)
		part      = 7
		wantType  = uint32(0x89abcdef)
		wantX     = float32(123.25)
		wantY     = float32(-456.5)
	)

	typeOffset := uintptr(playerRespawnCorpseTypesOffset53FBC0 +
		playerRespawnCorpseTypeSize53FBC0*int(direction) + 4*part)
	pointOffset := uintptr(playerRespawnCorpsePointsOff53FBC0 +
		playerRespawnCorpseDirectionSize53FBC0*int(direction) + 8*part)

	typeValue := memmap.PtrUint32(playerRespawnSettingsBase4F7EF0, typeOffset)
	xValue := memmap.PtrFloat32(playerRespawnCorpsePointsBase53FBC0, pointOffset)
	yValue := memmap.PtrFloat32(playerRespawnCorpsePointsBase53FBC0, pointOffset+4)
	oldType, oldX, oldY := *typeValue, *xValue, *yValue
	t.Cleanup(func() {
		*typeValue, *xValue, *yValue = oldType, oldX, oldY
	})

	*typeValue, *xValue, *yValue = wantType, wantX, wantY
	if got := respawnPlayerTypeIndex53FBC0(direction, part); got != wantType {
		t.Fatalf("corpse type = %#x, want %#x", got, wantType)
	}
	if got := respawnPlayerOffset53FBC0(direction, part); got.X != wantX || got.Y != wantY {
		t.Fatalf("corpse offset = %+v, want {%v %v}", got, wantX, wantY)
	}
}
