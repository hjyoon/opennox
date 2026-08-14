package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestChakramRetargetNative4EB250UsesNativeUpdateAndMapObjects(t *testing.T) {
	owner := &Object{}
	last := &Object{}
	update := &ChakramUpdateData{LastHit: last}
	source := &Object{
		ObjOwner:   owner,
		PosVec:     types.Pointf{X: 10, Y: 20},
		SpeedCur:   11,
		UpdateData: unsafe.Pointer(update),
	}
	excludedOwner := owner
	excludedOwner.ObjClass = object.ClassPlayer
	excludedLast := last
	excludedLast.ObjClass = object.ClassMonster
	target := &Object{ObjClass: object.ClassPlayer, PosVec: types.Pointf{X: 13, Y: 24}}

	var gotRect types.Rectf
	var mapPairs [][2]*Object
	got := chakramRetargetNative4EB250(source, chakramRetargetNativeDeps4EB250{
		eachInRect: func(rect types.Rectf, callback func(*Object)) {
			gotRect = rect
			callback(excludedOwner)
			callback(excludedLast)
			callback(target)
		},
		mapCheck: func(candidate, gotSource *Object) bool {
			mapPairs = append(mapPairs, [2]*Object{candidate, gotSource})
			return true
		},
	})
	if got != target {
		t.Fatalf("target = %p, want %p", got, target)
	}
	wantRect := types.Rectf{
		Min: types.Pointf{X: -390, Y: -380},
		Max: types.Pointf{X: 410, Y: 420},
	}
	if gotRect != wantRect {
		t.Fatalf("rect = %+v, want %+v", gotRect, wantRect)
	}
	if len(mapPairs) != 1 || mapPairs[0] != [2]*Object{target, source} {
		t.Fatalf("map pairs = %v, want target/source only", mapPairs)
	}
	if update.ReturnState != chakramReturnStateSeek4EAF00 {
		t.Fatalf("state = %d, want seek", update.ReturnState)
	}
	if source.VelVec.X == 0 || source.VelVec.Y == 0 {
		t.Fatalf("velocity = %+v, want both components", source.VelVec)
	}
}
