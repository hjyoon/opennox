package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
)

func TestChakramUpdateNative53DCC0UsesNativeObjectAndUpdateFields(t *testing.T) {
	item := &Object{}
	owner := &Object{PosVec: types.Pointf{X: 13, Y: 24}}
	update := &ChakramUpdateData{}
	source := &Object{
		InvFirstItem: item,
		ObjOwner:     owner,
		PosVec:       types.Pointf{X: 10, Y: 20},
		SpeedCur:     11,
		Field32:      100,
		UpdateData:   unsafe.Pointer(update),
	}
	var mapSource, mapOwner *Object
	chakramUpdateNative53DCC0(source, chakramUpdateNativeDeps53DCC0{
		mapCheck: func(gotSource, gotOwner *Object) bool {
			mapSource, mapOwner = gotSource, gotOwner
			return false
		},
		frame:         func() uint32 { return 120 },
		frameRate:     func() uint32 { return 30 },
		delayedDelete: func(*Object) { t.Fatal("unexpected delete") },
	})
	if mapSource != source || mapOwner != owner {
		t.Fatalf("map args = (%p, %p), want (%p, %p)", mapSource, mapOwner, source, owner)
	}
	if update.OwnerPos != owner.PosVec || update.ReturnTarget != nil || update.ReturnState != 0 {
		t.Fatalf("update = %+v, want owner snapshot and nil target", update)
	}
	if source.VelVec.X == 0 || source.VelVec.Y == 0 {
		t.Fatalf("velocity = %+v, want both components", source.VelVec)
	}
}

func TestChakramUpdateNative53DCC0InventoryFailureDeletesNativeSource(t *testing.T) {
	update := &ChakramUpdateData{}
	source := &Object{UpdateData: unsafe.Pointer(update)}
	var deleted *Object
	chakramUpdateNative53DCC0(source, chakramUpdateNativeDeps53DCC0{
		delayedDelete: func(obj *Object) { deleted = obj },
	})
	if deleted != source {
		t.Fatalf("deleted = %p, want %p", deleted, source)
	}
}
