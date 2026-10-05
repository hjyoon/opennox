package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"
)

func TestBotMonsterLink4FAAF0NativeWidthAndUnchangedLayout(t *testing.T) {
	var update MonsterUpdateData
	field := reflect.TypeOf(update.Field545)
	if field.Kind() != reflect.Pointer || field != reflect.TypeOf((*PlayerUpdateData)(nil)) {
		t.Fatalf("PE32 +2180 bot-player link type = %v, want *PlayerUpdateData", field)
	}
	wantSize, wantLink, want546, want547, want548 := uintptr(2200), uintptr(2180), uintptr(2184), uintptr(2188), uintptr(2192)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantSize, wantLink, want546, want547, want548 = 2960, 2928, 2936, 2940, 2944
	}
	if unsafe.Sizeof(update) != wantSize || unsafe.Offsetof(update.Field545) != wantLink ||
		unsafe.Offsetof(update.Field546) != want546 || unsafe.Offsetof(update.Field547) != want547 ||
		unsafe.Offsetof(update.Field548) != want548 || unsafe.Sizeof(update.Field545) != unsafe.Sizeof(uintptr(0)) {
		t.Fatalf("monster layout: size=%d link=%d/%d fields546..548=%d/%d/%d", unsafe.Sizeof(update),
			unsafe.Offsetof(update.Field545), unsafe.Sizeof(update.Field545), unsafe.Offsetof(update.Field546),
			unsafe.Offsetof(update.Field547), unsafe.Offsetof(update.Field548))
	}
	player := new(PlayerUpdateData)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(player)) <= math.MaxUint32 {
		t.Fatalf("bot player record=%p, want above 4 GiB", player)
	}
	reflect.ValueOf(&update).Elem().FieldByName("Field545").Set(reflect.ValueOf(player))
	if got := *(*unsafe.Pointer)(unsafe.Add(unsafe.Pointer(&update), wantLink)); got != unsafe.Pointer(player) {
		t.Fatalf("native bot link=%p, want %p", got, player)
	}
	runtime.KeepAlive(player)
}
