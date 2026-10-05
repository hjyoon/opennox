package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"
)

func TestBotPlayerLink4FAAC0NativeWidthAndUnchangedLayout(t *testing.T) {
	var update PlayerUpdateData
	field := reflect.TypeOf(update.Field73)
	if field.Kind() != reflect.Pointer || field != reflect.TypeOf((*MonsterUpdateData)(nil)) {
		t.Fatalf("PE32 +292 bot-monster link type = %v, want *MonsterUpdateData", field)
	}
	wantSize, wantLink, wantWall := uintptr(556), uintptr(292), uintptr(296)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantSize, wantLink, wantWall = 656, 368, 376
	}
	if unsafe.Sizeof(update) != wantSize || unsafe.Offsetof(update.Field73) != wantLink ||
		unsafe.Offsetof(update.CollisionWall) != wantWall || unsafe.Sizeof(update.Field73) != unsafe.Sizeof(uintptr(0)) {
		t.Fatalf("player layout: size=%d link=%d/%d wall=%d", unsafe.Sizeof(update),
			unsafe.Offsetof(update.Field73), unsafe.Sizeof(update.Field73), unsafe.Offsetof(update.CollisionWall))
	}
	monster := new(MonsterUpdateData)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(monster)) <= math.MaxUint32 {
		t.Fatalf("bot monster record=%p, want above 4 GiB", monster)
	}
	reflect.ValueOf(&update).Elem().FieldByName("Field73").Set(reflect.ValueOf(monster))
	if got := *(*unsafe.Pointer)(unsafe.Add(unsafe.Pointer(&update), wantLink)); got != unsafe.Pointer(monster) {
		t.Fatalf("native bot link=%p, want %p", got, monster)
	}
	runtime.KeepAlive(monster)
}
